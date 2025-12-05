package cmd

import (
	"fmt"
	"os"
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
		version := f.Version
		if version == "" {
			version = "-"
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
	if f.Version != "" {
		fmt.Printf("  Version:   %s\n", f.Version)
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

	spinner := ui.NewSpinner("Restarting formation...")
	spinner.Start()

	err = client.RestartFormation(formationID)
	if err != nil {
		spinner.StopWithError("Restart failed")
		return err
	}

	spinner.StopWithSuccess("Restarted formation")

	fmt.Println()
	ui.Success(fmt.Sprintf("Restarted %s", formationID))
	fmt.Println()

	return nil
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

	resp, err := client.GetFormationLogs(formationID, lines, stream)
	if err != nil {
		return err
	}

	if len(resp.Lines) == 0 {
		fmt.Println()
		ui.Dimmed("  No logs available")
		fmt.Println()
		return nil
	}

	// Print logs
	for _, line := range resp.Lines {
		fmt.Println(line)
	}

	return nil
}
