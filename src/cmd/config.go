package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/scaffold"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure formation settings",
	Long: `Configure settings in formation.yaml.

This command provides interactive wizards for configuring various
formation-level settings like A2A communication, LLM providers,
logging, and more.

Must be run inside a formation directory.`,
}

var configAsyncCmd = &cobra.Command{
	Use:   "async",
	Short: "Configure async response settings",
	Long: `Configure async response settings in formation.yaml.

This command provides an interactive wizard for:
  - Enabling/disabling async responses
  - Setting the time threshold before switching to async
  - Configuring webhook URL for async delivery
  - Setting retry count and delay

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard
  muxi config async`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scaffold.ConfigureAsync(); err != nil {
			return fmt.Errorf("failed to configure async: %w", err)
		}
		return nil
	},
}

var configLLMCmd = &cobra.Command{
	Use:   "llm",
	Short: "Configure LLM providers and models",
	Long: `Configure LLM providers, models, and settings in formation.yaml.

This command provides an interactive wizard for:
  - Adding/updating API keys for LLM providers
  - Configuring models for different capabilities (text, vision, audio, etc.)
  - Global LLM settings (temperature, tokens, caching)

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard
  muxi config llm`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scaffold.ConfigureLLM(); err != nil {
			return fmt.Errorf("failed to configure LLM: %w", err)
		}
		return nil
	},
}

var configMemoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Configure memory settings",
	Long: `Configure memory settings in formation.yaml.

This command provides an interactive wizard for:
  - Working memory (in-memory vector storage, local or remote FAISSx)
  - Buffer memory (conversation context)
  - Persistent memory (PostgreSQL or SQLite database)

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard
  muxi config memory`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scaffold.ConfigureMemory(); err != nil {
			return fmt.Errorf("failed to configure memory: %w", err)
		}
		return nil
	},
}

var configOverlordCmd = &cobra.Command{
	Use:   "overlord",
	Short: "Configure Overlord behavior and settings",
	Long: `Configure Overlord settings in formation.yaml.

This command provides an interactive wizard for:
  - Persona (identity and communication style)
  - Response options (format, streaming, progress)
  - Workflow behavior (routing, decomposition, timeouts)
  - Clarification settings (question style, limits)

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard
  muxi config overlord`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scaffold.ConfigureOverlord(); err != nil {
			return fmt.Errorf("failed to configure overlord: %w", err)
		}
		return nil
	},
}

var configLoggingCmd = &cobra.Command{
	Use:   "logging",
	Short: "Configure logging streams",
	Long: `Configure logging streams in formation.yaml.

This command provides an interactive wizard for:
  - Adding log streams (stdout, file, http, kafka)
  - Viewing and editing existing streams
  - Removing streams

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard
  muxi config logging`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scaffold.ConfigureLogging(); err != nil {
			return fmt.Errorf("failed to configure logging: %w", err)
		}
		return nil
	},
}

var configSecurityCmd = &cobra.Command{
	Use:   "security",
	Short: "Configure user credential handling",
	Long: `Configure user credential handling in formation.yaml.

This command provides an interactive wizard for:
  - Redirect mode (production) - redirect to external credential management
  - Dynamic mode (development) - collect credentials inline with encryption

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard
  muxi config security`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scaffold.ConfigureSecurity(); err != nil {
			return fmt.Errorf("failed to configure security: %w", err)
		}
		return nil
	},
}

var configA2ACmd = &cobra.Command{
	Use:   "a2a",
	Short: "Configure A2A (Agent-to-Agent) communication",
	Long: `Configure Agent-to-Agent communication in formation.yaml.

This command modifies the a2a section in formation.yaml to enable
inbound or outbound A2A communication.

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard (asks for direction)
  muxi config a2a
  
  # Configure inbound A2A (skip direction question)
  muxi config a2a --inbound
  
  # Configure outbound A2A (skip direction question)
  muxi config a2a --outbound`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		inbound, _ := cmd.Flags().GetBool("inbound")
		outbound, _ := cmd.Flags().GetBool("outbound")
		noWizard, _ := cmd.Flags().GetBool("no-wizard")
		
		if err := scaffold.ConfigureA2A(inbound, outbound, noWizard); err != nil {
			return fmt.Errorf("failed to configure A2A: %w", err)
		}

		return nil
	},
}

func init() {
	// Add flags to config a2a
	configA2ACmd.Flags().Bool("inbound", false, "Configure inbound A2A (skip direction question)")
	configA2ACmd.Flags().Bool("outbound", false, "Configure outbound A2A (skip direction question)")
	configA2ACmd.Flags().Bool("no-wizard", false, "Skip interactive prompts")

	// Add subcommands to config
	configCmd.AddCommand(configA2ACmd)
	configCmd.AddCommand(configAsyncCmd)
	configCmd.AddCommand(configLLMCmd)
	configCmd.AddCommand(configLoggingCmd)
	configCmd.AddCommand(configMemoryCmd)
	configCmd.AddCommand(configOverlordCmd)
	configCmd.AddCommand(configSecurityCmd)

	// Add config to root
	rootCmd.AddCommand(configCmd)
}
