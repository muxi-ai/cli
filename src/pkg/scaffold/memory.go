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
)

// Known embedding model dimensions
var embeddingModelDimensions = map[string]int{
	// OpenAI
	"openai/text-embedding-3-large": 3072,
	"openai/text-embedding-3-small": 1536,
	"openai/text-embedding-ada-002": 1536,
	// Cohere
	"cohere/embed-english-v3.0":      1024,
	"cohere/embed-multilingual-v3.0": 1024,
	"cohere/embed-english-light-v3.0": 384,
	// Voyage
	"voyage/voyage-large-2": 1536,
	"voyage/voyage-code-2":  1536,
	// Local default
	"all-MiniLM-L6-v2": 384,
}

// ConfigureMemory runs the memory configuration wizard
func ConfigureMemory() error {
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
│ [⚙] Configure Memory                                    MUXI │
│──────────────────────────────────────────────────────────────│
│ ℹ Memory enables agents to remember context and information. │
│ Working memory is always on. Persistent memory needs a DB.   │
╰──────────────────────────────────────────────────────────────╯`)

	// Step 1: What to configure
	options := []wizard.SelectOption{
		{Value: "working", Label: "Working memory (in-memory, always enabled)"},
		{Value: "buffer", Label: "Buffer memory (conversation context)"},
		{Value: "persistent", Label: "Persistent memory (database-backed, long-term)"},
	}

	choice, err := wizard.PromptSelect("What would you like to configure?", options, 0)
	if err != nil {
		return err
	}

	switch choice {
	case "working":
		return configureWorkingMemory(ctx.RootDir)
	case "buffer":
		return configureBufferMemory(ctx.RootDir)
	case "persistent":
		return configurePersistentMemory(ctx.RootDir)
	}

	return nil
}

// configureWorkingMemory handles Flow 1: Working Memory
func configureWorkingMemory(rootDir string) error {
	fmt.Println()
	ui.Bold("Working Memory")
	fmt.Println()
	ui.Dimmed("  Short-term vector memory for semantic search within a session.")
	fmt.Println()

	// Select mode
	modeOptions := []wizard.SelectOption{
		{Value: "local", Label: "Local (default, in-process)"},
		{Value: "remote", Label: "Remote (FAISSx server)"},
	}

	mode, err := wizard.PromptSelect("  Select mode", modeOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Mode", mode)

	if mode == "local" {
		return configureLocalWorkingMemory(rootDir)
	}
	return configureRemoteWorkingMemory(rootDir)
}

// configureLocalWorkingMemory configures local working memory
func configureLocalWorkingMemory(rootDir string) error {
	fmt.Println()
	ui.Dimmed("  \"auto\" uses 10% of RAM (min 64MB, max 1GB)")
	maxMemory, err := wizard.PromptString("  Max memory", "auto", nil)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Max memory", maxMemory)

	// Auto-detect vector dimension from embedding model
	vectorDim, embeddingModel := getEmbeddingVectorDimension(rootDir)
	if embeddingModel != "" {
		ui.Dimmed(fmt.Sprintf("  Vector dimension from embedding model: %s", embeddingModel))
	} else {
		ui.Dimmed("  Vector dimension (no embedding model configured, using default)")
	}
	ui.PromptSuccess("  Vector dimension", vectorDim)

	ui.Dimmed("  How often to clean up old vectors (FIFO)")
	fifoInterval, err := wizard.PromptString("  Cleanup interval (minutes)", "5", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Cleanup interval", fifoInterval+" min")

	// Update formation.yaml
	return updateWorkingMemoryInFormation(rootDir, "local", maxMemory, vectorDim, fifoInterval, "", "", "")
}

// configureRemoteWorkingMemory configures remote FAISSx working memory
func configureRemoteWorkingMemory(rootDir string) error {
	fmt.Println()
	ui.Dimmed("  FAISSx server endpoint (tcp:// or tcps://)")
	serverURL, err := wizard.PromptString("  Server URL", "tcp://localhost:8000", validateFAISSxURL)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Server URL", serverURL)

	// API Key
	fmt.Println()
	ui.Dimmed("  API key for FAISSx authentication")
	apiKey, err := promptForFAISSxSecret(rootDir, "FAISSX_API_KEY", "API Key")
	if err != nil {
		return err
	}

	// Tenant ID
	fmt.Println()
	ui.Dimmed("  Tenant identifier for multi-tenant FAISSx")
	tenantID, err := promptForFAISSxSecret(rootDir, "FAISSX_TENANT_ID", "Tenant ID")
	if err != nil {
		return err
	}

	// Memory settings
	fmt.Println()
	ui.Dimmed("  Remote mode requires explicit memory limit")
	maxMemory, err := wizard.PromptString("  Max memory (MB)", "256", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Max memory", maxMemory+" MB")

	// Auto-detect vector dimension from embedding model
	vectorDim, embeddingModel := getEmbeddingVectorDimension(rootDir)
	if embeddingModel != "" {
		ui.Dimmed(fmt.Sprintf("  Vector dimension from embedding model: %s", embeddingModel))
	} else {
		ui.Dimmed("  Vector dimension (no embedding model configured, using default)")
	}
	ui.PromptSuccess("  Vector dimension", vectorDim)

	ui.Dimmed("  How often to clean up old vectors (FIFO)")
	fifoInterval, err := wizard.PromptString("  Cleanup interval (minutes)", "5", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Cleanup interval", fifoInterval+" min")

	// Update formation.yaml
	return updateWorkingMemoryInFormation(rootDir, "remote", maxMemory, vectorDim, fifoInterval, serverURL, apiKey, tenantID)
}

// promptForFAISSxSecret prompts for a FAISSx secret with env var detection
func promptForFAISSxSecret(rootDir, secretName, displayName string) (string, error) {
	var value string

	// Check for existing env var
	envValue := os.Getenv(secretName)
	if envValue != "" {
		masked := maskSecretValue(envValue)
		useEnv, err := wizard.PromptConfirm(fmt.Sprintf("  ℹ Found %s in environment [%s]. Use it?", secretName, masked), true)
		if err != nil {
			return "", err
		}
		if useEnv {
			value = envValue
		}
	}

	// If not using env var, prompt for value
	if value == "" {
		var err error
		if secretName == "FAISSX_API_KEY" {
			value, err = wizard.PromptPassword("  Enter "+displayName, true)
		} else {
			value, err = wizard.PromptString("  "+displayName, "", nil)
		}
		if err != nil {
			return "", err
		}

		if value == "" {
			ui.PromptSkipped("  " + displayName)
			return "", nil
		}
	}

	// Save to secrets
	secretsManager := secrets.NewManager(rootDir)
	if err := secretsManager.Set(secretName, value, true); err != nil {
		return "", fmt.Errorf("failed to save secret: %w", err)
	}
	ui.Success(fmt.Sprintf("  Saved %s to secrets", secretName))

	return value, nil
}

// configureBufferMemory handles Flow 2: Buffer Memory
func configureBufferMemory(rootDir string) error {
	fmt.Println()
	ui.Bold("Buffer Memory")
	fmt.Println()
	ui.Dimmed("  Conversation context that persists across messages in a session.")
	fmt.Println()

	ui.Dimmed("  Number of recent messages to keep in context")
	windowSize, err := wizard.PromptString("  Context window size", "10", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Context window size", windowSize)

	ui.Dimmed("  Multiplier for total buffer (total = size × multiplier)")
	multiplier, err := wizard.PromptString("  Buffer multiplier", "10", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Buffer multiplier", multiplier)

	ui.Dimmed("  Use vector similarity to find relevant past messages")
	vectorSearch, err := wizard.PromptConfirm("  Enable vector similarity search?", true)
	if err != nil {
		return err
	}
	if vectorSearch {
		ui.PromptSuccess("  Vector search", "enabled")
	} else {
		ui.PromptSkipped("  Vector search")
	}

	// Update formation.yaml
	return updateBufferMemoryInFormation(rootDir, windowSize, multiplier, vectorSearch)
}

// configurePersistentMemory handles Flow 3: Persistent Memory
func configurePersistentMemory(rootDir string) error {
	fmt.Println()
	ui.Bold("Persistent Memory")
	fmt.Println()
	ui.Dimmed("  Long-term memory stored in a database, persists across sessions.")
	fmt.Println()

	// Select database type
	dbOptions := []wizard.SelectOption{
		{Value: "postgres", Label: "PostgreSQL (multi-user; requires PostgreSQL + pgvector)"},
		{Value: "sqlite", Label: "SQLite (single-user; good for development)"},
	}

	dbType, err := wizard.PromptSelect("  Select database type", dbOptions, 0)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Database type", dbType)

	if dbType == "postgres" {
		return configurePostgresMemory(rootDir)
	}
	return configureSQLiteMemory(rootDir)
}

// configurePostgresMemory configures PostgreSQL persistent memory
func configurePostgresMemory(rootDir string) error {
	fmt.Println()
	ui.Dimmed("  Requires PostgreSQL 17+ with pgvector extension installed.")
	fmt.Println()

	// Check for DATABASE_URL env var
	var connectionString string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" && (strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://")) {
		masked := maskConnectionString(dbURL)
		useEnv, err := wizard.PromptConfirm(fmt.Sprintf("  ℹ Found DATABASE_URL in environment [%s]. Use it?", masked), true)
		if err != nil {
			return err
		}
		if useEnv {
			connectionString = dbURL
		}
	}

	if connectionString == "" {
		ui.Dimmed("  Enter a full connection string (postgres://...) or just a hostname")
		input, err := wizard.PromptString("  Connection string or hostname", "", nil)
		if err != nil {
			return err
		}

		if input == "" {
			ui.PromptSkipped("  Connection")
			return nil
		}

		// Check if it's a connection string or hostname
		if strings.HasPrefix(input, "postgres://") || strings.HasPrefix(input, "postgresql://") {
			connectionString = input
			ui.PromptSuccess("  Connection string", maskConnectionString(input))
		} else {
			// It's a hostname - prompt for details
			connectionString, err = buildPostgresConnectionString(input)
			if err != nil {
				return err
			}
		}
	}

	// Save connection string to secrets
	secretsManager := secrets.NewManager(rootDir)
	if err := secretsManager.Set("PERSISTENT_DB_CONNECTION_STRING", connectionString, true); err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}
	ui.Success("  Saved PERSISTENT_DB_CONNECTION_STRING to secrets")

	// PostgreSQL-specific settings
	fmt.Println()
	ui.Dimmed("  Maximum time to wait for database queries")
	queryTimeout, err := wizard.PromptString("  Query timeout (seconds)", "30", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Query timeout", queryTimeout+"s")

	ui.Dimmed("  Generate LLM-synthesized summaries of user context")
	enableSynopsis, err := wizard.PromptConfirm("  Enable user synopsis?", true)
	if err != nil {
		return err
	}

	var synopsisTTL string
	if enableSynopsis {
		ui.PromptSuccess("  User synopsis", "enabled")

		ui.Dimmed("  How long to cache generated synopses")
		synopsisTTL, err = wizard.PromptString("  Synopsis cache TTL (seconds)", "3600", validatePositiveInt)
		if err != nil {
			return err
		}
		ui.PromptSuccess("  Synopsis cache TTL", synopsisTTL+"s")
	} else {
		ui.PromptSkipped("  User synopsis")
	}

	// Update formation.yaml
	return updatePersistentMemoryInFormation(rootDir, "postgres", queryTimeout, enableSynopsis, synopsisTTL)
}

// buildPostgresConnectionString prompts for details and builds a connection string
func buildPostgresConnectionString(host string) (string, error) {
	ui.PromptSuccess("  Host", host)

	ui.Dimmed("  PostgreSQL port (default: 5432)")
	port, err := wizard.PromptString("  Port", "5432", validatePositiveInt)
	if err != nil {
		return "", err
	}
	ui.PromptSuccess("  Port", port)

	ui.Dimmed("  Database name")
	database, err := wizard.PromptString("  Database", "", validateNotEmpty)
	if err != nil {
		return "", err
	}
	ui.PromptSuccess("  Database", database)

	ui.Dimmed("  Database username")
	username, err := wizard.PromptString("  Username", "", validateNotEmpty)
	if err != nil {
		return "", err
	}
	ui.PromptSuccess("  Username", username)

	ui.Dimmed("  Database password")
	password, err := wizard.PromptPassword("  Password", true)
	if err != nil {
		return "", err
	}
	ui.PromptSuccess("  Password", "********")

	// Build connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
	fmt.Println()
	ui.Dimmed(fmt.Sprintf("  Built connection string: postgres://%s:***@%s:%s/%s", username, host, port, database))

	return connStr, nil
}

// configureSQLiteMemory configures SQLite persistent memory
func configureSQLiteMemory(rootDir string) error {
	fmt.Println()
	ui.Dimmed("  Path to SQLite database file (will be created if doesn't exist)")
	dbPath, err := wizard.PromptString("  Database path", "./memory.db", nil)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Database path", dbPath)

	// Update formation.yaml (no secrets for SQLite)
	return updateSQLiteMemoryInFormation(rootDir, dbPath)
}

// Validation functions

func validateFAISSxURL(input string) error {
	if !strings.HasPrefix(input, "tcp://") && !strings.HasPrefix(input, "tcps://") {
		return fmt.Errorf("URL must start with tcp:// or tcps://")
	}
	return nil
}

func validateNotEmpty(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

// Helper functions

// getEmbeddingVectorDimension reads the embedding model from formation.yaml and returns its dimension
func getEmbeddingVectorDimension(rootDir string) (string, string) {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return "1536", "" // Default to OpenAI's common dimension
	}

	// Look for embedding model in formation.yaml
	// Pattern: - embedding: "model/name"
	contentStr := string(content)
	embeddingPattern := regexp.MustCompile(`-\s*embedding:\s*"?([^"\s\n]+)"?`)
	matches := embeddingPattern.FindStringSubmatch(contentStr)

	if len(matches) > 1 {
		model := matches[1]
		if dim, ok := embeddingModelDimensions[model]; ok {
			return fmt.Sprintf("%d", dim), model
		}
		// Unknown model, default to 1536
		return "1536", model
	}

	// No embedding model configured - check if using local fallback
	// Default local model is all-MiniLM-L6-v2 with 384 dimensions
	return "384", "all-MiniLM-L6-v2 (local)"
}

