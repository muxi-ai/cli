package cmd

import (
	"fmt"
	"text/tabwriter"
	"os"

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
	Use:   "show <name>",
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

	resp, err := client.GetTriggers()
	if err != nil {
		return fmt.Errorf("failed to get triggers: %w", err)
	}

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No triggers defined in this formation")
		fmt.Println()
		return nil
	}

	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "    %s\t%s\n", ui.BoldText("NAME"), ui.BoldText("DESCRIPTION"))

	for _, t := range resp.Triggers {
		desc := t.Description
		if desc == "" {
			desc = ui.DimmedText("-")
		}
		fmt.Fprintf(w, "    %s\t%s\n", t.Name, desc)
	}
	w.Flush()
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
		return fmt.Errorf("failed to get trigger: %w", err)
	}

	fmt.Println()
	fmt.Printf("  Trigger: %s\n", ui.BoldText(trigger.Name))
	if trigger.Description != "" {
		fmt.Printf("  Description: %s\n", trigger.Description)
	}
	fmt.Println()

	if trigger.Template != "" {
		fmt.Println("  Template:")
		fmt.Println("  ─────────")
		// Indent each line of the template
		for _, line := range splitLines(trigger.Template) {
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
