package server

import (
	"strings"
	"testing"
)

func TestGenerateHMACSignature(t *testing.T) {
	sig, ts := GenerateHMACSignature("secret-key", "GET", "/api/test")

	if sig == "" {
		t.Error("signature should not be empty")
	}
	if ts == 0 {
		t.Error("timestamp should not be zero")
	}

	// Same inputs should produce different signatures (due to timestamp)
	// but within same second they might match - just verify format
	if len(sig) < 20 {
		t.Errorf("signature seems too short: %s", sig)
	}
}

func TestBuildAuthHeader(t *testing.T) {
	header := BuildAuthHeader("key-id-123", "secret-key", "POST", "/api/formations")

	// Header format: MUXI-HMAC key=..., timestamp=..., signature=...
	if !strings.HasPrefix(header, "MUXI-HMAC ") {
		t.Errorf("header should start with 'MUXI-HMAC ', got: %s", header)
	}
	if !strings.Contains(header, "key=") {
		t.Errorf("header should contain 'key=', got: %s", header)
	}
	if !strings.Contains(header, "timestamp=") {
		t.Errorf("header should contain 'timestamp=', got: %s", header)
	}
	if !strings.Contains(header, "signature=") {
		t.Errorf("header should contain 'signature=', got: %s", header)
	}
}

func TestBuildAuthHeader_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		header := BuildAuthHeader("key", "secret", method, "/path")
		if header == "" {
			t.Errorf("header for %s should not be empty", method)
		}
	}
}
