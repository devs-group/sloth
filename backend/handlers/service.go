package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/devs-group/sloth/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (h *Handler) HandleStreamServiceLogs(c *gin.Context) {
	cfg := config.GetConfig()

	userID := userIDFromSession(c)
	upn := services.UPN(c.Param("upn"))
	s := c.Param("usn")

	p := services.Project{
		UserID: userID,
		UPN:    upn,
		Path:   upn.GetProjectPath(),
		Hook:   fmt.Sprintf("%s/v1/hook/%s", cfg.BackendUrl, upn),
	}

	err := h.service.SelectProjectByUPNOrAccessToken(&p)
	if err != nil {
		h.abortWithError(c, http.StatusBadRequest, "unable to find project by upn", err)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to upgrade http to ws", err)
		return
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			h.abortWithError(c, http.StatusInternalServerError, "unable to close websocket connection", err)
			return
		}
	}(conn)

	pPath := upn.GetProjectPath()
	out := make(chan string)
	go func() {
		err := compose.Logs(pPath, s, out)
		if err != nil {
			msg := fmt.Sprintf("unable to stream logs for service %s", s)
			h.abortWithError(c, http.StatusInternalServerError, msg, err)
			return
		}
	}()

	line := 0
	for o := range out {
		line++
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d %s", line, o)))
	}
}

func (h *Handler) HandleStreamShell(ctx *gin.Context) {
	userID := userIDFromSession(ctx)
	usn := ctx.Param("usn")
	projectID, err := strconv.Atoi(ctx.Param("projectID"))
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		h.abortWithError(ctx, http.StatusInternalServerError, "unable to upgrade http to ws", err)
		return
	}

	out := make(chan []byte)
	in := make(chan []byte)

	p, err := h.service.SelectProjectByIDAndUserID(projectID, userID)
	if err != nil {
		h.abortWithError(ctx, http.StatusNotFound, "unable to find project", err)
		return
	}

	c, cancel := context.WithCancel(ctx)

	go func() {
		err := compose.Shell(c, p.Path, string(p.UPN), usn, in, out)
		if err != nil {
			slog.Error("unable to interact with the shell", "err", err)
		}
	}()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				slog.Info("error reading from websocket:", "err", err)
				cancel()
				break
			}
			in <- message
		}
	}()

	go func() {
		for o := range out {
			err = conn.WriteMessage(websocket.TextMessage, []byte(o))
			if err != nil {
				slog.Info("error writing to websocket:", "err", err)
				cancel()
				conn.Close()
			}
		}
	}()
}
