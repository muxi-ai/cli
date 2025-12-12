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
	fmt.Printf("  Request ID:      %s\n", status.RequestID)

	// Status with color
	var statusDisplay string
	switch status.Status {
	case "completed":
		statusDisplay = ui.GreenText("completed")
	case "processing":
		statusDisplay = ui.CyanText("processing")
	case "pending":
		statusDisplay = "pending"
	case "failed":
		statusDisplay = ui.RedText("failed")
	default:
		statusDisplay = status.Status
	}
	fmt.Printf("  Status:          %s\n", statusDisplay)

	if status.FormationID != "" {
		fmt.Printf("  Formation:       %s\n", status.FormationID)
	}
	if status.UserID != "" {
		fmt.Printf("  User:            %s\n", status.UserID)
	}
	if status.ProcessingMode != "" {
		fmt.Printf("  Mode:            %s\n", status.ProcessingMode)
	}
	if status.ProcessingTime > 0 {
		fmt.Printf("  Processing Time: %.2fs\n", status.ProcessingTime)
	}
	fmt.Println()

	return nil
}
