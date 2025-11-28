package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/scaffold"
	"github.com/spf13/cobra"
)

var (
	noWizard bool
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create formation or components",
	Long: `Create formation scaffolding or component files.

Examples:
  muxi new formation              # Interactive wizard
  muxi new formation my-bot       # Create with name
  muxi new formation my-bot --no-wizard  # Skip wizard
  muxi new agent weather          # Create agent (in formation dir)
  muxi new mcp postgres           # Create MCP server
  muxi new sop onboarding         # Create SOP document`,
}

var newFormationCmd = &cobra.Command{
	Use:   "formation [name]",
	Short: "Create a new formation",
	Long: `Create a new MUXI formation with complete directory structure.

This creates:
  - Directory structure (agents/, mcps/, a2a/, sops/, triggers/, knowledge/)
  - Configuration files (.gitignore, .muxi, formation.yaml)
  - Encryption key (.key)
  - Secrets template (secrets)
  - README.md with getting started guide

Examples:
  muxi new formation              # Interactive wizard (asks for name)
  muxi new formation my-bot       # Create with name (runs wizard)
  muxi new formation my-bot --no-wizard  # Skip wizard entirely`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		// Create formation
		if err := scaffold.CreateFormation(name, noWizard); err != nil {
			return fmt.Errorf("failed to create formation: %w", err)
		}

		return nil
	},
}

var newAgentCmd = &cobra.Command{
	Use:   "agent [name]",
	Short: "Create a new agent",
	Long: `Create a new agent configuration file in agents/ directory.

Must be run inside a formation directory.

Examples:
  muxi new agent              # Interactive wizard (prompts for ID)
  muxi new agent weather      # Create with ID
  muxi new agent weather --no-wizard`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		// Don't wrap error - ErrorBlock already shown
		return scaffold.CreateAgent(name, noWizard)
	},
}

var newMcpCmd = &cobra.Command{
	Use:   "mcp [name]",
	Short: "Create a new MCP server",
	Long: `Create a new MCP server configuration file in mcps/ directory.

Must be run inside a formation directory.

Examples:
  muxi new mcp                        # Interactive wizard (prompts for ID)
  muxi new mcp weather-api            # Create formation-level MCP
  muxi new mcp --agent weather-bot    # Create agent-specific MCP
  muxi new mcp postgres --no-wizard   # Non-interactive`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		agentID, _ := cmd.Flags().GetString("agent")

		// Don't wrap error - ErrorBlock already shown
		return scaffold.CreateMCP(name, agentID, noWizard)
	},
}

var newSopCmd = &cobra.Command{
	Use:   "sop [name]",
	Short: "Create a new SOP document",
	Long: `Create a new Standard Operating Procedure (SOP) in sops/ directory.

SOPs define workflows that agents follow to complete complex tasks.
They include frontmatter (name, description, mode, tags) and step-by-step
instructions with agent assignments and tool references.

Must be run inside a formation directory.

Examples:
  muxi new sop                              # Interactive wizard
  muxi new sop customer-onboarding          # With ID
  muxi new sop "Customer Onboarding"        # Spaces are normalized
  muxi new sop customer-onboarding --no-wizard`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		if err := scaffold.CreateSOP(name, noWizard); err != nil {
			return fmt.Errorf("failed to create SOP: %w", err)
		}

		return nil
	},
}

var newTriggerCmd = &cobra.Command{
	Use:   "trigger [name]",
	Short: "Create a new trigger",
	Long: `Create a new trigger prompt template in triggers/ directory.

Triggers are markdown templates invoked via webhooks. Use ${{ data.xxx }}
syntax to access values from the webhook payload.

Must be run inside a formation directory.

Examples:
  muxi new trigger                          # Interactive wizard
  muxi new trigger github-issue             # With ID
  muxi new trigger "GitHub Issue"           # Spaces are normalized
  muxi new trigger github-issue --no-wizard`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		if err := scaffold.CreateTrigger(name, noWizard); err != nil {
			return fmt.Errorf("failed to create trigger: %w", err)
		}

		return nil
	},
}

var newA2AServiceCmd = &cobra.Command{
	Use:   "a2a-service [name]",
	Short: "Create a new A2A service configuration",
	Long: `Create a new A2A service configuration file in a2a/ directory.

A2A services define external agent-to-agent endpoints that your formation
can communicate with. This includes connection details, authentication,
and optional rate limiting settings.

Must be run inside a formation directory.

Examples:
  muxi new a2a-service                    # Interactive wizard
  muxi new a2a-service external-billing   # Create with ID
  muxi new a2a-service external-billing --no-wizard`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		return scaffold.CreateA2AService(name, noWizard)
	},
}

func init() {
	// Add --no-wizard flag to all new commands
	newFormationCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newAgentCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newMcpCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newSopCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newTriggerCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newA2AServiceCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")

	// Add MCP-specific flags
	newMcpCmd.Flags().String("agent", "", "Create MCP for specific agent (agent-specific MCP)")

	// Add subcommands to new
	newCmd.AddCommand(newFormationCmd)
	newCmd.AddCommand(newAgentCmd)
	newCmd.AddCommand(newMcpCmd)
	newCmd.AddCommand(newSopCmd)
	newCmd.AddCommand(newTriggerCmd)
	newCmd.AddCommand(newA2AServiceCmd)
}
