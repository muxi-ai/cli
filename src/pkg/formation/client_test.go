package formation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	if resp.GetResponseText() != "Hi there!" {
		t.Errorf("Response = %q, want 'Hi there!'", resp.GetResponseText())
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

func TestGetAgent(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/test-agent" {
			t.Errorf("expected path /agents/test-agent, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":   "test-agent",
				"name": "Test Agent",
			},
		})
	})
	defer server.Close()

	result, err := client.GetAgent("test-agent")
	if err != nil {
		t.Fatalf("GetAgent() error: %v", err)
	}
	if result["id"] != "test-agent" {
		t.Errorf("id = %v, want 'test-agent'", result["id"])
	}
}

func TestGetMCPServers(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp/servers" {
			t.Errorf("expected path /mcp/servers, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"servers": []map[string]interface{}{
					{"id": "mcp-1", "name": "MCP Server"},
				},
				"count": 1,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetMCPServers()
	if err != nil {
		t.Fatalf("GetMCPServers() error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("Count = %d, want 1", resp.Count)
	}
}

func TestGetTrigger(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/triggers/on-message" {
			t.Errorf("expected path /triggers/on-message, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"name":    "on-message",
				"enabled": true,
			},
		})
	})
	defer server.Close()

	result, err := client.GetTrigger("on-message")
	if err != nil {
		t.Fatalf("GetTrigger() error: %v", err)
	}
	if result.Name != "on-message" {
		t.Errorf("Name = %q, want 'on-message'", result.Name)
	}
}

func TestGetSOP(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sops/test-sop" {
			t.Errorf("expected path /sops/test-sop, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"name":        "test-sop",
				"description": "Test SOP",
			},
		})
	})
	defer server.Close()

	result, err := client.GetSOP("test-sop")
	if err != nil {
		t.Fatalf("GetSOP() error: %v", err)
	}
	if result.Name != "test-sop" {
		t.Errorf("Name = %q, want 'test-sop'", result.Name)
	}
}

func TestGetSessionMessages(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
				"count": 1,
			},
		})
	})
	defer server.Close()

	resp, err := client.GetSessionMessages("session-123", "user-1")
	if err != nil {
		t.Fatalf("GetSessionMessages() error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("Count = %d, want 1", resp.Count)
	}
}

func TestLinkUserIdentifier(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    map[string]interface{}{},
		})
	})
	defer server.Close()

	err := client.LinkUserIdentifier("email@test.com", "user-123", "email")
	if err != nil {
		t.Fatalf("LinkUserIdentifier() error: %v", err)
	}
}

func TestUnlinkUserIdentifier(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
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

	err := client.UnlinkUserIdentifier("email@test.com")
	if err != nil {
		t.Fatalf("UnlinkUserIdentifier() error: %v", err)
	}
}

func TestResolveUserIdentifier(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"identifier": "email@test.com",
				"user_id":    "user-123",
				"type":       "email",
			},
		})
	})
	defer server.Close()

	result, err := client.ResolveUserIdentifier("email@test.com")
	if err != nil {
		t.Fatalf("ResolveUserIdentifier() error: %v", err)
	}
	if result.Identifier != "email@test.com" {
		t.Errorf("Identifier = %q, want 'email@test.com'", result.Identifier)
	}
}

func TestResolveUser(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"identifier":       "user-123",
				"muxi_user_id":     "muxi-123",
				"internal_user_id": 1,
			},
		})
	})
	defer server.Close()

	result, err := client.ResolveUser("user-123", false)
	if err != nil {
		t.Fatalf("ResolveUser() error: %v", err)
	}
	if result.Identifier != "user-123" {
		t.Errorf("Identifier = %q, want 'user-123'", result.Identifier)
	}
}

func TestCancelAsyncJob(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
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

	err := client.CancelAsyncJob("job-123")
	if err != nil {
		t.Fatalf("CancelAsyncJob() error: %v", err)
	}
}

