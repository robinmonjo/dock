package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/robinmonjo/dock/notifier"
)

type hookServer struct {
	c chan notifier.PsStatus
	t *testing.T
}

func (s *hookServer) start() (host, port string) {
	host = "192.168.99.1" //WARNING, made to work with docker machine and virtualbox on a mac
	port = "8080"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		var payload notifier.HookPayload

		if err := decoder.Decode(&payload); err != nil {
			s.t.Fatal(err)
		}

		s.c <- payload.Ps.Status
	})

	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			s.t.Fatal(err)
		}
	}()
	return
}
