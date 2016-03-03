package procfs

import (
	"testing"
)

func TestReadNet(t *testing.T) {
	Mountpoint = "./assets/proc"
	sockets, err := ReadNet()
	if err != nil {
		t.Fatal(err)
	}

	if len(sockets) != 237 {
		t.Fatalf("expected 237 sockets, got %d", len(sockets))
	}

	//first tcp is supposed to bin port 9999 and have inode 84336181
	//sockets are ordered by protocols
	for _, s := range sockets {
		if s.Protocol == "tcp" {
			if s.LocalPort != "9999" {
				t.Fatalf("expected first tcp sockets to bind port 9999, got %s", s.LocalPort)
			}
			if s.Inode != "84336181" {
				t.Fatalf("expected first tcp sockets to have inode 84336181, got %s", s.Inode)
			}
			break
		}
	}

}
