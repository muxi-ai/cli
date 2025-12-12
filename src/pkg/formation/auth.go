package formation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/secrets"
	"github.com/muxi-ai/cli/pkg/server"
	"gopkg.in/yaml.v3"
)

const (
	AdminKeySecret  = "FORMATION_ADMIN_API_KEY"
	ClientKeySecret = "FORMATION_CLIENT_API_KEY"
)

// ResolveKeys attempts to get API keys from various sources
// Priority: 1. Environment variables, 2. Formation secrets.enc, 3. Cached keys
func ResolveKeys(formationDir string) (adminKey, clientKey string, err error) {
	// 1. Check environment variables
	adminKey = os.Getenv("MUXI_ADMIN_KEY")
	clientKey = os.Getenv("MUXI_CLIENT_KEY")
	if adminKey != "" || clientKey != "" {
		return adminKey, clientKey, nil
	}

	// 2. Try to load from formation's secrets.enc
	if formationDir != "" {
		adminKey, clientKey, err = loadKeysFromSecrets(formationDir)
		if err == nil && (adminKey != "" || clientKey != "") {
			return adminKey, clientKey, nil
		}
	}

	// 3. Try current directory if no formationDir specified
	if formationDir == "" {
		ctx, err := context.DetectFormation()
		if err == nil {
			adminKey, clientKey, _ = loadKeysFromSecrets(ctx.RootDir)
			if adminKey != "" || clientKey != "" {
				return adminKey, clientKey, nil
			}
		}
	}

	return "", "", fmt.Errorf("no API keys found - set MUXI_ADMIN_KEY/MUXI_CLIENT_KEY or configure in secrets.enc")
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
// Returns: "http://localhost:7890/api/my-formation/v1"
func BuildFormationURL(serverURL, formationID string) string {
	return fmt.Sprintf("%s/api/%s/v1", serverURL, formationID)
}

// NewClientFromContext creates a Formation API client using context detection
// It resolves: server profile, formation ID, and API keys
func NewClientFromContext(profile, formationID string) (*Client, error) {
	// Resolve server profile
	profileEntry, err := server.GetProfile(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Resolve formation ID
	if formationID == "" {
		ctx, err := context.DetectFormation()
		if err != nil {
			return nil, fmt.Errorf("formation ID required: %w", err)
		}
		formationID = ctx.ID
	}

	// Resolve API keys
	var formationDir string
	if ctx, err := context.DetectFormation(); err == nil {
		formationDir = ctx.RootDir
	}
	adminKey, clientKey, err := ResolveKeys(formationDir)
	if err != nil {
		return nil, err
	}

	// Build client
	baseURL := BuildFormationURL(profileEntry.URL, formationID)
	return NewClient(baseURL, adminKey, clientKey), nil
}

// DotMuxi represents the .muxi file for formation-level settings
type DotMuxi struct {
	Profile  string `yaml:"profile,omitempty"`
	Registry string `yaml:"registry,omitempty"`
	UserID   string `yaml:"user_id,omitempty"`
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
		return "", fmt.Errorf("formation ID required - use -F flag or run from formation directory")
	}

	if ctx.ID == "" {
		return "", fmt.Errorf("formation.yaml missing 'id' field")
	}

	return ctx.ID, nil
}
