package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var triggersCmd = &cobra.Command{
	Use:     "triggers",
	Short:   "Manage and run formation triggers",
	GroupID: "formation",
	Long: `Manage and run triggers defined in the formation.

Triggers are entry points that can be invoked to start workflows.
Use 'muxi triggers run <name>' to invoke a trigger.

Requires client API key (from secrets.enc or MUXI_CLIENT_KEY).`,
}

var triggersListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all triggers",
	Long: `List all triggers defined in the formation.

Displays trigger name and description.`,
	RunE: runTriggers,
}

var triggersShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show trigger details",
	Long: `Display a trigger's template content with markdown rendering.

Use --raw to output the raw template without rendering, suitable for
piping to files or other tools.`,
	Args: RequireArgs(1),
	RunE: runTriggersShow,
}

var triggersRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run a trigger",
	Long: `Run a trigger on a formation with optional data.

Triggers are predefined entry points that can process structured data
and invoke agents or workflows.

Examples:
  muxi triggers run github-issue --data '{"issue": {"number": 123}}'
  muxi triggers run webhook --file event.json
  muxi triggers run daily-report --async`,
	Args: RequireArgs(1),
	RunE: runTriggersRun,
}

func init() {
	rootCmd.AddCommand(triggersCmd)
	triggersCmd.AddCommand(triggersListCmd)
	triggersCmd.AddCommand(triggersShowCmd)
	triggersCmd.AddCommand(triggersRunCmd)

	formation.AddFormationFlag(triggersListCmd)
	formation.AddProfileFlag(triggersListCmd)

	formation.AddFormationFlag(triggersShowCmd)
	formation.AddProfileFlag(triggersShowCmd)
	triggersShowCmd.Flags().Bool("raw", false, "Show raw content without markdown rendering")

	formation.AddFormationFlag(triggersRunCmd)
	formation.AddProfileFlag(triggersRunCmd)
	triggersRunCmd.Flags().String("data", "", "JSON data to send with the trigger")
	triggersRunCmd.Flags().String("file", "", "JSON file to read trigger data from")
	triggersRunCmd.Flags().Bool("async", false, "Run trigger asynchronously (returns job ID)")
}

func runTriggers(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	resp, err := client.GetTriggers()
	if err != nil {
		return fmt.Errorf("failed to get triggers: %w", err)
	}

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No triggers found")
		fmt.Println()
		return nil
	}

	fmt.Println()
	if resp.Count == 1 {
		fmt.Println("  1 trigger found:")
	} else {
		fmt.Printf("  %d triggers found:\n", resp.Count)
	}
	fmt.Println()
	for _, name := range resp.Triggers {
		fmt.Printf("    • %s\n", name)
	}
	fmt.Println()
	ui.Dimmed("  Use 'muxi triggers show <name>' for details")
	fmt.Println()

	return nil
}

func runTriggersShow(cmd *cobra.Command, args []string) error {
	triggerName := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	trigger, err := client.GetTrigger(triggerName)
	if err != nil {
		fmt.Println()
		fmt.Printf("  Trigger '%s' not found\n", triggerName)
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Printf("  Trigger: %s\n", ui.BoldText(trigger.Name))
	fmt.Println()

	if len(trigger.DataFields) > 0 {
		fmt.Println("  Data Fields:")
		for _, field := range trigger.DataFields {
			fmt.Printf("    • ${{ data.%s }}\n", field)
		}
		fmt.Println()
	}

	if trigger.Content != "" {
		raw, _ := cmd.Flags().GetBool("raw")
		fmt.Println("  Content:")
		fmt.Println("  " + ui.DimmedText("────────────────────────────────────────"))
		if raw {
			fmt.Println(trigger.Content)
		} else {
			fmt.Println(ui.RenderMarkdown(trigger.Content))
		}
		fmt.Println("  " + ui.DimmedText("────────────────────────────────────────"))
	}
	fmt.Println()

	return nil
}

func runTriggersRun(cmd *cobra.Command, args []string) error {
	triggerName := args[0]
	dataStr, _ := cmd.Flags().GetString("data")
	filePath, _ := cmd.Flags().GetString("file")
	async, _ := cmd.Flags().GetBool("async")

	// Validate that we have either --data or --file (or neither), but not both
	if dataStr != "" && filePath != "" {
		ui.ErrorBlock(
			"Invalid flags",
			"Cannot use both --data and --file flags.",
			"Use one or the other.",
		)
		os.Exit(1)
	}

	// Read data from file if specified
	var data json.RawMessage
	if filePath != "" {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		// Validate it's valid JSON
		if !json.Valid(fileData) {
			ui.ErrorBlock(
				"Invalid JSON",
				fmt.Sprintf("File '%s' does not contain valid JSON.", filePath),
				"",
			)
			os.Exit(1)
		}
		data = json.RawMessage(fileData)
	} else if dataStr != "" {
		// Validate JSON string
		if !json.Valid([]byte(dataStr)) {
			ui.ErrorBlock(
				"Invalid JSON",
				"The --data value is not valid JSON.",
				"Example: --data '{\"key\": \"value\"}'",
			)
			os.Exit(1)
		}
		data = json.RawMessage(dataStr)
	} else {
		// No data provided, use empty object
		data = json.RawMessage("{}")
	}

	// Create client
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	// Run trigger
	spinner := ui.NewSpinner(fmt.Sprintf("Running trigger '%s'...", triggerName))
	spinner.Start()

	resp, err := client.TriggerTrigger(triggerName, data, async)
	if err != nil {
		spinner.StopWithError("Trigger failed")
		fmt.Println()
		// Extract just the message part (after "CODE: prefix: ")
		errMsg := err.Error()
		if idx := strings.Index(errMsg, ": "); idx != -1 {
			errMsg = errMsg[idx+2:]
			// Check for second colon (e.g., "INVALID_REQUEST: Template rendering failed: actual message")
			if idx2 := strings.Index(errMsg, ": "); idx2 != -1 {
				errMsg = errMsg[idx2+2:]
			}
		}
		fmt.Printf("  %s\n", errMsg)
		fmt.Println()
		os.Exit(1)
	}

	spinner.StopWithSuccess("Trigger fired")

	// Display result
	fmt.Println()
	if async {
		fmt.Printf("  Status:     %s\n", ui.CyanText("async"))
		fmt.Printf("  Job ID:     %s\n", resp.JobID)
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  Track progress: muxi jobs list"))
	} else {
		fmt.Printf("  Status:     %s\n", ui.GreenText(resp.Status))
		fmt.Printf("  Request ID: %s\n", resp.RequestID)
		if resp.Response != "" {
			fmt.Println()
			fmt.Printf("  Response:\n")
			fmt.Printf("  %s\n", resp.Response)
		}
	}
	fmt.Println()

	return nil
}
