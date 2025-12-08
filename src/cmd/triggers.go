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

func init() {
	rootCmd.AddCommand(triggersCmd)
	formation.AddFormationFlag(triggersCmd)
	formation.AddProfileFlag(triggersCmd)
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
