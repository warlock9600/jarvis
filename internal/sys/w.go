package sys

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type ProcessInfo struct {
	PID      int     `json:"pid"`
	User     string  `json:"user"`
	Command  string  `json:"command"`
	CPU      float64 `json:"cpu"`
	MemPct   float64 `json:"mem_pct"`
	RSSKB    int64   `json:"rss_kb"`
}

type WSnapshot struct {
	LoadAverage   []float64     `json:"load_average"`
	WOutput       string        `json:"w_output"`
	WMessage      string        `json:"w_message,omitempty"`
	FreeOutput    string        `json:"free_output,omitempty"`
	FreeSupported bool          `json:"free_supported"`
	FreeMessage   string        `json:"free_message,omitempty"`
	Processes     []ProcessInfo `json:"processes"`
}

func CollectWSnapshot(topN int) (WSnapshot, error) {
	if topN <= 0 {
		topN = 15
	}

	out := WSnapshot{}

	load, err := loadAverage()
	if err != nil {
		return out, err
	}
	out.LoadAverage = load

	wOut, err := runCommand("w")
	if err != nil {
		if strings.TrimSpace(wOut) == "" {
			return out, fmt.Errorf("run w: %w", err)
		}
		out.WMessage = fmt.Sprintf("w returned warning: %v", err)
	}
	out.WOutput = strings.TrimRight(wOut, "\n")

	freeOut, freeMsg, freeOK := freeOutput()
	out.FreeOutput = strings.TrimRight(freeOut, "\n")
	out.FreeMessage = freeMsg
	out.FreeSupported = freeOK

	processes, err := topProcesses(topN)
	if err != nil {
		return out, err
	}
	out.Processes = processes
	return out, nil
}

func loadAverage() ([]float64, error) {
	if runtime.GOOS == "linux" {
		b, err := os.ReadFile("/proc/loadavg")
		if err == nil {
			fields := strings.Fields(string(b))
			if len(fields) >= 3 {
				vals := make([]float64, 0, 3)
				for i := 0; i < 3; i++ {
					v, err := strconv.ParseFloat(fields[i], 64)
					if err != nil {
						break
					}
					vals = append(vals, v)
				}
				if len(vals) == 3 {
					return vals, nil
				}
			}
		}
	}

	if runtime.GOOS == "darwin" {
		out, err := runCommand("sysctl", "-n", "vm.loadavg")
		if err == nil {
			t := strings.TrimSpace(out)
			t = strings.TrimPrefix(t, "{")
			t = strings.TrimSuffix(t, "}")
			fields := strings.Fields(t)
			if len(fields) >= 3 {
				vals := make([]float64, 0, 3)
				for i := 0; i < 3; i++ {
					v, err := strconv.ParseFloat(fields[i], 64)
					if err != nil {
						break
					}
					vals = append(vals, v)
				}
				if len(vals) == 3 {
					return vals, nil
				}
			}
		}
	}

	out, err := runCommand("uptime")
	if err == nil {
		t := strings.ToLower(strings.TrimSpace(out))
		idx := strings.LastIndex(t, "load average")
		if idx < 0 {
			idx = strings.LastIndex(t, "load averages")
		}
		if idx >= 0 {
			part := strings.TrimSpace(out[idx:])
			if colon := strings.Index(part, ":"); colon >= 0 {
				part = strings.TrimSpace(part[colon+1:])
			}
			part = strings.ReplaceAll(part, ",", " ")
			fields := strings.Fields(part)
			if len(fields) >= 3 {
				vals := make([]float64, 0, 3)
				for i := 0; i < 3; i++ {
					v, err := strconv.ParseFloat(fields[i], 64)
					if err != nil {
						break
					}
					vals = append(vals, v)
				}
				if len(vals) == 3 {
					return vals, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("cannot read load average on this OS")
}

func freeOutput() (string, string, bool) {
	if runtime.GOOS != "linux" {
		return "", "free -m is not supported on this OS", false
	}
	out, err := runCommand("free", "-m")
	if err != nil {
		return "", fmt.Sprintf("free -m failed: %v", err), false
	}
	return out, "", true
}

func topProcesses(topN int) ([]ProcessInfo, error) {
	out, err := runCommand("ps", "-axo", "pid,user,comm,%cpu,%mem,rss")
	if err != nil {
		return nil, fmt.Errorf("run ps: %w", err)
	}

	var items []ProcessInfo
	s := bufio.NewScanner(strings.NewReader(out))
	lineNo := 0
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		lineNo++
		if lineNo == 1 && strings.Contains(strings.ToLower(line), "pid") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		cpu, err2 := strconv.ParseFloat(fields[3], 64)
		mem, err3 := strconv.ParseFloat(fields[4], 64)
		rss, err4 := strconv.ParseInt(fields[5], 10, 64)
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			continue
		}
		items = append(items, ProcessInfo{
			PID:     pid,
			User:    fields[1],
			Command: fields[2],
			CPU:     cpu,
			MemPct:  mem,
			RSSKB:   rss,
		})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].RSSKB == items[j].RSSKB {
			return items[i].CPU > items[j].CPU
		}
		return items[i].RSSKB > items[j].RSSKB
	})

	if len(items) > topN {
		items = items[:topN]
	}
	return items, nil
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	b, err := cmd.CombinedOutput()
	out := string(b)
	if err != nil {
		return out, fmt.Errorf("%w: %s", err, strings.TrimSpace(out))
	}
	return out, nil
}
