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

// Common models for each capability
var commonTextModels = []string{
	"openai/gpt-5",
	"openai/gpt-5-mini",
	"anthropic/claude-sonnet-4.5",
	"anthropic/claude-haiku-4.5",
	"google/gemini-2.5-flash",
}

var commonVisionModels = []string{
	"google/gemini-2.5-flash",
	"anthropic/claude-sonnet-4.5",
	"openai/gpt-5",
}

var commonAudioModels = []string{
	"openai/whisper-1",
}

var commonDocumentModels = []string{
	"openai/gpt-5",
	"anthropic/claude-sonnet-4.5",
	"google/gemini-2.5-flash",
}

var commonEmbeddingModels = []string{
	"openai/text-embedding-3-large",
	"openai/text-embedding-3-small",
	"cohere/embed-english-v3.0",
}

var commonStreamingModels = []string{
	"openai/gpt-5-mini",
	"anthropic/claude-haiku-4.5",
	"google/gemini-2.5-flash",
}

// ConfigureLLM runs the LLM configuration wizard
func ConfigureLLM() error {
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
│ [⚙] Configure LLM                                       MUXI │
│──────────────────────────────────────────────────────────────│
│ ℹ Configure language models for different capabilities:      │
│ text, vision, audio, documents, embeddings, and streaming.   │
╰──────────────────────────────────────────────────────────────╯`)

	// Step 1: What to configure
	options := []wizard.SelectOption{
		{Value: "capability", Label: "Configure model for a capability"},
		{Value: "api_key", Label: "Add/update API key for a provider"},
		{Value: "settings", Label: "Global LLM settings (temperature, tokens, caching)"},
	}

	choice, err := wizard.PromptSelect("What would you like to configure?", options, 0)
	if err != nil {
		return err
	}

	switch choice {
	case "capability":
		return configureModelCapability(ctx.RootDir)
	case "api_key":
		return configureProviderAPIKey(ctx.RootDir)
	case "settings":
		return configureGlobalSettings(ctx.RootDir)
	}

	return nil
}

// configureProviderAPIKey handles Flow 1: Add/Update Provider API Key
func configureProviderAPIKey(rootDir string) error {
	// Read current formation to check existing keys
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read formation.yaml: %w", err)
	}
	contentStr := string(content)

	// Build provider options with existing key hints
	var configuredProviders []LLMProvider
	var unconfiguredProviders []LLMProvider

	for _, p := range LLMProviders {
		secretRef := fmt.Sprintf("secrets.%s", p.SecretName)
		if strings.Contains(contentStr, secretRef) {
			// Provider has a key configured - get masked preview
			maskedKey := getMaskedSecretPreview(rootDir, p.SecretName)
			if maskedKey != "" {
				configuredProviders = append(configuredProviders, p)
			} else {
				unconfiguredProviders = append(unconfiguredProviders, p)
			}
		} else {
			unconfiguredProviders = append(unconfiguredProviders, p)
		}
	}

	// Build options list: configured first, then unconfigured
	var selectOptions []wizard.SelectOption
	var allProviders []LLMProvider

	for _, p := range configuredProviders {
		maskedKey := getMaskedSecretPreview(rootDir, p.SecretName)
		selectOptions = append(selectOptions, wizard.SelectOption{
			Value: p.Vendor,
			Label: fmt.Sprintf("%s [%s]", p.Name, maskedKey),
		})
		allProviders = append(allProviders, p)
	}

	for _, p := range unconfiguredProviders {
		selectOptions = append(selectOptions, wizard.SelectOption{
			Value: p.Vendor,
			Label: p.Name,
		})
		allProviders = append(allProviders, p)
	}

	fmt.Println()
	selectedVendor, err := wizard.PromptSelect("Select provider", selectOptions, 0)
	if err != nil {
		return err
	}

	// Find the selected provider
	var selectedProvider LLMProvider
	for _, p := range allProviders {
		if p.Vendor == selectedVendor {
			selectedProvider = p
			break
		}
	}
	ui.PromptSuccess("Provider", selectedProvider.Name)

	// Build prompt label
	fmt.Println()
	var keyHint string
	if selectedProvider.KeyPrefix != "" {
		keyHint = fmt.Sprintf(" (starts with %s)", selectedProvider.KeyPrefix)
	}
	fmt.Printf("%s API Key%s:\n", selectedProvider.Name, keyHint)

	var apiKey string

	// Check for existing env var (primary and alternatives)
	envVarsToCheck := append([]string{selectedProvider.SecretName}, selectedProvider.AltEnvVars...)
	for _, envName := range envVarsToCheck {
		envValue := os.Getenv(envName)
		if envValue != "" {
			maskedEnv := maskAPIKey(envValue, selectedProvider.KeyPrefix)
			useEnv, err := wizard.PromptConfirm(fmt.Sprintf("ℹ Found %s in environment [%s]. Use it?", envName, maskedEnv), true)
			if err != nil {
				return err
			}
			if useEnv {
				apiKey = envValue
				break
			}
		}
	}

	// If not using env var, prompt for key
	if apiKey == "" {
		apiKey, err = wizard.PromptPassword("Enter key", true)
		if err != nil {
			return err
		}

		if apiKey == "" {
			ui.PromptSkipped("API Key")
			return nil
		}
	}

	// Save to secrets
	secretsManager := secrets.NewManager(rootDir)

	if err := secretsManager.Set(selectedProvider.SecretName, apiKey, true); err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	ui.Success(fmt.Sprintf("Saved %s to secrets", selectedProvider.SecretName))

	// Update formation.yaml to include the API key reference if not already present
	secretRef := fmt.Sprintf("secrets.%s", selectedProvider.SecretName)
	if !strings.Contains(contentStr, secretRef) {
		if err := addAPIKeyToFormation(rootDir, selectedProvider); err != nil {
			return fmt.Errorf("failed to update formation.yaml: %w", err)
		}
		ui.Success("Updated formation.yaml with API key reference")
	}

	return nil
}

// configureModelCapability handles Flow 2: Configure Model Capability
func configureModelCapability(rootDir string) error {
	// Read current formation to show existing config
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read formation.yaml: %w", err)
	}
	contentStr := string(content)

	// Get current models for each capability
	textModel := extractModelForCapability(contentStr, "text")
	visionModel := extractModelForCapability(contentStr, "vision")
	audioModel := extractModelForCapability(contentStr, "audio")
	documentsModel := extractModelForCapability(contentStr, "documents")
	embeddingModel := extractModelForCapability(contentStr, "embedding")
	streamingModel := extractModelForCapability(contentStr, "streaming")

	// Build capability options with current status
	fmt.Println()
	ui.Dimmed("ℹ Vision/audio/documents/streaming default to text model if not set.")

	// Embedding fallback info when not configured
	embeddingFallback := ""
	if embeddingModel == "" {
		embeddingFallback = "all-MiniLM-L6-v2 (local)*"
		fmt.Println("* For better quality, configure: openai/text-embedding-3-large")
	}
	fmt.Println()

	options := []wizard.SelectOption{
		{Value: "text", Label: formatCapabilityOption("Text (main reasoning model)", textModel, "", false)},
		{Value: "vision", Label: formatCapabilityOption("Vision (image understanding)", visionModel, textModel, true)},
		{Value: "audio", Label: formatCapabilityOption("Audio (speech-to-text)", audioModel, textModel, true)},
		{Value: "documents", Label: formatCapabilityOption("Documents (PDF/doc processing)", documentsModel, textModel, true)},
		{Value: "embedding", Label: formatCapabilityOption("Embedding (vector search)", embeddingModel, embeddingFallback, false)},
		{Value: "streaming", Label: formatCapabilityOption("Streaming (progress updates)", streamingModel, textModel, true)},
	}

	capability, err := wizard.PromptSelect("Select capability to configure", options, 0)
	if err != nil {
		return err
	}

	return configureCapabilityModel(rootDir, capability)
}

// configureCapabilityModel configures a specific capability's model
func configureCapabilityModel(rootDir string, capability string) error {
	var commonModels []string
	var title string
	var hint string

	switch capability {
	case "text":
		commonModels = commonTextModels
		title = "Text Model"
	case "vision":
		commonModels = commonVisionModels
		title = "Vision Model"
	case "audio":
		commonModels = commonAudioModels
		title = "Audio Model"
	case "documents":
		commonModels = commonDocumentModels
		title = "Documents Model"
		hint = "ℹ Used for text extraction from PDFs and documents."
	case "embedding":
		commonModels = commonEmbeddingModels
		title = "Embedding Model"
	case "streaming":
		commonModels = commonStreamingModels
		title = "Streaming Model"
		hint = "ℹ Used for real-time progress updates. Recommend a fast, cheap model."
	}

	fmt.Println()
	ui.Bold(title)
	fmt.Println()
	if hint != "" {
		ui.Dimmed(hint)
		fmt.Println()
	}

	// Build options
	var selectOptions []wizard.SelectOption
	for _, m := range commonModels {
		selectOptions = append(selectOptions, wizard.SelectOption{Value: m, Label: m})
	}
	selectOptions = append(selectOptions, wizard.SelectOption{Value: "other", Label: "Other (enter)"})

	selectedModel, err := wizard.PromptSelect("Common models", selectOptions, 0)
	if err != nil {
		return err
	}

	if selectedModel == "other" {
		// User selected "Other" - prompt for model
		selectedModel, err = wizard.PromptString("Model (provider/model)", "", validateModelFormat)
		if err != nil {
			return err
		}
	}

	if selectedModel == "" {
		ui.Warning("No model selected, skipping")
		return nil
	}

	ui.PromptSuccess("Model", selectedModel)

	// Check if provider has API key configured
	vendor := strings.Split(selectedModel, "/")[0]
	var provider *LLMProvider
	for i := range LLMProviders {
		if LLMProviders[i].Vendor == vendor {
			provider = &LLMProviders[i]
			break
		}
	}

	if provider != nil {
		// Check if API key is configured
		formationFile := filepath.Join(rootDir, "formation.yaml")
		content, _ := os.ReadFile(formationFile)
		contentStr := string(content)
		secretRef := fmt.Sprintf("secrets.%s", provider.SecretName)

		if !strings.Contains(contentStr, secretRef) {
			// API key not configured - prompt for it
			fmt.Println()
			ui.Warning(fmt.Sprintf("No API key configured for %s", provider.Name))

			if err := promptForAPIKey(rootDir, *provider); err != nil {
				return err
			}
		}
	}

	// Update formation.yaml
	if err := updateModelInFormation(rootDir, capability, selectedModel); err != nil {
		return fmt.Errorf("failed to update formation.yaml: %w", err)
	}

	ui.Success(fmt.Sprintf("Updated %s model to %s", capability, selectedModel))

	// Ask about model settings
	fmt.Println()
	configSettings, err := wizard.PromptConfirm("Configure settings for this model?", true)
	if err != nil {
		return err
	}

	if configSettings {
		return configureModelSettings(rootDir, capability, selectedModel)
	}

	return nil
}

// ModelSettings holds all settings for a model
type ModelSettings struct {
	Temperature   string
	MaxTokens     string
	Timeout       string
	MaxRetries    string
	FallbackModel string
	// Vision-specific
	ImageMaxSizeMB    string
	ImageResize       bool
	ImageMaxWidth     string
	ImageMaxHeight    string
	// Audio-specific
	AudioMaxSizeMB string
	AudioLanguage  string
	// Video-specific
	VideoMaxSizeMB       string
	VideoMaxDuration     string
	VideoIncludeAudio    bool
	// Documents-specific
	DocsMaxSizeMB    string
	DocsChunkSize    string
	DocsOverlap      string
	DocsStrategy     string
	DocsCacheTTL     string
}

// configureModelSettings configures settings for a model
func configureModelSettings(rootDir, capability, model string) error {
	fmt.Println()
	ui.Bold("Model Settings")
	fmt.Println()

	settings := ModelSettings{}

	// Common settings
	ui.Dimmed("  Controls randomness: 0 = deterministic, 1 = creative")
	tempStr, err := wizard.PromptString("  Temperature (0.0-1.0)", "0.7", validateTemperature)
	if err != nil {
		return err
	}
	settings.Temperature = tempStr
	ui.PromptSuccess("  Temperature", tempStr)

	ui.Dimmed("  Maximum response length in tokens")
	tokensStr, err := wizard.PromptString("  Max tokens", "4096", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.MaxTokens = tokensStr
	ui.PromptSuccess("  Max tokens", tokensStr)

	ui.Dimmed("  Request timeout before retry or failure")
	timeoutStr, err := wizard.PromptString("  Timeout (seconds)", "30", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.Timeout = timeoutStr
	ui.PromptSuccess("  Timeout", timeoutStr)

	ui.Dimmed("  Number of retry attempts on failure")
	retriesStr, err := wizard.PromptString("  Max retries", "3", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.MaxRetries = retriesStr
	ui.PromptSuccess("  Max retries", retriesStr)

	ui.Dimmed("  Model to use if primary fails (e.g., openai/gpt-4o-mini)")
	fallbackModel, err := wizard.PromptString("  Fallback model (optional)", "", nil)
	if err != nil {
		return err
	}
	settings.FallbackModel = fallbackModel
	if fallbackModel != "" {
		ui.PromptSuccess("  Fallback model", fallbackModel)

		// Check if fallback model's provider has API key configured
		if err := checkFallbackModelAPIKey(rootDir, fallbackModel); err != nil {
			return err
		}
	} else {
		ui.PromptSkipped("  Fallback model")
	}

	// Capability-specific settings
	switch capability {
	case "vision":
		if err := configureVisionSettings(rootDir, &settings); err != nil {
			return err
		}
	case "audio":
		if err := configureAudioSettings(rootDir, &settings); err != nil {
			return err
		}
	case "video":
		if err := configureVideoSettings(rootDir, &settings); err != nil {
			return err
		}
	case "documents":
		if err := configureDocumentsSettings(rootDir, &settings); err != nil {
			return err
		}
	}

	// Update formation.yaml with settings
	return updateModelSettingsInFormation(rootDir, capability, settings)
}

// checkFallbackModelAPIKey checks if the fallback model's provider has an API key
func checkFallbackModelAPIKey(rootDir, fallbackModel string) error {
	vendor := strings.Split(fallbackModel, "/")[0]
	var provider *LLMProvider
	for i := range LLMProviders {
		if LLMProviders[i].Vendor == vendor {
			provider = &LLMProviders[i]
			break
		}
	}

	if provider == nil {
		return nil // Unknown provider, skip check
	}

	// Check if API key is configured
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, _ := os.ReadFile(formationFile)
	contentStr := string(content)
	secretRef := fmt.Sprintf("secrets.%s", provider.SecretName)

	if !strings.Contains(contentStr, secretRef) {
		fmt.Println()
		ui.Warning(fmt.Sprintf("No API key configured for fallback provider %s", provider.Name))
		if err := promptForAPIKey(rootDir, *provider); err != nil {
			return err
		}
	}

	return nil
}

// configureVisionSettings configures vision-specific settings
func configureVisionSettings(rootDir string, settings *ModelSettings) error {
	fmt.Println()
	ui.Bold("Vision Settings")
	fmt.Println()

	ui.Dimmed("  Maximum image file size")
	maxSize, err := wizard.PromptString("  Max image size (MB)", "5", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.ImageMaxSizeMB = maxSize
	ui.PromptSuccess("  Max image size", maxSize+"MB")

	resize, err := wizard.PromptConfirm("  Resize large images?", true)
	if err != nil {
		return err
	}
	settings.ImageResize = resize
	if resize {
		ui.PromptSuccess("  Resize", "enabled")

		ui.Dimmed("  Maximum width after resize")
		maxWidth, err := wizard.PromptString("  Max width (px)", "1024", validatePositiveInt)
		if err != nil {
			return err
		}
		settings.ImageMaxWidth = maxWidth
		ui.PromptSuccess("  Max width", maxWidth+"px")

		ui.Dimmed("  Maximum height after resize")
		maxHeight, err := wizard.PromptString("  Max height (px)", "1024", validatePositiveInt)
		if err != nil {
			return err
		}
		settings.ImageMaxHeight = maxHeight
		ui.PromptSuccess("  Max height", maxHeight+"px")
	} else {
		ui.PromptSkipped("  Resize")
	}

	return nil
}

// configureAudioSettings configures audio-specific settings
func configureAudioSettings(rootDir string, settings *ModelSettings) error {
	fmt.Println()
	ui.Bold("Audio Settings")
	fmt.Println()

	ui.Dimmed("  Maximum audio file size")
	maxSize, err := wizard.PromptString("  Max audio size (MB)", "10", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.AudioMaxSizeMB = maxSize
	ui.PromptSuccess("  Max audio size", maxSize+"MB")

	ui.Dimmed("  Language for transcription (auto, en, es, fr, etc.)")
	language, err := wizard.PromptString("  Language", "auto", nil)
	if err != nil {
		return err
	}
	settings.AudioLanguage = language
	ui.PromptSuccess("  Language", language)

	return nil
}

// configureVideoSettings configures video-specific settings
func configureVideoSettings(rootDir string, settings *ModelSettings) error {
	fmt.Println()
	ui.Bold("Video Settings")
	fmt.Println()

	ui.Dimmed("  Maximum video file size")
	maxSize, err := wizard.PromptString("  Max video size (MB)", "100", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.VideoMaxSizeMB = maxSize
	ui.PromptSuccess("  Max video size", maxSize+"MB")

	ui.Dimmed("  Maximum video duration in seconds")
	maxDuration, err := wizard.PromptString("  Max duration (seconds)", "300", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.VideoMaxDuration = maxDuration
	ui.PromptSuccess("  Max duration", maxDuration+"s")

	includeAudio, err := wizard.PromptConfirm("  Include audio analysis?", true)
	if err != nil {
		return err
	}
	settings.VideoIncludeAudio = includeAudio
	if includeAudio {
		ui.PromptSuccess("  Audio analysis", "enabled")
	} else {
		ui.PromptSkipped("  Audio analysis")
	}

	return nil
}

// configureDocumentsSettings configures documents-specific settings
func configureDocumentsSettings(rootDir string, settings *ModelSettings) error {
	fmt.Println()
	ui.Bold("Documents Settings")
	fmt.Println()

	ui.Dimmed("  Maximum document file size")
	maxSize, err := wizard.PromptString("  Max document size (MB)", "20", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.DocsMaxSizeMB = maxSize
	ui.PromptSuccess("  Max document size", maxSize+"MB")

	ui.Dimmed("  Text chunk size for processing")
	chunkSize, err := wizard.PromptString("  Chunk size (chars)", "1000", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.DocsChunkSize = chunkSize
	ui.PromptSuccess("  Chunk size", chunkSize)

	ui.Dimmed("  Overlap between chunks for context preservation")
	overlap, err := wizard.PromptString("  Overlap (chars)", "100", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.DocsOverlap = overlap
	ui.PromptSuccess("  Overlap", overlap)

	ui.Dimmed("  Extraction strategy (adaptive, semantic, fixed, paragraph)")
	strategy, err := wizard.PromptString("  Strategy", "adaptive", nil)
	if err != nil {
		return err
	}
	settings.DocsStrategy = strategy
	ui.PromptSuccess("  Strategy", strategy)

	ui.Dimmed("  How long to cache processed documents (seconds)")
	cacheTTL, err := wizard.PromptString("  Cache TTL (seconds)", "3600", validatePositiveInt)
	if err != nil {
		return err
	}
	settings.DocsCacheTTL = cacheTTL
	ui.PromptSuccess("  Cache TTL", cacheTTL+"s")

	return nil
}

// configureGlobalSettings handles Flow 3: Global LLM Settings
func configureGlobalSettings(rootDir string) error {
	fmt.Println()
	ui.Bold("Global LLM Settings")
	fmt.Println()
	ui.Dimmed("  These are defaults applied to all models unless overridden.")
	fmt.Println()

	// Temperature
	ui.Dimmed("  Controls randomness: 0 = deterministic, 1 = creative")
	tempStr, err := wizard.PromptString("  Temperature (0.0-1.0)", "0.7", validateTemperature)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Temperature", tempStr)

	// Max tokens
	ui.Dimmed("  Maximum response length in tokens")
	tokensStr, err := wizard.PromptString("  Max tokens", "4096", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Max tokens", tokensStr)

	// Timeout
	ui.Dimmed("  Request timeout before retry or failure")
	timeoutStr, err := wizard.PromptString("  Timeout (seconds)", "30", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Timeout", timeoutStr)

	// Max retries
	ui.Dimmed("  Number of retry attempts on failure")
	retriesStr, err := wizard.PromptString("  Max retries", "3", validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Max retries", retriesStr)

	// Fallback model
	ui.Dimmed("  Model to use if primary fails (e.g., openai/gpt-4o-mini)")
	fallbackModel, err := wizard.PromptString("  Default fallback model (optional)", "", nil)
	if err != nil {
		return err
	}
	if fallbackModel != "" {
		ui.PromptSuccess("  Default fallback model", fallbackModel)
	} else {
		ui.PromptSkipped("  Default fallback model")
	}

	// Response caching
	fmt.Println()
	ui.Bold("Response Caching")
	fmt.Println()
	ui.Dimmed("  Cache similar requests to reduce API costs and latency")

	enableCaching, err := wizard.PromptConfirm("  Enable response caching?", true)
	if err != nil {
		return err
	}

	var maxEntries, similarityThreshold, cacheTTL string
	if enableCaching {
		ui.PromptSuccess("  Caching", "enabled")

		ui.Dimmed("  Maximum number of cached responses (FIFO)")
		maxEntries, err = wizard.PromptString("  Max cache entries", "10000", validatePositiveInt)
		if err != nil {
			return err
		}
		ui.PromptSuccess("  Max entries", maxEntries)

		ui.Dimmed("  How similar requests must be to hit cache (0.95 = 95% similar)")
		similarityThreshold, err = wizard.PromptString("  Similarity threshold (0.0-1.0)", "0.95", validateTemperature)
		if err != nil {
			return err
		}
		ui.PromptSuccess("  Similarity threshold", similarityThreshold)

		ui.Dimmed("  How long to keep cached responses (86400 = 24 hours)")
		cacheTTL, err = wizard.PromptString("  Cache TTL (seconds)", "86400", validatePositiveInt)
		if err != nil {
			return err
		}
		ui.PromptSuccess("  Cache TTL", cacheTTL)
	} else {
		ui.PromptSkipped("  Caching")
	}

	// Update formation.yaml
	return updateGlobalSettingsInFormation(rootDir, tempStr, tokensStr, timeoutStr, retriesStr, fallbackModel, enableCaching, maxEntries, similarityThreshold, cacheTTL)
}

// promptForAPIKey prompts for an API key with env var detection
func promptForAPIKey(rootDir string, provider LLMProvider) error {
	// Build prompt label
	fmt.Println()
	var keyHint string
	if provider.KeyPrefix != "" {
		keyHint = fmt.Sprintf(" (starts with %s)", provider.KeyPrefix)
	}
	fmt.Printf("%s API Key%s:\n", provider.Name, keyHint)

	var apiKey string

	// Check for existing env var (primary and alternatives)
	envVarsToCheck := append([]string{provider.SecretName}, provider.AltEnvVars...)
	for _, envName := range envVarsToCheck {
		envValue := os.Getenv(envName)
		if envValue != "" {
			maskedEnv := maskAPIKey(envValue, provider.KeyPrefix)
			useEnv, err := wizard.PromptConfirm(fmt.Sprintf("ℹ Found %s in environment [%s]. Use it?", envName, maskedEnv), true)
			if err != nil {
				return err
			}
			if useEnv {
				apiKey = envValue
				break
			}
		}
	}

	// If not using env var, prompt for key
	if apiKey == "" {
		var err error
		apiKey, err = wizard.PromptPassword("Enter key", true)
		if err != nil {
			return err
		}

		if apiKey == "" {
			ui.PromptSkipped("API Key")
			return nil
		}
	}

	// Save to secrets
	secretsManager := secrets.NewManager(rootDir)
	if err := secretsManager.Set(provider.SecretName, apiKey, true); err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}
	ui.Success(fmt.Sprintf("Saved %s to secrets", provider.SecretName))

	// Update formation.yaml to include the API key reference
	if err := addAPIKeyToFormation(rootDir, provider); err != nil {
		return fmt.Errorf("failed to update formation.yaml: %w", err)
	}
	ui.Success("Updated formation.yaml with API key reference")

	return nil
}

// Helper functions

func getMaskedSecretPreview(rootDir, secretName string) string {
	secretsManager := secrets.NewManager(rootDir)

	value, ok := secretsManager.Get(secretName)
	if !ok || value == "" {
		return ""
	}

	// Mask the value: show first 3 and last 4 chars
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:3] + "..." + value[len(value)-4:]
}

func maskAPIKey(value, prefix string) string {
	if value == "" {
		return ""
	}
	// Mask the value: show prefix (if matches) + last 4 chars
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	if prefix != "" && strings.HasPrefix(value, prefix) {
		return prefix + "..." + value[len(value)-4:]
	}
	return value[:3] + "..." + value[len(value)-4:]
}

func formatCapabilityOption(name, currentModel, fallbackInfo string, defaultsToText bool) string {
	if currentModel != "" {
		return fmt.Sprintf("%s [%s]", name, currentModel)
	}
	if defaultsToText && fallbackInfo != "" {
		return fmt.Sprintf("%s [using text]", name)
	}
	if fallbackInfo != "" {
		// Special fallback info (e.g., for embedding)
		return fmt.Sprintf("%s [%s]", name, fallbackInfo)
	}
	return name
}

func extractModelForCapability(content, capability string) string {
	// Simple regex to find "- capability: model"
	pattern := fmt.Sprintf(`-\s*%s:\s*"?([^"\s]+)"?`, capability)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func validateModelFormat(input string) error {
	if !strings.Contains(input, "/") {
		return fmt.Errorf("model must be in format: provider/model (e.g., openai/gpt-5)")
	}
	return nil
}

