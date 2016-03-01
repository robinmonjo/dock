package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/robinmonjo/dock/notifier"
)

var (
	testImage string //must match with what is in the Makefile

	server    *hookServer
	serverURL string
)

func init() {
	testImage = os.Getenv("TEST_IMAGE")

	server = &hookServer{}
	h, p := server.start()
	serverURL = fmt.Sprintf("http://%s:%s", h, p)
}

func TestSimpleCommand(t *testing.T) {
	d := newDocker()
	if err := d.start(true, "run", testImage, "dock", "--debug", "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
}

func TestOrphanProcessReaping(t *testing.T) {
	d := newDocker()
	// using --debug, dock will have a 999 exit code if more than one process exists when exiting
	if err := d.start(true, "run", testImage, "dock", "--debug", "bash", "/go/src/github.com/robinmonjo/dock/integration/assets/spawn_orphaned.sh"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
}

func TestWebHook(t *testing.T) {
	c := make(chan notifier.PsStatus, 3)

	server.c = c
	server.t = t

	d := newDocker()

	if err := d.start(false, "run", testImage, "dock", "--debug", "--web-hook", serverURL, "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}

	for _, status := range []notifier.PsStatus{notifier.StatusStarting, notifier.StatusRunning, notifier.StatusCrashed} {
		s := <-server.c
		if s != status {
			t.Fatalf("expected status %q, got %q", notifier.StatusStarting, status)
		}
	}
}

func TestPortBindingHook(t *testing.T) {
	c := make(chan notifier.PsStatus, 3)

	server.c = c
	server.t = t
	port := "9999"

	d := newDocker()

	if err := d.start(false, "run", testImage, "dock", "--debug", "--web-hook", serverURL, "--bind-port", port, "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
	// ls will never bind the port, should never see the "running status"
	for _, status := range []notifier.PsStatus{notifier.StatusStarting, notifier.StatusCrashed} {
		s := <-server.c
		if s != status {
			t.Fatalf("expected status %q, got %q", notifier.StatusStarting, status)
		}
	}
}

func TestPortBinding(t *testing.T) {
	c := make(chan notifier.PsStatus, 3)

	server.c = c
	server.t = t

	d := newDocker()
	port := "9999"
	name := "dock-test-container"

	if err := d.start(true, "run", "-d", "--name", name, testImage, "dock", "--debug", "--web-hook", serverURL, "--bind-port", port, "python", "-m", "SimpleHTTPServer", port); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}

	defer d.start(false, "rm", name)

	s := <-server.c
	if s != notifier.StatusStarting {
		t.Fatalf("expected status %q, got %q", notifier.StatusStarting, s)
	}

	s = <-server.c
	if s != notifier.StatusRunning {
		t.Fatalf("expected status %q, got %q", notifier.StatusRunning, s)
	}

	if err := d.start(false, "stop", name); err != nil {
		t.Fatal(err)
	}
	s = <-server.c
	if s != notifier.StatusCrashed {
		t.Fatalf("expected status %q, got %q", notifier.StatusCrashed, s)
	}
}
