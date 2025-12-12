package cmd

import (
	"fmt"

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
  muxi sops show <id>    Show SOP details

Requires client API key (from secrets.enc or MUXI_CLIENT_KEY).`,
	RunE: runSopsList,
}

var sopsShowCmd = &cobra.Command{
	Use:   "show <id>",
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
	sopsShowCmd.Flags().Bool("raw", false, "Show raw content without markdown rendering")
}

func runSopsList(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	resp, err := client.GetSOPs()
	if err != nil {
		return fmt.Errorf("failed to get SOPs: %w", err)
	}

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No SOPs found")
		fmt.Println()
		return nil
	}

	fmt.Println()
	if resp.Count == 1 {
		fmt.Println("  1 SOP found:")
	} else {
		fmt.Printf("  %d SOPs found:\n", resp.Count)
	}
	fmt.Println()

	for _, s := range resp.SOPs {
		sopType := s.Type
		if sopType == "" {
			sopType = "template"
		}
		fmt.Printf("    • %s (%s)\n", s.Name, sopType)
		desc := s.Description
		if desc == "" {
			desc = "—"
		} else if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("      %s\n", ui.DimmedText(desc))
		fmt.Println()
	}

	ui.Dimmed("  Use 'muxi sops show <name>' for details")
	fmt.Println()

	return nil
}

func runSopsShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	sop, err := client.GetSOP(name)
	if err != nil {
		fmt.Println()
		fmt.Printf("  SOP '%s' not found\n", name)
		fmt.Println()
		return nil
	}

	sopType := sop.Type
	if sopType == "" {
		sopType = "template"
	}

	// Front matter (unrendered)
	fmt.Println()
	fmt.Printf("  %s (%s)\n", ui.BoldText(sop.Name), sopType)
	if sop.Description != "" {
		fmt.Printf("  %s\n", sop.Description)
	}
	if sop.Steps > 0 {
		fmt.Printf("  Steps:   %d\n", sop.Steps)
	}
	if len(sop.Agents) > 0 {
		fmt.Printf("  Agents:  %v\n", sop.Agents)
	}

	// Content
	if sop.Content != "" {
		raw, _ := cmd.Flags().GetBool("raw")
		fmt.Println()
		fmt.Println("  Content:")
		fmt.Println("  " + ui.DimmedText("────────────────────────────────────────"))
		fmt.Println()
		if raw {
			fmt.Println(sop.Content)
		} else {
			fmt.Println(ui.RenderMarkdown(sop.Content))
		}
		fmt.Println("  " + ui.DimmedText("────────────────────────────────────────"))
	}
	fmt.Println()

	return nil
}
