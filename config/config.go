package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Env string

const (
	Production  Env = "production"
	Development Env = "development"
)

var Environment Env = "production"
var GithubClientKey string
var GithubSecret string
var GithubAuthCallbackURL string
var SessionSecret string
var Host string
var ProjectsDir string
var FrontendHost string
var Version = "latest"

// LoadConfig loads config from .env file on development. Otherwise, we rely on build flags.
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("unable to load config from .env file")
		slog.Info("current config",
			"host", Host,
			"projects_dir", ProjectsDir,
			"frontend_host", FrontendHost,
			"version", Version,
		)
		return
	}

	Environment = Env(os.Getenv("ENVIRONMENT"))
	GithubClientKey = os.Getenv("GITHUB_CLIENT_KEY")
	GithubSecret = os.Getenv("GITHUB_SECRET")
	GithubAuthCallbackURL = os.Getenv("GITHUB_AUTH_CALLBACK_URL")
	SessionSecret = os.Getenv("SESSION_SECRET")
	Host = os.Getenv("HOST")
	ProjectsDir = os.Getenv("PROJECTS_DIR")
	FrontendHost = os.Getenv("FRONTEND_HOST")

	slog.Info("config from .env has been loaded")
}
