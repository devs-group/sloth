package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/devs-group/sloth/backend/config"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

const UserSessionKey = "user"

func assignProvider(c *gin.Context) *http.Request {
	q := c.Request.URL.Query()
	q.Add(":provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()
	return c.Request
}

func enableCors(w gin.ResponseWriter) {
	(w).Header().Set("Access-Control-Allow-Origin", config.FrontendHost)
	(w).Header().Set("Access-Control-Allow-Credentials", "true")
}

func (h *Handler) HandleGETAuthenticate(c *gin.Context) {
	// try to get the user without re-authenticating
	enableCors(c.Writer)
	c.Request = assignProvider(c)
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
}

func (h *Handler) HandleGETAuthenticateCallback(c *gin.Context) {
	enableCors(c.Writer)
	c.Request = assignProvider(c)
	u, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		slog.Error("unable to obtain user data", "provider", c.Param("provider"), "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = storeUserInSession(&u, c.Request, c.Writer)
	if err != nil {
		slog.Error("unable to store user data in session", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
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

func (h *Handler) HandleGETLogout(c *gin.Context) {
	c.Request = assignProvider(c)
	enableCors(c.Writer)
	err := gothic.Logout(c.Writer, c.Request)
	if err != nil {
		slog.Error("unable to logout user", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
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

func storeUserInSession(u *goth.User, req *http.Request, res http.ResponseWriter) error {
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return gothic.StoreInSession("auth", string(b), req, res)
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
