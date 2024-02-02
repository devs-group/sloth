package pkg

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/goombaio/namegenerator"

	"github.com/devs-group/sloth/config"
	"github.com/devs-group/sloth/pkg/compose"
	"github.com/devs-group/sloth/pkg/docker"
	"github.com/devs-group/sloth/repository"
)

type ContainerState struct {
	State  string `json:"state"`
	Status string `json:"status"`
}

const PersistentVolumeDirectoryName = "data"
const DockerComposeFileName = "docker-compose.yml"
const DockerConfigFileName = "config.json"

func PrepareProject(p *repository.Project, upn repository.UPN) error {
	if err := CreateProjectDirectories(p, upn); err != nil {
		return err
	}

	dc, err := generateDockerCompose(p, upn)
	if err != nil {
		return err
	}

	return SaveDockerComposeFile(upn, *dc)
}

func CreateProjectDirectories(p *repository.Project, upn repository.UPN) error {
	if HasVolumesInRequest(*p) {
		if _, err := CreateFolderIfNotExists(path.Join(upn.GetProjectPath(), PersistentVolumeDirectoryName)); err != nil {
			return err
		}
	}
	return nil
}

func generateDockerCompose(p *repository.Project, upn repository.UPN) (*compose.DockerCompose, error) {
	dc := GenerateDockerCompose(p, upn)

	dcj, err := dc.ToJSONString()
	if err != nil {
		return nil, err
	}

	p.DCJ = dcj

	return &dc, nil
}

func SaveDockerComposeFile(upn repository.UPN, dc compose.DockerCompose) error {
	dcy, err := dc.ToYAML()
	if err != nil {
		return err
	}
	return CreateDockerComposeFile(upn, dcy)
}

func DeleteBackupFiles(upn repository.UPN) {
	if err := DeleteFile(fmt.Sprintf("%s.tmp", DockerComposeFileName), upn); err != nil {
		slog.Error("unable to delete temp docker-compose file", "upn", upn, "err", err)
	}
	if err := DeleteFile(fmt.Sprintf("%s.tmp", DockerConfigFileName), upn); err != nil {
		slog.Error("unable to delete temp docker-compose file", "upn", upn, "err", err)
	}
}

func BackupCurrentFiles(upn repository.UPN) error {
	if err := CreateTempFile(DockerComposeFileName, upn); err != nil {
		return err
	}
	if err := CreateTempFile(DockerConfigFileName, upn); err != nil {
		RollbackFromTempFile(DockerComposeFileName, upn)
		return err
	}
	return nil
}

func RollbackToPreviousState(upn repository.UPN) {
	RollbackFromTempFile(DockerComposeFileName, upn)
	RollbackFromTempFile(DockerConfigFileName, upn)
	StartContainers(upn.GetProjectPath(), nil, nil)
}

func CreateTempFile(filename string, upn repository.UPN) error {
	oldPath := path.Join(filepath.Clean(config.ProjectsDir), string(upn), filename)
	newPath := path.Join(filepath.Clean(config.ProjectsDir), string(upn), fmt.Sprintf("%s.tmp", filename))
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

// rollbackFromTempFile renames filename.tmp file to filename file
func RollbackFromTempFile(filename string, upn repository.UPN) error {
	tmpPath := path.Join(filepath.Clean(config.ProjectsDir), string(upn), fmt.Sprintf("%s.tmp", filename))
	newPath := path.Join(filepath.Clean(config.ProjectsDir), string(upn), filename)
	_, err := os.Stat(tmpPath)
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, newPath)
}

func GetContainersState(upn repository.UPN) (map[string]ContainerState, error) {
	containers, err := docker.GetContainersByDirectory(upn.GetProjectPath())
	if err != nil {
		return nil, err
	}
	state := make(map[string]ContainerState)
	for i := range containers {
		c := containers[i]
		sn := c.Labels["com.docker.compose.service"]
		state[sn] = ContainerState{
			State:  c.State,
			Status: c.Status,
		}
	}
	return state, nil
}