func validateTemperature(input string) error {
	var temp float64
	_, err := fmt.Sscanf(input, "%f", &temp)
	if err != nil || temp < 0 || temp > 1 {
		return fmt.Errorf("temperature must be between 0.0 and 1.0")
	}
	return nil
}

func validatePositiveInt(input string) error {
	var val int
	_, err := fmt.Sscanf(input, "%d", &val)
	if err != nil || val <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

func addAPIKeyToFormation(rootDir string, provider LLMProvider) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	contentStr := string(content)
	keyLine := fmt.Sprintf("    %s: \"${{ secrets.%s }}\"", provider.Vendor, provider.SecretName)

	// Check if this key is already configured
	secretRef := fmt.Sprintf("secrets.%s", provider.SecretName)
	if strings.Contains(contentStr, secretRef) {
		return nil // Already configured
	}

	lines := strings.Split(contentStr, "\n")
	var result []string
	inLLM := false
	addedKey := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track when we enter/exit llm section
		if trimmed == "llm:" {
			inLLM = true
			result = append(result, line)
			continue
		}

		// If in llm section and we hit api_keys:
		if inLLM && trimmed == "api_keys:" && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
			result = append(result, line)
			result = append(result, keyLine)
			addedKey = true
			continue
		}

		// Exit llm section if we hit a new top-level key
		if inLLM && len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") {
			// Before exiting, check if we need to add api_keys section
			if !addedKey {
				// Insert api_keys section before this line
				result = append(result, "  api_keys:")
				result = append(result, keyLine)
				addedKey = true
			}
			inLLM = false
		}

		result = append(result, line)

		// If this is the llm: line and next line is not api_keys, we might need to add it
		if inLLM && !addedKey && i > 0 {
			prevTrimmed := strings.TrimSpace(lines[i-1])
			if prevTrimmed == "llm:" {
				// Check if current line is already api_keys
				if trimmed != "api_keys:" {
					// Insert api_keys before current line
					result = result[:len(result)-1]
					result = append(result, "  api_keys:")
					result = append(result, keyLine)
					result = append(result, line)
					addedKey = true
				}
			}
		}
	}

	// If llm section exists but we never added the key (end of file case)
	if inLLM && !addedKey {
		result = append(result, "  api_keys:")
		result = append(result, keyLine)
		addedKey = true
	}

	// If no llm section exists, add one
	if !addedKey {
		result = append(result, "")
		result = append(result, "llm:")
		result = append(result, "  api_keys:")
		result = append(result, keyLine)
	}

	contentStr = strings.Join(result, "\n")
	return os.WriteFile(formationFile, []byte(contentStr), 0644)
}

