package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var agentsCmd = &cobra.Command{
	Use:     "agents",
	Short:   "Manage formation agents",
	GroupID: "formation",
	Long:    `List and manage agents configured in a formation.`,
}

var agentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List agents in the formation",
	RunE:    runAgentsList,
}

var agentsShowCmd = &cobra.Command{
	Use:   "show <agent-id>",
	Short: "Show agent configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runAgentsShow,
}

func init() {
	rootCmd.AddCommand(agentsCmd)

	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsShowCmd)

	// Flags for list command
	formation.AddCommonFlags(agentsListCmd)
	agentsListCmd.Flags().BoolP("verbose", "v", false, "Show detailed agent information")

	// Flags for show command
	formation.AddFormationFlag(agentsShowCmd)
	formation.AddProfileFlag(agentsShowCmd)
}

func runAgentsList(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	resp, err := client.GetAgents()
	if err != nil {
		return err
	}

	if len(resp.Agents) == 0 {
		fmt.Println()
		ui.Dimmed("  No agents configured")
		fmt.Println()
		return nil
	}

	fmt.Println()

	if verbose {
		// Verbose: show detailed info for each agent
		for i, agent := range resp.Agents {
			if i > 0 {
				fmt.Println()
				fmt.Println("  " + ui.DimmedText("───────────────────────────────────────"))
				fmt.Println()
			}
			displayAgentVerbose(agent)
		}
	} else {
		// Table format
		fmt.Printf("  %-20s %-14s %-13s %s\n",
			"ID", "ROLE", "STATUS", "MODEL")
		fmt.Printf("  %-20s %-14s %-13s %s\n",
			"──────────────────", "────────────", "───────────", "─────────────────────")

		for _, agent := range resp.Agents {
			statusIcon := ui.GreenText("●")
			statusText := "active"
			if agent.Status == "disabled" {
				statusIcon = ui.DimmedText("○")
				statusText = "disabled"
			} else if agent.Status != "" && agent.Status != "active" {
				statusIcon = ui.DimmedText("○")
				statusText = agent.Status
			}

			model := agent.Model
			if model == "" {
				model = ui.DimmedText("-")
			}

			fmt.Printf("  %-20s %-14s %s %-10s %s\n",
				truncateStr(agent.ID, 20),
				truncateStr(agent.Role, 14),
				statusIcon,
				statusText,
				model)
		}
	}

	fmt.Println()
	return nil
}

func displayAgentVerbose(agent formation.Agent) {
	fmt.Printf("  %s\n", ui.BoldText(agent.ID))
	if agent.Name != "" && agent.Name != agent.ID {
		fmt.Printf("   Name:        %s\n", agent.Name)
	}
	fmt.Printf("   Role:        %s\n", agent.Role)
	if agent.Description != "" {
		fmt.Printf("   Description: %s\n", agent.Description)
	}

	// Status
	statusIcon := ui.GreenText("●")
	statusText := "active"
	if agent.Status == "disabled" {
		statusIcon = ui.DimmedText("○")
		statusText = "disabled"
	} else if agent.Status != "" && agent.Status != "active" {
		statusIcon = ui.DimmedText("○")
		statusText = agent.Status
	}
	fmt.Printf("   Status:      %s %s\n", statusIcon, statusText)

	// Model info
	if agent.Model != "" {
		fmt.Printf("   Model:       %s\n", agent.Model)
	}
	if agent.Provider != "" {
		fmt.Printf("   Provider:    %s\n", agent.Provider)
	}

	// Tools and MCP servers
	if len(agent.Tools) > 0 {
		fmt.Printf("   Tools:       %d configured\n", len(agent.Tools))
	}
	if len(agent.MCPServers) > 0 {
		fmt.Printf("   MCP Servers: %d connected\n", len(agent.MCPServers))
	}
}

// truncateStr truncates a string to max length with ellipsis
func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func runAgentsShow(cmd *cobra.Command, args []string) error {
	agentID := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	config, err := client.GetAgent(agentID)
	if err != nil {
		fmt.Println()
		fmt.Printf("  Agent '%s' not found\n", agentID)
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

	// Render with syntax highlighting
	fmt.Println()
	fmt.Println(ui.RenderYAML(string(yamlBytes)))

	return nil
}
