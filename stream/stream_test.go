package stream

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"testing"
)

func Test_remoteStream(t *testing.T) {
	fmt.Printf("Remote TCP stream ... ")

	//create a simple tcp server
	ln, err := net.Listen("tcp", ":9999")
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
	s, err := NewStream("tcp://localhost:9999", "", NoColor)
	if err != nil {
		t.Fatal(err)
	}

	mess := []byte("foo bar\n")
	if _, err := s.Write(mess); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	received, err := bufio.NewReader(s).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}

	if string(mess) != received {
		t.Fatal("expected to receive %s, got %s", string(mess), received)
	}

	//make sure close channel is called
	wg.Add(1)
	go func() {
		<-s.CloseCh
		wg.Done()
	}()

	s.Close()

	wg.Wait()

	fmt.Println("done")
}

func Test_fileStream(t *testing.T) {
	fmt.Printf("File stream with prefix ... ")
	s, err := NewStream("file:///tmp/psdock_test.log", "prefix ", NoColor)
	if err != nil {
		t.Fatal(err)
	}
	if s.Input != nil {
		t.Fatal("file stream are not expected to have input streams")
	}
	s.Write([]byte("foo bar"))
	s.Close()

	content, err := ioutil.ReadFile("/tmp/psdock_test.log")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove("/tmp/psdock_test.log")
	if string(content) != "prefix foo bar" {
		t.Fatalf("expecting \"prefix foo bar\" got \"%s\"", string(content))
	}

	fmt.Println("done")
}
