package cmd

import (
	"testing"
)

// executeCommand runs a cobra command - just verify no panic
func executeCommand(args ...string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func TestCommandsNoPanic(t *testing.T) {
	// Test that various commands don't panic
	commands := [][]string{
		{"--help"},
		{"remote", "--help"},
		{"secrets", "--help"},
		{"config", "--help"},
		{"new", "--help"},
		{"registry", "--help"},
		{"deploy", "--help"},
		{"chat", "--help"},
		{"validate", "--help"},
		{"bump", "--help"},
		{"agents", "--help"},
		{"sessions", "--help"},
		{"triggers", "--help"},
		{"mcp", "--help"},
		{"sops", "--help"},
		{"users", "--help"},
		{"logs", "--help"},
		{"requests", "--help"},
		{"memory", "--help"},
		{"history", "--help"},
		{"scheduler", "--help"},
		{"credentials", "--help"},
	}

	for _, args := range commands {
		t.Run(args[0], func(t *testing.T) {
			// Just verify no panic - ignore errors (expected for most)
			_ = executeCommand(args...)
		})
	}
}

func TestRequireArgs(t *testing.T) {
	validator := RequireArgs(2)

	// Test with enough args
	err := validator(rootCmd, []string{"arg1", "arg2"})
	if err != nil {
		t.Error("RequireArgs(2) should pass with 2 args")
	}

	// Test with not enough args - returns empty error after showing help
	err = validator(rootCmd, []string{"arg1"})
	if err == nil {
		t.Error("RequireArgs(2) should fail with 1 arg")
	}
}

func TestFormationMetadataFields(t *testing.T) {
	meta := FormationMetadata{
		ID:      "test-id",
		Name:    "Test Formation",
		Version: "1.0.0",
	}

	if meta.ID != "test-id" {
		t.Errorf("ID = %q, want 'test-id'", meta.ID)
	}
	if meta.Version != "1.0.0" {
		t.Errorf("Version = %q, want '1.0.0'", meta.Version)
	}
	if meta.Name != "Test Formation" {
		t.Errorf("Name = %q, want 'Test Formation'", meta.Name)
	}
}
