// Package procfs provides primitives for interacting with the linux proc
// pseudo file system

package procfs

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

//sample ps like tool

func main() {

	showSockets := true // display or not UDP / TCP sockets attached to a process
	allUsers := true    // display process of all users

	uid := os.Getuid()

	var (
		sockets []*Socket
		err     error
	)

	if showSockets {
		sockets, err = ReadNet()
		if err != nil {
			panic(err)
		}
		sort.Sort(Sockets(sockets)) //sort output by inode for faster search
	}

	WalkProcs(func(p *Proc) (bool, error) {

		st, err := p.Status()
		if err != nil {
			if os.IsNotExist(err) {
				return true, nil
			}
			return false, err
		}

		if !allUsers && st.Uid != uid {
			return true, nil
		}

		//get back user
		user, err := st.User()
		if err != nil {
			return false, err
		}

		//get back process name
		n := st.Name
		args, err := p.CmdLine()
		if err != nil {
			return false, err
		}
		if args != nil {
			n = strings.Join(args, " ")
		}

		//print basic infos
		fmt.Printf("%s %d %d %v", user.Username, p.Pid, st.PPid, n)

		if showSockets {
			//get back port bound by the process
			if err := printSockets(p, sockets); err != nil {
				return false, err
			}
		}

		fmt.Printf("\n")
		return true, nil
	})
}

func printSockets(p *Proc, sockets []*Socket) error {
	fds, err := p.Fds()
	if err != nil {
		if !os.IsPermission(err) {
			return err
		}
	}

	inodes := []string{}

	for _, fd := range fds {
		inode := fd.SocketInode()
		if inode != "" {
			inodes = append(inodes, inode)
		}
	}

	str := []string{}
	for _, inode := range inodes {
		if s := Sockets(sockets).Find(inode); s != nil {
			str = append(str, fmt.Sprintf("%s %v %s", s.Protocol, s.LocalIP, s.LocalPort))
		}
	}
	fmt.Printf(" %s", strings.Join(str, ", "))
	return nil
}
