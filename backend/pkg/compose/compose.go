package compose

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"sync"

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