func TestGetLLMSettings(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/llm/settings" {
			t.Errorf("expected path /llm/settings, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"model": "gpt-4",
			},
		})
	})
	defer server.Close()

	result, err := client.GetLLMSettings()
	if err != nil {
		t.Fatalf("GetLLMSettings() error: %v", err)
	}
	_ = result
}

func TestGetMemoryConfig(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/memory" {
			t.Errorf("expected path /memory, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"enabled": true,
			},
		})
	})
	defer server.Close()

	result, err := client.GetMemoryConfig()
	if err != nil {
		t.Fatalf("GetMemoryConfig() error: %v", err)
	}
	_ = result
}

func TestGetSchedulerConfig(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/scheduler" {
			t.Errorf("expected path /scheduler, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"enabled": true,
			},
		})
	})
	defer server.Close()

	result, err := client.GetSchedulerConfig()
	if err != nil {
		t.Fatalf("GetSchedulerConfig() error: %v", err)
	}
	_ = result
}

func TestGetSchedulerJobs(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"jobs":  []interface{}{},
				"count": 0,
			},
		})
	})
	defer server.Close()

	result, err := client.GetSchedulerJobs("")
	if err != nil {
		t.Fatalf("GetSchedulerJobs() error: %v", err)
	}
	if result.Count != 0 {
		t.Errorf("Count = %d, want 0", result.Count)
	}
}

func TestGetSchedulerJobs_RuntimeFieldMapping(t *testing.T) {
	scheduledFor := "2026-04-16T15:00:00Z"
	lastRunAt := "2026-04-16T14:00:00Z"

	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"jobs": []interface{}{
					map[string]interface{}{
						"id":              "job-recurring",
						"status":          "ACTIVE",
						"is_recurring":    true,
						"cron_expression": "0 * * * *",
						"scheduled_for":   nil,
						"original_prompt": "Hourly check-in",
						"last_run_at":     lastRunAt,
						"total_failures":  1,
					},
					map[string]interface{}{
						"id":              "job-once",
						"status":          "PAUSED",
						"is_recurring":    false,
						"cron_expression": nil,
						"scheduled_for":   scheduledFor,
						"original_prompt": "Follow up tomorrow",
					},
				},
				"count": 2,
			},
		})
	})
	defer server.Close()

	result, err := client.GetSchedulerJobs("")
	if err != nil {
		t.Fatalf("GetSchedulerJobs() error: %v", err)
	}

	if result.Count != 2 {
		t.Fatalf("Count = %d, want 2", result.Count)
	}

	recurring := result.Jobs[0]
	if recurring.Type != "recurring" {
		t.Errorf("Type = %q, want recurring", recurring.Type)
	}
	if recurring.Schedule != "0 * * * *" {
		t.Errorf("Schedule = %q, want cron expression", recurring.Schedule)
	}
	if recurring.Status != "ACTIVE" {
		t.Errorf("Status = %q, want ACTIVE", recurring.Status)
	}
	if !recurring.Enabled {
		t.Error("Enabled = false, want true for ACTIVE job")
	}
	if recurring.Message != "Hourly check-in" {
		t.Errorf("Message = %q, want original_prompt", recurring.Message)
	}
	if recurring.FailureCount != 1 {
		t.Errorf("FailureCount = %d, want 1", recurring.FailureCount)
	}
	if recurring.NextRun.IsZero() {
		t.Error("NextRun should be computed for recurring jobs")
	}

	expectedLastRun, err := time.Parse(time.RFC3339, lastRunAt)
	if err != nil {
		t.Fatalf("time.Parse() error: %v", err)
	}
	if !recurring.LastRun.Equal(expectedLastRun) {
		t.Errorf("LastRun = %v, want %v", recurring.LastRun, expectedLastRun)
	}

	oneTime := result.Jobs[1]
	if oneTime.Type != "one_time" {
		t.Errorf("Type = %q, want one_time", oneTime.Type)
	}
	if oneTime.Schedule != scheduledFor {
		t.Errorf("Schedule = %q, want scheduled_for", oneTime.Schedule)
	}
	if oneTime.Status != "PAUSED" {
		t.Errorf("Status = %q, want PAUSED", oneTime.Status)
	}
	if oneTime.Enabled {
		t.Error("Enabled = true, want false for PAUSED job")
	}

	expectedRunAt, err := time.Parse(time.RFC3339, scheduledFor)
	if err != nil {
		t.Fatalf("time.Parse() error: %v", err)
	}
	if !oneTime.RunAt.Equal(expectedRunAt) {
		t.Errorf("RunAt = %v, want %v", oneTime.RunAt, expectedRunAt)
	}
	if !oneTime.NextRun.Equal(expectedRunAt) {
		t.Errorf("NextRun = %v, want %v", oneTime.NextRun, expectedRunAt)
	}
	if oneTime.Message != "Follow up tomorrow" {
		t.Errorf("Message = %q, want original_prompt", oneTime.Message)
	}
}

