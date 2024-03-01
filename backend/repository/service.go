package repository

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/jmoiron/sqlx"
)

type Public struct {
	Enabled  bool     `json:"enabled"`
	Hosts    []string `json:"hosts" binding:"required"`
	Port     string   `json:"port" binding:"required,numeric"`
	SSL      bool     `json:"ssl"`
	Compress bool     `json:"compress"`
}

type Service struct {
	Ports    []string   `json:"ports" binding:"gt=0"`
	Image    string     `json:"image" binding:"required"`
	ImageTag string     `json:"image_tag" binding:"required"`
	Command  string     `json:"command"`
	Public   Public     `json:"public"`
	EnvVars  [][]string `json:"env_vars"`
	Volumes  []string   `json:"volumes" binding:"dive,dirpath"`

	Name      string `json:"name" binding:"required" db:"name"`
	Usn       string `json:"usn" db:"usn"`
	ProjectID int    `json:"-" db:"project_id"`
	DCJ       string `json:"-" db:"dcj"`
}

func DeleteMissingServices(upn UPN, projectID int, services []Service, tx *sqlx.Tx) error {
	usn := make([]string, len(services))
	for i, s := range services {
		usn[i] = s.Usn
	}
	usnJSON, err := json.Marshal(usn)
	if err != nil {
		return err
	}

	query := `
	DELETE FROM services 
	WHERE 
		project_id = $1 AND
		name IN ( 
			SELECT s.name 
			FROM services s, json_each(s.dcj)
			WHERE key NOT IN (SELECT value FROM json_each($2))
		)
	RETURNING( SELECT key FROM json_each(dcj, '$') ) as key;
	`
	var deletedServices []string
	err = tx.Select(&deletedServices, query, projectID, usnJSON)
	if err != nil {
		slog.Error("can't delete services")
		return err
	}

	for _, folder := range deletedServices {
		utils.DeleteFolder(path.Join(upn.GetProjectPath(), config.PersistentVolumeDirectoryName, folder))
	}

	return nil
}

func SelectServices(projectID int, tx *sqlx.Tx) ([]Service, error) {
	services := make([]Service, 0)
	slog.Info("SEL", "services", projectID)
	query := `
	SELECT json_extract(dcj, '$."' || key || '"') AS dcj, key as usn, project_id, name
	FROM services,
		 json_each(json_extract(dcj, '$')) 
	WHERE project_id = $1
	ORDER BY project_id DESC
    `

	err := tx.Select(&services, query, projectID)
	if err != nil {
		slog.Error("your database state is corrupted - check dcj for invalid json fields")
		return nil, err
	}

	for id := range services {
		service, err := services[id].ReadServiceFromDCJ(services[id].DCJ)
		if err != nil {
			slog.Error("error read service from dcj", "err", err)
			continue
		}
		services[id] = *service
	}

	return services, nil
}

func (s *Service) ReadServiceFromDCJ(dcj string) (*Service, error) {
	var sc compose.Container
	err := compose.FromString(dcj, &s)
	if err != nil {
		slog.Error("unable to parse docker compose json string", "err", err)
		return nil, err
	}
	hosts, err := sc.Labels.GetHosts()
	if err != nil {
		slog.Error("unable to get host from labels", "err", err)
	}
	// When no hosts are set, response with empty string
	if len(hosts) == 0 {
		hosts = []string{""}
	}

	image := strings.Split(s.Image, ":")
	if len(image) < 2 {
		return nil, fmt.Errorf("unsuported image, expected 'image:tag' format got: %s", s.Image)
	}

	envVars := make([][]string, len(sc.Environment))
	for i, e := range sc.Environment {
		kv := strings.Split(e, "=")
		envVars[i] = kv
	}

	// When no env vars are set, response with empty tuple
	if len(sc.Environment) == 0 {
		envVars = [][]string{{"", ""}}
	}

	volumes := make([]string, len(s.Volumes))
	for i, v := range s.Volumes {
		volumes[i] = strings.Split(v, ":")[1]
	}

	// When no volumes are set, response with empty string
	if len(s.Volumes) == 0 {
		volumes = []string{""}
	}

	port, err := sc.Labels.GetPort()
	if err != nil {
		slog.Error("unable to get port from labels", "err", err)
	}

	service := Service{
		Name:     s.Name,
		Usn:      s.Usn,
		Ports:    s.Ports,
		Command:  s.Command,
		Image:    image[0],
		ImageTag: image[1],
		EnvVars:  envVars,
		Volumes:  volumes,
		Public: Public{
			Enabled:  sc.Labels.IsPublic(),
			Hosts:    hosts,
			Port:     port,
			SSL:      sc.Labels.IsSSL(),
			Compress: sc.Labels.IsCompress(),
		},
	}

	return &service, nil
}

