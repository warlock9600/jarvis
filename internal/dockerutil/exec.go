package dockerutil

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func RunningContainers() ([]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}\t{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run docker ps: %w", err)
	}

	set := map[string]struct{}{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) > 0 {
			id := strings.TrimSpace(parts[0])
			if id != "" {
				set[id] = struct{}{}
			}
		}
		if len(parts) > 1 {
			name := strings.TrimSpace(parts[1])
			if name != "" {
				set[name] = struct{}{}
			}
		}
	}

	items := make([]string, 0, len(set))
	for x := range set {
		items = append(items, x)
	}
	sort.Strings(items)
	return items, nil
}

func PickAvailableShell(container string, preferred []string) (string, error) {
	if len(preferred) == 0 {
		preferred = []string{"bash", "sh", "ash"}
	}

	var lastErr error
	for _, shell := range preferred {
		ok, err := shellAvailable(container, shell)
		if err != nil {
			lastErr = err
			continue
		}
		if ok {
			return shell, nil
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("no usable shell found (tried %v): %w", preferred, lastErr)
	}
	return "", fmt.Errorf("no usable shell found (tried %v)", preferred)
}

func shellAvailable(container, shell string) (bool, error) {
	cmd := exec.Command("docker", "exec", container, shell, "-c", "exit 0")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return true, nil
	}

	msg := strings.ToLower(strings.TrimSpace(string(out)))
	if strings.Contains(msg, "executable file not found") || strings.Contains(msg, "not found") || strings.Contains(msg, "no such file") {
		return false, nil
	}
	return false, fmt.Errorf("check shell %s failed: %w: %s", shell, err, strings.TrimSpace(string(out)))
}

func ExecInteractive(container, shell string) error {
	cmd := exec.Command("docker", "exec", "-it", container, shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
