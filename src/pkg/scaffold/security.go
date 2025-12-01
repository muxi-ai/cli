package scaffold

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/secrets"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
	"gopkg.in/yaml.v3"
)

// ConfigureSecurity runs the security configuration wizard
func ConfigureSecurity() error {
	// Must be in formation directory
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation\n\nOr create a new one:\n  muxi new formation",
		)
		os.Exit(1)
	}

	// Show banner
	ui.Banner(`╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure Security                                  MUXI │
│──────────────────────────────────────────────────────────────│
│ Configure how the formation handles user credentials when    │
│ MCP tools or services request authentication.                │
╰──────────────────────────────────────────────────────────────╯`)

	// Get current mode
	currentMode := getCurrentSecurityValue(ctx.RootDir, "mode")

	// Step 1: Credential mode selection
	ui.Bold("User Credentials")
	fmt.Println()
	ui.Dimmed("  Controls how the formation handles credential requests from tools.")
	fmt.Println()

	modeOptions := []wizard.SelectOption{
		{Value: "redirect", Label: addCurrentIndicator("Redirect (always redirect to external credential management)", currentMode == "redirect" || currentMode == "")},
		{Value: "dynamic", Label: addCurrentIndicator("Dynamic (collect credentials inline when safe)", currentMode == "dynamic")},
	}

	defaultModeIdx := 0
	if currentMode == "dynamic" {
		defaultModeIdx = 1
	}

	mode, err := wizard.PromptSelect("  Credential handling mode", modeOptions, defaultModeIdx)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Mode", mode)

	switch mode {
	case "redirect":
		return configureRedirectMode(ctx.RootDir)
	case "dynamic":
		return configureDynamicMode(ctx.RootDir)
	}

	return nil
}

// configureRedirectMode handles Flow 1: Redirect Mode (Production)
func configureRedirectMode(rootDir string) error {
	fmt.Println()
	ui.Bold("Redirect Mode")
	fmt.Println()
	ui.Dimmed("  Users will be redirected to configure credentials on an external system,")
	ui.Dimmed("  where you can collect credentials and store them securely using the SDKs.")
	ui.Dimmed("  This is the recommended mode for production deployments.")
	fmt.Println()

	// Get current values
	currentURL := getCurrentSecurityValue(rootDir, "redirect_url")
	currentMessage := getCurrentSecurityValue(rootDir, "redirect_message")

	// Redirect URL
	ui.Dimmed("  URL where users configure their credentials")
	urlDefault := currentURL
	if urlDefault == "" {
		urlDefault = "https://yourcompany.com/manage/credentials"
	}
	var redirectURL string
	for {
		input, err := wizard.PromptString("  Redirect URL", urlDefault, nil)
		if err != nil {
			return err
		}
		redirectURL = normalizeURL(input)
		if err := validateURL(redirectURL); err != nil {
			ui.PromptError("  Redirect URL", input, err)
			continue
		}
		break
	}
	ui.PromptSuccess("  Redirect URL", redirectURL)

	// Custom message (optional)
	ui.Dimmed("  Custom message to show when credentials are needed (optional)")
	fmt.Println()

	if currentMessage == "" {
		ui.Dimmed("  Current (default):")
		ui.Dimmed("    Runtime generates: \"For security, please configure your credentials at <URL>.\"")
	} else {
		ui.Dimmed("  Current:")
		ui.Dimmed("    \"" + currentMessage + "\"")
	}
	fmt.Println()

	message, err := wizard.PromptString("  Redirect message [keep current]", "", nil)
	if err != nil {
		return err
	}

	if message == "" {
		if currentMessage == "" {
			ui.PromptSuccess("  Redirect message", "using default")
		} else {
			message = currentMessage
			ui.PromptSuccess("  Redirect message", "kept current")
		}
	} else {
		ui.PromptSuccess("  Redirect message", truncateForDisplay(message, 50))
	}

	return updateSecurityRedirectInFormation(rootDir, redirectURL, message)
}