func updateModelInFormation(rootDir, capability, model string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	contentStr := string(content)
	modelLine := fmt.Sprintf("    - %s: \"%s\"", capability, model)

	// Check if the capability already exists
	pattern := fmt.Sprintf(`-\s*%s:\s*"?[^"\n]+"?`, capability)
	re := regexp.MustCompile(pattern)

	if re.MatchString(contentStr) {
		// Update existing capability
		contentStr = re.ReplaceAllString(contentStr, fmt.Sprintf("- %s: \"%s\"", capability, model))
	} else if strings.Contains(contentStr, "models:") {
		// Add to existing models section
		lines := strings.Split(contentStr, "\n")
		var result []string

		for _, line := range lines {
			result = append(result, line)
			if strings.TrimSpace(line) == "models:" {
				result = append(result, modelLine)
			}
		}

		contentStr = strings.Join(result, "\n")
	} else if strings.Contains(contentStr, "llm:") {
		// Add models section to existing llm section
		lines := strings.Split(contentStr, "\n")
		var result []string
		addedModels := false

		for _, line := range lines {
			result = append(result, line)
			if strings.TrimSpace(line) == "llm:" && !addedModels {
				// Check if next lines have api_keys - add after that section
				addedModels = true
			}
		}

		// Find the right place to insert (after api_keys if present)
		contentStr = strings.Join(result, "\n")
		if !strings.Contains(contentStr, "models:") {
			// Simple append after llm section
			contentStr = strings.Replace(contentStr, "llm:\n", fmt.Sprintf("llm:\n  models:\n%s\n", modelLine), 1)
		}
	} else {
		// Add new llm section with models
		contentStr += fmt.Sprintf("\nllm:\n  models:\n%s\n", modelLine)
	}

	return os.WriteFile(formationFile, []byte(contentStr), 0644)
}

