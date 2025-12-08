package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Colors for help output
var (
	gold   = color.New(color.FgHiYellow)
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
	Use:   "muxi",
	Short: "MUXI CLI - Formation development and server management",
	Long:  "Build, deploy, and manage MUXI formations with ease.",
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
	rootCmd.AddGroup(&cobra.Group{ID: "server", Title: "Server Commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "config", Title: "Setup Commands"})

	// Override the help function with custom formatting
	rootCmd.SetHelpFunc(customHelp)

	// Add subcommands here as we build them
	rootCmd.AddCommand(newCmd)
}

// customHelp provides a beautifully formatted help output
func customHelp(cmd *cobra.Command, args []string) {
	// If help is requested for a subcommand, use default help
	if cmd != rootCmd {
		cmd.SetHelpFunc(nil)
		cmd.Help()
		return
	}

	// Version header outside box
	fmt.Print("MUXI CLI • ")
	dimmed.Println(version)

	// Header box
	fmt.Println("╭──────────────────────────────────────────────────────────────╮")
	fmt.Println("│ Build, deploy, and manage formations with ease               │")
	printBoxDivider()
	printBoxLine("Examples:")
	printColoredCommand(" $ muxi new formation my-bot")
	printColoredCommand(" $ muxi deploy --profile production")
	printColoredCommand(" $ muxi agents list")
	printBoxDivider()
	printBoxLine("Usage:")
	printColoredCommand(" $ muxi [command]")
	printBoxBottom()
	fmt.Println()

	// Formation Commands
	printBoxWithSubtitle("Formation Commands", "MUXI")
	printColoredCommand("$ muxi formation [command]")
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
	printColoredCommand("$ muxi server [command]")
	printBoxBottom()
	printCommandGroup(cmd, "server")
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
		if c.GroupID == "" && c.Name() != "help" && c.Name() != "completion" {
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
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
