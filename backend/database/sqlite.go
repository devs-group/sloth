package database

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/devs-group/sloth/backend/config"
)

var DB *sqlx.DB

type Store struct {
	DB *sqlx.DB
}

func initDB(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create directory: %w", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err := os.Create(path); err != nil {
			return fmt.Errorf("unable to create database file: %w", err)
		}
		slog.Info("Database file created", "path", path)
	} else if err != nil {
		return fmt.Errorf("error checking database file: %w", err)
	}

	return nil
}

func connectDB(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to sqlite db: %w", err)
	}

	// Enable foreign key support
	if _, err = db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("unable to enable foreign key support: %w", err)
	}

	return db, nil
}

func runMigrations(db *sqlx.DB, migrationsPath string) error {
	if err := goose.SetDialect(string(goose.DialectSQLite3)); err != nil {
		return fmt.Errorf("setting database dialect for migrations failed: %w", err)
	}
	slog.Info("Applying database migrations if required...")
	if err := goose.Up(db.DB, migrationsPath); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}

func connect() error {
	if err := initDB(config.DBPath); err != nil {
		return err
	}

	db, err := connectDB(config.DBPath)
	if err != nil {
		return err
	}

	if err := runMigrations(db, config.DBMigrationsPath); err != nil {
		return err
	}

	DB = db
	return nil
}

func NewStore() *Store {
	if DB == nil {
		if err := connect(); err != nil {
			panic(err)
		}
	}
	return &Store{DB: DB}
}
