package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func mockRegistryClient(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := &Client{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	return server, client
}

func TestIsAuthenticated(t *testing.T) {
	client := &Client{Token: "test-token"}
	if !client.IsAuthenticated() {
		t.Error("IsAuthenticated() should return true when token is set")
	}

	client = &Client{Token: ""}
	if client.IsAuthenticated() {
		t.Error("IsAuthenticated() should return false when token is empty")
	}
}

func TestGetAuthURL(t *testing.T) {
	client := &Client{BaseURL: "https://registry.example.com"}

	// With callback port
	url := client.GetAuthURL(8080)
	if url != "https://registry.example.com/auth/cli/authorize?callback=http://localhost:8080/auth" {
		t.Errorf("GetAuthURL(8080) = %q, unexpected", url)
	}

	// Without callback port
	url = client.GetAuthURL(0)
	if url != "https://registry.example.com/auth/cli/authorize" {
		t.Errorf("GetAuthURL(0) = %q, unexpected", url)
	}
}

func TestValidateToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/auth/validate" {
				t.Errorf("expected path /api/auth/validate, got %s", r.URL.Path)
			}

			auth := r.Header.Get("Authorization")
			if auth != "Bearer valid-token" {
				t.Errorf("Authorization = %q, want 'Bearer valid-token'", auth)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
			})
		})
		defer server.Close()

		info, err := client.ValidateToken("valid-token")
		if err != nil {
			t.Fatalf("ValidateToken() error: %v", err)
		}
		if info.Username != "testuser" {
			t.Errorf("Username = %q, want 'testuser'", info.Username)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
		defer server.Close()

		_, err := client.ValidateToken("invalid-token")
		if err == nil {
			t.Error("ValidateToken() should return error for invalid token")
		}
	})
}

func TestMyFormations(t *testing.T) {
	server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/formations" {
			t.Errorf("expected path /api/formations, got %s", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want 'Bearer test-token'", auth)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"formations": []map[string]interface{}{
				{"name": "test-formation", "owner": "testuser"},
			},
		})
	})
	defer server.Close()

	result, err := client.MyFormations()
	if err != nil {
		t.Fatalf("MyFormations() error: %v", err)
	}
	if len(result.Formations) != 1 {
		t.Errorf("len(formations) = %d, want 1", len(result.Formations))
	}
}

func TestGetFormation(t *testing.T) {
	server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/formations/@testuser/test-formation" {
			t.Errorf("expected path /api/formations/@testuser/test-formation, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":        "test-formation",
			"owner":       "testuser",
			"description": "Test description",
		})
	})
	defer server.Close()

	formation, err := client.GetFormation("testuser/test-formation", false)
	if err != nil {
		t.Fatalf("GetFormation() error: %v", err)
	}
	if formation.Name != "test-formation" {
		t.Errorf("Name = %q, want 'test-formation'", formation.Name)
	}
}

func TestGetVersions(t *testing.T) {
	server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
		// Path uses @owner format
		if r.URL.Path != "/api/formations/@testuser/test-formation/versions" {
			t.Errorf("expected path /api/formations/@testuser/test-formation/versions, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		// Returns array directly
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"version": "1.0.0", "published_at": "2024-01-01"},
		})
	})
	defer server.Close()

	versions, err := client.GetVersions("testuser/test-formation")
	if err != nil {
		t.Fatalf("GetVersions() error: %v", err)
	}
	if len(versions) != 1 {
		t.Errorf("len(versions) = %d, want 1", len(versions))
	}
}

func TestSearch(t *testing.T) {
	server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search" {
			t.Errorf("expected path /api/search, got %s", r.URL.Path)
		}

		query := r.URL.Query().Get("q")
		if query != "test" {
			t.Errorf("query = %q, want 'test'", query)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"name": "test-formation", "user": "testuser"},
			},
			"total": 1,
		})
	})
	defer server.Close()

	result, err := client.Search("test", "", 10)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(result.Results) != 1 {
		t.Errorf("len(results) = %d, want 1", len(result.Results))
	}
}

func TestGetPullInfo(t *testing.T) {
	server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
		// Path uses @ prefix for user
		if r.URL.Path != "/api/formations/@testuser/test-formation" {
			t.Errorf("expected path /api/formations/@testuser/test-formation, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"download_url": "https://example.com/download",
			"version":      "1.0.0",
		})
	})
	defer server.Close()

	result, err := client.GetPullInfo("testuser/test-formation")
	if err != nil {
		t.Fatalf("GetPullInfo() error: %v", err)
	}
	if result.DownloadURL != "https://example.com/download" {
		t.Errorf("DownloadURL = %q, want 'https://example.com/download'", result.DownloadURL)
	}
}

func TestDownloadFormation(t *testing.T) {
	// Create a test server that returns a zip file
	server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		// Write minimal zip header for testing
		w.Write([]byte("PK\x03\x04"))
	})
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := tmpDir + "/test.zip"

	err := client.DownloadFormation(server.URL+"/download", destPath)
	if err != nil {
		t.Fatalf("DownloadFormation() error: %v", err)
	}
}

func TestPublish(t *testing.T) {
	server, client := mockRegistryClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Formation published",
			"formation": map[string]interface{}{
				"name":    "test-formation",
				"version": "1.0.0",
				"user":    "testuser",
			},
		})
	})
	defer server.Close()

	// Create a temp zip file
	tmpDir := t.TempDir()
	zipPath := tmpDir + "/test.zip"
	os.WriteFile(zipPath, []byte("PK\x03\x04"), 0644)

	result, err := client.Publish(zipPath, "testorg")
	if err != nil {
		t.Fatalf("Publish() error: %v", err)
	}
	if result.Formation.Name != "test-formation" {
		t.Errorf("Formation.Name = %q, want 'test-formation'", result.Formation.Name)
	}
}
