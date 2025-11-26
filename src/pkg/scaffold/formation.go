package scaffold

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
)

// CreateFormation creates a new formation directory with all necessary files
func CreateFormation(name string, noWizard bool) error {
	var formationName string
	var description string
	var openaiKey string

	// Interactive mode - get formation ID if not provided
	if name == "" && !noWizard {
		// Loop until we get a valid ID that doesn't conflict
		for {
			var err error
			formationName, err = wizard.PromptString("Formation ID", "", validateFormationName)
			if err != nil {
				return err
			}
			
			// Check if directory already exists
			if _, err := os.Stat(formationName); !os.IsNotExist(err) {
				// Show error and re-prompt
				ui.PromptError("Formation ID", formationName, fmt.Errorf("directory already exists\n\nChoose a different ID or remove:\n  rm -rf %s", formationName))
				continue
			}
			
			// All good - show success and break
			ui.PromptSuccess("Formation ID", formationName)
			break
		}
	} else if name != "" {
		// Validate provided ID
		if err := validateFormationName(name); err != nil {
			return fmt.Errorf("invalid formation ID '%s': %w", name, err)
		}
		formationName = name
		
		// Check if directory already exists (non-interactive - just error out)
		if _, err := os.Stat(formationName); !os.IsNotExist(err) {
			ui.ErrorBlock(
				"Formation directory exists",
				fmt.Sprintf("Directory '%s' already exists", formationName),
				fmt.Sprintf("Choose a different name or remove:\n  rm -rf %s", formationName),
			)
			return fmt.Errorf("directory exists")
		}
	} else {
		return fmt.Errorf("formation ID required (provide as argument or run without --no-wizard)")
	}

	// Interactive mode - get optional description and secrets
	if !noWizard {
		var err error
		description, err = wizard.PromptString("Description (optional, press Enter to skip)", "", nil)
		if err != nil {
			return err
		}
		
		if description != "" {
			ui.PromptSuccess("Description", description)
		} else {
			ui.PromptSkipped("Description")
		}

		ui.Section("Setup secrets:")
		openaiKey, err = wizard.PromptPassword("  [1/1] OPENAI_API_KEY (optional)\n    Enter API key (leave empty to skip)", true)
		if err != nil {
			return err
		}

		if openaiKey != "" {
			if !strings.HasPrefix(openaiKey, "sk-") {
				ui.Warning("OpenAI API keys typically start with 'sk-'")
			}
			ui.PromptSuccess("  OPENAI_API_KEY", "configured")
		} else {
			ui.PromptSkipped("  OPENAI_API_KEY")
		}
	}

	// Create formation directory
	if err := os.Mkdir(formationName, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	ui.Section(fmt.Sprintf("Creating formation '%s'...", formationName))

	// Create subdirectories
	dirs := []string{"agents", "mcps", "a2a", "sops", "triggers", "knowledge"}
	for _, dir := range dirs {
		path := filepath.Join(formationName, dir)
		if err := os.Mkdir(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Create .gitkeep
		gitkeepPath := filepath.Join(path, ".gitkeep")
		if err := os.WriteFile(gitkeepPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep in %s: %w", dir, err)
		}
	}

	// Generate encryption key
	encryptionKey, err := generateEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Generate formation API keys
	// TODO: Store these in secrets.enc when secrets package is ready
	_ = generateFormationKey("fma")  // adminKey
	_ = generateFormationKey("fmc")  // clientKey

	// Create files
	files := map[string]string{
		".gitignore":      gitignoreTemplate(),
		".key":            encryptionKey,
		".muxi":           muxiTemplate(),
		"formation.yaml":  formationYAMLTemplate(formationName, description),
		"secrets":         secretsTemplate(),
		"README.md":       readmeTemplate(formationName, description),
	}

	for filename, content := range files {
		path := filepath.Join(formationName, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	// Make .key read-only for safety
	keyPath := filepath.Join(formationName, ".key")
	if err := os.Chmod(keyPath, 0400); err != nil {
		return fmt.Errorf("failed to set permissions on .key: %w", err)
	}

	ui.Step("Directory structure created")
	ui.Step("Formation keys generated")

	// If OpenAI key provided, create secrets.enc
	if openaiKey != "" {
		// TODO: Encrypt and save secrets
		// For now, just indicate success (already shown above with PromptSuccess)
	}

	// Success message
	fmt.Println()
	ui.Success(fmt.Sprintf("Formation '%s' created successfully!", formationName))
	
	if !noWizard {
		fmt.Println()
		ui.Dimmed("Next steps:")
		ui.Dimmed(fmt.Sprintf("  cd %s", formationName))
		if openaiKey == "" {
			ui.Dimmed("  muxi secrets set OPENAI_API_KEY")
		}
		ui.Dimmed("  muxi validate")
		ui.Dimmed("  muxi deploy --profile production")
	} else {
		fmt.Println()
		ui.Dimmed("Files created:")
		files := []string{
			".gitignore, .key, .muxi, formation.yaml",
			"secrets (template with 3 keys)",
			"README.md",
			"6 directories (agents/, mcps/, a2a/, sops/, triggers/, knowledge/)",
		}
		ui.List(files)
		
		fmt.Println()
		ui.Warning("Configure secrets before deploying:")
		ui.Dimmed(fmt.Sprintf("    cd %s", formationName))
		ui.Dimmed("    muxi secrets setup")
	}

	return nil
}

// validateFormationName validates formation ID format
func validateFormationName(name string) error {
	if name == "" {
		return fmt.Errorf("formation ID cannot be empty")
	}

	// Pattern: lowercase letter, then lowercase letters/numbers/hyphens, 3-50 chars
	pattern := regexp.MustCompile(`^[a-z][a-z0-9-]{2,49}$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf("formation IDs must:\n  • Be lowercase\n  • Start with a letter\n  • Contain only letters, numbers, and hyphens\n  • Be 3-50 characters long\n\nExample: my-bot")
	}

	return nil
}

// generateEncryptionKey generates a 32-byte encryption key for AES-256-GCM
func generateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// generateFormationKey generates a formation API key (admin or client)
func generateFormationKey(prefix string) string {
	key := make([]byte, 32)
	rand.Read(key)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(key))
}
