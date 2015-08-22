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
		log.Infof("signal: %v", s)

		switch s {
		case syscall.SIGCHLD:
			log.Info("child exit")

			var (
				exit int
				ws   syscall.WaitStatus
				rus  syscall.Rusage
			)
			log.Info("Waiting for children to finish")
			for {
				runPsef()
				pid, err := syscall.Wait4(-1, &ws, 0, &rus)
				log.Infof("wait4 on pid: %d, err: %v", pid, err)
				if err != nil {
					if err == syscall.ECHILD {
						//no more child
						log.Info("ECHILD")
						break
					}
					log.Error(err)
					break
				}
				if pid <= 0 {
					break
				}
				if pid == cpid {
					exit = exitStatus(ws)
					p.wait() //just to make sure go cleanup everything
				}
				runPsef()
			}
			log.Info("all clear")
			return exit
		default:
			if err := p.signal(s); err != nil {
				log.Errorf("error forwarding: %v", err)
			}
		}
	}

	panic("-- this line should never been executed --")
}