func DeleteFile(filename string, upn repository.UPN) error {
	p := path.Join(filepath.Clean(config.ProjectsDir), string(upn), filename)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Remove(p)
}

func CreateFolderIfNotExists(p string) (string, error) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to crete folder in path %s, err: %v", p, err)
		} else {
			slog.Debug("folder has been created successfully", "path", p)
			return p, nil
		}
	} else if err != nil {
		return "", fmt.Errorf("unable to check if folder exists in path %s, err: %v", p, err)
	} else {
		return p, nil
	}
}

func HasVolumesInRequest(p repository.Project) bool {
	hasVolumes := false
	for _, s := range p.Services {
		if len(s.Volumes) > 0 {
			hasVolumes = true
		}
	}
	return hasVolumes
}

func RenameFolder(oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func DeleteFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

func GenerateRandomName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	return nameGenerator.Generate()
}

func RunDockerLogin(ppath string, credentials []repository.DockerCredential) error {
	if len(credentials) == 0 {
		return nil
	}
	for _, dc := range credentials {
		err := docker.Login(dc.Username, dc.Password, dc.Registry, ppath)
		if err != nil {
			return err
		}
	}
	return nil
}

func StartContainers(ppath string, services compose.Services, credentials []repository.DockerCredential) error {
	err := RunDockerLogin(ppath, credentials)
	if err != nil {
		slog.Error("unable to run docker login", "path", ppath, "err", err)
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(services))
	for _, s := range services {
		wg.Add(1)
		go func(service *compose.Container) {
			defer wg.Done()
			err := docker.Pull(service.Image, ppath)
			if err != nil {
				errCh <- err
			}
		}(s)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return fmt.Errorf("unable to pull containers: %v", errors)
	}

	for _, dc := range credentials {
		err = docker.Logout(ppath, dc.Registry)
		if err != nil {
			return fmt.Errorf("unable to run docker logout for registry %s: %v", dc.Registry, err)
		}
	}

	if err := compose.Up(ppath); err != nil {
		return fmt.Errorf("unable to start containers: %v", err)
	}

	return nil
}

func StopContainers(ppath string) error {
	if err := compose.Down(ppath); err != nil {
		return fmt.Errorf("unable to shut down containers: %v", err)
	}
	return nil
}

func RestartContainers(ppath string, services compose.Services, credentials []repository.DockerCredential) error {
	if err := StopContainers(ppath); err != nil {
		return err
	}

	if err := StartContainers(ppath, services, credentials); err != nil {
		return err
	}
	return nil
}

func CreateDockerComposeFile(upn repository.UPN, yaml string) error {
	p := fmt.Sprintf("%s/%s/%s", filepath.Clean(config.ProjectsDir), upn, DockerComposeFileName)
	filePerm := 0600
	err := os.WriteFile(p, []byte(yaml), os.FileMode(filePerm))
	if err != nil {
		return fmt.Errorf("unable to write file %s: err %v", p, err)
	}
	return nil
}

func GenerateDockerCompose(p *repository.Project, upn repository.UPN) compose.DockerCompose {
	services := make(map[string]*compose.Container)
	for _, s := range p.Services {
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
				c.Volumes = append(c.Volumes, fmt.Sprintf("./%s/%s/%s:%s", PersistentVolumeDirectoryName, sanitizedServiceName, dataPath, v))
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

		services[sanitizedServiceName] = c
	}

	// External networks refer to pre-existing networks on the host machine.
	// In a production environment, this network is typically established during Traefik setup.
	// However, in development environments, this network may not be present by default.
	isWebExternalNetwork := true
	if config.Environment == config.Development {
		isWebExternalNetwork = false
	}

	dc := compose.DockerCompose{
		Version: "3.9",
		Networks: map[string]*compose.Network{
			"web": {
				External: isWebExternalNetwork,
			},
			"default": {
				Driver:   "bridge",
				External: false,
			},
		},
		Services: services,
	}
	return dc
}

func RandStringRunes(n int) (string, error) {
	var runes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(runes))))
		if err != nil {
			return "", err
		}
		b[i] = runes[n.Int64()]
	}
	return string(b), nil
}

func sanitizeName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