func maskSecretValue(value string) string {
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:3] + "..." + value[len(value)-4:]
}

func maskConnectionString(connStr string) string {
	// Mask password in connection string
	// postgres://user:password@host:port/db -> postgres://user:***@host:port/db
	re := regexp.MustCompile(`(postgres(?:ql)?://[^:]+:)([^@]+)(@.+)`)
	return re.ReplaceAllString(connStr, "${1}***${3}")
}

// Formation update functions

func updateWorkingMemoryInFormation(rootDir, mode, maxMemory, vectorDim, fifoInterval, serverURL, apiKey, tenantID string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	// Build working memory YAML
	var memoryYAML strings.Builder
	memoryYAML.WriteString("  working:\n")
	memoryYAML.WriteString(fmt.Sprintf("    mode: \"%s\"\n", mode))
	memoryYAML.WriteString(fmt.Sprintf("    max_memory_mb: \"%s\"\n", maxMemory))
	memoryYAML.WriteString(fmt.Sprintf("    vector_dimension: %s\n", vectorDim))
	memoryYAML.WriteString(fmt.Sprintf("    fifo_interval_min: %s\n", fifoInterval))

	if mode == "remote" {
		memoryYAML.WriteString("    remote:\n")
		memoryYAML.WriteString(fmt.Sprintf("      url: \"%s\"\n", serverURL))
		memoryYAML.WriteString("      api_key: \"${{ secrets.FAISSX_API_KEY }}\"\n")
		memoryYAML.WriteString("      tenant: \"${{ secrets.FAISSX_TENANT_ID }}\"\n")
	}

	return updateMemorySectionInFormation(formationFile, content, "working", memoryYAML.String())
}

