package cmd

import (
	"fmt"
	"time"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:     "jobs",
	Short:   "Manage async jobs",
	GroupID: "formation",
	Long: `List and manage async jobs for a user.

Jobs are created when triggers or chat requests are processed asynchronously.`,
}

var jobsListCmd = &cobra.Command{
	Use:   "list [user-id]",
	Aliases: []string{"ls"},
	Short: "List async jobs",
	Long: `List async jobs for a user.

User ID can be provided as argument or will use default user if set.`,
	Example: `  muxi jobs list
  muxi jobs list alice`,
	RunE: runJobsList,
}

var jobsCancelCmd = &cobra.Command{
	Use:   "cancel <job-id> [user-id]",
	Short: "Cancel an async job",
	Long: `Cancel an async job.

User ID can be provided as second argument or will use default user if set.`,
	Example: `  muxi jobs cancel job_123
  muxi jobs cancel job_123 alice`,
	Args: RequireArgs(1),
	RunE: runJobsCancel,
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.AddCommand(jobsListCmd)
	jobsCmd.AddCommand(jobsCancelCmd)

	// Formation and profile flags only (user from arg or default)
	formation.AddFormationFlag(jobsListCmd)
	formation.AddProfileFlag(jobsListCmd)
	formation.AddFormationFlag(jobsCancelCmd)
	formation.AddProfileFlag(jobsCancelCmd)
}

func runJobsList(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	// User from arg or default
	var userID string
	if len(args) > 0 {
		userID = args[0]
	} else {
		userID, err = formation.MustResolveUserID("")
		if err != nil {
			return fmt.Errorf("user ID required: provide as argument or set default with: muxi set default user")
		}
	}

	resp, err := client.GetJobs(userID)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed("  No jobs found")
		fmt.Println()
		return nil
	}

	// Header
	fmt.Println()
	fmt.Printf("  %-14s %-12s %-10s %s\n",
		"JOB ID", "STATUS", "PROGRESS", "CREATED")
	fmt.Printf("  %-14s %-12s %-10s %s\n",
		"──────────────", "────────────", "──────────", "────────────")

	for _, job := range resp.Jobs {
		// Status with color
		var statusDisplay string
		switch job.Status {
		case "completed":
			statusDisplay = ui.GreenText("completed")
		case "processing":
			statusDisplay = ui.CyanText("processing")
		case "pending":
			statusDisplay = "pending"
		case "failed":
			statusDisplay = ui.RedText("failed")
		case "cancelled":
			statusDisplay = ui.DimmedText("cancelled")
		default:
			statusDisplay = job.Status
		}

		// Progress
		progressDisplay := "-"
		if job.Progress > 0 {
			progressDisplay = fmt.Sprintf("%d%%", job.Progress)
		}

		// Created time
		createdDisplay := "-"
		if job.CreatedAt != nil {
			createdDisplay = formatRelativeTime(job.CreatedAt.Time)
		}

		// Truncate job ID for display
		jobID := job.ID
		if len(jobID) > 12 {
			jobID = jobID[:12] + ".."
		}

		fmt.Printf("  %-14s %-12s %-10s %s\n",
			jobID, statusDisplay, progressDisplay, createdDisplay)
	}

	fmt.Println()
	return nil
}

func runJobsCancel(cmd *cobra.Command, args []string) error {
	jobID := args[0]

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	// User from second arg or default
	var userID string
	if len(args) > 1 {
		userID = args[1]
	} else {
		userID, err = formation.MustResolveUserID("")
		if err != nil {
			return fmt.Errorf("user ID required: provide as second argument or set default with: muxi set default user")
		}
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Cancelling job '%s'...", jobID))
	spinner.Start()

	err = client.CancelJob(userID, jobID)
	if err != nil {
		spinner.StopWithError("Cancel failed")
		return err
	}

	spinner.StopWithSuccess("Job cancelled")
	fmt.Println()
	ui.Success(fmt.Sprintf("Cancelled job %s", jobID))
	fmt.Println()

	return nil
}

// formatRelativeTime formats a time as relative (e.g., "2 minutes ago")
func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
