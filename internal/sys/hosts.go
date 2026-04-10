package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultHostsMarker = "# --- jarvis managed hosts ---"

type HostsEntry struct {
	Address  string `json:"address"`
	Hostname string `json:"hostname"`
	Enabled  bool   `json:"enabled"`
	Raw      string `json:"raw"`
}

type HostsManager struct {
	Path   string
	Marker string
}

func NewHostsManager(path, marker string) *HostsManager {
	if strings.TrimSpace(path) == "" {
		path = "/etc/hosts"
	}
	if strings.TrimSpace(marker) == "" {
		marker = DefaultHostsMarker
	}
	return &HostsManager{Path: path, Marker: marker}
}

func (m *HostsManager) TakeControl() (bool, error) {
	lines, hadNewline, err := m.readLines()
	if err != nil {
		return false, err
	}
	if findMarker(lines, m.Marker) >= 0 {
		return false, nil
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
		lines = append(lines, "")
	}
	lines = append(lines, m.Marker)
	return true, m.writeLines(lines, hadNewline)
}

func (m *HostsManager) DisableAll() (int, error) {
	lines, hadNewline, err := m.readLines()
	if err != nil {
		return 0, err
	}
	idx := findMarker(lines, m.Marker)
	if idx < 0 {
		return 0, fmt.Errorf("marker not found. run 'jarvis sys hosts take-control' first")
	}
	changed := 0
	for i := idx + 1; i < len(lines); i++ {
		next := commentLine(lines[i])
		if next != lines[i] {
			changed++
			lines[i] = next
		}
	}
	return changed, m.writeLines(lines, hadNewline)
}

func (m *HostsManager) EnableAll() (int, error) {
	lines, hadNewline, err := m.readLines()
	if err != nil {
		return 0, err
	}
	idx := findMarker(lines, m.Marker)
	if idx < 0 {
		return 0, fmt.Errorf("marker not found. run 'jarvis sys hosts take-control' first")
	}
	changed := 0
	for i := idx + 1; i < len(lines); i++ {
		next := uncommentLine(lines[i])
		if next != lines[i] {
			changed++
			lines[i] = next
		}
	}
	return changed, m.writeLines(lines, hadNewline)
}

func (m *HostsManager) Add(address, hostname string) (bool, error) {
	lines, hadNewline, err := m.readLines()
	if err != nil {
		return false, err
	}
	idx := findMarker(lines, m.Marker)
	if idx < 0 {
		return false, fmt.Errorf("marker not found. run 'jarvis sys hosts take-control' first")
	}
	for i := idx + 1; i < len(lines); i++ {
		e, ok := parseEntry(lines[i])
		if ok && e.Address == address && e.Hostname == hostname {
			if !e.Enabled {
				lines[i] = formatEntry(address, hostname, true)
				return true, m.writeLines(lines, hadNewline)
			}
			return false, nil
		}
	}
	lines = append(lines, formatEntry(address, hostname, true))
	return true, m.writeLines(lines, hadNewline)
}

func (m *HostsManager) Delete(address, hostname string) (int, error) {
	lines, hadNewline, err := m.readLines()
	if err != nil {
		return 0, err
	}
	idx := findMarker(lines, m.Marker)
	if idx < 0 {
		return 0, fmt.Errorf("marker not found. run 'jarvis sys hosts take-control' first")
	}

	out := lines[:idx+1]
	deleted := 0
	for i := idx + 1; i < len(lines); i++ {
		e, ok := parseEntry(lines[i])
		if ok && e.Address == address && e.Hostname == hostname {
			deleted++
			continue
		}
		out = append(out, lines[i])
	}
	if deleted == 0 {
		return 0, nil
	}
	return deleted, m.writeLines(out, hadNewline)
}

func (m *HostsManager) Entries() ([]HostsEntry, error) {
	lines, _, err := m.readLines()
	if err != nil {
		return nil, err
	}
	idx := findMarker(lines, m.Marker)
	if idx < 0 {
		return nil, fmt.Errorf("marker not found. run 'jarvis sys hosts take-control' first")
	}
	entries := make([]HostsEntry, 0)
	for i := idx + 1; i < len(lines); i++ {
		e, ok := parseEntry(lines[i])
		if ok {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func (m *HostsManager) Clean() (int, error) {
	lines, hadNewline, err := m.readLines()
	if err != nil {
		return 0, err
	}
	idx := findMarker(lines, m.Marker)
	if idx < 0 {
		return 0, fmt.Errorf("marker not found. run 'jarvis sys hosts take-control' first")
	}
	count := len(lines[idx+1:])
	lines = lines[:idx+1]
	return count, m.writeLines(lines, hadNewline)
}

func (m *HostsManager) readLines() ([]string, bool, error) {
	b, err := os.ReadFile(m.Path)
	if err != nil {
		return nil, false, wrapPermissionErr(m.Path, err)
	}
	raw := string(b)
	hadNewline := strings.HasSuffix(raw, "\n")
	raw = strings.TrimSuffix(raw, "\n")
	if raw == "" {
		return []string{}, hadNewline, nil
	}
	return strings.Split(raw, "\n"), hadNewline, nil
}

func (m *HostsManager) writeLines(lines []string, hadNewline bool) error {
	content := strings.Join(lines, "\n")
	if hadNewline || len(lines) > 0 {
		content += "\n"
	}

	fi, err := os.Stat(m.Path)
	if err != nil {
		return wrapPermissionErr(m.Path, err)
	}
	mode := fi.Mode().Perm()

	tmp := m.Path + ".jarvis.tmp"
	if err := os.WriteFile(tmp, []byte(content), mode); err != nil {
		return wrapPermissionErr(m.Path, err)
	}
	if err := os.Rename(tmp, m.Path); err != nil {
		_ = os.Remove(tmp)
		return wrapPermissionErr(m.Path, err)
	}
	return nil
}

func findMarker(lines []string, marker string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) == marker {
			return i
		}
	}
	return -1
}

func parseEntry(line string) (HostsEntry, bool) {
	raw := line
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return HostsEntry{}, false
	}
	enabled := true
	if strings.HasPrefix(trimmed, "#") {
		enabled = false
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
	}
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return HostsEntry{}, false
	}
	fields := strings.Fields(trimmed)
	if len(fields) < 2 {
		return HostsEntry{}, false
	}
	return HostsEntry{Address: fields[0], Hostname: fields[1], Enabled: enabled, Raw: raw}, true
}

func formatEntry(address, hostname string, enabled bool) string {
	if enabled {
		return fmt.Sprintf("%s %s", address, hostname)
	}
	return fmt.Sprintf("# %s %s", address, hostname)
}

func commentLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}
	return "# " + trimmed
}

func uncommentLine(line string) string {
	left := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(left, "#") {
		return line
	}
	left = strings.TrimPrefix(left, "#")
	left = strings.TrimLeft(left, " \t")
	return left
}

func wrapPermissionErr(path string, err error) error {
	if os.IsPermission(err) {
		return fmt.Errorf("permission denied for %s (try with sudo): %w", filepath.Clean(path), err)
	}
	return err
}
