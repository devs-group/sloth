package compose

import (
	"log/slog"
	"regexp"
	"strings"
)

type Services map[string]*Container

type DockerCompose struct {
	Networks map[string]*Network `json:"networks,omitempty"`
	Services Services            `json:"services"`
}

type Condition struct {
	Condition string `json:"condition"`
}

type Network struct {
	External bool    `json:"external"`
	Driver   *string `json:"driver,omitempty"`
}

type HealthCheck struct {
	Test        string `json:"test"`
	Interval    string `json:"interval"`
	Timeout     string `json:"timeout"`
	Retries     int    `json:"retries"`
	StartPeriod string `json:"start_period"`
}

type Labels []string

type RestartPolicy struct {
	Condition   *string `json:"condition,omitempty"`
	Delay       *string `json:"delay,omitempty"`
	MaxAttempts *int    `json:"max_attempts,omitempty"`
	Window      *string `json:"window,omitempty"`
}

type Reservations struct {
	CPUs   *string `json:"cpus,omitempty"`
	Memory *string `json:"memory,omitempty"`
}

type Limits struct {
	CPUs   *string `json:"cpus,omitempty"`
	Memory *string `json:"memory,omitempty"`
	PIDs   *int    `json:"pids,omitempty"`
}

type Resources struct {
	Limits       *Limits       `json:"limits,omitempty"`
	Reservations *Reservations `json:"reservations,omitempty"`
}

type Deploy struct {
	Mode          *string        `json:"mode,omitempty"`
	Replicas      *int           `json:"replicas,omitempty"`
	EndPointMode  *string        `json:"endpoint_mode,omitempty"`
	Resources     *Resources     `json:"resources,omitempty"`
	RestartPolicy *RestartPolicy `json:"restart_policy,omitempty"`
}

type Container struct {
	Build       *Build               `json:"build,omitempty"`
	Command     string               `json:"command,omitempty"`
	CPU         int                  `json:"cpu_shares,omitempty"`
	DNS         []string             `json:"dns,omitempty"`
	DNSSearch   []string             `json:"dns_search,omitempty"`
	Entrypoint  string               `json:"entrypoint,omitempty"`
	EnvFile     []string             `json:"env_file,omitempty"`
	Environment []string             `json:"environment,omitempty"`
	Expose      []int                `json:"expose,omitempty"`
	Hostname    string               `json:"hostname,omitempty"`
	Image       string               `json:"image,omitempty"`
	Labels      Labels               `json:"labels,omitempty"`
	Links       []string             `json:"links,omitempty"`
	Logging     *Logging             `json:"logging,omitempty"`
	Memory      int                  `json:"mem_limit,omitempty"`
	Name        string               `json:"-"`
	Networks    []string             `json:"networks,omitempty"`
	NetworkMode string               `json:"network_mode,omitempty"`
	Pid         string               `json:"pid,omitempty"`
	Ports       []string             `json:"ports,omitempty"`
	Privileged  bool                 `json:"privileged,omitempty"`
	User        string               `json:"user,omitempty"`
	Volumes     []string             `json:"volumes,omitempty"`
	VolumesFrom []string             `json:"volumes_from,omitempty"`
	WorkDir     string               `json:"working_dir,omitempty"`
	Restart     string               `json:"restart"`
	HealthCheck *HealthCheck         `json:"healthcheck,omitempty"`
	Depends     map[string]Condition `json:"depends_on,omitempty"`
	Deploy      *Deploy              `json:"deploy,omitempty"`
}

type Build struct {
	Ctx BuildContext
}

type BuildContext struct {
	Context    string            `json:"context,omitempty"`
	Dockerfile string            `json:"dockerfile,omitempty"`
	Args       map[string]string `json:"args,omitempty"`
}

type Logging struct {
	Driver  string
	Options map[string]string
}

func (l Labels) IsPublic() bool {
	for _, label := range l {
		return strings.EqualFold(label, "traefik.enable=true")
	}
	return false
}

func (l Labels) IsSSL() bool {
	for _, label := range l {
		if strings.HasSuffix(label, "entrypoints=https") {
			return true
		}
	}
	return false
}

func (l Labels) IsCompress() bool {
	for _, label := range l {
		if strings.HasSuffix(label, "compress=true") {
			return true
		}
	}
	return false
}

func (l Labels) GetHosts() ([]string, error) {
	var hosts []string

	for _, label := range l {
		if strings.Contains(label, "rule=Host") {
			re, err := regexp.Compile(`Host\(` + "`([^`]+)`" + `\)`)
			if err != nil {
				return nil, err
			}
			submatches := re.FindAllStringSubmatch(label, -1)
			slog.Debug("host label", "matches", slog.AnyValue(submatches))
			for _, submatch := range submatches {
				if len(submatch) >= 2 {
					hosts = append(hosts, submatch[1])
				}
			}
		}
	}

	if len(hosts) > 0 {
		return hosts, nil
	}
	return nil, nil
}

func (l Labels) GetPort() (string, error) {
	for _, label := range l {
		if strings.Contains(label, "loadbalancer.server.port") {
			parts := strings.Split(label, "=")
			if len(parts) == 2 {
				return parts[1], nil
			}
		}
	}
	return "", nil
}

type channelWriter struct {
	ch chan<- []byte
}

func (cw channelWriter) Write(p []byte) (n int, err error) {
	copied := make([]byte, len(p))
	copy(copied, p)
	cw.ch <- copied
	return len(p), nil
}
