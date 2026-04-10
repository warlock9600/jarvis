package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/fsutil"

	"github.com/spf13/cobra"
	"github.com/jedib0t/go-pretty/v6/text"
)

func newFSCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fs",
		Short:   "Filesystem utilities",
		Long:    "Filesystem utilities for rich directory listings and automation-friendly output.",
		Example: "jarvis fs show .\njarvis fs show /var/log --sort time --recent 20 --json",
	}
	cmd.AddCommand(newFSShowCmd(state))
	return cmd
}

func newFSShowCmd(state *app.State) *cobra.Command {
	var sortBy string
	var reverse bool
	var hidden bool
	var typeFilter string
	var exts []string
	var largest int
	var recent int
	var tree bool
	var depth int
	var gitStatus bool

	cmd := &cobra.Command{
		Use:   "show [path]",
		Short: "Enhanced listing (replacement for ls alias)",
		Long: `Enhanced listing command with sorting, filtering, tree mode and optional git status.

Return codes:
- 0: listed successfully
- 1: listing failed`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := "."
			if len(args) == 1 {
				path = args[0]
			}
			entries, err := fsutil.Show(fsutil.Options{
				Path:      path,
				SortBy:    sortBy,
				Reverse:   reverse,
				Hidden:    hidden,
				Type:      typeFilter,
				Ext:       exts,
				Largest:   largest,
				Recent:    recent,
				Tree:      tree,
				Depth:     depth,
				GitStatus: gitStatus,
			})
			if err != nil {
				return common.NewExitError(common.ExitError, "fs show failed", err)
			}

			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"path": path, "entries": entries})
			}

			rows := make([][]string, 0, len(entries))
			for _, e := range entries {
				name := e.Name
				if tree {
					name = strings.Repeat("  ", e.Depth-1) + "- " + e.Name
				}
				name = colorizeFSName(state, e, name)
				rows = append(rows, []string{e.Mode, fsutil.HumanSize(e.SizeBytes), e.Modified.Format("2006-01-02 15:04:05"), name})
			}

			fmt.Fprintf(os.Stdout, "Path: %s\n", path)
			for _, r := range rows {
				fmt.Fprintf(os.Stdout, "%-11s %8s %s %s\n", r[0], r[1], r[2], r[3])
			}
			return nil
		},
		Example: "jarvis fs show .\njarvis fs show . --sort size --largest 15 --git-status\njarvis fs show . --tree --depth 2",
	}

	cmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "Sort by: name|size|time")
	cmd.Flags().BoolVarP(&reverse, "reverse", "o", false, "Reverse sort order")
	cmd.Flags().BoolVarP(&hidden, "hidden", "a", false, "Include hidden files")
	cmd.Flags().StringVarP(&typeFilter, "type", "y", "all", "Filter type: all|file|dir|link")
	cmd.Flags().StringSliceVarP(&exts, "ext", "e", nil, "Filter by extension(s), e.g. --ext go,yaml")
	cmd.Flags().IntVarP(&largest, "largest", "L", 0, "Show N largest entries")
	cmd.Flags().IntVarP(&recent, "recent", "R", 0, "Show N most recently modified entries")
	cmd.Flags().BoolVarP(&tree, "tree", "T", false, "Tree-like view")
	cmd.Flags().IntVarP(&depth, "depth", "D", 2, "Tree max depth (used with --tree)")
	cmd.Flags().BoolVarP(&gitStatus, "git-status", "g", false, "Show git status if inside repository")
	return cmd
}

func colorizeFSName(state *app.State, e fsutil.Entry, rendered string) string {
	if state.NoColor || !state.Printer.IsTTY {
		return rendered
	}
	if strings.HasPrefix(e.Name, ".") {
		return text.Colors{text.FgHiBlack}.Sprint(rendered)
	}
	if e.Type == "dir" {
		return text.Colors{text.FgGreen}.Sprint(rendered)
	}
	ext := strings.ToLower(filepath.Ext(e.Name))
	switch ext {
	case ".sh":
		return text.Colors{text.FgMagenta}.Sprint(rendered)
	}
	if e.Executable {
		return text.Colors{text.FgRed}.Sprint(rendered)
	}
	return rendered
}
