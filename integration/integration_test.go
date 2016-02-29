package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
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

func TestOrphanProcessReaping(t *testing.T) {
	d := newDocker()
	// using --debug, dock will have a 999 exit code if more than one process exists when exiting
	if err := d.start(true, "run", testImage, "dock", "--debug", "bash", "/go/src/github.com/robinmonjo/dock/integration/assets/spawn_orphaned.sh"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
}

func TestWebHook(t *testing.T) {
	port := "9999"
	host := "192.168.99.1" //WARNING, made to work with docker machine and virtualbox on a mac
	expectedStatus := []notifier.PsStatus{notifier.StatusStarting, notifier.StatusRunning, notifier.StatusCrashed}
	var wg sync.WaitGroup
	wg.Add(len(expectedStatus))
	cpt := 0

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		var payload notifier.HookPayload

		if err := decoder.Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Ps.Status != expectedStatus[cpt] {
			t.Fatalf("expected process status %q got %q", payload.Ps.Status, expectedStatus[cpt])
		}
		cpt++
		wg.Done()
	})

	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			t.Fatal(err)
		}
	}()

	d := newDocker()

	if err := d.start(false, "run", testImage, "dock", "--debug", "--web-hook", fmt.Sprintf("http://%s:%s", host, port), "ls"); err != nil {
		fmt.Println(d.debugInfo())
		t.Fatal(err)
	}
	//wait for the hook to be sent
	wg.Wait()
}
