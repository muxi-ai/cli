package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
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

var sessionsRestoreCmd = &cobra.Command{
	Use:   "restore <session-id>",
	Short: "Restore session from file",
	Long: `Restore a session from a JSON file.

The file should be in the same format as 'muxi sessions messages --json' output.`,
	Example: `  muxi sessions restore sess_abc123 --file backup.json
  muxi sessions restore sess_abc123 -f session.json`,
	Args: cobra.ExactArgs(1),
	RunE: runSessionsRestore,
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsShowCmd)
	sessionsCmd.AddCommand(sessionsMessagesCmd)
	sessionsCmd.AddCommand(sessionsRestoreCmd)

	sessionsListCmd.Flags().Int("limit", 10, "Maximum number of sessions to return")
	sessionsMessagesCmd.Flags().IntP("lines", "n", 0, "Limit number of messages (0 = all)")
	sessionsMessagesCmd.Flags().Bool("json", false, "Output as JSON")
	sessionsRestoreCmd.Flags().StringP("file", "f", "", "JSON file to restore from (required)")
	sessionsRestoreCmd.MarkFlagRequired("file")

	formation.AddCommonFlags(sessionsListCmd)
	formation.AddCommonFlags(sessionsShowCmd)
	formation.AddCommonFlags(sessionsMessagesCmd)
	formation.AddCommonFlags(sessionsRestoreCmd)
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

	// Styles
	goldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e48d20"))
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#c98b45"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

	// Markdown renderer
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(76),
	)

	// Print header
	fmt.Println()
	fmt.Printf("  Session: %s (%d messages)\n", ui.BoldText(sessionID), messages.Count)
	fmt.Println()

	// Print messages
	for _, m := range displayMessages {
		timestamp := dimStyle.Render(m.Timestamp.Format("15:04"))

		switch m.Role {
		case "user":
			// User message: >  content
			fmt.Printf("\n %s  %s  %s\n", timestamp, goldStyle.Render(">"), userStyle.Render(m.Content))

		case "assistant":
			// Assistant message: 𝐌 with markdown rendering
			rendered := m.Content
			if renderer != nil {
				if r, err := renderer.Render(m.Content); err == nil {
					rendered = strings.TrimSpace(r)
					// Indent each line
					lines := strings.Split(rendered, "\n")
					for i, line := range lines {
						if i > 0 {
							lines[i] = "      " + line
						}
					}
					rendered = strings.Join(lines, "\n")
				}
			}
			agent := "𝐌"
			if m.Agent != "" {
				agent = m.Agent
			}
			fmt.Printf("\n %s  %s %s\n", timestamp, goldStyle.Render(agent), rendered)

		case "system":
			// System message: dimmed
			fmt.Printf("\n %s  %s\n", timestamp, dimStyle.Render(m.Content))

		default:
			fmt.Printf("\n %s  %s: %s\n", timestamp, m.Role, m.Content)
		}
	}

	fmt.Println()

	return nil
}

func runSessionsRestore(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	filePath, _ := cmd.Flags().GetString("file")

	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON - accept both full response format and just messages array
	var messages []formation.Message

	// Try full format first (from --json output)
	var fullFormat formation.SessionMessagesResponse
	if err := json.Unmarshal(data, &fullFormat); err == nil && len(fullFormat.Messages) > 0 {
		messages = fullFormat.Messages
	} else {
		// Try messages array directly
		if err := json.Unmarshal(data, &messages); err != nil {
			return fmt.Errorf("invalid JSON format: expected messages array or session messages response")
		}
	}

	// Validate messages
	if len(messages) == 0 {
		return fmt.Errorf("no messages found in file")
	}

	for i, m := range messages {
		if m.Role == "" {
			return fmt.Errorf("message %d: missing 'role' field", i+1)
		}
		if m.Content == "" {
			return fmt.Errorf("message %d: missing 'content' field", i+1)
		}
		if m.Role != "user" && m.Role != "assistant" && m.Role != "system" {
			return fmt.Errorf("message %d: invalid role '%s' (must be user, assistant, or system)", i+1, m.Role)
		}
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Restoring %d messages...", len(messages)))
	spinner.Start()

	err = client.RestoreSession(sessionID, userID, messages)
	if err != nil {
		spinner.StopWithError("Failed to restore session")
		return err
	}

	spinner.StopWithSuccess("Session restored")
	fmt.Println()
	ui.Success(fmt.Sprintf("Restored %d messages to session %s", len(messages), sessionID))
	fmt.Println()

	return nil
}
