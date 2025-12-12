package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
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
  muxi logs --level error
  muxi logs --agent weather-bot
  muxi logs -u alice --level error`,
	RunE: runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	formation.AddFormationFlag(logsCmd)
	formation.AddProfileFlag(logsCmd)
	logsCmd.Flags().StringP("user", "u", "", "Filter by user ID")
	logsCmd.Flags().String("level", "", "Filter by log level (debug, info, warn, error)")
	logsCmd.Flags().String("agent", "", "Filter by agent ID")
	logsCmd.Flags().String("request", "", "Filter by request ID")
}

func runLogs(cmd *cobra.Command, args []string) error {
	userFilter, _ := cmd.Flags().GetString("user")
	level, _ := cmd.Flags().GetString("level")
	agent, _ := cmd.Flags().GetString("agent")
	requestID, _ := cmd.Flags().GetString("request")

	// Require at least one filter
	if userFilter == "" && level == "" && agent == "" && requestID == "" {
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

	// Start streaming in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- streamLogs(ctx, client, userFilter, level, agent, requestID)
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

func streamLogs(ctx context.Context, client *formation.Client, userID, level, agent, requestID string) error {
	resp, err := client.StreamLogsWithContext(ctx, userID, level, agent, requestID)
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

			var logEvent formation.LogStreamEvent
			if err := json.Unmarshal([]byte(data), &logEvent); err != nil {
				// Skip malformed events
				continue
			}

			displayLogEvent(logEvent)
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		return err
	}
	return ctx.Err()
}

func displayLogEvent(event formation.LogStreamEvent) {
	// Format timestamp (Unix ms to time)
	t := time.UnixMilli(event.Timestamp)
	timestamp := t.Format("15:04:05")

	// Color-code level
	var levelDisplay string
	switch strings.ToUpper(event.Level) {
	case "ERROR", "CRITICAL":
		levelDisplay = ui.RedText("ERROR")
	case "WARN", "WARNING":
		levelDisplay = ui.YellowText("WARN ")
	case "INFO":
		levelDisplay = ui.CyanText("INFO ")
	case "DEBUG":
		levelDisplay = ui.DimmedText("DEBUG")
	default:
		levelDisplay = fmt.Sprintf("%-5s", event.Level)
	}

	// Build context parts
	var context []string
	if event.UserID != "" {
		context = append(context, fmt.Sprintf("user=%s", event.UserID))
	}
	if event.AgentID != "" {
		context = append(context, fmt.Sprintf("agent=%s", event.AgentID))
	}
	if event.SessionID != "" {
		context = append(context, fmt.Sprintf("session=%s", event.SessionID))
	}
	if event.RequestID != "" {
		// Truncate request ID
		reqID := event.RequestID
		if len(reqID) > 12 {
			reqID = reqID[:12]
		}
		context = append(context, fmt.Sprintf("req=%s", reqID))
	}

	// Format context string
	contextStr := ""
	if len(context) > 0 {
		contextStr = " " + ui.DimmedText(strings.Join(context, " "))
	}

	// Print log line
	fmt.Printf("  [%s] %s  %s%s\n", timestamp, levelDisplay, event.Message, contextStr)
}
