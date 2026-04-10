package main

import (
	"fmt"
	"strings"

	"jarvis/internal/app"
	"jarvis/internal/config"
	"jarvis/internal/logger"
	"jarvis/internal/output"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	state := &app.State{}

	root := &cobra.Command{
		Use:           "jarvis",
		Short:         "Jarvis is a daily ops CLI for network, system, Docker, Kubernetes and data tasks",
		Long:          "Jarvis is a pragmatic CLI for daily engineering tasks: network diagnostics, system helpers, Docker, Kubernetes and data utilities. It supports both interactive and automation workflows.",
		Example:       "jarvis net ip --public\njarvis sys password --profile strict --length 40\njarvis k8s pods --namespace prod --restarts",
		Version:       buildVersionString(),
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, _ := cmd.Flags().GetString("config")
			cfg, resolvedPath, err := config.Load(cfgPath, cmd.Flags())
			if err != nil {
				return err
			}

			jsonOut, _ := cmd.Flags().GetBool("json")
			noColor, _ := cmd.Flags().GetBool("no-color")
			verbose, _ := cmd.Flags().GetBool("verbose")
			debug, _ := cmd.Flags().GetBool("debug")
			quiet, _ := cmd.Flags().GetBool("quiet")

			level := logger.LevelWarn
			switch {
			case quiet:
				level = logger.LevelError
			case debug:
				level = logger.LevelDebug
			case verbose:
				level = logger.LevelInfo
			}

			state.Config = cfg
			state.ConfigPath = resolvedPath
			state.JSON = jsonOut
			state.NoColor = noColor
			state.Printer = output.New(jsonOut, noColor)
			state.Logger = logger.New(level)
			return nil
		},
	}

	root.PersistentFlags().BoolP("json", "j", false, "Output command result in JSON")
	root.PersistentFlags().BoolP("no-color", "C", false, "Disable colored output")
	root.PersistentFlags().BoolP("verbose", "v", false, "Enable info logs")
	root.PersistentFlags().BoolP("debug", "d", false, "Enable debug logs")
	root.PersistentFlags().BoolP("quiet", "q", false, "Only print errors")
	root.PersistentFlags().StringP("config", "c", "", "Config file path (default: ~/.config/jarvis/config.yaml)")
	root.PersistentFlags().IntP("timeout", "t", 0, "Global network timeout in seconds (overrides config/env)")
	root.PersistentFlags().IntP("retries", "r", 0, "Global network retry count (overrides config/env)")
	root.PersistentFlags().StringSliceP("public-ip-provider", "P", nil, "Public IP provider URL(s) (overrides config/env)")
	root.PersistentFlags().StringP("speedtest-bin", "b", "", "Speedtest binary name or path (overrides config/env)")
	root.PersistentFlags().StringP("kubeconfig", "k", "", "Path to kubeconfig (overrides env/config)")

	root.SetHelpCommand(helpAliasCmd(root))

	root.AddCommand(newSysCmd(state))
	root.AddCommand(newNetCmd(state))
	root.AddCommand(newDockerCmd(state))
	root.AddCommand(newFSCmd(state))
	root.AddCommand(newK8sCmd(state))
	root.AddCommand(newDataCmd(state))
	root.AddCommand(newCatCmd(state))
	root.AddCommand(newScreensaverCmd(state))
	root.AddCommand(newConfigCmd(state))
	root.AddCommand(newJumpCmd())
	root.AddCommand(newCompletionCmd(root))

	return root
}

func helpAliasCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "help [command]",
		Short: "Show help for jarvis or a specific command",
		Long:  "Show detailed help for jarvis, a domain, or a subcommand. Use this as an alias for --help in scripts and interactive sessions.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return root.Help()
			}
			found, _, err := root.Find(args)
			if err != nil {
				return fmt.Errorf("cannot find help topic %q: %w", strings.Join(args, " "), err)
			}
			return found.Help()
		},
		Example: "jarvis help\njarvis help net ip",
	}
}

func newCompletionCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion script",
		Long:  "Generate shell completion script for bash, zsh, or fish. Source it in your shell profile for command and flag completion.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			switch shell {
			case "bash":
				return root.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			default:
				return fmt.Errorf("unsupported shell %q. expected bash|zsh|fish", shell)
			}
		},
		Example: "jarvis completion bash > /tmp/jarvis.bash\njarvis completion zsh > ~/.zfunc/_jarvis",
	}
	return cmd
}
