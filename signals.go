package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

const signalBufferSize = 2048

func newSignalsHandler(p *process) *signalsHandler {
	s := make(chan os.Signal, signalBufferSize)
	signal.Notify(s)

	return &signalsHandler{
		signals: s,
		process: p,
	}
}

type signalsHandler struct {
	process *process
	signals chan os.Signal
}

func (h *signalsHandler) forward() {
	for s := range h.signals {
		log.Infof("signal: %v", s)
		switch s {
		case syscall.SIGCHLD:
			// child exited
			log.Info("child exit")
			return
		default:
			if err := h.process.signal(s); err != nil {
				log.Errorf("error forwarding: %v", err)
			}
		}
	}
}
