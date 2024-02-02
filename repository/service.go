package repository

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/devs-group/sloth/pkg/compose"
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