// truncateForDisplay truncates a string for display purposes
func truncateForDisplay(s string, maxLen int) string {
	// Replace newlines with spaces for display
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// normalizeURL adds https:// if no protocol is specified
func normalizeURL(u string) string {
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return "https://" + u
	}
	return u
}

// validateURL checks if a URL is valid
func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}
	if parsed.Host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}
	// Check for valid host (must have at least one dot and valid TLD-like structure)
	if !strings.Contains(parsed.Host, ".") || strings.HasSuffix(parsed.Host, ".") {
		return fmt.Errorf("invalid URL: invalid host")
	}
	return nil
}

// configureDynamicMode handles Flow 2: Dynamic Mode (Development)
func configureDynamicMode(rootDir string) error {
	fmt.Println()
	ui.Bold("Dynamic Mode")
	fmt.Println()
	ui.Warning("  Dynamic credential collection should only be used in development.")
	ui.Dimmed("  User credentials will be collected inline when tools request them.")
	fmt.Println()

	// Get current URL
	currentURL := getCurrentSecurityValue(rootDir, "redirect_url")

	// Redirect URL (for fallback)
	ui.Dimmed("  Fallback URL when dynamic mode is not available")
	urlDefault := currentURL
	if urlDefault == "" {
		urlDefault = "https://example.com/credentials"
	}
	var redirectURL string
	for {
		input, err := wizard.PromptString("  Redirect URL", urlDefault, nil)
		if err != nil {
			return err
		}
		redirectURL = normalizeURL(input)
		if err := validateURL(redirectURL); err != nil {
			ui.PromptError("  Redirect URL", input, err)
			continue
		}
		break
	}
	ui.PromptSuccess("  Redirect URL", redirectURL)

	// Encryption settings
	fmt.Println()
	ui.Bold("Encryption Settings")
	fmt.Println()
	ui.Dimmed("  User credentials collected in dynamic mode are encrypted at rest using Fernet.")
	fmt.Println()

	// Prompt for encryption key
	ui.Dimmed("  Enter your Fernet key or leave empty to auto-generate one")
	encryptionKey, err := wizard.PromptPassword("  Encryption key [auto-generate]", true)
	if err != nil {
		return err
	}

	if encryptionKey == "" {
		// Auto-generate
		encryptionKey, err = generateFernetKey()
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		ui.PromptSuccess("  Encryption key", "auto-generated")
	} else {
		ui.PromptSuccess("  Encryption key", "provided")
	}

	// Save to secrets
	sm := secrets.NewManager(rootDir)
	if err := sm.Set("USER_CREDENTIALS_ENCRYPTION_KEY", encryptionKey, true); err != nil {
		return fmt.Errorf("failed to save encryption key: %w", err)
	}
	ui.PromptSuccess("  Saved", "USER_CREDENTIALS_ENCRYPTION_KEY")

	return updateSecurityDynamicInFormation(rootDir, redirectURL)
}

// generateFernetKey generates a 32-byte Fernet-compatible key
func generateFernetKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(key), nil
}

// getCurrentSecurityValue gets a value from user_credentials section
func getCurrentSecurityValue(rootDir string, key string) string {
	formationPath := filepath.Join(rootDir, "formation.yaml")
	data, err := os.ReadFile(formationPath)
	if err != nil {
		return ""
	}

	var formation map[string]interface{}
	if err := yaml.Unmarshal(data, &formation); err != nil {
		return ""
	}

	userCreds, ok := formation["user_credentials"].(map[string]interface{})
	if !ok {
		return ""
	}

	// Handle allowed_environments specially (it's an array)
	if key == "allowed_environments" {
		if envs, ok := userCreds[key].([]interface{}); ok {
			strs := make([]string, len(envs))
			for i, e := range envs {
				strs[i] = fmt.Sprintf("%v", e)
			}
			return strings.Join(strs, ", ")
		}
		return ""
	}

	if val, ok := userCreds[key]; ok {
		return fmt.Sprintf("%v", val)
	}

	return ""
}

