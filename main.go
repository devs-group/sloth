package main

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/devs-group/sloth/config"
	"github.com/devs-group/sloth/database"
	"github.com/devs-group/sloth/handlers"
	"io/fs"
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

//go:embed frontend/.output/public/*
var VueFiles embed.FS

func main() {
	config.LoadConfig()

	slog.Info(fmt.Sprintf("Starting sloth in %s mode", config.ENVIRONMENT))

	r := gin.Default()
	s := database.NewStore()
	h := handlers.NewHandler(s, VueFiles)

	cookieStore := cookie.NewStore([]byte(config.SESSION_SECRET))
	cookieStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
	})

	r.Use(sessions.Sessions("auth", cookieStore))
	gothic.Store = cookieStore

	goth.UseProviders(github.New(config.GITHUB_CLIENT_KEY, config.GITHUB_SECRET, config.GITHUB_AUTH_CALLBACK_URL))

	cfg := cors.DefaultConfig()
	if config.ENVIRONMENT == config.Development {
		cfg.AllowOrigins = append(cfg.AllowOrigins, "http://localhost:3000")
	}
	cfg.AllowCredentials = true
	cfg.AllowHeaders = append(cfg.AllowHeaders, "X-Access-Token")

	r.Use(cors.New(cfg))
	r.Use(gin.Recovery())

	r.GET("/info", h.HandleGETInfo)
	r.POST("v1/project", h.HandlePOSTProject)
	r.GET("v1/projects", h.HandleGETProjects)
	r.GET("v1/hook/:unique_project_name", h.HandleGETHook)
	r.GET("v1/auth/:provider", h.HandleGETAuthenticate)
	r.GET("v1/auth/:provider/callback", h.HandleGETAuthenticateCallback)
	r.GET("v1/auth/logout/:provider", h.HandleGETLogout)
	r.GET("v1/auth/user", h.HandleGETUser)

	// Serve frontend
	r.GET("/_/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")
		subFs, err := fs.Sub(VueFiles, "frontend/.output/public")
		if err != nil {
			slog.Error("unable to get subtree of frontend files")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		fileHandler := http.FileServer(http.FS(subFs))
		c.Request.URL.Path = path
		fileHandler.ServeHTTP(c.Writer, c.Request)
	})

	err := r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
