package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFormation(t *testing.T, dir, body string) {
	t.Helper()
	content := "schema: \"1.0.0\"\nid: \"my-formation\"\n" + body
	if err := os.WriteFile(filepath.Join(dir, "formation.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write formation: %v", err)
	}
}

func hasErrorContaining(result *Result, substr string) bool {
	for _, e := range result.Errors {
		if contains(e.Message, substr) {
			return true
		}
	}
	return false
}

func TestValidateKnowledgeSources(t *testing.T) {
	writeAgent := func(t *testing.T, knowledgeBlock string) *Result {
		tmpDir := t.TempDir()
		writeFormation(t, tmpDir, "")
		agentsDir := filepath.Join(tmpDir, "agents")
		os.Mkdir(agentsDir, 0755)
		content := "id: \"my-agent\"\nrole: \"helper\"\nknowledge:\n" + knowledgeBlock
		os.WriteFile(filepath.Join(agentsDir, "my-agent.yaml"), []byte(content), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return result
	}

	t.Run("local path source is valid", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - path: \"docs/\"\n")
		if hasErrorContaining(result, "knowledge") {
			t.Errorf("unexpected knowledge error: %+v", result.Errors)
		}
	})

	t.Run("supported remote scheme is valid", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - url: \"s3://bucket/docs\"\n")
		if hasErrorContaining(result, "scheme") {
			t.Errorf("unexpected scheme error: %+v", result.Errors)
		}
	})

	t.Run("unsupported remote scheme is flagged", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - url: \"gopher://host/docs\"\n")
		if !hasErrorContaining(result, "unsupported URL scheme") {
			t.Errorf("expected scheme error, got: %+v", result.Errors)
		}
	})

	t.Run("path and url together are flagged", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - path: \"docs/\"\n      url: \"https://example.com/docs.zip\"\n")
		if !hasErrorContaining(result, "not both") {
			t.Errorf("expected path/url error, got: %+v", result.Errors)
		}
	})

	t.Run("source without path or url is flagged", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - description: \"orphan\"\n")
		if !hasErrorContaining(result, "'path' or 'url'") {
			t.Errorf("expected missing-field error, got: %+v", result.Errors)
		}
	})

	t.Run("agent_tree regenerate enum is enforced", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - path: \"docs/\"\n      agent_tree:\n        regenerate: \"hourly\"\n")
		if !hasErrorContaining(result, "must be one of: manual, on-source-change, on-formation-load") {
			t.Errorf("expected regenerate enum error, got: %+v", result.Errors)
		}
	})

	t.Run("valid agent_tree regenerate passes", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - path: \"docs/\"\n      agent_tree:\n        regenerate: \"on-source-change\"\n")
		if hasErrorContaining(result, "agent_tree") {
			t.Errorf("unexpected agent_tree error: %+v", result.Errors)
		}
	})

	t.Run("agent_tree on remote source is flagged", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - url: \"s3://bucket/docs\"\n      agent_tree:\n        regenerate: \"manual\"\n")
		if !hasErrorContaining(result, "not supported on remote") {
			t.Errorf("expected remote agent_tree error, got: %+v", result.Errors)
		}
	})

	t.Run("unknown agent_tree key is flagged", func(t *testing.T) {
		result := writeAgent(t, "  sources:\n    - path: \"docs/\"\n      agent_tree:\n        depth: 3\n")
		if !hasErrorContaining(result, "not recognized") {
			t.Errorf("expected unknown-key error, got: %+v", result.Errors)
		}
	})
}

