package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"jarvis/internal/common"
	"jarvis/internal/jump"

	"github.com/spf13/cobra"
)

func newJumpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jump <host> [ssh_args...]",
		Short: "Connect to SSH host from ~/.ssh/config and ~/.ssh/config.d",
		Long: `Discover SSH hosts from ~/.ssh/config and ~/.ssh/config.d/*, offer shell completion,
then open SSH connection to selected host.

Return codes:
- 0: ssh session ended successfully
- 1: host not found, ssh failure, or config scan failure`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			host := args[0]
			known, err := jump.DiscoverHosts("")
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot discover SSH hosts", err)
			}
			if !contains(known, host) {
				return common.NewExitError(common.ExitError, fmt.Sprintf("unknown SSH host %q (not found in ~/.ssh/config or ~/.ssh/config.d/*)", host), nil)
			}

			sshArgs := []string{host}
			if len(args) > 1 {
				sshArgs = append(sshArgs, args[1:]...)
			}
			sshCmd := exec.Command("ssh", sshArgs...)
			sshCmd.Stdin = os.Stdin
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			if err := sshCmd.Run(); err != nil {
				return common.NewExitError(common.ExitError, "ssh connection failed", err)
			}
			return nil
		},
		Example: "jarvis jump bastion\njarvis jump app-prod -L 8080:localhost:8080",
		ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			hosts, err := jump.DiscoverHosts("")
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			filtered := make([]string, 0, len(hosts))
			for _, h := range hosts {
				if strings.HasPrefix(h, toComplete) {
					filtered = append(filtered, h)
				}
			}
			return filtered, cobra.ShellCompDirectiveNoFileComp
		},
	}
	return cmd
}

func contains(items []string, x string) bool {
	for _, item := range items {
		if item == x {
			return true
		}
	}
	return false
}
