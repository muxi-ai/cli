package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var triggerCmd = &cobra.Command{
	Use:     "trigger <name>",
	Short:   "Fire a trigger",
	GroupID: "formation",
	Long: `Fire a trigger on a formation with optional data.

Triggers are predefined entry points that can process structured data
and invoke agents or workflows.

Examples:
  muxi trigger github-issue --data '{"issue": {"number": 123}}'
  muxi trigger webhook --file event.json
  muxi trigger daily-report --async`,
	Args: cobra.ExactArgs(1),
	RunE: runTrigger,
}

func init() {
	rootCmd.AddCommand(triggerCmd)

	formation.AddFormationFlag(triggerCmd)
	formation.AddProfileFlag(triggerCmd)
	triggerCmd.Flags().String("data", "", "JSON data to send with the trigger")
	triggerCmd.Flags().String("file", "", "JSON file to read trigger data from")
	triggerCmd.Flags().Bool("async", false, "Fire trigger asynchronously (returns job ID)")
}

func runTrigger(cmd *cobra.Command, args []string) error {
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

	// Fire trigger
	spinner := ui.NewSpinner(fmt.Sprintf("Firing trigger '%s'...", triggerName))
	spinner.Start()

	resp, err := client.TriggerTrigger(triggerName, data, async)
	if err != nil {
		spinner.StopWithError("Trigger failed")
		return err
	}

	spinner.StopWithSuccess("Trigger fired")

	// Display result
	fmt.Println()
	if async {
		fmt.Printf("  Status:     %s\n", ui.CyanText("async"))
		fmt.Printf("  Job ID:     %s\n", resp.JobID)
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  Track progress: muxi jobs -u <user>"))
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
