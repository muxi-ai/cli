package cmd

import (
	"fmt"
	"net/url"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/telemetry"
	"github.com/muxi-ai/cli/pkg/ui"

	"github.com/spf13/cobra"
)

const consoleLaunchURL = "https://muxi.dev/launch"

var consoleCmd = &cobra.Command{
	Use:     "console",
	Short:   "Open MUXI Console in browser",
	GroupID: "formation",
	Long: `Open the MUXI Console web interface in your default browser.

When inside a formation directory with a server profile configured:
  Opens the Console pre-configured for that formation.

When NOT in a formation directory or no server configured:
  Opens the Console homepage.`,
	Example: `  # From formation directory
  cd my-bot
  muxi console

  # From anywhere
  muxi console`,
	RunE: runConsole,
}

func init() {
	rootCmd.AddCommand(consoleCmd)
	consoleCmd.Flags().StringP("profile", "p", "", "Server profile to use")
}

func runConsole(cmd *cobra.Command, args []string) error {
	profileFlag, _ := cmd.Flags().GetString("profile")

	// Check if we're in a formation directory
	ctx, ctxErr := context.DetectFormation()
	inFormationDir := ctxErr == nil

	// Build URL with machine ID
	params := url.Values{}
	params.Set("ic", telemetry.GetMachineID())

	if inFormationDir {
		// Try to resolve profile
		profile := formation.ResolveProfile(profileFlag)

		if profile != "" {
			// Get server URL from profile
			profileEntry, err := server.GetProfile(profile)
			if err == nil && profileEntry.URL != "" {
				params.Set("server", profileEntry.URL)
				params.Set("formation", ctx.ID)
			}
		}
	}

	// Build final URL - always use /launch endpoint
	targetURL := fmt.Sprintf("%s?%s", consoleLaunchURL, params.Encode())

	// Open in browser
	if err := openBrowser(targetURL); err != nil {
		// If browser open fails, just print the URL
		fmt.Println()
		ui.Success("Open this URL in your browser:")
		fmt.Println()
		fmt.Printf("  %s\n", targetURL)
		fmt.Println()
		return nil
	}

	fmt.Println()
	ui.Success("Opening Console in browser...")
	ui.Dimmed(fmt.Sprintf("  %s", targetURL))
	return nil
}
