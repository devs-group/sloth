package services

import (
	"fmt"
	"github.com/pkg/errors"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/pkg/compose"
	"github.com/devs-group/sloth/backend/pkg/docker"
	"github.com/devs-group/sloth/backend/utils"
)

type UPN string
type ContainerState struct {
	State  string `json:"state"`
	Status string `json:"status"`
}

func (upn *UPN) GetProjectPath() string {
	return path.Join(filepath.Clean(config.ProjectsDir), string(*upn))
}

func (upn *UPN) RunDockerLogin(credentials []DockerCredential) error {
	if len(credentials) == 0 {
		return nil
	}
	for _, dc := range credentials {
		err := docker.Login(dc.Username, dc.Password, dc.Registry, upn.GetProjectPath())
		if err != nil {
			return err
		}
	}
	return nil
}

func (upn *UPN) GetContainersState() (map[string]ContainerState, error) {
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

func (upn *UPN) StartContainers(services compose.Services, credentials []DockerCredential) error {
	err := upn.RunDockerLogin(credentials)
	if err != nil {
		slog.Error("unable to run docker login", "path", upn.GetProjectPath(), "err", err)
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(services))
	for _, s := range services {
		wg.Add(1)
		go func(service *compose.Container) {
			defer wg.Done()
			err := docker.Pull(service.Image, upn.GetProjectPath())
			if err != nil {
				errCh <- err
			}
		}(s)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("unable to pull containers: %v", errs)
	}

	for _, dc := range credentials {
		err = docker.Logout(upn.GetProjectPath(), dc.Registry)
		if err != nil {
			return fmt.Errorf("unable to run docker logout for registry %s: %v", dc.Registry, err)
		}
	}

	if err := compose.Up(upn.GetProjectPath()); err != nil {
		return fmt.Errorf("unable to start containers: %v", err)
	}

	return nil
}

func (upn *UPN) StopContainers() error {
	if err := compose.Down(upn.GetProjectPath()); err != nil {
		return fmt.Errorf("unable to shut down containers: %v", err)
	}
	return nil
}

func (upn *UPN) RestartContainers(services compose.Services, credentials []DockerCredential) error {
	if err := upn.StopContainers(); err != nil {
		return err
	}

	if err := upn.StartContainers(services, credentials); err != nil {
		return err
	}
	return nil
}

func (upn *UPN) DeleteBackupFiles() {
	if err := utils.DeleteFile(fmt.Sprintf("%s.tmp", config.DockerComposeFileName), upn.GetProjectPath()); err != nil {
		slog.Error("unable to delete temp docker-compose file", "upn", upn, "err", err)
	}
	if err := utils.DeleteFile(fmt.Sprintf("%s.tmp", config.DockerConfigFileName), upn.GetProjectPath()); err != nil {
		slog.Error("unable to delete temp docker-compose file", "upn", upn, "err", err)
	}
}

func (upn *UPN) BackupCurrentFiles() error {
	if err := upn.CreateTempFile(config.DockerComposeFileName); err != nil {
		return err
	}
	if err := upn.CreateTempFile(config.DockerConfigFileName); err != nil {
		err2 := upn.RollbackFromTempFile(config.DockerComposeFileName)
		if err2 != nil {
			err = errors.Wrap(err, err2.Error())
		}
		slog.Error("unable to backup current files", err)
		return err
	}
	return nil
}

func (upn *UPN) CreateTempFile(filename string) error {
	oldPath := path.Join(filepath.Clean(config.ProjectsDir), upn.GetProjectPath(), filename)
	newPath := path.Join(filepath.Clean(config.ProjectsDir), upn.GetProjectPath(), fmt.Sprintf("%s.tmp", filename))
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

func (upn *UPN) RollbackToPreviousState() {
	err := upn.RollbackFromTempFile(config.DockerComposeFileName)
	if err != nil {
		slog.Error("unable to rollback docker compose file", err)
	}
	err = upn.RollbackFromTempFile(config.DockerConfigFileName)
	if err != nil {
		slog.Error("unable to rollback docker config file", err)
	}
	err = upn.StartContainers(nil, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to start containers: %v", err))
	}
}

// RollbackFromTempFile renames filename.tmp file to filename file
func (upn *UPN) RollbackFromTempFile(filename string) error {
	tmpPath := path.Join(filepath.Clean(config.ProjectsDir), upn.GetProjectPath(), fmt.Sprintf("%s.tmp", filename))
	newPath := path.Join(filepath.Clean(config.ProjectsDir), upn.GetProjectPath(), filename)
	_, err := os.Stat(tmpPath)
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, newPath)
}

// IsOneContainerRunning checks for at least one container in the project if its running
func (upn *UPN) IsOneContainerRunning() (bool, error) {
	containerStates, err := upn.GetContainersState()
	if err != nil {
		return false, err
	}

	anyRunning := false
	for _, state := range containerStates {
		if state.State == "running" || state.State == "paused" {
			anyRunning = true
			break
		}
	}
	return anyRunning, nil
}
