package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/secrets"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
	"gopkg.in/yaml.v3"
)

// ConfigureLogging runs the logging configuration wizard
func ConfigureLogging() error {
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
│ [⚙] Configure Logging                                   MUXI │
│──────────────────────────────────────────────────────────────│
│ Configure where logs and events are sent. Supports multiple  │
│ destinations with different formats.                         │
╰──────────────────────────────────────────────────────────────╯`)

	// Step 1: Action selection
	fmt.Println()
	options := []wizard.SelectOption{
		{Value: "add", Label: "Add a new log stream"},
		{Value: "view", Label: "View/edit current streams"},
		{Value: "remove", Label: "Remove a stream"},
	}

	action, err := wizard.PromptSelect("What would you like to do?", options, 0)
	if err != nil {
		return err
	}

	switch action {
	case "add":
		return addLogStream(ctx.RootDir)
	case "view":
		return viewEditStreams(ctx.RootDir)
	case "remove":
		return removeStream(ctx.RootDir)
	}

	return nil
}

// addLogStream handles Flow 1: Add New Stream
func addLogStream(rootDir string) error {
	fmt.Println()
	ui.Bold("Add Log Stream")
	fmt.Println()

	// Select transport type
	transportOptions := []wizard.SelectOption{
		{Value: "stdout", Label: "stdout (console output)"},
		{Value: "file", Label: "file (local file)"},
		{Value: "http", Label: "http (webhook, datadog, splunk, etc.)"},
		{Value: "kafka", Label: "kafka (message queue)"},
	}

	transport, err := wizard.PromptSelect("  Select transport type", transportOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Transport", transport)

	switch transport {
	case "stdout":
		return configureStdoutStream(rootDir)
	case "file":
		return configureFileStream(rootDir)
	case "http":
		return configureHTTPStream(rootDir)
	case "kafka":
		return configureKafkaStream(rootDir)
	}

	return nil
}

// configureStdoutStream configures a stdout log stream
func configureStdoutStream(rootDir string) error {
	fmt.Println()
	ui.Bold("Console Output")
	fmt.Println()

	// Log level
	ui.Dimmed("  Log level (debug shows everything, error shows only errors)")
	levelOptions := []wizard.SelectOption{
		{Value: "debug", Label: "debug"},
		{Value: "info", Label: "info"},
		{Value: "warn", Label: "warn"},
		{Value: "error", Label: "error"},
	}
	level, err := wizard.PromptSelect("  Level", levelOptions, 1) // default to info
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Level", level)

	// Format
	ui.Dimmed("  Output format for log entries")
	formatOptions := []wizard.SelectOption{
		{Value: "jsonl", Label: "jsonl (structured, machine-readable)"},
		{Value: "text", Label: "text (human-readable)"},
	}
	format, err := wizard.PromptSelect("  Format", formatOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Format", format)

	// Build stream config
	stream := map[string]interface{}{
		"transport": "stdout",
		"level":     level,
		"format":    format,
	}

	if err := addStreamToFormation(rootDir, stream); err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Added stdout stream")
	return nil
}

// configureFileStream configures a file log stream
func configureFileStream(rootDir string) error {
	fmt.Println()
	ui.Bold("File Output")
	fmt.Println()

	// Path
	ui.Dimmed("  Path to log file (will be created if doesn't exist)")
	path, err := wizard.PromptString("  Path", "/var/log/formation.log", nil)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Path", path)

	// Log level
	ui.Dimmed("  Log level (debug shows everything, error shows only errors)")
	levelOptions := []wizard.SelectOption{
		{Value: "debug", Label: "debug"},
		{Value: "info", Label: "info"},
		{Value: "warn", Label: "warn"},
		{Value: "error", Label: "error"},
	}
	level, err := wizard.PromptSelect("  Level", levelOptions, 1)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Level", level)

	// Format
	ui.Dimmed("  Output format for log entries")
	formatOptions := []wizard.SelectOption{
		{Value: "jsonl", Label: "jsonl (structured, machine-readable)"},
		{Value: "text", Label: "text (human-readable)"},
	}
	format, err := wizard.PromptSelect("  Format", formatOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Format", format)

	// Build stream config
	stream := map[string]interface{}{
		"transport":   "file",
		"destination": path,
		"level":       level,
		"format":      format,
	}

	if err := addStreamToFormation(rootDir, stream); err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Added file stream")
	return nil
}

// configureHTTPStream configures an HTTP log stream
func configureHTTPStream(rootDir string) error {
	fmt.Println()
	ui.Bold("HTTP Stream")
	fmt.Println()

	// Format (logging service)
	ui.Dimmed("  Format (choose based on your logging service)")
	formatOptions := []wizard.SelectOption{
		{Value: "jsonl", Label: "jsonl (generic JSON lines)"},
		{Value: "datadog", Label: "Datadog"},
		{Value: "splunk", Label: "Splunk (HEC)"},
		{Value: "elastic", Label: "Elastic"},
		{Value: "loki", Label: "Grafana (Loki)"},
		{Value: "newrelic", Label: "New Relic"},
		{Value: "otlp", Label: "OpenTelemetry"},
	}
	format, err := wizard.PromptSelect("  Format", formatOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Format", format)

	// Check for env vars based on format
	checkLoggingEnvVars(format)

	// Endpoint URL
	ui.Dimmed("  Endpoint URL for log ingestion")
	var destination string
	for {
		input, err := wizard.PromptString("  URL", "", nil)
		if err != nil {
			return err
		}
		destination = normalizeURL(input)
		if err := validateURL(destination); err != nil {
			ui.PromptError("  URL", input, err)
			continue
		}
		break
	}
	ui.PromptSuccess("  URL", destination)

	// Authentication
	ui.Dimmed("  Authentication method")
	authOptions := []wizard.SelectOption{
		{Value: "none", Label: "None"},
		{Value: "bearer", Label: "Bearer token"},
		{Value: "apikey", Label: "API key header"},
		{Value: "basic", Label: "Basic auth"},
	}
	authType, err := wizard.PromptSelect("  Auth", authOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Auth", authType)

	// Build stream config
	stream := map[string]interface{}{
		"transport":   "http",
		"destination": destination,
		"format":      format,
	}

	// Handle auth based on type
	if authType != "none" {
		auth := map[string]interface{}{
			"type": authType,
		}

		sm := secrets.NewManager(rootDir)

		switch authType {
		case "bearer":
			ui.Dimmed("  Bearer token for authentication")
			token, err := wizard.PromptString("  Token", "", nil)
			if err != nil {
				return err
			}
			if err := sm.Set("LOGGING_BEARER_TOKEN", token, true); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
			auth["token"] = "${{ secrets.LOGGING_BEARER_TOKEN }}"
			ui.PromptSuccess("  Saved", "LOGGING_BEARER_TOKEN")

		case "apikey":
			ui.Dimmed("  API key header name")
			headerName, err := wizard.PromptString("  Header name", "X-API-Key", nil)
			if err != nil {
				return err
			}
			ui.PromptSuccess("  Header name", headerName)

			ui.Dimmed("  API key value")
			apiKey, err := wizard.PromptString("  API key", "", nil)
			if err != nil {
				return err
			}
			if err := sm.Set("LOGGING_API_KEY", apiKey, true); err != nil {
				return fmt.Errorf("failed to save API key: %w", err)
			}
			auth["header"] = headerName
			auth["key"] = "${{ secrets.LOGGING_API_KEY }}"
			ui.PromptSuccess("  Saved", "LOGGING_API_KEY")

		case "basic":
			ui.Dimmed("  Username for basic auth")
			username, err := wizard.PromptString("  Username", "", nil)
			if err != nil {
				return err
			}
			ui.PromptSuccess("  Username", username)

			ui.Dimmed("  Password for basic auth")
			password, err := wizard.PromptString("  Password", "", nil)
			if err != nil {
				return err
			}
			if err := sm.Set("LOGGING_BASIC_PASSWORD", password, true); err != nil {
				return fmt.Errorf("failed to save password: %w", err)
			}
			auth["username"] = username
			auth["password"] = "${{ secrets.LOGGING_BASIC_PASSWORD }}"
			ui.PromptSuccess("  Saved", "LOGGING_BASIC_PASSWORD")
		}

		stream["auth"] = auth
	}

	if err := addStreamToFormation(rootDir, stream); err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Added http stream")
	return nil
}

// configureKafkaStream configures a Kafka log stream
func configureKafkaStream(rootDir string) error {
	fmt.Println()
	ui.Bold("Kafka Stream")
	fmt.Println()

	// Brokers
	ui.Dimmed("  Broker addresses (comma, space, or newline-separated)")
	brokersInput, err := wizard.PromptString("  Brokers", "localhost:9092", nil)
	if err != nil {
		return err
	}
	brokers := parseBrokerList(brokersInput)
	ui.PromptSuccess("  Brokers", strings.Join(brokers, ", "))

	// Topic
	ui.Dimmed("  Topic name for log messages")
	topic, err := wizard.PromptString("  Topic", "formation-logs", nil)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Topic", topic)

	// Format
	ui.Dimmed("  Output format for messages")
	formatOptions := []wizard.SelectOption{
		{Value: "jsonl", Label: "jsonl"},
		{Value: "msgpack", Label: "msgpack"},
	}
	format, err := wizard.PromptSelect("  Format", formatOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Format", format)

	// Authentication
	ui.Dimmed("  Authentication method")
	authOptions := []wizard.SelectOption{
		{Value: "none", Label: "None"},
		{Value: "sasl", Label: "SASL (username/password)"},
	}
	authType, err := wizard.PromptSelect("  Auth", authOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Auth", authType)

	// Build stream config
	stream := map[string]interface{}{
		"transport": "kafka",
		"brokers":   brokers,
		"topic":     topic,
		"format":    format,
	}

	// Handle SASL auth
	if authType == "sasl" {
		ui.Dimmed("  SASL username")
		username, err := wizard.PromptString("  Username", "", nil)
		if err != nil {
			return err
		}
		ui.PromptSuccess("  Username", username)

		ui.Dimmed("  SASL password")
		password, err := wizard.PromptString("  Password", "", nil)
		if err != nil {
			return err
		}

		sm := secrets.NewManager(rootDir)
		if err := sm.Set("KAFKA_PASSWORD", password, true); err != nil {
			return fmt.Errorf("failed to save password: %w", err)
		}
		ui.PromptSuccess("  Saved", "KAFKA_PASSWORD")

		stream["auth"] = map[string]interface{}{
			"type":     "sasl",
			"username": username,
			"password": "${{ secrets.KAFKA_PASSWORD }}",
		}
	}

	if err := addStreamToFormation(rootDir, stream); err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Added kafka stream")
	return nil
}

// viewEditStreams handles Flow 2: View/Edit Current Streams
func viewEditStreams(rootDir string) error {
	streams, err := getLoggingStreams(rootDir)
	if err != nil {
		return err
	}

	if len(streams) == 0 {
		fmt.Println()
		ui.Warning("No logging streams configured")
		ui.Dimmed("  Use 'muxi config logging' and select 'Add a new log stream' to add one.")
		return nil
	}

	fmt.Println()
	ui.Bold("Current Streams")
	fmt.Println()

	// Build options from streams
	options := make([]wizard.SelectOption, len(streams)+1)
	for i, stream := range streams {
		options[i] = wizard.SelectOption{
			Value: fmt.Sprintf("%d", i),
			Label: formatStreamLabel(stream),
		}
	}
	options[len(streams)] = wizard.SelectOption{
		Value: "back",
		Label: "← Back",
	}

	choice, err := wizard.PromptSelect("  Select a stream to edit", options, len(streams))
	if err != nil {
		return err
	}

	if choice == "back" {
		return nil
	}

	// For now, just show the stream details
	var idx int
	fmt.Sscanf(choice, "%d", &idx)
	if idx >= 0 && idx < len(streams) {
		fmt.Println()
		ui.Dimmed("  Stream details:")
		streamYAML, _ := yaml.Marshal(streams[idx])
		for _, line := range strings.Split(string(streamYAML), "\n") {
			if line != "" {
				fmt.Printf("    %s\n", line)
			}
		}
		fmt.Println()
		ui.Dimmed("  (Edit formation.yaml directly to modify stream settings)")
	}

	return nil
}

// removeStream handles Flow 3: Remove Stream
func removeStream(rootDir string) error {
	streams, err := getLoggingStreams(rootDir)
	if err != nil {
		return err
	}

	if len(streams) == 0 {
		fmt.Println()
		ui.Warning("No logging streams configured")
		return nil
	}

	fmt.Println()
	ui.Bold("Remove Stream")
	fmt.Println()

	// Build options from streams
	options := make([]wizard.SelectOption, len(streams)+1)
	for i, stream := range streams {
		options[i] = wizard.SelectOption{
			Value: fmt.Sprintf("%d", i),
			Label: formatStreamLabel(stream),
		}
	}
	options[len(streams)] = wizard.SelectOption{
		Value: "back",
		Label: "← Cancel",
	}

	choice, err := wizard.PromptSelect("  Select stream to remove", options, len(streams))
	if err != nil {
		return err
	}

	if choice == "back" {
		return nil
	}

	var idx int
	fmt.Sscanf(choice, "%d", &idx)

	// Confirm removal
	confirm, err := wizard.PromptConfirm("  Remove this stream?", false)
	if err != nil {
		return err
	}

	if !confirm {
		ui.PromptSkipped("  Cancelled")
		return nil
	}

	// Remove the stream
	if err := removeStreamFromFormation(rootDir, idx); err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Stream removed")
	return nil
}

// Helper functions

func checkLoggingEnvVars(format string) {
	var envVar, secretName string

	switch format {
	case "datadog":
		envVar = "DATADOG_API_KEY"
		secretName = "LOGGING_BEARER_TOKEN"
	case "splunk":
		envVar = "SPLUNK_HEC_TOKEN"
		secretName = "LOGGING_BEARER_TOKEN"
	default:
		return
	}

	if val := os.Getenv(envVar); val != "" {
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  Found %s in environment", envVar))
	}
	_ = secretName // Will be used when importing
}

func parseBrokerList(input string) []string {
	// Replace newlines and commas with spaces, then split
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, ",", " ")

	parts := strings.Fields(input)
	brokers := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			brokers = append(brokers, p)
		}
	}
	return brokers
}

func formatStreamLabel(stream map[string]interface{}) string {
	transport, _ := stream["transport"].(string)

	switch transport {
	case "stdout":
		level, _ := stream["level"].(string)
		format, _ := stream["format"].(string)
		return fmt.Sprintf("stdout (%s, %s)", level, format)

	case "file":
		dest, _ := stream["destination"].(string)
		level, _ := stream["level"].(string)
		format, _ := stream["format"].(string)
		return fmt.Sprintf("file → %s (%s, %s)", dest, level, format)

	case "http":
		dest, _ := stream["destination"].(string)
		format, _ := stream["format"].(string)
		return fmt.Sprintf("http → %s (%s)", dest, format)

	case "kafka":
		topic, _ := stream["topic"].(string)
		return fmt.Sprintf("kafka → %s", topic)

	default:
		return transport
	}
}

func getLoggingStreams(rootDir string) ([]map[string]interface{}, error) {
	formationPath := filepath.Join(rootDir, "formation.yaml")
	data, err := os.ReadFile(formationPath)
	if err != nil {
		return nil, err
	}

	var formation map[string]interface{}
	if err := yaml.Unmarshal(data, &formation); err != nil {
		return nil, err
	}

	logging, ok := formation["logging"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	streamsRaw, ok := logging["streams"].([]interface{})
	if !ok {
		return nil, nil
	}

	streams := make([]map[string]interface{}, 0, len(streamsRaw))
	for _, s := range streamsRaw {
		if stream, ok := s.(map[string]interface{}); ok {
			streams = append(streams, stream)
		}
	}

	return streams, nil
}

func addStreamToFormation(rootDir string, stream map[string]interface{}) error {
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

	// Find or create logging section
	var loggingNode *yaml.Node
	for i := 0; i < len(docNode.Content); i += 2 {
		if docNode.Content[i].Value == "logging" {
			loggingNode = docNode.Content[i+1]
			break
		}
	}

	if loggingNode == nil {
		// Create logging section
		loggingNode = &yaml.Node{Kind: yaml.MappingNode}
		docNode.Content = append(docNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "logging"},
			loggingNode,
		)
	}

	// Find or create streams array
	var streamsNode *yaml.Node
	for i := 0; i < len(loggingNode.Content); i += 2 {
		if loggingNode.Content[i].Value == "streams" {
			streamsNode = loggingNode.Content[i+1]
			break
		}
	}

	if streamsNode == nil {
		streamsNode = &yaml.Node{Kind: yaml.SequenceNode}
		loggingNode.Content = append(loggingNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "streams"},
			streamsNode,
		)
	}

	// Convert stream map to YAML node
	streamNode := mapToYAMLNode(stream)
	streamsNode.Content = append(streamsNode.Content, streamNode)

	// Write back
	output, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("failed to marshal formation.yaml: %w", err)
	}

	if err := os.WriteFile(formationPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write formation.yaml: %w", err)
	}

	return nil
}

func removeStreamFromFormation(rootDir string, idx int) error {
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

	// Find logging section
	for i := 0; i < len(docNode.Content); i += 2 {
		if docNode.Content[i].Value == "logging" {
			loggingNode := docNode.Content[i+1]

			// Find streams array
			for j := 0; j < len(loggingNode.Content); j += 2 {
				if loggingNode.Content[j].Value == "streams" {
					streamsNode := loggingNode.Content[j+1]

					// Remove the stream at idx
					if idx >= 0 && idx < len(streamsNode.Content) {
						streamsNode.Content = append(
							streamsNode.Content[:idx],
							streamsNode.Content[idx+1:]...,
						)
					}
					break
				}
			}
			break
		}
	}

	// Write back
	output, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("failed to marshal formation.yaml: %w", err)
	}

	if err := os.WriteFile(formationPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write formation.yaml: %w", err)
	}

	return nil
}

func mapToYAMLNode(m map[string]interface{}) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}

	// Define order for stream fields
	orderedKeys := []string{"transport", "destination", "level", "format", "brokers", "topic", "auth"}

	for _, key := range orderedKeys {
		if val, ok := m[key]; ok {
			node.Content = append(node.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: key},
				valueToYAMLNode(val),
			)
		}
	}

	// Add any remaining keys not in the ordered list
	for key, val := range m {
		found := false
		for _, ok := range orderedKeys {
			if key == ok {
				found = true
				break
			}
		}
		if !found {
			node.Content = append(node.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: key},
				valueToYAMLNode(val),
			)
		}
	}

	return node
}

func valueToYAMLNode(val interface{}) *yaml.Node {
	switch v := val.(type) {
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: v}
	case int:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", v), Tag: "!!int"}
	case bool:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%t", v)}
	case []string:
		node := &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
		for _, s := range v {
			node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: s})
		}
		return node
	case map[string]interface{}:
		return mapToYAMLNode(v)
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", v)}
	}
}

// validateBroker checks if a broker address is valid (host:port format)
func validateBroker(broker string) error {
	// Simple validation: must contain colon and have non-empty parts
	parts := strings.Split(broker, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid broker format (expected host:port)")
	}
	// Check port is numeric
	matched, _ := regexp.MatchString(`^\d+$`, parts[1])
	if !matched {
		return fmt.Errorf("invalid port number")
	}
	return nil
}
