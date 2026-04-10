package main

import (
	"fmt"
	"os"
	"strings"

	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/sys"

	"github.com/spf13/cobra"
)

func newSysHostsCmd(state *app.State) *cobra.Command {
	var hostsPath string
	var marker string

	cmd := &cobra.Command{
		Use:   "hosts",
		Short: "Manage jarvis-controlled section in hosts file",
		Long: `Manage a dedicated marker section in hosts file (default /etc/hosts).

Workflow:
1) take-control adds marker to file end
2) add/delete/cat manage entries after marker
3) disable/enable comments or uncomments managed entries
4) clean removes all lines after marker`,
		Example: "jarvis sys hosts take-control\njarvis sys hosts add 10.10.10.10 internal.local",
	}
	cmd.PersistentFlags().StringVarP(&hostsPath, "hosts-file", "H", "/etc/hosts", "Hosts file path")
	cmd.PersistentFlags().StringVarP(&marker, "marker", "m", sys.DefaultHostsMarker, "Marker comment to separate jarvis-managed block")

	manager := func() *sys.HostsManager {
		return sys.NewHostsManager(hostsPath, marker)
	}

	cmd.AddCommand(
		newSysHostsTakeControlCmd(state, manager),
		newSysHostsDisableCmd(state, manager),
		newSysHostsEnableCmd(state, manager),
		newSysHostsAddCmd(state, manager),
		newSysHostsDeleteCmd(state, manager),
		newSysHostsCatCmd(state, manager),
		newSysHostsCleanCmd(state, manager),
	)
	return cmd
}

func newSysHostsTakeControlCmd(state *app.State, manager func() *sys.HostsManager) *cobra.Command {
	return &cobra.Command{
		Use:   "take-control",
		Short: "Append jarvis marker to hosts file",
		Long: `Append marker comment to end of hosts file if missing.

Return codes:
- 0: marker exists or created
- 1: file access failure`,
		RunE: func(_ *cobra.Command, _ []string) error {
			changed, err := manager().TakeControl()
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot take control of hosts file", err)
			}
			status := "marker already present"
			if changed {
				status = "marker appended"
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"status": status, "changed": changed})
			}
			fmt.Fprintln(os.Stdout, status)
			return nil
		},
		Example: "jarvis sys hosts take-control\njarvis sys hosts take-control --hosts-file /tmp/hosts",
	}
}

func newSysHostsDisableCmd(state *app.State, manager func() *sys.HostsManager) *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Comment out all lines after marker",
		Long: `Comment out every non-empty managed line after marker.

Return codes:
- 0: section disabled
- 1: marker missing or file access failure`,
		RunE: func(_ *cobra.Command, _ []string) error {
			count, err := manager().DisableAll()
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot disable hosts entries", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"changed": count})
			}
			fmt.Fprintf(os.Stdout, "Disabled %d line(s)\n", count)
			return nil
		},
		Example: "jarvis sys hosts disable\njarvis sys hosts disable --hosts-file /tmp/hosts",
	}
}

func newSysHostsEnableCmd(state *app.State, manager func() *sys.HostsManager) *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Uncomment all lines after marker",
		Long: `Uncomment every line after marker.

Return codes:
- 0: section enabled
- 1: marker missing or file access failure`,
		RunE: func(_ *cobra.Command, _ []string) error {
			count, err := manager().EnableAll()
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot enable hosts entries", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"changed": count})
			}
			fmt.Fprintf(os.Stdout, "Enabled %d line(s)\n", count)
			return nil
		},
		Example: "jarvis sys hosts enable\njarvis sys hosts enable --hosts-file /tmp/hosts",
	}
}

func newSysHostsAddCmd(state *app.State, manager func() *sys.HostsManager) *cobra.Command {
	return &cobra.Command{
		Use:   "add <address> <hostname>",
		Short: "Add entry after marker",
		Long: `Add address/hostname pair into jarvis-managed block.

Return codes:
- 0: entry added or already exists
- 1: marker missing or file access failure`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			changed, err := manager().Add(args[0], args[1])
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot add hosts entry", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"changed": changed, "address": args[0], "hostname": args[1]})
			}
			if changed {
				fmt.Fprintf(os.Stdout, "Added %s %s\n", args[0], args[1])
			} else {
				fmt.Fprintf(os.Stdout, "Entry already exists: %s %s\n", args[0], args[1])
			}
			return nil
		},
		Example: "jarvis sys hosts add 127.0.0.1 local.test\njarvis sys hosts add 10.0.0.5 dev.internal --hosts-file /tmp/hosts",
	}
}

func newSysHostsDeleteCmd(state *app.State, manager func() *sys.HostsManager) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <address> <hostname>",
		Short: "Delete entry after marker",
		Long: `Delete all matching address/hostname pairs from managed block.

Return codes:
- 0: delete completed (or nothing to delete)
- 1: marker missing or file access failure`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			deleted, err := manager().Delete(args[0], args[1])
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot delete hosts entry", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"deleted": deleted, "address": args[0], "hostname": args[1]})
			}
			fmt.Fprintf(os.Stdout, "Deleted %d line(s) for %s %s\n", deleted, args[0], args[1])
			return nil
		},
		Example: "jarvis sys hosts delete 127.0.0.1 local.test\njarvis sys hosts delete 10.0.0.5 dev.internal --hosts-file /tmp/hosts",
	}
}

func newSysHostsCatCmd(state *app.State, manager func() *sys.HostsManager) *cobra.Command {
	return &cobra.Command{
		Use:   "cat",
		Short: "Show managed entries after marker",
		Long: `Display parsed hosts entries located after marker.

Return codes:
- 0: entries printed
- 1: marker missing or file access failure`,
		RunE: func(_ *cobra.Command, _ []string) error {
			entries, err := manager().Entries()
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot read managed hosts entries", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"entries": entries})
			}
			if len(entries) == 0 {
				fmt.Fprintln(os.Stdout, "No entries after marker")
				return nil
			}
			rows := make([][]string, 0, len(entries))
			for _, e := range entries {
				rows = append(rows, []string{e.Address, e.Hostname, map[bool]string{true: "enabled", false: "disabled"}[e.Enabled]})
			}
			state.Printer.PrintTable([]string{"Address", "Hostname", "State"}, rows)
			return nil
		},
		Example: "jarvis sys hosts cat\njarvis sys hosts cat --json",
	}
}

func newSysHostsCleanCmd(state *app.State, manager func() *sys.HostsManager) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove all lines after marker",
		Long: `Remove entire managed block body (all lines after marker).

Safety:
- asks for confirmation unless --force is set.

Return codes:
- 0: clean completed
- 1: marker missing, aborted, or file access failure`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if !force {
				fmt.Fprint(os.Stderr, "Clean all managed hosts entries after marker? [y/N]: ")
				var answer string
				_, _ = fmt.Fscanln(os.Stdin, &answer)
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					return common.NewExitError(common.ExitError, "aborted by user", nil)
				}
			}
			removed, err := manager().Clean()
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot clean managed hosts entries", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"removed": removed})
			}
			fmt.Fprintf(os.Stdout, "Removed %d line(s)\n", removed)
			return nil
		},
		Example: "jarvis sys hosts clean --force\njarvis sys hosts clean --hosts-file /tmp/hosts",
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Run without confirmation")
	return cmd
}
