package scaffold

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	ui.Dimmed("  Custom message to show when credentials are needed")
	fmt.Println()

	// Get current redirect message
	currentMessage := getCurrentSecurityValue(rootDir, "redirect_message")
	defaultMessage := "For security, please configure your credentials at <REDIRECT_URL>."

	// Show current or default message
	displayMessage := currentMessage
	if currentMessage == "" {
		ui.Dimmed("  Current (default):")
		displayMessage = defaultMessage
	} else {
		ui.Dimmed("  Current:")
	}
	ui.Dimmed("    \"" + displayMessage + "\"")
	fmt.Println()

	// Prompt for new message
	message, err := wizard.PromptString("  Redirect message [keep current]", "", nil)
	if err != nil {
		return err
	}

	// If empty, keep current/default
	if message == "" {
		message = displayMessage
		ui.PromptSuccess("  Redirect message", "kept current")
	} else {
		ui.PromptSuccess("  Redirect message", truncateForDisplay(message, 50))
	}

	return updateSecurityRedirectInFormation(rootDir, message)
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

// configureDynamicMode handles Flow 2: Dynamic Mode (Development)
func configureDynamicMode(rootDir string) error {
	fmt.Println()
	ui.Bold("Dynamic Mode")
	fmt.Println()
	ui.Warning("  Dynamic credential collection should only be used in development.")
	ui.Dimmed("  User credentials will be collected inline when tools request them.")
	fmt.Println()

	// Get current values
	currentEnvs := getCurrentSecurityValue(rootDir, "allowed_environments")
	currentHTTPS := getCurrentSecurityValue(rootDir, "require_https")
	currentTTL := getCurrentSecurityValue(rootDir, "credential_ttl_minutes")
	currentAttempts := getCurrentSecurityValue(rootDir, "max_attempts")

	// Allowed environments
	ui.Dimmed("  Environments where dynamic mode is allowed (comma-separated)")
	envsDefault := "development, staging"
	if currentEnvs != "" {
		envsDefault = currentEnvs
	}
	envs, err := wizard.PromptString("  Allowed environments", envsDefault, nil)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Allowed environments", envs)

	// Require HTTPS
	ui.Dimmed("  Require HTTPS for credential collection (recommended)")
	httpsDefault := currentHTTPS != "false"
	requireHTTPS, err := wizard.PromptConfirm("  Require HTTPS?", httpsDefault)
	if err != nil {
		return err
	}
	if requireHTTPS {
		ui.PromptSuccess("  Require HTTPS", "enabled")
	} else {
		ui.PromptSkipped("  Require HTTPS")
	}

	// Credential TTL
	ui.Dimmed("  How long to cache collected credentials (minutes)")
	ttlDefault := "60"
	if currentTTL != "" {
		ttlDefault = currentTTL
	}
	ttl, err := wizard.PromptString("  Credential TTL", ttlDefault, validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Credential TTL", ttl+" minutes")

	// Max attempts
	ui.Dimmed("  Maximum failed auth attempts before lockout (1-10)")
	attemptsDefault := "3"
	if currentAttempts != "" {
		attemptsDefault = currentAttempts
	}
	attempts, err := wizard.PromptString("  Max attempts", attemptsDefault, validateMaxAttempts)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Max attempts", attempts)

	// Encryption settings
	fmt.Println()
	ui.Bold("Encryption Settings")
	fmt.Println()
	ui.Dimmed("  User credentials collected in dynamic mode are encrypted at rest using Fernet.")
	ui.Dimmed("  An encryption key will be auto-generated and stored in secrets.")
	fmt.Println()

	// Generate encryption key
	encryptionKey, err := generateFernetKey()
	if err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Save to secrets
	sm := secrets.NewManager(rootDir)
	if err := sm.Set("USER_CREDENTIALS_ENCRYPTION_KEY", encryptionKey, true); err != nil {
		return fmt.Errorf("failed to save encryption key: %w", err)
	}
	ui.PromptSuccess("  Generated", "USER_CREDENTIALS_ENCRYPTION_KEY")

	// Parse environments into array
	envList := parseEnvironmentList(envs)

	// Convert TTL and attempts to int
	ttlInt, _ := strconv.Atoi(ttl)
	attemptsInt, _ := strconv.Atoi(attempts)

	return updateSecurityDynamicInFormation(rootDir, envList, requireHTTPS, ttlInt, attemptsInt)
}

// generateFernetKey generates a 32-byte Fernet-compatible key
func generateFernetKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(key), nil
}

// parseEnvironmentList parses a comma-separated list of environments
func parseEnvironmentList(envs string) []string {
	parts := strings.Split(envs, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// validateMaxAttempts validates max attempts (1-10)
func validateMaxAttempts(input string) error {
	val, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if val < 1 || val > 10 {
		return fmt.Errorf("must be between 1 and 10")
	}
	return nil
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
func updateSecurityRedirectInFormation(rootDir string, message string) error {
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
	userCredsNode := buildRedirectNode(message)

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

	// Write back
	output, err := yaml.Marshal(&root)
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
func buildRedirectNode(message string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}

	// mode
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "mode"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: "redirect"},
	)

	// redirect_message (literal block style for multi-line)
	msgNode := &yaml.Node{Kind: yaml.ScalarNode, Value: message}
	if strings.Contains(message, "\n") {
		msgNode.Style = yaml.LiteralStyle
	}
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "redirect_message"},
		msgNode,
	)

	return node
}

// updateSecurityDynamicInFormation updates formation.yaml with dynamic mode settings
func updateSecurityDynamicInFormation(rootDir string, envs []string, requireHTTPS bool, ttl int, maxAttempts int) error {
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
	userCredsNode := buildDynamicNode(envs, requireHTTPS, ttl, maxAttempts)

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

	// Write back
	output, err := yaml.Marshal(&root)
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
func buildDynamicNode(envs []string, requireHTTPS bool, ttl int, maxAttempts int) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}

	// mode
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "mode"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: "dynamic"},
	)

	// allowed_environments
	envsNode := &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
	for _, env := range envs {
		envsNode.Content = append(envsNode.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: env})
	}
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "allowed_environments"},
		envsNode,
	)

	// require_https
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "require_https"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%t", requireHTTPS)},
	)

	// credential_ttl_minutes
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "credential_ttl_minutes"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", ttl), Tag: "!!int"},
	)

	// max_attempts
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "max_attempts"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", maxAttempts), Tag: "!!int"},
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
