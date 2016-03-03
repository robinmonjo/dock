package procfs

import (
	"testing"
)

func TestCountRunningProcs(t *testing.T) {
	Mountpoint = "./assets/proc"
	c, err := CountRunningProcs()
	if err != nil {
		t.Fatal(err)
	}
	if c != 4 {
		t.Fatalf("expected 2 running processes, got %d", c)
	}
}
