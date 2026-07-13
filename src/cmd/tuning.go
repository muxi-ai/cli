package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var tuningCmd = &cobra.Command{
	Use:     "tuning",
	Short:   "Review formation self-improvement state",
	GroupID: "formation",
	Long: `Review the formation's self-improvement (tuning) state.

The tuner accumulates learnings in MUXI.md. With auto_apply disabled it
writes suggested revisions to PENDING-MUXI.md instead, awaiting review:
show the suggestion with 'pending', then 'apply' or 'dismiss' it.

Requires admin API key (from secrets.enc or MUXI_ADMIN_KEY).`,
}

var tuningShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the live MUXI.md",
	Long:  `Display the formation's live MUXI.md content.`,
	RunE:  runTuningShow,
}

var tuningPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "Show the pending MUXI.md suggestion",
	Long:  `Display the pending MUXI.md suggestion awaiting review, if any.`,
	RunE:  runTuningPending,
}

var tuningApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the pending suggestion",
	Long:  `Promote the pending suggestion to the live MUXI.md and open its watch windows.`,
	RunE:  runTuningApply,
}

var tuningDismissCmd = &cobra.Command{
	Use:   "dismiss",
	Short: "Dismiss the pending suggestion",
	Long:  `Discard the pending suggestion. Dismissed learnings are never re-proposed.`,
	RunE:  runTuningDismiss,
}

func init() {
	rootCmd.AddCommand(tuningCmd)
	tuningCmd.AddCommand(tuningShowCmd)
	tuningCmd.AddCommand(tuningPendingCmd)
	tuningCmd.AddCommand(tuningApplyCmd)
	tuningCmd.AddCommand(tuningDismissCmd)

	for _, c := range []*cobra.Command{tuningShowCmd, tuningPendingCmd, tuningApplyCmd, tuningDismissCmd} {
		formation.AddFormationFlag(c)
		formation.AddProfileFlag(c)
	}
	tuningShowCmd.Flags().Bool("raw", false, "Show raw content without markdown rendering")
	tuningPendingCmd.Flags().Bool("raw", false, "Show raw content without markdown rendering")
}

func printTuningContent(cmd *cobra.Command, title string, content *string, path string, emptyHint string) {
	fmt.Println()
	if content == nil {
		ui.Dimmed("  " + emptyHint)
		fmt.Println()
		return
	}

	raw, _ := cmd.Flags().GetBool("raw")
	fmt.Printf("  %s (%s)\n", ui.BoldText(title), ui.DimmedText(path))
	fmt.Println()
	if raw {
		fmt.Println(*content)
	} else {
		fmt.Println(ui.RenderMarkdown(*content))
	}
	fmt.Println()
}

func runTuningShow(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	tuning, err := client.GetTuning()
	if err != nil {
		return fmt.Errorf("failed to get MUXI.md: %w", err)
	}

	printTuningContent(cmd, "MUXI.md", tuning.Content, tuning.Path, "No MUXI.md yet - the tuner has not accumulated any learnings")
	return nil
}

func runTuningPending(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	pending, err := client.GetTuningPending()
	if err != nil {
		return fmt.Errorf("failed to get pending suggestion: %w", err)
	}

	printTuningContent(cmd, "PENDING-MUXI.md", pending.Content, pending.Path, "No pending suggestion")
	if pending.Content != nil {
		ui.Dimmed("  Accept with: muxi tuning apply    Discard with: muxi tuning dismiss")
		fmt.Println()
	}
	return nil
}

func runTuningApply(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	result, err := client.ApplyTuningPending()
	if err != nil {
		return fmt.Errorf("failed to apply pending suggestion: %w", err)
	}

	fmt.Println()
	fmt.Printf("  ✓ Suggestion applied to %s\n", result.Path)
	if result.LearningsActivated > 0 {
		fmt.Printf("  %d learning(s) now in their watch window\n", result.LearningsActivated)
	}
	fmt.Println()
	return nil
}

func runTuningDismiss(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	result, err := client.DismissTuningPending()
	if err != nil {
		return fmt.Errorf("failed to dismiss pending suggestion: %w", err)
	}

	fmt.Println()
	fmt.Printf("  ✓ Suggestion dismissed (%d learning(s) will not be re-proposed)\n", result.LearningsDismissed)
	fmt.Println()
	return nil
}