func TestClearUserBuffer(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"cleared": true,
			},
		})
	})
	defer server.Close()

	result, err := client.ClearUserBuffer("user-123")
	if err != nil {
		t.Fatalf("ClearUserBuffer() error: %v", err)
	}
	_ = result
}

func TestClearAllBuffers(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"cleared": true,
			},
		})
	})
	defer server.Close()

	result, err := client.ClearAllBuffers()
	if err != nil {
		t.Fatalf("ClearAllBuffers() error: %v", err)
	}
	_ = result
}

func TestClearSessionBuffer(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"cleared": true,
			},
		})
	})
	defer server.Close()

	result, err := client.ClearSessionBuffer("session-123", "user-123")
	if err != nil {
		t.Fatalf("ClearSessionBuffer() error: %v", err)
	}
	_ = result
}

func TestListCredentialServices(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"services": []interface{}{},
				"count":    0,
			},
		})
	})
	defer server.Close()

	result, err := client.ListCredentialServices()
	if err != nil {
		t.Fatalf("ListCredentialServices() error: %v", err)
	}
	if result.Count != 0 {
		t.Errorf("Count = %d, want 0", result.Count)
	}
}

func TestListCredentials(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"credentials": []interface{}{},
				"count":       0,
			},
		})
	})
	defer server.Close()

	result, err := client.ListCredentials("user-123")
	if err != nil {
		t.Fatalf("ListCredentials() error: %v", err)
	}
	if result.Count != 0 {
		t.Errorf("Count = %d, want 0", result.Count)
	}
}

func TestTriggerTrigger(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"status":  "processing",
				"content": "",
			},
			"request": map[string]interface{}{
				"id": "req-123",
			},
		})
	})
	defer server.Close()

	result, err := client.TriggerTrigger("on-message", nil, false, "user-123")
	if err != nil {
		t.Fatalf("TriggerTrigger() error: %v", err)
	}
	if result.Status != "processing" {
		t.Errorf("Status = %q, want 'processing'", result.Status)
	}
}

func TestCreateSchedulerJob(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":       "job-123",
				"type":     "recurring",
				"schedule": "0 9 * * *",
				"status":   "active",
			},
		})
	})
	defer server.Close()

	result, err := client.CreateSchedulerJob("recurring", "0 9 * * *", "Daily check", "user-123")
	if err != nil {
		t.Fatalf("CreateSchedulerJob() error: %v", err)
	}
	if result.ID != "job-123" {
		t.Errorf("ID = %q, want 'job-123'", result.ID)
	}
}

