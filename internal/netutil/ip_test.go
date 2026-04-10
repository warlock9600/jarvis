package netutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublicIPFallback(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()

	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "203.0.113.11")
	}))
	defer good.Close()

	ips, errs := PublicIP([]string{bad.URL, good.URL}, 2*time.Second, 0, true, false)
	if len(ips) != 1 {
		t.Fatalf("expected 1 ip, got %d", len(ips))
	}
	if ips[0].Address != "203.0.113.11" {
		t.Fatalf("unexpected ip: %s", ips[0].Address)
	}
	if len(errs) == 0 {
		t.Fatalf("expected at least one error from fallback provider")
	}
}
