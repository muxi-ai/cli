// Package secrets provides encryption and decryption for MUXI formation secrets.
// It implements Fernet-compatible encryption matching the Python runtime.
package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fernet/fernet-go"
)

// Manager handles secrets encryption and storage for a formation.
type Manager struct {
	FormationDir    string
	KeyPath         string
	SecretsPath     string
	SecretsTemplate string
	key             *fernet.Key
	cache           map[string]string
}

// NewManager creates a new secrets manager for the given formation directory.
func NewManager(formationDir string) *Manager {
	return &Manager{
		FormationDir:    formationDir,
		KeyPath:         filepath.Join(formationDir, ".key"),
		SecretsPath:     filepath.Join(formationDir, "secrets.enc"),
		SecretsTemplate: filepath.Join(formationDir, "secrets"),
		cache:           make(map[string]string),
	}
}

// Initialize sets up encryption (creates key if needed) and loads existing secrets.
func (m *Manager) Initialize() error {
	// Load or create master key
	if err := m.loadOrCreateKey(); err != nil {
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}

	// Load existing secrets if file exists
	if _, err := os.Stat(m.SecretsPath); err == nil {
		if err := m.loadSecrets(); err != nil {
			return fmt.Errorf("failed to load secrets: %w", err)
		}
	}

	return nil
}

// loadOrCreateKey loads existing key or creates a new one.
func (m *Manager) loadOrCreateKey() error {
	if _, err := os.Stat(m.KeyPath); err == nil {
		// Load existing key
		keyData, err := os.ReadFile(m.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
		key, err := fernet.DecodeKey(string(keyData))
		if err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}
		m.key = key
	} else {
		// Generate new key
		key := new(fernet.Key)
		if err := key.Generate(); err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}
		m.key = key

		// Write key to file with restrictive permissions
		if err := os.WriteFile(m.KeyPath, []byte(key.Encode()), 0600); err != nil {
			return fmt.Errorf("failed to write key file: %w", err)
		}
	}
	return nil
}

// loadSecrets decrypts and loads secrets from file.
func (m *Manager) loadSecrets() error {
	encryptedData, err := os.ReadFile(m.SecretsPath)
	if err != nil {
		return fmt.Errorf("failed to read secrets file: %w", err)
	}

	decryptedData := fernet.VerifyAndDecrypt(encryptedData, 0, []*fernet.Key{m.key})
	if decryptedData == nil {
		return fmt.Errorf("failed to decrypt secrets (invalid key or corrupted data)")
	}

	if err := json.Unmarshal(decryptedData, &m.cache); err != nil {
		return fmt.Errorf("failed to parse secrets JSON: %w", err)
	}

	return nil
}

// saveSecrets encrypts and saves secrets to file.
func (m *Manager) saveSecrets() error {
	data, err := json.MarshalIndent(m.cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize secrets: %w", err)
	}

	encryptedData, err := fernet.EncryptAndSign(data, m.key)
	if err != nil {
		return fmt.Errorf("failed to encrypt secrets: %w", err)
	}

	if err := os.WriteFile(m.SecretsPath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}

	// Update secrets template file
	m.updateSecretsTemplate()

	return nil
}

// updateSecretsTemplate updates the secrets template file with all keys.
func (m *Manager) updateSecretsTemplate() {
	var keys []string
	for k := range m.cache {
		keys = append(keys, k)
	}

	var lines []string
	for _, k := range keys {
		lines = append(lines, k+"=")
	}

	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}

	os.WriteFile(m.SecretsTemplate, []byte(content), 0644)
}

// NormalizeName converts a secret name to uppercase with underscores.
func NormalizeName(name string) string {
	// Convert to uppercase
	name = strings.ToUpper(name)
	// Replace non-alphanumeric chars (except underscore) with underscore
	re := regexp.MustCompile(`[^A-Z0-9_]`)
	name = re.ReplaceAllString(name, "_")
	// Remove multiple consecutive underscores
	re = regexp.MustCompile(`_+`)
	name = re.ReplaceAllString(name, "_")
	// Trim leading/trailing underscores
	name = strings.Trim(name, "_")
	return name
}

