package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/telemetry"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

var logsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "Stream formation logs",
	GroupID: "formation",
	Long: `Stream real-time logs from a formation.

Streams Server-Sent Events (SSE) from the formation's log endpoint.
At least one filter flag is required.

Requires admin API key.`,
	Example: `  muxi logs -u alice
  muxi logs -s sess_abc123
  muxi logs --request req_xyz
  muxi logs --level error
  muxi logs --agent weather-bot
  muxi logs --event "chat.*"`,
	RunE: runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	formation.AddFormationFlag(logsCmd)
	formation.AddProfileFlag(logsCmd)
	logsCmd.Flags().StringP("user", "u", "", "Filter by user ID")
	logsCmd.Flags().StringP("session", "s", "", "Filter by session ID")
	logsCmd.Flags().String("request", "", "Filter by request ID")
	logsCmd.Flags().String("agent", "", "Filter by agent ID")
	logsCmd.Flags().String("level", "", "Filter by log level (debug, info, warn, error, critical)")
	logsCmd.Flags().String("event", "", "Filter by event type (supports wildcards, e.g. chat.*)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	// Track telemetry
	state := telemetry.Load()
	state.IncrementLogs()
	state.FlushIfDue()
	defer state.Save()

	userID, _ := cmd.Flags().GetString("user")
	sessionID, _ := cmd.Flags().GetString("session")
	requestID, _ := cmd.Flags().GetString("request")
	agentID, _ := cmd.Flags().GetString("agent")
	level, _ := cmd.Flags().GetString("level")
	eventType, _ := cmd.Flags().GetString("event")

	// Require at least one filter
	if userID == "" && sessionID == "" && requestID == "" && agentID == "" && level == "" && eventType == "" {
		return cmd.Help()
	}

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	// Show streaming header
	fmt.Println()
	fmt.Printf("  Streaming logs %s\n", ui.DimmedText("(Ctrl+C to stop)"))
	fmt.Println()

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Build filters
	filters := formation.LogFilters{
		UserID:    userID,
		SessionID: sessionID,
		RequestID: requestID,
		AgentID:   agentID,
		Level:     level,
		EventType: eventType,
	}

	// Start streaming in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- streamLogs(ctx, client, filters)
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		cancel() // Cancel the context to close the connection
		fmt.Println()
		fmt.Println()
		ui.Dimmed("  Stopped streaming")
		fmt.Println()
		return nil
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			return err
		}
		return nil
	}
}

func streamLogs(ctx context.Context, client *formation.Client, filters formation.LogFilters) error {
	resp, err := client.StreamLogsWithContext(ctx, filters)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("stream failed: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		// Parse SSE format
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			if data == "" || data == "[DONE]" {
				continue
			}

			// Extract timestamp using gjson (fast, no full parse)
			ts := gjson.Get(data, "timestamp").Int()
			t := time.UnixMilli(ts)
			timestamp := t.Format("02-Jan-2006 15:04:05 MST")

			fmt.Printf("  [%s] %s\n", ui.DimmedText(timestamp), data)
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		return err
	}
	return ctx.Err()
}
