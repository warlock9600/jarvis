package netutil

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

type TLSExpiry struct {
	Host       string    `json:"host"`
	Port       string    `json:"port"`
	NotAfter   time.Time `json:"not_after"`
	DaysLeft   int       `json:"days_left"`
	ServerName string    `json:"server_name"`
}

func TLSExpiryCheck(target string, timeout time.Duration) (TLSExpiry, error) {
	host, port := splitHostPort(target)
	addr := net.JoinHostPort(host, port)
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return TLSExpiry{}, err
	}
	defer conn.Close()
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return TLSExpiry{}, fmt.Errorf("no peer certificates received")
	}
	cert := state.PeerCertificates[0]
	days := int(time.Until(cert.NotAfter).Hours() / 24)
	return TLSExpiry{Host: host, Port: port, NotAfter: cert.NotAfter, DaysLeft: days, ServerName: cert.Subject.CommonName}, nil
}

func splitHostPort(target string) (string, string) {
	if strings.Contains(target, ":") {
		h, p, err := net.SplitHostPort(target)
		if err == nil {
			return h, p
		}
	}
	return target, "443"
}
