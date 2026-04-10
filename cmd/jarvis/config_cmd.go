package main

import (
	"fmt"
	"os"

	"jarvis/internal/app"

	"github.com/spf13/cobra"
)

func newConfigCmd(state *app.State) *cobra.Command {
	cfg := &cobra.Command{
		Use:     "config",
		Short:   "Configuration commands",
		Long:    "Inspect effective configuration and config file location.",
		Example: "jarvis config show\njarvis config path",
	}
	cfg.AddCommand(newConfigShowCmd(state), newConfigPathCmd(state))
	return cfg
}

func newConfigShowCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show effective configuration (secrets masked)",
		Long: `Show effective configuration built from file, environment, and flags.

Secrets are masked in output.

Return codes:
- 0: config printed
- 1: print failure`,
		RunE: func(_ *cobra.Command, _ []string) error {
			safe := map[string]any{
				"config_path": state.ConfigPath,
				"net": map[string]any{
					"public_ip_providers": state.Config.Net.PublicIPProviders,
					"timeout_seconds":     state.Config.Net.TimeoutSeconds,
					"retries":             state.Config.Net.Retries,
				},
				"speedtest": map[string]any{"bin": state.Config.Speedtest.Bin},
				"k8s":       map[string]any{"kubeconfig": state.Config.K8s.Kubeconfig},
				"secrets": map[string]any{
					"registry_token": "********",
					"api_token":      "********",
				},
			}
			if state.JSON {
				return state.Printer.PrintJSON(safe)
			}
			return state.Printer.PrintJSON(safe)
		},
		Example: "jarvis config show\njarvis config show --json",
	}
	return cmd
}

func newConfigPathCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show resolved config file path",
		Long: `Show the config path jarvis resolved for this run.

Return codes:
- 0: path printed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Fprintln(os.Stdout, state.ConfigPath)
			return nil
		},
		Example: "jarvis config path\njarvis --config /tmp/jarvis.yaml config path",
	}
	return cmd
}
