package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

// resources: - https://github.com/opencontainers/runc/blob/master/signals.go
//            - https://github.com/Yelp/dumb-init/blob/master/dumb-init.c
//            - https://github.com/phusion/baseimage-docker/blob/master/image/bin/my_init

const (
	signalBufferSize = 2048
	killTimeout      = 5
)

type signalsHandler struct {
	signals chan os.Signal
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
			//child process died, the container will exits
			//sending sigterm to every remaining processes before calling wait4
			if err := signalAllExceptPid1(syscall.SIGTERM); err != nil {
				log.Debugf("failed to send sigterm signal: %v", err)
			}

			go func() {
				<-time.After(killTimeout * time.Second)
				log.Debugf("kill timed out")
				if err := signalAllExceptPid1(syscall.SIGKILL); err != nil {
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
					//p.wait() //should work and be cleaner but sometimes it hangs because some pipes are stil open. Need to test with future version of go
					return e.status
				}
			}

		default:

			if sig, ok := s.(syscall.Signal); ok {
				logHowSignalIsHandled(p.pid(), sig)
			}

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
