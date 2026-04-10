//go:build !linux && !darwin

package netutil

import "errors"

func FlushDNS(dryRun bool) (FlushResult, error) {
	_ = dryRun
	return FlushResult{OS: "unsupported", Message: "dns flush is not supported on this OS"}, errors.New("dns flush is not supported on this OS")
}
