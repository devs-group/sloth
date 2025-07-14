package authprovider

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/devs-group/sloth/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

func CreateUserResponse(u *UserSession) gin.H {
	return gin.H{
		"user": gin.H{
			"id":                    u.BackendUserID,
			"currentOrganisationID": u.CurrentOrganisationID,
			"email":                 u.GothUser.Email,
			"name":                  u.GothUser.Name,
			"first_name":            u.GothUser.FirstName,
			"last_name":             u.GothUser.LastName,
			"nickname":              u.GothUser.NickName,
			"location":              u.GothUser.Location,
			"avatar_url":            u.GothUser.AvatarURL,
		},
	}
}

type UserSession struct {
	BackendUserID         int        `json:"userID"`
	CurrentOrganisationID int        `json:"currentOrganisationID"`
	GothUser              *goth.User `json:"gothUser"`
}

func UpdateSession(provider string, u *goth.User, tx *sqlx.Tx, c *gin.Context) (int, error) {
	sessionIDs, err := services.UpsertUserBySocialIDAndMethod(provider, u, tx)
	if err != nil || sessionIDs == nil {
		if err != nil {
			slog.Error("error occurred during user upsert", "err", err)
			return http.StatusBadGateway, err
		}
		// User ID is 0
		slog.Error("can't insert new user")
		return http.StatusBadGateway, fmt.Errorf("cant insert new user")
	}

	session, err := StoreUserInSession(sessionIDs.UserID, sessionIDs.CurrentOrganisationID, u, c.Request, c.Writer)
	if err != nil {
		slog.Error("unable to store user data in session", "err", err)
		return http.StatusInternalServerError, err
	}

	c.JSON(http.StatusOK, CreateUserResponse(session))
	return http.StatusOK, nil
}

func StoreUserInSession(backendUserID, currentOrganisationID int, u *goth.User, req *http.Request, res http.ResponseWriter) (*UserSession, error) {
	session := UserSession{
		BackendUserID:         backendUserID,
		CurrentOrganisationID: currentOrganisationID,
		GothUser:              u,
	}

	b, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	if err := gothic.StoreInSession("auth", string(b), req, res); err != nil {
		return nil, err
	}

	return &session, nil
}

func GetUserSession(req *http.Request) (*UserSession, error) {
	s, err := gothic.GetFromSession("auth", req)
	if err != nil {
		return nil, err
	}

	var session UserSession
	err = json.Unmarshal([]byte(s), &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
