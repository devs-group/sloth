package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/models"
	"github.com/devs-group/sloth/backend/pkg/email"
	"github.com/devs-group/sloth/backend/services"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func (h *Handler) HandlePOSTOrganisation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisation := models.Organisation{
		OwnerID: userID,
	}

	if err := ctx.BindJSON(&organisation); err != nil {
		slog.Error("unable to parse organisation request body", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "unable to parse request body"})
		return
	}

	o, err := h.service.CreateOrganisation(organisation)
	if err != nil {
		slog.Error("unable to create organisation", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "unable to create organisation"})
		return
	}

	ctx.JSON(http.StatusOK, o)
	return
}

func (h *Handler) HandleGETOrganisations(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		organisations, err := services.SelectOrganisations(userID, tx)
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
		organisation, err := services.SelectOrganisation(tx, organisationID, userID)
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
		g := services.Organisation{
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
	organisationID := ctx.Param("id")
	memberID := ctx.Param("member_id")

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := services.DeleteMember(userID, memberID, organisationID, tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandlePUTInvitation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	var invite services.Invitation

	if err := ctx.BindJSON(&invite); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
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
		// Returns status forbidden if userId is not ownerID
		if err := services.PutInvitation(userID, invite.Email, invite.OrganisationID, invitationToken, tx); err != nil {
			return http.StatusForbidden, err
		}

		if err = services.StoreNotification(userID, "Invitation", "Invitation Content", invite.Email, "INVITATION", tx); err != nil {
			slog.Error("Unable to store notification from invitation")
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandlePUTMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	memberID := ctx.Param("member_id")

	var invite services.Invitation

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

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := services.PutMember(memberID, invite.OrganisationID, tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETInvitations(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisationID, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		ctx.Status(http.StatusBadRequest)
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		invites, err := services.GetInvitations(userID, organisationID, tx)
		if err != nil {
			return http.StatusForbidden, err
		}
		ctx.JSON(http.StatusOK, invites)
		return http.StatusOK, nil
	})
}

// func (h *Handler) HandleGETMembersForInvitation(ctx *gin.Context) {
// 	userID := userIDFromSession(ctx)
// 	organisationName := ctx.Param("organisation_name")
// 	memberSearch := ctx.Param("member_search")

// 	if len(memberSearch) < 3 {
// 		h.abortWithError(ctx, http.StatusBadRequest, "Search query must be at least 3 characters long", nil)
// 		return
// 	}

// 	if !h.validateOrganisationName(ctx, organisationName) {
// 		return
// 	}
// 	userHasRights := false
// 	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
// 		if userHasRights = services.CheckIsMemberOfOrganisation(userID, organisationName, tx); !userHasRights {
// 			return http.StatusForbidden, fmt.Errorf("insufficient rights userID: %s organisation: %s", userID, organisationName)
// 		}
// 		return http.StatusOK, nil
// 	})

// 	if !userHasRights {
// 		// if user does not have enough rights error was already send
// 		return
// 	}

// 	res, err := github.SearchGitHubUsers(memberSearch)
// 	if err != nil {
// 		h.abortWithError(ctx, http.StatusInternalServerError, "Github fetch failed", err)
// 	}

// 	ctx.JSON(http.StatusOK, res)
// }

func (h *Handler) HandleDELETEWithdrawInvitation(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	type WithdrawInvitation struct {
		Email          string `json:"email"`
		OrganisationID int    `json:"organisation_id"`
	}

	var withdrawInvitation WithdrawInvitation
	if err := ctx.BindJSON(&withdrawInvitation); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	slog.Info("+dsfs", withdrawInvitation)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		err := services.WithdrawInvitation(userID, withdrawInvitation.Email, withdrawInvitation.OrganisationID, tx)
		if err != nil {
			return http.StatusForbidden, err
		}
		return http.StatusOK, nil
	})

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

		ok, err := services.AcceptInvitation(userID, userMailFromSession(ctx), acceptRequest.InvitationToken, tx)
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

	invitation, err := services.GetInvitation(loggedInEmail, token, tx)
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

func (h *Handler) HandleGETOrganisationProjects(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	organisationID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Status(http.StatusBadRequest)
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		projects, err := services.GetProjectsByOrganisationID(userID, organisationID, tx)
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
		UPN            string `json:"upn"`
		OrganisationID int    `json:"organisation_id"`
	}
	var g OrganisationProjectPut

	if err := ctx.BindJSON(&g); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		ok, err := services.AddOrganisationProjectByUPN(userID, g.OrganisationID, g.UPN, tx)
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

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {

		fmt.Printf("userID '%s', organisationID '%d', upn '%s'", userID, g.OrganisationID, g.UPN)

		if err := services.DeleteProject(userID, g.OrganisationID, g.UPN, tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}
