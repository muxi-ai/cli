package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var memoryCmd = &cobra.Command{
	Use:     "memory",
	Short:   "View and manage memory",
	GroupID: "formation",
	Long: `View memory configuration and manage user memories.

Memory stores context and preferences for users across sessions.`,
}

var memoryStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show memory configuration",
	Long:  `Display memory buffer and working memory configuration.`,
	RunE:  runMemoryStatus,
}

var memoryListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List user memories",
	Long: `List all stored memories for a user.

Requires -u flag to specify the user.`,
	RunE: runMemoryList,
}

var memoryAddCmd = &cobra.Command{
	Use:   "add <content>",
	Short: "Add a memory for a user",
	Long: `Add a new memory entry for a user.

Examples:
  muxi memory add -u alice "Prefers dark mode"
  muxi memory add -u alice "Uses TypeScript for all projects"`,
	Args: RequireArgs(1),
	RunE: runMemoryAdd,
}

var memoryDeleteCmd = &cobra.Command{
	Use:   "delete <memory-id>",
	Short: "Delete a memory",
	Long: `Delete a specific memory by ID.

Example:
  muxi memory delete -u alice mem_abc123`,
	Args: RequireArgs(1),
	RunE: runMemoryDelete,
}

// Buffer subcommand
var memoryBufferCmd = &cobra.Command{
	Use:   "buffer",
	Short: "Manage memory buffers",
	Long:  `View and manage memory buffers for users and sessions.`,
}

var memoryBufferListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List user's buffer",
	Long: `List buffer data for a user including sessions and message counts.

Requires -u flag to specify the user.`,
	RunE: runMemoryBufferList,
}

var memoryBufferStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show aggregate buffer stats",
	Long:  `Display aggregate buffer statistics across all users. Admin only.`,
	RunE:  runMemoryBufferStatus,
}

var memoryBufferClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear buffer",
	Long: `Clear buffer memory.

Examples:
  muxi memory buffer clear -u alice          # Clear user's buffer
  muxi memory buffer clear -s sess_abc123    # Clear session buffer
  muxi memory buffer clear --all             # Clear ALL buffers (requires confirmation)
  muxi memory buffer clear --all --force     # Clear ALL without confirmation`,
	RunE: runMemoryBufferClear,
}

func init() {
	rootCmd.AddCommand(memoryCmd)
	memoryCmd.AddCommand(memoryStatusCmd)
	memoryCmd.AddCommand(memoryListCmd)
	memoryCmd.AddCommand(memoryAddCmd)
	memoryCmd.AddCommand(memoryDeleteCmd)
	memoryCmd.AddCommand(memoryBufferCmd)

	// Buffer subcommands
	memoryBufferCmd.AddCommand(memoryBufferListCmd)
	memoryBufferCmd.AddCommand(memoryBufferStatusCmd)
	memoryBufferCmd.AddCommand(memoryBufferClearCmd)

	formation.AddFormationFlag(memoryCmd)
	formation.AddProfileFlag(memoryCmd)

	formation.AddFormationFlag(memoryStatusCmd)
	formation.AddProfileFlag(memoryStatusCmd)

	formation.AddCommonFlags(memoryListCmd)
	formation.AddCommonFlags(memoryAddCmd)
	formation.AddCommonFlags(memoryDeleteCmd)

	// Buffer command flags
	formation.AddCommonFlags(memoryBufferListCmd)

	formation.AddFormationFlag(memoryBufferStatusCmd)
	formation.AddProfileFlag(memoryBufferStatusCmd)

	formation.AddFormationFlag(memoryBufferClearCmd)
	formation.AddProfileFlag(memoryBufferClearCmd)
	formation.AddUserFlag(memoryBufferClearCmd)
	memoryBufferClearCmd.Flags().StringP("session", "s", "", "Session ID to clear")
	memoryBufferClearCmd.Flags().Bool("all", false, "Clear ALL buffers (admin)")
	memoryBufferClearCmd.Flags().Bool("force", false, "Skip confirmation for --all")
}

