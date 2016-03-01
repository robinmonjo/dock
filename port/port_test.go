package port

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestIsPortBound(t *testing.T) {
	port := "8080"
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
		if err != nil {
			t.Fatal(err)
		}
	}()

	maxTry := 10
	for i := 0; i < maxTry; i++ {
		pid, err := IsPortBound(port, []int{os.Getpid()})
		if err != nil {
			t.Fatal(err)
		}
		if pid == -1 {
			continue //port not bound yet
		}
		//port bound
		if pid != os.Getpid() {
			t.Fatal("expect port to be bound by %d, got %d", os.Getpid(), pid)
		} else {
			return // ok
		}
	}
	t.Fatal("port never bound :(")
}
