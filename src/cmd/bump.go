package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var bumpCmd = &cobra.Command{
	Use:   "bump [patch|minor|major]",
	Short: "Bump the formation version",
	Long: `Bump the formation version before deploying updates.

Examples:
  muxi bump              # Bump patch: 1.0.0 -> 1.0.1
  muxi bump minor        # Bump minor: 1.0.0 -> 1.1.0
  muxi bump major        # Bump major: 1.0.0 -> 2.0.0
  muxi bump --set 2.0.0  # Set specific version

If no version exists, defaults to 1.0.0 then applies the bump.`,
	Args:    cobra.MaximumNArgs(1),
	GroupID: "formation",
	RunE:    runBump,
}

var setVersion string

func init() {
	bumpCmd.Flags().StringVar(&setVersion, "set", "", "Set specific version (e.g., 2.0.0)")
	rootCmd.AddCommand(bumpCmd)
}

func runBump(cmd *cobra.Command, args []string) error {
	// Must be in formation directory
	ctx, err := context.DetectFormation()
	if err != nil || ctx == nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation",
		)
		return fmt.Errorf("not in formation directory")
	}

	// Find the formation file
	formationFile, found := context.FindFormationFile(ctx.RootDir)
	if !found {
		ui.ErrorBlock(
			"Formation config not found",
			"Could not find formation.afs or formation.yaml",
			"Create a formation first:\n  muxi new formation",
		)
		return fmt.Errorf("formation config not found")
	}

	// Read the file
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read formation config: %w", err)
	}

	contentStr := string(content)

	// Extract current version
	currentVersion := extractVersion(contentStr)
	if currentVersion == "" {
		currentVersion = "0.0.0" // Will become 1.0.0 after patch bump, or as specified
	}

	// Determine new version
	var newVersion string
	if setVersion != "" {
		// Validate the set version
		if !isValidVersion(setVersion) {
			ui.ErrorBlock(
				"Invalid version",
				fmt.Sprintf("'%s' is not a valid semver version.", setVersion),
				"Use format: major.minor.patch (e.g., 2.0.0)",
			)
			return fmt.Errorf("invalid version")
		}
		newVersion = setVersion
	} else {
		// Determine bump type
		bumpType := "patch"
		if len(args) > 0 {
			bumpType = strings.ToLower(args[0])
			if bumpType != "patch" && bumpType != "minor" && bumpType != "major" {
				ui.ErrorBlock(
					"Invalid bump type",
					fmt.Sprintf("'%s' is not a valid bump type.", args[0]),
					"Use: patch, minor, or major",
				)
				return fmt.Errorf("invalid bump type")
			}
		}

		// If no version exists and bumping patch, start at 1.0.0
		if currentVersion == "0.0.0" && bumpType == "patch" {
			newVersion = "1.0.0"
		} else if currentVersion == "0.0.0" {
			// For minor/major on missing version, start from 1.0.0 then bump
			currentVersion = "1.0.0"
			newVersion = bumpVersion(currentVersion, bumpType)
		} else {
			newVersion = bumpVersion(currentVersion, bumpType)
		}
	}

	// Update the file
	var newContent string
	versionRegex := regexp.MustCompile(`(?m)^version:\s*["']?([^"'\n]+)["']?\s*$`)
	if versionRegex.MatchString(contentStr) {
		// Replace existing version
		newContent = versionRegex.ReplaceAllString(contentStr, fmt.Sprintf(`version: "%s"`, newVersion))
	} else {
		// Add version after id field, or at the beginning if no id
		idRegex := regexp.MustCompile(`(?m)^id:\s*["']?[^"'\n]+["']?\s*$`)
		if idRegex.MatchString(contentStr) {
			// Add after id line
			newContent = idRegex.ReplaceAllStringFunc(contentStr, func(match string) string {
				return match + fmt.Sprintf("\nversion: \"%s\"", newVersion)
			})
		} else {
			// Add at the beginning (after schema if present)
			schemaRegex := regexp.MustCompile(`(?m)^schema:\s*["']?[^"'\n]+["']?\s*$`)
			if schemaRegex.MatchString(contentStr) {
				newContent = schemaRegex.ReplaceAllStringFunc(contentStr, func(match string) string {
					return match + fmt.Sprintf("\nversion: \"%s\"", newVersion)
				})
			} else {
				// Just prepend
				newContent = fmt.Sprintf("version: \"%s\"\n", newVersion) + contentStr
			}
		}
	}

	// Write back
	if err := os.WriteFile(formationFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write formation config: %w", err)
	}

	// Display result
	fmt.Println()
	ui.Success("Formation version updated!")
	if currentVersion == "0.0.0" {
		fmt.Printf("  (none) => %s\n", newVersion)
	} else {
		fmt.Printf("  %s => %s\n", currentVersion, newVersion)
	}

	return nil
}

func extractVersion(content string) string {
	versionRegex := regexp.MustCompile(`(?m)^version:\s*["']?([^"'\n]+)["']?\s*$`)
	matches := versionRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func isValidVersion(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	return true
}

func bumpVersion(current, bumpType string) string {
	parts := strings.Split(current, ".")
	if len(parts) != 3 {
		return "1.0.0"
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
