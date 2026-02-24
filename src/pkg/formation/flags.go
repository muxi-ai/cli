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
	Draft       bool
}

// AddCommonFlags adds -f, -p, -u, --draft flags to a command
func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("formation", "f", "", "Formation ID (default: from formation.yaml)")
	cmd.Flags().StringP("profile", "p", "", "Server profile (default: from .muxi or global)")
	cmd.Flags().StringP("user", "u", "", "User ID for user-scoped operations")
	cmd.Flags().Bool("draft", false, "Use draft mode (local dev via muxi up)")
}

// AddFormationFlag adds just the -f flag
func AddFormationFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("formation", "f", "", "Formation ID (default: from formation.yaml)")
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
	draft, _ := cmd.Flags().GetBool("draft")

	return CommonFlags{
		FormationID: formationID,
		Profile:     profile,
		UserID:      userID,
		Draft:       draft,
	}
}

// ResolveUserID resolves the effective user ID
// Priority: 1. Flag value, 2. .muxi file, 3. Global default, 4. Draft → "tester", 5. No postgres → "default"
func ResolveUserID(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	// Check .muxi in formation directory
	ctx, ctxErr := context.DetectFormation()
	if ctxErr == nil {
		if dotMuxi, err := LoadDotMuxi(ctx.RootDir); err == nil {
			if dotMuxi.UserID != "" {
				return dotMuxi.UserID
			}
		}
	}

	// Check global default
	if globalUID := defaults.GetUserID(); globalUID != "" {
		return globalUID
	}

	// Auto-default for draft mode
	if ctxErr == nil {
		if dotMuxi, err := LoadDotMuxi(ctx.RootDir); err == nil && dotMuxi.Draft {
			return "tester"
		}
	}

	// Auto-default for single-user formations (no postgres)
	if ctxErr == nil && !UsesPostgres(ctx.RootDir) {
		return "default"
	}

	return ""
}

// ResolveUserIDWithFormation resolves user ID considering saved formation config
// Priority: 1. Flag value, 2. .muxi file, 3. Saved formation default, 4. Global default, 5. Draft → "tester", 6. No postgres → "default"
func ResolveUserIDWithFormation(flagValue, formationID string) string {
	if flagValue != "" {
		return flagValue
	}

	// Check .muxi in formation directory
	ctx, ctxErr := context.DetectFormation()
	if ctxErr == nil {
		if dotMuxi, err := LoadDotMuxi(ctx.RootDir); err == nil && dotMuxi.UserID != "" {
			return dotMuxi.UserID
		}
	}

	// Check saved formation config
	if formationID != "" {
		if entry, err := defaults.GetFormation(formationID); err == nil && entry.DefaultUserID != "" {
			return entry.DefaultUserID
		}
	}

	// Check global default
	if globalUID := defaults.GetUserID(); globalUID != "" {
		return globalUID
	}

	// Auto-default for draft mode
	if ctxErr == nil {
		if dotMuxi, err := LoadDotMuxi(ctx.RootDir); err == nil && dotMuxi.Draft {
			return "tester"
		}
	}

	// Auto-default for single-user formations (no postgres)
	if ctxErr == nil && !UsesPostgres(ctx.RootDir) {
		return "default"
	}

	return ""
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

	draft := ResolveDraftMode(flags.Draft)
	return NewClientFromContext(profile, formationID, draft)
}

// ClientAndUserFromFlags creates a client and resolves user ID from flags
func ClientAndUserFromFlags(cmd *cobra.Command) (*Client, string, error) {
	flags := GetCommonFlags(cmd)

	client, err := ClientFromFlags(cmd)
	if err != nil {
		return nil, "", err
	}

	// Use formation-aware user ID resolution
	formationID, _ := ResolveFormationID(flags.FormationID)
	userID := ResolveUserIDWithFormation(flags.UserID, formationID)
	if userID == "" {
		return nil, "", &UserIDRequiredError{}
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

	fmt.Println(dim(" ╭" + strings.Repeat("─", visualWidth) + "╮"))
	fmt.Println(dim(" │") + content + dim("│"))
	fmt.Println(dim(" ╰" + strings.Repeat("─", visualWidth) + "╯"))
}

// PrintBadgeFromFlags prints the badge using resolved flag values
func PrintBadgeFromFlags(cmd *cobra.Command) {
	flags := GetCommonFlags(cmd)
	formationID, _ := ResolveFormationID(flags.FormationID)
	serverName := ResolveProfile(flags.Profile)
	PrintBadge(formationID, serverName)
}
