package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate autocompletion script for your shell",
	Long: `Generate shell autocompletion script for muxi.

Examples:
  muxi completion zsh              # Show setup instructions
  muxi completion zsh --install    # Auto-install to ~/.zshrc
  muxi completion zsh --show       # Output completion script
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.Flags().Bool("install", false, "Install completion to shell config file")
	completionCmd.Flags().Bool("show", false, "Output completion script to stdout")
}

func runCompletion(cmd *cobra.Command, args []string) error {
	install, _ := cmd.Flags().GetBool("install")
	show, _ := cmd.Flags().GetBool("show")
	shell := args[0]

	if install {
		return installCompletion(shell)
	}

	if show {
		// Output completion script to stdout
		switch shell {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	}

	// Default: show instructions
	return showCompletionInstructions(shell)
}

func showCompletionInstructions(shell string) error {
	fmt.Println()
	ui.Bold(fmt.Sprintf("  Setup %s completion for muxi", shell))
	fmt.Println()

	switch shell {
	case "zsh":
		fmt.Println("  Option 1: Auto-install (recommended)")
		fmt.Println()
		fmt.Println("    muxi completion zsh --install")
		fmt.Println()
		fmt.Println("  Option 2: Manual setup")
		fmt.Println()
		fmt.Println("    # Add to ~/.zshrc:")
		fmt.Println("    source <(muxi completion zsh --show)")
		fmt.Println()
	case "bash":
		fmt.Println("  Option 1: Auto-install (recommended)")
		fmt.Println()
		fmt.Println("    muxi completion bash --install")
		fmt.Println()
		fmt.Println("  Option 2: Manual setup")
		fmt.Println()
		fmt.Println("    # Add to ~/.bashrc or ~/.bash_profile:")
		fmt.Println("    source <(muxi completion bash --show)")
		fmt.Println()
	case "fish":
		fmt.Println("  Option 1: Auto-install (recommended)")
		fmt.Println()
		fmt.Println("    muxi completion fish --install")
		fmt.Println()
		fmt.Println("  Option 2: Manual setup")
		fmt.Println()
		fmt.Println("    muxi completion fish --show > ~/.config/fish/completions/muxi.fish")
		fmt.Println()
	case "powershell":
		fmt.Println("  Add to your PowerShell profile:")
		fmt.Println()
		fmt.Println("    muxi completion powershell --show | Out-String | Invoke-Expression")
		fmt.Println()
	}

	return nil
}

func installCompletion(shell string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var configFile, sourceLine string

	switch shell {
	case "zsh":
		configFile = filepath.Join(home, ".zshrc")
		sourceLine = "source <(muxi completion zsh --show)"
	case "bash":
		// Check for .bash_profile (macOS) or .bashrc (Linux)
		configFile = filepath.Join(home, ".bashrc")
		if _, err := os.Stat(filepath.Join(home, ".bash_profile")); err == nil {
			configFile = filepath.Join(home, ".bash_profile")
		}
		sourceLine = "source <(muxi completion bash --show)"
	case "fish":
		configFile = filepath.Join(home, ".config", "fish", "config.fish")
		sourceLine = "muxi completion fish --show | source"
	case "powershell":
		fmt.Println()
		ui.Dimmed("  PowerShell auto-install not supported.")
		fmt.Println("  Add this to your PowerShell profile:")
		fmt.Println()
		fmt.Println("    muxi completion powershell | Out-String | Invoke-Expression")
		fmt.Println()
		return nil
	}

	// Check if already installed
	if containsLine(configFile, sourceLine) {
		fmt.Println()
		ui.Success(fmt.Sprintf("Completion already installed in %s", configFile))
		fmt.Println()
		return nil
	}

	// Append to config file
	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", configFile, err)
	}
	defer f.Close()

	// Add newline before if file doesn't end with one
	stat, _ := f.Stat()
	if stat.Size() > 0 {
		f.WriteString("\n")
	}

	f.WriteString("# MUXI CLI completion\n")
	f.WriteString(sourceLine + "\n")

	fmt.Println()
	ui.Success(fmt.Sprintf("Installed completion to %s", configFile))
	fmt.Println()
	fmt.Println("  To activate now, run:")
	fmt.Printf("    source %s\n", configFile)
	fmt.Println()

	return nil
}

func containsLine(filename, line string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strings.TrimSpace(line) {
			return true
		}
	}
	return false
}
