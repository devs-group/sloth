package compose

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"sync"

	"github.com/devs-group/sloth/backend/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
)

func Down(pPath string) error {
	command := []string{"down", "--remove-orphans"}
	return ExecuteDockerComposeCommand(pPath, command...)
}

func Up(pPath string) error {
	command := []string{"up", "-d"}
	return ExecuteDockerComposeCommand(pPath, command...)
}

func ExecuteDockerComposeCommand(pPath string, command ...string) error {
	messages, errChan, err := cmd(pPath, command...)
	if err != nil {
		slog.Error("error", "error starting docker-compose", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var errorList []error
	var mutex sync.Mutex

	go func() {
		defer wg.Done()
		for msg := range messages {
			if strings.Contains(msg, "Error response from daemon") {
				mutex.Lock()
				errorList = append(errorList, fmt.Errorf("%v", msg))
				mutex.Unlock()
			}
			slog.Info("info", "docker-compose message", msg)
		}
	}()

	go func() {
		defer wg.Done()
		var msgs error
		for err := range errChan {
			if err != nil {
				if msgs == nil {
					msgs = err
				} else {
					msgs = errors.Wrap(err, msgs.Error())
				}
				slog.Error("error", "error from docker-compose", err)
			}
		}
		if msgs != nil {
			mutex.Lock()
			errorList = append(errorList, msgs)
			mutex.Unlock()
		}
	}()

	wg.Wait()

	if len(errorList) > 0 {
		var combinedError error
		for _, err := range errorList {
			if combinedError == nil {
				combinedError = err
			} else {
				combinedError = errors.Wrap(err, combinedError.Error())
			}
		}
		return combinedError
	}

	return nil
}

func Logs(pPath, service string, ch chan string) error {
	cmd := exec.Command("docker-compose", "logs", "-f", service)
	cmd.Dir = pPath

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
				slog.Error("error", "error reading log", err)
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

func Shell(ctx context.Context, ppath string, project string, service string, in, out chan []byte) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()
	defer func() {
		if err != nil {
			slog.Error("error writing to websocket:", "err", err)
		}
	}()

	containerID, err := docker.GetContainerIDByService(project, service)
	if err != nil {
		return err
	}

	resp, err := cli.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh"},
	})
	if err != nil {
		return err
	}

	hijackResp, err := cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{Detach: false, Tty: false})
	if err != nil {
		return err
	}

	stdoutWriter := channelWriter{ch: out}
	stderrWriter := channelWriter{ch: out}
	go func() {
		for {
			select {
			case <-ctx.Done():
				hijackResp.Close()
				return
			case data := <-in:
				if _, err := hijackResp.Conn.Write(data); err != nil {
					slog.Error("Error writing to exec stdin:", err)
					return
				} else {
					slog.Info("Executing Command from websocket", "cmd", string(data))
				}
			}
		}
	}()

	if _, err := stdcopy.StdCopy(stdoutWriter, stderrWriter, hijackResp.Reader); err != nil && err != io.EOF {
		slog.Error("Error reading from exec stdout/stderr:", err)
		return err
	} else {
		slog.Info("Executed")
	}

	return nil
}

func cmd(pPath string, args ...string) (<-chan string, <-chan error, error) {
	messages := make(chan string)
	errorChan := make(chan error, 1)

	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = pPath

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)

	go func() {
		defer close(messages)
		for scanner.Scan() {
			messages <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			errorChan <- err
		}
		errorChan <- cmd.Wait()
		close(errorChan)
	}()

	return messages, errorChan, nil
}
