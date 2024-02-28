package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/devs-group/sloth/backend/repository"
	"github.com/gin-gonic/gin"
)

func (h *Handler) HandlePOSTGroup(ctx *gin.Context) {
	var g repository.Group
	userID := userIDFromSession(ctx)
	g.OwnerID = userID

	if err := ctx.BindJSON(&g); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to parse request body", err)
		return
	}

	if strings.TrimSpace(g.Name) == "" {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", fmt.Errorf("malformed request"))
		return
	}

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	if err := g.CreateGroup(tx); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to create group", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to store group", err)
	}

	ctx.JSON(http.StatusOK, g)
}

func (h *Handler) HandleGETGroups(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	groups, err := repository.SelectGroups(userID, tx)
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to select projects", err)
		return
	}

	ctx.JSON(http.StatusOK, groups)
}

func (h *Handler) HandleDELETEGroup(ctx *gin.Context) {
	var g repository.Group
	userID := userIDFromSession(ctx)

	g.Name = ctx.Param("group_name")
	g.OwnerID = userID

	if strings.TrimSpace(g.Name) == "" {
		h.abortWithError(ctx, http.StatusBadRequest, "unable to parse request body", fmt.Errorf("malformed request"))
		return
	}

	tx, err := h.store.DB.Beginx()
	defer tx.Rollback()

	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	if err := g.DeleteGroup(tx); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to delete group", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to delete group", err)
		return
	}

	ctx.Status(http.StatusOK)
}
