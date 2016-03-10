package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/robinmonjo/dock/procfs"
)

// resources: - https://github.com/opencontainers/runc/blob/master/signals.go
//            - https://github.com/Yelp/dumb-init/blob/master/dumb-init.c
//            - https://github.com/phusion/baseimage-docker/blob/master/image/bin/my_init

const (
	signalBufferSize = 2048
	killTimeout      = 5
)

type signalsHandler struct {
	signals   chan os.Signal
	authority bool
}

func newSignalsHandler() *signalsHandler {
	s := make(chan os.Signal, signalBufferSize)
	signal.Notify(s)

	return &signalsHandler{
		signals: s,
	}
}

func (h *signalsHandler) forward(p *process) int {

	pid1 := p.pid()

	for s := range h.signals {
		log.Debugf("signal: %q", s)

		switch s {
		case syscall.SIGWINCH:
			p.resizePty()

		case syscall.SIGCHLD:
			//child process died, dock will exit
			//sending sigterm to every remaining processes before calling wait4
			if err := signalAllDescendants(syscall.SIGTERM); err != nil {
				log.Debugf("failed to send sigterm signal: %v", err)
			}

			go func() {
				<-time.After(killTimeout * time.Second)
				log.Debugf("kill timed out")
				if err := signalAllDescendants(syscall.SIGKILL); err != nil {
					log.Debugf("failed to send sigkill signal: %v", err)
				}
			}()

			//waiting for all processes to die
			log.Debug("reaping all children")
			exits, err := reap()
			log.Debug("children reaped")
			if err != nil {
				log.Error(err)
			}

			for _, e := range exits {
				if e.pid == pid1 {
					p.wait()
					return e.status
				}
			}

		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			fallthrough
		case syscall.SIGQUIT:
			//stopping signals
			sigToForward := s

			if h.authority {
				blocked, err := isSignalBlocked(pid1, s)
				if err != nil {
					log.Error(err)
					goto forward
				}
				ignored, err := isSignalIgnored(pid1, s)
				if err != nil {
					log.Error(err)
					goto forward
				}
				if blocked || ignored {
					sigToForward = os.Signal(syscall.SIGKILL)
				}
			}

		forward:
			if err := p.signal(sigToForward); err != nil {
				log.Error(err)
			}

		default:

			//simply forward the signal to the process
			if err := p.signal(s); err != nil {
				log.Error(err)
			}
		}
	}

	panic("-- this line should never been executed --")
}

// exit models a process exit status with the pid and exit status.
type exit struct {
	pid    int
	status int
}

func reap() (exits []exit, err error) {
	var (
		ws  syscall.WaitStatus
		rus syscall.Rusage
	)
	for {
		pid, err := syscall.Wait4(-1, &ws, 0, &rus)
		if err != nil {
			if err == syscall.ECHILD || err == syscall.ESRCH {
				return exits, nil
			}
			return nil, err
		}
		if pid <= 0 {
			return exits, nil
		}
		log.Debugf("process with PID %d died", pid)
		exits = append(exits, exit{
			pid:    pid,
			status: exitStatus(ws),
		})
	}
}

// Send the given signal to every processes except for the PID 1
func signalAllDescendants(sig syscall.Signal) error {
	self := procfs.Self()
	pses, err := self.Descendants()
	if err != nil {
		return err
	}
	for _, ps := range pses {
		err = syscall.Kill(ps.Pid, sig)
	}
	return err
}

// tell if given pid blocks the given signal
func isSignalBlocked(pid int, s os.Signal) (bool, error) {
	status, err := procStatus(pid)
	if err != nil {
		return false, err
	}
	return include(status.SigBlk, s), nil
}

// tell if the given pid ignore the given signal
func isSignalIgnored(pid int, s os.Signal) (bool, error) {
	status, err := procStatus(pid)
	if err != nil {
		return false, err
	}
	return include(status.SigIgn, s), nil
}

func procStatus(pid int) (*procfs.ProcStatus, error) {
	p := &procfs.Proc{
		Pid: pid,
	}
	return p.Status()
}

func include(signals []syscall.Signal, s os.Signal) bool {
	if sig, ok := s.(syscall.Signal); ok {
		for _, s := range signals {
			if s == sig {
				return true
			}
		}
		return false
	}
	return false
}
