package scaffold

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/secrets"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
)

// FormationConfig holds all configuration gathered during the wizard
type FormationConfig struct {
	Name             string
	Description      string
	EnableStreaming  bool
	EnableAsync      bool
	WebhookURL       string
	ProviderType     string // "cloud", "local", "enterprise"
	Provider         *LLMProvider
	LocalProvider    *LocalProvider
	LocalBaseURL     string
	LocalModel       string
	EnterpriseProvider *EnterpriseProvider
	APIKey           string
}

// CreateFormation creates a new formation directory with all necessary files
func CreateFormation(name string, noWizard bool) error {
	config := &FormationConfig{}

	// Check if we're already inside a formation directory
	if _, err := context.DetectFormation(); err == nil {
		ui.ErrorBlock(
			"Already in formation directory",
			"Cannot create a formation inside another formation.",
			"Navigate outside the current formation:\n  cd ..\n  muxi new formation",
		)
		os.Exit(1)
	}

	// Show logo and banner in interactive mode
	if !noWizard {
		fmt.Println()
		ui.Gold(`  ███╗   ███╗██╗   ██╗██╗  ██╗██╗
  ████╗ ████║██║   ██║╚██╗██╔╝██║
  ██╔████╔██║██║   ██║ ╚███╔╝ ██║
  ██║╚██╔╝██║██║   ██║ ██╔██╗ ██║
  ██║ ╚═╝ ██║╚██████╔╝██╔╝ ██╗██║
  ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝`)
		ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [+] Creating new formation                              MUXI │\n│──────────────────────────────────────────────────────────────│\n│ ℹ A formation is a deployable unit containing agents, MCPs,  │\n│ SOPs, and configuration for your AI system.                  │\n╰──────────────────────────────────────────────────────────────╯")
	}

	// Step 1: Get formation ID
	if name == "" && !noWizard {
		for {
			inputName, err := wizard.PromptString("Formation ID", "", nil)
			if err != nil {
				return err
			}
			
			config.Name = normalizeFormationName(inputName)
			
			if err := validateFormationName(config.Name); err != nil {
				ui.PromptError("Formation ID", inputName, err)
				continue
			}
			
			if _, err := os.Stat(config.Name); !os.IsNotExist(err) {
				ui.PromptError("Formation ID", inputName, fmt.Errorf("directory already exists\n\nChoose a different ID or remove:\n  rm -rf %s", config.Name))
				continue
			}
			
			ui.PromptSuccess("Formation ID", config.Name)
			break
		}
	} else if name != "" {
		config.Name = normalizeFormationName(name)
		
		if err := validateFormationName(config.Name); err != nil {
			return fmt.Errorf("invalid formation ID '%s': %w", name, err)
		}
		
		if _, err := os.Stat(config.Name); !os.IsNotExist(err) {
			ui.ErrorBlock(
				"Formation directory exists",
				fmt.Sprintf("Directory '%s' already exists", config.Name),
				fmt.Sprintf("Choose a different name or remove:\n  rm -rf %s", config.Name),
			)
			return fmt.Errorf("directory exists")
		}
	} else {
		return fmt.Errorf("formation ID required (provide as argument or run without --no-wizard)")
	}

	// Interactive mode - gather all configuration
	if !noWizard {
		var err error
		
		// Step 2: Description
		config.Description, err = wizard.PromptString("Description (optional, press Enter to skip)", "", nil)
		if err != nil {
			return err
		}
		if config.Description != "" {
			ui.PromptSuccess("Description", config.Description)
		} else {
			ui.PromptSkipped("Description")
		}

		// Step 3: Streaming responses
		config.EnableStreaming, err = wizard.PromptConfirm("Enable streaming responses?", false)
		if err != nil {
			return err
		}
		if config.EnableStreaming {
			ui.PromptSuccess("Streaming", "enabled")
		} else {
			ui.PromptSkipped("Streaming")
		}

		// Step 4: Async responses
		config.EnableAsync, err = wizard.PromptConfirm("Enable async responses for long-running tasks?", false)
		if err != nil {
			return err
		}
		
		if config.EnableAsync {
			// Get webhook URL (required for async)
			fmt.Println()
			ui.Dimmed("  Async responses are delivered via webhook for long-running tasks.")
			fmt.Println()
			
			for {
				webhookURL, err := wizard.PromptString("Webhook URL", "", nil)
				if err != nil {
					return err
				}
				
				if webhookURL == "" {
					// Empty URL - ask if they want to disable async
					disable, err := wizard.PromptConfirm("Disable async responses?", false)
					if err != nil {
						return err
					}
					if disable {
						config.EnableAsync = false
						ui.PromptSkipped("Async")
						break
					}
					continue
				}
				
				// Auto-prepend https:// if no protocol
				if !strings.HasPrefix(webhookURL, "http://") && !strings.HasPrefix(webhookURL, "https://") {
					webhookURL = "https://" + webhookURL
				}
				
				config.WebhookURL = webhookURL
				ui.PromptSuccess("Webhook URL", webhookURL)
				break
			}
		} else {
			ui.PromptSkipped("Async")
		}

		// Step 5: LLM Provider selection
		fmt.Println()
		ui.Section("LLM Provider Setup")
		ui.Dimmed("You need at least one LLM provider for the formation to work.")
		ui.Dimmed("You can add more later using 'muxi config llm'.")
		fmt.Println()
		
		if err := promptLLMProvider(config); err != nil {
			return err
		}
	}

	// Create formation directory
	if err := os.Mkdir(config.Name, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	ui.Section(fmt.Sprintf("Creating formation '%s'...", config.Name))

	// Create subdirectories
	dirs := []string{"agents", "mcps", "a2a", "sops", "triggers", "knowledge"}
	for _, dir := range dirs {
		path := filepath.Join(config.Name, dir)
		if err := os.Mkdir(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		gitkeepPath := filepath.Join(path, ".gitkeep")
		if err := os.WriteFile(gitkeepPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep in %s: %w", dir, err)
		}
	}

	// Initialize secrets manager (creates .key with Fernet encryption)
	secretsMgr := secrets.NewManager(config.Name)
	if err := secretsMgr.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize secrets: %w", err)
	}

	// Store API key in encrypted secrets if provided
	if config.APIKey != "" && config.Provider != nil {
		if err := secretsMgr.Set(config.Provider.SecretName, config.APIKey, true); err != nil {
			return fmt.Errorf("failed to store API key: %w", err)
		}
		ui.Step("API key encrypted and stored")
	} else {
		// Create secrets template file with expected keys (even if no values yet)
		secretsTemplate := generateSecretsTemplate(config)
		secretsPath := filepath.Join(config.Name, "secrets")
		if err := os.WriteFile(secretsPath, []byte(secretsTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create secrets template: %w", err)
		}
	}

	// Make .key read-only for safety
	keyPath := filepath.Join(config.Name, ".key")
	if err := os.Chmod(keyPath, 0400); err != nil {
		return fmt.Errorf("failed to set permissions on .key: %w", err)
	}

	// Generate formation API keys
	_ = generateFormationKey("fma")
	_ = generateFormationKey("fmc")

	// Create files with dynamic content based on config
	files := map[string]string{
		".gitignore":     gitignoreTemplate(),
		".muxi":          muxiTemplate(),
		"formation.yaml": generateFormationYAML(config),
		"README.md":      readmeTemplate(config.Name, config.Description),
	}

	for filename, content := range files {
		path := filepath.Join(config.Name, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	ui.Step("Directory structure created")
	ui.Step("Formation keys generated")

	// Success message
	fmt.Println()
	ui.Success(fmt.Sprintf("Formation '%s' created successfully!", config.Name))
	
	// Show next steps based on provider type
	if !noWizard {
		fmt.Println()
		
		if config.ProviderType == "enterprise" && config.EnterpriseProvider != nil {
			ui.Dimmed("Next steps:")
			for i, step := range config.EnterpriseProvider.NextSteps {
				ui.Dimmed(fmt.Sprintf("  %d. %s", i+1, step))
			}
		} else {
			ui.Dimmed("Next steps:")
			ui.Dimmed(fmt.Sprintf("  cd %s", config.Name))
			if config.APIKey == "" && config.ProviderType != "local" {
				if config.Provider != nil {
					ui.Dimmed(fmt.Sprintf("  muxi secrets set %s", config.Provider.SecretName))
				}
			}
			ui.Dimmed("  muxi new agent       # add agents")
			ui.Dimmed("  muxi new mcp         # add tools")
			ui.Dimmed("  muxi config overlord # configure orchestration")
			ui.Dimmed("  muxi validate")
			ui.Dimmed("  muxi deploy")
		}
	} else {
		fmt.Println()
		ui.Dimmed("Files created:")
		fileList := []string{
			".gitignore, .key, .muxi, formation.yaml",
			"secrets (template)",
			"README.md",
			"6 directories (agents/, mcps/, a2a/, sops/, triggers/, knowledge/)",
		}
		ui.List(fileList)
		
		fmt.Println()
		ui.Warning("Configure secrets before deploying:")
		ui.Dimmed(fmt.Sprintf("    cd %s", config.Name))
		ui.Dimmed("    muxi secrets setup")
	}

	return nil
}

// promptLLMProvider handles the LLM provider selection wizard
func promptLLMProvider(config *FormationConfig) error {
	// Build the options list for PromptSelect
	var options []wizard.SelectOption
	
	// Cloud providers (1-17)
	for _, p := range LLMProviders {
		options = append(options, wizard.SelectOption{
			Value: fmt.Sprintf("cloud:%d", len(options)),
			Label: p.Name,
		})
	}
	
	// Local provider
	options = append(options, wizard.SelectOption{
		Value: "local",
		Label: "Local (Ollama/llama_cpp)",
	})
	
	// Enterprise providers
	for i, p := range EnterpriseProviders {
		options = append(options, wizard.SelectOption{
			Value: fmt.Sprintf("enterprise:%d", i),
			Label: p.Name,
		})
	}
	
	// Show selection
	selection, err := wizard.PromptSelect("Select provider", options, 0)
	if err != nil {
		return err
	}
	
	// Parse selection
	if strings.HasPrefix(selection, "cloud:") {
		idx, _ := strconv.Atoi(strings.TrimPrefix(selection, "cloud:"))
		provider := LLMProviders[idx]
		config.ProviderType = "cloud"
		config.Provider = &provider
		
		// Show which provider was selected
		ui.PromptSuccess("Provider", provider.Name)
		
		// Prompt for API key (visible while typing, shown as *** after)
		fmt.Println()
		apiKey, err := wizard.PromptString(fmt.Sprintf("%s API Key", provider.Name), "", nil)
		if err != nil {
			return err
		}
		
		if apiKey != "" {
			config.APIKey = apiKey
			ui.PromptSuccess("API Key", "***")
			ui.PromptSuccess("Model", fmt.Sprintf("%s/%s", provider.Vendor, provider.DefaultModel))
		} else {
			ui.PromptSkipped("API Key")
			ui.PromptSuccess("Model", fmt.Sprintf("%s/%s (add API key later)", provider.Vendor, provider.DefaultModel))
		}
		return nil
	}
	
	if selection == "local" {
		ui.PromptSuccess("Provider", "Local")
		config.ProviderType = "local"
		return promptLocalProvider(config)
	}
	
	if strings.HasPrefix(selection, "enterprise:") {
		idx, _ := strconv.Atoi(strings.TrimPrefix(selection, "enterprise:"))
		provider := EnterpriseProviders[idx]
		config.ProviderType = "enterprise"
		config.EnterpriseProvider = &provider
		
		fmt.Println()
		ui.Success(fmt.Sprintf("%s template added to formation.yaml", provider.Name))
		return nil
	}
	
	return fmt.Errorf("invalid selection")
}

// promptLocalProvider handles local provider configuration
func promptLocalProvider(config *FormationConfig) error {
	fmt.Println()
	fmt.Println("Local LLM Setup")
	fmt.Println("───────────────")
	fmt.Println()
	ui.Dimmed("For Ollama, llama.cpp, or other local inference servers.")
	fmt.Println()
	
	// Select local provider type
	fmt.Println("Provider:")
	for i, p := range LocalProviders {
		fmt.Printf("  [%d] %s (default: %s)\n", i+1, p.Name, p.DefaultURL)
	}
	fmt.Println()
	
	for {
		selection, err := wizard.PromptString(fmt.Sprintf("Select (1-%d)", len(LocalProviders)), "", nil)
		if err != nil {
			return err
		}
		
		num, err := strconv.Atoi(selection)
		if err != nil || num < 1 || num > len(LocalProviders) {
			ui.PromptError("Selection", selection, fmt.Errorf("please enter 1 or 2"))
			continue
		}
		
		localProvider := LocalProviders[num-1]
		config.LocalProvider = &localProvider
		
		// Get base URL
		baseURL, err := wizard.PromptString(fmt.Sprintf("Base URL [%s]", localProvider.DefaultURL), localProvider.DefaultURL, nil)
		if err != nil {
			return err
		}
		if baseURL == "" {
			baseURL = localProvider.DefaultURL
		}
		config.LocalBaseURL = baseURL
		
		// Get model name
		modelName, err := wizard.PromptString("Model name (e.g., llama3, mistral, phi3)", "", nil)
		if err != nil {
			return err
		}
		if modelName == "" {
			modelName = "llama3"
		}
		config.LocalModel = modelName
		
		fmt.Println()
		ui.Success(fmt.Sprintf("Local LLM configured: %s/%s at %s", localProvider.Vendor, modelName, baseURL))
		return nil
	}
}

// normalizeFormationName converts user input to valid formation name format
// Converts spaces to hyphens, lowercases, removes extra hyphens
func normalizeFormationName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	
	// Replace multiple hyphens with single hyphen
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	
	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")
	
	return name
}

// validateFormationName validates formation ID format
func validateFormationName(name string) error {
	if name == "" {
		return fmt.Errorf("formation ID cannot be empty")
	}

	// Pattern: lowercase letter, then lowercase letters/numbers/hyphens, 3-50 chars
	pattern := regexp.MustCompile(`^[a-z][a-z0-9-]{2,49}$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf("formation IDs must:\n  • Start with a letter\n  • Contain only letters, numbers, and hyphens\n  • Be 3-50 characters long\n\nExample: my-bot")
	}

	return nil
}

// generateFormationKey generates a formation API key (admin or client)
func generateFormationKey(prefix string) string {
	key := make([]byte, 32)
	rand.Read(key)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(key))
}
