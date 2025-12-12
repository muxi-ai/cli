package cmd

import (
	"bytes"
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
	Long: `List and manage agents configured in a running formation.

Requires connection to a Formation API server. Use -F to specify a formation
and -p to specify a server profile.`,
}

var agentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List agents in the formation",
	Long: `List all agents configured in the formation.

Displays agent ID, role, status, and model. Use -v for detailed output
including description and system message preview.`,
	RunE: runAgentsList,
}

var agentsShowCmd = &cobra.Command{
	Use:   "show <agent-id>",
	Short: "Show agent configuration",
	Long: `Display an agent's full configuration as YAML with syntax highlighting.

Use --raw to output plain YAML without formatting, suitable for piping
to files or other tools.`,
	Args: cobra.ExactArgs(1),
	RunE: runAgentsShow,
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
	agentsShowCmd.Flags().Bool("raw", false, "Output raw YAML without formatting (for piping)")
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

	// Use AgentList if Agents is empty (per spec)
	agents := resp.Agents
	if len(agents) == 0 {
		agents = resp.AgentList
	}

	if len(agents) == 0 {
		fmt.Println()
		ui.Dimmed("  No agents configured")
		fmt.Println()
		return nil
	}

	fmt.Println()

	if verbose {
		// Verbose: show detailed info for each agent
		for i, agent := range agents {
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

		for _, agent := range agents {
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
	raw, _ := cmd.Flags().GetBool("raw")

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	if !raw {
		formation.PrintBadgeFromFlags(cmd)
	}

	config, err := client.GetAgent(agentID)
	if err != nil {
		if raw {
			return fmt.Errorf("agent '%s' not found", agentID)
		}
		fmt.Println()
		fmt.Printf("  Agent '%s' not found\n", agentID)
		fmt.Println()
		return nil
	}

	// Remove "source" key
	delete(config, "source")

	// Convert to YAML with 2-space indent
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(config); err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}
	yamlBytes := buf.Bytes()

	if raw {
		// Raw output for piping
		fmt.Print(string(yamlBytes))
	} else {
		// Formatted output with syntax highlighting and indentation
		fmt.Println()
		fmt.Println(ui.IndentString(ui.RenderYAML(string(yamlBytes)), 2))
	}

	return nil
}