func updateBufferMemoryInFormation(rootDir, windowSize, multiplier string, vectorSearch bool) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	// Build buffer memory YAML
	var memoryYAML strings.Builder
	memoryYAML.WriteString("  buffer:\n")
	memoryYAML.WriteString(fmt.Sprintf("    size: %s\n", windowSize))
	memoryYAML.WriteString(fmt.Sprintf("    multiplier: %s\n", multiplier))
	memoryYAML.WriteString(fmt.Sprintf("    vector_search: %t\n", vectorSearch))

	return updateMemorySectionInFormation(formationFile, content, "buffer", memoryYAML.String())
}

func updatePersistentMemoryInFormation(rootDir, dbType, queryTimeout string, enableSynopsis bool, synopsisTTL string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	// Build persistent memory YAML
	var memoryYAML strings.Builder
	memoryYAML.WriteString("  persistent:\n")
	memoryYAML.WriteString("    connection_string: \"${{ secrets.PERSISTENT_DB_CONNECTION_STRING }}\"\n")
	memoryYAML.WriteString(fmt.Sprintf("    query_timeout_seconds: %s\n", queryTimeout))
	
	if enableSynopsis {
		memoryYAML.WriteString("    user_synopsis:\n")
		memoryYAML.WriteString("      enabled: true\n")
		memoryYAML.WriteString(fmt.Sprintf("      cache_ttl: %s\n", synopsisTTL))
	}

	return updateMemorySectionInFormation(formationFile, content, "persistent", memoryYAML.String())
}

