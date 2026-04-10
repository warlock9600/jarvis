package netutil

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type IPResult struct {
	Local  []IPItem `json:"local,omitempty"`
	Public []IPItem `json:"public,omitempty"`
}

type IPItem struct {
	Interface string `json:"interface,omitempty"`
	Version   string `json:"version"`
	Address   string `json:"address"`
	Source    string `json:"source,omitempty"`
}

func LocalIPs(v4, v6 bool) ([]IPItem, error) {
	if !v4 && !v6 {
		v4, v6 = true, true
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var out []IPItem
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP
			if ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil && v4 {
				out = append(out, IPItem{Interface: iface.Name, Version: "ipv4", Address: ip.String()})
			}
			if ip.To4() == nil && v6 {
				out = append(out, IPItem{Interface: iface.Name, Version: "ipv6", Address: ip.String()})
			}
		}
	}
	return out, nil
}

func PublicIP(providers []string, timeout time.Duration, retries int, v4, v6 bool) ([]IPItem, []error) {
	if !v4 && !v6 {
		v4, v6 = true, true
	}
	client := &http.Client{Timeout: timeout}
	var out []IPItem
	var errs []error

	for _, p := range providers {
		var body string
		var err error
		for i := 0; i <= retries; i++ {
			body, err = fetchURL(client, p)
			if err == nil {
				break
			}
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("provider %s: %w", p, err))
			continue
		}
		ip := net.ParseIP(strings.TrimSpace(body))
		if ip == nil {
			errs = append(errs, fmt.Errorf("provider %s returned invalid IP", p))
			continue
		}
		if ip.To4() != nil && v4 {
			out = append(out, IPItem{Version: "ipv4", Address: ip.String(), Source: p})
		}
		if ip.To4() == nil && v6 {
			out = append(out, IPItem{Version: "ipv6", Address: ip.String(), Source: p})
		}
		if len(out) > 0 {
			return out, errs
		}
	}
	return out, errs
}

func fetchURL(client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
