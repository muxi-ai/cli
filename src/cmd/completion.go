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

Use --install to automatically add completion to your shell config.

Examples:
  # Auto-install (recommended)
  muxi completion zsh --install
  muxi completion bash --install

  # Manual: output script to stdout
  muxi completion zsh
  muxi completion bash

  # Manual: source in current session
  source <(muxi completion zsh)
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.Flags().Bool("install", false, "Install completion to shell config file")
}

func runCompletion(cmd *cobra.Command, args []string) error {
	install, _ := cmd.Flags().GetBool("install")
	shell := args[0]

	if install {
		return installCompletion(shell)
	}

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

func installCompletion(shell string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var configFile, sourceLine string

	switch shell {
	case "zsh":
		configFile = filepath.Join(home, ".zshrc")
		sourceLine = "source <(muxi completion zsh)"
	case "bash":
		// Check for .bash_profile (macOS) or .bashrc (Linux)
		configFile = filepath.Join(home, ".bashrc")
		if _, err := os.Stat(filepath.Join(home, ".bash_profile")); err == nil {
			configFile = filepath.Join(home, ".bash_profile")
		}
		sourceLine = "source <(muxi completion bash)"
	case "fish":
		configFile = filepath.Join(home, ".config", "fish", "config.fish")
		sourceLine = "muxi completion fish | source"
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
	ui.Dimmed("  Restart your terminal or run: source " + configFile)
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
