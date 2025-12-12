package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var triggersCmd = &cobra.Command{
	Use:     "triggers",
	Short:   "List formation triggers",
	GroupID: "formation",
	Long: `List all triggers defined in the formation.

Triggers are entry points that can be invoked to start workflows.
Use 'muxi trigger <name>' to invoke a trigger.

Requires client API key (from secrets.enc or MUXI_CLIENT_KEY).`,
	RunE: runTriggers,
}

var triggersShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show trigger details",
	Long: `Show detailed information about a specific trigger.

Displays the trigger template content and metadata.`,
	Args: cobra.ExactArgs(1),
	RunE: runTriggersShow,
}

func init() {
	rootCmd.AddCommand(triggersCmd)
	triggersCmd.AddCommand(triggersShowCmd)

	formation.AddFormationFlag(triggersCmd)
	formation.AddProfileFlag(triggersCmd)
	formation.AddFormationFlag(triggersShowCmd)
	formation.AddProfileFlag(triggersShowCmd)
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

	trigger, err := client.GetTrigger(triggerName)
	if err != nil {
		fmt.Println()
		fmt.Printf("  Trigger '%s' not found\n", triggerName)
		fmt.Println()
		return nil
	}

	formation.PrintBadgeFromFlags(cmd)

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
		fmt.Println("  Template:")
		fmt.Println("  ─────────")
		// Indent each line of the template
		for _, line := range splitLines(trigger.Content) {
			fmt.Printf("  %s\n", line)
		}
	}
	fmt.Println()

	return nil
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
