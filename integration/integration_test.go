package integration

import (
	"fmt"
	"os"
	"testing"
)

var testImage string //must match with what is in the Makefile

func init() {
	testImage = os.Getenv("TEST_IMAGE")
}

func TestSimple(t *testing.T) {
	d := newDocker()
	if err := d.start(true, "run", testImage, "dock", "--debug", "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
}

func TestInteractive(t *testing.T) {
	d := newDocker()
	if err := d.start(true, "run", "-t", testImage, "dock", "-i", "--debug", "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
}

func TestInteractiveNot(t *testing.T) {
	d := newDocker()
	if err := d.start(true, "run", testImage, "dock", "-i", "--debug", "ls"); err == nil {
		t.Fatal(err)
	}
}
