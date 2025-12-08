package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:     "clear",
	Short:   "Clear a session",
	GroupID: "formation",
	Long: `Delete a session and its message history.

Requires a session ID (-s) and user ID (-u).
Use -f to skip confirmation prompt.`,
	Example: `  muxi clear -s sess_abc123 -u test-user
  muxi clear -s sess_abc123 -u test-user -f`,
	RunE: runClear,
}

func init() {
	rootCmd.AddCommand(clearCmd)

	formation.AddCommonFlags(clearCmd)
	clearCmd.Flags().StringP("session", "s", "", "Session ID (required)")
	clearCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	clearCmd.MarkFlagRequired("session")
}

func runClear(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	sessionID, _ := cmd.Flags().GetString("session")
	force, _ := cmd.Flags().GetBool("force")

	// Confirm deletion
	if !force {
		fmt.Printf("\n  Clear session '%s'? This will delete all messages. [y/N]: ", sessionID)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	spinner := ui.NewSpinner("Clearing session...")
	spinner.Start()

	err = client.DeleteSession(sessionID, userID)
	if err != nil {
		spinner.StopWithError("Failed to clear session")
		return err
	}

	spinner.StopWithSuccess("Cleared session")

	fmt.Println()
	ui.Success(fmt.Sprintf("Cleared session %s", sessionID))
	fmt.Println()

	return nil
}
