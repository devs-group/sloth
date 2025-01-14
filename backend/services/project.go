package services

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Project struct {
	ID           int    `json:"id" db:"id"`
	UPN          UPN    `json:"upn" db:"unique_name"`
	AccessToken  string `json:"access_token" db:"access_token"`
	Name         string `json:"name" binding:"required" db:"name"`
	UserID       string `json:"-" db:"user_id"`
	Path         string `json:"-" db:"path"`
	Organisation string `json:"organisation_name" db:"organisation_name"`
	// Ignored in DB operations - populated separately
	Hook              string             `json:"hook"`
	Services          []Service          `json:"services"`
	DockerCredentials []DockerCredential `json:"docker_credentials"`

	//Ignore in both - populated internal
	ComposeServices compose.Services `json:"-"`
}

func (s *S) PrepareProject(p *Project) error {
	if _, err := utils.CreateFolderIfNotExists(path.Join(p.UPN.GetProjectPath())); err != nil {
		return err
	}

	if err := s.CreateProjectServiceDirectories(p); err != nil {
		return err
	}

	dc, err := s.GenerateDockerCompose(p)
	if err != nil {
		return err
	}
	p.ComposeServices = dc.Services
	return s.SaveDockerComposeFile(p.UPN, *dc)
}

func (s *S) GenerateDockerCompose(p *Project) (*compose.DockerCompose, error) {
	services := make(map[string]*compose.Container)
	for _, service := range p.Services {
		if service.Usn == "" {
			service.Usn = utils.GenerateRandomName()
		}
		container, _, err := generateServiceCompose(&service)
		if err != nil {
			return nil, err
		}
		services[service.Usn] = container
	}

	networks := map[string]*compose.Network{
		"web": {
			External: utils.IsProduction(), // Adjust based on environment
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

func (s *S) SaveDockerComposeFile(upn UPN, dc compose.DockerCompose) error {
	dcy, err := dc.ToYAML()
	if err != nil {
		return err
	}
	return s.CreateDockerComposeFile(upn, dcy)
}

func (s *S) CreateDockerComposeFile(upn UPN, yaml string) error {
	cfg := config.GetConfig()

	p := fmt.Sprintf("%s/%s/%s", filepath.Clean(cfg.ProjectsDir), upn, cfg.DockerComposeFileName)
	filePerm := 0600
	err := os.WriteFile(p, []byte(yaml), os.FileMode(filePerm))
	if err != nil {
		return fmt.Errorf("unable to write file %s: err %v", p, err)
	}
	return nil
}

func (s *S) CreateProjectServiceDirectories(p *Project) error {
	cfg := config.GetConfig()

	if s.HasVolumesInRequest(p) {
		for _, service := range p.Services {
			if _, err := utils.CreateFolderIfNotExists(path.Join(p.UPN.GetProjectPath(), cfg.PersistentVolumeDirectoryName, service.Usn)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *S) HasVolumesInRequest(p *Project) bool {
	hasVolumes := false
	for i := range p.Services {
		if len(p.Services[i].Volumes) > 0 {
			hasVolumes = true
		}
	}
	return hasVolumes
}

func (s *S) SelectProjects(userID string) ([]Project, error) {
	projects := make([]Project, 0)
	query := `SELECT DISTINCT p.id, p.unique_name, p.access_token, p.user_id
	FROM projects p
	LEFT JOIN projects_in_organisations pg ON p.id = pg.project_id
	LEFT JOIN organisations o ON pg.organisation_id = o.id
	LEFT JOIN organisation_members om ON om.organisation_id = o.id
	WHERE p.user_id = $1 OR om.user_id = $1
	`

	err := s.dbService.GetConn().Select(&projects, query, userID)
	if err != nil {
		return nil, err
	}

	for i := range projects {
		err := s.SelectProjectByUPNOrAccessToken(&projects[i])
		if err != nil {
			return nil, err
		}
	}

	return projects, nil
}

func (s *S) SelectProjectByIDAndUserID(projectID int, userID string) (*Project, error) {
	q := `
		SELECT p.id, p.unique_name, p.access_token, p.name, p.user_id, p.path
		FROM projects AS p
		WHERE p.id = $1 AND p.user_id = $2
	`

	var project Project
	err := s.dbService.GetConn().Get(&project, q, projectID, userID)
	if err != nil {
		return nil, err
	}

	project.DockerCredentials, err = s.SelectDockerCredentials(project.UserID)
	if err != nil {
		return nil, err
	}

	project.Services, err = s.SelectServices(project.ID)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *S) SelectProjectByIDAndAccessToken(projectID int, accessToken string) (*Project, error) {
	query := `
		SELECT p.id, p.unique_name, p.access_token, p.name, p.user_id, p.path
		FROM projects AS p
		WHERE p.id = $1 AND p.access_token = $2
	`

	var project Project
	err := s.dbService.GetConn().Get(&project, query, projectID, accessToken)
	if err != nil {
		return nil, err
	}

	project.DockerCredentials, err = s.SelectDockerCredentials(project.UserID)
	if err != nil {
		return nil, err
	}

	project.Services, _ = s.SelectServices(project.ID)

	return &project, nil
}

func (s *S) SelectProjectByUPNOrAccessToken(p *Project) error {
	query := `
	SELECT
		p.id,
		p.unique_name,
		p.access_token,
		p.name,
		p.user_id,
		p.path,
		COALESCE(o.name, '') AS organisation_name
FROM
		projects p
		LEFT JOIN projects_in_organisations pg ON pg.project_id = p.id
		LEFT JOIN organisations o ON pg.organisation_id = o.id
		LEFT JOIN organisation_members om ON o.id = om.organisation_id
WHERE
		p.unique_name = $1 AND (
				p.access_token = $2 OR
				p.user_id = $3 OR
				om.user_id = $3
		)
GROUP BY
		p.id,
		p.unique_name,
		p.access_token,
		p.name,
		p.user_id,
		p.path,
		o.name
		`

	err := s.dbService.GetConn().Get(p, query, string(p.UPN), p.AccessToken, p.UserID)
	if err != nil {
		return err
	}

	p.DockerCredentials, err = s.SelectDockerCredentials(p.UserID)
	if err != nil {
		return err
	}
	p.Services, err = s.SelectServices(p.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *S) SaveProject(p *Project) error {
	q1 := `
	INSERT INTO projects (name, unique_name, access_token, user_id, path)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`
	err := s.dbService.GetConn().Get(&p.ID, q1, p.Name, p.UPN, p.AccessToken, p.UserID, p.Path)
	if err != nil {
		return err
	}

	for id := range p.Services {
		if err := s.SaveService(&p.Services[id], p.UPN, p.ID); err != nil {
			return err
		}
	}

	for _, dc := range p.DockerCredentials {
		q2 := `
        INSERT INTO docker_credentials (username, password, registry, project_id)
        VALUES ($1, $2, $3, $4)
    	`
		_, err = s.dbService.GetConn().Exec(q2, dc.Username, dc.Password, dc.Registry, p.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *S) UpdateProject(p *Project) error {
	return s.WithTransaction(func(tx *sqlx.Tx) error {
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
		for _, svc := range p.Services {
			if svc.Usn == "" {
				svc.Usn = utils.GenerateRandomName()
				query := `INSERT INTO services (name, project_id, dcj)	VALUES ($1, $2, $3)`
				_, serviceJson, err := generateServiceCompose(&svc)
				if err != nil {
					return errors.Wrap(err, "unable to generate service for compose")
				}
				_, err = tx.Exec(query, svc.Name, p.ID, serviceJson)
				if err != nil {
					return errors.Wrap(err, "unable to save a new service")
				}
			} else {
				err := s.UpdateService(tx, &svc, p.UPN, p.ID)
				if err != nil {
					return errors.Wrap(err, "unable to update service")
				}
			}
		}

		// TODO: Fix this. When service is removed, please delete if from the db
		/**
		if err := s.DeleteMissingServices(p.UPN, p.ID, p.Services, tx); err != nil {
			return err
		}
		*/
		return nil
	})
}

func (s *S) DeleteProjectByIDAndUserID(projectID int, userID string) error {
	q := `
	DELETE FROM projects
	WHERE
		id = $1 AND
		user_id = $2 AND
		NOT EXISTS (
			SELECT 1 FROM projects_in_organisations
			WHERE projects_in_organisations.project_id = projects.id
		);
	`
	res, err := s.dbService.GetConn().Exec(q, projectID, userID)
	if err != nil {
		return err
	}

	delCount, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("can't get affected rows %v", err)
	}

	if delCount == 0 {
		return fmt.Errorf("can't remove project! Verify that this project isn't used by any organisation")
	}

	return nil
}
