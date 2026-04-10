package main

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func buildVersionString() string {
	return fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
}
