package formation

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/defaults"
	"github.com/spf13/cobra"
)

// CommonFlags holds the common flags for Formation API commands
type CommonFlags struct {
	FormationID string
	Profile     string
	UserID      string
}

// AddCommonFlags adds -F, -p, -u flags to a command
func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("formation", "F", "", "Formation ID (default: from formation.yaml)")
	cmd.Flags().StringP("profile", "p", "", "Server profile (default: from .muxi or global)")
	cmd.Flags().StringP("user", "u", "", "User ID for user-scoped operations")
}

// AddFormationFlag adds just the -F flag
func AddFormationFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("formation", "F", "", "Formation ID (default: from formation.yaml)")
}

// AddProfileFlag adds just the -p flag
func AddProfileFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("profile", "p", "", "Server profile (default: from .muxi or global)")
}

// AddUserFlag adds just the -u flag
func AddUserFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("user", "u", "", "User ID for user-scoped operations")
}

// GetCommonFlags extracts common flags from a command
func GetCommonFlags(cmd *cobra.Command) CommonFlags {
	formationID, _ := cmd.Flags().GetString("formation")
	profile, _ := cmd.Flags().GetString("profile")
	userID, _ := cmd.Flags().GetString("user")

	return CommonFlags{
		FormationID: formationID,
		Profile:     profile,
		UserID:      userID,
	}
}

// ResolveUserID resolves the effective user ID
// Priority: 1. Flag value, 2. .muxi file, 3. Global default
func ResolveUserID(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	// Check .muxi in formation directory
	var formationUserID string
	if ctx, err := context.DetectFormation(); err == nil {
		if dotMuxi, err := LoadDotMuxi(ctx.RootDir); err == nil {
			formationUserID = dotMuxi.UserID
		}
	}

	return defaults.GetEffectiveUserID(formationUserID)
}

// MustResolveUserID resolves user ID or returns error if not set
func MustResolveUserID(flagValue string) (string, error) {
	userID := ResolveUserID(flagValue)
	if userID == "" {
		return "", &UserIDRequiredError{}
	}
	return userID, nil
}

// UserIDRequiredError indicates user ID is required but not set
type UserIDRequiredError struct{}

func (e *UserIDRequiredError) Error() string {
	return "user ID required - use -u flag or set default with: muxi set default user"
}

// ClientFromFlags creates a Formation API client from command flags
func ClientFromFlags(cmd *cobra.Command) (*Client, error) {
	flags := GetCommonFlags(cmd)
	
	profile := ResolveProfile(flags.Profile)
	formationID, err := ResolveFormationID(flags.FormationID)
	if err != nil {
		return nil, err
	}

	return NewClientFromContext(profile, formationID)
}

// ClientAndUserFromFlags creates a client and resolves user ID from flags
func ClientAndUserFromFlags(cmd *cobra.Command) (*Client, string, error) {
	client, err := ClientFromFlags(cmd)
	if err != nil {
		return nil, "", err
	}

	flags := GetCommonFlags(cmd)
	userID, err := MustResolveUserID(flags.UserID)
	if err != nil {
		return nil, "", err
	}

	return client, userID, nil
}

// PrintBadge prints a compact badge showing formation and server
// ╭────────────────────────────────────╮
// │ ⌬ test-formation │ ⚙︎ server-name │
// ╰────────────────────────────────────╯
func PrintBadge(formationID, serverName string) {
	dim := color.New(color.Faint).SprintFunc()

	// Build the content line (what goes between the │ characters)
	// Format: " ⌬ {formation} │ ⚙︎ {server} "
	content := fmt.Sprintf(" ⌬ %s │ ⚙︎ %s ", formationID, serverName)

	// Calculate visual width using rune count
	// Subtract 1 for the variation selector in ⚙︎ (U+FE0E)
	visualWidth := utf8.RuneCountInString(content) - 1

	fmt.Println(dim("  ╭" + strings.Repeat("─", visualWidth) + "╮"))
	fmt.Println(dim("  │") + content + dim("│"))
	fmt.Println(dim("  ╰" + strings.Repeat("─", visualWidth) + "╯"))
}

// PrintBadgeFromFlags prints the badge using resolved flag values
func PrintBadgeFromFlags(cmd *cobra.Command) {
	flags := GetCommonFlags(cmd)
	formationID, _ := ResolveFormationID(flags.FormationID)
	serverName := ResolveProfile(flags.Profile)
	PrintBadge(formationID, serverName)
}
