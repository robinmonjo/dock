package system

import (
	"os"
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
		signals, err := decodeSigMask(mask)
		if err != nil {
			t.Fatal(err)
		}
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
	f, err := os.Open("./assets/proc/1/status")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p := &Process{
		Pid: 9,
	}
	ps, err := p.parseStatusFile(f)
	if err != nil {
		t.Fatal(err)
	}

	if ps.Name != "bash" {
		t.Fatalf("expected Name bash, got %q", ps.Name)
	}
	if ps.PPid != 1 {
		t.Fatalf("expected PPid 1, got %d", ps.PPid)
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

func TestIntifySlice(t *testing.T) {
	res, err := intifySlice([]string{"1", "2", "3", "19", "2345", "8999"})
	if err != nil {
		t.Fatal(err)
	}
	expected := []int{1, 2, 3, 19, 2345, 8999}
	if !reflect.DeepEqual(expected, res) {
		t.Fatalf("expected %v, got %v", expected, res)
	}
}

func TestChildren(t *testing.T) {
	//TODO
}
