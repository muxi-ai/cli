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

		if err := scaffold.CreateAgent(name, noWizard); err != nil {
			return fmt.Errorf("failed to create agent: %w", err)
		}

		return nil
	},
}

var newMcpCmd = &cobra.Command{
	Use:   "mcp <name>",
	Short: "Create a new MCP server",
	Long: `Create a new MCP server configuration file in mcps/ directory.

Must be run inside a formation directory.

Examples:
  muxi new mcp postgres
  muxi new mcp postgres --no-wizard`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := scaffold.CreateMCP(name, noWizard); err != nil {
			return fmt.Errorf("failed to create MCP server: %w", err)
		}

		return nil
	},
}

var newSopCmd = &cobra.Command{
	Use:   "sop <name>",
	Short: "Create a new SOP document",
	Long: `Create a new Standard Operating Procedure document in sops/ directory.

Must be run inside a formation directory.

Examples:
  muxi new sop customer-onboarding
  muxi new sop customer-onboarding --no-wizard`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := scaffold.CreateSOP(name, noWizard); err != nil {
			return fmt.Errorf("failed to create SOP: %w", err)
		}

		return nil
	},
}

var newTriggerCmd = &cobra.Command{
	Use:   "trigger <name>",
	Short: "Create a new trigger",
	Long: `Create a new trigger configuration file in triggers/ directory.

Must be run inside a formation directory.

Examples:
  muxi new trigger webhook
  muxi new trigger webhook --no-wizard`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := scaffold.CreateTrigger(name, noWizard); err != nil {
			return fmt.Errorf("failed to create trigger: %w", err)
		}

		return nil
	},
}

var newA2ACmd = &cobra.Command{
	Use:   "a2a <name>",
	Short: "Create a new A2A configuration",
	Long: `Create a new Agent-to-Agent communication configuration in a2a/ directory.

Must be run inside a formation directory.

Examples:
  muxi new a2a external-api
  muxi new a2a external-api --no-wizard`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := scaffold.CreateA2A(name, noWizard); err != nil {
			return fmt.Errorf("failed to create A2A config: %w", err)
		}

		return nil
	},
}

func init() {
	// Add --no-wizard flag to all new commands
	newFormationCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newAgentCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newMcpCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newSopCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newTriggerCmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")
	newA2ACmd.Flags().BoolVar(&noWizard, "no-wizard", false, "Skip interactive prompts")

	// Add subcommands to new
	newCmd.AddCommand(newFormationCmd)
	newCmd.AddCommand(newAgentCmd)
	newCmd.AddCommand(newMcpCmd)
	newCmd.AddCommand(newSopCmd)
	newCmd.AddCommand(newTriggerCmd)
	newCmd.AddCommand(newA2ACmd)
}
