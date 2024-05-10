package handlers

import (
	"fmt"
	"net/http"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/devs-group/sloth/backend/repository"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (h *Handler) HandleStreamServiceLogs(c *gin.Context) {
	userID := userIDFromSession(c)
	upn := repository.UPN(c.Param("upn"))
	s := c.Param("service")

	p := repository.Project{
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
	tx.Rollback()

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
