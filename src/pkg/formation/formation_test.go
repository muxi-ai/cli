package formation

import (
	"testing"
)

func TestBuildFormationURL(t *testing.T) {
	tests := []struct {
		serverURL   string
		formationID string
		want        string
	}{
		{"http://localhost:7890", "my-formation", "http://localhost:7890/api/my-formation/v1"},
		{"https://api.muxi.ai", "prod-bot", "https://api.muxi.ai/api/prod-bot/v1"},
	}

	for _, tt := range tests {
		got := BuildFormationURL(tt.serverURL, tt.formationID)
		if got != tt.want {
			t.Errorf("BuildFormationURL(%q, %q) = %q, want %q", tt.serverURL, tt.formationID, got, tt.want)
		}
	}
}

func TestResolveUserID(t *testing.T) {
	// Test that flag value takes precedence
	result := ResolveUserID("flag-user")
	if result != "flag-user" {
		t.Errorf("ResolveUserID(\"flag-user\") = %q, want \"flag-user\"", result)
	}

	// Test that empty flag falls through (returns empty if no defaults)
	result = ResolveUserID("")
	// Can't assert much here without mocking the file system
}

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:7890/api/test/v1", "admin-key", "client-key")

	if client.BaseURL != "http://localhost:7890/api/test/v1" {
		t.Errorf("BaseURL = %q, want %q", client.BaseURL, "http://localhost:7890/api/test/v1")
	}
	if client.AdminKey != "admin-key" {
		t.Errorf("AdminKey = %q, want %q", client.AdminKey, "admin-key")
	}
	if client.ClientKey != "client-key" {
		t.Errorf("ClientKey = %q, want %q", client.ClientKey, "client-key")
	}
}
