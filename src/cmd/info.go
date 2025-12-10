package cmd

import (
	"fmt"
	"time"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show formation status and info",
	GroupID: "formation",
	Long: `Show status and configuration for a formation.

Displays runtime status including agents, MCP servers, memory usage, and stats.
Use --full to include detailed configuration information.

Requires admin API key (from secrets.enc or MUXI_ADMIN_KEY).`,
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
	formation.AddCommonFlags(infoCmd)
	infoCmd.Flags().Bool("full", false, "Show full configuration details")
}

func runInfo(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	full, _ := cmd.Flags().GetBool("full")

	// Get health status first
	health, err := client.Health()
	if err != nil {
		return fmt.Errorf("failed to get health: %w", err)
	}

	// Get detailed status
	status, err := client.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Print formation info
	fmt.Println()
	fmt.Printf("  Formation: %s\n", ui.BoldText(status.Formation.Name))
	fmt.Println()

	// Status line with colored indicator based on health
	statusColor := ui.GreenText("●")
	statusText := health.Status
	switch health.Status {
	case "healthy":
		statusColor = ui.GreenText("●")
	case "degraded":
		statusColor = ui.YellowText("●")
	default: // unhealthy or unknown
		statusColor = ui.RedText("●")
	}
	fmt.Printf("    Status:     %s %s\n", statusColor, statusText)
	fmt.Printf("    Version:    %s\n", status.Formation.Version)
	fmt.Printf("    Uptime:     %s\n", formatUptime(status.Server.UptimeSeconds))
	fmt.Println()

	// Agents and MCP
	agentStatus := fmt.Sprintf("%d", status.Agents.Count)
	if status.Agents.Active > 0 && status.Agents.Active != status.Agents.Count {
		agentStatus = fmt.Sprintf("%d (%d active)", status.Agents.Count, status.Agents.Active)
	}
	fmt.Printf("    Agents:     %s\n", agentStatus)

	mcpStatus := fmt.Sprintf("%d servers", status.MCPServers.Count)
	if status.MCPServers.Active > 0 {
		mcpStatus = fmt.Sprintf("%d servers connected", status.MCPServers.Active)
	}
	fmt.Printf("    MCP:        %s\n", mcpStatus)

	// Memory
	memoryInfo := fmt.Sprintf("%.0f MB working", status.Stats.Memory.WorkingMemoryMB)
	if status.Stats.Memory.MemoryUsageMB > 0 {
		memoryInfo = fmt.Sprintf("%.0f MB working, %.1f MB usage",
			status.Stats.Memory.WorkingMemoryMB, status.Stats.Memory.MemoryUsageMB)
	}
	fmt.Printf("    Memory:     %s\n", memoryInfo)
	fmt.Println()

	// Stats
	fmt.Println("    Stats:")
	requestInfo := fmt.Sprintf("%d total", status.Stats.Requests.Total)
	if status.Stats.Requests.Active > 0 {
		requestInfo = fmt.Sprintf("%d total (%d active)", status.Stats.Requests.Total, status.Stats.Requests.Active)
	}
	fmt.Printf("      Requests:   %s\n", requestInfo)
	fmt.Printf("      CPU:        %.0f%%\n", status.Stats.CPUPercent)
	fmt.Println()

	// Full config if requested
	if full {
		config, err := client.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		fmt.Println("    Configuration:")
		fmt.Printf("      Schema:     %s\n", config.SchemaVersion)
		fmt.Printf("      Agents:     %d total\n", config.Agents.Total)
		fmt.Printf("      Secrets:    %d total\n", config.Secrets.Total)
		fmt.Printf("      MCP:        %d servers (timeout: %ds, retries: %d)\n",
			config.MCP.Servers.Total, config.MCP.DefaultTimeoutSeconds, config.MCP.DefaultRetryAttempts)
		fmt.Println()
	}

	return nil
}

func formatUptime(seconds int64) string {
	if seconds == 0 {
		return "-"
	}

	d := time.Duration(seconds) * time.Second
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
