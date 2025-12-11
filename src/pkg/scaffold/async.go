package scaffold

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
	"gopkg.in/yaml.v3"
)

// ConfigureAsync runs the async configuration wizard
func ConfigureAsync() error {
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
│ [⚙] Configure Async Responses                           MUXI │
│──────────────────────────────────────────────────────────────│
│ Configure how long-running tasks are handled. When a task    │
│ exceeds the threshold, the response is delivered via webhook.│
╰──────────────────────────────────────────────────────────────╯`)

	// Check current async config
	currentConfig := getAsyncConfig(ctx.RootDir)

	// Show options menu
	options := []wizard.SelectOption{
		{Value: "configure", Label: "Configure async settings"},
		{Value: "disable", Label: "Disable async responses"},
	}

	action, err := wizard.PromptSelect("What would you like to do?", options, 0)
	if err != nil {
		return err
	}

	switch action {
	case "configure":
		return configureAsyncSettings(ctx.RootDir, currentConfig)
	case "disable":
		if currentConfig == nil {
			fmt.Println()
			ui.Dimmed("  Async responses are not configured")
			return nil
		}
		return disableAsync(ctx.RootDir)
	}

	return nil
}

// AsyncConfig holds the async configuration
type AsyncConfig struct {
	ThresholdSeconds  int
	WebhookURL        string
	RetryCount        int
	RetryDelaySeconds int
}

// getAsyncConfig reads current async config from formation.yaml
func getAsyncConfig(rootDir string) *AsyncConfig {
	formationPath, _ := context.FindFormationFile(rootDir)
	data, err := os.ReadFile(formationPath)
	if err != nil {
		return nil
	}

	var formation map[string]interface{}
	if err := yaml.Unmarshal(data, &formation); err != nil {
		return nil
	}

	asyncRaw, ok := formation["async"].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &AsyncConfig{
		ThresholdSeconds:  30,
		RetryCount:        3,
		RetryDelaySeconds: 5,
	}

	if threshold, ok := asyncRaw["threshold_seconds"].(int); ok {
		config.ThresholdSeconds = threshold
	}
	if url, ok := asyncRaw["webhook_url"].(string); ok {
		config.WebhookURL = url
	}
	if retry, ok := asyncRaw["retry"].(map[string]interface{}); ok {
		if count, ok := retry["count"].(int); ok {
			config.RetryCount = count
		}
		if delay, ok := retry["delay_seconds"].(int); ok {
			config.RetryDelaySeconds = delay
		}
	}

	return config
}

// configureAsyncSettings handles the async configuration flow
func configureAsyncSettings(rootDir string, current *AsyncConfig) error {
	ui.Bold("Async Response Configuration")
	fmt.Println()

	// Set defaults
	if current == nil {
		current = &AsyncConfig{
			ThresholdSeconds:  30,
			RetryCount:        3,
			RetryDelaySeconds: 5,
		}
	}

	// Threshold
	ui.Dimmed("  Switch to async mode if processing time exceeds this threshold")
	thresholdOptions := []wizard.SelectOption{
		{Value: "15", Label: "15 seconds"},
		{Value: "30", Label: "30 seconds"},
		{Value: "60", Label: "60 seconds"},
		{Value: "120", Label: "120 seconds"},
		{Value: "custom", Label: "Custom"},
	}

	// Mark current value
	currentThresholdIdx := 1 // default to 30
	for i, opt := range thresholdOptions {
		if opt.Value == strconv.Itoa(current.ThresholdSeconds) {
			thresholdOptions[i].Label += " [current]"
			currentThresholdIdx = i
			break
		}
	}

	thresholdChoice, err := wizard.PromptSelect("  Threshold", thresholdOptions, currentThresholdIdx)
	if err != nil {
		return err
	}

	var threshold int
	if thresholdChoice == "custom" {
		for {
			input, err := wizard.PromptString("  Threshold (seconds)", strconv.Itoa(current.ThresholdSeconds), nil)
			if err != nil {
				return err
			}
			threshold, err = strconv.Atoi(input)
			if err != nil || threshold < 1 || threshold > 3600 {
				ui.PromptError("  Threshold", input, fmt.Errorf("must be 1-3600 seconds"))
				continue
			}
			break
		}
	} else {
		threshold, _ = strconv.Atoi(thresholdChoice)
	}
	ui.PromptSuccess("  Threshold", fmt.Sprintf("%d seconds", threshold))

	// Webhook URL
	fmt.Println()
	ui.Dimmed("  Webhook URL for async response delivery")
	var webhookURL string
	for {
		defaultURL := current.WebhookURL
		input, err := wizard.PromptString("  Webhook URL", defaultURL, nil)
		if err != nil {
			return err
		}
		if input == "" {
			ui.PromptError("  Webhook URL", "", fmt.Errorf("webhook URL is required for async responses"))
			continue
		}
		webhookURL = normalizeURL(input)
		if err := validateURL(webhookURL); err != nil {
			ui.PromptError("  Webhook URL", input, err)
			continue
		}
		break
	}
	ui.PromptSuccess("  Webhook URL", webhookURL)

	// Retry count
	fmt.Println()
	ui.Dimmed("  Number of delivery retry attempts (0 to disable)")
	retryOptions := []wizard.SelectOption{
		{Value: "0", Label: "0 (no retries)"},
		{Value: "3", Label: "3 retries"},
		{Value: "5", Label: "5 retries"},
		{Value: "10", Label: "10 retries"},
	}

	currentRetryIdx := 1 // default to 3
	for i, opt := range retryOptions {
		if opt.Value == strconv.Itoa(current.RetryCount) {
			retryOptions[i].Label += " [current]"
			currentRetryIdx = i
			break
		}
	}

	retryChoice, err := wizard.PromptSelect("  Retry count", retryOptions, currentRetryIdx)
	if err != nil {
		return err
	}
	retryCount, _ := strconv.Atoi(retryChoice)
	ui.PromptSuccess("  Retry count", retryChoice)

	// Retry delay (only if retries enabled)
	retryDelay := 5
	if retryCount > 0 {
		fmt.Println()
		ui.Dimmed("  Delay between retry attempts")
		delayOptions := []wizard.SelectOption{
			{Value: "1", Label: "1 second"},
			{Value: "5", Label: "5 seconds"},
			{Value: "10", Label: "10 seconds"},
			{Value: "30", Label: "30 seconds"},
		}

		currentDelayIdx := 1 // default to 5
		for i, opt := range delayOptions {
			if opt.Value == strconv.Itoa(current.RetryDelaySeconds) {
				delayOptions[i].Label += " [current]"
				currentDelayIdx = i
				break
			}
		}

		delayChoice, err := wizard.PromptSelect("  Retry delay", delayOptions, currentDelayIdx)
		if err != nil {
			return err
		}
		retryDelay, _ = strconv.Atoi(delayChoice)
		ui.PromptSuccess("  Retry delay", fmt.Sprintf("%s seconds", delayChoice))
	}

	// Save configuration
	if err := saveAsyncConfig(rootDir, threshold, webhookURL, retryCount, retryDelay); err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Async configuration saved to formation.yaml")
	return nil
}

// disableAsync removes async configuration
func disableAsync(rootDir string) error {
	fmt.Println()
	ui.Bold("Disable Async Responses")
	fmt.Println()

	ui.Warning("This will remove the async configuration from formation.yaml")
	fmt.Println()

	confirm, err := wizard.PromptConfirm("  Disable async responses?", false)
	if err != nil {
		return err
	}

	if !confirm {
		ui.PromptSkipped("  Cancelled")
		return nil
	}

	if err := removeAsyncConfig(rootDir); err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Async responses disabled")
	return nil
}

// saveAsyncConfig writes async config to formation.yaml
func saveAsyncConfig(rootDir string, threshold int, webhookURL string, retryCount, retryDelay int) error {
	formationPath, _ := context.FindFormationFile(rootDir)
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

	// Build async node
	asyncNode := buildAsyncNode(threshold, webhookURL, retryCount, retryDelay)

	// Find or create async section
	found := false
	for i := 0; i < len(docNode.Content); i += 2 {
		if docNode.Content[i].Value == "async" {
			docNode.Content[i+1] = asyncNode
			found = true
			break
		}
	}

	if !found {
		docNode.Content = append(docNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "async"},
			asyncNode,
		)
	}

	// Write back with 2-space indentation
	output, err := marshalYAML(&root)
	if err != nil {
		return fmt.Errorf("failed to marshal formation.yaml: %w", err)
	}

	if err := os.WriteFile(formationPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write formation.yaml: %w", err)
	}

	// Clean up formatting
	content, err := os.ReadFile(formationPath)
	if err == nil {
		cleaned := ensureBlankLineBeforeTopLevel(string(content))
		os.WriteFile(formationPath, []byte(cleaned), 0644)
	}

	return nil
}

// buildAsyncNode creates the async YAML node
func buildAsyncNode(threshold int, webhookURL string, retryCount, retryDelay int) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}

	// threshold_seconds
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "threshold_seconds"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: strconv.Itoa(threshold), Tag: "!!int"},
	)

	// webhook_url
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "webhook_url"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: webhookURL, Style: yaml.DoubleQuotedStyle},
	)

	// retry (only if count > 0)
	if retryCount > 0 {
		retryNode := &yaml.Node{Kind: yaml.MappingNode}
		retryNode.Content = append(retryNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "count"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: strconv.Itoa(retryCount), Tag: "!!int"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "delay_seconds"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: strconv.Itoa(retryDelay), Tag: "!!int"},
		)
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "retry"},
			retryNode,
		)
	}

	return node
}

// removeAsyncConfig removes async section from formation.yaml
func removeAsyncConfig(rootDir string) error {
	formationPath, _ := context.FindFormationFile(rootDir)
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

	// Remove async section
	newContent := make([]*yaml.Node, 0, len(docNode.Content))
	for i := 0; i < len(docNode.Content); i += 2 {
		if docNode.Content[i].Value != "async" {
			newContent = append(newContent, docNode.Content[i], docNode.Content[i+1])
		}
	}
	docNode.Content = newContent

	// Write back
	output, err := marshalYAML(&root)
	if err != nil {
		return fmt.Errorf("failed to marshal formation.yaml: %w", err)
	}

	// Clean up multiple blank lines
	lines := strings.Split(string(output), "\n")
	var result []string
	prevEmpty := false
	for _, line := range lines {
		isEmpty := strings.TrimSpace(line) == ""
		if isEmpty && prevEmpty {
			continue
		}
		result = append(result, line)
		prevEmpty = isEmpty
	}

	if err := os.WriteFile(formationPath, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write formation.yaml: %w", err)
	}

	return nil
}