func TestAddMemory(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":     "mem-123",
				"type":   "fact",
				"detail": "User likes coffee",
			},
		})
	})
	defer server.Close()

	result, err := client.AddMemory("user-123", "fact", "User likes coffee")
	if err != nil {
		t.Fatalf("AddMemory() error: %v", err)
	}
	if result.ID != "mem-123" {
		t.Errorf("ID = %q, want 'mem-123'", result.ID)
	}
}

func TestBulkLinkIdentifiers(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"muxi_user_id":           "muxi-123",
				"internal_user_id":       1,
				"identifiers_associated": 2,
				"new_identifiers":        []string{"email@test.com"},
			},
		})
	})
	defer server.Close()

	result, err := client.BulkLinkIdentifiers("user-123", []string{"email@test.com", "phone:123"})
	if err != nil {
		t.Fatalf("BulkLinkIdentifiers() error: %v", err)
	}
	if result.IdentifiersAssociated != 2 {
		t.Errorf("IdentifiersAssociated = %d, want 2", result.IdentifiersAssociated)
	}
}

func TestCreateCredential(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"credential_id":      "cred-123",
				"service":            "github",
				"name":               "my-github",
				"credential_preview": "ghp_***",
				"created_at":         "2024-01-01T00:00:00Z",
			},
		})
	})
	defer server.Close()

	req := &CreateCredentialRequest{
		Service:    "github",
		Name:       "my-github",
		Credential: map[string]interface{}{"token": "token123"},
	}
	result, err := client.CreateCredential("user-123", req)
	if err != nil {
		t.Fatalf("CreateCredential() error: %v", err)
	}
	if result.CredentialID != "cred-123" {
		t.Errorf("CredentialID = %q, want 'cred-123'", result.CredentialID)
	}
}

func TestRestoreSession(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    map[string]interface{}{},
		})
	})
	defer server.Close()

	messages := []Message{
		{Text: "Hello"},
		{Text: "Hi there!"},
	}
	err := client.RestoreSession("session-123", "user-123", messages)
	if err != nil {
		t.Fatalf("RestoreSession() error: %v", err)
	}
}

func TestGetAsyncJobs(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"jobs":  []interface{}{},
				"count": 0,
			},
		})
	})
	defer server.Close()

	result, err := client.GetAsyncJobs()
	if err != nil {
		t.Fatalf("GetAsyncJobs() error: %v", err)
	}
	if result.Count != 0 {
		t.Errorf("Count = %d, want 0", result.Count)
	}
}

func TestGetAsyncJob(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":     "job-123",
				"status": "completed",
			},
		})
	})
	defer server.Close()

	result, err := client.GetAsyncJob("job-123")
	if err != nil {
		t.Fatalf("GetAsyncJob() error: %v", err)
	}
	if result.ID != "job-123" {
		t.Errorf("ID = %q, want 'job-123'", result.ID)
	}
}

func TestDeleteMemory(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
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

	err := client.DeleteMemory("user-123", "mem-123")
	if err != nil {
		t.Fatalf("DeleteMemory() error: %v", err)
	}
}

func TestGetMCPServer(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":   "mcp-123",
				"name": "Test MCP",
			},
		})
	})
	defer server.Close()

	result, err := client.GetMCPServer("mcp-123")
	if err != nil {
		t.Fatalf("GetMCPServer() error: %v", err)
	}
	if result["id"] != "mcp-123" {
		t.Errorf("id = %v, want 'mcp-123'", result["id"])
	}
}

func TestGetAsyncConfig(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"enabled": true,
			},
		})
	})
	defer server.Close()

	result, err := client.GetAsyncConfig()
	if err != nil {
		t.Fatalf("GetAsyncConfig() error: %v", err)
	}
	_ = result
}

func TestDeleteSchedulerJob(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
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

	err := client.DeleteSchedulerJob("job-123")
	if err != nil {
		t.Fatalf("DeleteSchedulerJob() error: %v", err)
	}
}

func TestDeleteCredential(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
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

	_, err := client.DeleteCredential("user-123", "cred-123")
	if err != nil {
		t.Fatalf("DeleteCredential() error: %v", err)
	}
}

