package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/muxi-ai/cli/cmd"
	"github.com/muxi-ai/cli/pkg/scaffold"
	"github.com/muxi-ai/cli/pkg/server"
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

	// Set version info in root command and packages
	cmd.SetVersionInfo(version, commit, date)
	scaffold.Version = version
	server.Version = version

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
