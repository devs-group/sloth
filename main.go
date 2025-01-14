package main

import (
	"fmt"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/joho/godotenv"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
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

func main() {
	if !utils.IsProduction() {
		// During development, we load from .env file
		slog.Info("Loading environemnt variables from .env file")
		err := godotenv.Load(".env")
		if err != nil {
			slog.Error("Error loading .env file", "err", err)
			os.Exit(1)
		}
	} else {
		slog.Info("Loading environment variables from system")
	}

	cfg := config.GetConfig()

	var port int
	app := &cli.App{
		Version:              cfg.Version,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Executes the application",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Value:       9090,
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
	cfg := config.GetConfig()

	if utils.IsProduction() {
		slog.Info(fmt.Sprintf(`Starting sloth in "production" mode`))
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})))
	} else {
		slog.Info(fmt.Sprintf(`Starting sloth in "development" mode`))
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})))
	}

	r := gin.Default()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		// Suppress listing all available routes for less log spamming
	}
	dbService := database.NewDatabaseService(cfg.DBPath, cfg.DBMigrationsPath)
	err := dbService.Setup(false)
	if err != nil {
		log.Fatal("Failed to setup database", err)
	}

	h := handlers.New(dbService, VueFiles)

	cookieStore := cookie.NewStore([]byte(cfg.SessionSecret))
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
			cfg.GitHubConfig.GithubClientKey,
			cfg.GitHubConfig.GithubSecret,
			cfg.GitHubConfig.GithubAuthCallbackURL,
			"user:email",
		),
		google.New(
			cfg.GoogleConfig.GoogleClientKey,
			cfg.GoogleConfig.GoogleSecret,
			cfg.GoogleConfig.GoogleAuthCallbackURL,
		),
	)

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = append(corsCfg.AllowOrigins, cfg.FrontendHost)
	corsCfg.AllowCredentials = true
	corsCfg.AllowHeaders = append(corsCfg.AllowHeaders, "X-Access-Token")

	r.Use(cors.New(corsCfg))
	r.Use(gin.Recovery())

	v1 := r.Group("v1")
	// Add Backend Endpoints to Application
	h.RegisterEndpoints(v1)

	// Serve frontend
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/_/")
	})

	if !utils.IsProduction() {
		targetURL, err := url.Parse(cfg.FrontendHost)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		r.GET("/_/*proxyPath", func(c *gin.Context) {
			// Update the request's URL to the target URL
			c.Request.URL.Scheme = targetURL.Scheme
			c.Request.URL.Host = targetURL.Host

			// Update the Host header for the target server
			c.Request.Host = targetURL.Host

			// Forward the request to the target server
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	} else {
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
	}

	slog.Info("Starting server", "frontend", fmt.Sprintf("%s/_/", cfg.FrontendHost))
	slog.Info("Port", "p", port)

	return r.Run(fmt.Sprintf(":%d", port))
}
