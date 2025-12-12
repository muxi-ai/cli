package cmd

import (
	"fmt"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var mcpCmd = &cobra.Command{
	Use:     "mcp",
	Short:   "Manage MCP servers",
	GroupID: "formation",
	Long:    `List and manage MCP (Model Context Protocol) servers configured in a formation.`,
}

var mcpListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List MCP servers in the formation",
	RunE:    runMCPList,
}

var mcpShowCmd = &cobra.Command{
	Use:   "show <server-id>",
	Short: "Show MCP server configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runMCPShow,
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.AddCommand(mcpListCmd)
	mcpCmd.AddCommand(mcpShowCmd)

	// Flags for list command
	formation.AddCommonFlags(mcpListCmd)
	mcpListCmd.Flags().BoolP("verbose", "v", false, "Show detailed MCP server information including tools")

	// Flags for show command
	formation.AddFormationFlag(mcpShowCmd)
	formation.AddProfileFlag(mcpShowCmd)
	mcpShowCmd.Flags().Bool("raw", false, "Output raw YAML without formatting (for piping)")
}

func runMCPList(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	resp, err := client.GetMCPServers()
	if err != nil {
		return err
	}

	// Handle both "servers" and "mcp_servers" field names from API
	servers := resp.Servers
	if len(servers) == 0 && len(resp.MCPServers) > 0 {
		servers = resp.MCPServers
	}

	if len(servers) == 0 {
		fmt.Println()
		ui.Dimmed("  No MCP servers configured")
		fmt.Println()
		return nil
	}

	fmt.Println()

	if verbose {
		// Verbose: show detailed info for each server
		for i, server := range servers {
			if i > 0 {
				fmt.Println()
				fmt.Println("  " + ui.DimmedText("───────────────────────────────────────"))
				fmt.Println()
			}
			displayMCPServerVerbose(server)
		}
	} else {
		// Table format
		fmt.Printf("  %-16s %-10s %-14s %s\n",
			"ID", "TYPE", "STATUS", "TOOLS")
		fmt.Printf("  %-16s %-10s %-14s %s\n",
			"──────────────", "────────", "────────────", "─────")

		for _, server := range servers {
			statusIcon := ui.GreenText("●")
			statusText := "connected"
			if server.Status == "disabled" {
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

			fmt.Printf("  %-16s %-10s %s %-12s %s\n",
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
	fmt.Printf("  %s\n", ui.BoldText(server.ID))
	if server.Name != "" && server.Name != server.ID {
		fmt.Printf("   Name:        %s\n", server.Name)
	}
	fmt.Printf("   Type:        %s\n", server.Type)
	if server.Description != "" {
		fmt.Printf("   Description: %s\n", server.Description)
	}

	// Status
	statusIcon := ui.GreenText("●")
	statusText := "connected"
	if server.Status == "disabled" {
		statusIcon = ui.DimmedText("○")
		statusText = "disabled"
	} else if server.Status != "" && server.Status != "connected" && server.Status != "active" {
		statusIcon = ui.DimmedText("○")
		statusText = server.Status
	}
	fmt.Printf("   Status:      %s %s\n", statusIcon, statusText)

	// Tools
	toolsCount := server.ToolsCount
	if toolsCount == 0 && len(server.Tools) > 0 {
		toolsCount = len(server.Tools)
	}

	if toolsCount > 0 {
		fmt.Printf("   Tools:       %d available\n", toolsCount)
		if len(server.Tools) > 0 {
			// Show first few tools
			maxShow := 5
			if len(server.Tools) <= maxShow {
				fmt.Printf("                %s\n", strings.Join(server.Tools, ", "))
			} else {
				fmt.Printf("                %s, +%d more\n",
					strings.Join(server.Tools[:maxShow], ", "),
					len(server.Tools)-maxShow)
			}
		}
	}
}

func runMCPShow(cmd *cobra.Command, args []string) error {
	serverID := args[0]
	raw, _ := cmd.Flags().GetBool("raw")

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	if !raw {
		formation.PrintBadgeFromFlags(cmd)
	}

	config, err := client.GetMCPServer(serverID)
	if err != nil {
		if raw {
			return fmt.Errorf("MCP server '%s' not found", serverID)
		}
		fmt.Println()
		fmt.Printf("  MCP server '%s' not found\n", serverID)
		fmt.Println()
		return nil
	}

	// Remove "source" key
	delete(config, "source")

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	if raw {
		// Raw output for piping
		fmt.Print(string(yamlBytes))
	} else {
		// Formatted output with syntax highlighting and indentation
		fmt.Println()
		fmt.Println(ui.IndentString(ui.RenderYAML(string(yamlBytes)), 3))
	}

	return nil
}
