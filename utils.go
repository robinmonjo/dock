package main

import (
	"fmt"
	"os/exec"
	"syscall"
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

//run ps -ef and print the output (used for debugging)
func runPsef() {
	out, _ := exec.Command("/bin/ps", "-ef").Output()
	fmt.Printf("%s\n", out)
}
