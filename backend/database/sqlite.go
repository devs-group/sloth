package database

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

type Store struct {
	DB *sqlx.DB
}

func connect() {
	var err error
	path := "./database/database.sqlite"

	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		slog.Error("unable to create directory", "err", err)
		panic(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err := os.Create(path); err != nil {
			slog.Error("unable to create file", "err", err)
			panic(err)
		}
		slog.Info("created directory and file", "path", path)
	} else if err != nil {
		slog.Error("unable to check file", "err", err)
		panic(err)
	}

	slog.Info("connecting to sqlite db...")

	// Connect to the SQLite database
	DB, err = sqlx.Open("sqlite", "./database/database.sqlite")
	if err != nil {
		slog.Error("unable to connect to sqlite db", "err", err)
		panic(err)
	}

	// Enable foreign key support
	_, err = DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		slog.Error("unable to enable foreign key support", "err", err)
		panic(err)
	}

	// Create tables if they don't exist
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS projects (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            unique_name VARCHAR(255),
            access_token VARCHAR(255),
            name VARCHAR(255),
            user_id VARCHAR(255),
            path VARCHAR(255)
        );
        CREATE TABLE IF NOT EXISTS docker_credentials (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username VARCHAR(255),
            password VARCHAR(255),
            registry VARCHAR(255),
            project_id INTEGER,
            FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
        );
        CREATE TABLE IF NOT EXISTS services (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            unique_name VARCHAR(255) UNIQUE,
            project_id INTEGER,
            dcj JSON,
            FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
        );
    `)
	if err != nil {
		slog.Error("unable to create tables", "err", err)
		panic(err)
	}
}

func NewStore() *Store {
	if DB == nil {
		connect()
	}
	return &Store{
		DB: DB,
	}
}
