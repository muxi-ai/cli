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

func init() {
	rootCmd.AddCommand(formationCmd)

	formationCmd.AddCommand(formationListCmd)
	formationCmd.AddCommand(formationGetCmd)

	// Flags
	formationListCmd.Flags().String("profile", "", "Server profile to use")
	formationGetCmd.Flags().String("profile", "", "Server profile to use")
	formationGetCmd.Flags().BoolP("verbose", "v", false, "Show internal details (port, pid)")
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
		fmt.Println("  Deploy a formation: muxi deploy")
		fmt.Println()
		return nil
	}

	// Header
	fmt.Println()
	fmt.Printf("  %-20s %-10s %s\n",
		"ID", "STATUS", "UPTIME")
	fmt.Printf("  %-20s %-10s %s\n",
		"──────────────────", "────────", "──────")

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

		fmt.Printf("  %-20s %s %-8s %s\n",
			f.ID, statusIcon, statusText, uptimeStr)
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
				"muxi formation list",
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
		fmt.Printf("  Name:       %s\n", f.Name)
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

	// Health
	var healthDisplay string
	if f.Healthy {
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
