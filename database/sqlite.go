package database

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

type Store struct {
	DB *sqlx.DB
}

func connect() {
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

func NewStore() *Store {
	if DB == nil {
		connect()
	}
	return &Store{
		DB: DB,
	}
}

type Project struct {
	ID          int    `db:"id"`
	UniqueName  string `db:"unique_name"`
	DCJ         string `db:"dcj"`
	AccessToken string `db:"access_token"`
}

func (s *Store) GetProjectByNameAndAccessToken(upn string, accessToken string) (*Project, error) {
	var p Project
	err := s.DB.Get(&p, "SELECT id, unique_name, dcj, access_token FROM projects WHERE unique_name=$1 AND access_token=$2", upn, accessToken)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) InsertProjectWithTx(upn string, accessToken string, dcj string, cb func() error) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("INSERT INTO projects (unique_name, access_token, dcj) VALUES ($1, $2, $3)", upn, accessToken, dcj)
	if err != nil {
		return err
	}
	err = cb()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) SelectProjects() ([]Project, error) {
	var projects []Project
	err := s.DB.Select(&projects, "SELECT id, unique_name, dcj, access_token FROM projects")
	if err != nil {
		return nil, err
	}
	return projects, nil
}
