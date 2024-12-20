package services

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
	ID                int                          `json:"id" db:"id"`
	Ports             []string                     `json:"ports" binding:"gt=0"`
	Image             string                       `json:"image" binding:"required"`
	ImageTag          string                       `json:"image_tag" binding:"required"`
	Command           string                       `json:"command"`
	Public            Public                       `json:"public"`
	EnvVars           [][]string                   `json:"env_vars"`
	Volumes           []string                     `json:"volumes" binding:"dive,dirpath"`
	Name              string                       `json:"name" binding:"required" db:"name"`
	HealthCheck       *compose.HealthCheck         `json:"healthcheck,omitempty" `
	Depends           map[string]compose.Condition `json:"depends_on,omitempty"`
	Deploy            *compose.Deploy              `json:"deploy,omitempty"`
	Usn               string                       `json:"usn" db:"usn"`
	ProjectID         int                          `json:"-" db:"project_id"`
	DCJ               string                       `json:"-" db:"dcj"`
	PostDeployActions []PostDeployAction           `json:"post_deploy_actions"`
}

func (s *S) DeleteMissingServices(upn UPN, projectID int, services []Service, tx *sqlx.Tx) error {
	usn := make([]string, len(services))
	for i, s := range services {
		usn[i] = s.Usn

	}

	if ok, err := SearchNotInElementsDependsOn(usn, projectID, tx); err != nil || !ok {
		if err != nil {
			return err
		}
		return fmt.Errorf("can't delete service, service is in use: %s", usn)
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
		err := utils.DeleteFolder(path.Join(upn.GetProjectPath(), config.PersistentVolumeDirectoryName, folder))
		if err != nil {
			slog.Error("can't delete folder", err)
			return err
		}
	}

	return nil
}

