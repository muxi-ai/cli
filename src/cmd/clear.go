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
	Short:   "Clear session or memory buffer",
	GroupID: "formation",
	Long: `Delete a session and its message history, or clear memory buffer.

Use -s to clear a specific session.
Use --buffer to clear the user's memory buffer.
Use --all-buffers to clear all buffers (admin only).

Requires user ID (-u) for session and buffer operations.
Use -f to skip confirmation prompt.`,
	Example: `  muxi clear -s sess_abc123 -u test-user      # Clear session
  muxi clear --buffer -u test-user             # Clear user's buffer  
  muxi clear --buffer -s sess_abc123 -u alice  # Clear session buffer
  muxi clear --all-buffers                     # Clear all buffers (admin)`,
	RunE: runClear,
}

func init() {
	rootCmd.AddCommand(clearCmd)

	formation.AddCommonFlags(clearCmd)
	clearCmd.Flags().StringP("session", "s", "", "Session ID to clear")
	clearCmd.Flags().Bool("buffer", false, "Clear memory buffer instead of session")
	clearCmd.Flags().Bool("all-buffers", false, "Clear all buffers (admin only)")
	clearCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runClear(cmd *cobra.Command, args []string) error {
	sessionID, _ := cmd.Flags().GetString("session")
	buffer, _ := cmd.Flags().GetBool("buffer")
	allBuffers, _ := cmd.Flags().GetBool("all-buffers")
	force, _ := cmd.Flags().GetBool("force")

	// Handle --all-buffers (admin only, no user required)
	if allBuffers {
		return runClearAllBuffers(cmd, force)
	}

	// Handle --buffer (clear user's buffer)
	if buffer {
		return runClearBuffer(cmd, sessionID, force)
	}

	// Default: clear session (requires -s)
	if sessionID == "" {
		return fmt.Errorf("session ID required (-s), or use --buffer to clear memory buffer")
	}

	return runClearSession(cmd, sessionID, force)
}

func runClearSession(cmd *cobra.Command, sessionID string, force bool) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

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

func runClearBuffer(cmd *cobra.Command, sessionID string, force bool) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	// If session ID provided, clear specific session buffer
	if sessionID != "" {
		if !force {
			fmt.Printf("\n  Clear buffer for session '%s'? [y/N]: ", sessionID)
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

		spinner := ui.NewSpinner("Clearing session buffer...")
		spinner.Start()

		resp, err := client.ClearSessionBuffer(userID, sessionID)
		if err != nil {
			spinner.StopWithError("Failed to clear buffer")
			return err
		}

		spinner.StopWithSuccess("Buffer cleared")
		fmt.Println()
		ui.Success(fmt.Sprintf("Cleared %d messages from session %s", resp.MessagesCleared, sessionID))
		fmt.Println()
		return nil
	}

	// Clear entire user buffer
	if !force {
		fmt.Printf("\n  Clear all buffer messages for user '%s'? [y/N]: ", userID)
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

	spinner := ui.NewSpinner("Clearing buffer...")
	spinner.Start()

	resp, err := client.ClearMemoryBuffer(userID)
	if err != nil {
		spinner.StopWithError("Failed to clear buffer")
		return err
	}

	spinner.StopWithSuccess("Buffer cleared")
	fmt.Println()
	ui.Success(fmt.Sprintf("Cleared %d messages from %d sessions", resp.MessagesCleared, resp.SessionsCleared))
	fmt.Println()

	return nil
}

func runClearAllBuffers(cmd *cobra.Command, force bool) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	if !force {
		fmt.Println()
		fmt.Printf("  %s\n", ui.RedText("Clear ALL memory buffers for ALL users?"))
		fmt.Print("  Type 'clear all' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "clear all" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	spinner := ui.NewSpinner("Clearing all buffers...")
	spinner.Start()

	resp, err := client.ClearAllBuffers()
	if err != nil {
		spinner.StopWithError("Failed to clear buffers")
		return err
	}

	spinner.StopWithSuccess("All buffers cleared")
	fmt.Println()
	ui.Success(fmt.Sprintf("Cleared %d messages", resp.MessagesCleared))
	fmt.Println()

	return nil
}
