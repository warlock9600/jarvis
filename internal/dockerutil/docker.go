package dockerutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Image struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	Created string `json:"created"`
}

type PruneResult struct {
	Candidates int    `json:"candidates"`
	Deleted    int    `json:"deleted"`
	DryRun     bool   `json:"dry_run"`
	Output     string `json:"output,omitempty"`
}

func Images() ([]Image, error) {
	cmd := exec.Command("docker", "image", "ls", "--format", "{{json .}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run docker image ls: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	res := make([]Image, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row map[string]string
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		res = append(res, Image{
			Name:    fmt.Sprintf("%s:%s", row["Repository"], row["Tag"]),
			Size:    row["Size"],
			Created: row["CreatedSince"],
		})
	}
	return res, nil
}

func PruneDangling(dryRun bool) (PruneResult, error) {
	idsCmd := exec.Command("docker", "images", "-f", "dangling=true", "-q")
	idsOut, err := idsCmd.Output()
	if err != nil {
		return PruneResult{}, fmt.Errorf("inspect dangling images: %w", err)
	}
	ids := strings.Fields(string(idsOut))
	result := PruneResult{Candidates: len(ids), DryRun: dryRun}
	if dryRun || len(ids) == 0 {
		return result, nil
	}

	cmd := exec.Command("docker", "image", "prune", "-f")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return result, fmt.Errorf("docker prune failed: %w", err)
	}
	result.Output = out.String()
	result.Deleted = result.Candidates
	return result, nil
}
