package main

import (
	"fmt"
	"io"
	"os"

	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/datautil"

	"github.com/spf13/cobra"
)

func newDataCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "data",
		Short:   "Data transformation helpers",
		Long:    "Data transformation helpers for base64 and JWT payload inspection.",
		Example: "jarvis data b64 encode --input 'hello'\njarvis data jwt decode <token>",
	}
	cmd.AddCommand(newDataB64Cmd(state), newDataJWTCmd(state))
	return cmd
}

func newDataB64Cmd(state *app.State) *cobra.Command {
	var input string
	b64 := &cobra.Command{
		Use:     "b64",
		Short:   "Base64 encode/decode",
		Long:    "Base64 helpers for encoding and decoding either --input value or stdin stream.",
		Example: "jarvis data b64 encode --input secret\ncat token.b64 | jarvis data b64 decode",
	}

	encode := &cobra.Command{
		Use:   "encode",
		Short: "Encode text to base64",
		Long: `Encode plain text to base64.

Return codes:
- 0: encoded
- 1: input read failure`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			val, err := getInput(input, cmd)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot read input", err)
			}
			out := datautil.B64Encode(val)
			if state.JSON {
				return state.Printer.PrintJSON(map[string]string{"encoded": out})
			}
			fmt.Fprintln(os.Stdout, out)
			return nil
		},
		Example: "jarvis data b64 encode --input 'hello'\necho -n 'hello' | jarvis data b64 encode",
	}
	encode.Flags().StringVarP(&input, "input", "i", "", "Input value; if empty, stdin is used")

	decode := &cobra.Command{
		Use:   "decode",
		Short: "Decode base64 to text",
		Long: `Decode base64 input.

Return codes:
- 0: decoded
- 1: invalid base64 or read failure`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			val, err := getInput(input, cmd)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot read input", err)
			}
			out, err := datautil.B64Decode(val)
			if err != nil {
				return common.NewExitError(common.ExitError, "invalid base64", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]string{"decoded": out})
			}
			fmt.Fprintln(os.Stdout, out)
			return nil
		},
		Example: "jarvis data b64 decode --input aGVsbG8=\necho aGVsbG8= | jarvis data b64 decode",
	}
	decode.Flags().StringVarP(&input, "input", "i", "", "Input value; if empty, stdin is used")

	b64.AddCommand(encode, decode)
	return b64
}

func newDataJWTCmd(state *app.State) *cobra.Command {
	jwtCmd := &cobra.Command{
		Use:     "jwt",
		Short:   "JWT token helpers",
		Long:    "JWT token helpers for non-validating decoding of header and payload.",
		Example: "jarvis data jwt decode <token>\njarvis data jwt decode <token> --json",
	}
	jwtCmd.AddCommand(newDataJWTDecodeCmd(state))
	return jwtCmd
}

func newDataJWTDecodeCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode <token>",
		Short: "Decode JWT header and payload without signature validation",
		Long: `Decode JWT token and print header/payload as pretty JSON.

Security note:
- Signature is not validated by default.

Return codes:
- 0: token decoded
- 1: invalid format`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			decoded, err := datautil.DecodeJWT(args[0])
			if err != nil {
				return common.NewExitError(common.ExitError, "JWT decode failed", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(decoded)
			}
			fmt.Fprintln(os.Stdout, "Header:")
			if err := state.Printer.PrintJSON(decoded.Header); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Payload:")
			return state.Printer.PrintJSON(decoded.Payload)
		},
		Example: "jarvis data jwt decode eyJ...\njarvis data jwt decode eyJ... --json",
	}
	return cmd
}

func getInput(flagValue string, cmd *cobra.Command) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	b, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return "", err
	}
	return string(b), nil
}
