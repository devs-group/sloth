package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/devs-group/sloth/backend/pkg/github"
	"github.com/devs-group/sloth/backend/repository"
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

	h.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		err := group.CreateGroup(tx)
		if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, group)
		return nil
	})
}

func (h *Handler) HandleGETGroups(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		groups, err := repository.SelectGroups(userID, tx)
		if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, groups)
		return nil
	})
}

func (h *Handler) HandleGETGroup(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	groupName := ctx.Param("group_name")

	h.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		g := repository.Group{
			OwnerID: userID,
			Name:    groupName,
		}

		if err := g.SelectGroup(tx); err != nil {
			return err
		}
		ctx.JSON(http.StatusOK, g)
		return nil
	})
}

func (h *Handler) HandleDELETEGroup(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	groupName := ctx.Param("group_name")

	if !h.validateGroupName(ctx, groupName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		g := repository.Group{
			Name:    groupName,
			OwnerID: userID,
		}

		if err := g.DeleteGroup(tx); err != nil {
			return err
		}

		ctx.Status(http.StatusOK)
		return nil
	})
}

func (h *Handler) HandleDELETEMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	groupName := ctx.Param("group_name")
	memberID := ctx.Param("member_id")

	if !h.validateGroupName(ctx, groupName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		if err := repository.DeleteMember(userID, memberID, groupName, tx); err != nil {
			return err
		}

		ctx.Status(http.StatusOK)
		return nil
	})
}

func (h *Handler) HandlePUTMember(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	groupName := ctx.Param("group_name")
	memberID := ctx.Param("member_id")

	if !h.validateGroupName(ctx, groupName) {
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		if err := repository.PutMember(userID, memberID, groupName, tx); err != nil {
			return err
		}

		ctx.Status(http.StatusOK)
		return nil
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
	h.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		if userHasRights = repository.CheckMemberOfGroup(userID, groupName, tx); !userHasRights {
			return fmt.Errorf("Insufficient rights userID: %s group: %s", userID, groupName)
		}
		return nil
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
