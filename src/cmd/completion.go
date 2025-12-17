package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate autocompletion script for your shell",
	Long: `Generate shell autocompletion script for muxi.

To load completions:

Bash:
  $ source <(muxi completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ muxi completion bash > /etc/bash_completion.d/muxi
  # macOS:
  $ muxi completion bash > $(brew --prefix)/etc/bash_completion.d/muxi

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ muxi completion zsh > "${fpath[1]}/_muxi"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ muxi completion fish | source

  # To load completions for each session, execute once:
  $ muxi completion fish > ~/.config/fish/completions/muxi.fish

PowerShell:
  PS> muxi completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> muxi completion powershell > muxi.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
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
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
