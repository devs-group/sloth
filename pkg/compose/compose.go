package compose

import (
	"bufio"
	"fmt"
	"os/exec"
)

func Pull(ppath string) error {
	return cmd(ppath, "pull")
}

func Down(ppath string) error {
	return cmd(ppath, "down")
}

func Up(ppath string) error {
	return cmd(ppath, "up", "-d")
}

func cmd(ppath string, arg ...string) error {
	cmd := exec.Command("docker-compose", arg...)
	cmd.Dir = ppath
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	cmd.Wait()
	return nil
}
