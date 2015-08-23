package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// count the number of process currently running
func countRunningPses() (int, error) {
	cpt := 0
	err := filepath.Walk("/proc", func(path string, fi os.FileInfo, err error) error {
		if path == "/proc" {
			return nil
		}

		if filepath.Dir(path) != "/proc" {
			return nil
		}
		if _, err := strconv.Atoi(fi.Name()); err == nil {
			cpt++
		}
		return nil
	})
	return cpt, err
}
