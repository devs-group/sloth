package authprovider

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

func CreateUserResponse(u *goth.User) gin.H {
	return gin.H{
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
	}
}

type UserSession struct {
	BackendUserID int        `json:"userID"`
	GothUser      *goth.User `json:"gothUser"`
}

func StoreUserInSession(backendUserID int, u *goth.User, req *http.Request, res http.ResponseWriter) error {
	session := UserSession{
		BackendUserID: backendUserID,
		GothUser:      u,
	}

	b, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return gothic.StoreInSession("auth", string(b), req, res)
}

func GetUserFromSession(req *http.Request) (*UserSession, error) {
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
