package handlers

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/devs-group/sloth/config"
	"github.com/devs-group/sloth/pkg"
	"github.com/devs-group/sloth/repository"

	"github.com/devs-group/sloth/database"
	"github.com/devs-group/sloth/pkg/compose"
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
	projects, err := repository.SelectProjects(userID, h.store)
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

	err := p.SelectProjectByUPNOrAccessToken(h.store)
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

	accessToken, err := pkg.RandStringRunes(accessTokenLen)
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to generate access token", err)
		return
	}

	upnSuffix, err := pkg.RandStringRunes(uniqueProjectSuffixLen)
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to generate unique project name suffix", err)
		return
	}

	upn := repository.UPN(fmt.Sprintf("%s-%s", pkg.GenerateRandomName(), upnSuffix))

	p.UPN = upn
	p.AccessToken = accessToken
	p.Path = upn.GetProjectPath()

	if err := pkg.PrepareProject(&p, upn); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "Failed to prepare project", err)
		return
	}

	err = p.SaveProject(h.store)
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to create project", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":        accessToken,
		"unique_project_name": upn,
	})
}

func (h *Handler) HandlePUTProject(c *gin.Context) {
	userID := userIDFromSession(c)
	upn := repository.UPN(c.Param("upn"))

	var p repository.Project
	p.UserID = userID
	p.UPN = upn
	p.Hook = fmt.Sprintf("%s/v1/hook/%s", config.Host, p.UPN)

	if err := c.BindJSON(&p); err != nil {
		h.abortWithError(c, http.StatusBadRequest, "Failed to parse request body", err)
		return
	}

	if err := h.upateAndRestartContainers(c, &p, upn); err != nil {
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

	if err := p.SelectProjectByUPNOrAccessToken(h.store); err != nil {
		h.abortWithError(ctx, http.StatusUnauthorized, "unable to find project by name and access token", err)
		return
	}

	queryParams := ctx.Request.URL.Query()

	for i, service := range p.Services {
		for key, values := range queryParams {
			if service.Name == key {
				slog.Info("Servicename", "tag", values[0], "image", p.Services[i].Image)
				p.Services[i].ImageTag = values[0]
			}
		}
	}

	if err := h.upateAndRestartContainers(ctx, &p, upn); err != nil {
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

	err := p.SelectProjectByUPNOrAccessToken(h.store)
	if err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to find project by upn", err)
		return
	}

	state, err := pkg.GetContainersState(upn)
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
		Hook:   fmt.Sprintf("%s/v1/hook/%s", config.Host, upn),
	}

	if err := pkg.StopContainers(p.Path); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to stop containers", err)
		return
	}

	err := p.DeleteProjectByUPNWithTx(h.store, func() error {
		return pkg.RenameFolder(ppath, deletedProjectPath)
	})
	if err != nil {
		err = pkg.RenameFolder(deletedProjectPath, ppath)
		if err != nil {
			slog.Error("unable to rename folder", "err", err)
		}
		h.abortWithError(c, http.StatusInternalServerError, "unable to delete project", err)
		return
	}

	// Delete the temp folder in background
	go func() {
		err := pkg.DeleteFolder(deletedProjectPath)
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

	err := p.SelectProjectByUPNOrAccessToken(h.store)
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

func (h *Handler) upateAndRestartContainers(c *gin.Context, p *repository.Project, upn repository.UPN) error {
	if err := pkg.StopContainers(upn.GetProjectPath()); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "Failed to stop containers", err)
		return err
	}

	if err := pkg.BackupCurrentFiles(upn); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "Failed to backup current files", err)
		return err
	}
	defer pkg.DeleteBackupFiles(upn)

	if err := pkg.PrepareProject(p, upn); err != nil {
		pkg.RollbackToPreviousState(upn)
		h.abortWithError(c, http.StatusInternalServerError, "Failed to prepare project", err)
		return err
	}

	dc, err := compose.FromString(p.DCJ)
	if err != nil {
		pkg.RollbackToPreviousState(upn)
		h.abortWithError(c, http.StatusInternalServerError, "Failed to create compose file", err)
		return err
	}

	if err := pkg.StartContainers(upn.GetProjectPath(), dc.Services, p.DockerCredentials); err != nil {
		pkg.RollbackToPreviousState(upn)
		h.abortWithError(c, http.StatusInternalServerError, "Failed to start containers", err)
		return err
	}

	if err := p.UpdateProject(h.store); err != nil {
		pkg.RollbackToPreviousState(upn)
		h.abortWithError(c, http.StatusInternalServerError, "Failed to update project", err)
		return err
	}

	return nil
}
