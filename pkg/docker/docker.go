package docker

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func GetContainersByDirectory(dir string) ([]types.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	cntnrs := make([]types.Container, 0)
	for i := range containers {
		workDir, ok := containers[i].Labels["com.docker.compose.project.working_dir"]
		if !ok {
			continue
		}
		if !strings.HasSuffix(workDir, dir) {
			continue
		}
		cntnrs = append(cntnrs, containers[i])
	}
	return cntnrs, nil
}

func Login(username, password, registryURL, path string) error {
	cmd := exec.Command("docker", "--config", "./", "login", registryURL, "-u", username, "-p", password)
	cmd.Dir = path

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func Pull(image, path string) error {
	cmd := exec.Command("docker", "--config", "./", "pull", image)
	cmd.Dir = path

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func Logout(path, registry string) error {
	cmd := exec.Command("docker", "--config", "./", "logout", registry)
	cmd.Dir = path

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func GetContainerIDByService(upn, service string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer cli.Close()
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return "", err
	}

	for _, container := range containers {

		projectName, ok := container.Labels["com.docker.compose.project"]
		if !ok {
			continue
		}
		serviceName, ok := container.Labels["com.docker.compose.service"]
		if !ok {
			continue
		}
		if projectName == upn && serviceName == service {
			return container.ID, nil
		}

	}

	return "", errors.New(fmt.Sprintf("unable to find container with for service: %s in project: %s", service, upn))
}
