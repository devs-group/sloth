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
)

func (h *Handler) HandleGetProjectState(ctx *gin.Context) {
	cfg := config.GetConfig()

	organisationID := currentOrganisationIDFromSession(ctx)
	idParam := ctx.Param("id")

	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	project, err := h.service.SelectProjectByIDAndOrganisationID(projectID, organisationID)
	if err != nil {
		h.abortWithError(ctx, http.StatusNotFound, "unable to find project", err)
		return
	}
	project.Hook = fmt.Sprintf("%s/v1/hook/%d", cfg.BackendUrl, project.ID)

	state, err := project.UPN.GetContainersState()
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to get container state", err)
		return
	}

	ctx.JSON(http.StatusOK, state)
}

func (h *Handler) HandleListProjects(ctx *gin.Context) {
	cfg := config.GetConfig()

	userID := userIDFromSession(ctx)
	organisationID := currentOrganisationIDFromSession(ctx)

	projects, err := h.service.ListProjects(userID, organisationID)
	if err != nil {
		h.abortWithError(ctx, http.StatusNotFound, "unable to find projects", err)
		return
	}
	for i := range projects {
		projects[i].Hook = fmt.Sprintf("%s/v1/hook/%d", cfg.BackendUrl, projects[i].ID)
	}
	ctx.JSON(http.StatusOK, projects)
}

func (h *Handler) HandleGetProject(ctx *gin.Context) {
	cfg := config.GetConfig()

	organisationID := currentOrganisationIDFromSession(ctx)
	idParam := ctx.Param("id")

	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	project, err := h.service.SelectProjectByIDAndOrganisationID(projectID, organisationID)
	if err != nil {
		h.abortWithError(ctx, http.StatusNotFound, "unable to find project", err)
		return
	}
	project.Hook = fmt.Sprintf("%s/v1/hook/%d", cfg.BackendUrl, project.ID)

	slog.Info("services", "project.Services", project.Services)
	ctx.JSON(http.StatusOK, project)

}

func (h *Handler) HandleDeleteProject(ctx *gin.Context) {
	organisationID := currentOrganisationIDFromSession(ctx)
	idParam := ctx.Param("id")

	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	project, err := h.service.SelectProjectByIDAndOrganisationID(projectID, organisationID)
	if err != nil {
		h.abortWithError(ctx, http.StatusNotFound, "unable to find project", err)
		return
	}

	pPath := project.UPN.GetProjectPath()

	err = h.service.DeleteProjectByIDAndOrganisationID(projectID, organisationID)
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to delete project", err)
		return
	}

	if err := project.UPN.StopContainers(); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to stop containers", err)
		return
	}

	if err := utils.DeleteFolder(pPath); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to delete folder", err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *Handler) HandleCreateProject(c *gin.Context) {
	var p services.Project
	currentOrganisationID := currentOrganisationIDFromSession(c)

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

	err = h.service.SaveProject(&p, currentOrganisationID)
	if err != nil {
		slog.Error("unable to save project", "err", err)
		h.abortWithError(c, http.StatusInternalServerError, "unable to save project", err)
		err = utils.DeleteFolder(p.UPN.GetProjectPath())
		if err != nil {
			h.abortWithError(c, http.StatusInternalServerError, "unable to delete folder after save project failed", err)
			return
		}
		return
	}

	if err := h.service.PrepareProject(&p); err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to prepare project", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": p.ID,
	})
}

func (h *Handler) HandleUpdateProject(c *gin.Context) {
	currentOrganisationID := currentOrganisationIDFromSession(c)
	var p services.Project
	if err := c.BindJSON(&p); err != nil {
		h.abortWithError(c, http.StatusBadRequest, "failed to parse request body", err)
		return
	}
	p.OrganisationID = currentOrganisationID
	if err := h.updateAndRestartContainers(c, &p); err != nil {
	  slog.Error("unable to update and restart containers", "err", err)
		h.abortWithError(c, http.StatusInternalServerError, "", err)
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

	project, err := h.service.SelectProjectByIDAndAccessToken(projectID, accessToken)
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

	if err := h.updateAndRestartContainers(ctx, project); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to update and restart containers", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id": project.ID,
	})
}

func (h *Handler) updateAndRestartContainers(c *gin.Context, p *services.Project) error {
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

	slog.Debug("project before", "id", slog.Any("services", p.Services))
	if err := h.service.UpdateProject(p); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to update project")
	}
	slog.Debug("project after", "id", slog.Any("services", p.Services))

	if err := h.service.PrepareProject(p); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to prepare project")
	}

	if err := p.UPN.StartContainers(p.ComposeServices, p.DockerCredentials); err != nil {
		p.UPN.RollbackToPreviousState()
		return errors.Wrap(err, "unable to start containers")
	}

	return nil
}
