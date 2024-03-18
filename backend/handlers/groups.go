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

// Validate the group name is not empty or just whitespace.
func (h *Handler) validateGroupName(ctx *gin.Context, groupName string) bool {
	if strings.TrimSpace(groupName) == "" {
		h.abortWithError(ctx, http.StatusBadRequest, "Group name cannot be empty or whitespace", fmt.Errorf("malformed request"))
		return false
	}
	return true
}

func (h *Handler) HandlePOSTGroup(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	group := repository.Group{
		OwnerID: userID,
	}

	if err := ctx.BindJSON(&group); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	if !h.validateGroupName(ctx, group.Name) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		err := group.CreateGroup(tx)
		if err != nil {
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, group)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETGroups(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		groups, err := repository.SelectGroups(userID, tx)
		if err != nil {

			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, groups)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleGETGroup(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	groupName := ctx.Param("group_name")

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		g := repository.Group{
			OwnerID: userID,
			Name:    groupName,
		}

		if err := g.SelectGroup(tx); err != nil {
			return http.StatusForbidden, err
		}
		ctx.JSON(http.StatusOK, g)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEGroup(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	groupName := ctx.Param("group_name")

	if !h.validateGroupName(ctx, groupName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		g := repository.Group{
			Name:    groupName,
			OwnerID: userID,
		}

		if err := g.DeleteGroup(tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandleDELETEMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	groupName := ctx.Param("group_name")
	memberID := ctx.Param("member_id")

	if !h.validateGroupName(ctx, groupName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := repository.DeleteMember(userID, memberID, groupName, tx); err != nil {
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
	if !h.validateGroupName(ctx, invite.GroupName) {
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
		if err := repository.PutInvitation(userID, invite.Email, invite.GroupName, invitationToken, tx); err != nil {
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

	if !h.validateGroupName(ctx, invite.GroupName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := repository.PutMember(memberID, invite.GroupName, tx); err != nil {
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
	groupName := ctx.Param("group_name")
	memberSearch := ctx.Param("member_search")

	if len(memberSearch) < 3 {
		h.abortWithError(ctx, http.StatusBadRequest, "Search query must be at least 3 characters long", nil)
		return
	}

	if !h.validateGroupName(ctx, groupName) {
		return
	}
	userHasRights := false
	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if userHasRights = repository.CheckIsMemberOfGroup(userID, groupName, tx); !userHasRights {
			return http.StatusForbidden, fmt.Errorf("Insufficient rights userID: %s group: %s", userID, groupName)
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
		if !isInvitatedToGroup(acceptRequest.InvitationToken, tx, ctx) {
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

func isInvitatedToGroup(token string, tx *sqlx.Tx, ctx *gin.Context) bool {
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

func (h *Handler) HandleGETLeaveGroup(ctx *gin.Context) {
	// TODO
}

func (h *Handler) HandleGetGroupProjects(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	groupName := ctx.Param("group_name")

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		projects, err := repository.GetProjectsByGroupName(userID, groupName, tx)
		if err != nil {
			slog.Error("error", "cant get projects", err)
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, projects)
		return http.StatusOK, nil
	})
}

func (h *Handler) HandlePUTGroupProject(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	type GroupProjectPut struct {
		UPN       string `json:"upn"`
		GroupName string `json:"group_name"`
	}
	var g GroupProjectPut

	if err := ctx.BindJSON(&g); err != nil {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", err)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		ok, err := repository.AddGroupProjectByUPN(userID, g.GroupName, g.UPN, tx)
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

func (h *Handler) HandleDELETEGroupProject(ctx *gin.Context) {
	// TODO
}
