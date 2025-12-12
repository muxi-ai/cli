package cmd

import (
	"fmt"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:     "users",
	Short:   "Manage user identifiers",
	GroupID: "formation",
	Long: `Manage user identifier mappings.

User identifiers map external IDs (email, phone, etc.) to MUXI user IDs.

Requires client API key.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var usersIdentifiersCmd = &cobra.Command{
	Use:   "identifiers",
	Short: "List user identifiers",
	Long: `List all identifiers for a user.

Shows identifier value, type, and when it was created.`,
	RunE: runUsersIdentifiers,
}

var usersLinkCmd = &cobra.Command{
	Use:   "link <identifier>",
	Short: "Link an identifier to a user",
	Long: `Link an external identifier to a MUXI user.

Examples:
  muxi users link -u alice "alice@company.com"
  muxi users link -u alice "+1234567890" --type phone`,
	Args: RequireArgs(1),
	RunE: runUsersLink,
}

var usersUnlinkCmd = &cobra.Command{
	Use:   "unlink <identifier>",
	Short: "Unlink an identifier",
	Long: `Remove an identifier mapping.

Example:
  muxi users unlink "alice@company.com"`,
	Args: RequireArgs(1),
	RunE: runUsersUnlink,
}

var usersResolveCmd = &cobra.Command{
	Use:   "resolve <identifier>",
	Short: "Resolve identifier to user ID",
	Long: `Look up which MUXI user ID an identifier maps to.

Example:
  muxi users resolve "alice@company.com"`,
	Args: RequireArgs(1),
	RunE: runUsersResolve,
}

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersIdentifiersCmd)
	usersCmd.AddCommand(usersLinkCmd)
	usersCmd.AddCommand(usersUnlinkCmd)
	usersCmd.AddCommand(usersResolveCmd)

	formation.AddCommonFlags(usersCmd)
	formation.AddCommonFlags(usersIdentifiersCmd)
	formation.AddCommonFlags(usersLinkCmd)

	formation.AddFormationFlag(usersUnlinkCmd)
	formation.AddProfileFlag(usersUnlinkCmd)

	formation.AddFormationFlag(usersResolveCmd)
	formation.AddProfileFlag(usersResolveCmd)

	usersLinkCmd.Flags().String("type", "", "Identifier type (email, phone, external_id)")
}

func runUsersIdentifiers(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	resp, err := client.GetUserIdentifiersForUser(userID)
	if err != nil {
		return fmt.Errorf("failed to list identifiers: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	if resp.Count == 0 {
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  No identifiers for user '%s'", userID))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Printf("  Identifiers for user '%s':\n\n", userID)
	fmt.Printf("  %-35s %-15s %s\n", "IDENTIFIER", "TYPE", "CREATED")
	fmt.Printf("  %s\n", strings.Repeat("─", 70))

	for _, id := range resp.Identifiers {
		idType := id.Type
		if idType == "" {
			idType = "-"
		}

		created := "-"
		if !id.CreatedAt.IsZero() {
			created = id.CreatedAt.Format("Jan 2, 2006")
		}

		fmt.Printf("  %-35s %-15s %s\n",
			truncate(id.Identifier, 35),
			idType,
			created)
	}
	fmt.Println()

	return nil
}

func runUsersLink(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	identifier := args[0]
	idType, _ := cmd.Flags().GetString("type")

	err = client.LinkUserIdentifier(identifier, userID, idType)
	if err != nil {
		return fmt.Errorf("failed to link identifier: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Linked '%s' to user '%s'", identifier, userID))
	fmt.Println()

	return nil
}

func runUsersUnlink(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	identifier := args[0]

	err = client.UnlinkUserIdentifier(identifier)
	if err != nil {
		return fmt.Errorf("failed to unlink identifier: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Unlinked '%s'", identifier))
	fmt.Println()

	return nil
}

func runUsersResolve(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	identifier := args[0]

	result, err := client.ResolveUserIdentifier(identifier)
	if err != nil {
		return fmt.Errorf("failed to resolve identifier: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)

	fmt.Println()
	fmt.Printf("  Identifier: %s\n", identifier)
	fmt.Printf("  User ID:    %s\n", result.UserID)
	if result.Type != "" {
		fmt.Printf("  Type:       %s\n", result.Type)
	}
	fmt.Println()

	return nil
}
