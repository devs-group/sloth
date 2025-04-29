package services

import (
	"database/sql"
	"encoding/json"

	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/devs-group/sloth/backend/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Public struct {
	Enabled  bool     `json:"enabled"`
	Hosts    []string `json:"hosts" binding:"required"`
	Port     string   `json:"port" binding:"required,numeric"`
	SSL      bool     `json:"ssl"`
	Compress bool     `json:"compress"`
}

type Service struct {
	ID          int                          `json:"id" db:"id"`
	Ports       []string                     `json:"ports" binding:"gt=0"`
	Image       string                       `json:"image" binding:"required"`
	ImageTag    string                       `json:"image_tag" binding:"required"`
	Command     string                       `json:"command"`
	Public      Public                       `json:"public"`
	EnvVars     [][]string                   `json:"env_vars"`
	Volumes     []string                     `json:"volumes" binding:"dive,dirpath"`
	Name        string                       `json:"name" binding:"required" db:"name"`
	HealthCheck *compose.HealthCheck         `json:"healthcheck,omitempty" `
	Depends     map[string]compose.Condition `json:"depends_on,omitempty"`
	Deploy      *compose.Deploy              `json:"deploy,omitempty"`
	Usn         string                       `json:"usn" db:"usn"`
	ProjectID   int                          `json:"-" db:"project_id"`
	DCJ         string                       `json:"-" db:"dcj"`
}

func (s *S) DeleteMissingServices(upn UPN, projectID int, services []Service, tx *sqlx.Tx) error {
	cfg := config.GetConfig()

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
			AND project_id = $1
		)
	RETURNING( SELECT key FROM json_each(dcj, '$') ) as key;
	`
	var deletedServices []string
	err = tx.Select(&deletedServices, query, projectID, string(usnJSON))
	if err != nil {
		slog.Error("can't delete services")
		return err
	}

	for _, folder := range deletedServices {
		err := utils.DeleteFolder(path.Join(upn.GetProjectPath(), cfg.PersistentVolumeDirectoryName, folder))
		if err != nil {
			slog.Error("can't delete folder", "err", err)
			return err
		}
	}

	return nil
}

func (s *S) SelectServices(projectID int) ([]*Service, error) {
	services := make([]*Service, 0)
	query := `
	SELECT json_extract(dcj, '$."' || key || '"') AS dcj, key as usn, project_id, name, services.id
	FROM services,
		 json_each(json_extract(dcj, '$'))
	WHERE project_id = $1
	ORDER BY project_id DESC
    `

	err := s.dbService.GetConn().Select(&services, query, projectID)
	if err != nil {
		slog.Error("your database state is corrupted - check dcj for invalid json fields")
		return nil, err
	}

	for id, dbService := range services {
		slog.Info("Service", "s", dbService.DCJ)
		service, err := s.ReadServiceFromDCJ(*dbService)
		if err != nil {
			slog.Error("error read service from dcj", "err", err)
			continue
		}
		rowID := dbService.ID
		services[id] = service
		services[id].ID = rowID
	}

	return services, nil
}

func (s *S) ReadServiceFromDCJ(service Service) (*Service, error) {
	var sc compose.Container
	err := compose.FromString(service.DCJ, &sc)
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

	image := strings.Split(sc.Image, ":")
	if len(image) < 2 {
		return nil, fmt.Errorf("unsuported image, expected 'image:tag' format got: %s", service.Image)
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

	volumes := make([]string, len(sc.Volumes))
	for i, v := range sc.Volumes {
		volumes[i] = strings.Split(v, ":")[1]
	}

	// When no volumes are set, response with empty string
	if len(sc.Volumes) == 0 {
		volumes = []string{""}
	}

	port, err := sc.Labels.GetPort()
	if err != nil {
		slog.Error("unable to get port from labels", "err", err)
	}

	return &Service{
		Name:        service.Name,
		Usn:         service.Usn,
		Ports:       sc.Ports,
		Command:     sc.Command,
		Image:       image[0],
		ImageTag:    image[1],
		EnvVars:     envVars,
		Volumes:     volumes,
		HealthCheck: sc.HealthCheck,
		Depends:     sc.Depends,
		Deploy:      sc.Deploy,
		Public: Public{
			Enabled:  sc.Labels.IsPublic(),
			Hosts:    hosts,
			Port:     port,
			SSL:      sc.Labels.IsSSL(),
			Compress: sc.Labels.IsCompress(),
		},
	}, nil
}

// UpdateService updates the service
func (s *S) UpdateService(tx *sqlx.Tx, service *Service, upn UPN, projectID int) error {
	query := `
	SELECT COALESCE (json_extract(value, '$.volumes'), "[]" ) as volumes
	FROM services, json_each(dcj, '$')
	WHERE  project_id = $1 AND json_extract(dcj, ('$."' || $2 || '"')) IS NOT NULL;
	`
	var v string
	err := tx.Get(&v, query, projectID, service.Usn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error("can't get volumes from dcj", "err", err)
		return err
	}

	var volumes []string
	if v != "" {
		if err := json.Unmarshal([]byte(v), &volumes); err != nil {
			slog.Error("can't unmarshal volumes", "err", err)
			return err
		}
	}

	_, serviceJSON, err := generateServiceCompose(service)
	if err != nil {
		return err
	}
	query = `
    		UPDATE services SET dcj = $3, name = $2
    			WHERE project_id = $1 AND json_extract(dcj, ('$."' || $4 || '"')) IS NOT NULL;
	`
	_, err = tx.Exec(query, projectID, service.Name, serviceJSON, service.Usn)
	if err != nil {
		slog.Error("error updating services", "err", err)
		return err
	}

	//volumeName := fmt.Sprintf("%s-%s", strings.ToLower(service.Name), strings.ToLower(service.Usn))
	newVolumesMap := make(map[string]bool)
	for _, vol := range service.Volumes {
		newVolumesMap["./"+path.Join(service.getServicePath(), vol)] = true
	}

	for _, origVolume := range volumes {
		vPath := strings.Split(origVolume, ":")[0]
		if _, exists := newVolumesMap[vPath]; !exists {
			err := utils.DeleteFolder(path.Join(upn.GetProjectPath(), vPath))
			if err != nil {
				slog.Error("can't delete folder", "err", err)
				return err
			}
		}
	}
	return nil
}

// SaveService inserts a new service with its DCJ for a given projectID into the database.
func (s *S) SaveService(service *Service, upn UPN, projectID int) error {
	if service.Usn != "" {
		return fmt.Errorf("service already have an USN - update the service")
	}

	service.Usn = utils.GenerateRandomName()

	_, serviceJSON, err := generateServiceCompose(service)
	if err != nil {
		return errors.Wrap(err, "unable to generate service compose")
	}

	query := `INSERT INTO services (name, project_id, dcj)	VALUES ($1, $2, $3)`
	_, err = s.dbService.GetConn().Exec(query, service.Name, projectID, serviceJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getServicePath() string {
	cfg := config.GetConfig()
	return fmt.Sprintf("./%s/%s", cfg.PersistentVolumeDirectoryName, sanitizeName(s.Usn))
}

func generateServiceCompose(service *Service) (*compose.Container, string, error) {
	cfg := config.GetConfig()

	c := &compose.Container{
		Image:    fmt.Sprintf("%s:%s", service.Image, service.ImageTag),
		Restart:  "always",
		Networks: []string{"traefik", "default"},
		Ports:    service.Ports,
	}

	if service.Depends != nil {
		c.Depends = service.Depends
	}

	// TODO: Add back when we have a proper solution
	//if service.HealthCheck != nil {
	//	c.HealthCheck = service.HealthCheck
	//}

	if service.Command != "" {
		c.Command = service.Command
	}

	if service.Deploy != nil {
		c.Deploy = service.Deploy
	}

	if c.Deploy == nil {
		c.Deploy = new(compose.Deploy)
	}

	if c.Deploy.Resources == nil {
		c.Deploy.Resources = new(compose.Resources)
	}

	c.Deploy.Resources.Limits = &compose.Limits{
		CPUs:   &cfg.DockerContainerLimits.CPUs,
		Memory: &cfg.DockerContainerLimits.Memory,
	}

	c.Deploy.Resources.Reservations = &compose.Reservations{
		CPUs:   &cfg.DockerContainerLimits.CPUs,
		Memory: &cfg.DockerContainerLimits.Memory,
	}
	c.Deploy.Replicas = &cfg.DockerContainerReplicas

	for _, ev := range service.EnvVars {
		if len(ev) == 2 && ev[0] != "" && ev[1] != "" {
			c.Environment = append(c.Environment, fmt.Sprintf("%s=%s", ev[0], ev[1]))
		}
	}

	if len(service.Volumes) > 0 && service.Volumes[0] != "" {
		for _, v := range service.Volumes {
			dataPath := v

			if strings.HasPrefix(v, "/") {
				dataPath, _ = strings.CutPrefix(v, "/")
			}
			c.Volumes = append(c.Volumes, fmt.Sprintf("%s/%s:%s", service.getServicePath(), dataPath, v))
		}
	}

	usn := sanitizeName(service.Usn)
	if service.Public.Enabled {
		hosts := []string{fmt.Sprintf("Host(`%s.%s`)", usn, cfg.BackendHost)}
		if len(service.Public.Hosts) > 0 && service.Public.Hosts[0] != "" {
			hosts = make([]string, len(service.Public.Hosts))
			for idx, h := range service.Public.Hosts {
				hosts[idx] = fmt.Sprintf("Host(`%s`)", strings.ToLower(h))
			}
		}

		labels := []string{
			"traefik.enable=true",
			// It's weird but yaml parser creates a new-line in yaml when we use || with empty spaces between hosts.
			fmt.Sprintf("traefik.http.routers.%s.rule=%s", usn, strings.Join(hosts, "||")),
			fmt.Sprintf("traefik.http.routers.%s.service=%s", usn, usn),
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%s", usn, service.Public.Port),
		}

		if utils.IsProduction() && service.Public.SSL {
			labels = append(
				labels,
				fmt.Sprintf("traefik.http.routers.%s.entrypoints=https", usn),
				fmt.Sprintf("traefik.http.routers.%s.tls=true", usn),
				fmt.Sprintf("traefik.http.routers.%s.tls.certresolver=le", usn),
			)
		} else {
			labels = append(
				labels,
				fmt.Sprintf("traefik.http.routers.%s.entrypoints=http", usn),
			)
		}

		if service.Public.Compress {
			labels = append(
				labels,
				fmt.Sprintf("traefik.http.middlewares.%s-compress.compress=true", usn),
			)
		}

		c.Labels = labels
	}

	ctn, err := json.Marshal(c)
	if err != nil {
		return nil, "", err
	}

	serviceJson := "{\"" + usn + "\":" + string(ctn) + "}"
	return c, serviceJson, nil
}

func sanitizeName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
