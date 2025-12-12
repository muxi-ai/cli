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

var auditCmd = &cobra.Command{
	Use:     "audit",
	Short:   "View audit log",
	GroupID: "formation",
	Long: `View the formation's audit log.

The audit log records administrative actions like agent creation,
configuration changes, and user management operations.

Requires admin API key.`,
	RunE: runAudit,
}

func init() {
	// TODO: Uncomment when audit endpoints are implemented in runtime
	// rootCmd.AddCommand(auditCmd)
	//
	// formation.AddFormationFlag(auditCmd)
	// formation.AddProfileFlag(auditCmd)
	// auditCmd.Flags().IntP("lines", "n", 50, "Number of entries to show")
	// auditCmd.Flags().Bool("clear", false, "Clear the audit log")
}

func runAudit(cmd *cobra.Command, args []string) error {
	clear, _ := cmd.Flags().GetBool("clear")
	lines, _ := cmd.Flags().GetInt("lines")

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	// Handle clear
	if clear {
		return runAuditClear(client)
	}

	// Get audit log
	resp, err := client.GetAuditLog()
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No audit entries")
		fmt.Println()
		return nil
	}

	// Limit entries
	entries := resp.Entries
	if len(entries) > lines {
		entries = entries[len(entries)-lines:]
	}

	// Header
	fmt.Println()
	fmt.Printf("  %-20s %-22s %-18s %s\n",
		"TIMESTAMP", "ACTION", "RESOURCE", "USER")
	fmt.Printf("  %-20s %-22s %-18s %s\n",
		"────────────────────", "────────────────────", "────────────────", "──────────")

	for _, entry := range entries {
		// Format timestamp
		timestamp := "-"
		if entry.Timestamp != nil {
			timestamp = entry.Timestamp.Format("2006-01-02 15:04:05")
		}

		// Resource is type/id
		resource := entry.ResourceID
		if entry.ResourceType != "" && entry.ResourceID != "" {
			resource = entry.ResourceType + "/" + entry.ResourceID
		} else if entry.ResourceType != "" {
			resource = entry.ResourceType
		}

		// Truncate fields for display
		action := truncateString(entry.Action, 20)
		resource = truncateString(resource, 16)
		user := truncateString(entry.User, 10)
		if user == "" {
			user = "-"
		}

		fmt.Printf("  %-20s %-22s %-18s %s\n",
			timestamp, action, resource, user)
	}

	// Show if truncated
	if resp.Count > lines {
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  Showing last %d of %d entries", lines, resp.Count))
	}

	fmt.Println()
	return nil
}

func runAuditClear(client *formation.Client) error {
	// Confirm before clearing
	fmt.Println()
	fmt.Printf("  Clear the audit log?\n")
	fmt.Println()
	fmt.Printf("  %s\n", ui.RedText("This action cannot be undone!"))
	fmt.Print("  Type 'clear' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	confirm = strings.TrimSpace(confirm)

	if confirm != "clear" {
		fmt.Println()
		ui.Dimmed("  Cancelled")
		fmt.Println()
		return nil
	}

	spinner := ui.NewSpinner("Clearing audit log...")
	spinner.Start()

	err = client.ClearAuditLog()
	if err != nil {
		spinner.StopWithError("Clear failed")
		return err
	}

	spinner.StopWithSuccess("Audit log cleared")
	fmt.Println()

	return nil
}

// truncateString truncates a string to maxLen with ".." suffix
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}
