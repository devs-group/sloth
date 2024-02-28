package handlers

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"

	"github.com/devs-group/sloth/backend/database"
)

const accessTokenLen = 12
const uniqueProjectSuffixLen = 10

type Handler struct {
	store    *database.Store
	vueFiles embed.FS
	upgrader websocket.Upgrader
}

func New(store *database.Store, vueFiles embed.FS) Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	// TODO: Loop over list of trusted origins instead returning true for all origins.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return Handler{
		store:    store,
		vueFiles: vueFiles,
		upgrader: upgrader,
	}
}

func (h *Handler) abortWithError(c *gin.Context, statusCode int, message string, err error) {
	slog.Error(message, "err", err)
	c.AbortWithStatus(statusCode)
}