// Set stores a secret (overwrites if exists when overwrite=true).
func (m *Manager) Set(name, value string, overwrite bool) error {
	if m.key == nil {
		if err := m.Initialize(); err != nil {
			return err
		}
	}

	normalizedName := NormalizeName(name)

	if _, exists := m.cache[normalizedName]; exists && !overwrite {
		return fmt.Errorf("secret '%s' already exists (use overwrite to replace)", normalizedName)
	}

	m.cache[normalizedName] = value
	return m.saveSecrets()
}

// Get retrieves a secret by name.
func (m *Manager) Get(name string) (string, bool) {
	if m.key == nil {
		if err := m.Initialize(); err != nil {
			return "", false
		}
	}

	normalizedName := NormalizeName(name)
	value, exists := m.cache[normalizedName]
	return value, exists
}

// Delete removes a secret by name.
func (m *Manager) Delete(name string) (bool, error) {
	if m.key == nil {
		if err := m.Initialize(); err != nil {
			return false, err
		}
	}

	normalizedName := NormalizeName(name)
	if _, exists := m.cache[normalizedName]; !exists {
		return false, nil
	}

	delete(m.cache, normalizedName)
	return true, m.saveSecrets()
}

// List returns all secret names.
func (m *Manager) List() ([]string, error) {
	if m.key == nil {
		if err := m.Initialize(); err != nil {
			return nil, err
		}
	}

	var names []string
	for k := range m.cache {
		names = append(names, k)
	}
	return names, nil
}

// Exists checks if a secret exists.
func (m *Manager) Exists(name string) bool {
	_, exists := m.Get(name)
	return exists
}

// Import loads secrets from a map (for bulk import).
func (m *Manager) Import(secrets map[string]string, overwrite bool) error {
	if m.key == nil {
		if err := m.Initialize(); err != nil {
			return err
		}
	}

	for name, value := range secrets {
		normalizedName := NormalizeName(name)
		if _, exists := m.cache[normalizedName]; exists && !overwrite {
			continue // Skip existing if not overwriting
		}
		m.cache[normalizedName] = value
	}

	return m.saveSecrets()
}

// Export returns all secrets as a map.
func (m *Manager) Export() (map[string]string, error) {
	if m.key == nil {
		if err := m.Initialize(); err != nil {
			return nil, err
		}
	}

	// Return a copy
	result := make(map[string]string)
	for k, v := range m.cache {
		result[k] = v
	}
	return result, nil
}

// Clear removes all secrets.
func (m *Manager) Clear() error {
	if m.key == nil {
		if err := m.Initialize(); err != nil {
			return err
		}
	}

	m.cache = make(map[string]string)
	return m.saveSecrets()
}

// GenerateKey creates a new random Fernet key (base64 encoded).
func GenerateKey() (string, error) {
	key := new(fernet.Key)
	if err := key.Generate(); err != nil {
		return "", err
	}
	return key.Encode(), nil
}

// Encrypt encrypts data with the given key.
func Encrypt(data []byte, keyStr string) ([]byte, error) {
	key, err := fernet.DecodeKey(keyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}
	return fernet.EncryptAndSign(data, key)
}

// Decrypt decrypts data with the given key.
func Decrypt(data []byte, keyStr string) ([]byte, error) {
	key, err := fernet.DecodeKey(keyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}
	result := fernet.VerifyAndDecrypt(data, 0, []*fernet.Key{key})
	if result == nil {
		return nil, fmt.Errorf("decryption failed (invalid key or corrupted data)")
	}
	return result, nil
}

// InterpolateSecrets replaces ${{ secrets.NAME }} patterns in a string.
func (m *Manager) InterpolateSecrets(text string) (string, error) {
	if m.key == nil {
		if err := m.Initialize(); err != nil {
			return "", err
		}
	}

	// Pattern: ${{ secrets.NAME }} with flexible whitespace
	re := regexp.MustCompile(`\$\{\{\s*secrets\.([A-Z0-9_]+)\s*\}\}`)

	var lastErr error
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		// Extract secret name
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		secretName := submatches[1]

		value, exists := m.cache[NormalizeName(secretName)]
		if !exists {
			lastErr = fmt.Errorf("secret '%s' not found", secretName)
			return match
		}
		return value
	})

	if lastErr != nil {
		return "", lastErr
	}
	return result, nil
}

// ParseSecretsFile parses a KEY=VALUE format file (like .env).
func ParseSecretsFile(content string) map[string]string {
	secrets := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first =
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Skip if value is empty (template file)
		if value == "" {
			continue
		}

		secrets[key] = value
	}

	return secrets
}

