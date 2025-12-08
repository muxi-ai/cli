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
	Long: `List sessions for a user in the current formation.

Requires a user ID (via -u flag, .muxi file, or global default).`,
	Example: `  muxi sessions -u test-user
  muxi sessions --active`,
	RunE: runSessions,
}

func init() {
	rootCmd.AddCommand(sessionsCmd)

	formation.AddCommonFlags(sessionsCmd)
	sessionsCmd.Flags().Bool("active", false, "Show only active sessions")
}

func runSessions(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	activeOnly, _ := cmd.Flags().GetBool("active")

	spinner := ui.NewSpinner("Fetching sessions...")
	spinner.Start()

	sessions, err := client.GetSessions(userID)
	if err != nil {
		spinner.StopWithError("Failed to fetch sessions")
		return err
	}

	spinner.Stop()

	if sessions.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No sessions found")
		fmt.Println()
		return nil
	}

	// Filter if --active
	displaySessions := sessions.Sessions
	if activeOnly {
		var filtered []formation.Session
		for _, s := range sessions.Sessions {
			if s.Status == "active" {
				filtered = append(filtered, s)
			}
		}
		displaySessions = filtered
		if len(filtered) == 0 {
			fmt.Println()
			ui.Dimmed("  No active sessions found")
			fmt.Println()
			return nil
		}
	}

	// Print header
	fmt.Println()
	fmt.Printf("  %-20s  %-10s  %-18s  %s\n",
		ui.BoldText("SESSION ID"),
		ui.BoldText("MESSAGES"),
		ui.BoldText("LAST ACTIVITY"),
		ui.BoldText("STATUS"))

	// Print sessions
	for _, s := range displaySessions {
		statusIcon := "○"
		if s.Status == "active" {
			statusIcon = ui.GreenText("●")
		}

		fmt.Printf("  %-20s  %-10d  %-18s  %s %s\n",
			s.ID,
			s.MessageCount,
			formatTimeAgo(s.LastActivity),
			statusIcon,
			s.Status)
	}

	fmt.Println()
	ui.Dimmed(fmt.Sprintf("  %d session(s) for user %s", len(displaySessions), userID))
	fmt.Println()

	return nil
}
