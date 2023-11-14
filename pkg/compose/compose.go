package compose

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/devs-group/sloth/pkg/docker"
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

type channelWriter struct {
	ch chan<- []byte
}

func (cw channelWriter) Write(p []byte) (n int, err error) {
	copied := make([]byte, len(p))
	copy(copied, p)
	cw.ch <- copied
	return len(p), nil
}

func Exec(ctx context.Context, upn, service string, in, out chan []byte) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()

	containerID, err := docker.GetContainerIDByService(upn, service)
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
