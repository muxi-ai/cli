package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

// historyCmd is a deprecated alias for "sessions messages"
var historyCmd = &cobra.Command{
	Use:        "history",
	Short:      "View session message history (use 'sessions messages' instead)",
	GroupID:    "formation",
	Deprecated: "use 'muxi sessions messages <session-id>' instead",
	Hidden:     true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		ui.Warning("'muxi history' is deprecated. Use 'muxi sessions messages <session-id>' instead.")
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
