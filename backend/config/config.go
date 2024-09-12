package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/joho/godotenv"
)

type Env string

const (
	Production  Env = "production"
	Development Env = "development"
)

var Environment Env = "production"

var SessionSecret string
var Host string
var ProjectsDir string
var FrontendHost string
var Version = "latest"

var DBPath = "./database/database.sqlite"
var DBMigrationsPath = "./database/migrations/"

const PersistentVolumeDirectoryName = "data"
const DockerComposeFileName = "docker-compose.yml"
const DockerConfigFileName = "config.json"

var SMTPFrom string
var SMTPPort string
var SMTPHost string
var SMTPPassword string
var EmailInvitationURL string
var EmailInvitationMaxValid time.Duration

var DockerContainerLimits compose.Limits
var DockerContainerReplicas int

func initializeDependency() error {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("unable to load config from .env file")
		slog.Info("current config",
			"host", Host,
			"projects_dir", ProjectsDir,
			"frontend_host", FrontendHost,
			"version", Version,
		)
		return err
	}
	return nil
}

// LoadConfig loads config from .env file on development. Otherwise, we rely on build flags.
func LoadConfig() {
	if err := initializeDependency(); err != nil {
		return
	}

	Environment = Env(os.Getenv("ENVIRONMENT"))
	AuthProviderConfig = *NewAuthProvider()
	SessionSecret = os.Getenv("SESSION_SECRET")
	Host = os.Getenv("HOST")
	ProjectsDir = os.Getenv("PROJECTS_DIR")
	FrontendHost = os.Getenv("FRONTEND_HOST")

	SMTPFrom = os.Getenv("SMTP_FROM")
	SMTPHost = os.Getenv("SMTP_HOST")
	SMTPPort = os.Getenv("SMTP_PORT")
	SMTPPassword = os.Getenv("SMTP_PASSWORD")

	EmailInvitationMaxValid = 7 * 24 * time.Hour

	EmailInvitationURL = os.Getenv("EMAIL_INVITATION_URL")
	if val := os.Getenv("DATABASE_PATH"); val != "" {
		DBPath = val
	}

	if val := os.Getenv("DATABASE_MIGRATIONS_PATH"); val != "" {
		DBMigrationsPath = val
	}

	maxCpus := os.Getenv("DOCKER_CONTAINER_MAX_CPUS")
	maxMemory := os.Getenv("DOCKER_CONTAINER_MAX_MEMORY")
	DockerContainerLimits = compose.Limits{
		CPUs:   &maxCpus,
		Memory: &maxMemory,
	}

	var err error
	DockerContainerReplicas, err = strconv.Atoi(os.Getenv("DOCKER_CONTAINER_MAX_REPLICAS"))
	if err != nil {
		slog.Info("cant parse or find 'DOCKER_CONTAINER_MAX_REPLICAS'")
		panic(err)
	}

	slog.Info("config from .env has been loaded")
}