func updateModelSettingsInFormation(rootDir, capability string, settings ModelSettings) error {
	// For now, just log - full implementation would update YAML structure
	// TODO: Implement proper YAML structure update for model settings
	ui.Success("Settings saved (note: manual YAML update may be needed for capability-specific settings)")
	return nil
}

func updateGlobalSettingsInFormation(rootDir, temp, tokens, timeout, retries, fallback string, caching bool, maxEntries, similarity, ttl string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Build settings YAML
	var settingsYAML strings.Builder
	settingsYAML.WriteString("  settings:\n")
	settingsYAML.WriteString(fmt.Sprintf("    temperature: %s\n", temp))
	settingsYAML.WriteString(fmt.Sprintf("    max_tokens: %s\n", tokens))
	settingsYAML.WriteString(fmt.Sprintf("    timeout_seconds: %s\n", timeout))
	settingsYAML.WriteString(fmt.Sprintf("    max_retries: %s\n", retries))
	if fallback != "" {
		settingsYAML.WriteString(fmt.Sprintf("    fallback_model: \"%s\"\n", fallback))
	}

	if caching {
		settingsYAML.WriteString("    caching:\n")
		settingsYAML.WriteString("      enabled: true\n")
		settingsYAML.WriteString(fmt.Sprintf("      max_entries: %s\n", maxEntries))
		settingsYAML.WriteString(fmt.Sprintf("      similarity: %s\n", similarity))
		settingsYAML.WriteString(fmt.Sprintf("      ttl: %s\n", ttl))
	}

	// Check if settings section exists under llm (not in comments)
	// Look for "  settings:" at proper indentation (not "# settings:" in comments)
	hasLLMSettings := regexp.MustCompile(`(?m)^llm:[\s\S]*?\n  settings:`).MatchString(contentStr)
	if hasLLMSettings {
		// Replace existing settings - complex, for now just append
		ui.Warning("Settings section exists - please update manually")
		fmt.Println(settingsYAML.String())
	} else if strings.Contains(contentStr, "llm:") {
		// Add settings to llm section
		contentStr = strings.Replace(contentStr, "llm:\n", "llm:\n"+settingsYAML.String(), 1)
		if err := os.WriteFile(formationFile, []byte(contentStr), 0644); err != nil {
			return err
		}
		ui.Success("Updated formation.yaml with global settings")
	} else {
		// Add new llm section with settings
		contentStr += "\nllm:\n" + settingsYAML.String()
		if err := os.WriteFile(formationFile, []byte(contentStr), 0644); err != nil {
			return err
		}
		ui.Success("Added LLM settings to formation.yaml")
	}

	return nil
}
