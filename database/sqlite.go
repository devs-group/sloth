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

	DB, err = sqlx.Open("sqlite", "./database/database.sqlite")
	if err != nil {
		slog.Error("unable to connect to sqlite db", "err", err)
		panic(err)
	}
	DB.MustExec(`	
	CREATE TABLE IF NOT EXISTS
	projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		unique_name VARCHAR(255),
		access_token VARCHAR(255),
		dcj JSON,
		name VARCHAR(255),
		user_id VARCHAR(255),
		path VARCHAR(255)
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
	UserID      string `db:"user_id"`
	Name        string `db:"name"`
	UniqueName  string `db:"unique_name"`
	DCJ         string `db:"dcj"`
	AccessToken string `db:"access_token"`
	Path        string `db:"path"`
}

func (s *Store) GetProjectByNameAndAccessToken(upn, accessToken string) (*Project, error) {
	var p Project
	q := `
	SELECT id, name, unique_name, dcj, access_token FROM projects WHERE unique_name=$1 AND access_token=$2
	`
	err := s.DB.Get(&p, q, upn, accessToken)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) InsertProjectWithTx(userID, name, upn, accessToken, dcj, path string, cb func() error) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	q := `
	INSERT INTO projects (name, unique_name, access_token, dcj, user_id, path) VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(q, name, upn, accessToken, dcj, userID, path)
	if err != nil {
		return err
	}
	err = cb()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpdateProjectWithTx(userID, upn, name, dcj string, cb func() error) (*Project, error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck
	var p Project
	q := `
    UPDATE projects
    SET
        name = $3,
    	dcj = $4
    WHERE user_id = $1 AND unique_name = $2
    RETURNING *;
	`
	err = tx.Get(&p, q, userID, upn, name, dcj)
	if err != nil {
		return nil, err
	}
	err = cb()
	if err != nil {
		return nil, err
	}
	return &p, tx.Commit()
}

func (s *Store) SelectProjects(userID string) ([]Project, error) {
	var projects []Project
	err := s.DB.Select(&projects, "SELECT * FROM projects WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *Store) SelectProjectByUPN(userID, upn string) (*Project, error) {
	var p Project
	q := `
		SELECT *
		FROM projects
		WHERE user_id = $1 AND unique_name = $2;
	`
	err := s.DB.Get(&p, q, userID, upn)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) DeleteProjectByUPNWithTx(userID, upn string, cb func() error) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	q := `
		DELETE
		FROM projects
		WHERE user_id = $1 AND unique_name = $2;
	`
	_, err = s.DB.Exec(q, userID, upn)
	if err != nil {
		return err
	}
	err = cb()
	if err != nil {
		return err
	}
	return tx.Commit()
}
