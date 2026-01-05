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
		fmt.Println("    # Add to ~/.zshrc (caches completion, regenerates on upgrade):")
		fmt.Println(`    if command -v muxi >/dev/null; then
        muxi_cache=~/.muxi/completions/zsh/_muxi
        muxi_bin=$(command -v muxi); muxi_bin=${muxi_bin:A}  # resolve symlinks
        [[ -d ${muxi_cache:h} ]] || mkdir -p ${muxi_cache:h}
        [[ ! -f $muxi_cache || $muxi_bin -nt $muxi_cache ]] && muxi completion zsh --show >| $muxi_cache
        source $muxi_cache
    fi`)
		fmt.Println()
	case "bash":
		fmt.Println("  Option 1: Auto-install (recommended)")
		fmt.Println()
		fmt.Println("    muxi completion bash --install")
		fmt.Println()
		fmt.Println("  Option 2: Manual setup")
		fmt.Println()
		fmt.Println("    # Add to ~/.bashrc (caches completion, regenerates on upgrade):")
		fmt.Println(`    if command -v muxi >/dev/null; then
        muxi_cache=~/.muxi/completions/bash/muxi
        muxi_bin=$(realpath "$(command -v muxi)" 2>/dev/null || readlink -f "$(command -v muxi)" 2>/dev/null || command -v muxi)
        mkdir -p "$(dirname "$muxi_cache")"
        [[ ! -f "$muxi_cache" || "$muxi_bin" -nt "$muxi_cache" ]] && muxi completion bash --show > "$muxi_cache"
        source "$muxi_cache"
    fi`)
		fmt.Println()
	case "fish":
		fmt.Println("  Option 1: Auto-install (recommended)")
		fmt.Println()
		fmt.Println("    muxi completion fish --install")
		fmt.Println()
		fmt.Println("  Option 2: Manual setup")
		fmt.Println()
		fmt.Println("    # Run once (fish auto-loads from completions dir):")
		fmt.Println("    muxi completion fish --show > ~/.config/fish/completions/muxi.fish")
		fmt.Println()
	case "powershell":
		fmt.Println("  Add to your PowerShell profile ($PROFILE):")
		fmt.Println()
		fmt.Println(`    $muxiCache = "$env:USERPROFILE\.muxi\completions\powershell\muxi.ps1"
    if (Get-Command muxi -ErrorAction SilentlyContinue) {
        $muxiBin = (Get-Command muxi).Source
        while ((Get-Item $muxiBin).LinkType) { $muxiBin = (Get-Item $muxiBin).Target }  # resolve symlinks
        if (!(Test-Path $muxiCache) -or ((Get-Item $muxiBin).LastWriteTime -gt (Get-Item $muxiCache).LastWriteTime)) {
            New-Item -ItemType Directory -Path (Split-Path $muxiCache) -Force | Out-Null
            muxi completion powershell --show | Out-File $muxiCache
        }
        . $muxiCache
    }`)
		fmt.Println()
	}

	return nil
}

func installCompletion(shell string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var configFile, completionBlock, marker string

	switch shell {
	case "zsh":
		configFile = filepath.Join(home, ".zshrc")
		marker = "# MUXI CLI completion"
		completionBlock = `# MUXI CLI completion
if command -v muxi >/dev/null; then
    muxi_cache=~/.muxi/completions/zsh/_muxi
    muxi_bin=$(command -v muxi); muxi_bin=${muxi_bin:A}  # resolve symlinks
    [[ -d ${muxi_cache:h} ]] || mkdir -p ${muxi_cache:h}
    [[ ! -f $muxi_cache || $muxi_bin -nt $muxi_cache ]] && muxi completion zsh --show >| $muxi_cache
    source $muxi_cache
fi`
	case "bash":
		// Check for .bash_profile (macOS) or .bashrc (Linux)
		configFile = filepath.Join(home, ".bashrc")
		if _, err := os.Stat(filepath.Join(home, ".bash_profile")); err == nil {
			configFile = filepath.Join(home, ".bash_profile")
		}
		marker = "# MUXI CLI completion"
		completionBlock = `# MUXI CLI completion
if command -v muxi >/dev/null; then
    muxi_cache=~/.muxi/completions/bash/muxi
    muxi_bin=$(realpath "$(command -v muxi)" 2>/dev/null || readlink -f "$(command -v muxi)" 2>/dev/null || command -v muxi)
    mkdir -p "$(dirname "$muxi_cache")"
    [[ ! -f "$muxi_cache" || "$muxi_bin" -nt "$muxi_cache" ]] && muxi completion bash --show > "$muxi_cache"
    source "$muxi_cache"
fi`
	case "fish":
		// Fish uses a completions directory, just write directly
		completionsDir := filepath.Join(home, ".config", "fish", "completions")
		completionFile := filepath.Join(completionsDir, "muxi.fish")
		
		if err := os.MkdirAll(completionsDir, 0755); err != nil {
			return fmt.Errorf("failed to create completions directory: %w", err)
		}
		
		f, err := os.Create(completionFile)
		if err != nil {
			return fmt.Errorf("failed to create completion file: %w", err)
		}
		defer f.Close()
		
		if err := rootCmd.GenFishCompletion(f, true); err != nil {
			return fmt.Errorf("failed to generate fish completion: %w", err)
		}
		
		fmt.Println()
		ui.Success(fmt.Sprintf("Installed completion to %s", completionFile))
		fmt.Println()
		fmt.Println("  Fish will auto-load completions on next shell start.")
		fmt.Println()
		return nil
	case "powershell":
		fmt.Println()
		ui.Dimmed("  PowerShell auto-install not supported.")
		fmt.Println()
		fmt.Println("  Add this to your PowerShell profile ($PROFILE):")
		fmt.Println()
		fmt.Println(`    $muxiCache = "$env:USERPROFILE\.muxi\completions\powershell\muxi.ps1"
    if (Get-Command muxi -ErrorAction SilentlyContinue) {
        $muxiBin = (Get-Command muxi).Source
        while ((Get-Item $muxiBin).LinkType) { $muxiBin = (Get-Item $muxiBin).Target }  # resolve symlinks
        if (!(Test-Path $muxiCache) -or ((Get-Item $muxiBin).LastWriteTime -gt (Get-Item $muxiCache).LastWriteTime)) {
            New-Item -ItemType Directory -Path (Split-Path $muxiCache) -Force | Out-Null
            muxi completion powershell --show | Out-File $muxiCache
        }
        . $muxiCache
    }`)
		fmt.Println()
		return nil
	}

	// Check if already installed (look for marker comment)
	if containsLine(configFile, marker) {
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

	f.WriteString(completionBlock + "\n")

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