func (s *S) SelectServices(projectID int, tx *sqlx.Tx) ([]Service, error) {
	services := make([]Service, 0)
	query := `
	SELECT json_extract(dcj, '$."' || key || '"') AS dcj, key as usn, project_id, name, services.id
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

	for id, dbService := range services {
		service, err := s.ReadServiceFromDCJ(services[id])
		if err != nil {
			slog.Error("error read service from dcj", "err", err)
			continue
		}
		rowID := dbService.ID
		services[id] = *service
		services[id].ID = rowID
	}

	return services, nil
}

func (s *S) ReadServiceFromDCJ(service Service) (*Service, error) {
	var sc compose.Container
	err := compose.FromString(service.DCJ, &s)
	if err != nil {
		slog.Error("unable to parse docker compose json string", "err", err)
		return nil, err
	}

	err = compose.FromString(service.DCJ, &sc)
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

	image := strings.Split(service.Image, ":")
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

	volumes := make([]string, len(service.Volumes))
	for i, v := range service.Volumes {
		volumes[i] = strings.Split(v, ":")[1]
	}

	// When no volumes are set, response with empty string
	if len(service.Volumes) == 0 {
		volumes = []string{""}
	}

	port, err := sc.Labels.GetPort()
	if err != nil {
		slog.Error("unable to get port from labels", "err", err)
	}

	return &Service{
		Name:        service.Name,
		Usn:         service.Usn,
		Ports:       service.Ports,
		Command:     service.Command,
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

// UpsertService inserts a new service with its DCJ for a given projectID into the database.
func (s *S) UpsertService(service *Service, upn UPN, projectID int, tx *sqlx.Tx) error {
	if service.Usn == "" {
		return s.SaveService(service, upn, projectID, tx)
	} else {
		if _, err := service.GenerateServiceCompose(upn); err != nil {
			return err
		}

		query := `
		SELECT COALESCE (json_extract(value, '$.volumes'), "[]" ) as volumes
		FROM services, json_each(dcj, '$')
		WHERE  project_id = $1 AND json_extract(dcj, ('$."' || $2 || '"')) IS NOT NULL;
		`
		var dbVolumes string
		if err := tx.Get(&dbVolumes, query, projectID, service.Usn); err != nil {
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
		_, err := tx.Exec(query, projectID, service.Name, service.DCJ, service.Usn)
		if err != nil {
			slog.Error("Error", "error updating services", err)
			return err
		}

		newVolumesMap := make(map[string]bool)
		for _, vol := range service.Volumes {
			newVolumesMap["./"+path.Join(service.getServicePath(), vol)] = true
		}

		for _, origVolume := range volumes {
			vPath := strings.Split(origVolume, ":")[0]
			if _, exists := newVolumesMap[vPath]; !exists {
				err := utils.DeleteFolder(path.Join(upn.GetProjectPath(), vPath))
				if err != nil {
					slog.Error("can't delete folder", err)
					return err
				}
			}
		}
	}
	return nil
}

func SearchNotInElementsDependsOn(usns []string, projectID int, tx *sqlx.Tx) (bool, error) {
	query := `
	WITH dependants AS (
		SELECT json_extract(value, '$.depends_on') as d, obj.key as child
		FROM services,
			json_each(json_extract(dcj, '$')) as obj
		WHERE project_id = $1
		ORDER BY project_id
	)
	SELECT
		1
	FROM
		dependants
	CROSS JOIN
		json_each(dependants.d)
	WHERE key NOT IN (SELECT value FROM json_each($2))
	LIMIT 1
    `
	usnJSON, err := json.Marshal(usns)
	if err != nil {
		return false, err
	}

	hasDependants := make([]int, 0)
	if err := tx.Select(&hasDependants, query, projectID, usnJSON); err != nil {
		slog.Error("Error", "", err)
		return false, err
	}

	if len(hasDependants) > 0 && hasDependants[0] > 0 {
		return false, nil
	}

	return true, nil
}

func (s *S) DependsOnExists(service *Service, projectID int, tx *sqlx.Tx) bool {
	query := `
	SELECT coalesce(sum(1),0)
	FROM services,
		 json_each(json_extract(dcj, '$'))
	WHERE project_id = $1 AND key IN (SELECT value FROM json_each($2))
	ORDER BY project_id DESC
    `

	parentsUsnJSON := make([]string, len(service.Depends))
	for i := range service.Depends {
		parentsUsnJSON = append(parentsUsnJSON, i)
	}

	parents, err := json.Marshal(parentsUsnJSON)
	if err != nil {
		slog.Error("Error", "cant marshal parent's usn's", err)
		return false
	}

	var hasParents int
	if err := tx.Get(&hasParents, query, projectID, parents); err != nil || hasParents != len(service.Depends) {
		return false
	}

	return true
}

// SaveService inserts a new service with its DCJ for a given projectID into the database.
func (s *S) SaveService(service *Service, upn UPN, projectID int, tx *sqlx.Tx) error {
	if service.Usn != "" {
		return fmt.Errorf("service already have an USN - update the service")
	}

	if !s.DependsOnExists(service, projectID, tx) {
		return fmt.Errorf("depends on service does not exis")
	}

	service.Usn = utils.GenerateRandomName()
	query := `INSERT INTO services (name, project_id, dcj)	VALUES ($1, $2, $3)`
	if _, err := service.GenerateServiceCompose(upn); err != nil {
		return err
	}

	_, err := tx.Exec(query, service.Name, projectID, service.DCJ)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getServicePath() string {
	return fmt.Sprintf("./%s/%s", config.PersistentVolumeDirectoryName, sanitizeName(s.Usn))
}

func (s *Service) GenerateServiceCompose(upn UPN) (*compose.Container, error) {
	sanitizedServiceName := sanitizeName(s.Usn)
	c := &compose.Container{
		Image:    fmt.Sprintf("%s:%s", s.Image, s.ImageTag),
		Restart:  "always",
		Networks: []string{"web", "default"},
		Ports:    s.Ports,
	}

	if s.Depends != nil {
		c.Depends = s.Depends
	}

	if s.HealthCheck != nil {
		c.HealthCheck = s.HealthCheck
	}

	if s.Command != "" {
		c.Command = s.Command
	}

	if s.Deploy != nil {
		c.Deploy = s.Deploy
	}

	if c.Deploy == nil {
		c.Deploy = new(compose.Deploy)
	}

	if c.Deploy.Resources == nil {
		c.Deploy.Resources = new(compose.Resources)
	}

	c.Deploy.Resources.Limits = &config.DockerContainerLimits
	c.Deploy.Replicas = &config.DockerContainerReplicas

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
