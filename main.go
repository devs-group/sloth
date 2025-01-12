package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/urfave/cli/v2"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/database"
	"github.com/devs-group/sloth/backend/handlers"
)

//go:embed all:frontend/.output/public/*
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
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Value:       8080,
						Usage:       "Port at which the application should run on",
						Destination: &port,
					},
				},
				Action: func(ctx *cli.Context) error {
					return run(port)
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

	if config.Environment == config.Development {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})))
	} else {
		gin.SetMode(gin.ReleaseMode)
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})))
	}

	r := gin.Default()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		// Suppress listing all available routes for less log spamming
	}
	s := database.NewStore()
	h := handlers.New(s, VueFiles)

	cookieStore := cookie.NewStore([]byte(config.SessionSecret))
	cookieStore.Options(sessions.Options{
		Path: "/",
		// 7 Days validity
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
	})

	r.Use(sessions.Sessions("auth", cookieStore))

	gothic.Store = cookieStore

	goth.UseProviders(
		github.New(
			config.AuthProviderConfig.GitHubConfig.GithubClientKey,
			config.AuthProviderConfig.GitHubConfig.GithubSecret,
			config.AuthProviderConfig.GitHubConfig.GithubAuthCallbackURL,
			"user:email",
		),
		google.New(
			config.AuthProviderConfig.GoogleConfig.GoogleClientKey,
			config.AuthProviderConfig.GoogleConfig.GoogleSecret,
			config.AuthProviderConfig.GoogleConfig.GoogleAuthCallbackURL,
		),
	)

	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = append(cfg.AllowOrigins, config.FrontendHost)
	cfg.AllowCredentials = true
	cfg.AllowHeaders = append(cfg.AllowHeaders, "X-Access-Token")

	r.Use(cors.New(cfg))
	r.Use(gin.Recovery())

	v1 := r.Group("v1")
	// Add Backend Endpoints to Application
	h.RegisterEndpoints(v1)

	// Serve frontend
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/_/")
	})

	r.GET("/_/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")
		subFs, err := fs.Sub(VueFiles, "frontend/.output/public")
		if err != nil {
			slog.Error("unable to get subtree of frontend files")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		fileHandler := http.FileServer(http.FS(subFs))
		if strings.HasPrefix(path, "/_nuxt") {
			c.Request.URL.Path = path
		} else {
			c.Request.URL.Path = "/"
		}
		fileHandler.ServeHTTP(c.Writer, c.Request)
	})

	slog.Info("Starting server", "frontend", fmt.Sprintf("%s/_/", config.FrontendHost))
	slog.Info("Port", "p", port)

	return r.Run(fmt.Sprintf(":%d", port))
}