func TestGetSchedulerJob(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":       "job-123",
				"type":     "recurring",
				"schedule": "0 9 * * *",
				"status":   "active",
			},
		})
	})
	defer server.Close()

	result, err := client.GetSchedulerJob("job-123")
	if err != nil {
		t.Fatalf("GetSchedulerJob() error: %v", err)
	}
	if result.ID != "job-123" {
		t.Errorf("ID = %q, want 'job-123'", result.ID)
	}
}

func TestGetSchedulerJob_RuntimeFieldMapping(t *testing.T) {
	scheduledFor := "2026-04-17T09:30:00Z"
	lastRunAt := "2026-04-16T09:30:00Z"

	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":              "job-456",
				"status":          "COMPLETED",
				"is_recurring":    false,
				"cron_expression": nil,
				"scheduled_for":   scheduledFor,
				"original_prompt": "Send wrap-up",
				"last_run_at":     lastRunAt,
				"total_failures":  2,
			},
		})
	})
	defer server.Close()

	result, err := client.GetSchedulerJob("job-456")
	if err != nil {
		t.Fatalf("GetSchedulerJob() error: %v", err)
	}

	if result.ID != "job-456" {
		t.Errorf("ID = %q, want job-456", result.ID)
	}
	if result.Type != "one_time" {
		t.Errorf("Type = %q, want one_time", result.Type)
	}
	if result.Schedule != scheduledFor {
		t.Errorf("Schedule = %q, want scheduled_for", result.Schedule)
	}
	if result.Status != "COMPLETED" {
		t.Errorf("Status = %q, want COMPLETED", result.Status)
	}
	if result.Enabled {
		t.Error("Enabled = true, want false for COMPLETED job")
	}
	if result.Message != "Send wrap-up" {
		t.Errorf("Message = %q, want original_prompt", result.Message)
	}
	if result.FailureCount != 2 {
		t.Errorf("FailureCount = %d, want 2", result.FailureCount)
	}

	expectedRunAt, err := time.Parse(time.RFC3339, scheduledFor)
	if err != nil {
		t.Fatalf("time.Parse() error: %v", err)
	}
	if !result.RunAt.Equal(expectedRunAt) {
		t.Errorf("RunAt = %v, want %v", result.RunAt, expectedRunAt)
	}
	if !result.NextRun.Equal(expectedRunAt) {
		t.Errorf("NextRun = %v, want %v", result.NextRun, expectedRunAt)
	}

	expectedLastRun, err := time.Parse(time.RFC3339, lastRunAt)
	if err != nil {
		t.Fatalf("time.Parse() error: %v", err)
	}
	if !result.LastRun.Equal(expectedLastRun) {
		t.Errorf("LastRun = %v, want %v", result.LastRun, expectedLastRun)
	}
}

func TestGetCredential(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"credential_id":      "cred-123",
				"service":            "github",
				"name":               "my-github",
				"credential_preview": "ghp_***",
			},
		})
	})
	defer server.Close()

	result, err := client.GetCredential("user-123", "cred-123")
	if err != nil {
		t.Fatalf("GetCredential() error: %v", err)
	}
	if result.CredentialID != "cred-123" {
		t.Errorf("CredentialID = %q, want 'cred-123'", result.CredentialID)
	}
}

func TestResolveUserNew(t *testing.T) {
	server, client := mockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"identifier":       "email:test@example.com",
				"muxi_user_id":     "user-123",
				"internal_user_id": 1,
			},
		})
	})
	defer server.Close()

	result, err := client.ResolveUser("email:test@example.com", false)
	if err != nil {
		t.Fatalf("ResolveUser() error: %v", err)
	}
	if result.MuxiUserID != "user-123" {
		t.Errorf("MuxiUserID = %q, want 'user-123'", result.MuxiUserID)
	}
}
