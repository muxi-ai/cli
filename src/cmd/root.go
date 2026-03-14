package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/muxi-ai/cli/pkg/telemetry"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/updates"
	"github.com/spf13/cobra"
)

// RequireArgs returns a custom Args validator that shows help instead of error
func RequireArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			cmd.Help()
			return fmt.Errorf("") // Empty error to suppress additional output
		}
		return nil
	}
}

// Colors for help output
var (
	gold   = color.RGB(216, 137, 62) // Brand color #d8893e
	cyan   = color.New(color.FgCyan)
	dimmed = color.New(color.Faint)
)

var (
	version string
	commit  string
	date    string
)

const boxWidth = 64

var rootCmd = &cobra.Command{
	Use:           "muxi",
	Short:         "MUXI CLI - Formation development and server management",
	Long:          "Build, deploy, and manage MUXI formations with ease.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Add --version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(fmt.Sprintf("muxi version %s\n", version))

	// Define command groups (used internally for organization)
	rootCmd.AddGroup(&cobra.Group{ID: "formation", Title: "Formation Commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "registry", Title: "Registry Commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "remote", Title: "Remote Commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "config", Title: "Setup Commands"})

	// Override the help function with custom formatting and telemetry
	rootCmd.SetHelpFunc(helpWithTelemetry)

	// Add subcommands here as we build them
	rootCmd.AddCommand(newCmd)
}

// helpWithTelemetry wraps customHelp with telemetry tracking
func helpWithTelemetry(cmd *cobra.Command, args []string) {
	// Track --help usage
	state := telemetry.Load()
	state.IncrementHelp(cmd.Name())
	state.FlushIfDue()
	state.Save()

	customHelp(cmd, args)
}

// customHelp provides a beautifully formatted help output
func customHelp(cmd *cobra.Command, args []string) {
	// If help is requested for a subcommand, use default Cobra help
	if cmd != rootCmd {
		// Print Long description if available
		if cmd.Long != "" {
			cmd.Println(cmd.Long)
			cmd.Println()
		}
		// Generate default help template
		cmd.Println(cmd.UsageString())
		return
	}

	// MUXI header with logo
	ui.MUXIHeader(version, runtime.GOARCH)

	// Formation Commands
	printBoxWithSubtitle("Formation Commands", "MUXI")
	printColoredCommand("$ muxi remote [command]")
	printBoxLineDimmed("  or (from formation dir):")
	printColoredCommand("$ muxi [command]")
	printBoxBottom()
	printCommandGroup(cmd, "formation")
	fmt.Println()

	// Registry Commands
	printBoxWithSubtitle("Registry Commands", "MUXI")
	printColoredCommand("$ muxi registry [command]")
	printBoxBottom()
	printCommandGroup(cmd, "registry")
	fmt.Println()

	// Server Commands
	printBoxWithSubtitle("Server Commands", "MUXI")
	printColoredCommand("$ muxi remote [command]")
	printBoxBottom()
	printCommandGroup(cmd, "remote")
	fmt.Println()

	// Setup Commands
	printBoxSimple("Setup Commands", "MUXI")
	printCommandGroup(cmd, "config")
	fmt.Println()

	// Additional Commands (ungrouped)
	printBoxSimple("Additional Commands", "MUXI")
	printUngroupedCommands(cmd)
	fmt.Println()

	// Flags
	printBoxSimple("Flags", "MUXI")
	fmt.Printf("  -h, --help       %s\n", dimmed.Sprint("Help for MUXI CLI"))
	fmt.Printf("  -v, --version    %s\n", dimmed.Sprint("Version for MUXI CLI"))
	fmt.Println()

	// Footer
	printDivider()
	fmt.Println()
	dimmed.Println("Use \"muxi [command] --help\" for more information about a command.")
	fmt.Println()
}

// Box drawing helpers
func printBox(title, subtitle string) {
	fmt.Print("╭")
	fmt.Print(strings.Repeat("─", boxWidth-2))
	fmt.Println("╮")
	printBoxLine(title)
}

func printBoxWithSubtitle(title, subtitle string) {
	fmt.Print("╭")
	fmt.Print(strings.Repeat("─", boxWidth-2))
	fmt.Println("╮")

	// Title with colored subtitle right-aligned
	content := title
	padding := boxWidth - 4 - len(content) - len(subtitle)
	if padding < 1 {
		padding = 1
	}
	fmt.Printf("│ %s%s%s │\n", content, strings.Repeat(" ", padding), gold.Sprint(subtitle))

	printBoxDivider()
}

func printBoxSimple(title, subtitle string) {
	fmt.Print("╭")
	fmt.Print(strings.Repeat("─", boxWidth-2))
	fmt.Println("╮")

	content := title
	padding := boxWidth - 4 - len(content) - len(subtitle)
	if padding < 1 {
		padding = 1
	}
	fmt.Printf("│ %s%s%s │\n", content, strings.Repeat(" ", padding), gold.Sprint(subtitle))

	fmt.Print("╰")
	fmt.Print(strings.Repeat("─", boxWidth-2))
	fmt.Println("╯")
}

func printBoxLine(content string) {
	padding := boxWidth - 4 - len(content)
	if padding < 0 {
		padding = 0
		content = content[:boxWidth-4]
	}
	fmt.Printf("│ %s%s │\n", content, strings.Repeat(" ", padding))
}

func printBoxLineDimmed(content string) {
	padding := boxWidth - 4 - len(content)
	if padding < 0 {
		padding = 0
		content = content[:boxWidth-4]
	}
	fmt.Printf("│ %s%s │\n", dimmed.Sprint(content), strings.Repeat(" ", padding))
}

func printBoxDivider() {
	fmt.Print("│")
	fmt.Print(strings.Repeat("─", boxWidth-2))
	fmt.Println("│")
}

func printBoxBottom() {
	fmt.Print("╰")
	fmt.Print(strings.Repeat("─", boxWidth-2))
	fmt.Println("╯")
}

func printDivider() {
	fmt.Println(strings.Repeat("─", boxWidth))
}

func printColoredCommand(content string) {
	// Color $ as dimmed, muxi as cyan
	colored := strings.Replace(content, "$", dimmed.Sprint("$"), 1)
	colored = strings.Replace(colored, "muxi", cyan.Sprint("muxi"), 1)

	// Calculate padding (account for ANSI codes not taking visual space)
	visualLen := len(content)
	padding := boxWidth - 4 - visualLen
	if padding < 0 {
		padding = 0
	}
	fmt.Printf("│ %s%s │\n", colored, strings.Repeat(" ", padding))
}

func printCommandGroup(cmd *cobra.Command, groupID string) {
	for _, c := range cmd.Commands() {
		if c.GroupID == groupID && c.Name() != "help" {
			fmt.Printf("  %-12s%s\n", c.Name(), dimmed.Sprint(c.Short))
		}
	}
}

func printUngroupedCommands(cmd *cobra.Command) {
	for _, c := range cmd.Commands() {
		if c.GroupID == "" && c.Name() != "help" && c.Name() != "completion" && !c.Hidden {
			fmt.Printf("  %-12s%s\n", c.Name(), dimmed.Sprint(c.Short))
		}
	}
	// Always show these
	fmt.Printf("  %-12s%s\n", "completion", dimmed.Sprint("Generate autocompletion script for your shell"))
	fmt.Printf("  %-12s%s\n", "help", dimmed.Sprint("Help about any command"))
}

// SetVersionInfo sets version information from main
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = v
	rootCmd.SetVersionTemplate(fmt.Sprintf("muxi version %s\n", v))
	updates.SetCurrentVersion(v)
}

// Execute runs the root command
func Execute() error {
	// Check for updates from cache (instant, no network)
	updateInfo := updates.CheckCachedUpdate()

	// Run the command
	err := rootCmd.Execute()

	// Show update notification if available (after command output)
	if updateInfo != nil && os.Getenv("MUXI_NO_UPDATE_CHECK") == "" {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "[*] MUXI CLI %s available (current: %s)\n", updateInfo.LatestVersion, updateInfo.CurrentVersion)
		fmt.Fprintln(os.Stderr, "    Run: muxi upgrade")
	}

	// Refresh cache in background for next run (fire-and-forget)
	go updates.RefreshCache()

	return err
}
