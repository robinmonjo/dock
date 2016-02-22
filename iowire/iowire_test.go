package iowire

import (
	"bufio"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"testing"
)

func Test_remoteWire(t *testing.T) {
	//create a simple tcp server
	ln, err := net.Listen("tcp", ":9998")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatal(err)
		}

		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		conn.Write([]byte(message))
		wg.Done()
	}()

	//create a stream on this server
	wire, err := NewWire("tcp://127.0.0.1:9998", "", NoColor)
	if err != nil {
		t.Fatal(err)
	}

	mess := []byte("foo bar\n")
	if _, err := wire.Write(mess); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	received, err := bufio.NewReader(wire).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}

	if string(mess) != received {
		t.Fatal("expected to receive %s, got %s", string(mess), received)
	}

	//make sure close channel is called
	wg.Add(1)
	go func() {
		<-wire.CloseCh
		wg.Done()
	}()

	wire.Close()

	wg.Wait()
}

func Test_fileWire(t *testing.T) {
	wire, err := NewWire("file:///tmp/dock_test.log", "prefix ", NoColor)
	if err != nil {
		t.Fatal(err)
	}

	wire.Write([]byte("foo bar"))
	wire.Close()

	content, err := ioutil.ReadFile("/tmp/dock_test.log")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove("/tmp/dock_test.log")
	if string(content) != "prefix foo bar" {
		t.Fatalf("expecting \"prefix foo bar\" got \"%s\"", string(content))
	}
}
