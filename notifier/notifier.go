package notifier

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

type PsStatus string

const (
	StatusStarting PsStatus = "starting"
	StatusRunning  PsStatus = "running"
	StatusCrashed  PsStatus = "crashed"
)

var WebHook string

type Ps struct {
	Status        PsStatus        `json:"status"`
	NetInterfaces []*NetInterface `json:"net_interfaces"`
}

type NetInterface struct {
	Name string `json:"name"`
	IPv4 string `json:"ipv4,omitempty"`
	IPv6 string `json:"ipv6,omitempty"`
}

type HookPayload struct {
	Ps *Ps `json:"ps"`
}

func NotifyHook(status PsStatus) error {
	payload := &HookPayload{
		&Ps{
			Status:        status,
			NetInterfaces: netInterfaces(),
		},
	}

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

func netInterfaces() (netInterfaces []*NetInterface) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Error(err)
		return
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Error(err)
			continue
		}

		netInterface := &NetInterface{
			Name: iface.Name,
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.To4() != nil {
				netInterface.IPv4 = ip.String()
			} else {
				netInterface.IPv6 = ip.String()
			}
		}
		netInterfaces = append(netInterfaces, netInterface)
	}
	return
}
