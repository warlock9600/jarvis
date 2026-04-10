package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"jarvis/internal/common"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Net       NetConfig       `yaml:"net"`
	Speedtest SpeedtestConfig `yaml:"speedtest"`
	K8s       K8sConfig       `yaml:"k8s"`
	Secrets   SecretConfig    `yaml:"secrets"`
}

type NetConfig struct {
	PublicIPProviders []string `yaml:"public_ip_providers"`
	TimeoutSeconds    int      `yaml:"timeout_seconds"`
	Retries           int      `yaml:"retries"`
}

type SpeedtestConfig struct {
	Bin string `yaml:"bin"`
}

type K8sConfig struct {
	Kubeconfig string `yaml:"kubeconfig"`
}

type SecretConfig struct {
	RegistryToken string `yaml:"registry_token"`
	APIToken      string `yaml:"api_token"`
}

func DefaultConfig() Config {
	return Config{
		Net: NetConfig{
			PublicIPProviders: []string{"https://api.ipify.org", "https://ifconfig.me/ip"},
			TimeoutSeconds:    5,
			Retries:           2,
		},
		Speedtest: SpeedtestConfig{Bin: "speedtest"},
	}
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./config.yaml"
	}
	return filepath.Join(home, ".config", "jarvis", "config.yaml")
}

func Load(path string, flags *pflag.FlagSet) (Config, string, error) {
	cfg := DefaultConfig()
	if path == "" {
		path = DefaultPath()
	}

	if b, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return cfg, path, fmt.Errorf("parse config file %s: %w", path, err)
		}
	}

	applyEnv(&cfg)
	applyFlags(&cfg, flags)

	return cfg, path, nil
}

func applyEnv(cfg *Config) {
	if s := os.Getenv("JARVIS_PUBLIC_IP_PROVIDERS"); s != "" {
		parts := strings.Split(s, ",")
		cfg.Net.PublicIPProviders = nil
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.Net.PublicIPProviders = append(cfg.Net.PublicIPProviders, p)
			}
		}
	}
	if s := os.Getenv("JARVIS_HTTP_TIMEOUT_SECONDS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			cfg.Net.TimeoutSeconds = v
		}
	}
	if s := os.Getenv("JARVIS_HTTP_RETRIES"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 {
			cfg.Net.Retries = v
		}
	}
	if s := os.Getenv("JARVIS_SPEEDTEST_BIN"); s != "" {
		cfg.Speedtest.Bin = s
	}
	if s := os.Getenv("KUBECONFIG"); s != "" {
		cfg.K8s.Kubeconfig = s
	}
	if s := os.Getenv("JARVIS_KUBECONFIG"); s != "" {
		cfg.K8s.Kubeconfig = s
	}
	if s := os.Getenv("JARVIS_REGISTRY_TOKEN"); s != "" {
		cfg.Secrets.RegistryToken = s
	}
	if s := os.Getenv("JARVIS_API_TOKEN"); s != "" {
		cfg.Secrets.APIToken = s
	}
}

func applyFlags(cfg *Config, flags *pflag.FlagSet) {
	if flags == nil {
		return
	}
	if f := flags.Lookup("timeout"); f != nil && f.Changed {
		v, _ := flags.GetInt("timeout")
		cfg.Net.TimeoutSeconds = v
	}
	if f := flags.Lookup("retries"); f != nil && f.Changed {
		v, _ := flags.GetInt("retries")
		cfg.Net.Retries = v
	}
	if f := flags.Lookup("public-ip-provider"); f != nil && f.Changed {
		v, _ := flags.GetStringSlice("public-ip-provider")
		cfg.Net.PublicIPProviders = v
	}
	if f := flags.Lookup("speedtest-bin"); f != nil && f.Changed {
		v, _ := flags.GetString("speedtest-bin")
		cfg.Speedtest.Bin = v
	}
	if f := flags.Lookup("kubeconfig"); f != nil && f.Changed {
		v, _ := flags.GetString("kubeconfig")
		cfg.K8s.Kubeconfig = v
	}
}

func SafeMap(cfg Config, path string) map[string]any {
	out := map[string]any{
		"config_path": path,
		"net": map[string]any{
			"public_ip_providers": cfg.Net.PublicIPProviders,
			"timeout_seconds":     cfg.Net.TimeoutSeconds,
			"retries":             cfg.Net.Retries,
		},
		"speedtest": map[string]any{
			"bin": cfg.Speedtest.Bin,
		},
		"k8s": map[string]any{
			"kubeconfig": cfg.K8s.Kubeconfig,
		},
		"secrets": map[string]any{
			"registry_token": mask("registry_token", cfg.Secrets.RegistryToken),
			"api_token":      mask("api_token", cfg.Secrets.APIToken),
		},
	}
	return out
}

func mask(key, value string) string {
	if common.ShouldMaskKey(key) {
		return common.MaskValue(value)
	}
	return value
}
