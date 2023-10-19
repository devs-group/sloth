package docker

import (
	"context"
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
