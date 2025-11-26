package main

import (
	"fmt"
	"os"

	"github.com/muxi-ai/cli/cmd"
)

var (
	// Version info - set via ldflags during build
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version info in root command
	cmd.SetVersionInfo(version, commit, date)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
