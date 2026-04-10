package main

import (
	"fmt"
	"strconv"

	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/sys"

	"github.com/spf13/cobra"
)

func newSysCmd(state *app.State) *cobra.Command {
	sysCmd := &cobra.Command{
		Use:     "sys",
		Short:   "System utility commands",
		Long:    "System utility commands for secure password generation and local OS diagnostics.",
		Example: "jarvis sys password\njarvis sys hosts take-control\njarvis sys hosts add 10.10.10.10 internal.local",
	}
	sysCmd.AddCommand(newSysPasswordCmd(state), newSysHostsCmd(state), newSysWCmd(state))
	return sysCmd
}

func newSysPasswordCmd(state *app.State) *cobra.Command {
	var length int
	var lower, upper, digits, symbols bool
	var profile string
	var noAmbiguous bool

	cmd := &cobra.Command{
		Use:   "password",
		Short: "Generate a random password",
		Long: `Generate cryptographically secure passwords for infra and user accounts.

Profiles:
- infra: >=24 chars, includes lower/upper/digits/symbols.
- human: >=16 chars, excludes symbols and ambiguous characters.
- strict: >=32 chars, all sets, no ambiguous characters.

Return codes:
- 0: password generated
- 1: invalid options or generation failure`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := sys.PasswordOptions{
				Length:      length,
				Lower:       lower,
				Upper:       upper,
				Digits:      digits,
				Symbols:     symbols,
				Profile:     profile,
				NoAmbiguous: noAmbiguous,
			}
			if profile == "" && !lower && !upper && !digits && !symbols {
				opts.Lower, opts.Upper, opts.Digits = true, true, true
			}
			pwd, err := sys.GeneratePassword(opts)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot generate password", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"password": pwd, "length": len(pwd), "profile": profile})
			}
			fmt.Fprintln(cmd.OutOrStdout(), pwd)
			return nil
		},
		Example: "jarvis sys password\njarvis sys password --profile strict --length 40 --no-ambiguous",
	}

	cmd.Flags().IntVarP(&length, "length", "l", 24, "Password length")
	cmd.Flags().BoolVarP(&lower, "lower", "L", false, "Include lowercase letters")
	cmd.Flags().BoolVarP(&upper, "upper", "U", false, "Include uppercase letters")
	cmd.Flags().BoolVarP(&digits, "digits", "D", false, "Include digits")
	cmd.Flags().BoolVarP(&symbols, "symbols", "S", false, "Include symbols")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Preset profile: infra|human|strict")
	cmd.Flags().BoolVarP(&noAmbiguous, "no-ambiguous", "A", false, "Exclude ambiguous characters like O/0/I/l")
	return cmd
}

func newSysWCmd(state *app.State) *cobra.Command {
	var top int

	cmd := &cobra.Command{
		Use:   "w",
		Short: "Show system session/load, memory and heavy processes",
		Long: `Aggregate output of:
- w (user sessions)
- load average
- free -m (Linux only)
- top N heaviest processes by RSS

Return codes:
- 0: data collected
- 1: command execution failed`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			snapshot, err := sys.CollectWSnapshot(top)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot collect system snapshot", err)
			}

			if state.JSON {
				return state.Printer.PrintJSON(snapshot)
			}

			if len(snapshot.LoadAverage) == 3 {
				fmt.Fprintf(cmd.OutOrStdout(), "Load Average: %.2f %.2f %.2f\n\n", snapshot.LoadAverage[0], snapshot.LoadAverage[1], snapshot.LoadAverage[2])
			}

			fmt.Fprintln(cmd.OutOrStdout(), "w:")
			fmt.Fprintln(cmd.OutOrStdout(), snapshot.WOutput)
			if snapshot.WMessage != "" {
				fmt.Fprintln(cmd.OutOrStdout(), snapshot.WMessage)
			}
			fmt.Fprintln(cmd.OutOrStdout())

			fmt.Fprintln(cmd.OutOrStdout(), "free -m:")
			if snapshot.FreeSupported {
				fmt.Fprintln(cmd.OutOrStdout(), snapshot.FreeOutput)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), snapshot.FreeMessage)
			}
			fmt.Fprintln(cmd.OutOrStdout())

			rows := make([][]string, 0, len(snapshot.Processes))
			for _, p := range snapshot.Processes {
				rows = append(rows, []string{
					strconv.Itoa(p.PID),
					p.User,
					p.Command,
					fmt.Sprintf("%.1f", p.CPU),
					fmt.Sprintf("%.1f", p.MemPct),
					strconv.FormatInt(p.RSSKB, 10),
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Top %d Processes by RSS (KB):\n", len(rows))
			state.Printer.PrintTable([]string{"PID", "USER", "COMMAND", "CPU%", "MEM%", "RSS KB"}, rows)
			return nil
		},
		Example: "jarvis sys w\njarvis sys w --top 15 --json",
	}
	cmd.Flags().IntVarP(&top, "top", "W", 15, "Number of heavy processes to show")
	return cmd
}
