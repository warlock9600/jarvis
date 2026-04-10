package netutil

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type SpeedtestResult struct {
	PingMS     float64 `json:"ping_ms"`
	JitterMS   float64 `json:"jitter_ms"`
	DownloadMB float64 `json:"download_mbps"`
	UploadMB   float64 `json:"upload_mbps"`
	Raw        any     `json:"raw,omitempty"`
}

func RunSpeedtest(bin string) (SpeedtestResult, error) {
	cmd := exec.Command(bin, "--format=json")
	out, err := cmd.Output()
	if err != nil {
		return SpeedtestResult{}, fmt.Errorf("failed to run speedtest CLI (%s). Install Ookla speedtest or set --speedtest-bin: %w", bin, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return SpeedtestResult{}, fmt.Errorf("parse speedtest output: %w", err)
	}

	res := SpeedtestResult{Raw: raw}
	res.PingMS = readFloat(raw, "ping", "latency")
	res.JitterMS = readFloat(raw, "ping", "jitter")
	res.DownloadMB = readFloat(raw, "download", "bandwidth") * 8 / 1_000_000
	res.UploadMB = readFloat(raw, "upload", "bandwidth") * 8 / 1_000_000
	if res.DownloadMB == 0 {
		res.DownloadMB = readFloat(raw, "download", "bandwidthMbps")
	}
	if res.UploadMB == 0 {
		res.UploadMB = readFloat(raw, "upload", "bandwidthMbps")
	}
	return res, nil
}

func readFloat(data map[string]any, keys ...string) float64 {
	var cur any = data
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return 0
		}
		cur = m[k]
	}
	if f, ok := cur.(float64); ok {
		return f
	}
	return 0
}
