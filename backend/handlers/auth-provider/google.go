package authprovider

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/devs-group/sloth/backend/repository"
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
		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"email":      u.Email,
				"name":       u.Name,
				"id":         u.UserID,
				"first_name": u.FirstName,
				"last_name":  u.LastName,
				"nickname":   u.NickName,
				"location":   u.Location,
				"avatar_url": u.AvatarURL,
			},
		})
	} else {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
	return nil
}

func (p *GoogleProvider) HandleGETAuthenticateCallback(tx *sqlx.Tx, c *gin.Context) (int, error) {
	u, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		slog.Error("unable to obtain user data - google", "provider", c.Param("provider"), "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return http.StatusBadGateway, err
	}

	err = storeUserInSession(&u, c.Request, c.Writer)
	if err != nil {
		slog.Error("unable to store user data in session", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return http.StatusBadGateway, err
	}

	if ok, err := repository.UpsertUserBySocialIDAndMethod(&u, tx); err != nil || !ok {
		if err != nil {
			slog.Error("error occurred during user upsert", err)
			return http.StatusBadGateway, err
		}
		if !ok {
			slog.Error("can't insert new user")
			return http.StatusBadGateway, fmt.Errorf("cant insert new user")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"email":      u.Email,
			"name":       u.Name,
			"id":         u.UserID,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"nickname":   u.NickName,
			"location":   u.Location,
			"avatar_url": u.AvatarURL,
		},
	})

	return http.StatusOK, nil
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
