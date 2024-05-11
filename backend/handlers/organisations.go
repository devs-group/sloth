package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/email"
	"github.com/devs-group/sloth/backend/pkg/github"
	"github.com/devs-group/sloth/backend/repository"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Validate the organisation name is not empty or just whitespace.
func (h *Handler) validateOrganisationName(ctx *gin.Context, organisationName string) bool {
	if strings.TrimSpace(organisationName) == "" {
		h.abortWithError(ctx, http.StatusBadRequest, "Organisation name cannot be empty or whitespace", fmt.Errorf("malformed request"))
		return false
	}
	return true
}

func (h *Handler) HandlePOSTOrganisation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisation := repository.Organisation{
		OwnerID: userID,
	}

	if err := ctx.BindJSON(&organisation); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	if !h.validateOrganisationName(ctx, organisation.Name) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		err := organisation.CreateOrganisation(tx)
		if err != nil {
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, organisation)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETOrganisations(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		organisations, err := repository.SelectOrganisations(userID, tx)
		if err != nil {
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, organisations)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETOrganisation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	idParam := ctx.Param("id")

	organisationID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		organisation, err := repository.SelectOrganisation(tx, organisationID, userID)
		if err != nil {
			return http.StatusNotFound, err
		}
		ctx.JSON(http.StatusOK, organisation)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEOrganisation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	idParam := ctx.Param("id")

	organisationID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		g := repository.Organisation{
			ID:      organisationID,
			OwnerID: userID,
		}

		if err := g.DeleteOrganisation(tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisationName := ctx.Param("organisation_name")
	memberID := ctx.Param("member_id")

	if !h.validateOrganisationName(ctx, organisationName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := repository.DeleteMember(userID, memberID, organisationName, tx); err != nil {
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
	if !h.validateOrganisationName(ctx, invite.OrganisationName) {
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
		if err := repository.PutInvitation(userID, invite.Email, invite.OrganisationName, invitationToken, tx); err != nil {
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

	if !h.validateOrganisationName(ctx, invite.OrganisationName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := repository.PutMember(memberID, invite.OrganisationName, tx); err != nil {
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
	organisationName := ctx.Param("organisation_name")
	memberSearch := ctx.Param("member_search")

	if len(memberSearch) < 3 {
		h.abortWithError(ctx, http.StatusBadRequest, "Search query must be at least 3 characters long", nil)
		return
	}

	if !h.validateOrganisationName(ctx, organisationName) {
		return
	}
	userHasRights := false
	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if userHasRights = repository.CheckIsMemberOfOrganisation(userID, organisationName, tx); !userHasRights {
			return http.StatusForbidden, fmt.Errorf("insufficient rights userID: %s organisation: %s", userID, organisationName)
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
		UserID          int    `json:"user_id"`
		InvitationToken string `json:"invitation_token"`
	}

	var acceptRequest AcceptInvitationRequest
	if err := ctx.BindJSON(&acceptRequest); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	if userID != strconv.Itoa(acceptRequest.UserID) {
		h.abortWithError(ctx, http.StatusForbidden, "not authorized", fmt.Errorf("not authorized to accept inviation"))
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if ok, err := isInvitedToOrganisation(acceptRequest.InvitationToken, tx, ctx); err != nil || !ok {
			if err != nil {
				return http.StatusForbidden, err
			}
			slog.Info("Error", "err", "user does not have rights")
			return http.StatusForbidden, fmt.Errorf("insufficient rights")
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

func isInvitedToOrganisation(token string, tx *sqlx.Tx, ctx *gin.Context) (bool, error) {
	loggedInEmail := userMailFromSession(ctx)
	if loggedInEmail == "" {
		slog.Info("email is empty - check login scopes for any social login")
		return false, fmt.Errorf("can't verify email address")
	}

	invitation, err := repository.GetInvitation(loggedInEmail, token, tx)
	if err != nil {
		return false, err
	}

	if invitation != nil {
		return true, nil
	}

	return false, fmt.Errorf("user is not invited to the organisation")
}

func (h *Handler) HandleGETLeaveOrganisation(ctx *gin.Context) {
	// TODO
}

func (h *Handler) HandleGetOrganisationProjects(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisationName := ctx.Param("organisation_name")

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		projects, err := repository.GetProjectsByOrganisationName(userID, organisationName, tx)
		if err != nil {
			slog.Error("error", "cant get projects", err)
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, projects)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandlePUTOrganisationProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	type OrganisationProjectPut struct {
		UPN              string `json:"upn"`
		OrganisationName string `json:"organisation_name"`
	}
	var g OrganisationProjectPut

	if err := ctx.BindJSON(&g); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		ok, err := repository.AddOrganisationProjectByUPN(userID, g.OrganisationName, g.UPN, tx)
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

func (h *Handler) HandleDELETEOrganisationProject(ctx *gin.Context) {
	slog.Info("METHOD NOT IMPLEMENTED", "NOT IMPLEMENTED", "DELETE GROUP PROJECT")
	// TODO
}
