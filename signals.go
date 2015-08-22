package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

const signalBufferSize = 2048

type signalsListener struct {
	signals chan os.Signal
}

func newSignalsListener() *signalsListener {
	s := make(chan os.Signal, signalBufferSize)
	signal.Notify(s)

	return &signalsListener{
		signals: s,
	}
}

func (l *signalsListener) forward(p *process) int {

	cpid := p.pid()

	for s := range l.signals {
		log.Info(s)
		switch s {
		case syscall.SIGCHLD:
			//child exiteld, sending a sigterm to all pses and collecting waits TODO: kill timeout
			syscall.Kill(-1, syscall.SIGTERM)
			exits, err := l.reap()
			if err != nil {
				log.Error(err)
			}
			for _, e := range exits {
				if e.pid == cpid {
					p.wait()
					return e.status
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
func (l *signalsListener) reap() ([]exit, error) {
	var (
		exits []exit
		ws    syscall.WaitStatus
		rus   syscall.Rusage
	)
	for {
		runPsef()
		pid, err := syscall.Wait4(-1, &ws, 0, &rus)
		log.Infof("wait 4 on PID: %d", pid)
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
		runPsef()
	}
}
