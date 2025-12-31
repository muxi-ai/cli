package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientFromEntry(t *testing.T) {
	entry := &ProfileEntry{
		URL:       "http://localhost:7890",
		KeyID:     "test-key-id",
		SecretKey: "test-secret",
	}

	client := NewClientFromEntry(entry)

	if client.BaseURL != "http://localhost:7890" {
		t.Errorf("BaseURL = %q, want 'http://localhost:7890'", client.BaseURL)
	}
	if client.KeyID != "test-key-id" {
		t.Errorf("KeyID = %q, want 'test-key-id'", client.KeyID)
	}
	if client.SecretKey != "test-secret" {
		t.Errorf("SecretKey = %q, want 'test-secret'", client.SecretKey)
	}
	if client.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func mockServerClient(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := &Client{
		BaseURL:    server.URL,
		KeyID:      "test-key",
		SecretKey:  "test-secret",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	return server, client
}

func TestClientGet(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// Check auth header exists
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("Authorization header should be set")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	defer server.Close()

	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestClientPost(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want 'application/json'", contentType)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	resp, err := client.Post("/test", nil, "application/json")
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestClientPut(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	resp, err := client.Put("/test", nil, "")
	if err != nil {
		t.Fatalf("Put() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestClientDelete(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	resp, err := client.Delete("/test")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestPing(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ping" {
			t.Errorf("expected path /ping, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("1704067200000")) // Unix millis
	})
	defer server.Close()

	_, err := client.Ping()
	if err != nil {
		t.Fatalf("Ping() error: %v", err)
	}
}

func TestHealth(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("expected path /health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"status":     "ok",
				"formations": 2,
			},
		})
	})
	defer server.Close()

	health, err := client.Health()
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if health.Data.Status != "ok" {
		t.Errorf("Data.Status = %q, want 'ok'", health.Data.Status)
	}
}

func TestListFormations(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/formations" {
			t.Errorf("expected path /rpc/formations, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"formations": []map[string]interface{}{
					{"id": "formation-1", "name": "Test Formation"},
				},
				"total": 1,
			},
		})
	})
	defer server.Close()

	resp, err := client.ListFormations()
	if err != nil {
		t.Fatalf("ListFormations() error: %v", err)
	}
	if len(resp.Formations) != 1 {
		t.Errorf("len(formations) = %d, want 1", len(resp.Formations))
	}
}

func TestGetFormation(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/formations/test-formation" {
			t.Errorf("expected path /rpc/formations/test-formation, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":     "test-formation",
				"name":   "Test Formation",
				"status": "running",
			},
		})
	})
	defer server.Close()

	formation, err := client.GetFormation("test-formation")
	if err != nil {
		t.Fatalf("GetFormation() error: %v", err)
	}
	if formation.ID != "test-formation" {
		t.Errorf("ID = %q, want 'test-formation'", formation.ID)
	}
}

func TestStartFormation(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/formations/test-formation/start" {
			t.Errorf("expected path /rpc/formations/test-formation/start, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.StartFormation("test-formation")
	if err != nil {
		t.Fatalf("StartFormation() error: %v", err)
	}
}

func TestStopFormation(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/formations/test-formation/stop" {
			t.Errorf("expected path /rpc/formations/test-formation/stop, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.StopFormation("test-formation")
	if err != nil {
		t.Fatalf("StopFormation() error: %v", err)
	}
}

func TestRestartFormation(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/formations/test-formation/restart" {
			t.Errorf("expected path /rpc/formations/test-formation/restart, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.RestartFormation("test-formation")
	if err != nil {
		t.Fatalf("RestartFormation() error: %v", err)
	}
}

func TestDeleteFormation(t *testing.T) {
	server, client := mockServerClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/formations/test-formation" {
			t.Errorf("expected path /rpc/formations/test-formation, got %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.DeleteFormation("test-formation")
	if err != nil {
		t.Fatalf("DeleteFormation() error: %v", err)
	}
}
