package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func TestLoadConfigPriority(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := []byte("net:\n  timeout_seconds: 3\n  retries: 1\n")
	if err := os.WriteFile(cfgPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("JARVIS_HTTP_TIMEOUT_SECONDS", "9"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("JARVIS_HTTP_TIMEOUT_SECONDS") })

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int("timeout", 0, "")
	_ = flags.Set("timeout", "15")

	cfg, _, err := Load(cfgPath, flags)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Net.TimeoutSeconds != 15 {
		t.Fatalf("expected flag override timeout=15, got %d", cfg.Net.TimeoutSeconds)
	}
}
