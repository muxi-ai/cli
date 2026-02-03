package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/validate"
	"github.com/spf13/cobra"
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

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds int64
		wantLen int // Just check it's not empty
	}{
		{0, 1},
		{30, 1},
		{60, 1},
		{3600, 1},
		{86400, 1},
	}

	for _, tt := range tests {
		got := formatDuration(tt.seconds)
		if len(got) < tt.wantLen {
			t.Errorf("formatDuration(%d) = %q, too short", tt.seconds, got)
		}
	}
}

func TestFormatStatus(t *testing.T) {
	statuses := []string{"running", "stopped", "error", "starting", "unknown"}

	for _, status := range statuses {
		got := formatStatus(status)
		if got == "" {
			t.Errorf("formatStatus(%q) returned empty", status)
		}
	}
}

func TestCountFiles(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-*")
	defer os.RemoveAll(tmpDir)

	// Create some test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.go"), []byte("test"), 0644)

	count := countFiles(tmpDir, "*.txt")
	if count != 2 {
		t.Errorf("countFiles(*.txt) = %d, want 2", count)
	}

	count = countFiles(tmpDir, "*.go")
	if count != 1 {
		t.Errorf("countFiles(*.go) = %d, want 1", count)
	}

	count = countFiles("/nonexistent", "*.txt")
	if count != 0 {
		t.Errorf("countFiles(nonexistent) = %d, want 0", count)
	}
}

