package handlers

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/repository"
	"github.com/devs-group/sloth/backend/utils"

	"github.com/devs-group/sloth/backend/database"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/gin-gonic/gin"
)

const accessTokenLen = 12
const uniqueProjectSuffixLen = 10

type Handler struct {
	store    *database.Store
	vueFiles embed.FS
	upgrader websocket.Upgrader
}

func New(store *database.Store, vueFiles embed.FS) Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	// TODO: Loop over list of trusted origins instead returning true for all origins.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return Handler{
		store:    store,
		vueFiles: vueFiles,
		upgrader: upgrader,
	}
}

func (h *Handler) abortWithError(c *gin.Context, statusCode int, message string, err error) {
	slog.Error(message, "err", err)
	c.AbortWithStatus(statusCode)
}

func (h *Handler) HandleGETProjects(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	projects, err := repository.SelectProjects(userID, tx)
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to select projects", err)
		return
	}
	ctx.JSON(http.StatusOK, projects)
}

func (h *Handler) HandleGETProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	upn := repository.UPN(ctx.Param("upn"))

	p := repository.Project{
		UserID: userID,
		UPN:    upn,
		Hook:   fmt.Sprintf("%s/v1/hook/%s", config.Host, upn),
	}

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	err = p.SelectProjectByUPNOrAccessToken(tx)
	if err != nil {
		h.abortWithError(ctx, http.StatusNotFound, "unable to select project", err)
		return
	}

	ctx.JSON(http.StatusOK, p)
}

func (h *Handler) HandlePOSTProject(c *gin.Context) {
	var p repository.Project
	userID := userIDFromSession(c)
	p.UserID = userID

	if err := c.BindJSON(&p); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to parse request body", err)
		return
	}

	accessToken, err := utils.RandStringRunes(accessTokenLen)
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to generate access token", err)
		return
	}

	upnSuffix, err := utils.RandStringRunes(uniqueProjectSuffixLen)
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to generate unique project name suffix", err)
		return
	}

	p.AccessToken = accessToken
	p.UPN = repository.UPN(fmt.Sprintf("%s-%s", utils.GenerateRandomName(), upnSuffix))
	p.Path = p.UPN.GetProjectPath()

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	err = p.SaveProject(tx)
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to store project", err)
		utils.DeleteFolder(p.UPN.GetProjectPath())
		return
	}

	if err := p.PrepareProject(); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "Failed to prepare project", err)
		return
	}

	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to initiate transaction", err)
		utils.DeleteFolder(p.UPN.GetProjectPath())
		return
	}

	if err := tx.Commit(); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to store data to database", err)
		utils.DeleteFolder(p.UPN.GetProjectPath())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":        accessToken,
		"unique_project_name": p.UPN,
	})
}

func (h *Handler) HandlePUTProject(c *gin.Context) {
	userID := userIDFromSession(c)

	var p repository.Project
	if err := c.BindJSON(&p); err != nil {
		h.abortWithError(c, http.StatusBadRequest, "Failed to parse request body", err)
		return
	}

	p.UserID = userID
	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()
	if err != nil {
		h.abortWithError(c, http.StatusNotFound, "unable to start transaction project", err)
		return
	}

	if err := h.updateAndRestartContainers(c, &p, tx); err != nil {
		return // error and log already handled in function
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) HandleGetHook(ctx *gin.Context) {
	accessToken := ctx.GetHeader("X-Access-Token")
	if accessToken == "" {
		h.abortWithError(ctx, http.StatusUnauthorized, "X-Access-Token header is required", nil)
		return
	}

	upn := repository.UPN(ctx.Param("upn"))
	p := repository.Project{
		AccessToken: accessToken,
		UPN:         upn,
	}

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	if err := p.SelectProjectByUPNOrAccessToken(tx); err != nil {
		h.abortWithError(ctx, http.StatusUnauthorized, "unable to find project by name and access token", err)
		return
	}

	queryParams := ctx.Request.URL.Query()

	for i, service := range p.Services {
		for key, values := range queryParams {
			if service.Name == key {
				p.Services[i].ImageTag = values[0]
			}
		}
	}

	if err := h.updateAndRestartContainers(ctx, &p, tx); err != nil {
		return // error and log already handled in function
	}

	ctx.JSON(http.StatusOK, gin.H{
		"upn": p.UPN,
	})
}

func (h *Handler) HandleGETProjectState(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	upn := repository.UPN(ctx.Param("upn"))
	p := repository.Project{
		UserID: userID,
		UPN:    upn,
		Hook:   fmt.Sprintf("%s/v1/hook/%s", config.Host, upn),
	}

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	err = p.SelectProjectByUPNOrAccessToken(tx)
	if err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to find project by upn", err)
		return
	}

	state, err := p.UPN.GetContainersState()
	if err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to get project state", err)
		return
	}

	ctx.JSON(http.StatusOK, state)
}

