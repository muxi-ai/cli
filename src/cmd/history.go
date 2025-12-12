package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history",
	Short:   "View session message history",
	GroupID: "formation",
	Long: `View message history for a session.

Requires a session ID (-s) and user ID (-u).`,
	Example: `  muxi history -s sess_abc123 -u test-user
  muxi history -s sess_abc123 -u test-user --lines 50
  muxi history -s sess_abc123 -u test-user --json`,
	RunE: runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)

	formation.AddCommonFlags(historyCmd)
	historyCmd.Flags().StringP("session", "s", "", "Session ID (required)")
	historyCmd.Flags().IntP("lines", "n", 0, "Limit number of messages (0 = all)")
	historyCmd.Flags().Bool("json", false, "Output as JSON")

	historyCmd.MarkFlagRequired("session")
}

func runHistory(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	sessionID, _ := cmd.Flags().GetString("session")
	lines, _ := cmd.Flags().GetInt("lines")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	spinner := ui.NewSpinner("Fetching messages...")
	spinner.Start()

	messages, err := client.GetSessionMessages(sessionID, userID)
	if err != nil {
		spinner.StopWithError("Failed to fetch messages")
		return err
	}

	spinner.Stop()

	// JSON output (no badge)
	if jsonOutput {
		data, err := json.MarshalIndent(messages, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	formation.PrintBadgeFromFlags(cmd)

	if messages.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No messages in session")
		fmt.Println()
		return nil
	}

	// Apply line limit
	displayMessages := messages.Messages
	if lines > 0 && lines < len(displayMessages) {
		displayMessages = displayMessages[len(displayMessages)-lines:]
	}

	// Print header
	fmt.Println()
	fmt.Printf("  Session: %s (%d messages)\n", ui.BoldText(sessionID), messages.Count)
	fmt.Println()

	// Print messages
	for _, m := range displayMessages {
		timestamp := m.Timestamp.Format("15:04:05")
		role := m.Role

		// Color role
		var roleDisplay string
		switch role {
		case "user":
			roleDisplay = ui.CyanText("user")
		case "assistant":
			if m.Agent != "" {
				roleDisplay = ui.GreenText(m.Agent)
			} else {
				roleDisplay = ui.GreenText("assistant")
			}
		case "system":
			roleDisplay = ui.DimmedText("system")
		default:
			roleDisplay = role
		}

		fmt.Printf("  [%s] %s: %s\n", ui.DimmedText(timestamp), roleDisplay, m.Content)
	}

	fmt.Println()

	return nil
}