// GenerateRandomBytes generates cryptographically secure random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString generates a random base64-encoded string.
func GenerateRandomString(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetTemplateKeys reads the secrets template file and returns all keys.
func (m *Manager) GetTemplateKeys() ([]string, error) {
	content, err := os.ReadFile(m.SecretsTemplate)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read secrets template: %w", err)
	}

	var keys []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Extract key (everything before =)
		parts := strings.SplitN(line, "=", 2)
		if len(parts) >= 1 && parts[0] != "" {
			keys = append(keys, strings.TrimSpace(parts[0]))
		}
	}
	return keys, nil
}

// UpdateTemplate updates the secrets template file with the given keys.
func (m *Manager) UpdateTemplate(keys []string) error {
	var lines []string
	for _, k := range keys {
		lines = append(lines, k+"=")
	}

	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}

	return os.WriteFile(m.SecretsTemplate, []byte(content), 0644)
}

// ScanFormationFiles scans all formation YAML/MD files for ${{ secrets.XXX }} patterns.
// Returns a slice of unique secret names found.
func ScanFormationFiles(rootDir string) ([]string, error) {
	// Pattern: ${{ secrets.NAME }} with flexible whitespace and case
	re := regexp.MustCompile(`\$\{\{\s*secrets\.([A-Za-z0-9_]+)\s*\}\}`)

	foundSecrets := make(map[string]bool)

	// Directories and file patterns to scan
	scanPaths := []struct {
		dir     string
		pattern string
	}{
		{"", "formation.afs"},
		{"", "formation.yaml"},
		{"agents", "*.afs"},
		{"agents", "*.yaml"},
		{"mcps", "*.afs"},
		{"mcps", "*.yaml"},
		{"a2a", "*.afs"},
		{"a2a", "*.yaml"},
		{"sops", "*.md"},
		{"triggers", "*.md"},
	}

	for _, sp := range scanPaths {
		searchDir := rootDir
		if sp.dir != "" {
			searchDir = filepath.Join(rootDir, sp.dir)
		}

		// Check if directory exists
		if sp.dir != "" {
			if _, err := os.Stat(searchDir); os.IsNotExist(err) {
				continue
			}
		}

		// Find matching files
		var files []string
		if sp.dir == "" {
			// Single file in root
			filePath := filepath.Join(rootDir, sp.pattern)
			if _, err := os.Stat(filePath); err == nil {
				files = append(files, filePath)
			}
		} else {
			// Glob pattern in subdirectory
			matches, err := filepath.Glob(filepath.Join(searchDir, sp.pattern))
			if err != nil {
				continue
			}
			files = matches
		}

		// Scan each file
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			// Filter out comment lines before scanning
			var activeLines []string
			for _, line := range strings.Split(string(content), "\n") {
				trimmed := strings.TrimSpace(line)
				if !strings.HasPrefix(trimmed, "#") {
					activeLines = append(activeLines, line)
				}
			}
			filteredContent := strings.Join(activeLines, "\n")

			// Find all matches in non-commented content
			matches := re.FindAllStringSubmatch(filteredContent, -1)
			for _, match := range matches {
				if len(match) >= 2 {
					// Normalize the secret name
					secretName := NormalizeName(match[1])
					foundSecrets[secretName] = true
				}
			}
		}
	}

	// Convert map to slice
	var result []string
	for name := range foundSecrets {
		result = append(result, name)
	}

	return result, nil
}

// DeleteFromTemplate removes a key from the secrets template file.
func (m *Manager) DeleteFromTemplate(name string) error {
	keys, err := m.GetTemplateKeys()
	if err != nil {
		return err
	}

	// Filter out the key to delete
	var newKeys []string
	for _, k := range keys {
		if k != name {
			newKeys = append(newKeys, k)
		}
	}

	return m.UpdateTemplate(newKeys)
}

// AddToTemplate adds a key to the secrets template file if not already present.
func (m *Manager) AddToTemplate(name string) error {
	keys, err := m.GetTemplateKeys()
	if err != nil {
		return err
	}

	// Check if already exists
	for _, k := range keys {
		if k == name {
			return nil // Already exists
		}
	}

	keys = append(keys, name)
	return m.UpdateTemplate(keys)
}
