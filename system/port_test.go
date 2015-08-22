package system

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_portBinder(t *testing.T) {
	fmt.Printf("port binder ... ")
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("nc", "-l", port)
	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	time.Sleep(100 * time.Millisecond) //just make sure cmd has time to bind the port

	binder, err := portBinder(port)
	if err != nil {
		t.Fatal(err)
	}

	if binder != cmd.Process.Pid {
		t.Fatalf("wrong port binder, expected %d got %d", cmd.Process.Pid, binder)
	}
	fmt.Println("done")
}

func Test_isPortBoundFalse(t *testing.T) {
	fmt.Printf("is port bound faillure ... ")
	pids := []int{2344, 2445, 1}

	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}

	bound, err := IsPortBound(port, pids) // none of these pid should have bound this port
	if err != nil {
		t.Fatal(err)
	}
	if bound {
		t.Fatalf("port %s must not be reported as bound by one of these processes %v", port, pids)
	}
	fmt.Println("done")
}

func Test_isPortBoundTrue(t *testing.T) {
	fmt.Printf("is port bound success ... ")

	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("nc", "-l", port)
	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := fmt.Sprintf("%d", cmd.Process.Pid)

	p, _ := strconv.Atoi(pid)

	pids := []int{2344, 2445, 1, p, 890}
	bound, err := IsPortBound(port, pids)
	if err != nil {
		t.Fatal(err)
	}
	if !bound {
		t.Fatalf("port %s should be reported has bound by %d", port, p)
	}
	fmt.Println("done")
}

//helper

//return a free port
func freePort() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return strings.TrimPrefix(l.Addr().String(), "[::]:"), nil
}
