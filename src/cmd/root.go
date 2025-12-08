package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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

	// Header box
	printBox("MUXI CLI • Build, deploy, and manage formations with ease", "")
	printBoxLine("Examples:")
	printBoxLine(" $ muxi new formation my-bot")
	printBoxLine(" $ muxi deploy --profile production")
	printBoxLine(" $ muxi agents list")
	printBoxDivider()
	printBoxLine("Usage:")
	printBoxLine(" $ muxi [command]")
	printBoxBottom()
	fmt.Println()

	// Formation Commands
	printBoxWithSubtitle("Formation Commands", "MUXI")
	printBoxLine("$ muxi formation [command]")
	printBoxLine("  or (from formation dir):")
	printBoxLine("$ muxi [command]")
	printBoxBottom()
	printCommandGroup(cmd, "formation")
	fmt.Println()

	// Registry Commands
	printBoxWithSubtitle("Registry Commands", "MUXI")
	printBoxLine("$ muxi registry [command]")
	printBoxBottom()
	printCommandGroup(cmd, "registry")
	fmt.Println()

	// Server Commands
	printBoxWithSubtitle("Server Commands", "MUXI")
	printBoxLine("$ muxi server [command]")
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
	fmt.Println("  -h, --help       Help for muxi")
	fmt.Println("  -v, --version    Version for muxi")
	fmt.Println()

	// Footer
	printDivider()
	fmt.Println()
	fmt.Println("Use \"muxi [command] --help\" for more information about a command.")
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
	
	// Title with subtitle right-aligned
	content := title
	padding := boxWidth - 4 - len(content) - len(subtitle)
	if padding < 1 {
		padding = 1
	}
	fmt.Printf("│ %s%s%s │\n", content, strings.Repeat(" ", padding), subtitle)
	
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
	fmt.Printf("│ %s%s%s │\n", content, strings.Repeat(" ", padding), subtitle)
	
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

func printCommandGroup(cmd *cobra.Command, groupID string) {
	for _, c := range cmd.Commands() {
		if c.GroupID == groupID && c.Name() != "help" {
			fmt.Printf("  %-12s%s\n", c.Name(), c.Short)
		}
	}
}

func printUngroupedCommands(cmd *cobra.Command) {
	for _, c := range cmd.Commands() {
		if c.GroupID == "" && c.Name() != "help" && c.Name() != "completion" {
			fmt.Printf("  %-12s%s\n", c.Name(), c.Short)
		}
	}
	// Always show these
	fmt.Printf("  %-12s%s\n", "completion", "Generate autocompletion script for your shell")
	fmt.Printf("  %-12s%s\n", "help", "Help about any command")
}

// SetVersionInfo sets version information from main
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = v
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
