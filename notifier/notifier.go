package notifier

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type PsStatus string

const (
	StatusStarting PsStatus = "starting"
	StatusRunning  PsStatus = "running"
	StatusCrashed  PsStatus = "crashed"
)

var WebHook string

type Ps struct {
	Status PsStatus `json:"status"`
}

type HookPayload struct {
	Ps *Ps `json:"ps"`
}

func NotifyHook(status PsStatus) error {
	payload := &HookPayload{&Ps{Status: status}}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", WebHook, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	u, err := url.Parse(WebHook)
	if err != nil {
		return err
	}

	var client *http.Client

	if u.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("bad status code expected 200 .. 299 got %d", resp.Status)
	}
	return nil
}
