package compose

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os/exec"

	"github.com/devs-group/sloth/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func Pull(ppath string) error {
	out, err := cmd(ppath, "pull")
	if err != nil {
		return err
	}
	go func() {
		for o := range out {
			slog.Debug(o)
		}
	}()
	return nil
}

func Down(ppath string) error {
	out, err := cmd(ppath, "down", "--remove-orphans")
	if err != nil {
		return err
	}
	go func() {
		for o := range out {
			slog.Debug(o)
		}
	}()
	return nil
}

func Up(ppath string) error {
	out, err := cmd(ppath, "up", "-d")
	if err != nil {
		return err
	}
	go func() {
		for o := range out {
			slog.Debug(o)
		}
	}()
	return nil
}

func Logs(ppath, service string, ch chan string) error {
	cmd := exec.Command("docker-compose", "logs", "-f", service)
	cmd.Dir = ppath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		defer close(ch)
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Error reading log:", err)
				break
			}
			ch <- line
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func Exec(upn, service string, in chan []byte, out chan []byte) error {
	// Connect to the Docker daemon
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()

	containerID, err := docker.GetContainerIDByService(upn, service)
	if err != nil {
		return err
	}

	// Create a container exec instance
	resp, err := cli.ContainerExecCreate(context.Background(), containerID, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh"}, // or any other shell command you want to run
	})
	if err != nil {
		return err
	}

	hijackResp, err := cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{
		Tty: true,
	})
	if err != nil {
		return err
	}

	go func() {
		for i := range in {
			_, err = hijackResp.Conn.Write(i)
			if err != nil {
				log.Println("Error writing to shell:", err)
				break
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := hijackResp.Reader.Read(buf)
		if err != nil {
			log.Println("Error reading from shell:", err)
			break
		}
		out <- buf[:n]
	}

	return nil
}

func cmd(ppath string, arg ...string) (<-chan string, error) {
	messages := make(chan string)
	cmd := exec.Command("docker-compose", arg...)
	cmd.Dir = ppath
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanWords)
	go func(s *bufio.Scanner) {
		for s.Scan() {
			fmt.Println(s.Text())
			messages <- s.Text()
		}
	}(scanner)
	return messages, cmd.Wait()
}
