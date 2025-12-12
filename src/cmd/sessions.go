package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var sessionsCmd = &cobra.Command{
	Use:     "sessions",
	Short:   "List user sessions",
	GroupID: "formation",
	Long: `List active sessions for a user in the current formation.

Sessions are ephemeral and stored in memory. Only active sessions with 
recent activity are displayed. Inactive sessions are automatically 
cleaned up by the formation.

Requires a user ID (via -u flag, .muxi file, or global default).`,
	Example: `  muxi sessions -u alice`,
	RunE:    runSessions,
}

var sessionsShowCmd = &cobra.Command{
	Use:   "show <session-id>",
	Short: "Show session details",
	Long:  `Display detailed information about a specific session.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionsShow,
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
	sessionsCmd.AddCommand(sessionsShowCmd)

	formation.AddCommonFlags(sessionsCmd)
	formation.AddCommonFlags(sessionsShowCmd)
}

func runSessions(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	spinner := ui.NewSpinner("Fetching sessions...")
	spinner.Start()

	sessions, err := client.GetSessions(userID)
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
