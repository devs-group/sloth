package database

import (
	"fmt"
	"github.com/devs-group/sloth/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"log/slog"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

type IDatabaseService interface {
	Setup(force bool) error
	create(force bool) error
	connect() error
	migrate() error
	Delete() error

	GetConn() *sqlx.DB
	GetDBPath() string
	GetMigrationsPath() string
}

type DatabaseService struct {
	DBPath         string
	MigrationsPath string
	Conn           *sqlx.DB
}

func NewDatabaseService(dbPath, migrationsPath string) IDatabaseService {
	return &DatabaseService{
		DBPath:         dbPath,
		MigrationsPath: migrationsPath,
	}
}

func (d *DatabaseService) Setup(force bool) error {
	err := d.create(force)
	if err != nil {
		return err
	}
	err = d.connect()
	if err != nil {
		return err
	}
	err = d.migrate()
	if err != nil {
		return err
	}
	return nil
}

func (d *DatabaseService) create(force bool) error {
	if err := os.MkdirAll(filepath.Dir(d.DBPath), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create directory: %w", err)
	}

	if _, err := os.Stat(d.DBPath); force || os.IsNotExist(err) {
		if _, err := os.Create(d.DBPath); err != nil {
			return fmt.Errorf("unable to create database file: %w", err)
		}
		slog.Info("Database file created", "dbPath", d.DBPath)
	} else if err != nil {
		return fmt.Errorf("error creating database file: %w", err)
	} else {
		slog.Info("Database file exists already at", "dbPath", d.DBPath)
	}

	return nil
}

func (d *DatabaseService) connect() error {
	db, err := sqlx.Open("sqlite", d.DBPath)
	if err != nil {
		return fmt.Errorf("unable to connect to sqlite db: %w", err)
	}

	// Enable foreign key support
	if _, err = db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("unable to enable foreign key support: %w", err)
	}

	// Set the db also on the struct
	d.Conn = db

	return nil
}

func (d *DatabaseService) migrate() error {
	if err := goose.SetDialect(string(goose.DialectSQLite3)); err != nil {
		return errors.Wrapf(err, `failed to set goose dialect to "%s"`, goose.DialectSQLite3)
	}

	staticPath := filepath.Join(d.MigrationsPath, "static")
	if _, err := os.Stat(staticPath); err == nil {
		slog.Info("Applying static database migrations if required", "path", staticPath)
		if err := goose.Up(d.Conn.DB, staticPath); err != nil {
			return errors.Wrap(err, "failed to apply static migrations")
		}
	}

	// If available, apply also dynamic migrations
	dynamicPath := filepath.Join(d.MigrationsPath, "dynamic")
	if _, err := os.Stat(dynamicPath); err == nil {
		slog.Info("Applying dynamic database migrations if required", "path", dynamicPath)
		goose.SetLogger(goose.NopLogger())
		if err := goose.DownTo(d.Conn.DB, dynamicPath, 0, goose.WithNoVersioning()); err != nil {
			return errors.Wrapf(err, "failed to revert dynamic migrations")
		}
		goose.SetLogger(logger.GooseLogger{})
		if err := goose.Up(d.Conn.DB, dynamicPath, goose.WithNoVersioning()); err != nil {
			return errors.Wrapf(err, "failed to apply dynamic migrations")
		}
	}
	return nil
}

func (d *DatabaseService) Delete() error {
	slog.Info("Trying to delete db file", "dbPath", d.DBPath)
	if _, err := os.Stat(d.DBPath); err == nil {
		if err := os.Remove(d.DBPath); err != nil {
			return fmt.Errorf("unable to delete database file: %w", err)
		}
		slog.Info("Database file deleted", "dbPath", d.DBPath)
	} else {
		return fmt.Errorf("error deleting database file: %w", err)
	}

	return nil
}

func (d *DatabaseService) GetConn() *sqlx.DB {
	return d.Conn
}

func (d *DatabaseService) GetDBPath() string {
	return d.DBPath
}

func (d *DatabaseService) GetMigrationsPath() string {
	return d.MigrationsPath
}
