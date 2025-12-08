package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/muxi-ai/cli/cmd"
)

//go:embed .version
var embeddedVersion string

var (
	// Version info - can be overridden via ldflags during build
	version = ""
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Use embedded version from .version file if not set via ldflags
	if version == "" {
		version = strings.TrimSpace(embeddedVersion)
	}

	// Set version info in root command
	cmd.SetVersionInfo(version, commit, date)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
