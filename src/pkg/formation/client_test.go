package formation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockServer creates a test server with the given handler
func mockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient(server.URL, "test-admin-key", "test-client-key")
	return server, client
}

func TestHealth(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("expected path /health, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		
		// Health endpoint returns plain JSON (no envelope)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"version": "1.0.0",
		})
	})
	defer server.Close()

	resp, err := client.Health()
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if resp.Status != "healthy" {
		t.Errorf("Status = %q, want 'healthy'", resp.Status)
	}
}

func TestGetStatus(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			t.Errorf("expected path /status, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"formation": map[string]interface{}{
					"id":      "test-formation",
					"name":    "Test Formation",
					"version": "1.0.0",
				},
				"server": map[string]interface{}{
					"version":        "1.0.0",
					"uptime_seconds": 3600,
				},
			},
		})
	})
	defer server.Close()

	resp, err := client.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus() error: %v", err)
	}
	if resp.Formation.ID != "test-formation" {
		t.Errorf("Formation.ID = %q, want 'test-formation'", resp.Formation.ID)
	}
}

func TestGetConfig(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config" {
			t.Errorf("expected path /config, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"formation_id": "test-formation",
				"version":      "1.0.0",
			},
		})
	})
	defer server.Close()

	resp, err := client.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error: %v", err)
	}
	if resp.FormationID != "test-formation" {
		t.Errorf("FormationID = %q, want 'test-formation'", resp.FormationID)
	}
}

func TestGetAgents(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents" {
			t.Errorf("expected path /agents, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"agents": []map[string]interface{}{
					{"id": "agent-1", "name": "Weather Bot"},
					{"id": "agent-2", "name": "Helper Bot"},
				},
				"count": 2,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetAgents()
	if err != nil {
		t.Fatalf("GetAgents() error: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("Count = %d, want 2", resp.Count)
	}
}

func TestGetSecrets(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/secrets" {
			t.Errorf("expected path /secrets, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"secrets": map[string]string{
					"API_KEY":      "sk-****",
					"SECRET_TOKEN": "tok-****",
				},
				"count": 2,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetSecrets()
	if err != nil {
		t.Fatalf("GetSecrets() error: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("Count = %d, want 2", resp.Count)
	}
}

func TestGetTriggers(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/triggers" {
			t.Errorf("expected path /triggers, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"triggers": []string{"on-message", "on-join"},
				"count":    2,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetTriggers()
	if err != nil {
		t.Fatalf("GetTriggers() error: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("Count = %d, want 2", resp.Count)
	}
}

func TestGetSOPs(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sops" {
			t.Errorf("expected path /sops, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"sops":  []map[string]interface{}{},
				"count": 0,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetSOPs()
	if err != nil {
		t.Fatalf("GetSOPs() error: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("Count = %d, want 0", resp.Count)
	}
}

func TestGetSessions(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions" {
			t.Errorf("expected path /sessions, got %s", r.URL.Path)
		}
		
		// Check user header
		userID := r.Header.Get("X-Muxi-User-ID")
		if userID != "alice" {
			t.Errorf("X-Muxi-User-ID = %q, want 'alice'", userID)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"sessions": []map[string]interface{}{
					{"session_id": "sess_123", "user_id": "alice"},
				},
				"count": 1,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetSessions("alice", 10)
	if err != nil {
		t.Fatalf("GetSessions() error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("Count = %d, want 1", resp.Count)
	}
}

func TestGetRequests(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/requests" {
			t.Errorf("expected path /requests, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"requests": []map[string]interface{}{
					{"request_id": "req_123", "status": "completed"},
				},
				"count": 1,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetRequests("alice")
	if err != nil {
		t.Fatalf("GetRequests() error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("Count = %d, want 1", resp.Count)
	}
}

func TestCancelRequest(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/requests/req_123" {
			t.Errorf("expected path /requests/req_123, got %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    map[string]interface{}{},
		})
	})
	defer server.Close()

	err := client.CancelRequest("req_123", "alice")
	if err != nil {
		t.Fatalf("CancelRequest() error: %v", err)
	}
}

func TestGetAuditLog(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audit" {
			t.Errorf("expected path /audit, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"entries": []map[string]interface{}{},
				"count":   0,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetAuditLog()
	if err != nil {
		t.Fatalf("GetAuditLog() error: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("Count = %d, want 0", resp.Count)
	}
}

func TestClearAuditLog(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audit" {
			t.Errorf("expected path /audit, got %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    map[string]interface{}{},
		})
	})
	defer server.Close()

	err := client.ClearAuditLog()
	if err != nil {
		t.Fatalf("ClearAuditLog() error: %v", err)
	}
}

func TestChat(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat" {
			t.Errorf("expected path /chat, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Message != "Hello" {
			t.Errorf("Message = %q, want 'Hello'", req.Message)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"response":   "Hi there!",
				"session_id": "sess_123",
			},
		})
	})
	defer server.Close()

	req := &ChatRequest{Message: "Hello"}
	resp, err := client.Chat(req)
	if err != nil {
		t.Fatalf("Chat() error: %v", err)
	}
	if resp.Response != "Hi there!" {
		t.Errorf("Response = %q, want 'Hi there!'", resp.Response)
	}
}

func TestSetSecret(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/secrets/API_KEY" {
			t.Errorf("expected path /secrets/API_KEY, got %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    map[string]interface{}{},
		})
	})
	defer server.Close()

	err := client.SetSecret("API_KEY", "secret-value")
	if err != nil {
		t.Fatalf("SetSecret() error: %v", err)
	}
}

func TestDeleteSecret(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/secrets/API_KEY" {
			t.Errorf("expected path /secrets/API_KEY, got %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    map[string]interface{}{},
		})
	})
	defer server.Close()

	err := client.DeleteSecret("API_KEY")
	if err != nil {
		t.Fatalf("DeleteSecret() error: %v", err)
	}
}

func TestHTTPErrors(t *testing.T) {
	t.Run("500 error", func(t *testing.T) {
		server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error": map[string]interface{}{
					"message": "Internal server error",
				},
			})
		})
		defer server.Close()

		_, err := client.Health()
		if err == nil {
			t.Error("expected error for 500 response")
		}
	})

	t.Run("401 unauthorized", func(t *testing.T) {
		server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error": map[string]interface{}{
					"message": "Unauthorized",
				},
			})
		})
		defer server.Close()

		_, err := client.GetConfig()
		if err == nil {
			t.Error("expected error for 401 response")
		}
	})
}

func TestAPIKeyHeaders(t *testing.T) {
	t.Run("admin key header", func(t *testing.T) {
		server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
			adminKey := r.Header.Get("X-MUXI-ADMIN-KEY")
			if adminKey != "test-admin-key" {
				t.Errorf("X-MUXI-ADMIN-KEY = %q, want 'test-admin-key'", adminKey)
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    map[string]interface{}{},
			})
		})
		defer server.Close()

		client.GetConfig()
	})

	t.Run("client key header", func(t *testing.T) {
		server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
			clientKey := r.Header.Get("X-MUXI-CLIENT-KEY")
			if clientKey != "test-client-key" {
				t.Errorf("X-MUXI-CLIENT-KEY = %q, want 'test-client-key'", clientKey)
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"triggers": []string{},
					"count":    0,
				},
			})
		})
		defer server.Close()

		// GetTriggers uses client key (via GetClient)
		client.GetTriggers()
	})
}
