package main

import (
	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/screensaver"

	"github.com/spf13/cobra"
)

func newScreensaverCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screensaver",
		Short: "Matrix-style terminal screensaver with date/time",
		Long: `Run matrix-style screensaver in terminal with current date/time overlay.

Controls:
- q: quit
- Ctrl-C: quit

Return codes:
- 0: exited by user
- 1: terminal unsupported or runtime error`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := screensaver.Run(screensaver.Options{NoColor: state.NoColor}); err != nil {
				return common.NewExitError(common.ExitError, "screensaver failed", err)
			}
			return nil
		},
		Example: "jarvis screensaver\njarvis screensaver --no-color",
	}
	return cmd
}
