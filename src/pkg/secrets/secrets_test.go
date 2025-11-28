package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"openai-api-key", "OPENAI_API_KEY"},
		{"database_url", "DATABASE_URL"},
		{"MySecret123", "MYSECRET123"},
		{"test--double", "TEST_DOUBLE"},
		{"_leading_", "LEADING"},
		{"ALREADY_UPPER", "ALREADY_UPPER"},
	}

	for _, tt := range tests {
		result := NormalizeName(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestManagerBasicOperations(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "secrets-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager
	mgr := NewManager(tmpDir)

	// Initialize
	if err := mgr.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify key file was created
	if _, err := os.Stat(filepath.Join(tmpDir, ".key")); os.IsNotExist(err) {
		t.Error("Key file was not created")
	}

	// Set a secret
	if err := mgr.Set("test-key", "test-value", false); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the secret
	value, exists := mgr.Get("test-key")
	if !exists {
		t.Error("Secret not found after Set")
	}
	if value != "test-value" {
		t.Errorf("Got %q, want %q", value, "test-value")
	}

	// Verify secrets.enc was created
	if _, err := os.Stat(filepath.Join(tmpDir, "secrets.enc")); os.IsNotExist(err) {
		t.Error("secrets.enc was not created")
	}

	// List secrets
	names, err := mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(names) != 1 || names[0] != "TEST_KEY" {
		t.Errorf("List returned %v, want [TEST_KEY]", names)
	}

	// Delete secret
	deleted, err := mgr.Delete("test-key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !deleted {
		t.Error("Delete returned false")
	}

	// Verify deletion
	_, exists = mgr.Get("test-key")
	if exists {
		t.Error("Secret still exists after Delete")
	}
}

func TestManagerPersistence(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "secrets-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and initialize first manager
	mgr1 := NewManager(tmpDir)
	if err := mgr1.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Set secrets
	mgr1.Set("api-key", "sk-1234567890", false)
	mgr1.Set("database-url", "postgres://localhost/test", false)

	// Create new manager (simulates restart)
	mgr2 := NewManager(tmpDir)
	if err := mgr2.Initialize(); err != nil {
		t.Fatalf("Second initialize failed: %v", err)
	}

	// Verify secrets persisted
	val1, _ := mgr2.Get("api-key")
	val2, _ := mgr2.Get("database-url")

	if val1 != "sk-1234567890" {
		t.Errorf("api-key = %q, want %q", val1, "sk-1234567890")
	}
	if val2 != "postgres://localhost/test" {
		t.Errorf("database-url = %q, want %q", val2, "postgres://localhost/test")
	}
}

func TestInterpolateSecrets(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "secrets-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager and set secrets
	mgr := NewManager(tmpDir)
	mgr.Initialize()
	mgr.Set("API_KEY", "sk-secret-key", false)
	mgr.Set("TOKEN", "bearer-token-123", false)

	tests := []struct {
		input    string
		expected string
	}{
		{"${{ secrets.API_KEY }}", "sk-secret-key"},
		{"Bearer ${{ secrets.TOKEN }}", "Bearer bearer-token-123"},
		{"Key: ${{ secrets.API_KEY }}, Token: ${{ secrets.TOKEN }}", "Key: sk-secret-key, Token: bearer-token-123"},
		{"No secrets here", "No secrets here"},
		{"${{secrets.API_KEY}}", "sk-secret-key"}, // No spaces
	}

	for _, tt := range tests {
		result, err := mgr.InterpolateSecrets(tt.input)
		if err != nil {
			t.Errorf("InterpolateSecrets(%q) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("InterpolateSecrets(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseSecretsFile(t *testing.T) {
	content := `# Comment
OPENAI_API_KEY=sk-1234567890
DATABASE_URL=postgres://localhost/test

# Another comment
EMPTY_VALUE=
GITHUB_TOKEN=ghp_abcdef123456
`

	secrets := ParseSecretsFile(content)

	if len(secrets) != 3 {
		t.Errorf("Got %d secrets, want 3", len(secrets))
	}

	if secrets["OPENAI_API_KEY"] != "sk-1234567890" {
		t.Errorf("OPENAI_API_KEY = %q", secrets["OPENAI_API_KEY"])
	}
	if secrets["DATABASE_URL"] != "postgres://localhost/test" {
		t.Errorf("DATABASE_URL = %q", secrets["DATABASE_URL"])
	}
	if secrets["GITHUB_TOKEN"] != "ghp_abcdef123456" {
		t.Errorf("GITHUB_TOKEN = %q", secrets["GITHUB_TOKEN"])
	}
	// EMPTY_VALUE should be skipped
	if _, exists := secrets["EMPTY_VALUE"]; exists {
		t.Error("EMPTY_VALUE should be skipped")
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Fernet keys are 32 bytes base64 encoded = 44 chars with padding
	if len(key) != 44 {
		t.Errorf("Key length = %d, want 44", len(key))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("Hello, MUXI secrets!")

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted = %q, want %q", decrypted, plaintext)
	}
}
