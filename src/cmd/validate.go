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
	Use:   "validate",
	Short: "Validate formation configuration",
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
			location := issue.File
			if issue.Field != "" {
				location = fmt.Sprintf("%s → %s", issue.File, issue.Field)
			}
			if location == "" {
				location = issue.Field
			}

			fmt.Printf("  %s %s\n", ui.RedText("✗"), location)
			printIssueMessage(issue.Message)
			fmt.Println()
		}
	}

	// Show warnings
	if len(result.Warnings) > 0 {
		ui.Warning(fmt.Sprintf("Found %d warning(s)", len(result.Warnings)))
		fmt.Println()

		for _, issue := range result.Warnings {
			location := issue.File
			if issue.Field != "" {
				location = fmt.Sprintf("%s → %s", issue.File, issue.Field)
			}
			if location == "" {
				location = issue.Field
			}

			fmt.Printf("  %s %s\n", ui.YellowText("⚠"), location)
			printIssueMessage(issue.Message)
			fmt.Println()
		}
	}

	// Summary
	if result.IsValid() {
		ui.Success(fmt.Sprintf("Formation '%s' is valid (with warnings)", formationID))
	} else {
		fmt.Printf("  Run %s to fix configuration issues\n", ui.BoldText("muxi config <section>"))
		fmt.Printf("  Run %s to set missing secrets\n", ui.BoldText("muxi secrets set <KEY>"))
	}
	fmt.Println()
}

// printIssueMessage prints the issue message with hints dimmed
func printIssueMessage(message string) {
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "*") {
			// Hint line - print dimmed
			fmt.Printf("    %s\n", ui.DimmedText(line))
		} else {
			// Regular line
			fmt.Printf("    %s\n", line)
		}
	}
}
