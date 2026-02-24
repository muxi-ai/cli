package formation

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SaveArtifact saves an artifact to ~/.muxi/cli/outputs/{formationID}/
// Returns the saved file path
func SaveArtifact(artifact Artifact, formationID string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	outputDir := filepath.Join(home, ".muxi", "cli", "outputs", formationID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := resolveFilename(outputDir, artifact.Filename)
	destPath := filepath.Join(outputDir, filename)

	// Text artifacts: save content directly
	if artifact.Content != nil && *artifact.Content != "" {
		if err := os.WriteFile(destPath, []byte(*artifact.Content), 0644); err != nil {
			return "", fmt.Errorf("failed to write file: %w", err)
		}
		return destPath, nil
	}

	// Binary artifacts: decode data URL
	if artifact.DataURL != nil && *artifact.DataURL != "" {
		data, err := decodeDataURL(*artifact.DataURL)
		if err != nil {
			return "", fmt.Errorf("failed to decode artifact: %w", err)
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return "", fmt.Errorf("failed to write file: %w", err)
		}
		return destPath, nil
	}

	return "", fmt.Errorf("artifact has no content or data_url")
}

// resolveFilename handles collisions by appending -1, -2, etc.
func resolveFilename(dir, filename string) string {
	if _, err := os.Stat(filepath.Join(dir, filename)); os.IsNotExist(err) {
		return filename
	}

	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-%d%s", base, i, ext)
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}

	return filename
}

// decodeDataURL parses "data:<mime>;base64,<data>" and returns raw bytes
func decodeDataURL(dataURL string) ([]byte, error) {
	// Format: data:<mime>;base64,<encoded>
	idx := strings.Index(dataURL, ",")
	if idx < 0 {
		return nil, fmt.Errorf("invalid data URL format")
	}
	encoded := dataURL[idx+1:]
	return base64.StdEncoding.DecodeString(encoded)
}

// FormatArtifactSize formats bytes to human-readable string
func FormatArtifactSize(bytes int) string {
	if bytes == 0 {
		return ""
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := unit, 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
