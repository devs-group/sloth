package authprovider

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth/gothic"
)

type GoogleProvider struct {
	URL     string
	Request *http.Request
}

func (p *GoogleProvider) SetRequest(req *http.Request) error {
	p.Request = req
	return nil
}

func (p *GoogleProvider) HandleGETAuthenticate(c *gin.Context) error {
	u, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err == nil {
		session, err := GetUserSession(c.Request)
		if err != nil {
			return nil
		}
		if u.UserID == session.GothUser.UserID && u.Provider == session.GothUser.Provider {
			c.JSON(http.StatusOK, CreateUserResponse(session))
		}
	} else {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
	return nil
}

func (p *GoogleProvider) HandleGETAuthenticateCallback(tx *sqlx.Tx, c *gin.Context) (int, error) {
	u, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		slog.Error("unable to obtain user data - google", "provider", c.Param("provider"), "err", err)
		return http.StatusUnauthorized, err
	}
	return UpdateSession(&u, tx, c)
}

func (p *GoogleProvider) HandleLogout(c *gin.Context) error {
	err := gothic.Logout(c.Writer, c.Request)
	if err != nil {
		slog.Error("unable to logout user", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
	return nil
}
