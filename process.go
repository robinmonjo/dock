package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

const (
	exitSignalOffset = 128
)

type process struct {
	args   []string
	cmd    *exec.Cmd
	stdin  io.Reader
	stdout io.Writer
}

func (p *process) start() error {
	n := len(p.args)
	if n == 0 {
		return fmt.Errorf("args can't be empty")
	}

	arg1 := p.args[0]
	var args []string
	if n > 1 {
		args = p.args[1:n]
	}

	cmd := exec.Command(arg1, args...)
	cmd.Stdin = p.stdin
	cmd.Stdout = p.stdout
	p.cmd = cmd
	return cmd.Start()
}

func (p *process) wait() (int, error) {
	err := p.cmd.Wait()
	status := p.cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitCode := status.ExitStatus()
	if status.Signaled() {
		exitCode += exitSignalOffset
	}
	return exitCode, err
}

func (p *process) pid() int {
	return p.cmd.Process.Pid
}

func (p *process) signal(sig os.Signal) error {
	return p.cmd.Process.Signal(sig)
}
