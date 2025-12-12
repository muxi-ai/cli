package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/scaffold"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Configure formation settings",
	GroupID: "formation",
	Long: `Configure settings in formation.yaml.

This command provides interactive wizards for configuring various
formation-level settings like A2A communication, LLM providers,
logging, and more.

Use --remote to fetch configuration from a running Formation via the API (read-only).

Must be run inside a formation directory.

Examples:
  # Interactive configuration (local)
  muxi config llm

  # Fetch full config from remote Formation
  muxi config --remote

  # Fetch specific config section from remote
  muxi config llm --remote
  muxi config memory --remote
  muxi config overlord --remote`,
	RunE: runConfig,
}

func runConfig(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetBool("remote")

	if remote {
		return runRemoteConfig(cmd, args)
	}

	// Default: show help since there's no default action for config without subcommand
	return cmd.Help()
}

func runRemoteConfig(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	configResp, err := client.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	return printConfigOutput(configResp, output)
}

func printConfigOutput(data interface{}, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	default:
		// Convert via JSON to preserve snake_case field names from json tags
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		var mapData map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &mapData); err != nil {
			return err
		}

		// Rename schema_version to schema for cleaner output
		if v, ok := mapData["schema_version"]; ok {
			mapData["schema"] = v
			delete(mapData, "schema_version")
		}

		// Pretty print as yaml with syntax highlighting and indentation
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(mapData); err != nil {
			return err
		}
		fmt.Println()
		fmt.Println(ui.IndentString(ui.RenderYAML(buf.String()), 2))
		return nil
	}
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

Use --remote to fetch LLM settings from a running Formation (read-only).

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard (local)
  muxi config llm

  # Fetch LLM settings from remote Formation
  muxi config llm --remote`,
	Args: cobra.NoArgs,
	RunE: runConfigLLM,
}

func runConfigLLM(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetBool("remote")

	if remote {
		client, err := formation.ClientFromFlags(cmd)
		if err != nil {
			return err
		}

		formation.PrintBadgeFromFlags(cmd)

		llmSettings, err := client.GetLLMSettings()
		if err != nil {
			return fmt.Errorf("failed to get LLM settings: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		return printConfigOutput(llmSettings, output)
	}

	if err := scaffold.ConfigureLLM(); err != nil {
		return fmt.Errorf("failed to configure LLM: %w", err)
	}
	return nil
}

var configMemoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Configure memory settings",
	Long: `Configure memory settings in formation.yaml.

This command provides an interactive wizard for:
  - Working memory (in-memory vector storage, local or remote FAISSx)
  - Buffer memory (conversation context)
  - Persistent memory (PostgreSQL or SQLite database)

Use --remote to fetch memory settings from a running Formation (read-only).

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard (local)
  muxi config memory

  # Fetch memory settings from remote Formation
  muxi config memory --remote`,
	Args: cobra.NoArgs,
	RunE: runConfigMemory,
}

func runConfigMemory(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetBool("remote")

	if remote {
		client, err := formation.ClientFromFlags(cmd)
		if err != nil {
			return err
		}

		formation.PrintBadgeFromFlags(cmd)

		memoryConfig, err := client.GetMemoryConfig()
		if err != nil {
			return fmt.Errorf("failed to get memory config: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		return printConfigOutput(memoryConfig, output)
	}

	if err := scaffold.ConfigureMemory(); err != nil {
		return fmt.Errorf("failed to configure memory: %w", err)
	}
	return nil
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

Use --remote to fetch overlord settings from a running Formation (read-only).

Must be run inside a formation directory.

Examples:
  # Configure with interactive wizard (local)
  muxi config overlord

  # Fetch overlord settings from remote Formation
  muxi config overlord --remote`,
	Args: cobra.NoArgs,
	RunE: runConfigOverlord,
}

func runConfigOverlord(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetBool("remote")

	if remote {
		client, err := formation.ClientFromFlags(cmd)
		if err != nil {
			return err
		}

		formation.PrintBadgeFromFlags(cmd)

		overlordConfig, err := client.GetOverlordConfig()
		if err != nil {
			return fmt.Errorf("failed to get overlord config: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		return printConfigOutput(overlordConfig, output)
	}

	if err := scaffold.ConfigureOverlord(); err != nil {
		return fmt.Errorf("failed to configure overlord: %w", err)
	}
	return nil
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
	// Add --remote and --output flags to config and subcommands
	configCmd.Flags().Bool("remote", false, "Fetch config from remote Formation API")
	configCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	formation.AddCommonFlags(configCmd)

	configLLMCmd.Flags().Bool("remote", false, "Fetch LLM settings from remote Formation API")
	configLLMCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	formation.AddCommonFlags(configLLMCmd)

	configMemoryCmd.Flags().Bool("remote", false, "Fetch memory config from remote Formation API")
	configMemoryCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	formation.AddCommonFlags(configMemoryCmd)

	configOverlordCmd.Flags().Bool("remote", false, "Fetch overlord config from remote Formation API")
	configOverlordCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	formation.AddCommonFlags(configOverlordCmd)

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
