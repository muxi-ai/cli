package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"

	"github.com/spf13/cobra"
)

var formationCmd = &cobra.Command{
	Use:   "formation",
	Short: "Manage deployed formations",
	Long:  `List, inspect, and manage formations deployed to a MUXI server.`,
}

var formationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List deployed formations",
	RunE:  runFormationList,
}

var formationGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get formation details",
	Args:  cobra.ExactArgs(1),
	RunE:  runFormationGet,
}

var formationDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a formation",
	Long: `Delete a formation from the server.

This will:
- Stop the formation process/container
- Release the allocated port
- Remove from registry
- Clean up formation directory`,
	Args: cobra.ExactArgs(1),
	RunE: runFormationDelete,
}

var formationStopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "Stop a running formation",
	Args:  cobra.ExactArgs(1),
	RunE:  runFormationStop,
}

var formationRestartCmd = &cobra.Command{
	Use:   "restart <id>",
	Short: "Restart a formation",
	Args:  cobra.ExactArgs(1),
	RunE:  runFormationRestart,
}

var formationRollbackCmd = &cobra.Command{
	Use:   "rollback <id>",
	Short: "Rollback to previous version",
	Long:  `Rollback a formation to its previous deployed version.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runFormationRollback,
}

var formationLogsCmd = &cobra.Command{
	Use:   "logs <id>",
	Short: "View formation logs",
	Args:  cobra.ExactArgs(1),
	RunE:  runFormationLogs,
}

func init() {
	rootCmd.AddCommand(formationCmd)

	formationCmd.AddCommand(formationListCmd)
	formationCmd.AddCommand(formationGetCmd)
	formationCmd.AddCommand(formationDeleteCmd)
	formationCmd.AddCommand(formationStopCmd)
	formationCmd.AddCommand(formationRestartCmd)
	formationCmd.AddCommand(formationRollbackCmd)
	formationCmd.AddCommand(formationLogsCmd)

	// Flags
	formationListCmd.Flags().String("profile", "", "Server profile to use")
	formationGetCmd.Flags().String("profile", "", "Server profile to use")
	formationGetCmd.Flags().BoolP("verbose", "v", false, "Show internal details (port, pid)")
	formationDeleteCmd.Flags().String("profile", "", "Server profile to use")
	formationDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	formationDeleteCmd.Flags().Bool("atomic", false, "Skip confirmation prompt (alias for --force)")
	formationStopCmd.Flags().String("profile", "", "Server profile to use")
	formationRestartCmd.Flags().String("profile", "", "Server profile to use")
	formationRollbackCmd.Flags().String("profile", "", "Server profile to use")
	formationLogsCmd.Flags().String("profile", "", "Server profile to use")
	formationLogsCmd.Flags().IntP("lines", "n", 100, "Number of lines to show")
	formationLogsCmd.Flags().String("stream", "", "Filter by stream (stdout, stderr)")
	formationLogsCmd.Flags().BoolP("follow", "f", false, "Stream new logs (like tail -f)")
}

// runFormationList handles muxi formation list
func runFormationList(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	resp, err := client.ListFormations()
	if err != nil {
		return err
	}

	if len(resp.Formations) == 0 {
		fmt.Println()
		ui.Dimmed("  No formations deployed")
		fmt.Println()
		fmt.Printf("  Deploy a formation: %s\n", ui.Command("muxi deploy"))
		fmt.Println()
		return nil
	}

	// Header
	fmt.Println()
	fmt.Printf("  %-20s %-10s %-10s %s\n",
		"ID", "VERSION", "STATUS", "UPTIME")
	fmt.Printf("  %-20s %-10s %-10s %s\n",
		"──────────────────", "────────", "────────", "──────")

	for _, f := range resp.Formations {
		// Status icon
		statusIcon := ui.GreenText("●")
		statusText := "running"
		if f.Status != "running" {
			statusIcon = ui.DimmedText("○")
			statusText = f.Status
		}

		// Uptime display
		uptimeStr := "-"
		if f.Uptime > 0 {
			uptimeStr = formatDurationShort(f.Uptime)
		}

		// Version display
		version := "-"
		if f.Version != nil && f.Version.Semantic != "" {
			version = f.Version.Semantic
		}

		fmt.Printf("  %-20s %-10s %s %-8s %s\n",
			f.ID, version, statusIcon, statusText, uptimeStr)
	}

	fmt.Println()
	return nil
}

// runFormationGet handles muxi formation get <id>
func runFormationGet(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	verbose, _ := cmd.Flags().GetBool("verbose")
	formationID := args[0]

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	f, err := client.GetFormation(formationID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not found",
				fmt.Sprintf("Formation '%s' does not exist on this server.", formationID),
				ui.Command("muxi formation list"),
			)
			os.Exit(1)
		}
		return err
	}

	// Get server name for display
	serverName := profile
	if serverName == "" {
		serverName = server.GetDefaultServer()
	}

	fmt.Println()
	fmt.Printf("  Formation: %s\n", f.ID)
	if f.Name != "" && f.Name != f.ID {
		fmt.Printf("  Name:      %s\n", f.Name)
	}
	if f.Version != nil && f.Version.Semantic != "" {
		fmt.Printf("  Version:   %s\n", f.Version.Semantic)
	}
	fmt.Println()

	// Status
	var statusDisplay string
	switch f.Status {
	case "running":
		statusDisplay = ui.GreenText("● running")
	case "starting":
		statusDisplay = ui.CyanText("● starting")
	case "stopped":
		statusDisplay = ui.RedText("○ stopped")
	default:
		statusDisplay = ui.RedText("○ " + f.Status)
	}
	fmt.Printf("  Status:     %s\n", statusDisplay)

	// Uptime
	if f.Uptime > 0 {
		fmt.Printf("  Uptime:     %s\n", formatDurationShort(f.Uptime))
	}

	// Health (show — when status is not "running")
	var healthDisplay string
	if f.Status != "running" {
		healthDisplay = ui.DimmedText("—")
	} else if f.Healthy {
		healthDisplay = ui.GreenText("✓ healthy")
	} else {
		healthDisplay = ui.RedText("✗ unhealthy")
	}
	fmt.Printf("  Health:     %s\n", healthDisplay)

	// Restarts
	fmt.Printf("  Restarts:   %d\n", f.RestartCount)

	fmt.Println()

	// Timestamps
	if f.DeployedAt != "" {
		fmt.Printf("  Deployed:   %s\n", formatTimestamp(f.DeployedAt))
	}
	if f.UpdatedAt != "" {
		fmt.Printf("  Updated:    %s\n", formatTimestamp(f.UpdatedAt))
	}

	// Verbose: internal details
	if verbose {
		fmt.Println()
		ui.Dimmed("  Internal:")
		if f.Port > 0 {
			fmt.Printf("  Port:       %d\n", f.Port)
		}
		if f.PID > 0 {
			fmt.Printf("  PID:        %d\n", f.PID)
		}
	}

	fmt.Println()

	return nil
}

// runFormationDelete handles muxi formation delete <id>
func runFormationDelete(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	force, _ := cmd.Flags().GetBool("force")
	atomic, _ := cmd.Flags().GetBool("atomic")
	skipConfirm := force || atomic
	formationID := args[0]

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Check if formation exists first
	_, err = client.GetFormation(formationID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not found",
				fmt.Sprintf("Formation '%s' does not exist on this server.", formationID),
				ui.Command("muxi formation list"),
			)
			os.Exit(1)
		}
		return err
	}

	// Confirm deletion unless --force or --atomic
	if !skipConfirm {
		fmt.Println()
		fmt.Printf("  Delete formation '%s'?\n", formationID)
		fmt.Println()
		fmt.Printf("  %s\n", ui.RedText("⚠ This action cannot be undone!"))
		fmt.Print("  Enter formation ID to confirm: ")

		var confirm string
		fmt.Scanln(&confirm)

		if confirm != formationID {
			fmt.Println()
			ui.Dimmed("  Deletion cancelled")
			fmt.Println()
			return nil
		}
	}

	// Delete formation
	spinner := ui.NewSpinner("Deleting formation...")
	spinner.Start()

	err = client.DeleteFormation(formationID)
	if err != nil {
		spinner.StopWithError("Delete failed")
		return err
	}

	spinner.StopWithSuccess("Deleted formation")

	fmt.Println()
	ui.Success(fmt.Sprintf("Deleted %s", formationID))
	fmt.Println()

	return nil
}

// formatDurationShort formats seconds into short duration (5d 12h)
func formatDurationShort(seconds int64) string {
	d := time.Duration(seconds) * time.Second

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

// formatTimestamp formats an ISO timestamp for display
func formatTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02 15:04:05")
}

// runFormationStop handles muxi formation stop <id>
func runFormationStop(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	formationID := args[0]

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Check if formation exists
	f, err := client.GetFormation(formationID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not found",
				fmt.Sprintf("Formation '%s' does not exist on this server.", formationID),
				ui.Command("muxi formation list"),
			)
			os.Exit(1)
		}
		return err
	}

	// Check if already stopped
	if f.Status == "stopped" {
		fmt.Println()
		ui.Warning(fmt.Sprintf("Formation '%s' is already stopped", formationID))
		fmt.Println()
		return nil
	}

	spinner := ui.NewSpinner("Stopping formation...")
	spinner.Start()

	err = client.StopFormation(formationID)
	if err != nil {
		spinner.StopWithError("Stop failed")
		return err
	}

	spinner.StopWithSuccess("Stopped formation")

	fmt.Println()
	ui.Success(fmt.Sprintf("Stopped %s", formationID))
	fmt.Println()

	return nil
}

// runFormationRestart handles muxi formation restart <id>
func runFormationRestart(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	formationID := args[0]

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Check if formation exists
	_, err = client.GetFormation(formationID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not found",
				fmt.Sprintf("Formation '%s' does not exist on this server.", formationID),
				ui.Command("muxi formation list"),
			)
			os.Exit(1)
		}
		return err
	}

	fmt.Println()
	fmt.Printf("  Restarting: %s\n", ui.BoldText(formationID))
	fmt.Println()

	spinner := ui.NewSpinner("Sending restart request...")
	spinner.Start()

	var lastStage string
	var lastProgress *server.DeployProgressEvent

	// SSE callback - called for each progress event
	callback := func(event server.SSEEvent) error {
		if event.Event == "progress" {
			var progress server.DeployProgressEvent
			if err := json.Unmarshal([]byte(event.Data), &progress); err != nil {
				return nil // Skip malformed events
			}

			// Same stage with updated progress - just update the message
			if lastStage == progress.Stage {
				spinner.UpdateMessage(formatRestartStageMessage(progress))
				lastProgress = &progress
				return nil
			}

			// First server event: finish "Sending restart request" and show it succeeded
			if lastStage == "" {
				spinner.StopWithSuccess("Restart initiated")
			} else {
				// Complete previous stage
				spinner.StopWithSuccess(formatRestartStageComplete(lastStage, lastProgress))
			}

			// Start new spinner for this stage
			lastStage = progress.Stage
			lastProgress = &progress
			// Use padded spinner for health_check to give margin from terminal bottom
			if progress.Stage == "health_check" {
				spinner = ui.NewSpinnerWithPadding(formatRestartStageMessage(progress), 1)
			} else {
				spinner = ui.NewSpinner(formatRestartStageMessage(progress))
			}
			spinner.Start()
		}
		return nil
	}

	// Execute restart with streaming
	_, restartErr := client.RestartFormationStreaming(formationID, callback)

	// Handle final state
	if restartErr != nil {
		if lastStage != "" {
			spinner.StopWithError(ui.DimmedText("[SERVER]") + " Failed")
		} else {
			spinner.StopWithError("Restart failed")
		}
		playRestartNotificationSound(false)
		fmt.Println()
		return restartErr
	}

	// Complete final stage
	if lastStage != "" {
		spinner.StopWithSuccess(formatRestartStageComplete(lastStage, lastProgress))
	} else {
		spinner.StopWithSuccess("Restart complete")
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Formation `%s` restarted!", formationID))
	fmt.Println()

	playRestartNotificationSound(true)

	return nil
}

// formatRestartStageMessage formats a progress event into a spinner message with [SERVER] prefix
func formatRestartStageMessage(p server.DeployProgressEvent) string {
	prefix := ui.DimmedText("[SERVER]") + " "

	switch p.Stage {
	case "stopping":
		return prefix + "Stopping current process..."
	case "validating":
		return prefix + "Validating formation files..."
	case "resolving_runtime":
		return prefix + "Resolving runtime version..."
	case "downloading_sif":
		if p.Progress > 0 {
			return prefix + fmt.Sprintf("Downloading runtime image... %d%%", p.Progress)
		}
		return prefix + "Downloading runtime image..."
	case "pulling_runner":
		return prefix + "Pulling runtime runner..."
	case "spawning":
		return prefix + "Starting formation..."
	case "health_check":
		if p.Attempt > 0 && p.MaxAttempts > 0 {
			remaining := p.MaxAttempts - p.Attempt
			return prefix + fmt.Sprintf("Waiting for formation to start (%d)", remaining)
		}
		return prefix + "Waiting for formation to start..."
	default:
		if p.Message != "" {
			return prefix + p.Message
		}
		return prefix + p.Stage + "..."
	}
}

// formatRestartStageComplete formats the completion message for a restart stage
func formatRestartStageComplete(stage string, p *server.DeployProgressEvent) string {
	prefix := ui.DimmedText("[SERVER]") + " "

	switch stage {
	case "stopping":
		return prefix + "Stopped current process"
	case "validating":
		return prefix + "Validated formation files"
	case "resolving_runtime":
		if p != nil && p.Version != "" {
			return prefix + "Resolved runtime " + p.Version
		}
		return prefix + "Resolved runtime version"
	case "downloading_sif":
		return prefix + "Downloaded runtime image"
	case "pulling_runner":
		return prefix + "Pulled runtime runner"
	case "spawning":
		return prefix + "Started formation"
	case "health_check":
		return prefix + "Formation started"
	default:
		return prefix + stage + " complete"
	}
}

// playRestartNotificationSound plays a sound to notify the user of restart completion
func playRestartNotificationSound(success bool) {
	if runtime.GOOS == "darwin" {
		sound := "/System/Library/Sounds/Glass.aiff"
		if !success {
			sound = "/System/Library/Sounds/Sosumi.aiff"
		}
		exec.Command("afplay", sound).Run()
	} else {
		fmt.Print("\a")
	}
}

// runFormationRollback handles muxi formation rollback <id>
func runFormationRollback(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	formationID := args[0]

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Check if formation exists
	_, err = client.GetFormation(formationID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not found",
				fmt.Sprintf("Formation '%s' does not exist on this server.", formationID),
				ui.Command("muxi formation list"),
			)
			os.Exit(1)
		}
		return err
	}

	spinner := ui.NewSpinner("Rolling back formation...")
	spinner.Start()

	resp, err := client.RollbackFormation(formationID)
	if err != nil {
		spinner.StopWithError("Rollback failed")
		return err
	}

	spinner.StopWithSuccess("Rolled back formation")

	fmt.Println()
	if resp != nil && resp.PreviousVersion != "" {
		ui.Success(fmt.Sprintf("Rolled back %s to version %s", formationID, resp.PreviousVersion))
	} else {
		ui.Success(fmt.Sprintf("Rolled back %s", formationID))
	}
	fmt.Println()

	return nil
}

// runFormationLogs handles muxi formation logs <id>
func runFormationLogs(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	lines, _ := cmd.Flags().GetInt("lines")
	stream, _ := cmd.Flags().GetString("stream")
	follow, _ := cmd.Flags().GetBool("follow")
	formationID := args[0]

	// Default stream to "all" if not specified
	if stream == "" {
		stream = "all"
	}

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Check if formation exists
	_, err = client.GetFormation(formationID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not found",
				fmt.Sprintf("Formation '%s' does not exist on this server.", formationID),
				ui.Command("muxi formation list"),
			)
			os.Exit(1)
		}
		return err
	}

	// Follow mode - stream logs via SSE
	if follow {
		return streamFormationLogs(client, formationID, stream)
	}

	// Non-follow mode - get recent logs
	resp, err := client.GetFormationLogs(formationID, lines, stream)
	if err != nil {
		return err
	}

	// Combine logs based on stream filter
	var allLogs []string
	if stream == "all" || stream == "stdout" {
		allLogs = append(allLogs, resp.Logs.Stdout...)
	}
	if stream == "all" || stream == "stderr" {
		allLogs = append(allLogs, resp.Logs.Stderr...)
	}

	if len(allLogs) == 0 {
		fmt.Println()
		ui.Dimmed("  No logs available")
		fmt.Println()
		return nil
	}

	// Print logs
	for _, line := range allLogs {
		fmt.Println(line)
	}

	return nil
}

// streamFormationLogs streams logs via SSE (follow mode)
func streamFormationLogs(client *server.Client, formationID, stream string) error {
	fmt.Printf("Streaming logs for %s (Ctrl+C to stop)...\n\n", formationID)

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		err := client.StreamFormationLogs(formationID, stream, func(event server.LogEvent) error {
			if event.Line != "" {
				fmt.Println(event.Line)
			}
			return nil
		})
		errChan <- err
	}()

	select {
	case <-sigChan:
		fmt.Println("\n\nStopped streaming.")
		return nil
	case err := <-errChan:
		return err
	}
}