// updateSecurityRedirectInFormation updates formation.yaml with redirect mode settings
func updateSecurityRedirectInFormation(rootDir string, redirectURL string, message string) error {
	formationPath := filepath.Join(rootDir, "formation.yaml")
	data, err := os.ReadFile(formationPath)
	if err != nil {
		return fmt.Errorf("failed to read formation.yaml: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse formation.yaml: %w", err)
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return fmt.Errorf("invalid formation.yaml structure")
	}

	docNode := root.Content[0]
	if docNode.Kind != yaml.MappingNode {
		return fmt.Errorf("formation.yaml root must be a mapping")
	}

	// Build user_credentials node
	userCredsNode := buildRedirectNode(redirectURL, message)

	// Find or create user_credentials
	found := false
	for i := 0; i < len(docNode.Content); i += 2 {
		if docNode.Content[i].Value == "user_credentials" {
			docNode.Content[i+1] = userCredsNode
			found = true
			break
		}
	}

	if !found {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "user_credentials"}
		docNode.Content = append(docNode.Content, keyNode, userCredsNode)
	}

	// Write back with 2-space indentation
	output, err := marshalYAML(&root)
	if err != nil {
		return fmt.Errorf("failed to marshal formation.yaml: %w", err)
	}

	if err := os.WriteFile(formationPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write formation.yaml: %w", err)
	}

	fmt.Println()
	ui.Success("Security configuration saved to formation.yaml")
	return nil
}

// buildRedirectNode creates the user_credentials node for redirect mode
func buildRedirectNode(redirectURL string, message string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}

	// mode
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "mode"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: "redirect"},
	)

	// redirect_url
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "redirect_url"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: redirectURL},
	)

	// redirect_message (only if custom message provided)
	if message != "" {
		msgNode := &yaml.Node{Kind: yaml.ScalarNode, Value: message}
		if strings.Contains(message, "\n") {
			msgNode.Style = yaml.LiteralStyle
		}
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "redirect_message"},
			msgNode,
		)
	}

	return node
}

// updateSecurityDynamicInFormation updates formation.yaml with dynamic mode settings
func updateSecurityDynamicInFormation(rootDir string, redirectURL string) error {
	formationPath := filepath.Join(rootDir, "formation.yaml")
	data, err := os.ReadFile(formationPath)
	if err != nil {
		return fmt.Errorf("failed to read formation.yaml: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse formation.yaml: %w", err)
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return fmt.Errorf("invalid formation.yaml structure")
	}

	docNode := root.Content[0]
	if docNode.Kind != yaml.MappingNode {
		return fmt.Errorf("formation.yaml root must be a mapping")
	}

	// Build user_credentials node
	userCredsNode := buildDynamicNode(redirectURL)

	// Find or create user_credentials
	found := false
	for i := 0; i < len(docNode.Content); i += 2 {
		if docNode.Content[i].Value == "user_credentials" {
			docNode.Content[i+1] = userCredsNode
			found = true
			break
		}
	}

	if !found {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "user_credentials"}
		docNode.Content = append(docNode.Content, keyNode, userCredsNode)
	}

	// Write back with 2-space indentation
	output, err := marshalYAML(&root)
	if err != nil {
		return fmt.Errorf("failed to marshal formation.yaml: %w", err)
	}

	if err := os.WriteFile(formationPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write formation.yaml: %w", err)
	}

	fmt.Println()
	ui.Success("Security configuration saved to formation.yaml")
	return nil
}

// buildDynamicNode creates the user_credentials node for dynamic mode
func buildDynamicNode(redirectURL string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}

	// mode
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "mode"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: "dynamic"},
	)

	// redirect_url (fallback)
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "redirect_url"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: redirectURL},
	)

	// encryption
	encNode := &yaml.Node{Kind: yaml.MappingNode}
	encNode.Content = append(encNode.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "key"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: "${{ secrets.USER_CREDENTIALS_ENCRYPTION_KEY }}"},
	)
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "encryption"},
		encNode,
	)

	return node
}
