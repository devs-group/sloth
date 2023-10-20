package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"log"
	"strings"
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

func Exec(conn *websocket.Conn) {
	// Connect to the Docker daemon
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// Specify the ID or name of the running container you want to interact with
	containerID := "73058f055b26"

	// Create a container exec instance
	resp, err := cli.ContainerExecCreate(context.Background(), containerID, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh"}, // or any other shell command you want to run
	})
	if err != nil {
		panic(err)
	}

	hijackResp, err := cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{
		Tty: true,
	})

	go func() {
		for {
			// _, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}

			_, err = hijackResp.Conn.Write([]byte("touch test_file"))
			_, err = hijackResp.Conn.Write([]byte("echo 'hello'"))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := hijackResp.Reader.Read(buf)
		if err != nil {
			log.Println("read:", err)
			break
		}

		err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
