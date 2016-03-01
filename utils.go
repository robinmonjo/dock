package main

import (
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/robinmonjo/dock/iowire"
	"github.com/robinmonjo/procfs"
)

const exitSignalOffset = 128

// ExitStatus returns the correct exit status for a process based on if it
// was signaled or existed cleanly.
func exitStatus(status syscall.WaitStatus) int {
	if status.Signaled() {
		return exitSignalOffset + int(status.Signal())
	}
	return status.ExitStatus()
}

// Print the current process tree
func printProcessTree() {
	procfs.WalkProcs(func(p *procfs.Proc) (bool, error) {
		status, err := p.Status()
		if err != nil {
			log.Printf("%d", p.Pid)
		} else {
			args, err := p.CmdLine()
			if err != nil {
				args = []string{status.Name}
			}
			log.Printf("%d\t%d\t%s\t%s", p.Pid, status.PPid, status.State, strings.Join(args, " "))
		}
		return true, nil
	})
}

// Send the given signal to every processes except for the PID 1
func signalAllExceptPid1(sig syscall.Signal) error {
	return procfs.WalkProcs(func(p *procfs.Proc) (bool, error) {
		return true, syscall.Kill(p.Pid, sig)
	})
}

// prefix args have the following format: --prefix some-prefix[:blue]
func parsePrefixArg(prefix string) (string, iowire.Color) {
	comps := strings.Split(prefix, ":")
	if len(comps) == 1 {
		return comps[0], iowire.NoColor
	}
	return comps[0], iowire.MapColor(comps[len(comps)-1])
}

func logHowSignalIsHandled(pid int, s syscall.Signal) {
	p := &procfs.Proc{
		Pid: pid,
	}
	status, err := p.Status()
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("is %d blocking signal ? %v", pid, include(status.SigBlk, s))
	log.Debugf("is %d ignoring signal ? %v", pid, include(status.SigIgn, s))
	log.Debugf("is %d caughting signal ? %v", pid, include(status.SigCgt, s))
}

func include(signals []syscall.Signal, sig syscall.Signal) bool {
	for _, s := range signals {
		if s == sig {
			return true
		}
	}
	return false
}
