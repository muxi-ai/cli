package cmd

import (
	"fmt"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var schedulerCmd = &cobra.Command{
	Use:     "scheduler",
	Short:   "View scheduler configuration and jobs",
	GroupID: "formation",
	Long: `View scheduler configuration and scheduled jobs.

The scheduler manages automated job execution on a schedule.

Requires admin API key.`,
}

var schedulerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show scheduler status and configuration",
	Long: `Show scheduler status and configuration.

Displays whether scheduler is enabled, timezone, check interval, and limits.`,
	RunE: runSchedulerStatus,
}

var schedulerListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List scheduled jobs",
	Long: `List all scheduled jobs.

Shows job ID, type, schedule, next run time, and status.`,
	RunE: runSchedulerList,
}

func init() {
	rootCmd.AddCommand(schedulerCmd)
	schedulerCmd.AddCommand(schedulerStatusCmd)
	schedulerCmd.AddCommand(schedulerListCmd)

	formation.AddFormationFlag(schedulerStatusCmd)
	formation.AddProfileFlag(schedulerStatusCmd)

	formation.AddFormationFlag(schedulerListCmd)
	formation.AddProfileFlag(schedulerListCmd)
}

func runSchedulerStatus(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	config, err := client.GetSchedulerConfig()
	if err != nil {
		return fmt.Errorf("failed to get scheduler config: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	ui.Bold("Scheduler Configuration")
	fmt.Println()

	enabledStr := ui.RedText("disabled")
	if config.Enabled {
		enabledStr = ui.GreenText("enabled")
	}
	fmt.Printf("  Status:              %s\n", enabledStr)
	fmt.Printf("  Timezone:            %s\n", config.Timezone)
	fmt.Printf("  Check Interval:      %d minutes\n", config.CheckIntervalMinutes)
	fmt.Printf("  Max Concurrent Jobs: %d\n", config.MaxConcurrentJobs)
	fmt.Printf("  Max Failures:        %d\n", config.MaxFailuresBeforePause)
	fmt.Println()

	return nil
}

func runSchedulerList(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	resp, err := client.GetSchedulerJobs()
	if err != nil {
		return fmt.Errorf("failed to list scheduler jobs: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No scheduled jobs")
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Printf("  %-20s %-12s %-15s %-20s %s\n",
		"ID", "TYPE", "SCHEDULE", "NEXT RUN", "STATUS")
	fmt.Printf("  %s\n", strings.Repeat("─", 80))

	for _, job := range resp.Jobs {
		schedule := job.Schedule
		if schedule == "" {
			schedule = "-"
		}

		nextRun := "-"
		if !job.NextRun.IsZero() {
			nextRun = job.NextRun.Format("Jan 2 3:04pm")
		}

		status := ui.RedText("● disabled")
		if job.Enabled {
			status = ui.GreenText("● enabled")
		}

		fmt.Printf("  %-20s %-12s %-15s %-20s %s\n",
			truncate(job.ID, 20),
			job.Type,
			truncate(schedule, 15),
			nextRun,
			status)
	}
	fmt.Println()

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
