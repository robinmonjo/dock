package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type process struct {
	argv   []string //args[0] mus be the path
	cmd    *exec.Cmd
	stdin  io.Reader
	stdout io.Writer
}

func (p *process) start() error {
	n := len(p.argv)
	if n == 0 {
		return fmt.Errorf("argv can't be empty")
	}

	path, err := exec.LookPath(p.argv[0])
	if err != nil {
		return err
	}

	var args []string
	if n > 1 {
		args = p.argv[1:n]
	}

	p.cmd = exec.Command(path, args...)
	p.cmd.Stdin = p.stdin
	p.cmd.Stdout = p.stdout
	p.cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

	return p.cmd.Start()
}

func (p *process) wait() error {
	return p.cmd.Wait()
}

func (p *process) pid() int {
	return p.cmd.Process.Pid
}

func (p *process) signal(sig os.Signal) error {
	return p.cmd.Process.Signal(sig)
}
