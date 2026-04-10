package main

import (
	"fmt"
	"os"
	"strconv"

	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/dockerutil"

	"github.com/spf13/cobra"
)

func newDockerCmd(state *app.State) *cobra.Command {
	dockerCmd := &cobra.Command{
		Use:     "docker",
		Short:   "Local Docker helpers",
		Long:    "Local Docker helper commands for images listing and safe cleanup.",
		Example: "jarvis docker images\njarvis docker exec my-container\njarvis docker prune --dangling --dry-run",
	}
	dockerCmd.AddCommand(newDockerImagesCmd(state), newDockerExecCmd(state), newDockerPruneCmd(state))
	return dockerCmd
}

func newDockerImagesCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "List local Docker images",
		Long: `List local Docker images with repository tag, size and creation time.

Return codes:
- 0: images listed
- 1: docker unavailable or command failed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			images, err := dockerutil.Images()
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot list docker images", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(images)
			}
			rows := make([][]string, 0, len(images))
			for _, img := range images {
				rows = append(rows, []string{img.Name, img.Size, img.Created})
			}
			state.Printer.PrintTable([]string{"Image", "Size", "Created"}, rows)
			return nil
		},
		Example: "jarvis docker images\njarvis docker images --json",
	}
	return cmd
}

func newDockerPruneCmd(state *app.State) *cobra.Command {
	var dangling bool
	var dryRun bool
	var force bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune Docker images",
		Long: `Prune dangling Docker images.

Safety:
- Use --dry-run to inspect candidates.
- Requires confirmation unless --force is set.

Return codes:
- 0: prune completed or dry-run reported
- 1: docker failure`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dangling {
				return common.NewExitError(common.ExitError, "only --dangling mode is supported in MVP", nil)
			}
			if !dryRun && !force {
				fmt.Fprint(os.Stderr, "Proceed with docker image prune? [y/N]: ")
				var answer string
				_, _ = fmt.Fscanln(cmd.InOrStdin(), &answer)
				if answer != "y" && answer != "Y" {
					return common.NewExitError(common.ExitError, "aborted by user", nil)
				}
			}

			res, err := dockerutil.PruneDangling(dryRun)
			if err != nil {
				return common.NewExitError(common.ExitError, "docker prune failed", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(res)
			}
			state.Printer.PrintKV(map[string]string{
				"candidates": strconv.Itoa(res.Candidates),
				"deleted":    strconv.Itoa(res.Deleted),
				"dry_run":    strconv.FormatBool(res.DryRun),
			})
			if res.Output != "" {
				fmt.Fprintln(os.Stdout, res.Output)
			}
			return nil
		},
		Example: "jarvis docker prune --dangling --dry-run\njarvis docker prune --dangling --force",
	}
	cmd.Flags().BoolVarP(&dangling, "dangling", "d", false, "Prune dangling images only")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be pruned")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Run without confirmation")
	return cmd
}
