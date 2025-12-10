package cmd

import (
	"fmt"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:     "mcp",
	Short:   "Manage MCP servers",
	GroupID: "formation",
	Long:    `List and manage MCP (Model Context Protocol) servers configured in a formation.`,
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List MCP servers in the formation",
	RunE:  runMCPList,
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.AddCommand(mcpListCmd)

	// Flags for list command
	formation.AddCommonFlags(mcpListCmd)
	mcpListCmd.Flags().BoolP("verbose", "v", false, "Show detailed MCP server information including tools")
}

func runMCPList(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	resp, err := client.GetMCPServers()
	if err != nil {
		return err
	}

	if len(resp.Servers) == 0 {
		fmt.Println()
		ui.Dimmed("  No MCP servers configured")
		fmt.Println()
		return nil
	}

	fmt.Println()

	if verbose {
		// Verbose: show detailed info for each server
		for i, server := range resp.Servers {
			if i > 0 {
				fmt.Println()
			}
			displayMCPServerVerbose(server)
		}
	} else {
		// Table format
		fmt.Printf("  %-16s %-10s %-12s %s\n",
			"ID", "TYPE", "STATUS", "TOOLS")
		fmt.Printf("  %-16s %-10s %-12s %s\n",
			"──────────────", "────────", "──────────", "─────")

		for _, server := range resp.Servers {
			statusIcon := ui.GreenText("●")
			statusText := "connected"
			if !server.Enabled {
				statusIcon = ui.DimmedText("○")
				statusText = "disabled"
			} else if server.Status != "" && server.Status != "connected" && server.Status != "active" {
				statusIcon = ui.DimmedText("○")
				statusText = server.Status
			}

			toolsCount := server.ToolsCount
			if toolsCount == 0 && len(server.Tools) > 0 {
				toolsCount = len(server.Tools)
			}
			toolsStr := fmt.Sprintf("%d tools", toolsCount)
			if toolsCount == 1 {
				toolsStr = "1 tool"
			}
			if toolsCount == 0 {
				toolsStr = ui.DimmedText("-")
			}

			fmt.Printf("  %-16s %-10s %s %-10s %s\n",
				truncateStr(server.ID, 16),
				server.Type,
				statusIcon,
				statusText,
				toolsStr)
		}
	}

	fmt.Println()
	return nil
}

func displayMCPServerVerbose(server formation.MCPServer) {
	fmt.Printf("  Server: %s\n", server.ID)
	if server.Name != "" && server.Name != server.ID {
		fmt.Printf("  Name:        %s\n", server.Name)
	}
	fmt.Printf("  Type:        %s\n", server.Type)
	if server.Description != "" {
		fmt.Printf("  Description: %s\n", server.Description)
	}
	fmt.Println()

	// Status
	statusIcon := ui.GreenText("●")
	statusText := "connected"
	if !server.Enabled {
		statusIcon = ui.DimmedText("○")
		statusText = "disabled"
	} else if server.Status != "" && server.Status != "connected" && server.Status != "active" {
		statusIcon = ui.DimmedText("○")
		statusText = server.Status
	}
	fmt.Printf("  Status:      %s %s\n", statusIcon, statusText)

	// Tools
	toolsCount := server.ToolsCount
	if toolsCount == 0 && len(server.Tools) > 0 {
		toolsCount = len(server.Tools)
	}

	if toolsCount > 0 {
		fmt.Printf("  Tools:       %d available\n", toolsCount)
		if len(server.Tools) > 0 {
			// Show first few tools
			maxShow := 5
			if len(server.Tools) <= maxShow {
				fmt.Printf("               %s\n", strings.Join(server.Tools, ", "))
			} else {
				fmt.Printf("               %s, +%d more\n",
					strings.Join(server.Tools[:maxShow], ", "),
					len(server.Tools)-maxShow)
			}
		}
	}
}
