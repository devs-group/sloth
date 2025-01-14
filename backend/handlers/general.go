package handlers

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"

	"github.com/gin-gonic/gin"

	"github.com/devs-group/sloth/backend/database"
	"github.com/devs-group/sloth/backend/services"
)

const accessTokenLen = 12
const uniqueProjectSuffixLen = 10

type Handler struct {
	dbService database.IDatabaseService
	vueFiles  embed.FS
	upgrader  websocket.Upgrader
	service   *services.S
}

type TransactionFunc func(*sqlx.Tx) (int, error)

func New(dbService database.IDatabaseService, vueFiles embed.FS) Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	// TODO: Loop over list of trusted origins instead returning true for all origins.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return Handler{
		dbService: dbService,
		vueFiles:  vueFiles,
		upgrader:  upgrader,
		service:   services.New(dbService),
	}
}

func (h *Handler) abortWithError(c *gin.Context, statusCode int, message string, err error) {
	slog.Error(message, "err", err)
	c.AbortWithStatus(statusCode)
}

func (h *Handler) WithTransaction(ctx *gin.Context, fn TransactionFunc) {
	tx, err := h.dbService.GetConn().Beginx()
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				h.abortWithError(ctx, http.StatusInternalServerError, "unable to commit transaction", err)
			}
		}
	}()

	statusCode, err := fn(tx)
	if err != nil {
		h.abortWithError(ctx, statusCode, "operation failed", err)
	}
}
