package main_tests

import (
	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	err := godotenv.Load(".env.test")
	if err != nil {
		slog.Error("Error loading .env file", "err", err)
		os.Exit(1)
	}

	gin.SetMode(gin.TestMode)

	code := m.Run()
	os.Exit(code)
}

func SetupTestEnvironment(t *testing.T) database.IDatabaseService {
	cfg := config.GetConfig()

	// Connect to the test database and run migrations
	dbService := database.NewDatabaseService(cfg.DBPath, cfg.DBMigrationsPath)
	err := dbService.Setup(true)
	if err != nil {
		return nil
	}

	return dbService
}
