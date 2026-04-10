//go:build darwin

package netutil

import (
	"bytes"
	"fmt"
	"os/exec"
)

func FlushDNS(dryRun bool) (FlushResult, error) {
	cmds := [][]string{{"dscacheutil", "-flushcache"}, {"killall", "-HUP", "mDNSResponder"}}
	res := FlushResult{OS: "darwin", Message: "DNS cache flushed on macOS"}
	for _, c := range cmds {
		step := FlushStep{Command: c[0] + " " + c[1]}
		if dryRun {
			step.Status = "planned"
			res.Steps = append(res.Steps, step)
			continue
		}
		cmd := exec.Command(c[0], c[1:]...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			step.Status = "failed"
			step.Output = stderr.String()
			res.Steps = append(res.Steps, step)
			return res, fmt.Errorf("%s: %w", step.Command, err)
		}
		step.Status = "done"
		res.Steps = append(res.Steps, step)
	}
	return res, nil
}
