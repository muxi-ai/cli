package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"hi", 5, "hi"},
		{"", 5, ""},
		{"exactly5", 5, "exac…"},
	}

	for _, tt := range tests {
		got := truncateStr(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncateStr(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		content string
		want    string
	}{
		{"version: 1.0.0\nother: stuff", "1.0.0"},
		{"version: 2.3.4", "2.3.4"},
		{"no version here", ""},
		{"version: v1.2.3", "v1.2.3"},
	}

	for _, tt := range tests {
		got := extractVersion(tt.content)
		if got != tt.want {
			t.Errorf("extractVersion(%q) = %q, want %q", tt.content, got, tt.want)
		}
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		v    string
		want bool
	}{
		{"1.0.0", true},
		{"0.0.1", true},
		{"10.20.30", true},
		{"1.0", false},
		{"v1.0.0", false},
		{"1.0.0.0", false},
		{"abc", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isValidVersion(tt.v)
		if got != tt.want {
			t.Errorf("isValidVersion(%q) = %v, want %v", tt.v, got, tt.want)
		}
	}
}

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		current  string
		bumpType string
		want     string
	}{
		{"1.0.0", "patch", "1.0.1"},
		{"1.0.0", "minor", "1.1.0"},
		{"1.0.0", "major", "2.0.0"},
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "major", "2.0.0"},
		{"0.0.9", "patch", "0.0.10"},
	}

	for _, tt := range tests {
		got := bumpVersion(tt.current, tt.bumpType)
		if got != tt.want {
			t.Errorf("bumpVersion(%q, %q) = %q, want %q", tt.current, tt.bumpType, got, tt.want)
		}
	}
}

func TestFormatDurationShort(t *testing.T) {
	tests := []struct {
		seconds int64
		want    string
	}{
		{30, "0m"},
		{60, "1m"},
		{90, "1m"},
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{86400, "1d 0h"},
		{90000, "1d 1h"},
	}

	for _, tt := range tests {
		got := formatDurationShort(tt.seconds)
		if got != tt.want {
			t.Errorf("formatDurationShort(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestGetAVContentType(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".mp3", "audio/mpeg"},
		{".m4a", "audio/mp4"},
		{".wav", "audio/wav"},
		{".mp4", "video/mp4"},
		{".mov", "video/quicktime"},
		{".webm", "audio/webm"},
		{".unknown", ""},
	}

	for _, tt := range tests {
		got := getAVContentType(tt.ext)
		if got != tt.want {
			t.Errorf("getAVContentType(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

func TestIsAVFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"audio.mp3", true},
		{"audio.m4a", true},
		{"audio.wav", true},
		{"video.mp4", true},
		{"video.mov", true},
		{"document.pdf", false},
		{"image.png", false},
		{"file.txt", false},
	}

	for _, tt := range tests {
		got := isAVFile(tt.path)
		if got != tt.want {
			t.Errorf("isAVFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestContainsLine(t *testing.T) {
	// Create temp file
	tmpDir, _ := os.MkdirTemp("", "test-*")
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line one\nline two\nline three\n"
	os.WriteFile(testFile, []byte(content), 0644)

	if !containsLine(testFile, "line two") {
		t.Error("containsLine should find 'line two'")
	}
	if containsLine(testFile, "line four") {
		t.Error("containsLine should not find 'line four'")
	}
	if containsLine("/nonexistent/file", "anything") {
		t.Error("containsLine should return false for missing file")
	}
}

func TestCleanConfigOutput(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{},
		"key3": "value3",
	}

	cleanConfigOutput(data)

	// Should still have key1 and key3
	if data["key1"] != "value1" {
		t.Error("key1 should remain")
	}
}

func TestRemoveEmptyObjects(t *testing.T) {
	data := map[string]interface{}{
		"filled": map[string]interface{}{"inner": "value"},
		"empty":  map[string]interface{}{},
		"string": "value",
	}

	removeEmptyObjects(data)

	if _, ok := data["empty"]; ok {
		t.Error("empty object should be removed")
	}
	if data["string"] != "value" {
		t.Error("string value should remain")
	}
}
