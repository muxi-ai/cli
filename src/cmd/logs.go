package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
Use filter flags to narrow down the logs you want to see.

Requires admin API key.

Examples:
  muxi logs
  muxi logs --level error
  muxi logs --agent weather-bot
  muxi logs -u alice`,
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

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	// Show streaming header
	fmt.Println()
	fmt.Printf("  Streaming logs %s\n", ui.DimmedText("(Ctrl+C to stop)"))
	fmt.Println()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start streaming in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- streamLogs(client, userFilter, level, agent, requestID)
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Println()
		fmt.Println()
		ui.Dimmed("  Stopped streaming")
		fmt.Println()
		return nil
	case err := <-errChan:
		if err != nil {
			return err
		}
		return nil
	}
}

func streamLogs(client *formation.Client, userID, level, agent, requestID string) error {
	resp, err := client.StreamLogs(userID, level, agent, requestID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("stream failed: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
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

	return scanner.Err()
}

func displayLogEvent(event formation.LogStreamEvent) {
	// Format timestamp
	timestamp := event.Timestamp.Format("15:04:05")

	// Color-code level
	var levelDisplay string
	switch strings.ToUpper(event.Level) {
	case "ERROR":
		levelDisplay = ui.RedText("ERROR")
	case "WARN", "WARNING":
		levelDisplay = ui.YellowText("WARN ")
	case "INFO":
		levelDisplay = ui.CyanText("INFO ")
	case "DEBUG":
		levelDisplay = ui.DimmedText("DEBUG")
	default:
		levelDisplay = event.Level
	}

	// Build context parts
	var context []string
	if event.User != "" {
		context = append(context, fmt.Sprintf("user=%s", event.User))
	}
	if event.Agent != "" {
		context = append(context, fmt.Sprintf("agent=%s", event.Agent))
	}
	if event.Session != "" {
		context = append(context, fmt.Sprintf("session=%s", event.Session))
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
