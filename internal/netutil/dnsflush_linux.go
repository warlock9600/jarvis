//go:build linux

package netutil

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

func FlushDNS(dryRun bool) (FlushResult, error) {
	candidates := [][][]string{
		{{"resolvectl", "flush-caches"}},
		{{"systemd-resolve", "--flush-caches"}},
		{{"service", "nscd", "restart"}},
	}

	res := FlushResult{OS: "linux"}
	for _, group := range candidates {
		ok := true
		var steps []FlushStep
		for _, c := range group {
			if _, err := exec.LookPath(c[0]); err != nil {
				ok = false
				break
			}
			step := FlushStep{Command: c[0] + " " + c[1]}
			if dryRun {
				step.Status = "planned"
				steps = append(steps, step)
				continue
			}
			cmd := exec.Command(c[0], c[1:]...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				step.Status = "failed"
				step.Output = stderr.String()
				return res, fmt.Errorf("%s: %w", step.Command, err)
			}
			step.Status = "done"
			steps = append(steps, step)
		}
		if ok {
			res.Steps = steps
			res.Message = "DNS cache flushed on Linux"
			return res, nil
		}
	}
	res.Message = "No supported DNS cache service found (systemd-resolved or nscd)"
	return res, errors.New(res.Message)
}
