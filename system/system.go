package system

import (
	"path/filepath"
	"strconv"
)

const DefaultProcfs = "procfs"

var procfs = DefaultProcfs

//count the number of running processes
func CountRunningProcesses() int {
	files, _ := filepath.Glob(filepath.Join(procfs, "*"))
	cpt := 0
	for _, f := range files {
		if _, err := strconv.Atoi(filepath.Base(f)); err == nil {
			cpt++
		}
	}
	return cpt
}
