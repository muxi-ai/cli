package cmd

import (
	"github.com/muxi-ai/cli/pkg/chat"
	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:     "chat [message]",
	Short:   "Chat with a formation",
	GroupID: "formation",
	Long: `Start an interactive chat session with a formation.

In interactive mode, you can send messages and receive streaming responses.
Use slash commands like /help, /agents, /exit for navigation.

One-shot mode: Pass a message as argument to get a single response and exit.`,
	Example: `  # Interactive chat
  muxi chat
  muxi chat -s sess_abc123           # Resume session

  # One-shot mode
  muxi chat "What's the weather?"
  echo "Analyze this" | muxi chat`,
	RunE: runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)

	formation.AddCommonFlags(chatCmd)
	chatCmd.Flags().StringP("session", "s", "", "Resume session ID")
	chatCmd.Flags().StringP("group", "g", "", "Agent group for routing")
	chatCmd.Flags().Bool("no-stream", false, "Disable streaming (wait for full response)")
	chatCmd.Flags().Bool("no-splash", false, "Skip welcome banner")
}

func runChat(cmd *cobra.Command, args []string) error {
	// Get flags
	flags := formation.GetCommonFlags(cmd)

	// Resolve formation ID
	formationID, err := formation.ResolveFormationID(flags.FormationID)
	if err != nil {
		// Use dummy for now if not in formation dir
		formationID = "my-formation"
	}

	// Resolve server profile
	profile := formation.ResolveProfile(flags.Profile)
	if profile == "" {
		profile = "local"
	}

	// Resolve user ID
	userID := formation.ResolveUserID(flags.UserID)
	if userID == "" {
		userID = "default-user"
	}

	// TODO: Handle one-shot mode when args provided
	// if len(args) > 0 {
	//     return runOneShotChat(args[0], ...)
	// }

	// Start interactive chat
	cfg := chat.Config{
		FormationID: formationID,
		ServerID:    profile,
		UserID:      userID,
	}

	return chat.Run(cfg)
}
