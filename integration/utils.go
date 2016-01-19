package integration

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	"github.com/robinmonjo/dock/notifier"
)

func statusFromHookBody(body io.Reader, t *testing.T) notifier.PsStatus {
	var payload notifier.HookPayload

	content, _ := ioutil.ReadAll(body)
	if err := json.Unmarshal(content, &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Ps.Status
}
