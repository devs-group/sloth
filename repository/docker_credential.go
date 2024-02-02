package repository

import (
	"github.com/devs-group/sloth/database"
)

type DockerCredential struct {
	ID        int    `json:"id" db:"id"`
	Username  string `json:"username" binding:"required" db:"username"`
	Password  string `json:"password" binding:"required" db:"password"`
	Registry  string `json:"registry" binding:"required,uri" db:"registry"`
	ProjectID int    `json:"-" db:"project_id"`
}

func SelectDockerCredentials(userID string, store *database.Store) ([]DockerCredential, error) {
	var dcs = make([]DockerCredential, 0)
	credsQuery := `SELECT * FROM docker_credentials WHERE project_id = $1`
	err := store.DB.Select(&dcs, credsQuery, userID)
	if err != nil {
		return nil, err
	}
	return dcs, nil
}
