package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var knowledgeCmd = &cobra.Command{
	Use:     "knowledge",
	Short:   "Manage agent knowledge",
	GroupID: "formation",
	Long: `Manage agent knowledge bases.

Requires admin API key (from secrets.enc or MUXI_ADMIN_KEY).`,
}

var knowledgeRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Force-rebuild per-agent knowledge trees",
	Long: `Force-rebuild persistent per-agent knowledge trees.

Rebuilds every registered agent-tree source. Limit the scope with
--agent (one agent) and/or --source (one source slug).`,
	RunE: runKnowledgeRebuild,
}

func init() {
	rootCmd.AddCommand(knowledgeCmd)
	knowledgeCmd.AddCommand(knowledgeRebuildCmd)

	formation.AddFormationFlag(knowledgeRebuildCmd)
	formation.AddProfileFlag(knowledgeRebuildCmd)
	knowledgeRebuildCmd.Flags().String("agent", "", "Rebuild trees for one agent only")
	knowledgeRebuildCmd.Flags().String("source", "", "Rebuild one source slug only")
}

func runKnowledgeRebuild(cmd *cobra.Command, args []string) error {
	agentID, _ := cmd.Flags().GetString("agent")
	sourceID, _ := cmd.Flags().GetString("source")

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	resp, err := client.RebuildKnowledge(agentID, sourceID)
	if err != nil {
		return fmt.Errorf("failed to rebuild knowledge trees: %w", err)
	}

	fmt.Println()
	if len(resp.Agents) == 0 {
		ui.Dimmed("  No agents with knowledge trees found")
		fmt.Println()
		return nil
	}

	for agent, report := range resp.Agents {
		fmt.Printf("  %s\n", ui.BoldText(agent))
		for _, rebuilt := range report.Rebuilt {
			fmt.Printf("    ✓ %s (%d nodes)\n", rebuilt.SourceID, rebuilt.NodeCount)
		}
		for _, failed := range report.Failed {
			fmt.Printf("    ✗ %s (failed)\n", failed)
		}
		for _, skipped := range report.Skipped {
			fmt.Printf("    - %s %s\n", skipped, ui.DimmedText("(skipped)"))
		}
		if len(report.Rebuilt) == 0 && len(report.Failed) == 0 && len(report.Skipped) == 0 {
			ui.Dimmed("    no registered tree sources")
		}
		fmt.Println()
	}

	return nil
}
