package main

import (
	"fmt"
	"os"
	"strings"

	"jarvis/internal/app"
	"jarvis/internal/catutil"
	"jarvis/internal/common"

	"github.com/spf13/cobra"
)

func newCatCmd(state *app.State) *cobra.Command {
	var plain bool
	var style string

	cmd := &cobra.Command{
		Use:   "cat <file> [file...]",
		Short: "Show file with automatic syntax highlighting",
		Long: `Show file content with automatic syntax detection and terminal highlighting when possible.

Highlighting is enabled in TTY by default and disabled for --json, --no-color or --plain.

Return codes:
- 0: file(s) printed
- 1: read/highlight errors`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if state.JSON {
				views := make([]catutil.FileView, 0, len(args))
				for _, p := range args {
					v, err := catutil.ReadFile(p)
					if err != nil {
						return common.NewExitError(common.ExitError, "cannot read file", err)
					}
					views = append(views, v)
				}
				return state.Printer.PrintJSON(map[string]any{"files": views})
			}

			outputs := make([]string, 0, len(args))
			for i, p := range args {
				v, err := catutil.ReadFile(p)
				if err != nil {
					return common.NewExitError(common.ExitError, "cannot read file", err)
				}
				if v.Binary && !plain {
					return common.NewExitError(common.ExitError, fmt.Sprintf("%s looks like a binary file; use --plain to print raw content", p), nil)
				}

				body := v.Content
				if !plain && !state.NoColor && state.Printer.IsTTY && !v.Binary {
					h, err := catutil.RenderHighlighted(v, style, false)
					if err == nil {
						body = h
					}
				}
				if len(args) > 1 {
					header := fmt.Sprintf("==> %s <==", p)
					if i == 0 {
						outputs = append(outputs, header+"\n"+strings.TrimRight(body, "\n"))
					} else {
						outputs = append(outputs, "\n"+header+"\n"+strings.TrimRight(body, "\n"))
					}
				} else {
					outputs = append(outputs, body)
				}
			}

			_, _ = fmt.Fprint(os.Stdout, catutil.JoinWithHeader(outputs))
			if len(outputs) == 1 && !strings.HasSuffix(outputs[0], "\n") {
				_, _ = fmt.Fprintln(os.Stdout)
			}
			return nil
		},
		Example: "jarvis cat main.go\njarvis cat config.yaml --style dracula\njarvis cat file1.go file2.go",
	}

	cmd.Flags().BoolVarP(&plain, "plain", "p", false, "Disable syntax highlighting")
	cmd.Flags().StringVarP(&style, "style", "s", "monokai", "Highlight style name")
	return cmd
}
