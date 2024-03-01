package config

import (
	"log/slog"
	"os"
	"strconv"

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

var DBPath = "./database/database.sqlite"
var DBMigrationsPath = "./database/migrations/"
var DBRunMigrations = true

const PersistentVolumeDirectoryName = "data"
const DockerComposeFileName = "docker-compose.yml"
const DockerConfigFileName = "config.json"

var SMTPFrom string
var SMTPPort string
var SMTPHost string
var SMTPPW string
var EmailInvitationURL string

func ReadBoolFromString(b string) bool {
	c, err := strconv.ParseBool(b)
	if err != nil {
		return false
	}
	return c
}

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

	SMTPFrom = os.Getenv("SMTP_FROM")
	SMTPHost = os.Getenv("SMTP_HOST")
	SMTPPort = os.Getenv("SMTP_PORT")
	SMTPPW = os.Getenv("SMTP_PW")

	EmailInvitationURL = os.Getenv("EMAIL_INVITATION_URL")
	if val := os.Getenv("DATABASE_PATH"); val != "" {
		DBPath = val
	}

	if val := os.Getenv("DATABASE_MIGRATIONS_PATH"); val != "" {
		DBMigrationsPath = val
	}

	if val := os.Getenv("DATABASE_RUN_MIGRATIONS"); val != "" {
		DBRunMigrations = ReadBoolFromString(val)
	}

	slog.Info("config from .env has been loaded")
}