func TestBuildOrderedConfigYAML(t *testing.T) {
	data := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
	}

	result := buildOrderedConfigYAML(data)
	if result == "" {
		t.Error("buildOrderedConfigYAML() returned empty string")
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hel.."},
		{"hi", 5, "hi"},
	}

	for _, tt := range tests {
		got := truncateString(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestIsVersionHigher(t *testing.T) {
	tests := []struct {
		newV string
		oldV string
		want bool
	}{
		{"2.0.0", "1.0.0", true},
		{"1.1.0", "1.0.0", true},
		{"1.0.1", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "2.0.0", false},
		{"1.0.0", "1.1.0", false},
	}

	for _, tt := range tests {
		got := isVersionHigher(tt.newV, tt.oldV)
		if got != tt.want {
			t.Errorf("isVersionHigher(%q, %q) = %v, want %v", tt.newV, tt.oldV, got, tt.want)
		}
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
	}{
		{"1.0.0", [3]int{1, 0, 0}},
		{"2.3.4", [3]int{2, 3, 4}},
		{"10.20.30", [3]int{10, 20, 30}},
	}

	for _, tt := range tests {
		got := parseVersion(tt.input)
		if got != tt.want {
			t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFormatTimeout(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{30, "0:30"},
		{60, "1:00"},
		{90, "1:30"},
		{3600, "60:00"},
	}

	for _, tt := range tests {
		got := formatTimeout(tt.seconds)
		if got != tt.want {
			t.Errorf("formatTimeout(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		seconds int64
		wantLen int
	}{
		{0, 1},
		{60, 1},
		{3600, 1},
		{86400, 1},
	}

	for _, tt := range tests {
		got := formatUptime(tt.seconds)
		if len(got) < tt.wantLen {
			t.Errorf("formatUptime(%d) returned too short: %q", tt.seconds, got)
		}
	}
}

func TestFormatLatency(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{time.Millisecond, "1ms"},
		{time.Second, "1.00s"},
		{500 * time.Microsecond, "500µs"},
	}

	for _, tt := range tests {
		got := formatLatency(tt.d)
		if got != tt.want {
			t.Errorf("formatLatency(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatTimestamp(t *testing.T) {
	// Test with valid timestamp
	got := formatTimestamp("2024-01-01T00:00:00Z")
	if got == "" {
		t.Error("formatTimestamp should not return empty for valid timestamp")
	}

	// Test with invalid timestamp
	got = formatTimestamp("invalid")
	if got != "invalid" {
		t.Errorf("formatTimestamp(invalid) = %q, want 'invalid'", got)
	}
}

func TestNormalizeService(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  spaces  ", "--spaces--"},
		{"UPPER", "upper"},
		{"Mixed Case", "mixed-case"},
		{"", ""},
	}

	for _, tt := range tests {
		got := normalizeService(tt.input)
		if got != tt.want {
			t.Errorf("normalizeService(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShouldExclude(t *testing.T) {
	tests := []struct {
		path      string
		isDir     bool
		includeDB bool
		want      bool
	}{
		{".git", true, false, true},
		{"node_modules", true, false, true},
		{"__pycache__", true, false, true},
		{".DS_Store", false, false, true},
		{"main.go", false, false, false},
		{"src", true, false, false},
		{"formation.yaml", false, false, false},
		// memory.db tests
		{"memory.db", false, false, true},       // excluded by default
		{"memory.db", false, true, false},       // included with flag
		{"other.db", false, false, false},       // other db files not excluded
	}

	for _, tt := range tests {
		got := shouldExclude(tt.path, tt.isDir, tt.includeDB)
		if got != tt.want {
			t.Errorf("shouldExclude(%q, %v, %v) = %v, want %v", tt.path, tt.isDir, tt.includeDB, got, tt.want)
		}
	}
}

func TestReplaceResourceWithCLI(t *testing.T) {
	data := map[string]interface{}{
		"resource": "some-value",
		"other":    "kept",
	}

	replaceResourceWithCLI(data)

	if _, ok := data["resource"]; ok {
		t.Error("resource key should be removed")
	}
	if data["other"] != "kept" {
		t.Error("other keys should remain")
	}
}

func TestWriteYAMLField(t *testing.T) {
	buf := new(bytes.Buffer)
	writeYAMLField(buf, "test", "value", 0)

	result := buf.String()
	if result == "" {
		t.Error("writeYAMLField should produce output")
	}
}

func TestWriteYAMLArrayItem(t *testing.T) {
	buf := new(bytes.Buffer)
	data := map[string]interface{}{"key": "value"}
	writeYAMLArrayItem(buf, data, 0)

	result := buf.String()
	if result == "" {
		t.Error("writeYAMLArrayItem should produce output")
	}
}

func TestParseFlexibleDate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2025-12-31T14:00:00Z", false},
		{"2025-12-31T14:00:00", false},
		{"2025-12-31 14:00:00", false},
		{"2025-12-31 14:00", false},
		{"2025-12-31", false},
		{"12/31/2025 14:00", false},
		{"12/31/2025", false},
		{"Dec 31, 2025 2:00pm", false},
		{"Dec 31, 2025", false},
		{"invalid-date", true},
		{"", true},
	}

	for _, tt := range tests {
		_, err := parseFlexibleDate(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseFlexibleDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestTruncateScheduler(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		input time.Time
	}{
		{now.Add(-time.Second * 30)},
		{now.Add(-time.Minute * 5)},
		{now.Add(-time.Hour * 2)},
		{now.Add(-time.Hour * 48)},
	}

	for _, tt := range tests {
		result := formatRelativeTime(tt.input)
		if result == "" {
			t.Errorf("formatRelativeTime() returned empty string")
		}
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()
	tests := []struct {
		input time.Time
		want  string
	}{
		{time.Time{}, "unknown"},
		{now.Add(-time.Second * 30), "just now"},
		{now.Add(-time.Minute), "1 minute ago"},
		{now.Add(-time.Minute * 5), "5 minutes ago"},
		{now.Add(-time.Hour), "1 hour ago"},
		{now.Add(-time.Hour * 5), "5 hours ago"},
	}

	for _, tt := range tests {
		got := formatTimeAgo(tt.input)
		if got != tt.want {
			t.Errorf("formatTimeAgo() = %q, want %q", got, tt.want)
		}
	}
}

func TestFormatDurationProfiles(t *testing.T) {
	tests := []struct {
		input int64
	}{
		{30},
		{90},
		{3600},
		{3661},
	}

	for _, tt := range tests {
		got := formatDuration(tt.input)
		if got == "" {
			t.Errorf("formatDuration(%d) returned empty string", tt.input)
		}
	}
}

func TestFormatInfoUptime(t *testing.T) {
	tests := []struct {
		input int64
	}{
		{30},
		{90},
		{3661},
		{86400},
	}

	for _, tt := range tests {
		got := formatUptime(tt.input)
		if got == "" {
			t.Errorf("formatUptime(%d) returned empty string", tt.input)
		}
	}
}

func TestFormatTriggerStatus(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"active"},
		{"inactive"},
		{"error"},
		{"unknown"},
	}

	for _, tt := range tests {
		got := formatStatus(tt.input)
		if got == "" {
			t.Errorf("formatStatus(%q) returned empty string", tt.input)
		}
	}
}

func TestPrintBox(t *testing.T) {
	printBox("Test Title", "Test Subtitle")
	printBoxWithSubtitle("Test Title", "Test Subtitle")
	printBoxSimple("Test Title", "Test Subtitle")
	printBoxLine("Test Content")
	printBoxLineDimmed("Test Content")
	printBoxDivider()
	printBoxBottom()
	printDivider()
}

func TestRemoveEmptyObjectsMore(t *testing.T) {
	data := map[string]interface{}{
		"id":     "test",
		"nested": map[string]interface{}{},
		"deep": map[string]interface{}{
			"empty": map[string]interface{}{},
		},
	}
	removeEmptyObjects(data)
}

func TestWriteYAMLValueForArrayItem(t *testing.T) {
	buf := new(bytes.Buffer)
	writeYAMLValueForArrayItem(buf, "string", 0)
	if buf.Len() == 0 {
		t.Error("writeYAMLValueForArrayItem should produce output")
	}

	buf.Reset()
	writeYAMLValueForArrayItem(buf, 123, 0)
	if buf.Len() == 0 {
		t.Error("writeYAMLValueForArrayItem should produce output for int")
	}

	buf.Reset()
	writeYAMLValueForArrayItem(buf, true, 0)
	if buf.Len() == 0 {
		t.Error("writeYAMLValueForArrayItem should produce output for bool")
	}

	buf.Reset()
	writeYAMLValueForArrayItem(buf, map[string]interface{}{"key": "value"}, 0)
	if buf.Len() == 0 {
		t.Error("writeYAMLValueForArrayItem should produce output for map")
	}
}

func TestPrintColoredCommand(t *testing.T) {
	printColoredCommand("muxi deploy")
	printColoredCommand("muxi formation list")
}

func TestWriteYAMLFieldComplex(t *testing.T) {
	buf := new(bytes.Buffer)

	// Test array value
	writeYAMLField(buf, "items", []interface{}{"a", "b", "c"}, 0)
	if buf.Len() == 0 {
		t.Error("writeYAMLField should produce output for array")
	}

	buf.Reset()
	// Test nested map
	writeYAMLField(buf, "nested", map[string]interface{}{
		"inner": "value",
	}, 0)
	if buf.Len() == 0 {
		t.Error("writeYAMLField should produce output for nested map")
	}

	buf.Reset()
	// Test nil value
	writeYAMLField(buf, "nil", nil, 0)
}

func TestFormatDurationShortMore(t *testing.T) {
	tests := []struct {
		input int64
	}{
		{0},
		{59},
		{60},
		{3600},
		{86400},
	}

	for _, tt := range tests {
		got := formatDurationShort(tt.input)
		if got == "" {
			t.Errorf("formatDurationShort(%d) returned empty string", tt.input)
		}
	}
}

func TestFormatTimestampMore(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"2025-12-31T14:00:00Z"},
		{"invalid"},
		{""},
	}

	for _, tt := range tests {
		formatTimestamp(tt.input)
	}
}

func TestFormatServerStageMessage(t *testing.T) {
	tests := []server.DeployProgressEvent{
		{Stage: "validating"},
		{Stage: "building"},
		{Stage: "deploying"},
		{Stage: "unknown"},
	}

	for _, tt := range tests {
		got := formatServerStageMessage(tt)
		if got == "" {
			t.Errorf("formatServerStageMessage(%q) returned empty string", tt.Stage)
		}
	}
}

func TestFormatStageComplete(t *testing.T) {
	event := &server.DeployProgressEvent{
		Stage:   "validating",
		Message: "Done",
	}

	stages := []string{"validating", "building", "deploying", "unknown"}
	for _, stage := range stages {
		got := formatStageComplete(stage, event)
		if got == "" {
			t.Errorf("formatStageComplete(%q) returned empty string", stage)
		}
	}
}

func TestCreateTarGzBundle(t *testing.T) {
	// Create a temp directory with a simple formation
	tmpDir := t.TempDir()

	// Create formation.yaml
	formationYAML := `schema_version: "1"
id: test-formation
version: "0.0.1"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "formation.yaml"), []byte(formationYAML), 0644); err != nil {
		t.Fatalf("Failed to create formation.yaml: %v", err)
	}

	// Create bundle
	bundlePath, fileCount, err := createTarGzBundle(tmpDir, "test-formation", false)
	if err != nil {
		t.Fatalf("createTarGzBundle() error: %v", err)
	}
	defer os.Remove(bundlePath)

	if fileCount < 1 {
		t.Errorf("fileCount = %d, want >= 1", fileCount)
	}
}

func TestPlayNotificationSound(t *testing.T) {
	playNotificationSound(true)
	playNotificationSound(false)
}

func TestFormatStartStageMessage(t *testing.T) {
	tests := []server.DeployProgressEvent{
		{Stage: "starting"},
		{Stage: "waiting"},
		{Stage: "connecting"},
		{Stage: "unknown"},
	}

	for _, tt := range tests {
		got := formatStartStageMessage(tt)
		if got == "" {
			t.Errorf("formatStartStageMessage(%q) returned empty string", tt.Stage)
		}
	}
}

func TestFormatStartStageComplete(t *testing.T) {
	event := &server.DeployProgressEvent{
		Stage:   "starting",
		Message: "Done",
	}

	stages := []string{"starting", "waiting", "connecting", "unknown"}
	for _, stage := range stages {
		got := formatStartStageComplete(stage, event)
		if got == "" {
			t.Errorf("formatStartStageComplete(%q) returned empty string", stage)
		}
	}
}

func TestPlayStartNotificationSound(t *testing.T) {
	playStartNotificationSound(true)
	playStartNotificationSound(false)
}

func TestFormatRestartStageMessage(t *testing.T) {
	tests := []server.DeployProgressEvent{
		{Stage: "stopping"},
		{Stage: "waiting"},
		{Stage: "starting"},
		{Stage: "unknown"},
	}

	for _, tt := range tests {
		got := formatRestartStageMessage(tt)
		if got == "" {
			t.Errorf("formatRestartStageMessage(%q) returned empty string", tt.Stage)
		}
	}
}

func TestFormatRestartStageComplete(t *testing.T) {
	event := &server.DeployProgressEvent{
		Stage:   "stopping",
		Message: "Done",
	}

	stages := []string{"stopping", "waiting", "starting", "unknown"}
	for _, stage := range stages {
		got := formatRestartStageComplete(stage, event)
		if got == "" {
			t.Errorf("formatRestartStageComplete(%q) returned empty string", stage)
		}
	}
}

func TestPlayRestartNotificationSound(t *testing.T) {
	playRestartNotificationSound(true)
	playRestartNotificationSound(false)
}

func TestFormatRollbackStageMessage(t *testing.T) {
	tests := []server.DeployProgressEvent{
		{Stage: "preparing"},
		{Stage: "restoring"},
		{Stage: "restarting"},
		{Stage: "unknown"},
	}

	for _, tt := range tests {
		got := formatRollbackStageMessage(tt)
		if got == "" {
			t.Errorf("formatRollbackStageMessage(%q) returned empty string", tt.Stage)
		}
	}
}

func TestFormatRollbackStageComplete(t *testing.T) {
	event := &server.DeployProgressEvent{
		Stage:   "preparing",
		Message: "Done",
	}

	stages := []string{"preparing", "restoring", "restarting", "unknown"}
	for _, stage := range stages {
		got := formatRollbackStageComplete(stage, event)
		if got == "" {
			t.Errorf("formatRollbackStageComplete(%q) returned empty string", stage)
		}
	}
}

func TestPlayRollbackNotificationSound(t *testing.T) {
	playRollbackNotificationSound(true)
	playRollbackNotificationSound(false)
}

func TestIsInFormationDir(t *testing.T) {
	// Test when not in a formation dir
	result := isInFormationDir()
	// Just verify it doesn't panic
	_ = result
}

func TestLoadDotMuxi(t *testing.T) {
	// Test loading when file doesn't exist
	_, err := loadDotMuxi()
	// Just verify it returns without panic (may or may not find file)
	_ = err
}

func TestDetermineScope(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("global", false, "")
	cmd.Flags().Bool("local", false, "")

	scope, err := determineScope(cmd)
	// Just verify it runs without panic
	_ = scope
	_ = err
}

func TestDisplayValidationResults(t *testing.T) {
	result := &validate.Result{
		Errors:   []validate.Issue{},
		Warnings: []validate.Issue{},
	}
	// Just verify it doesn't panic
	displayValidationResults("test-formation", result)
}

func TestDisplayValidationResultsWithErrors(t *testing.T) {
	result := &validate.Result{
		Errors: []validate.Issue{
			{Field: "id", Message: "Error 1"},
			{Field: "version", Message: "Error 2"},
		},
		Warnings: []validate.Issue{
			{Field: "description", Message: "Warning 1"},
		},
	}
	displayValidationResults("test-formation", result)
}

func TestSaveDotMuxi(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	config := &DotMuxi{
		Profile: "test-profile",
	}

	err := saveDotMuxi(config)
	if err != nil {
		t.Fatalf("saveDotMuxi() error: %v", err)
	}

	// Verify file was created
	_, err = os.Stat(".muxi")
	if os.IsNotExist(err) {
		t.Error(".muxi file was not created")
	}
}

func TestPrintCommandGroup(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	subCmd := &cobra.Command{
		Use:     "sub",
		Short:   "A subcommand",
		GroupID: "test-group",
	}
	rootCmd.AddCommand(subCmd)
	rootCmd.AddGroup(&cobra.Group{ID: "test-group", Title: "Test Group"})

	// Just verify it doesn't panic
	printCommandGroup(rootCmd, "test-group")
}

func TestPrintUngroupedCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	subCmd := &cobra.Command{
		Use:   "sub",
		Short: "A subcommand",
	}
	rootCmd.AddCommand(subCmd)

	// Just verify it doesn't panic
	printUngroupedCommands(rootCmd)
}

func TestCustomHelp(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	// Just verify it doesn't panic
	customHelp(rootCmd, []string{})
}
