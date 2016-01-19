package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/robinmonjo/dock/notifier"
)

var testImage string //must match with what is in the Makefile

func init() {
	testImage = os.Getenv("TEST_IMAGE")
}

func TestSimpleCommand(t *testing.T) {
	d := newDocker()
	if err := d.start(true, "run", testImage, "dock", "--debug", "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
}

func TestSimpleInteractiveCommand(t *testing.T) {
	d := newDocker()
	if err := d.start(true, "run", "-t", testImage, "dock", "-i", "--debug", "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
}

func TestInteractiveNoTerminal(t *testing.T) {
	d := newDocker()
	if err := d.start(true, "run", testImage, "dock", "-i", "--debug", "ls"); err == nil { //fail because no terminal in docker container
		t.Fatal(err)
	}
}

func TestOrphanProcessReaping(t *testing.T) {
	d := newDocker()
	// using --debug, dock will have a 999 exit code if more than one process exists when exiting
	if err := d.start(true, "run", testImage, "dock", "--debug", "bash", "/assets/spawn_orphaned.sh"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
}

//this is broken
func TestWebHooks(t *testing.T) {
	//TODO: only run this test on linux (where docker is native)
	cpt := 0
	expectedStatus := []notifier.PsStatus{notifier.StatusStarting, notifier.StatusRunning, notifier.StatusCrashed}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		status := statusFromHookBody(r.Body, t)
		if status != expectedStatus[cpt] {
			t.Fatalf("expecting status %v got %v", expectedStatus[cpt], status)
		}
		cpt++
	}))
	defer ts.Close()

	d := newDocker()
	// using --debug, dock will have a 999 exit code if more than one process exists when exiting
	if err := d.start(true, "run", "--net=host", testImage, "dock", "--debug", "-web-hook", ts.URL, "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}

	if cpt != 3 {
		fmt.Println(d.debugInfo())
		t.Fatalf("hook called %d times, should have been called 3 times", cpt)
	}
}
