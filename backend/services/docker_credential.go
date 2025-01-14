package services

type DockerCredential struct {
	ID        int    `json:"id" db:"id"`
	Username  string `json:"username" binding:"required" db:"username"`
	Password  string `json:"password" binding:"required" db:"password"`
	Registry  string `json:"registry" binding:"required,uri" db:"registry"`
	ProjectID int    `json:"-" db:"project_id"`
}

func (s *S) SelectDockerCredentials(userID string) ([]DockerCredential, error) {
	var dcs = make([]DockerCredential, 0)
	credsQuery := `SELECT * FROM docker_credentials WHERE project_id = $1`
	err := s.dbService.GetConn().Select(&dcs, credsQuery, userID)
	if err != nil {
		return nil, err
	}
	return dcs, nil
}
