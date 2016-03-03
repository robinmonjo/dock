package procfs

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDecodeSigMask(t *testing.T) {
	masks := []string{"fffffffe7ffbfeff", "00000000280b2603", "0000000000000000"}
	expected := [][]int{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 15, 16, 17, 18, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		[]int{1, 2, 10, 11, 14, 17, 18, 20, 28, 30},
		[]int{},
	}

	for i, mask := range masks {
		signals := decodeSigMask(mask)

		exp := expected[i]
		if len(exp) != len(signals) {
			t.Fatalf("expected %d signals got %d for mask %s", len(exp), len(signals), mask)
		}

		for j, sig := range signals {
			if int(sig) != exp[j] {
				t.Fatalf("expected sig %d, got sig number %d", sig, exp[j])
			}
		}
	}
}

func TestParseStatusFile(t *testing.T) {
	Mountpoint = "./assets/proc"
	p := &Proc{
		Pid: 1,
	}

	ps, err := p.Status()
	if err != nil {
		t.Fatal(err)
	}

	if ps.Name != "bash" {
		t.Fatalf("expected Name bash, got %q", ps.Name)
	}
	if ps.PPid != 0 {
		t.Fatalf("expected PPid 0, got %d", ps.PPid)
	}
	if ps.Uid != 0 {
		t.Fatalf("expected Uid 0, got %d", ps.Uid)
	}
	if ps.State != "S (sleeping)" {
		t.Fatalf("expected State S (sleeping), got %q", ps.State)
	}
	if len(ps.SigBlk) != 1 {
		t.Fatalf("expected 1 signal blocked, got %d", len(ps.SigBlk))
	}
	if len(ps.SigIgn) != 4 {
		t.Fatalf("expected 4 signals ignored, got %d", len(ps.SigIgn))
	}
	if len(ps.SigCgt) != 19 {
		t.Fatalf("expected 20 signals blocked, got %d", len(ps.SigCgt))
	}
}

func TestFds(t *testing.T) {
	Mountpoint = "./assets/proc"
	p := &Proc{
		Pid: 9,
	}
	fds, err := p.Fds()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		fd := fds[i]
		//expected <i> ../../symlinktargets/targ<i>
		expectedSource := fmt.Sprintf("%d", i)
		if fd.Source != expectedSource {
			t.Fatalf("expected source to be %q, got %q", expectedSource, fd.Source)
		}

		expectedTarget := fmt.Sprintf("../../symlinktargets/targ%d", i)
		if fd.Target != expectedTarget {
			t.Fatalf("expected target to be %q, got %q", expectedTarget, fd.Target)
		}
	}

	inode := fds[0].SocketInode()
	if inode != "" {
		t.Fatal("first file descriptor shouldn't be a socket")
	}

	inode = fds[3].SocketInode()
	if inode != "84336181" {
		t.Fatalf("expected inode to be 84336181, got %q", inode)
	}
}

func TestChildren(t *testing.T) {
	Mountpoint = "./assets/proc"
	p := &Proc{
		Pid: 9,
	}
	children, err := p.Children()
	if err != nil {
		t.Fatal(err)
	}
	expected := []*Proc{
		&Proc{Pid: 12},
		&Proc{Pid: 14},
	}
	if !reflect.DeepEqual(children, expected) {
		t.Fatalf("expected processes %#v, got %#v", expected, children)
	}
}

func TestNoChild(t *testing.T) {
	Mountpoint = "./assets/proc"
	p := &Proc{
		Pid: 12,
	}
	children, err := p.Children()
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 0 {
		t.Fatal("pid 12 should have no children")
	}
}

func TestDescendants(t *testing.T) {
	Mountpoint = "./assets/proc"
	p := &Proc{
		Pid: 1,
	}
	descendants, err := p.Descendants()
	if err != nil {
		t.Fatal(err)
	}

	expected := []*Proc{
		&Proc{Pid: 9},
		&Proc{Pid: 12},
		&Proc{Pid: 14},
	}
	if !reflect.DeepEqual(descendants, expected) {
		t.Fatalf("expected processes %#v, got %#v", expected, descendants)
	}
}

func TestNoDescendant(t *testing.T) {
	Mountpoint = "./assets/proc"
	p := &Proc{
		Pid: 14,
	}
	descendants, err := p.Descendants()
	if err != nil {
		t.Fatal(err)
	}
	if len(descendants) != 0 {
		t.Fatal("pid 14 should have no descendants")
	}
}
