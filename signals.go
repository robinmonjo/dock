package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	//"github.com/robinmonjo/dock/system"
)

const (
	signalBufferSize       = 2048
	childrenSigtermTimeout = 2  //seconds
	childrenKillTimeout    = 10 //seconds
)

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
	//procStatus, err := system.NewProcStatus(cpid)
	//if err != nil {
	//	log.Error(err)
	//}

	for s := range l.signals {
		log.Debug(s)
		switch s {
		case syscall.SIGCHLD:
			//child exited, sending a sigterm to all pses and collecting waits TODO: kill timeout
			if os.Getpid() == 1 { //if i'am the init process (supposed to be but I crashed my mac twice :) )
				go func() {
					<-time.After(childrenSigtermTimeout * time.Second)
					syscall.Kill(-1, syscall.SIGTERM)
				}()
			}

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
			//	log.Infof("signal is blocked %v", procStatus.SignalBlocked(s.(syscall.Signal)))
			//	log.Infof("signal is inored %v", procStatus.SignalIgnored(s.(syscall.Signal)))
			//	log.Infof("signal is caught %v", procStatus.SignalCaught(s.(syscall.Signal)))
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
