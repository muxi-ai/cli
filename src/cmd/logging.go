package cmd

import (
	"fmt"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var loggingCmd = &cobra.Command{
	Use:     "logging",
	Short:   "View logging configuration",
	GroupID: "formation",
	Long:    `View logging configuration and destinations.`,
}

var loggingStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show logging configuration",
	Long:  `Display logging configuration including enabled streams.`,
	RunE:  runLoggingStatus,
}

var loggingListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List logging destinations",
	Long:    `List all configured logging destinations.`,
	RunE:    runLoggingList,
}

func init() {
	rootCmd.AddCommand(loggingCmd)
	loggingCmd.AddCommand(loggingStatusCmd)
	loggingCmd.AddCommand(loggingListCmd)

	formation.AddFormationFlag(loggingStatusCmd)
	formation.AddProfileFlag(loggingStatusCmd)

	formation.AddFormationFlag(loggingListCmd)
	formation.AddProfileFlag(loggingListCmd)
}

func runLoggingStatus(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	config, err := client.GetLoggingConfig()
	if err != nil {
		return fmt.Errorf("failed to get logging config: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	fmt.Print("  ")
	ui.Bold("Logging Configuration")
	fmt.Println()

	// System Events section
	fmt.Println()
	fmt.Print("  ")
	ui.Bold("System Events")
	fmt.Println()
	fmt.Printf("    Level:       %s\n", config.System.Level)
	fmt.Printf("    Destination: %s\n", config.System.Destination)

	// Conversation Events section
	fmt.Println()
	fmt.Print("  ")
	ui.Bold("Conversation Events")
	fmt.Println()

	enabledStr := ui.RedText("✗ disabled")
	if config.Conversation.Enabled {
		enabledStr = ui.GreenText("✓ enabled")
	}
	fmt.Printf("    Status:  %s\n", enabledStr)

	// Count destinations by transport
	transportCounts := make(map[string]int)
	for _, stream := range config.Conversation.Streams {
		transportCounts[stream.Transport]++
	}

	fmt.Printf("    Streams: %d\n", len(config.Conversation.Streams))
	// Show in consistent order
	for _, transport := range []string{"stdout", "file", "stream"} {
		if count, ok := transportCounts[transport]; ok {
			fmt.Printf("      • %-6s (%d)\n", transport, count)
		}
	}

	// Build the hint command with resolved formation/profile
	flags := formation.GetCommonFlags(cmd)
	formationID, _ := formation.ResolveFormationID(flags.FormationID)
	profile := formation.ResolveProfile(flags.Profile)

	fmt.Println()
	ui.Dimmed("  List conversation destinations with:")
	ui.Dimmed(fmt.Sprintf("    muxi logging list -f %s -p %s", formationID, profile))
	fmt.Println()

	return nil
}

func runLoggingList(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	resp, err := client.GetLoggingDestinations()
	if err != nil {
		return fmt.Errorf("failed to get logging destinations: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	// System Events section
	fmt.Println()
	fmt.Print("  ")
	ui.Bold("System Events")
	fmt.Println()
	fmt.Printf("    Level:       %s\n", resp.System.Level)
	fmt.Printf("    Destination: %s\n", resp.System.Destination)

	// Conversation Destinations section
	fmt.Println()
	fmt.Print("  ")
	ui.Bold(fmt.Sprintf("Conversation Destinations (%d)", resp.Conversation.Count))
	fmt.Println()

	if resp.Conversation.Count == 0 {
		ui.Dimmed("    No conversation destinations configured")
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Printf("  %-15s %-10s %-25s %-8s %-8s %s\n",
		"ID", "TRANSPORT", "DESTINATION", "LEVEL", "FORMAT", "STATUS")
	fmt.Printf("  %s\n", strings.Repeat("─", 85))

	for _, dest := range resp.Conversation.Destinations {
		destination := dest.Destination
		if destination == "" {
			destination = "-"
		}
		if len(destination) > 25 {
			destination = destination[:22] + "..."
		}

		status := ui.RedText("● off")
		if dest.Enabled {
			status = ui.GreenText("● on")
		}

		fmt.Printf("  %-15s %-10s %-25s %-8s %-8s %s\n",
			truncate(dest.ID, 15),
			dest.Transport,
			destination,
			dest.Level,
			dest.Format,
			status)
	}
	fmt.Println()

	return nil
}