// UpdateService inserts a new service with its DCJ for a given projectID into the database.
func (s *Service) UpsertService(upn UPN, projectID int, tx *sqlx.Tx) error {
	if s.Usn == "" {
		return s.SaveService(upn, projectID, tx)
	} else {
		if _, err := s.GenerateServiceCompose(upn, projectID); err != nil {
			return err
		}

		query := `
		SELECT COALESCE (json_extract(value, '$.volumes'), "[]" ) as volumes 
		FROM services, json_each(dcj, '$') 
		WHERE  project_id = $1 AND json_extract(dcj, ('$."' || $2 || '"')) IS NOT NULL;
		`
		var dbVolumes string
		if err := tx.Get(&dbVolumes, query, projectID, s.Usn); err != nil {
			slog.Error("Error", "cant get volumes", err)
			return err
		}

		var volumes []string
		if err := json.Unmarshal([]byte(dbVolumes), &volumes); err != nil {
			slog.Error("Error", "cant unmarshal db volumes", err)
			return err
		}

		query = `
    		UPDATE services SET dcj = $3, name = $2
    			WHERE project_id = $1 AND json_extract(dcj, ('$."' || $4 || '"')) IS NOT NULL;
		`
		_, err := tx.Exec(query, projectID, s.Name, s.DCJ, s.Usn)
		if err != nil {
			slog.Error("Error", "error updating services", err)
			return err
		}

		newVolumesMap := make(map[string]bool)
		for _, vol := range s.Volumes {
			newVolumesMap["./"+path.Join(s.getServicePath(), vol)] = true
		}

		for _, origVolume := range volumes {
			vPath := strings.Split(origVolume, ":")[0]
			if _, exists := newVolumesMap[vPath]; !exists {
				utils.DeleteFolder((path.Join(upn.GetProjectPath(), vPath)))
			}
		}
	}
	return nil
}

// SaveService inserts a new service with its DCJ for a given projectID into the database.
func (s *Service) SaveService(upn UPN, projectID int, tx *sqlx.Tx) error {
	if s.Usn != "" {
		return fmt.Errorf("Service already have an USN - upsert the service!")
	}

	s.Usn = utils.GenerateRandomName()
	query := `INSERT INTO services (name, project_id, dcj)	VALUES ($1, $2, $3)`
	if _, err := s.GenerateServiceCompose(upn, projectID); err != nil {
		return err
	}

	_, err := tx.Exec(query, s.Name, projectID, s.DCJ)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getServicePath() string {
	return fmt.Sprintf("./%s/%s", config.PersistentVolumeDirectoryName, sanitizeName(s.Usn))
}

func (s *Service) GenerateServiceCompose(upn UPN, projectID int) (*compose.Container, error) {
	sanitizedServiceName := sanitizeName(s.Usn)
	c := &compose.Container{
		Image:    fmt.Sprintf("%s:%s", s.Image, s.ImageTag),
		Restart:  "always",
		Networks: []string{"web", "default"},
		Ports:    s.Ports,
	}

	if s.Command != "" {
		c.Command = s.Command
	}

	for _, ev := range s.EnvVars {
		if len(ev) == 2 && ev[0] != "" && ev[1] != "" {
			c.Environment = append(c.Environment, fmt.Sprintf("%s=%s", ev[0], ev[1]))
		}
	}

	if len(s.Volumes) > 0 && upn != "" && s.Volumes[0] != "" {
		for _, v := range s.Volumes {
			dataPath := v

			if strings.HasPrefix(v, "/") {
				dataPath, _ = strings.CutPrefix(v, "/")
			}
			c.Volumes = append(c.Volumes, fmt.Sprintf("%s/%s:%s", s.getServicePath(), dataPath, v))
		}
	}

	if s.Public.Enabled {
		usn := fmt.Sprintf("%s-%s", upn, sanitizedServiceName)
		hosts := []string{fmt.Sprintf("Host(`%s.devs-group.ch)", strings.ToLower(usn))}

		if len(s.Public.Hosts) > 0 && s.Public.Hosts[0] != "" {
			hosts = make([]string, len(s.Public.Hosts))
			for idx, h := range s.Public.Hosts {
				hosts[idx] = fmt.Sprintf("Host(`%s`)", strings.ToLower(h))
			}
		}

		labels := []string{
			"traefik.enable=true",
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%s", usn, s.Public.Port),
			// It's weird but yaml parser creates a new-line in yaml when we use || with empty spaces between hosts.
			fmt.Sprintf("traefik.http.routers.%s.rule=%s", usn, strings.Join(hosts, "||")),
		}

		if s.Public.SSL {
			labels = append(
				labels,
				fmt.Sprintf("traefik.http.routers.%s.entrypoints=https", usn),
				fmt.Sprintf("traefik.http.routers.%s.tls=true", usn),
				fmt.Sprintf("traefik.http.routers.%s.tls.certresolver=le", usn),
			)
		}

		if s.Public.Compress {
			labels = append(
				labels,
				fmt.Sprintf("traefik.http.middlewares.%s-compress.compress=true", usn),
				fmt.Sprintf("traefik.http.routers.%s.middlewares=%s-compress", usn, usn),
			)
		}

		c.Labels = labels
	}

	ctn, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	s.DCJ = "{\"" + s.Usn + "\":" + string(ctn) + "}"
	return c, nil
}

func sanitizeName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
