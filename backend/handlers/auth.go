package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/devs-group/sloth/backend/config"
	authprovider "github.com/devs-group/sloth/backend/handlers/auth-provider"
	"github.com/jmoiron/sqlx"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

const UserSessionKey = "user"

type AuthProvider interface {
	SetRequest(req *http.Request) error
	HandleGETAuthenticate(c *gin.Context) error
	HandleGETAuthenticateCallback(tx *sqlx.Tx, c *gin.Context) (int, error)
	HandleLogout(c *gin.Context) error
}

var providers = map[string]AuthProvider{
	"github": &authprovider.GitHubProvider{
		URL: "https://github.com/login/oauth/authorize",
	},
}

func assignProvider(c *gin.Context) *AuthProvider {
	providerKey := c.Param("provider")
	if provider, ok := providers[providerKey]; ok {
		q := c.Request.URL.Query()
		q.Set("provider", providerKey)
		c.Request.URL.RawQuery = q.Encode()
		provider.SetRequest(c.Request)
		return &provider
	}
	return nil
}

func enableCors(w gin.ResponseWriter) {
	(w).Header().Set("Access-Control-Allow-Origin", config.FrontendHost)
	(w).Header().Set("Access-Control-Allow-Credentials", "true")
}

func (h *Handler) HandleGETAuthenticate(c *gin.Context) {
	enableCors(c.Writer)
	p := assignProvider(c)
	if p != nil {
		(*p).HandleGETAuthenticate(c)
	}
}

func (h *Handler) HandleGETAuthenticateCallback(c *gin.Context) {
	enableCors(c.Writer)
	p := assignProvider(c)
	if p != nil {
		h.WithTransaction(c, func(tx *sqlx.Tx) (int, error) {
			res, err := (*p).HandleGETAuthenticateCallback(tx, c)
			slog.Info("Handled Authentication", "inf", res)
			slog.Info("Handled Authentication", "err", err)

			return res, err
		})
	} else {
		slog.Info("Unable to find auth-provider")
	}
}

func (h *Handler) HandleGETLogout(c *gin.Context) {
	p := assignProvider(c)
	if p != nil {
		(*p).HandleLogout(c)
	}
}

func (h *Handler) HandleGETUser(c *gin.Context) {
	enableCors(c.Writer)
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
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
}

func getUserFromSession(req *http.Request) (*goth.User, error) {
	var u goth.User
	s, err := gothic.GetFromSession("auth", req)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(s), &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := getUserFromSession(c.Request)
		if err != nil {
			slog.Error("unable to get user from session", "err", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(UserSessionKey, u.UserID)
		c.Next()
	}
}

func userIDFromSession(c *gin.Context) string {
	userID, _ := c.Get("user")
	return userID.(string)
}

func userMailFromSession(c *gin.Context) string {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		return ""
	}
	return u.Email
}
