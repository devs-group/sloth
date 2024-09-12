package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/models"
	"github.com/devs-group/sloth/backend/pkg/email"
	"github.com/devs-group/sloth/backend/services"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/gin-gonic/gin"
)

func (h *Handler) HandlePOSTOrganisation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisation := models.Organisation{
		OwnerID: userID,
	}
	if err := ctx.BindJSON(&organisation); err != nil {
		UnableToParseRequestBody(ctx, err)
		return
	}
	o, err := h.service.CreateOrganisation(organisation)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to create organisation", err)
		return
	}
	ctx.JSON(http.StatusOK, o)
	return
}

func (h *Handler) HandleDELETEOrganisation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	idParam := ctx.Param("id")
	organisationID, err := strconv.Atoi(idParam)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "invalid organisation id", err)
		return
	}
	err = h.service.DeleteOrganisation(userID, organisationID)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to delete organisation", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (h *Handler) HandleGETOrganisations(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisations, err := h.service.SelectOrganisations(userID)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to get organisation", err)
		return
	}
	ctx.JSON(http.StatusOK, organisations)
}

func (h *Handler) HandleGETOrganisation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	idParam := ctx.Param("id")
	organisationID, err := strconv.Atoi(idParam)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "invalid organisation id", err)
		return
	}
	organisation, err := h.service.SelectOrganisation(organisationID, userID)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to get organisation", err)
		return
	}
	ctx.JSON(http.StatusOK, organisation)
}

func (h *Handler) HandleDELETEMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisationID := ctx.Param("id")
	memberID := ctx.Param("member_id")
	err := h.service.DeleteMember(userID, memberID, organisationID)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to delete member", err)
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *Handler) HandlePUTInvitation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	var invite services.Invitation
	if err := ctx.BindJSON(&invite); err != nil {
		UnableToParseRequestBody(ctx, err)
		return
	}
	invitationToken, err := utils.RandStringRunes(256)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to generate random invitation token", err)
		return
	}
	err = email.SendMail(config.EmailInvitationURL, invitationToken, invite.Email)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to send invitation email", err)
		return
	}

	err = h.service.SaveInvitation(userID, invite.Email, invite.OrganisationID, invitationToken)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to save invitation", err)
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *Handler) HandlePUTMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	memberID := ctx.Param("member_id")

	var invite services.Invitation
	if err := ctx.BindJSON(&invite); err != nil {
		UnableToParseRequestBody(ctx, err)
		return
	}
	if userID != memberID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you don't have permissions to add this user"})
		return
	}
	err := h.service.PutMember(memberID, invite.OrganisationID)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to add member", err)
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *Handler) HandleGETInvitations(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisationID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to parse id param", err)
		return
	}
	invites, err := h.service.GetInvitations(userID, organisationID)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to get invitations", err)
		return
	}
	ctx.JSON(http.StatusOK, invites)
}

func (h *Handler) HandleDELETEWithdrawInvitation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	type WithdrawInvitation struct {
		Email          string `json:"email"`
		OrganisationID int    `json:"organisation_id"`
	}

	var withdrawInvitation WithdrawInvitation
	if err := ctx.BindJSON(&withdrawInvitation); err != nil {
		UnableToParseRequestBody(ctx, err)
		return
	}
	err := h.service.WithdrawInvitation(userID, withdrawInvitation.Email, withdrawInvitation.OrganisationID)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to withdraw invitation", err)
		return
	}
}

func (h *Handler) HandlePOSTAcceptInvitation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	type AcceptInvitationRequest struct {
		UserID          int    `json:"user_id"`
		InvitationToken string `json:"invitation_token"`
	}

	var acceptRequest AcceptInvitationRequest
	if err := ctx.BindJSON(&acceptRequest); err != nil {
		UnableToParseRequestBody(ctx, err)
		return
	}

	if userID != strconv.Itoa(acceptRequest.UserID) {
		HandleError(ctx, http.StatusForbidden, "unauthorized to accept invitation", errors.New("requested user id is not equal to logged in user id"))
		return
	}

	email, err := getUserMailFromSession(ctx)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to obtain user email", err)
		return
	}

	_, err = h.service.GetInvitation(email, acceptRequest.InvitationToken)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to get invitation", err)
		return
	}

	accepted, err := h.service.AcceptInvitation(userID, email, acceptRequest.InvitationToken)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to accept invitation", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"accepted": accepted})
}

func (h *Handler) HandleGETLeaveOrganisation(ctx *gin.Context) {
	// TODO
}

func (h *Handler) HandleGETOrganisationProjects(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisationID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Status(http.StatusBadRequest)
	}
	projects, err := h.service.GetProjectsByOrganisationID(userID, organisationID)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to get projects", err)
		return
	}
	ctx.JSON(http.StatusOK, projects)
}

func (h *Handler) HandlePUTOrganisationProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	type OrganisationProjectPut struct {
		UPN            string `json:"upn" binding:"required"`
		OrganisationID int    `json:"organisation_id" binding:"required"`
	}
	var g OrganisationProjectPut
	if err := ctx.BindJSON(&g); err != nil {
		UnableToParseRequestBody(ctx, err)
		return
	}
	err := h.service.AddProjectToOrganisationByUPN(userID, g.OrganisationID, g.UPN)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to add project to organisation", err)
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *Handler) HandleDELETEOrganisationProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	type OrganisationProjectDelete struct {
		UPN            string `json:"upn"`
		OrganisationID int    `json:"organisation_id"`
	}
	var g OrganisationProjectDelete
	if err := ctx.BindJSON(&g); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	err := h.service.DeleteProject(userID, g.OrganisationID, g.UPN)
	if err != nil {
		HandleError(ctx, http.StatusInternalServerError, "unable to delete project", err)
		return
	}
	ctx.Status(http.StatusOK)
}
