package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// check wether the given port is bound by one of the given pids
func IsPortBound(port string, pids []int) (bool, error) {
	for _, pid := range pids {
		binderPid, err := portBinder(port)
		if err != nil {
			return false, err
		}

		if binderPid == -1 {
			continue
		}

		if binderPid == pid {
			return true, nil
		}
	}
	return false, nil
}

// returns the pid of the process binding the given port.
// returns -1 if port is not bound or an error
func portBinder(port string) (int, error) {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf(":%s", port)).Output()
	if err != nil {
		if exitStatus(err) == 1 {
			//not binded
			return -1, nil
		}
		return -1, err
	}
	pidStr := strings.TrimSuffix(string(out), "\n")
	return strconv.Atoi(pidStr)
}

func exitStatus(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 555
}
