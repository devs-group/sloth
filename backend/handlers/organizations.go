package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/email"
	"github.com/devs-group/sloth/backend/pkg/github"
	"github.com/devs-group/sloth/backend/repository"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Validate the organization name is not empty or just whitespace.
func (h *Handler) validateOrganizationName(ctx *gin.Context, organizationName string) bool {
	if strings.TrimSpace(organizationName) == "" {
		h.abortWithError(ctx, http.StatusBadRequest, "Organization name cannot be empty or whitespace", fmt.Errorf("malformed request"))
		return false
	}
	return true
}

func (h *Handler) HandlePOSTOrganization(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organization := repository.Organization{
		OwnerID: userID,
	}

	if err := ctx.BindJSON(&organization); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	if !h.validateOrganizationName(ctx, organization.Name) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		err := organization.CreateOrganization(tx)
		if err != nil {
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, organization)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETOrganizations(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		organizations, err := repository.SelectOrganizations(userID, tx)
		if err != nil {

			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, organizations)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETOrganization(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organizationName := ctx.Param("organization_name")

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		g := repository.Organization{
			OwnerID: userID,
			Name:    organizationName,
		}

		if err := g.SelectOrganization(tx); err != nil {
			return http.StatusForbidden, err
		}
		ctx.JSON(http.StatusOK, g)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEOrganization(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organizationName := ctx.Param("organization_name")

	if !h.validateOrganizationName(ctx, organizationName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		g := repository.Organization{
			Name:    organizationName,
			OwnerID: userID,
		}

		if err := g.DeleteOrganization(tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organizationName := ctx.Param("organization_name")
	memberID := ctx.Param("member_id")

	if !h.validateOrganizationName(ctx, organizationName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := repository.DeleteMember(userID, memberID, organizationName, tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandlePUTInvitation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	var invite repository.Invitation

	if err := ctx.BindJSON(&invite); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}
	if !h.validateOrganizationName(ctx, invite.OrganizationName) {
		return
	}
	invitationToken, err := utils.RandStringRunes(256)
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "cant create invitation token", err)
		return
	}

	err = email.SendMail(config.EmailInvitationURL, invitationToken, invite.Email)
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "Cant send invitation mail", err)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := repository.PutInvitation(userID, invite.Email, invite.OrganizationName, invitationToken, tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandlePUTMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	memberID := ctx.Param("member_id")

	var invite repository.Invitation

	if err := ctx.BindJSON(&invite); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	if userID != memberID {
		h.abortWithError(ctx, http.StatusBadRequest,
			"you dont have this permission to do that.",
			fmt.Errorf("user try to add different user without permission"))
		return
	}

	if !h.validateOrganizationName(ctx, invite.OrganizationName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := repository.PutMember(memberID, invite.OrganizationName, tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETInvitations(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		invites, err := repository.GetInvitations(userID, tx)
		if err != nil {
			return http.StatusForbidden, err
		}
		ctx.JSON(http.StatusOK, invites)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETMembersForInvitation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organizationName := ctx.Param("organization_name")
	memberSearch := ctx.Param("member_search")

	if len(memberSearch) < 3 {
		h.abortWithError(ctx, http.StatusBadRequest, "Search query must be at least 3 characters long", nil)
		return
	}

	if !h.validateOrganizationName(ctx, organizationName) {
		return
	}
	userHasRights := false
	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if userHasRights = repository.CheckIsMemberOfOrganization(userID, organizationName, tx); !userHasRights {
			return http.StatusForbidden, fmt.Errorf("Insufficient rights userID: %s organization: %s", userID, organizationName)
		}
		return http.StatusOK, nil
	})

	if !userHasRights {
		// if user does not have enough rights error was already send
		return
	}

	res, err := github.SearchGitHubUsers(memberSearch)
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "Github fetch failed", err)
	}

	ctx.JSON(http.StatusOK, res)
}

func (h *Handler) HandlePOSTAcceptInvitation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	type AcceptInvitationRequest struct {
		UserID          string `json:"user_id"`
		InvitationToken string `json:"invitation_token"`
	}

	var acceptRequest AcceptInvitationRequest
	if err := ctx.BindJSON(&acceptRequest); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	if userID != acceptRequest.UserID {
		h.abortWithError(ctx, http.StatusForbidden, "not authorized", fmt.Errorf("not authorized to accept inviation"))
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if !isInvitatedToOrganization(acceptRequest.InvitationToken, tx, ctx) {
			slog.Info("Error", "err", "user does not have rights")
			return http.StatusForbidden, fmt.Errorf("not authorized to accept inviation")
		}

		ok, err := repository.AcceptInvitation(userID, userMailFromSession(ctx), acceptRequest.InvitationToken, tx)
		if err != nil || !ok {
			h.abortWithError(ctx, http.StatusForbidden, "unable to process invitation", fmt.Errorf("unable to process invitation"))
			return http.StatusForbidden, fmt.Errorf("unable to process invitation")
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func isInvitatedToOrganization(token string, tx *sqlx.Tx, ctx *gin.Context) bool {
	loggedInEmail := userMailFromSession(ctx)

	inviation, err := repository.GetInvitation(loggedInEmail, token, tx)
	if err != nil {
		return false
	}

	if inviation != nil {
		return true
	}

	return false
}

func (h *Handler) HandleGETLeaveOrganization(ctx *gin.Context) {
	// TODO
}

func (h *Handler) HandleGetOrganizationProjects(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organizationName := ctx.Param("organization_name")

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		projects, err := repository.GetProjectsByOrganizationName(userID, organizationName, tx)
		if err != nil {
			slog.Error("error", "cant get projects", err)
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, projects)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandlePUTOrganizationProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	type OrganizationProjectPut struct {
		UPN              string `json:"upn"`
		OrganizationName string `json:"organization_name"`
	}
	var g OrganizationProjectPut

	if err := ctx.BindJSON(&g); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		ok, err := repository.AddOrganizationProjectByUPN(userID, g.OrganizationName, g.UPN, tx)
		if err != nil {
			return http.StatusForbidden, err
		}

		if !ok {
			return http.StatusInternalServerError, fmt.Errorf("unable to add project")
		}

		ctx.JSON(http.StatusOK, userID)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEOrganizationProject(ctx *gin.Context) {
	slog.Info("METHOD NOT IMPLEMENTED", "NOT IMPLEMENTED", "DELETE GROUP PROJECT")
	// TODO
}
