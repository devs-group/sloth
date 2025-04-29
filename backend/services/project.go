package services

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"log/slog"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Project struct {
	ID             int    `json:"id" db:"id"`
	UPN            UPN    `json:"upn" db:"unique_name"`
	AccessToken    string `json:"access_token" db:"access_token"`
	Name           string `json:"name" binding:"required" db:"name"`
	OrganisationID string `json:"-" db:"organisation_id"`
	Path           string `json:"-" db:"path"`
	Organisation   string `json:"organisation_name" db:"organisation_name"`
	// Ignored in DB operations - populated separately
	Hook              string             `json:"hook"`
	Services          []*Service         `json:"services"`
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
		container, _, err := generateServiceCompose(service)
		if err != nil {
			return nil, err
		}
		services[service.Usn] = container
	}

	networks := map[string]*compose.Network{
		"traefik": {
			External: true,
		},
		"default": {
			Driver:   utils.StringAsPointer("bridge"),
			External: false,
		},
	}

	dc := &compose.DockerCompose{
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

func (s *S) ListProjects(userID, organisationID string) ([]Project, error) {
	projects := make([]Project, 0)
	query := `
		SELECT DISTINCT p.id, p.unique_name, p.access_token, p.name, p.organisation_id
		FROM projects p
		JOIN organisation_members om ON om.user_id = $1
		WHERE p.organisation_id = $2
	`

	err := s.dbService.GetConn().Select(&projects, query, userID, organisationID)
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

func (s *S) SelectProjectByIDAndOrganisationID(projectID int, currentOrganisationID string) (*Project, error) {
	q := `
		SELECT p.id, p.unique_name, p.access_token, p.name, p.organisation_id, p.path
		FROM projects AS p
		WHERE p.id = $1 AND p.organisation_id = $2
	`

	var project Project
	err := s.dbService.GetConn().Get(&project, q, projectID, currentOrganisationID)
	if err != nil {
		return nil, err
	}

	project.DockerCredentials, err = s.SelectDockerCredentials(project.OrganisationID)
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
		SELECT p.id, p.unique_name, p.access_token, p.name, p.organisation_id, p.path
		FROM projects AS p
		WHERE p.id = $1 AND p.access_token = $2
	`

	var project Project
	err := s.dbService.GetConn().Get(&project, query, projectID, accessToken)
	if err != nil {
		return nil, err
	}

	project.DockerCredentials, err = s.SelectDockerCredentials(project.OrganisationID)
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
		p.organisation_id,
		p.path,
		COALESCE(o.name, '') AS organisation_name
FROM
		projects p
		LEFT JOIN organisations o ON o.id = p.organisation_id
WHERE
    	p.unique_name = $1
		AND (
		    p.access_token = $2
    		OR p.organisation_id = $3
		)
GROUP BY
		p.id,
		p.unique_name,
		p.access_token,
		p.name,
		p.organisation_id,
		p.path,
		o.name
		`

	slog.Debug("Query Params", "unique_name", string(p.UPN), "access_token", p.AccessToken, "organisation_id", p.OrganisationID)
	err := s.dbService.GetConn().Get(p, query, string(p.UPN), p.AccessToken, p.OrganisationID)
	if err != nil {
		return err
	}

	p.DockerCredentials, err = s.SelectDockerCredentials(p.OrganisationID)
	if err != nil {
		return err
	}
	p.Services, err = s.SelectServices(p.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *S) SaveProject(p *Project, currentOrganisationID string) error {
	q1 := `
	INSERT INTO projects (name, unique_name, access_token, organisation_id, path)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`
	err := s.dbService.GetConn().Get(&p.ID, q1, p.Name, p.UPN, p.AccessToken, currentOrganisationID, p.Path)
	if err != nil {
		return err
	}

	for id := range p.Services {
		if err := s.SaveService(p.Services[id], p.UPN, p.ID); err != nil {
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
			WHERE organisation_id = $1 AND unique_name = $2;
		`
		_, err := tx.Exec(q1, p.OrganisationID, p.UPN, p.Name)
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

		// Get existing services from database
		existingServices, err := s.SelectServices(p.ID)
		if err != nil {
			return errors.Wrap(err, "unable to fetch existing services")
		}

		// Create a map of existing services by USN for easy lookup
		existingServiceMap := make(map[string]*Service)
		for _, svc := range existingServices {
			existingServiceMap[svc.Usn] = svc
		}

		// Create a map of updated services by USN
		updatedServiceMap := make(map[string]bool)
		for _, svc := range p.Services {
			if svc.Usn != "" {
				updatedServiceMap[svc.Usn] = true
			}
		}

		// Delete services that are no longer in the project
		for _, existingSvc := range existingServices {
			if !updatedServiceMap[existingSvc.Usn] {
				deleteQuery := `DELETE FROM services WHERE usn = $1 AND project_id = $2`
				_, err = tx.Exec(deleteQuery, existingSvc.Usn, p.ID)
				if err != nil {
					return errors.Wrap(err, "unable to delete removed service")
				}
				slog.Info("Deleted service", "usn", existingSvc.Usn, "projectID", p.ID)
			}
		}

		// Update or insert services
		for _, svc := range p.Services {
			if svc.Usn == "" {
				// Insert new service
				svc.Usn = utils.GenerateRandomName()
				query := `INSERT INTO services (name, usn, project_id, dcj) VALUES ($1, $2, $3, $4)`
				_, serviceJson, err := generateServiceCompose(svc)
				if err != nil {
					return errors.Wrap(err, "unable to generate service for compose")
				}
				_, err = tx.Exec(query, svc.Name, svc.Usn, p.ID, serviceJson)
				if err != nil {
					return errors.Wrap(err, "unable to save a new service")
				}
				slog.Info("Added new service", "name", svc.Name, "usn", svc.Usn, "projectID", p.ID)
			} else {
				// Update existing service
				err := s.UpdateService(tx, svc, p.UPN, p.ID)
				if err != nil {
					return errors.Wrap(err, "unable to update service")
				}
				slog.Info("Updated service", "name", svc.Name, "usn", svc.Usn, "projectID", p.ID)
			}
		}
		return nil
	})
}

func (s *S) DeleteProjectByIDAndOrganisationID(projectID int, organisationID string) error {
	q := `
	DELETE FROM projects
	WHERE
		id = $1 AND
		organisation_id = $2
	`
	res, err := s.dbService.GetConn().Exec(q, projectID, organisationID)
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
