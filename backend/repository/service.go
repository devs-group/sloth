package repository

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/database"
	"github.com/devs-group/sloth/backend/pkg/compose"
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
	Name     string     `json:"name" binding:"required"`
	Ports    []string   `json:"ports" binding:"gt=0"`
	Image    string     `json:"image" binding:"required"`
	ImageTag string     `json:"image_tag" binding:"required"`
	Command  string     `json:"command"`
	Public   Public     `json:"public"`
	EnvVars  [][]string `json:"env_vars"`
	Volumes  []string   `json:"volumes" binding:"dive,dirpath"`

	ProjetID int                `json:"-" db:"project_id"`
	DCJ      string             `json:"-" db:"dcj"`
	ctn      *compose.Container `json:"-"`
}

func SelectServices(projectID int, store *database.Store) ([]Service, error) {
	services := make([]Service, 0)

	query := `
        SELECT id, unique_name, access_token, name, user_id, path
        FROM projects
        WHERE projectID = $1
    `

	err := store.DB.Get(services, query, projectID)
	if err != nil {
		return nil, err
	}

	return services, nil
}

func ReadServicesFromDCJ(dcj string) ([]Service, error) {
	dc, err := compose.FromString(dcj)
	if err != nil {
		slog.Error("unable to parse docker compose json string", "err", err)
		return nil, err
	}
	services := make([]Service, len(dc.Services))
	idx := 0
	for k, s := range dc.Services {
		hosts, err := s.Labels.GetHosts()
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

		envVars := make([][]string, len(s.Environment))
		for i, e := range s.Environment {
			kv := strings.Split(e, "=")
			envVars[i] = kv
		}

		// When no env vars are set, response with empty tuple
		if len(s.Environment) == 0 {
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

		port, err := s.Labels.GetPort()
		if err != nil {
			slog.Error("unable to get port from labels", "err", err)
		}

		services[idx] = Service{
			Name:     k,
			Ports:    s.Ports,
			Command:  s.Command,
			Image:    image[0],
			ImageTag: image[1],
			EnvVars:  envVars,
			Volumes:  volumes,
			Public: Public{
				Enabled:  s.Labels.IsPublic(),
				Hosts:    hosts,
				Port:     port,
				SSL:      s.Labels.IsSSL(),
				Compress: s.Labels.IsCompress(),
			},
		}
		idx++
	}
	return services, nil
}

// SaveServiceDCJ inserts a new service with its DCJ for a given projectID into the database.
func (s *Service) SaveService(upn UPN, projectID int, tx *sqlx.Tx) error {
	query := `INSERT INTO services (unique_name, project_id, dcj)	VALUES ($1, $2, $3)`

	if err := s.GenerateServiceCompose(upn, projectID); err != nil {
		return err
	}

	_, err := tx.Exec(query, "ddtstd", projectID, s.DCJ)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GenerateServiceCompose(upn UPN, projectID int) error {
	sanitizedServiceName := sanitizeName(s.Name)
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
			c.Volumes = append(c.Volumes, fmt.Sprintf("./%s/%s/%s:%s", config.PersistentVolumeDirectoryName, sanitizedServiceName, dataPath, v))
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

	s.ctn = c

	ctn, err := json.Marshal(s.ctn)
	if err != nil {
		return err
	}

	s.DCJ = string(ctn)
	return nil
}

func sanitizeName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
