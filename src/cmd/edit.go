package cmd

import (
	"fmt"
	"os"

	"github.com/muxi-ai/cli/pkg/scaffold"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [component] [id]",
	Short: "Open component files in your editor",
	Long: `Open component configuration files in your default editor ($EDITOR).

Examples:
  muxi edit formation                   # Opens formation.yaml
  muxi edit agent weather               # Opens agents/weather.yaml
  muxi edit mcp weather-api             # Opens mcps/weather-api.yaml (or agent file if agent-level)
  muxi edit sop customer-onboarding     # Opens sops/customer-onboarding.md
  muxi edit trigger webhook-handler     # Opens triggers/webhook-handler.yaml
  muxi edit a2a-service external-billing # Opens a2a/external-billing.yaml

The editor used is determined by $EDITOR environment variable.
Falls back to: vim (Unix) or notepad (Windows) if $EDITOR is not set.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		component := args[0]
		
		var id string
		if len(args) > 1 {
			id = args[1]
		}
		
		if err := scaffold.EditComponent(component, id); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