func TestValidateToolFilters(t *testing.T) {
	writeMCP := func(t *testing.T, toolsBlock string) *Result {
		tmpDir := t.TempDir()
		writeFormation(t, tmpDir, "")
		mcpsDir := filepath.Join(tmpDir, "mcps")
		os.Mkdir(mcpsDir, 0755)
		content := "id: \"my-mcp\"\ntype: \"http\"\ntools:\n" + toolsBlock
		os.WriteFile(filepath.Join(mcpsDir, "my-mcp.yaml"), []byte(content), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return result
	}

	t.Run("canonical allow only is valid", func(t *testing.T) {
		result := writeMCP(t, "  allow:\n    - tool_a\n")
		if hasErrorContaining(result, "tools") {
			t.Errorf("unexpected tools error: %+v", result.Errors)
		}
	})

	t.Run("allow and deny are mutually exclusive", func(t *testing.T) {
		result := writeMCP(t, "  allow:\n    - tool_a\n  deny:\n    - tool_b\n")
		if !hasErrorContaining(result, "mutually exclusive") {
			t.Errorf("expected mutual-exclusivity error, got: %+v", result.Errors)
		}
	})

	t.Run("legacy whitelist and blacklist still flagged", func(t *testing.T) {
		result := writeMCP(t, "  whitelist:\n    - tool_a\n  blacklist:\n    - tool_b\n")
		if !hasErrorContaining(result, "mutually exclusive") {
			t.Errorf("expected mutual-exclusivity error, got: %+v", result.Errors)
		}
	})

	t.Run("canonical plus alias of same list is an error", func(t *testing.T) {
		result := writeMCP(t, "  allow:\n    - tool_a\n  whitelist:\n    - tool_a\n")
		if !hasErrorContaining(result, "declare only one spelling") {
			t.Errorf("expected same-list spelling error, got: %+v", result.Errors)
		}
	})

	t.Run("unknown tools key is an error", func(t *testing.T) {
		result := writeMCP(t, "  permit:\n    - tool_a\n")
		if !hasErrorContaining(result, "unknown key") {
			t.Errorf("expected unknown-key error, got: %+v", result.Errors)
		}
	})
}

func TestValidateTuning(t *testing.T) {
	run := func(t *testing.T, body string) *Result {
		tmpDir := t.TempDir()
		writeFormation(t, tmpDir, body)
		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return result
	}

	t.Run("boolean shorthand is valid", func(t *testing.T) {
		result := run(t, "tuning: false\n")
		if hasErrorContaining(result, "tuning") {
			t.Errorf("unexpected tuning error: %+v", result.Errors)
		}
	})

	t.Run("known keys are valid", func(t *testing.T) {
		result := run(t, "tuning:\n  active: true\n  interval_hours: 24\n  auto_apply: false\n")
		if hasErrorContaining(result, "tuning") {
			t.Errorf("unexpected tuning error: %+v", result.Errors)
		}
	})

	t.Run("unknown key is an error", func(t *testing.T) {
		result := run(t, "tuning:\n  cadence: daily\n")
		if !hasErrorContaining(result, "unknown key") {
			t.Errorf("expected unknown-key error, got: %+v", result.Errors)
		}
	})

	t.Run("non-boolean auto_apply is an error", func(t *testing.T) {
		result := run(t, "tuning:\n  auto_apply: \"yes\"\n")
		if !hasErrorContaining(result, "must be a boolean") {
			t.Errorf("expected boolean type error, got: %+v", result.Errors)
		}
	})

	t.Run("non-positive interval_hours is an error", func(t *testing.T) {
		result := run(t, "tuning:\n  interval_hours: 0\n")
		if !hasErrorContaining(result, "positive number") {
			t.Errorf("expected positive-number error, got: %+v", result.Errors)
		}
	})

	t.Run("scalar string is an error", func(t *testing.T) {
		result := run(t, "tuning: aggressive\n")
		if !hasErrorContaining(result, "boolean or a mapping") {
			t.Errorf("expected type error, got: %+v", result.Errors)
		}
	})
}

func TestValidateA2AInboundAuth(t *testing.T) {
	run := func(t *testing.T, authBlock string) *Result {
		tmpDir := t.TempDir()
		writeFormation(t, tmpDir, "a2a:\n  inbound:\n    auth:\n"+authBlock)
		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return result
	}

	t.Run("hmac with secret is valid", func(t *testing.T) {
		result := run(t, "      type: hmac\n      secret: \"${{ secrets.A2A_SECRET }}\"\n")
		if hasErrorContaining(result, "auth") {
			t.Errorf("unexpected auth error: %+v", result.Errors)
		}
	})

	t.Run("hmac without secret is an error", func(t *testing.T) {
		result := run(t, "      type: hmac\n")
		if !hasErrorContaining(result, "hmac auth requires 'secret'") {
			t.Errorf("expected missing-secret error, got: %+v", result.Errors)
		}
	})

	t.Run("openid requires issuer", func(t *testing.T) {
		result := run(t, "      type: openid\n")
		if !hasErrorContaining(result, "openid auth requires 'issuer'") {
			t.Errorf("expected missing-issuer error, got: %+v", result.Errors)
		}
	})

	t.Run("openid issuer must be a URL", func(t *testing.T) {
		result := run(t, "      type: openid\n      issuer: \"not-a-url\"\n")
		if !hasErrorContaining(result, "must be a URL") {
			t.Errorf("expected issuer URL error, got: %+v", result.Errors)
		}
	})

	t.Run("oauth2 inbound is rejected with hint", func(t *testing.T) {
		result := run(t, "      type: oauth2\n")
		if !hasErrorContaining(result, "outbound-only") {
			t.Errorf("expected direction hint error, got: %+v", result.Errors)
		}
	})
}

func TestValidateA2AServiceFiles(t *testing.T) {
	run := func(t *testing.T, authBlock string) *Result {
		tmpDir := t.TempDir()
		writeFormation(t, tmpDir, "")
		a2aDir := filepath.Join(tmpDir, "a2a")
		os.Mkdir(a2aDir, 0755)
		content := "id: \"remote-svc\"\nurl: \"https://example.com\"\nauth:\n" + authBlock
		os.WriteFile(filepath.Join(a2aDir, "remote-svc.yaml"), []byte(content), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return result
	}

	t.Run("oauth2 with all fields is valid", func(t *testing.T) {
		result := run(t, "  type: oauth2\n  client_id: \"cid\"\n  client_secret: \"${{ secrets.CS }}\"\n  token_url: \"https://idp/token\"\n")
		if hasErrorContaining(result, "auth") {
			t.Errorf("unexpected auth error: %+v", result.Errors)
		}
	})

	t.Run("oauth2 missing fields are errors", func(t *testing.T) {
		result := run(t, "  type: oauth2\n  client_id: \"cid\"\n")
		if !hasErrorContaining(result, "requires 'client_secret'") || !hasErrorContaining(result, "requires 'token_url'") {
			t.Errorf("expected missing oauth2 field errors, got: %+v", result.Errors)
		}
	})

	t.Run("openid outbound is rejected with hint", func(t *testing.T) {
		result := run(t, "  type: openid\n  issuer: \"https://idp\"\n")
		if !hasErrorContaining(result, "inbound-only") {
			t.Errorf("expected direction hint error, got: %+v", result.Errors)
		}
	})

	t.Run("unknown auth type is an error", func(t *testing.T) {
		result := run(t, "  type: kerberos\n")
		if !hasErrorContaining(result, "auth type 'kerberos' is invalid") {
			t.Errorf("expected invalid-type error, got: %+v", result.Errors)
		}
	})
}
