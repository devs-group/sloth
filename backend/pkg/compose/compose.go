package compose

import (
	"bufio"
	"io"
	"log/slog"
	"os/exec"
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
				slog.Error("Error reading log", "error:", err)
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
			messages <- s.Text()
		}
	}(scanner)
	return messages, cmd.Wait()
}
