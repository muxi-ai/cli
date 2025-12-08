package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/validate"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Validate formation configuration",
	GroupID: "formation",
	Long: `Validate the current formation's configuration files.

Checks for:
  - Required fields in formation.yaml
  - Valid schema version
  - Server configuration
  - LLM configuration
  - Secret references (ensures referenced secrets exist)
  - Agent file validity
  - MCP file validity

Must be run inside a formation directory.

Examples:
  # Validate current formation
  muxi validate

  # Validate and see all details
  muxi validate --verbose`,
	Args: cobra.NoArgs,
	RunE: runValidate,
}

var validateVerbose bool

func init() {
	validateCmd.Flags().BoolVarP(&validateVerbose, "verbose", "v", false, "Show detailed validation information")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Must be in formation directory
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation",
		)
		os.Exit(1)
	}

	// Run validation
	result, err := validate.Formation(ctx.RootDir)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Display results
	displayValidationResults(ctx.ID, result)

	// Exit with error code if validation failed
	if !result.IsValid() {
		os.Exit(1)
	}

	return nil
}

func displayValidationResults(formationID string, result *validate.Result) {
	fmt.Println()

	if result.IsValid() && len(result.Warnings) == 0 {
		ui.Success(fmt.Sprintf("Formation '%s' is valid", formationID))
		fmt.Println()
		return
	}

	// Show errors
	if len(result.Errors) > 0 {
		ui.Error(fmt.Sprintf("Found %d error(s)", len(result.Errors)))
		fmt.Println()

		for _, issue := range result.Errors {
			// Print first line of message with error symbol
			lines := strings.Split(issue.Message, "\n")
			fmt.Printf("  %s %s\n", ui.RedText("✗"), lines[0])
			// Print remaining lines (hints) dimmed
			for _, line := range lines[1:] {
				fmt.Printf("    %s\n", ui.DimmedText(strings.TrimPrefix(line, " ")))
			}
			fmt.Println()
		}
	}

	// Show warnings
	if len(result.Warnings) > 0 {
		ui.Warning(fmt.Sprintf("Found %d warning(s)", len(result.Warnings)))
		fmt.Println()

		for _, issue := range result.Warnings {
			// Print first line of message with warning symbol
			lines := strings.Split(issue.Message, "\n")
			fmt.Printf("  %s %s\n", ui.YellowText("⚠"), lines[0])
			// Print remaining lines (hints) dimmed
			for _, line := range lines[1:] {
				fmt.Printf("    %s\n", ui.DimmedText(strings.TrimPrefix(line, " ")))
			}
			fmt.Println()
		}
	}

	// Summary
	if result.IsValid() {
		ui.Success(fmt.Sprintf("Formation '%s' is valid (with warnings)", formationID))
	} else {
		fmt.Printf("  Run %s to fix configuration issues\n", ui.Command("muxi config <section>"))
		fmt.Printf("  Run %s to set missing secrets\n", ui.Command("muxi secrets set <KEY>"))
	}
	fmt.Println()
}
