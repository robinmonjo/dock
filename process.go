package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/docker/docker/pkg/term"
	"github.com/kr/pty"
	"github.com/robinmonjo/dock/iowire"
)

type process struct {
	argv      []string //argv[0] must be the path
	cmd       *exec.Cmd
	wire      *iowire.Wire
	pty       *os.File
	termState *termState
}

type termState struct {
	state *term.State
	fd    uintptr //file descriptor associated with the state
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

	if p.wire.Interactive() {
		go func() {
			<-p.wire.CloseCh
			//if interactive and stream closed, send a sigterm to the process
			p.signal(syscall.SIGTERM)
		}()
		return p.startInteractive()
	} else {
		return p.startNonInteractive()
	}
}

func (p *process) startNonInteractive() error {
	p.cmd.Stdin = p.wire
	p.cmd.Stdout = p.wire
	p.cmd.Stderr = p.wire

	return p.cmd.Start()
}

func (p *process) startInteractive() error {
	f, err := pty.Start(p.cmd)
	if err != nil {
		return err
	}
	p.pty = f

	if p.wire.Input == os.Stdin {
		// the current terminal shall pass everything to the console, make it ignores ctrl+C etc ...
		// this is done by making the terminal raw. The state is saved to reset user's terminal settings
		// when dock exits
		state, err := term.SetRawTerminal(os.Stdin.Fd())
		if err != nil {
			return err
		}
		p.termState = &termState{
			state: state,
			fd:    os.Stdin.Fd(),
		}
	} else {
		// wire.Input is a socket (tcp, tls ...). Obvioulsy, we can't set the remote user's terminal in raw mode, however we can at least
		// disable echo on the console
		state, err := term.SaveState(p.pty.Fd())
		if err != nil {
			return err
		}
		if err := term.DisableEcho(p.pty.Fd(), state); err != nil {
			return err
		}
		p.termState = &termState{
			state: state,
			fd:    p.pty.Fd(),
		}
	}

	p.resizePty()
	go io.Copy(p.wire, f)
	go io.Copy(f, p.wire)
	return nil
}

func (p *process) wait() error {
	return p.cmd.Wait()
}

func (p *process) cleanup() {
	p.wire.Close()
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
	return term.SetWinsize(p.pty.Fd(), ws)
}
