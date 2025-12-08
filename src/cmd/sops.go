package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var sopsCmd = &cobra.Command{
	Use:     "sops",
	Short:   "List and view SOPs",
	GroupID: "formation",
	Long: `List and view Standard Operating Procedures (SOPs) in the formation.

SOPs define workflows, templates, and guides for agents to follow.

Commands:
  muxi sops              List all SOPs
  muxi sops show <name>  Show SOP details

Requires client API key (from secrets.enc or MUXI_CLIENT_KEY).`,
	RunE: runSopsList,
}

var sopsShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show SOP details",
	Long:  `Show detailed information about a specific SOP including its content.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSopsShow,
}

func init() {
	rootCmd.AddCommand(sopsCmd)
	sopsCmd.AddCommand(sopsShowCmd)

	formation.AddFormationFlag(sopsCmd)
	formation.AddProfileFlag(sopsCmd)

	formation.AddFormationFlag(sopsShowCmd)
	formation.AddProfileFlag(sopsShowCmd)
}

func runSopsList(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	resp, err := client.GetSOPs()
	if err != nil {
		return fmt.Errorf("failed to get SOPs: %w", err)
	}

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No SOPs defined in this formation")
		fmt.Println()
		return nil
	}

	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "    %s\t%s\t%s\n", ui.BoldText("NAME"), ui.BoldText("TYPE"), ui.BoldText("DESCRIPTION"))

	for _, s := range resp.SOPs {
		sopType := s.Type
		if sopType == "" {
			sopType = "template"
		}
		desc := s.Description
		if desc == "" {
			desc = ui.DimmedText("-")
		}
		// Truncate long descriptions
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		fmt.Fprintf(w, "    %s\t%s\t%s\n", s.Name, sopType, desc)
	}
	w.Flush()
	fmt.Println()

	return nil
}

func runSopsShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	sop, err := client.GetSOP(name)
	if err != nil {
		return fmt.Errorf("failed to get SOP '%s': %w", name, err)
	}

	fmt.Println()
	fmt.Printf("  SOP: %s\n", ui.BoldText(sop.Name))
	fmt.Println()

	if sop.Description != "" {
		fmt.Printf("    Description:  %s\n", sop.Description)
	}
	if sop.Type != "" {
		fmt.Printf("    Type:         %s\n", sop.Type)
	}
	if sop.Steps > 0 {
		fmt.Printf("    Steps:        %d\n", sop.Steps)
	}
	if len(sop.Agents) > 0 {
		fmt.Printf("    Agents:       %v\n", sop.Agents)
	}
	fmt.Println()

	if sop.Content != "" {
		fmt.Println("    Content:")
		fmt.Println("    " + ui.DimmedText("─────────────────────────────────────────"))
		fmt.Println(ui.Indent(sop.Content, 2))
		fmt.Println("    " + ui.DimmedText("─────────────────────────────────────────"))
		fmt.Println()
	}

	return nil
}
