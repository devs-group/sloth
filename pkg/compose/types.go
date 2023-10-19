package compose

import (
	"regexp"
	"strings"
)

type DockerCompose struct {
	Version  string                `json:"version"` // json tags affect YAML field names too.
	Networks map[string]*Network   `json:"networks,omitempty"`
	Services map[string]*Container `json:"services"`
}

type Network struct {
	External bool   `json:"external"`
	Driver   string `json:"driver"`
}

type Labels []string

type Container struct {
	Build       *Build   `json:"build,omitempty"`
	Command     string   `json:"command,omitempty"`
	CPU         int      `json:"cpu_shares,omitempty"`
	DNS         []string `json:"dns,omitempty"`
	DNSSearch   []string `json:"dns_search,omitempty"`
	Entrypoint  string   `json:"entrypoint,omitempty"`
	EnvFile     []string `json:"env_file,omitempty"`
	Environment []string `json:"environment,omitempty"`
	Expose      []int    `json:"expose,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	Image       string   `json:"image,omitempty"`
	Labels      Labels   `json:"labels,omitempty"`
	Links       []string `json:"links,omitempty"`
	Logging     *Logging `json:"logging,omitempty"`
	Memory      int      `json:"mem_limit,omitempty"`
	Name        string   `json:"-"`
	Networks    []string `json:"networks,omitempty"`
	NetworkMode string   `json:"network_mode,omitempty"`
	Pid         string   `json:"pid,omitempty"`
	Ports       []string `json:"ports,omitempty"`
	Privileged  bool     `json:"privileged,omitempty"`
	User        string   `json:"user,omitempty"`
	Volumes     []string `json:"volumes,omitempty"`
	VolumesFrom []string `json:"volumes_from,omitempty"`
	WorkDir     string   `json:"working_dir,omitempty"`
	Restart     string   `json:"restart"`
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

func (l Labels) GetHost() (string, error) {
	for _, label := range l {
		if strings.Contains(label, "rule=Host") {
			re, err := regexp.Compile(`Host\(` + "`([^`]+)`" + `\)`)
			if err != nil {
				return "", err
			}
			submatch := re.FindStringSubmatch(label)
			if len(submatch) >= 2 {
				return submatch[1], nil
			}
			return "", nil
		}
	}
	return "", nil
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