func runMemoryStatus(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	config, err := client.GetMemoryConfig()
	if err != nil {
		return fmt.Errorf("failed to get memory config: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	ui.Bold("Memory Configuration")
	fmt.Println()

	fmt.Println("  Buffer:")
	fmt.Printf("    Size:          %d messages\n", config.Buffer.Size)
	fmt.Printf("    Multiplier:    %.1f\n", config.Buffer.Multiplier)
	vectorSearch := ui.RedText("disabled")
	if config.Buffer.VectorSearch {
		vectorSearch = ui.GreenText("enabled")
	}
	fmt.Printf("    Vector Search: %s\n", vectorSearch)
	fmt.Println()

	fmt.Println("  Working Memory:")
	fmt.Printf("    Max Size:      %d MB\n", config.Working.MaxMemoryMB)
	fmt.Printf("    FIFO Interval: %d minutes\n", config.Working.FIFOIntervalMin)
	fmt.Println()

	return nil
}

func runMemoryList(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	resp, err := client.GetMemories(userID)
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  No memories for user '%s'", userID))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Printf("  Memories for user '%s' (%d):\n\n", userID, resp.Count)
	fmt.Printf("  %-20s %-45s %s\n", "ID", "CONTENT", "CREATED")
	fmt.Printf("  %s\n", strings.Repeat("─", 80))

	for _, mem := range resp.Memories {
		created := "-"
		if !mem.CreatedAt.IsZero() {
			created = mem.CreatedAt.Format("Jan 2, 2006")
		}

		content := mem.Content
		if len(content) > 45 {
			content = content[:42] + "..."
		}

		fmt.Printf("  %-20s %-45s %s\n",
			truncate(mem.ID, 20),
			content,
			created)
	}
	fmt.Println()

	return nil
}

func runMemoryAdd(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	content := args[0]

	mem, err := client.AddMemory(userID, content)
	if err != nil {
		return fmt.Errorf("failed to add memory: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Added memory '%s' for user '%s'", mem.ID, userID))
	fmt.Println()

	return nil
}

func runMemoryDelete(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	memoryID := args[0]

	err = client.DeleteMemory(userID, memoryID)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Deleted memory '%s'", memoryID))
	fmt.Println()

	return nil
}

func runMemoryBufferList(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	resp, err := client.GetUserBuffer(userID)
	if err != nil {
		return fmt.Errorf("failed to get buffer: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	fmt.Printf("  Buffer for user '%s'\n\n", userID)
	fmt.Printf("    Total Messages: %d\n", resp.TotalMessages)
	fmt.Printf("    Buffer Size:    %.2f KB\n", resp.BufferSizeKB)

	if len(resp.Sessions) == 0 {
		fmt.Println()
		ui.Dimmed("    No sessions")
	} else {
		fmt.Println()
		fmt.Printf("    %-25s %-10s %s\n", "SESSION", "MESSAGES", "LAST ACTIVITY")
		fmt.Printf("    %s\n", strings.Repeat("─", 55))
		for _, sess := range resp.Sessions {
			lastActivity := "-"
			if !sess.LastActivity.IsZero() {
				lastActivity = sess.LastActivity.Format("Jan 2, 2006 3:04pm")
			}
			fmt.Printf("    %-25s %-10d %s\n",
				truncate(sess.SessionID, 25),
				sess.MessageCount,
				lastActivity)
		}
	}
	fmt.Println()

	return nil
}

func runMemoryBufferStatus(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	resp, err := client.GetBufferStats()
	if err != nil {
		return fmt.Errorf("failed to get buffer stats: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	fmt.Print("  ")
	ui.Bold("Buffer Statistics")
	fmt.Println()

	fmt.Printf("    Total Entries:  %d\n", resp.TotalEntries)
	fmt.Printf("    Total Users:    %d\n", resp.TotalUsers)
	fmt.Printf("    Total Sessions: %d\n", resp.TotalSessions)
	fmt.Printf("    Buffer Size:    %.2f KB\n", resp.BufferSizeKB)
	fmt.Printf("    Max Size:       %d\n", resp.MaxSize)
	fmt.Printf("    Utilization:    %.1f%%\n", resp.Utilization*100)
	fmt.Println()

	return nil
}

func runMemoryBufferClear(cmd *cobra.Command, args []string) error {
	userFlag, _ := cmd.Flags().GetString("user")
	sessionFlag, _ := cmd.Flags().GetString("session")
	allFlag, _ := cmd.Flags().GetBool("all")
	forceFlag, _ := cmd.Flags().GetBool("force")

	// Validate flags
	if userFlag == "" && sessionFlag == "" && !allFlag {
		fmt.Println()
		ui.ErrorBlock(
			"Missing required flag",
			"Specify -u <user>, -s <session>, or --all",
			"",
		)
		fmt.Println("Usage:")
		fmt.Println("  muxi memory buffer clear -u <user>      Clear a user's buffer")
		fmt.Println("  muxi memory buffer clear -s <session>   Clear a session's buffer")
		fmt.Println("  muxi memory buffer clear --all          Clear ALL buffers")
		fmt.Println()
		return nil
	}

	// Can't combine flags
	flagCount := 0
	if userFlag != "" {
		flagCount++
	}
	if sessionFlag != "" {
		flagCount++
	}
	if allFlag {
		flagCount++
	}
	if flagCount > 1 {
		ui.ErrorBlock(
			"Invalid flags",
			"Use only one of: -u <user>, -s <session>, or --all",
			"",
		)
		return nil
	}

	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	// Handle --all
	if allFlag {
		if !forceFlag {
			fmt.Println()
			fmt.Print("  Clear ALL buffers for ALL users? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println()
				ui.Dimmed("  Cancelled")
				fmt.Println()
				return nil
			}
		}

		resp, err := client.ClearAllBuffers()
		if err != nil {
			fmt.Printf("\n  %s Failed to clear buffers\n", ui.RedText("✗"))
			fmt.Printf("    %s\n\n", ui.DimmedText(err.Error()))
			return nil
		}

		fmt.Printf("\n  %s Cleared all buffers\n", ui.GreenText("✓"))
		fmt.Printf("    Messages cleared: %d\n", resp.MessagesCleared)
		fmt.Printf("    Sessions cleared: %d\n", resp.SessionsCleared)
		fmt.Println()
		return nil
	}

	// Handle -s <session>
	if sessionFlag != "" {
		userID := formation.ResolveUserID(userFlag)
		resp, err := client.ClearSessionBuffer(sessionFlag, userID)
		if err != nil {
			fmt.Printf("\n  %s Failed to clear session buffer\n", ui.RedText("✗"))
			fmt.Printf("    %s\n\n", ui.DimmedText(err.Error()))
			return nil
		}

		fmt.Printf("\n  %s Cleared session buffer: %s\n", ui.GreenText("✓"), sessionFlag)
		fmt.Printf("    Messages cleared: %d\n", resp.MessagesCleared)
		fmt.Println()
		return nil
	}

	// Handle -u <user>
	userID := formation.ResolveUserID(userFlag)
	if userID == "" {
		ui.ErrorBlock("User ID required", "Use -u flag to specify user", "")
		return nil
	}

	resp, err := client.ClearUserBuffer(userID)
	if err != nil {
		fmt.Printf("\n  %s Failed to clear user buffer\n", ui.RedText("✗"))
		fmt.Printf("    %s\n\n", ui.DimmedText(err.Error()))
		return nil
	}

	fmt.Printf("\n  %s Cleared buffer for user: %s\n", ui.GreenText("✓"), userID)
	fmt.Printf("    Messages cleared: %d\n", resp.MessagesCleared)
	fmt.Printf("    Sessions cleared: %d\n", resp.SessionsCleared)
	fmt.Println()

	return nil
}