func (h *Handler) HandleDELETEProject(c *gin.Context) {
	userID := userIDFromSession(c)
	upn := repository.UPN(c.Param("upn"))
	ppath := upn.GetProjectPath()
	deletedProjectPath := fmt.Sprintf("%s-deleted", ppath)

	p := repository.Project{
		UserID: userID,
		UPN:    upn,
		Path:   ppath,
	}

	if err := p.UPN.StopContainers(); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to stop containers", err)
		return
	}

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to start transaction", err)
	}
	err = p.DeleteProjectByUPNWithTx(tx, func() error {
		return utils.RenameFolder(ppath, deletedProjectPath)
	})

	if err != nil {
		err = utils.RenameFolder(deletedProjectPath, ppath)
		if err != nil {
			slog.Error("unable to rename folder", "err", err)
		}
		h.abortWithError(c, http.StatusInternalServerError, "unable to delete project", err)
		return
	}

	if err = tx.Commit(); err != nil {
		err = utils.RenameFolder(deletedProjectPath, ppath)
		if err != nil {
			slog.Error("unable to rename folder", "err", err)
		}
		h.abortWithError(c, http.StatusInternalServerError, "unable to delete project", err)
		return
	}

	// Delete the temp folder in background
	go func() {
		err := utils.DeleteFolder(deletedProjectPath)
		if err != nil {
			slog.Error("unable to delete folder", "path", deletedProjectPath, "err", err)
		}
	}()

	c.Status(http.StatusOK)
}

func (h *Handler) HandleStreamServiceLogs(c *gin.Context) {
	userID := userIDFromSession(c)
	upn := repository.UPN(c.Param("upn"))
	s := c.Param("service")

	p := repository.Project{
		UserID: userID,
		UPN:    upn,
		Path:   upn.GetProjectPath(),
		Hook:   fmt.Sprintf("%s/v1/hook/%s", config.Host, upn),
	}

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	err = p.SelectProjectByUPNOrAccessToken(tx)
	if err != nil {
		h.abortWithError(c, http.StatusBadRequest, "unable to find project by upn", err)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to upgrade http to ws", err)
		return
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			h.abortWithError(c, http.StatusInternalServerError, "unable to close websocket connection", err)
			return
		}
	}(conn)

	ppath := upn.GetProjectPath()
	out := make(chan string)
	go func() {
		err := compose.Logs(ppath, s, out)
		if err != nil {
			h.abortWithError(c, http.StatusInternalServerError, "unable to stream logs", err)
			return
		}
	}()

	line := 0
	for o := range out {
		line++
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d %s", line, o)))
	}
}

func (h *Handler) updateAndRestartContainers(c *gin.Context, p *repository.Project, tx *sqlx.Tx) error {
	if err := p.UPN.StopContainers(); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "Failed to stop containers", err)
		return err
	}

	if err := p.UPN.BackupCurrentFiles(); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "Failed to backup current files", err)
		return err
	}
	defer p.UPN.DeleteBackupFiles()

	if err := p.UpdateProject(tx); err != nil {
		tx.Rollback()
		p.UPN.RollbackToPreviousState()
		h.abortWithError(c, http.StatusInternalServerError, "Failed to update project", err)
		return err
	}

	if err := p.PrepareProject(); err != nil {
		p.UPN.RollbackToPreviousState()
		h.abortWithError(c, http.StatusInternalServerError, "Failed to prepare project", err)
		return err
	}

	if err := p.UPN.StartContainers(p.CTN, p.DockerCredentials); err != nil {
		p.UPN.RollbackToPreviousState()
		h.abortWithError(c, http.StatusInternalServerError, "Failed to start containers", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		p.UPN.RollbackToPreviousState()
		h.abortWithError(c, http.StatusInternalServerError, "Failed to update project", err)
		return err
	}
	return nil
}