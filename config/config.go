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

var ENVIRONMENT Env = "production"
var GITHUB_CLIENT_KEY string
var GITHUB_SECRET string
var GITHUB_AUTH_CALLBACK_URL string
var SESSION_SECRET string
var HOST string

// LoadConfig loads config from .env file on development. Otherwise we rely on build flags.
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("unable to load config from .env file")
		return
	}

	ENVIRONMENT = Env(os.Getenv("ENVIRONMENT"))
	GITHUB_CLIENT_KEY = os.Getenv("GITHUB_CLIENT_KEY")
	GITHUB_SECRET = os.Getenv("GITHUB_SECRET")
	GITHUB_AUTH_CALLBACK_URL = os.Getenv("GITHUB_AUTH_CALLBACK_URL")
	SESSION_SECRET = os.Getenv("SESSION_SECRET")
	HOST = os.Getenv("HOST")
}
