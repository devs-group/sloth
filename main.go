package main

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/devs-group/sloth/config"
	"github.com/devs-group/sloth/database"
	"github.com/devs-group/sloth/handlers"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

	cookieStore := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	cookieStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
	})

	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	l, _ := c.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	fmt.Print(l)

	r.Use(sessions.Sessions("auth", cookieStore))
	gothic.Store = cookieStore

	goth.UseProviders(github.New(os.Getenv("GITHUB_CLIENT_KEY"), os.Getenv("GITHUB_SECRET"), os.Getenv("GITHUB_AUTH_CALLBACK_URL")))

	config := cors.DefaultConfig()
	config.AllowOrigins = append(config.AllowOrigins, "http://localhost:3000")
	config.AllowCredentials = true
	config.AllowHeaders = append(config.AllowHeaders, "X-Access-Token")

	r.Use(cors.New(config))
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
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		fileHandler := http.FileServer(http.FS(subFs))
		c.Request.URL.Path = path
		fileHandler.ServeHTTP(c.Writer, c.Request)
	})

	r.Run()
}
