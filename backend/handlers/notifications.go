package handlers

import (
	"net/http"

	"github.com/devs-group/sloth/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func (h *Handler) HandlePUTNotification(ctx *gin.Context) {
	userId := userIDFromSession(ctx)

	var notification services.Notification

	if err := ctx.BindJSON(&notification); err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to parse request body", err)
		return
	}

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		if err := h.service.StoreNotification(userId, notification.Subject, notification.Content, notification.Recipient, notification.NotificationType, tx); err != nil {
			return http.StatusForbidden, err
		}

		ctx.Status(http.StatusOK)
		return http.StatusOK, nil
	})

}

func (h *Handler) HandleGETNotifications(ctx *gin.Context) {
	userID := userIDFromSession(ctx)

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		notifications, err := h.service.GetNotifications(userID, tx)
		if err != nil {
			return http.StatusForbidden, err
		}

		ctx.JSON(http.StatusOK, notifications)
		return http.StatusOK, nil

	})
}
