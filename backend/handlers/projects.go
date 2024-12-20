package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/services"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func (h *Handler) HandleGETProjectState(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	idParam := ctx.Param("id")

	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		project, err := h.service.SelectProjectByIDAndUserID(tx, projectID, userID)
		if err != nil {

			return http.StatusNotFound, err
		}
		project.Hook = fmt.Sprintf("%s/v1/hook/%d", config.Host, project.ID)

		state, err := project.UPN.GetContainersState()
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
		projects, err := h.service.SelectProjects(userID)
		if err != nil {
			return http.StatusForbidden, err
		}
		for i := range projects {
			projects[i].Hook = fmt.Sprintf("%s/v1/hook/%d", config.Host, projects[i].ID)
		}
		ctx.JSON(http.StatusOK, projects)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	idParam := ctx.Param("id")

	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		project, err := h.service.SelectProjectByIDAndUserID(tx, projectID, userID)
		if err != nil {
			return http.StatusNotFound, err
		}
		project.Hook = fmt.Sprintf("%s/v1/hook/%d", config.Host, project.ID)

		ctx.JSON(http.StatusOK, project)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	idParam := ctx.Param("id")

	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		project, err := h.service.SelectProjectByIDAndUserID(tx, projectID, userID)
		if err != nil {
			return http.StatusNotFound, err
		}

		pPath := project.UPN.GetProjectPath()

		err = h.service.DeleteProjectByIDAndUserID(tx, projectID, userID)
		if err != nil {
			slog.Error(fmt.Sprintf("unable to delete Project by id: %v", err))
			return http.StatusInternalServerError, err
		}

		if err := project.UPN.StopContainers(); err != nil {
			slog.Error(fmt.Sprintf("unable to stop containers: %v", err))
			return http.StatusInternalServerError, err
		}

		if err := utils.DeleteFolder(pPath); err != nil {
			slog.Error(fmt.Sprintf("unable to delete folder: %v", err))
			return http.StatusInternalServerError, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandlePOSTProject(c *gin.Context) {
	var p services.Project
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
	p.UPN = services.UPN(fmt.Sprintf("%s-%s", utils.GenerateRandomName(), upnSuffix))
	p.Path = p.UPN.GetProjectPath()

	tx, err := h.store.DB.Beginx()
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	err = h.service.SaveProject(&p, tx)
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

	if err := h.service.PrepareProject(&p); err != nil {
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

	c.JSON(http.StatusCreated, gin.H{
		"id": p.ID,
	})
}

func (h *Handler) HandlePUTProject(c *gin.Context) {
	userID := userIDFromSession(c)

	var p services.Project
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
	idParam := ctx.Param("id")

	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "invalid type", err)
		return
	}

	tx, err := h.store.DB.Beginx()
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	project, err := h.service.SelectProjectByIDAndAccessToken(tx, projectID, accessToken)
	if err != nil {
		h.abortWithError(ctx, http.StatusNotFound, "unable find project", err)
		return
	}

	queryParams := ctx.Request.URL.Query()

	for i, service := range project.Services {
		for key, values := range queryParams {
			if service.Name == key {
				project.Services[i].ImageTag = values[0]
			}
		}
	}

	if err := h.updateAndRestartContainers(ctx, project, tx); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		return
	}

	if err := tx.Commit(); err != nil {
		project.UPN.RollbackToPreviousState()
		h.abortWithError(ctx, http.StatusInternalServerError, "failed to commit project update", err)
		err = tx.Rollback()
		if err != nil {
			slog.Error("unable to rollback transaction", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id": project.ID,
	})
}

func (h *Handler) updateAndRestartContainers(c *gin.Context, p *services.Project, tx *sqlx.Tx) error {
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

	if err := h.service.UpdateProject(p, tx); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to update project")
	}

	if err := h.service.PrepareProject(p); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to prepare project")
	}

	if err := p.UPN.StartContainers(p.CTN, p.DockerCredentials); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to start containers")
	}

	return nil
}
