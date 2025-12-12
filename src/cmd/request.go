package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var requestCmd = &cobra.Command{
	Use:     "request <request-id>",
	Short:   "Get request status",
	GroupID: "formation",
	Long: `Get the status of an async request by its ID.

Useful for tracking long-running requests that were processed asynchronously.
Request IDs are returned when triggers or chat requests run in async mode.`,
	Example: `  muxi request req_abc123`,
	Args:    RequireArgs(1),
	RunE:    runRequest,
}

func init() {
	rootCmd.AddCommand(requestCmd)

	formation.AddFormationFlag(requestCmd)
	formation.AddProfileFlag(requestCmd)
}

func runRequest(cmd *cobra.Command, args []string) error {
	requestID := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	status, err := client.GetRequestStatus(requestID)
	if err != nil {
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
