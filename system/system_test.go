package system

import (
	"testing"
)

func TestCountRunningProcesses(t *testing.T) {
	procfs = "assets/proc"
	c := CountRunningProcesses()
	if c != 2 {
		t.Fatal("expected 2 running processes, got %d", c)
	}
}
