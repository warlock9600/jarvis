package main

import (
	"fmt"
	"os"

	"jarvis/internal/common"
)

func main() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(common.ExitCode(err))
	}
}
