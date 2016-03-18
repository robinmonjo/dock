package port

import (
	"os"
	"sort"

	"github.com/robinmonjo/dock/procfs"
)

// check whether the port is bound by one of the given PID. Return the PID of the process, 0 if no pids list specified or -1 if port is not bound
func IsPortBound(port string, pids []int) (int, error) {
	sockets, err := procfs.ReadNet()
	if err != nil {
		return -1, err
	}

	if len(pids) == 0 {
		//no PIDs specified, just check if the port is bound
		for _, s := range sockets {
			if s.LocalPort == port {
				return 0, nil //port is bound but we don't care about the PID
			}
		}
		return -1, nil
	}

	sort.Sort(procfs.Sockets(sockets)) //sort output by inode for faster search

	for _, pid := range pids {
		p := &procfs.Proc{
			Pid: pid,
		}

		// get back all file descriptors associated to this PID
		fds, err := p.Fds()
		if err != nil {
			if !os.IsPermission(err) {
				return -1, err
			}
		}

		// get back inodes
		inodes := []string{}

		for _, fd := range fds {
			inode := fd.SocketInode()
			if inode != "" {
				inodes = append(inodes, inode)
			}
		}

		for _, inode := range inodes {
			if s := procfs.Sockets(sockets).Find(inode); s != nil {
				if s.LocalPort == port {
					return p.Pid, nil
				}
			}
		}
	}
	return -1, nil
}
