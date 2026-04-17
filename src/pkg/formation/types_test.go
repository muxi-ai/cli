package formation

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserIDRequiredError(t *testing.T) {
	err := &UserIDRequiredError{}
	msg := err.Error()

	if msg == "" {
		t.Error("Error() should return non-empty message")
	}

	// Should mention user ID
	if !containsSubstring(msg, "user") {
		t.Errorf("Error() should mention user, got: %s", msg)
	}
}

func TestChatFile(t *testing.T) {
	file := ChatFile{
		Filename:    "test.mp3",
		Content:     "base64content",
		ContentType: "audio/mpeg",
		Size:        1024,
	}

	if file.Filename != "test.mp3" {
		t.Errorf("Filename = %q, want 'test.mp3'", file.Filename)
	}
	if file.Size != 1024 {
		t.Errorf("Size = %d, want 1024", file.Size)
	}
}

func TestChatRequest(t *testing.T) {
	req := ChatRequest{
		Message:   "Hello",
		SessionID: "sess_123",
		Stream:    true,
	}

	if req.Message != "Hello" {
		t.Errorf("Message = %q, want 'Hello'", req.Message)
	}
	if !req.Stream {
		t.Error("Stream should be true")
	}
}

func TestAudioChatRequest(t *testing.T) {
	req := AudioChatRequest{
		Files: []ChatFile{
			{Filename: "voice.m4a", ContentType: "audio/mp4"},
		},
		UserID:    "alice",
		SessionID: "sess_123",
		Stream:    true,
	}

	if len(req.Files) != 1 {
		t.Errorf("Files length = %d, want 1", len(req.Files))
	}
	if req.UserID != "alice" {
		t.Errorf("UserID = %q, want 'alice'", req.UserID)
	}
}

func TestFlexTimeUnmarshalJSON_ParsesFractionalOffsetTimestamp(t *testing.T) {
	var ft FlexTime
	raw := []byte(`"2026-04-16T09:00:00.123456+00:00"`)

	if err := json.Unmarshal(raw, &ft); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	expected, err := time.Parse(time.RFC3339Nano, "2026-04-16T09:00:00.123456+00:00")
	if err != nil {
		t.Fatalf("time.Parse() error: %v", err)
	}

	if !ft.Equal(expected) {
		t.Errorf("FlexTime = %v, want %v", ft.Time, expected)
	}
}

func TestFlexTimeUnmarshalJSON_ParsesFractionalTimestampWithoutColonOffset(t *testing.T) {
	var ft FlexTime
	raw := []byte(`"2026-04-16T09:00:00.123456+0000"`)

	if err := json.Unmarshal(raw, &ft); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	expected, err := time.Parse("2006-01-02T15:04:05.999999999Z0700", "2026-04-16T09:00:00.123456+0000")
	if err != nil {
		t.Fatalf("time.Parse() error: %v", err)
	}

	if !ft.Equal(expected) {
		t.Errorf("FlexTime = %v, want %v", ft.Time, expected)
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
