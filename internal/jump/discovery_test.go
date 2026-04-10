package jump

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDiscoverHosts(t *testing.T) {
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	cfgD := filepath.Join(sshDir, "config.d")
	if err := os.MkdirAll(cfgD, 0o755); err != nil {
		t.Fatal(err)
	}

	mainCfg := "\nHost app-prod app-stage\n  User ubuntu\nHost *\n  ForwardAgent yes\n"
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte(mainCfg), 0o600); err != nil {
		t.Fatal(err)
	}

	extraCfg := "\n# comment\nHost db-prod\nHost !bastion*\nHost bastion\n"
	if err := os.WriteFile(filepath.Join(cfgD, "team.conf"), []byte(extraCfg), 0o600); err != nil {
		t.Fatal(err)
	}

	hosts, err := DiscoverHosts(home)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"app-prod", "app-stage", "bastion", "db-prod"}
	if !slices.Equal(hosts, want) {
		t.Fatalf("hosts mismatch\n got: %v\nwant: %v", hosts, want)
	}
}
