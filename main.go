package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/urfave/cli/v2"

	"github.com/devs-group/sloth/config"
	"github.com/devs-group/sloth/database"
	"github.com/devs-group/sloth/handlers"
)

//go:embed frontend/.output/public/*
var VueFiles embed.FS

func main() {
	config.LoadConfig()

	var port int
	app := &cli.App{
		Version:              config.Version,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Executes the application",
				Action: func(ctx *cli.Context) error {
					return run(port)
				},
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Value:       8080,
						Usage:       "Port at which the application should run on",
						Destination: &port,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(port int) error {
	slog.Info(fmt.Sprintf("Starting sloth in %s mode", config.Environment))

	r := gin.Default()
	s := database.NewStore()
	h := handlers.NewHandler(s, VueFiles)

	cookieStore := cookie.NewStore([]byte(config.SessionSecret))
	cookieStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
	})

	r.Use(sessions.Sessions("auth", cookieStore))
	gothic.Store = cookieStore

	goth.UseProviders(github.New(config.GithubClientKey, config.GithubSecret, config.GithubAuthCallbackURL))

	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = append(cfg.AllowOrigins, config.FrontendHost)
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

	slog.Info("Starting server", "frontend", fmt.Sprintf("%s/_/", config.FrontendHost))

	slog.Info("Port", "p", port)
	return r.Run(fmt.Sprintf(":%d", port))
}
