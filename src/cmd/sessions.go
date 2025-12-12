package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var sessionsCmd = &cobra.Command{
	Use:     "sessions",
	Short:   "Manage user sessions",
	GroupID: "formation",
	Long: `Manage user sessions in the current formation.

Sessions are ephemeral and stored in memory. Inactive sessions are 
automatically cleaned up by the formation.`,
}

var sessionsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List user sessions",
	Long: `List active sessions for a user.

Requires a user ID (via -u flag, .muxi file, or global default).`,
	Example: `  muxi sessions list -u alice
  muxi sessions list --limit 20`,
	RunE: runSessionsList,
}

var sessionsShowCmd = &cobra.Command{
	Use:   "show <session-id>",
	Short: "Show session details",
	Long:  `Display detailed information about a specific session.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionsShow,
}

var sessionsMessagesCmd = &cobra.Command{
	Use:   "messages <session-id>",
	Short: "View session message history",
	Long:  `View message history for a session.`,
	Example: `  muxi sessions messages sess_abc123
  muxi sessions messages sess_abc123 --lines 50
  muxi sessions messages sess_abc123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runSessionsMessages,
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsShowCmd)
	sessionsCmd.AddCommand(sessionsMessagesCmd)

	sessionsListCmd.Flags().Int("limit", 10, "Maximum number of sessions to return")
	sessionsMessagesCmd.Flags().IntP("lines", "n", 0, "Limit number of messages (0 = all)")
	sessionsMessagesCmd.Flags().Bool("json", false, "Output as JSON")

	formation.AddCommonFlags(sessionsListCmd)
	formation.AddCommonFlags(sessionsShowCmd)
	formation.AddCommonFlags(sessionsMessagesCmd)
}

func runSessionsList(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")

	spinner := ui.NewSpinner("Fetching sessions...")
	spinner.Start()

	sessions, err := client.GetSessions(userID, limit)
	if err != nil {
		spinner.StopWithError("Failed to fetch sessions")
		return err
	}

	spinner.Stop()

	formation.PrintBadgeFromFlags(cmd)

	if sessions.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No active sessions")
		fmt.Println()
		return nil
	}

	// Print header
	fmt.Println()
	fmt.Printf("  %-20s  %-10s  %s\n",
		ui.BoldText("SESSION ID"),
		ui.BoldText("MESSAGES"),
		ui.BoldText("LAST ACTIVITY"))

	// Print sessions
	for _, s := range sessions.Sessions {
		fmt.Printf("  %-20s  %-10d  %s\n",
			s.ID,
			s.MessageCount,
			formatTimeAgo(s.LastActivity))
	}

	fmt.Println()
	ui.Dimmed(fmt.Sprintf("  %d session(s) for user %s", sessions.Count, userID))
	fmt.Println()

	return nil
}

func runSessionsShow(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	session, err := client.GetSession(sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	fmt.Printf("  Session ID:    %s\n", session.SessionID)
	fmt.Printf("  User ID:       %s\n", session.UserID)
	fmt.Printf("  Messages:      %d\n", session.MessageCount)
	if !session.CreatedAt.IsZero() {
		fmt.Printf("  Created:       %s\n", session.CreatedAt.Format("Jan 2, 2006 3:04pm"))
	}
	if !session.LastActivity.IsZero() {
		fmt.Printf("  Last Activity: %s\n", formatTimeAgo(session.LastActivity))
	}
	fmt.Println()

	return nil
}

func runSessionsMessages(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

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
