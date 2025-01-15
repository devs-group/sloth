package config

import (
	"os"
	"strconv"
	"time"
)

type GitHubConfig struct {
	GithubClientKey       string
	GithubSecret          string
	GithubAuthCallbackURL string
}

type GoogleConfig struct {
	GoogleClientKey       string
	GoogleSecret          string
	GoogleAuthCallbackURL string
}

type ComposeLimitConfig struct {
	CPUs   string
	Memory string
}

type Config struct {
	SessionSecret string
	BackendHost   string
	BackendUrl    string
	FrontendHost  string
	ProjectsDir   string
	Version       string

	SMTPFrom     string
	SMTPHost     string
	SMTPPort     string
	SMTPPassword string

	EmailInvitationMaxValid time.Duration
	EmailInvitationURL      string

	DBPath           string
	DBMigrationsPath string

	GitHubConfig *GitHubConfig
	GoogleConfig *GoogleConfig

	DockerContainerLimits   *ComposeLimitConfig
	DockerContainerReplicas int

	// Statics
	PersistentVolumeDirectoryName string
	DockerComposeFileName         string
	DockerConfigFileName          string
}

func GetConfig() Config {
	return Config{
		SessionSecret: getEnv("SESSION_SECRET", ""),
		BackendHost:   getEnv("BACKEND_HOST", "localhost"),
		BackendUrl:    getEnv("BACKEND_URL", "http://localhost"),
		FrontendHost:  getEnv("FRONTEND_HOST", "http://frontend:3000"),
		ProjectsDir:   getEnv("PROJECTS_DIR", "./projects"),
		Version:       getEnv("VERSION", "latest"),

		SMTPFrom:     getEnv("SMTP_FROM", ""),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		EmailInvitationMaxValid: 7 * 24 * time.Hour,
		EmailInvitationURL:      getEnv("EMAIL_INVITATION_URL", ""),

		DBPath:           getEnv("DATABASE_PATH", "db.sqlite"),
		DBMigrationsPath: getEnv("DATABASE_MIGRATIONS_PATH", "migrations"),

		GitHubConfig: &GitHubConfig{
			GithubClientKey:       getEnv("GITHUB_CLIENT_KEY", ""),
			GithubSecret:          getEnv("GITHUB_SECRET", ""),
			GithubAuthCallbackURL: getEnv("GITHUB_AUTH_CALLBACK_URL", ""),
		},

		GoogleConfig: &GoogleConfig{
			GoogleClientKey:       getEnv("GOOGLE_CLIENT_KEY", ""),
			GoogleSecret:          getEnv("GOOGLE_SECRET", ""),
			GoogleAuthCallbackURL: getEnv("GOOGLE_AUTH_CALLBACK_URL", ""),
		},

		DockerContainerLimits: &ComposeLimitConfig{
			CPUs:   getEnv("DOCKER_CONTAINER_MAX_CPUS", "1.0"),
			Memory: getEnv("DOCKER_CONTAINER_MAX_MEMORY", "256M"),
		},
		DockerContainerReplicas: getEnvInt("DOCKER_CONTAINER_MAX_REPLICAS", 1),

		PersistentVolumeDirectoryName: "data",
		DockerComposeFileName:         "docker-compose.yml",
		DockerConfigFileName:          "config.json",
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if stringValue, exists := os.LookupEnv(key); exists {
		value, err := strconv.Atoi(stringValue)
		if err != nil {
			return fallback
		}
		return value
	}
	return fallback
}
