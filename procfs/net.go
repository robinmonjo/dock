package procfs

import (
	"bufio"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	localAddrCol  = 1
	remoteAddrCol = 2
	inodeCol      = 9
)

var protocols = []string{"tcp", "tcp6", "udp", "udp6"}

type Socket struct {
	Protocol   string
	LocalIP    net.IP
	LocalPort  string
	RemoteIP   net.IP
	RemotePort string
	Inode      string
}

func ReadNet() ([]*Socket, error) {
	var (
		sockets = []*Socket{}
		err     error
		mutex   = &sync.Mutex{}
		wg      sync.WaitGroup
	)

	wg.Add(len(protocols))
	for _, proto := range protocols {
		go func(p string) {
			s, e := parseNetFile(p)
			mutex.Lock()
			if e != nil {
				err = e
			} else {
				sockets = append(sockets, s...)
			}
			mutex.Unlock()
			wg.Done()
		}(proto)
	}

	wg.Wait()
	return sockets, err
}

func parseNetFile(protocol string) ([]*Socket, error) {
	f, err := os.Open(filepath.Join(Mountpoint, "net", protocol))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sockets := []*Socket{}

	scanner := bufio.NewScanner(f)
	scanner.Scan() //flush file header

	for scanner.Scan() {
		sock := processLine(scanner.Text(), protocol)
		sockets = append(sockets, sock)
	}

	return sockets, scanner.Err()
}

func processLine(line, protocol string) *Socket {
	columns := strings.Fields(line)

	s := &Socket{
		Protocol: protocol,
	}

	for i, c := range columns {
		switch i {
		case localAddrCol:
			addr := strings.Split(c, ":")
			s.LocalIP = hexStringToIP(addr[0])
			s.LocalPort = hexStringToDecimalPort(addr[1])
		case remoteAddrCol:
			addr := strings.Split(c, ":")
			s.RemoteIP = hexStringToIP(addr[0])
			s.RemotePort = hexStringToDecimalPort(addr[1])
		case inodeCol:
			s.Inode = c
		}
	}

	return s
}

//sort warppers
type Sockets []*Socket

func (sockets Sockets) Len() int           { return len(sockets) }
func (sockets Sockets) Swap(i, j int)      { sockets[i], sockets[j] = sockets[j], sockets[i] }
func (sockets Sockets) Less(i, j int) bool { return sockets[i].Inode < sockets[j].Inode }

func (sockets Sockets) Find(inode string) *Socket {
	i := sort.Search(len(sockets), func(i int) bool {
		return sockets[i].Inode >= inode
	})
	if i < len(sockets) && sockets[i].Inode == inode {
		return sockets[i]
	}
	return nil
}

//utilities
func hexStringToIP(str string) net.IP {
	b, _ := hex.DecodeString(str)
	return net.IP(b)
}

func hexStringToDecimalPort(str string) string {
	p, _ := strconv.ParseInt(str, 16, 32)
	return strconv.Itoa(int(p))
}
