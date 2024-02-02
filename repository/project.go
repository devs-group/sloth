package repository

import (
	"github.com/devs-group/sloth/database"
)

type Project struct {
	ID          int    `json:"id" db:"id"`
	UPN         UPN    `json:"upn" db:"unique_name"`
	AccessToken string `json:"access_token" db:"access_token"`
	DCJ         string `json:"dcj" db:"dcj"`
	Name        string `json:"name" binding:"required" db:"name"`
	UserID      string `json:"-" db:"user_id"`
	Path        string `json:"-" db:"path"`

	// Ignored in DB operations - populated separately
	Hook              string             `json:"hook"`
	Services          []Service          `json:"services"`
	DockerCredentials []DockerCredential `json:"docker_credentials"`
}

func SelectProjects(userID string, store *database.Store) ([]Project, error) {
	var projects []Project
	query := `SELECT * FROM projects WHERE user_id = $1`
	err := store.DB.Select(&projects, query, userID)
	if err != nil {
		return nil, err
	}

	for i := range projects {
		err := projects[i].SelectProjectByUPNOrAccessToken(store)
		if err != nil {
			return nil, err
		}
	}

	return projects, nil
}

func (p *Project) SelectProjectByUPNOrAccessToken(store *database.Store) error {
	query := `
        SELECT id, unique_name, access_token, dcj, name, user_id, path
        FROM projects
        WHERE unique_name = $1 AND (
            access_token = $2 OR
            user_id = $3
        )
    `

	err := store.DB.Get(p, query, string(p.UPN), p.AccessToken, p.UserID)
	if err != nil {
		return err
	}

	p.DockerCredentials, err = SelectDockerCredentials(p.UserID, store)
	if err != nil {
		return err
	}

	p.Services, err = ReadServicesFromDCJ(p.DCJ)
	return err
}

func (p *Project) SaveProject(store *database.Store) error {
	tx, err := store.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	var projectID int
	q1 := `
	INSERT INTO projects (name, unique_name, access_token, dcj, user_id, path)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id
	`
	err = tx.Get(&projectID, q1, p.Name, p.UPN, p.AccessToken, p.DCJ, p.UserID, p.Path)
	if err != nil {
		return err
	}

	for _, dc := range p.DockerCredentials {
		q2 := `
        INSERT INTO docker_credentials (username, password, registry, project_id)
        VALUES ($1, $2, $3, $4)
    	`
		_, err = tx.Exec(q2, dc.Username, dc.Password, dc.Registry, projectID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *Project) UpdateProject(s *database.Store) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q1 := `
		UPDATE projects
		SET
			name = $3,
			dcj = $4
		WHERE user_id = $1 AND unique_name = $2;
	`

	_, err = tx.Exec(q1, p.UserID, p.UPN, p.Name, p.DCJ)
	if err != nil {
		return err
	}

	q2 := `DELETE FROM docker_credentials WHERE project_id = $1`
	_, err = tx.Exec(q2, p.ID)
	if err != nil {
		return err
	}

	for _, dc := range p.DockerCredentials {
		q3 := `
			INSERT INTO docker_credentials (username, password, registry, project_id)
			VALUES ($1, $2, $3, $4)
			`
		_, err = tx.Exec(q3, dc.Username, dc.Password, dc.Registry, p.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *Project) DeleteProjectByUPNWithTx(store *database.Store, cb func() error) error {
	tx, err := store.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := `
		DELETE
		FROM projects
		WHERE user_id = $1 AND unique_name = $2;
	`
	_, err = store.DB.Exec(q, p.UserID, p.UPN)
	if err != nil {
		return err
	}
	err = cb()
	if err != nil {
		return err
	}
	return tx.Commit()
}
