package main

import (
	"os/exec"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/robinmonjo/dock/iowire"
	"github.com/robinmonjo/dock/procfs"
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

func exitStatusFromError(err error) int {
	if msg, ok := err.(*exec.ExitError); ok {
		return msg.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return -1
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

// prefix args have the following format: --prefix some-prefix[:blue]
func parsePrefixArg(prefix string) (string, iowire.Color) {
	comps := strings.Split(prefix, ":")
	if len(comps) == 1 {
		return comps[0], iowire.NoColor
	}
	return comps[0], iowire.MapColor(comps[len(comps)-1])
}
