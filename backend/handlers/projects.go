package handlers

import (
	"fmt"
	"github.com/pkg/errors"
	"log/slog"
	"net/http"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/repository"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func (h *Handler) HandleGETProjectState(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	upnValue := ctx.Param("upn")
	upn := repository.UPN(upnValue)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		p := repository.Project{
			UserID: userID,
			UPN:    upn,
			Hook:   fmt.Sprintf("%s/v1/hook/%s", config.Host, upnValue),
		}
		if err := p.SelectProjectByUPNOrAccessToken(tx); err != nil {
			return http.StatusNotFound, err
		}

		state, err := p.UPN.GetContainersState()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		ctx.JSON(http.StatusOK, state)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETProjects(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		projects, err := repository.SelectProjects(userID, tx)
		if err != nil {
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, projects)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	upn := repository.UPN(ctx.Param("upn"))

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		p := repository.Project{
			UserID: userID,
			UPN:    upn,
			Hook:   fmt.Sprintf("%s/v1/hook/%s", config.Host, upn),
		}

		err := p.SelectProjectByUPNOrAccessToken(tx)
		if err != nil {
			return http.StatusNotFound, err
		}

		ctx.JSON(http.StatusOK, p)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEProject(c *gin.Context) {
	userID := userIDFromSession(c)
	upn := repository.UPN(c.Param("upn"))
	pPath := upn.GetProjectPath()

	p := repository.Project{
		UserID: userID,
		UPN:    upn,
		Path:   pPath,
	}

	h.WithTransaction(c, func(tx *sqlx.Tx) (int, error) {
		err := p.DeleteProjectByUPNWithTx(tx)
		if err != nil {
			slog.Error("Error", "unable to delete Project by upn", err)
			return http.StatusInternalServerError, err
		}

		if err := p.UPN.StopContainers(); err != nil {
			slog.Error("Error", "unable to stop containers", err)
			return http.StatusInternalServerError, err
		}

		if err := utils.DeleteFolder(pPath); err != nil {
			slog.Error("unable to delete folder", err)
			return http.StatusInternalServerError, err
		}

		c.Status(http.StatusOK)
		return http.StatusOK, nil
	})
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

	// TODO: @4ddev probably also use the withTransaction method here?
	tx, err := h.store.DB.Beginx()
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	err = p.SaveProject(tx)
	if err != nil {
		slog.Error("unable to save project", err)
		h.abortWithError(c, http.StatusInternalServerError, "unable to save project", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		err = utils.DeleteFolder(p.UPN.GetProjectPath())
		if err != nil {
			slog.Error("unable to delete folder, after saving project failed", err)
		}
		return
	}

	if err := p.PrepareProject(); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to prepare project", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to store data", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		err = utils.DeleteFolder(p.UPN.GetProjectPath())
		if err != nil {
			slog.Error("unable to delete folder, after transaction Commit failed", err)
		}
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
		h.abortWithError(c, http.StatusBadRequest, "failed to parse request body", err)
		return
	}

	p.UserID = userID
	tx, err := h.store.DB.Beginx()
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	if err := h.updateAndRestartContainers(c, &p, tx); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		p.UPN.RollbackToPreviousState()
		h.abortWithError(c, http.StatusInternalServerError, "failed to commit project update", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) HandleGetProjectHook(ctx *gin.Context) {
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
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	if err := p.SelectProjectByUPNOrAccessToken(tx); err != nil {
		h.abortWithError(ctx, http.StatusNotFound, "unable to find project by name and access token", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
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
		h.abortWithError(ctx, http.StatusInternalServerError, "", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		p.UPN.RollbackToPreviousState()
		h.abortWithError(ctx, http.StatusInternalServerError, "failed to commit project update", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"upn": p.UPN,
	})
}

func (h *Handler) updateAndRestartContainers(c *gin.Context, p *repository.Project, tx *sqlx.Tx) error {
	if isRunning, err := p.UPN.IsOneContainerRunning(); err != nil || isRunning {
		if err != nil {
			return errors.Wrap(err, "unable to receive container states")
		}
		if err := p.UPN.StopContainers(); err != nil {
			return errors.Wrap(err, "unable to stop containers")
		}
	}

	if err := p.UPN.BackupCurrentFiles(); err != nil {
		return errors.Wrap(err, "unable to backup current files")
	}
	defer p.UPN.DeleteBackupFiles()

	if err := p.UpdateProject(tx); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to update project")
	}

	if err := p.PrepareProject(); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to prepare project")
	}

	if err := p.UPN.StartContainers(p.CTN, p.DockerCredentials); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to start containers")
	}

	return nil
}
