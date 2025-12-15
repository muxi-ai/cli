package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var requestsCmd = &cobra.Command{
	Use:     "requests",
	Short:   "Manage chat requests",
	GroupID: "formation",
	Long: `List and manage chat requests.

Track status of async requests and cancel in-progress ones.`,
}

var requestsListCmd = &cobra.Command{
	Use:     "list [user-id]",
	Aliases: []string{"ls"},
	Short:   "List chat requests",
	Long: `List chat requests and their status.

Status can be: processing, completed, failed, or cancelled.`,
	Example: `  muxi requests list
  muxi requests list alice`,
	RunE: runRequestsList,
}

var requestsShowCmd = &cobra.Command{
	Use:   "show <request-id>",
	Short: "Get request status",
	Long:  `Get detailed status of a specific request.`,
	Example: `  muxi requests show req_abc123`,
	Args: RequireArgs(1),
	RunE: runRequestsShow,
}

var requestsCancelCmd = &cobra.Command{
	Use:   "cancel <request-id> [user-id]",
	Short: "Cancel a request",
	Long: `Cancel an in-progress request.

Only requests with status "processing" can be cancelled.`,
	Example: `  muxi requests cancel req_123
  muxi requests cancel req_123 alice`,
	Args: RequireArgs(1),
	RunE: runRequestsCancel,
}

func init() {
	rootCmd.AddCommand(requestsCmd)
	requestsCmd.AddCommand(requestsListCmd)
	requestsCmd.AddCommand(requestsShowCmd)
	requestsCmd.AddCommand(requestsCancelCmd)

	// Formation and profile flags
	formation.AddFormationFlag(requestsListCmd)
	formation.AddProfileFlag(requestsListCmd)
	formation.AddFormationFlag(requestsShowCmd)
	formation.AddProfileFlag(requestsShowCmd)
	formation.AddUserFlag(requestsShowCmd)
	formation.AddFormationFlag(requestsCancelCmd)
	formation.AddProfileFlag(requestsCancelCmd)
}

func runRequestsList(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	// User from arg or default
	var userID string
	if len(args) > 0 {
		userID = args[0]
	} else {
		userID, err = formation.MustResolveUserID("")
		if err != nil {
			return fmt.Errorf("user ID required: provide as argument or set default with: muxi set default user")
		}
	}

	resp, err := client.GetRequests(userID)
	if err != nil {
		// Handle not found gracefully
		if strings.Contains(err.Error(), "NOT_FOUND") {
			formation.PrintBadgeFromFlags(cmd)
			fmt.Println()
			ui.Dimmed("  No requests found")
			fmt.Println()
			return nil
		}
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No requests found")
		fmt.Println()
		return nil
	}

	// Header
	fmt.Println()
	fmt.Printf("  %-16s %-12s %-10s %s\n",
		"REQUEST ID", "STATUS", "PROGRESS", "CREATED")
	fmt.Printf("  %-16s %-12s %-10s %s\n",
		"────────────────", "────────────", "──────────", "────────────")

	for _, req := range resp.Requests {
		// Status with color
		var statusDisplay string
		switch req.Status {
		case "completed":
			statusDisplay = ui.GreenText("completed")
		case "processing":
			statusDisplay = ui.CyanText("processing")
		case "failed":
			statusDisplay = ui.RedText("failed")
		case "cancelled":
			statusDisplay = ui.DimmedText("cancelled")
		default:
			statusDisplay = req.Status
		}

		// Progress
		progressDisplay := "-"
		if req.Progress > 0 {
			progressDisplay = fmt.Sprintf("%d%%", req.Progress)
		}

		// Created time
		createdDisplay := "-"
		if req.CreatedAt != nil {
			createdDisplay = formatRelativeTime(req.CreatedAt.Time)
		}

		// Truncate request ID for display
		reqID := req.RequestID
		if len(reqID) > 14 {
			reqID = reqID[:12] + ".."
		}

		fmt.Printf("  %-16s %-12s %-10s %s\n",
			reqID, statusDisplay, progressDisplay, createdDisplay)
	}

	fmt.Println()
	return nil
}

func runRequestsShow(cmd *cobra.Command, args []string) error {
	requestID := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	userFlag, _ := cmd.Flags().GetString("user")
	userID, err := formation.MustResolveUserID(userFlag)
	if err != nil {
		return err
	}

	status, err := client.GetRequestStatus(requestID, userID)
	if err != nil {
		// Handle not found gracefully
		if strings.Contains(err.Error(), "NOT_FOUND") {
			formation.PrintBadgeFromFlags(cmd)
			fmt.Println()
			fmt.Printf("  Request '%s' not found\n", requestID)
			fmt.Println()
			ui.Dimmed("  The request may have expired or been deleted.")
			fmt.Println()
			return nil
		}
		return fmt.Errorf("failed to get request status: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	fmt.Printf("  %s\n", ui.BoldText(status.RequestID))

	// Status with icon
	var statusIcon, statusText string
	switch status.Status {
	case "completed":
		statusIcon = ui.GreenText("●")
		statusText = "completed"
	case "processing":
		statusIcon = ui.CyanText("●")
		statusText = "processing"
	case "failed":
		statusIcon = ui.RedText("●")
		statusText = "failed"
	case "cancelled":
		statusIcon = ui.DimmedText("●")
		statusText = "cancelled"
	default:
		statusIcon = ui.DimmedText("○")
		statusText = status.Status
	}
	fmt.Printf("   Status:    %s %s\n", statusIcon, statusText)

	if status.Progress != "" {
		fmt.Printf("   Progress:  %s\n", status.Progress)
	}
	if status.Error != "" {
		fmt.Printf("   Error:     %s\n", ui.RedText(status.Error))
	}
	if status.CompletedAt != nil {
		fmt.Printf("   Completed: %s\n", status.CompletedAt.Format("Jan 2, 2006 3:04pm"))
	}
	fmt.Println()

	return nil
}

func runRequestsCancel(cmd *cobra.Command, args []string) error {
	requestID := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	// User from second arg or default
	var userID string
	if len(args) > 1 {
		userID = args[1]
	} else {
		userID, err = formation.MustResolveUserID("")
		if err != nil {
			return fmt.Errorf("user ID required: provide as second argument or set default with: muxi set default user")
		}
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Cancelling request '%s'...", requestID))
	spinner.Start()

	err = client.CancelRequest(requestID, userID)
	if err != nil {
		spinner.StopWithError("Cancel failed")
		return err
	}

	spinner.StopWithSuccess("Request cancelled")
	fmt.Println()
	ui.Success(fmt.Sprintf("Cancelled request %s", requestID))
	fmt.Println()

	return nil
}

// formatRelativeTime formats a time as relative (e.g., "2 minutes ago")
func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
