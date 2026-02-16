package cmd

import (
	"fmt"
	"strings"
	"time"

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

var schedulerAddCmd = &cobra.Command{
	Use:   "add <once|recurring> <schedule> <message>",
	Short: "Add a scheduled job",
	Long: `Add a new scheduled job (one-time or recurring).

Types:
  once       One-time job that runs at a specific datetime
  recurring  Recurring job that runs on a cron schedule

Schedule formats:
  once:      Date/time (e.g., "2025-12-31", "2025-12-31 14:00", "2025-12-31T14:00:00Z")
  recurring: Cron expression (e.g., "0 9 * * 1" for Mondays at 9am)

Examples:
  muxi scheduler add once "2025-12-31" "Send new year greeting"
  muxi scheduler add once "2025-12-31 12:00" "Noon reminder"
  muxi scheduler add recurring "0 9 * * *" "Daily morning report"
  muxi scheduler add recurring "0 9 * * 1" "Weekly Monday report" -u alice`,
	Args: RequireArgs(3),
	RunE: runSchedulerAdd,
}

var schedulerShowCmd = &cobra.Command{
	Use:   "show <job-id>",
	Short: "Show scheduled job details",
	Long:  `Display detailed information about a specific scheduled job.`,
	Args:  RequireArgs(1),
	RunE:  runSchedulerShow,
}

var schedulerRemoveCmd = &cobra.Command{
	Use:     "remove <job-id>",
	Aliases: []string{"rm"},
	Short:   "Remove a scheduled job",
	Long:    `Remove a scheduled job by ID.`,
	Args:    RequireArgs(1),
	RunE:    runSchedulerRemove,
}

func init() {
	rootCmd.AddCommand(schedulerCmd)
	schedulerCmd.AddCommand(schedulerStatusCmd)
	schedulerCmd.AddCommand(schedulerListCmd)
	schedulerCmd.AddCommand(schedulerAddCmd)
	schedulerCmd.AddCommand(schedulerShowCmd)
	schedulerCmd.AddCommand(schedulerRemoveCmd)

	formation.AddFormationFlag(schedulerStatusCmd)
	formation.AddProfileFlag(schedulerStatusCmd)
	formation.AddUserFlag(schedulerStatusCmd)

	formation.AddFormationFlag(schedulerListCmd)
	formation.AddProfileFlag(schedulerListCmd)
	formation.AddUserFlag(schedulerListCmd)

	formation.AddFormationFlag(schedulerAddCmd)
	formation.AddProfileFlag(schedulerAddCmd)
	formation.AddUserFlag(schedulerAddCmd)

	formation.AddFormationFlag(schedulerShowCmd)
	formation.AddProfileFlag(schedulerShowCmd)

	formation.AddFormationFlag(schedulerRemoveCmd)
	formation.AddProfileFlag(schedulerRemoveCmd)
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

	enabledStr := ui.RedText("✗ disabled")
	if config.Enabled {
		enabledStr = ui.GreenText("✓ enabled")
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

	userFlag, _ := cmd.Flags().GetString("user")
	userID := formation.ResolveUserID(userFlag)

	resp, err := client.GetSchedulerJobs(userID)
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

func runSchedulerAdd(cmd *cobra.Command, args []string) error {
	jobType := strings.ToLower(args[0])
	schedule := args[1]
	message := args[2]

	// Validate job type
	if jobType != "once" && jobType != "recurring" {
		ui.ErrorBlock(
			"Invalid job type",
			fmt.Sprintf("'%s' is not a valid job type.", jobType),
			"Use 'once' for one-time jobs or 'recurring' for cron-based jobs.",
		)
		return nil
	}

	// Map "once" to API's "one_time"
	apiType := jobType
	if jobType == "once" {
		apiType = "one_time"
		// Parse flexible date format and convert to ISO
		parsedTime, err := parseFlexibleDate(schedule)
		if err != nil {
			ui.ErrorBlock(
				"Invalid date/time",
				fmt.Sprintf("Could not parse '%s' as a date/time.", schedule),
				"Examples: \"2025-12-31\", \"2025-12-31 14:00\", \"2025-12-31T14:00:00Z\"",
			)
			return nil
		}
		// Check if in the future
		if parsedTime.Before(time.Now()) {
			ui.ErrorBlock(
				"Invalid schedule",
				"Schedule datetime must be in the future.",
				"",
			)
			return nil
		}
		schedule = parsedTime.UTC().Format(time.RFC3339)
	}

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	userFlag, _ := cmd.Flags().GetString("user")
	userID := formation.ResolveUserID(userFlag)

	spinner := ui.NewSpinner("Creating scheduled job...")
	spinner.Start()

	resp, err := client.CreateSchedulerJob(apiType, schedule, message, userID)
	spinner.Stop()
	formation.PrintBadgeFromFlags(cmd)

	if err != nil {
		fmt.Printf("\n  %s Failed to create job\n", ui.RedText("✗"))
		// Extract error message
		errMsg := err.Error()
		if strings.Contains(errMsg, ": ") {
			parts := strings.SplitN(errMsg, ": ", 2)
			if len(parts) > 1 {
				errMsg = parts[1]
			}
		}
		fmt.Printf("    %s\n\n", ui.DimmedText(errMsg))
		return nil
	}

	fmt.Printf("\n  %s Scheduled job created\n", ui.GreenText("✓"))

	fmt.Println()
	fmt.Printf("  Job ID:   %s\n", ui.BoldText(resp.ID))
	fmt.Printf("  Type:     %s\n", resp.Type)
	fmt.Printf("  Schedule: %s\n", resp.Schedule)
	if !resp.NextRun.IsZero() {
		fmt.Printf("  Next Run: %s\n", resp.NextRun.Format("Jan 2, 2006 3:04pm MST"))
	}
	fmt.Println()

	return nil
}

// parseFlexibleDate parses various date formats into time.Time
func parseFlexibleDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,          // 2025-12-31T14:00:00Z
		"2006-01-02T15:04:05", // 2025-12-31T14:00:00
		"2006-01-02 15:04:05", // 2025-12-31 14:00:00
		"2006-01-02 15:04",    // 2025-12-31 14:00
		"2006-01-02",          // 2025-12-31 (midnight)
		"01/02/2006 15:04",    // 12/31/2025 14:00
		"01/02/2006",          // 12/31/2025
		"Jan 2, 2006 3:04pm",  // Dec 31, 2025 2:00pm
		"Jan 2, 2006",         // Dec 31, 2025
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// Try with local timezone
	loc := time.Local
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", s)
}

func runSchedulerShow(cmd *cobra.Command, args []string) error {
	jobID := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	job, err := client.GetSchedulerJob(jobID)
	if err != nil {
		formation.PrintBadgeFromFlags(cmd)
		fmt.Printf("\n  %s Job not found: %s\n\n", ui.RedText("✗"), jobID)
		return nil
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	ui.Bold(fmt.Sprintf("Job: %s", job.ID))
	fmt.Println()

	status := ui.RedText("● disabled")
	if job.Enabled {
		status = ui.GreenText("● enabled")
	}

	fmt.Printf("  Type:     %s\n", job.Type)
	fmt.Printf("  Status:   %s\n", status)
	if job.Schedule != "" {
		fmt.Printf("  Schedule: %s\n", job.Schedule)
	}
	fmt.Printf("  Message:  %s\n", job.Message)
	if job.UserID != "" {
		fmt.Printf("  User:     %s\n", job.UserID)
	}
	if !job.NextRun.IsZero() {
		fmt.Printf("  Next Run: %s\n", job.NextRun.Format("Jan 2, 2006 3:04pm MST"))
	}
	if !job.LastRun.IsZero() {
		fmt.Printf("  Last Run: %s\n", job.LastRun.Format("Jan 2, 2006 3:04pm MST"))
	}
	if job.FailureCount > 0 {
		fmt.Printf("  Failures: %s\n", ui.RedText(fmt.Sprintf("%d", job.FailureCount)))
	}
	fmt.Println()

	return nil
}

func runSchedulerRemove(cmd *cobra.Command, args []string) error {
	jobID := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	err = client.DeleteSchedulerJob(jobID)
	formation.PrintBadgeFromFlags(cmd)

	if err != nil {
		fmt.Printf("\n  %s Failed to remove job\n", ui.RedText("✗"))
		errMsg := err.Error()
		if strings.Contains(errMsg, ": ") {
			parts := strings.SplitN(errMsg, ": ", 2)
			if len(parts) > 1 {
				errMsg = parts[1]
			}
		}
		fmt.Printf("    %s\n\n", ui.DimmedText(errMsg))
		return nil
	}

	fmt.Printf("\n  %s Removed job: %s\n\n", ui.GreenText("✓"), jobID)
	return nil
}
