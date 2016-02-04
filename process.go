package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/kr/pty"
	"github.com/robinmonjo/dock/term"
)

type process struct {
	argv      []string //argv[0] must be the path
	cmd       *exec.Cmd
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	pty       *os.File
	termState *termState
}

type termState struct {
	state *term.State
	fd    uintptr //file descriptor associated with the state
}

func (p *process) beforeStart() error {
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

	p.cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

	return nil
}

func (p *process) start() error {
	if err := p.beforeStart(); err != nil {
		return err
	}

	p.cmd.Stdin = p.stdin
	p.cmd.Stdout = p.stdout
	p.cmd.Stderr = p.stderr

	return p.cmd.Start()
}

func (p *process) startInteractive() error {
	if err := p.beforeStart(); err != nil {
		return err
	}

	f, err := pty.Start(p.cmd)
	if err != nil {
		return err
	}
	p.pty = f

	//if stdin, set raw on stdin
	//if network, disable echo on p.pty

	state, err := term.SetRawTerminal(os.Stdin.Fd())
	if err != nil {
		return err
	}
	p.termState = &termState{
		state: state,
		fd:    os.Stdin.Fd(),
	}
	p.resizePty()
	go io.Copy(p.stdout, f)
	go io.Copy(p.stderr, f)
	go io.Copy(f, p.stdin)
	return nil
}

func (p *process) wait() error {
	return p.cmd.Wait()
}

func (p *process) cleanup() {
	if p.pty != nil {
		p.pty.Close()
	}
	if p.termState != nil {
		term.RestoreTerminal(p.termState.fd, p.termState.state)
	}
}

func (p *process) pid() int {
	return p.cmd.Process.Pid
}

func (p *process) signal(sig os.Signal) error {
	return p.cmd.Process.Signal(sig)
}

func (p *process) resizePty() error {
	if p.pty == nil {
		return nil
	}
	ws, err := term.GetWinsize(os.Stdin.Fd())
	if err != nil {
		return err
	}
	if err := term.SetWinsize(p.pty.Fd(), ws); err != nil {
		return err
	}
	return nil
}
