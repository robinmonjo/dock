package system

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	localAddrColumn = 1
	inodeColumn     = 9
)

var (
	mountPoint = "/proc"
	protocols  = []string{"tcp", "tcp6", "udp", "udp6"}
)

type Socket struct {
	Protocol string
	BindPort string
	Inode    string
}

func Net() ([]*Socket, error) {
	var (
		sockets = []*Socket{}
		err     error
		mutex   = &sync.Mutex{}
		wg      sync.WaitGroup
	)

	wg.Add(len(protocols))
	for _, proto := range protocols {
		go func(p string) {
			s, e := parseNetfile(p)
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

func parseNetfile(protocol string) ([]*Socket, error) {
	f, err := os.Open(filepath.Join(mountPoint, "net", protocol))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sockets := []*Socket{}

	scanner := bufio.NewScanner(f)

	flushedHeader := false
	for scanner.Scan() {
		line := scanner.Text()
		if !flushedHeader {
			flushedHeader = true
			continue
		}
		sock := processLine(line, protocol)
		sockets = append(sockets, sock)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sockets, nil
}

func processLine(line, protocol string) *Socket {
	columns := strings.Split(line, " ")

	s := &Socket{
		Protocol: protocol,
	}

	c := 0
	for _, _c := range columns {
		if _c == "" {
			continue
		}
		switch c {
		case localAddrColumn: // "0.0.0.0:9999"
			hexPort := strings.Split(_c, ":")[1]
			i, _ := strconv.ParseInt(hexPort, 16, 32)
			s.BindPort = strconv.Itoa(int(i))
		case inodeColumn:
			s.Inode = _c
		}
		c++
	}

	return s
}
