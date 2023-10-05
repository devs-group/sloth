package database

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func Connect() {
	var err error
	DB, err = sqlx.Open("sqlite3", "./database/database.sqlite")
	DB.MustExec(`	
	CREATE TABLE IF NOT EXISTS
	projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		unique_name VARCHAR(255),
		access_token VARCHAR(255),
		dcj JSON
	);
	`)
	if err != nil {
		slog.Error("unable to connect to sqlite db", "err", err)
		panic(err)
	}
}
