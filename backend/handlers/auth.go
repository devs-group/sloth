package handlers

import (
	"fmt"
	"github.com/devs-group/sloth/backend/config"
	authprovider "github.com/devs-group/sloth/backend/handlers/auth-provider"
	"github.com/devs-group/sloth/backend/models"
	"github.com/jmoiron/sqlx"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const UserSessionKey = "user"
const UserCurrentOrganisationIDKey = "currentOrganisationID"

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
		err := provider.SetRequest(c.Request)
		if err != nil {
			return nil
		}
		return &provider
	}
	return nil
}

func enableCors(w gin.ResponseWriter) {
	cfg := config.GetConfig()
	(w).Header().Set("Access-Control-Allow-Origin", cfg.FrontendHost)
	(w).Header().Set("Access-Control-Allow-Credentials", "true")
}

func userIDFromSession(c *gin.Context) string {
	userID, _ := c.Get(UserSessionKey)
	return fmt.Sprintf("%v", userID)
}

func currentOrganisationIDFromSession(c *gin.Context) string {
	userID, _ := c.Get(UserCurrentOrganisationIDKey)
	return fmt.Sprintf("%v", userID)
}

func getUserMailFromSession(c *gin.Context) (string, error) {
	u, err := authprovider.GetUserSession(c.Request)
	if err != nil {
		return "", fmt.Errorf("unable to obtain user email from the session")
	}
	return u.GothUser.Email, nil
}

func (h *Handler) HandleGETAuthenticate(c *gin.Context) {
	enableCors(c.Writer)
	p := assignProvider(c)
	if p != nil {
		err := (*p).HandleGETAuthenticate(c)
		if err != nil {
			slog.Error("HandleGETAuthenticate error", "err", err)
		}
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
		err := (*p).HandleLogout(c)
		if err != nil {
			slog.Error("HandleGETLogout error", "err", err)
		}
	}
}

// AuthMiddleware retrieves the user from the current Goth session storage.
// If a user exists with a matching social ID and provider combination, their user ID is fetched
// and assigned for the current session.
//
// Important: Avoid changing the user ID elsewhere as this could disrupt SQL relationships
// in tables such as `organisations` and `projects`.
// Instead, to modify user associations, update the user ID consistently across the entire session.
//
// Parameters:
//   - h: A Handler instance to handle the HTTP request.
//
// Returns:
//   - A Gin HandlerFunc that manages the request authentication.
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		u, err := authprovider.GetUserSession(ctx.Request)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Additionally check that user exists in our database
		query := `
			SELECT current_organisation_id
			FROM users
			WHERE user_id = ?
		`
		var currentOrganisationID int
		err = h.dbService.GetConn().Get(&currentOrganisationID, query, u.BackendUserID)
		if err != nil || currentOrganisationID == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		slog.Debug("VerifyUserSession", "currentOrganisationID", currentOrganisationID)

		ctx.Set(UserSessionKey, u.BackendUserID)
		ctx.Set(UserCurrentOrganisationIDKey, currentOrganisationID)
		ctx.Next()
	}
}

func (h *Handler) GetUser(c *gin.Context) {
	enableCors(c.Writer)
	u, err := authprovider.GetUserSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, authprovider.CreateUserResponse(u))
}

func (h *Handler) SetCurrentOrganisation(c *gin.Context) {
	userID := userIDFromSession(c)
	u, err := authprovider.GetUserSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var payload models.UserSetCurrentOrganisation
	if err := c.BindJSON(&payload); err != nil {
		UnableToParseRequestBody(c, err)
		return
	}
	query := `
			UPDATE users
			SET current_organisation_id = $1
			WHERE user_id = $2
		`
	slog.Debug("SetCurrentOrganisation", "payload", payload.ID, "userID", userID)
	_, err = h.dbService.GetConn().Exec(query, payload.ID, userID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	userIDAsInt, err := strconv.Atoi(userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	_, err = authprovider.StoreUserInSession(userIDAsInt, payload.ID, u.GothUser, c.Request, c.Writer)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

// VerifyUserSession simply responds OK in case the AuthMiddleware protecting is not preventing that
func (h *Handler) VerifyUserSession(c *gin.Context) {
	c.Status(http.StatusOK)
}
