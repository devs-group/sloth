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
	"github.com/jmoiron/sqlx"
)

func (h *Handler) HandleStreamServiceLogs(c *gin.Context) {
	userID := userIDFromSession(c)
	upn := services.UPN(c.Param("upn"))
	s := c.Param("usn")

	p := services.Project{
		UserID: userID,
		UPN:    upn,
		Path:   upn.GetProjectPath(),
		Hook:   fmt.Sprintf("%s/v1/hook/%s", config.Host, upn),
	}

	tx, err := h.store.DB.Beginx()
	if err != nil {
		h.abortWithError(c, http.StatusInternalServerError, "unable to initiate transaction", err)
		return
	}

	err = p.SelectProjectByUPNOrAccessToken(tx)
	if err != nil {
		h.abortWithError(c, http.StatusBadRequest, "unable to find project by upn", err)
		return
	}
	// TODO: @4ddev why is this rolled back here?
	tx.Commit()

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
	serviceID := ctx.Param("service")
	projectID, err := strconv.Atoi(ctx.Param("id"))
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

	h.WithTransaction(ctx, func(tx *sqlx.Tx) (int, error) {
		p, err := services.SelectProjectByIDAndUserID(tx, projectID, userID)
		if err != nil {
			return http.StatusNotFound, err
		}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			err := compose.Shell(ctx, p.Path, string(p.UPN), serviceID, in, out)
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
		return http.StatusOK, nil
	})
}
