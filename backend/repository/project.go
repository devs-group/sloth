package repository

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/jmoiron/sqlx"
)

type Project struct {
	ID          int    `json:"id" db:"id"`
	UPN         UPN    `json:"upn" db:"unique_name"`
	AccessToken string `json:"access_token" db:"access_token"`
	Name        string `json:"name" binding:"required" db:"name"`
	UserID      string `json:"-" db:"user_id"`
	Path        string `json:"-" db:"path"`
	Group       string `json:"group_name" db:"group_name"`
	// Ignored in DB operations - populated separately
	Hook              string             `json:"hook"`
	Services          []Service          `json:"services"`
	DockerCredentials []DockerCredential `json:"docker_credentials"`

	//Ignore in both - populated internal
	CTN compose.Services `json:"-"`
}

func (p *Project) PrepareProject() error {
	if err := p.CreateProjectServiceDirectories(); err != nil {
		return err
	}

	dc, err := p.GenerateDockerCompose()
	if err != nil {
		return err
	}
	p.CTN = dc.Services
	return SaveDockerComposeFile(p.UPN, *dc)
}

func (p *Project) GenerateDockerCompose() (*compose.DockerCompose, error) {
	// Initialize the services map
	services := make(map[string]*compose.Container)
	for _, serv := range p.Services {
		srv, err := serv.GenerateServiceCompose(p.UPN, p.ID)
		if err != nil {
			return nil, err
		}
		services[serv.Name] = srv
	}

	networks := map[string]*compose.Network{
		"web": {
			External: config.Environment == config.Production, // Adjust based on environment
		},
		"default": {
			Driver:   "bridge",
			External: false,
		},
	}

	dc := &compose.DockerCompose{
		Version:  "3.9",
		Networks: networks,
		Services: services,
	}
	return dc, nil
}

func SaveDockerComposeFile(upn UPN, dc compose.DockerCompose) error {
	dcy, err := dc.ToYAML()
	if err != nil {
		return err
	}
	return CreateDockerComposeFile(upn, dcy)
}

func CreateDockerComposeFile(upn UPN, yaml string) error {
	p := fmt.Sprintf("%s/%s/%s", filepath.Clean(config.ProjectsDir), upn, config.DockerComposeFileName)
	filePerm := 0600
	err := os.WriteFile(p, []byte(yaml), os.FileMode(filePerm))
	if err != nil {
		return fmt.Errorf("unable to write file %s: err %v", p, err)
	}
	return nil
}

func (p *Project) CreateProjectServiceDirectories() error {
	if p.HasVolumesInRequest() {
		for _, service := range p.Services {
			if _, err := utils.CreateFolderIfNotExists(path.Join(p.UPN.GetProjectPath(), config.PersistentVolumeDirectoryName, service.Usn)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Project) HasVolumesInRequest() bool {
	hasVolumes := false
	for i := range p.Services {
		if len(p.Services[i].Volumes) > 0 {
			hasVolumes = true
		}
	}
	return hasVolumes
}

func SelectProjects(userID string, tx *sqlx.Tx) ([]Project, error) {
	var projects []Project
	query := `SELECT p.unique_name, p.access_token, p.user_id 
			  FROM projects p 
				LEFT JOIN groups g ON g.id = p.group_id 
				LEFT JOIN group_members gm ON gm.group_id = g.id 
			  WHERE p.user_id = $1 OR gm.user_id = $1 GROUP BY p.user_id`

	err := tx.Select(&projects, query, userID)
	if err != nil {
		return nil, err
	}

	// extract details of each project from database
	for i := range projects {
		err := projects[i].SelectProjectByUPNOrAccessToken(tx)
		if err != nil {
			return nil, err
		}
	}

	return projects, nil
}

func (p *Project) SelectProjectByUPNOrAccessToken(tx *sqlx.Tx) error {

	query := `
        SELECT p.id, p.unique_name, p.access_token, 
			   p.name, p.user_id, p.path, 
			   COALESCE( o.name, "" ) as group_name FROM projects p
		LEFT JOIN groups o ON p.group_id = o.id 
		LEFT JOIN group_members gm ON o.id = gm.group_id
        WHERE p.unique_name = $1 AND (
            p.access_token = $2 OR
            p.user_id = $3 OR 
			gm.user_id = $3
        )
    `

	err := tx.Get(p, query, string(p.UPN), p.AccessToken, p.UserID)
	if err != nil {
		return err
	}

	p.DockerCredentials, err = SelectDockerCredentials(p.UserID, tx)
	if err != nil {
		return err
	}
	p.Services, err = SelectServices(p.ID, tx)
	return err
}

func (p *Project) SaveProject(tx *sqlx.Tx) error {
	q1 := `
	INSERT INTO projects (name, unique_name, access_token, user_id, path)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`
	err := tx.Get(&p.ID, q1, p.Name, p.UPN, p.AccessToken, p.UserID, p.Path)
	if err != nil {
		return err
	}

	for id := range p.Services {
		if err := p.Services[id].SaveService(p.UPN, p.ID, tx); err != nil {
			return err
		}
	}

	for _, dc := range p.DockerCredentials {
		q2 := `
        INSERT INTO docker_credentials (username, password, registry, project_id)
        VALUES ($1, $2, $3, $4)
    	`
		_, err = tx.Exec(q2, dc.Username, dc.Password, dc.Registry, p.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) UpdateProject(tx *sqlx.Tx) error {
	q1 := `
		UPDATE projects
		SET name = $3
		WHERE user_id = $1 AND unique_name = $2;
	`

	_, err := tx.Exec(q1, p.UserID, p.UPN, p.Name)
	if err != nil {
		return err
	}

	q2 := `DELETE FROM docker_credentials WHERE project_id = $1`
	_, err = tx.Exec(q2, p.ID)
	if err != nil {
		return err
	}

	for id := range p.Services {
		if err := p.Services[id].UpsertService(p.UPN, p.ID, tx); err != nil {
			return err
		}
	}

	if err := DeleteMissingServices(p.UPN, p.ID, p.Services, tx); err != nil {
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

	return nil
}

func (p *Project) DeleteProjectByUPNWithTx(tx *sqlx.Tx) error {
	q := `
		DELETE
		FROM projects
		WHERE user_id = $1 AND unique_name = $2;
	`
	res, err := tx.Exec(q, p.UserID, p.UPN)
	if err != nil {
		return err
	}

	delCount, err := res.RowsAffected()
	if delCount != 1 {
		return fmt.Errorf("Cant remove project!")
	}

	return nil
}
