package netutil

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

type DNSRecord struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func Lookup(name, server, recordType string, timeout time.Duration) ([]DNSRecord, error) {
	recordType = strings.ToUpper(recordType)
	r := &net.Resolver{}
	if server != "" {
		r = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
				d := net.Dialer{Timeout: timeout}
				return d.DialContext(ctx, "udp", net.JoinHostPort(server, "53"))
			},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var out []DNSRecord
	switch recordType {
	case "A", "AAAA":
		ips, err := r.LookupIP(ctx, "ip", name)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			if (recordType == "A" && ip.To4() != nil) || (recordType == "AAAA" && ip.To4() == nil) {
				out = append(out, DNSRecord{Type: recordType, Value: ip.String()})
			}
		}
	case "MX":
		mx, err := r.LookupMX(ctx, name)
		if err != nil {
			return nil, err
		}
		for _, m := range mx {
			out = append(out, DNSRecord{Type: "MX", Value: fmt.Sprintf("%s (pref=%d)", m.Host, m.Pref)})
		}
	case "TXT":
		txt, err := r.LookupTXT(ctx, name)
		if err != nil {
			return nil, err
		}
		for _, t := range txt {
			out = append(out, DNSRecord{Type: "TXT", Value: t})
		}
	default:
		return nil, fmt.Errorf("unsupported record type %s", recordType)
	}
	return out, nil
}
