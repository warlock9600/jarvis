package jump

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DiscoverHosts(homeDir string) ([]string, error) {
	if strings.TrimSpace(homeDir) == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot resolve home directory: %w", err)
		}
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	paths := []string{filepath.Join(sshDir, "config")}
	matches, err := filepath.Glob(filepath.Join(sshDir, "config.d", "*"))
	if err != nil {
		return nil, fmt.Errorf("cannot scan ~/.ssh/config.d: %w", err)
	}
	paths = append(paths, matches...)

	set := map[string]struct{}{}
	for _, p := range paths {
		hosts, err := parseHostsFile(p)
		if err != nil {
			return nil, err
		}
		for _, h := range hosts {
			set[h] = struct{}{}
		}
	}

	out := make([]string, 0, len(set))
	for h := range set {
		out = append(out, h)
	}
	sort.Strings(out)
	return out, nil
}

func parseHostsFile(path string) ([]string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot stat %s: %w", path, err)
	}
	if fi.IsDir() {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	var hosts []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := stripComment(strings.TrimSpace(s.Text()))
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.EqualFold(fields[0], "host") {
			continue
		}
		for _, h := range fields[1:] {
			if isPatternHost(h) {
				continue
			}
			hosts = append(hosts, h)
		}
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	return hosts, nil
}

func stripComment(s string) string {
	if idx := strings.Index(s, "#"); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}

func isPatternHost(h string) bool {
	return strings.ContainsAny(h, "*?!")
}
