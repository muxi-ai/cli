package formation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/defaults"
	"github.com/muxi-ai/cli/pkg/secrets"
	"github.com/muxi-ai/cli/pkg/server"
	"gopkg.in/yaml.v3"
)

const (
	AdminKeySecret  = "FORMATION_ADMIN_API_KEY"
	ClientKeySecret = "FORMATION_CLIENT_API_KEY"
)

// ResolveKeys attempts to get API keys from various sources
// Priority: 1. Formation secrets.enc, 2. Saved formations config
func ResolveKeys(formationDir string) (adminKey, clientKey string, err error) {
	// 1. Try to load from formation's secrets.enc
	if formationDir != "" {
		adminKey, clientKey, err = loadKeysFromSecrets(formationDir)
		if err == nil && (adminKey != "" || clientKey != "") {
			return adminKey, clientKey, nil
		}
	}

	// 2. Try current directory if no formationDir specified
	if formationDir == "" {
		ctx, err := context.DetectFormation()
		if err == nil {
			adminKey, clientKey, _ = loadKeysFromSecrets(ctx.RootDir)
			if adminKey != "" || clientKey != "" {
				return adminKey, clientKey, nil
			}
		}
	}

	return "", "", fmt.Errorf("no API keys found - configure in secrets.enc or use 'muxi formations add'")
}

// ResolveKeysFromSaved attempts to get API keys from saved formations config
func ResolveKeysFromSaved(formationID string) (adminKey, clientKey string, err error) {
	entry, err := defaults.GetFormation(formationID)
	if err != nil {
		return "", "", err
	}
	return entry.AdminKey, entry.ClientKey, nil
}

// loadKeysFromSecrets loads API keys from a formation's secrets.enc file
func loadKeysFromSecrets(formationDir string) (adminKey, clientKey string, err error) {
	secretsPath := filepath.Join(formationDir, "secrets.enc")
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("secrets.enc not found")
	}

	mgr := secrets.NewManager(formationDir)
	if err := mgr.Initialize(); err != nil {
		return "", "", fmt.Errorf("failed to load secrets: %w", err)
	}

	adminKey, _ = mgr.Get(AdminKeySecret)
	clientKey, _ = mgr.Get(ClientKeySecret)

	return adminKey, clientKey, nil
}

// BuildFormationURL constructs the Formation API base URL
// serverURL: e.g., "http://localhost:7890"
// formationID: e.g., "my-formation"
// draft: if true, uses /draft/ prefix instead of /api/
// Returns: "http://localhost:7890/api/my-formation/v1" or "http://localhost:7890/draft/my-formation/v1"
func BuildFormationURL(serverURL, formationID string, draft bool) string {
	prefix := "api"
	if draft {
		prefix = "draft"
	}
	return fmt.Sprintf("%s/%s/%s/v1", serverURL, prefix, formationID)
}

// ResolveDraftMode resolves whether to use draft mode
// Priority: 1. --draft flag, 2. .muxi file
func ResolveDraftMode(flagValue bool) bool {
	if flagValue {
		return true
	}

	if ctx, err := context.DetectFormation(); err == nil {
		if dotMuxi, err := LoadDotMuxi(ctx.RootDir); err == nil && dotMuxi.Draft {
			return true
		}
	}

	return false
}

// NewClientFromContext creates a Formation API client using context detection
// It resolves: server profile, formation ID, API keys, and draft mode
// When a formation ID is provided and not in a formation directory, it checks
// saved formations in ~/.muxi/cli/formations.yaml
func NewClientFromContext(profile, formationID string, draft bool) (*Client, error) {
	// Check if we're in a formation directory
	ctx, ctxErr := context.DetectFormation()
	inFormationDir := ctxErr == nil

	// Resolve formation ID
	if formationID == "" {
		if !inFormationDir {
			return nil, fmt.Errorf("formation ID required - use -f flag or run from formation directory")
		}
		formationID = ctx.ID
	}

	// Try to use saved formation config when not in formation directory
	if !inFormationDir && formationID != "" {
		return newClientFromSavedFormation(profile, formationID, draft)
	}

	// In formation directory - use secrets.enc for keys
	profileEntry, err := server.GetProfile(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	adminKey, clientKey, err := ResolveKeys(ctx.RootDir)
	if err != nil {
		return nil, err
	}

	baseURL := BuildFormationURL(profileEntry.URL, formationID, draft)
	return NewClient(baseURL, adminKey, clientKey), nil
}

// newClientFromSavedFormation creates a client using saved formation config
func newClientFromSavedFormation(profileOverride, formationID string, draft bool) (*Client, error) {
	// Get saved formation entry
	entry, err := defaults.GetFormation(formationID)
	if err != nil {
		return nil, fmt.Errorf("formation '%s' not found - use 'muxi formations add %s' to configure", formationID, formationID)
	}

	// Use profile override if provided, otherwise use saved default
	profileName := profileOverride
	if profileName == "" {
		profileName = entry.DefaultProfile
	}

	profileEntry, err := server.GetProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile '%s': %w", profileName, err)
	}

	if entry.ClientKey == "" {
		return nil, fmt.Errorf("formation '%s' has no client key configured", formationID)
	}

	baseURL := BuildFormationURL(profileEntry.URL, formationID, draft)
	return NewClient(baseURL, entry.AdminKey, entry.ClientKey), nil
}

// UsesPostgres checks if the formation uses PostgreSQL for persistent storage
func UsesPostgres(formationDir string) bool {
	formationFile, found := context.FindFormationFile(formationDir)
	if !found {
		return false
	}
	data, err := os.ReadFile(formationFile)
	if err != nil {
		return false
	}
	content := string(data)

	// Look for postgres in connection_string under persistent: section
	inPersistent := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "persistent:" {
			inPersistent = true
			continue
		}
		if inPersistent && len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		if inPersistent && strings.HasPrefix(trimmed, "connection_string:") {
			return strings.Contains(trimmed, "postgres")
		}
	}
	return false
}

// DotMuxi represents the .muxi file for formation-level settings
type DotMuxi struct {
	Profile  string `yaml:"profile,omitempty"`
	Registry string `yaml:"registry,omitempty"`
	UserID   string `yaml:"user_id,omitempty"`
	Draft    bool   `yaml:"draft,omitempty"`
}

// LoadDotMuxi loads the .muxi file from a directory
func LoadDotMuxi(dir string) (*DotMuxi, error) {
	path := filepath.Join(dir, ".muxi")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &DotMuxi{}, nil
		}
		return nil, err
	}

	var config DotMuxi
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveDotMuxi saves the .muxi file to a directory
func SaveDotMuxi(dir string, config *DotMuxi) error {
	path := filepath.Join(dir, ".muxi")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal .muxi: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ResolveProfile resolves the server profile to use
// Priority: 1. Flag value, 2. .muxi file, 3. Global default
func ResolveProfile(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	// Check .muxi in formation directory
	if ctx, err := context.DetectFormation(); err == nil {
		if dotMuxi, err := LoadDotMuxi(ctx.RootDir); err == nil && dotMuxi.Profile != "" {
			return dotMuxi.Profile
		}
	}

	// Fall back to global default
	return server.GetDefaultProfile()
}

// ResolveFormationID resolves the formation ID to use
// Priority: 1. Flag value, 2. Formation context (formation.yaml)
func ResolveFormationID(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	ctx, err := context.DetectFormation()
	if err != nil {
		return "", fmt.Errorf("formation ID required - use -f flag or run from formation directory")
	}

	if ctx.ID == "" {
		return "", fmt.Errorf("formation.yaml missing 'id' field")
	}

	return ctx.ID, nil
}
