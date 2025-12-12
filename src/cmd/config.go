package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/scaffold"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
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

		// Clean up the config output
		cleanConfigOutput(mapData)

		// Build ordered output (schema, version, id, description first)
		orderedYAML := buildOrderedConfigYAML(mapData)

		fmt.Println()
		fmt.Println(ui.IndentString(ui.RenderYAML(orderedYAML), 2))
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

// cleanConfigOutput removes "resource" keys and renames fields for cleaner output
func cleanConfigOutput(data map[string]interface{}) {
	// Rename schema_version to schema
	if v, ok := data["schema_version"]; ok {
		data["schema"] = v
		delete(data, "schema_version")
	}

	// Rename formation_id to id
	if v, ok := data["formation_id"]; ok {
		data["id"] = v
		delete(data, "formation_id")
	}

	// Replace "resource" keys with CLI commands
	replaceResourceWithCLI(data)
}

// resourceToCLI maps API resource paths to CLI commands
var resourceToCLI = map[string]string{
	"/v1/agents":       "muxi agents ls",
	"/v1/mcp/servers":  "muxi mcp ls",
	"/v1/secrets":      "muxi secrets ls --remote",
	"/v1/llm/settings": "muxi config llm --remote",
	"/v1/memory":       "muxi config memory --remote",
	"/v1/overlord":     "muxi config overlord --remote",
	"/v1/a2a":          "muxi config a2a --remote",
	"/v1/async":        "muxi config async --remote",
	"/v1/scheduler":    "muxi scheduler status",
	"/v1/logging":      "muxi config logging --remote",
}

// replaceResourceWithCLI recursively replaces "resource" keys with "cli" commands
func replaceResourceWithCLI(data map[string]interface{}) {
	if resource, ok := data["resource"].(string); ok {
		if cli, found := resourceToCLI[resource]; found {
			data["cli"] = cli
		}
		delete(data, "resource")
	}

	for _, v := range data {
		if nested, ok := v.(map[string]interface{}); ok {
			replaceResourceWithCLI(nested)
		}
	}
}

// buildOrderedConfigYAML creates YAML with specific field order
func buildOrderedConfigYAML(data map[string]interface{}) string {
	var buf bytes.Buffer

	// Priority fields in order
	priorityKeys := []string{"schema", "version", "id", "description"}

	// Write priority fields first
	for _, key := range priorityKeys {
		if v, ok := data[key]; ok {
			writeYAMLField(&buf, key, v, 0)
			delete(data, key)
		}
	}

	// Write remaining fields (sorted for consistency)
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		writeYAMLField(&buf, key, data[key], 0)
	}

	return buf.String()
}

// writeYAMLField writes a single YAML field with proper indentation
func writeYAMLField(buf *bytes.Buffer, key string, value interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)

	switch v := value.(type) {
	case map[string]interface{}:
		buf.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
		// Sort nested keys
		nestedKeys := make([]string, 0, len(v))
		for k := range v {
			nestedKeys = append(nestedKeys, k)
		}
		sort.Strings(nestedKeys)
		for _, nk := range nestedKeys {
			writeYAMLField(buf, nk, v[nk], indent+1)
		}
	case []interface{}:
		buf.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
		for _, item := range v {
			buf.WriteString(fmt.Sprintf("%s  - %v\n", indentStr, item))
		}
	case string:
		buf.WriteString(fmt.Sprintf("%s%s: %s\n", indentStr, key, v))
	case float64:
		if v == float64(int(v)) {
			buf.WriteString(fmt.Sprintf("%s%s: %d\n", indentStr, key, int(v)))
		} else {
			buf.WriteString(fmt.Sprintf("%s%s: %v\n", indentStr, key, v))
		}
	case bool:
		buf.WriteString(fmt.Sprintf("%s%s: %t\n", indentStr, key, v))
	default:
		buf.WriteString(fmt.Sprintf("%s%s: %v\n", indentStr, key, v))
	}
}
