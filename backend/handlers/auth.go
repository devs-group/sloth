package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/devs-group/sloth/backend/config"
	authprovider "github.com/devs-group/sloth/backend/handlers/auth-provider"
	"github.com/jmoiron/sqlx"

	"github.com/gin-gonic/gin"
)

const UserSessionKey = "user"

type AuthProvider interface {
	SetRequest(req *http.Request) error
	HandleGETAuthenticate(c *gin.Context) error
	HandleGETAuthenticateCallback(tx *sqlx.Tx, c *gin.Context) (int, error)
	HandleLogout(c *gin.Context) error
}

var providers = map[string]AuthProvider{
	"github": &authprovider.GitHubProvider{},
	"google": &authprovider.GoogleProvider{},
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

func userIDFromSession(c *gin.Context) string {
	userID, _ := c.Get(UserSessionKey)
	return fmt.Sprintf("%v", userID)
}

func userMailFromSession(c *gin.Context) string {
	u, err := authprovider.GetUserFromSession(c.Request)
	if err != nil {
		return ""
	}
	return u.GothUser.Email
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

// AuthMiddleware retrieves the user from the current Goth session storage.
// If a user exists with a matching social ID and provider combination, their user ID is fetched
// and assigned for the current session.
//
// Important: Avoid changing the user ID elsewhere as this could disrupt SQL relationships
// in tables such as `organizations` and `projects`.
// Instead, to modify user associations, update the user ID consistently across the entire session.
//
// Parameters:
//   - h: A Handler instance to handle the HTTP request.
//
// Returns:
//   - A Gin HandlerFunc that manages the request authentication.
func AuthMiddleware(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := authprovider.GetUserFromSession(c.Request)
		if err != nil {
			slog.Error("unable to get user from session", "err", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(UserSessionKey, u.BackendUserID)
		c.Next()
	}
}

func (h *Handler) HandleGETUser(c *gin.Context) {
	enableCors(c.Writer)
	u, err := authprovider.GetUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, authprovider.CreateUserResponse(u.GothUser))
}
