package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
	date    string
)

var rootCmd = &cobra.Command{
	Use:   "muxi",
	Short: "MUXI CLI - Formation development and server management",
	Long: `MUXI CLI - Formation development and server management

Build, deploy, and manage MUXI formations with ease.

Examples:
  muxi new formation my-bot
  muxi deploy --profile production
  muxi agent list`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Add --version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(fmt.Sprintf("muxi version %s\n", version))

	// Add subcommands here as we build them
	rootCmd.AddCommand(newCmd)
}

// SetVersionInfo sets version information from main
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = v
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
