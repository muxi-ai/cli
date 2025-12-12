package cmd

import (
	"fmt"
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
	RunE: runMemoryStatus,
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
	Args: cobra.ExactArgs(1),
	RunE: runMemoryAdd,
}

var memoryDeleteCmd = &cobra.Command{
	Use:   "delete <memory-id>",
	Short: "Delete a memory",
	Long: `Delete a specific memory by ID.

Example:
  muxi memory delete -u alice mem_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runMemoryDelete,
}

func init() {
	rootCmd.AddCommand(memoryCmd)
	memoryCmd.AddCommand(memoryStatusCmd)
	memoryCmd.AddCommand(memoryListCmd)
	memoryCmd.AddCommand(memoryAddCmd)
	memoryCmd.AddCommand(memoryDeleteCmd)

	formation.AddFormationFlag(memoryCmd)
	formation.AddProfileFlag(memoryCmd)

	formation.AddFormationFlag(memoryStatusCmd)
	formation.AddProfileFlag(memoryStatusCmd)

	formation.AddCommonFlags(memoryListCmd)
	formation.AddCommonFlags(memoryAddCmd)
	formation.AddCommonFlags(memoryDeleteCmd)
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
