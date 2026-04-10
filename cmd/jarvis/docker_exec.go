package main

import (
	"fmt"
	"strings"

	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/dockerutil"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newDockerExecCmd(state *app.State) *cobra.Command {
	var preferredShell string

	cmd := &cobra.Command{
		Use:   "exec <container>",
		Short: "Open interactive shell inside running container",
		Long: `Open interactive shell inside running container.

Shell selection priority:
- bash
- sh
- ash

If --shell is provided, it is tried first, then fallback list.

Return codes:
- 0: shell session exited successfully
- 1: container not running, no usable shell, or docker exec failure`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.JSON {
				return common.NewExitError(common.ExitError, "--json is not supported for interactive docker exec", nil)
			}
			if !term.IsTerminal(0) || !term.IsTerminal(1) {
				return common.NewExitError(common.ExitError, "interactive terminal required for docker exec", nil)
			}

			container := args[0]
			running, err := dockerutil.RunningContainers()
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot list running containers", err)
			}
			if !strContains(running, container) {
				return common.NewExitError(common.ExitError, fmt.Sprintf("container %q is not running", container), nil)
			}

			order := shellOrder(preferredShell)
			shell, err := dockerutil.PickAvailableShell(container, order)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot determine usable shell in container", err)
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Opening %s in %s...\n", shell, container)
			if err := dockerutil.ExecInteractive(container, shell); err != nil {
				return common.NewExitError(common.ExitError, "docker exec failed", err)
			}
			return nil
		},
		Example: "jarvis docker exec my-container\njarvis docker exec api --shell sh",
		ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			containers, err := dockerutil.RunningContainers()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			res := make([]string, 0, len(containers))
			for _, c := range containers {
				if strings.HasPrefix(c, toComplete) {
					res = append(res, c)
				}
			}
			return res, cobra.ShellCompDirectiveNoFileComp
		},
	}
	cmd.Flags().StringVarP(&preferredShell, "shell", "s", "", "Preferred shell to try first (fallback: bash, sh, ash)")
	return cmd
}

func shellOrder(preferred string) []string {
	base := []string{"bash", "sh", "ash"}
	if strings.TrimSpace(preferred) == "" {
		return base
	}
	out := []string{preferred}
	for _, x := range base {
		if x != preferred {
			out = append(out, x)
		}
	}
	return out
}

func strContains(items []string, target string) bool {
	for _, x := range items {
		if x == target {
			return true
		}
	}
	return false
}