func updateSQLiteMemoryInFormation(rootDir, dbPath string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	// Build SQLite memory YAML
	var memoryYAML strings.Builder
	memoryYAML.WriteString("  persistent:\n")
	memoryYAML.WriteString(fmt.Sprintf("    connection_string: \"sqlite:///%s\"\n", dbPath))

	return updateMemorySectionInFormation(formationFile, content, "persistent", memoryYAML.String())
}

func updateMemorySectionInFormation(formationFile string, content []byte, section, sectionYAML string) error {
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Find memory section and the specific subsection
	memoryLineIdx := -1
	sectionStartIdx := -1
	sectionEndIdx := -1
	inMemory := false
	inSection := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Found memory: section
		if trimmed == "memory:" {
			memoryLineIdx = i
			inMemory = true
			continue
		}

		// Exit memory section if we hit a new top-level key
		if inMemory && len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") && trimmed != "" {
			inMemory = false
			if inSection {
				sectionEndIdx = i
				inSection = false
			}
			continue
		}

		// Found the specific subsection (e.g., "  working:", "  buffer:", "  persistent:")
		if inMemory && trimmed == section+":" && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
			sectionStartIdx = i
			inSection = true
			continue
		}

		// Exit subsection if we hit another subsection or end of memory
		if inSection && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && trimmed != "" {
			sectionEndIdx = i
			inSection = false
		}
	}

	// If section exists, mark end at EOF if not found
	if inSection && sectionEndIdx == -1 {
		sectionEndIdx = len(lines)
	}

	var result []string

	if sectionStartIdx >= 0 && sectionEndIdx >= 0 {
		// Replace existing section
		result = append(result, lines[:sectionStartIdx]...)
		result = append(result, strings.TrimSuffix(sectionYAML, "\n"))
		result = append(result, lines[sectionEndIdx:]...)
	} else if memoryLineIdx >= 0 {
		// Memory exists but subsection doesn't - add it
		for i, line := range lines {
			result = append(result, line)
			if i == memoryLineIdx {
				result = append(result, strings.TrimSuffix(sectionYAML, "\n"))
			}
		}
	} else {
		// No memory section - add at end
		result = lines
		result = append(result, "")
		result = append(result, "memory:")
		result = append(result, strings.TrimSuffix(sectionYAML, "\n"))
	}

	output := strings.Join(result, "\n")
	if err := os.WriteFile(formationFile, []byte(output), 0644); err != nil {
		return err
	}

	ui.Success("Updated formation.yaml with memory configuration")
	return nil
}
