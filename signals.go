package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/robinmonjo/dock/system"
)

const (
	signalBufferSize    = 2048
	childrenTermTimeout = 2 //seconds
	childrenKillTimeout = 5 //seconds
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

func (l *signalsHandler) forward(p *process) int {

	cpid := p.pid()

	for s := range l.signals {
		log.Debug(s)

		switch s {
		case syscall.SIGWINCH:
			p.resizePty()

		case syscall.SIGCHLD:
			done := make(chan bool, 1)

			if os.Getpid() == 1 { //if i'am the init process (supposed to be but I crashed my mac twice :) )
				go func() {
					timeouts := []time.Duration{childrenTermTimeout, childrenKillTimeout}
					sigs := []syscall.Signal{syscall.SIGTERM, syscall.SIGKILL}
					for i := 0; i < 2; i++ {
						select {
						case <-time.After(timeouts[i] * time.Second):
							syscall.Kill(-1, sigs[i])
						case <-done:
							return
						}
					}
				}()
			}

			exits, err := l.reap()
			done <- true
			if err != nil {
				log.Error(err)
			}
			for _, e := range exits {
				if e.pid == cpid {
					p.wait()
					return e.status
				}
			}

		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			ps, err := system.NewProcStatus(cpid)
			if err != nil {
				log.Error(err)
			}

			if !ps.SignalAsEffect(s.(syscall.Signal)) {
				//you won't ignore or block my interupts, I'm your boss
				if err := p.signal(syscall.SIGKILL); err != nil {
					log.Error(err)
				}
			} else {
				if err := p.signal(s); err != nil {
					log.Error(err)
				}
			}

		default:
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

// this may block if child processes doesn't respond to their parent death
func (l *signalsHandler) reap() ([]exit, error) {
	var (
		exits []exit
		ws    syscall.WaitStatus
		rus   syscall.Rusage
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
		exits = append(exits, exit{
			pid:    pid,
			status: exitStatus(ws),
		})
	}
}
