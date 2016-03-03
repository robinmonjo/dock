package procfs

import (
	"os"
	"strconv"
)

// DefaultMountpoint define the default mount point of the proc file system
const DefaultMountpoint = "/proc"

// MountPoint is the path of the proc file system mount point.
// Default to DefaultMountPoint
var Mountpoint = DefaultMountpoint

// CountRunningProces return the number of running processes or an error if any
func CountRunningProcs() (int, error) {
	cpt := 0
	err := WalkProcs(func(process *Proc) (bool, error) {
		cpt++
		return true, nil
	})
	return cpt, err
}

// WalkFunc WalkFunc is the type of the function called for each process visited by WalkProcs.
// The process argument contains the current process. If the function return false or an error,
// the WalkProcs func stop, and returns the eventual error
type WalkFunc func(process *Proc) (bool, error)

// WalkProcs walks all the processes and call walk on each process
func WalkProcs(walk WalkFunc) error {
	d, err := os.Open(Mountpoint)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		if pid, err := strconv.Atoi(name); err == nil {
			loop, err := walk(&Proc{
				Pid: pid,
			})
			if err != nil {
				return err
			}
			if !loop {
				return nil
			}
		}
	}
	return nil
}
