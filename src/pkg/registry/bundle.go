package registry

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// BundleInfo contains information about a created bundle
type BundleInfo struct {
	Path      string
	FileCount int
	Size      int64
	Warnings  []string
}

// ExcludedPatterns are patterns to exclude from bundles
var ExcludedPatterns = []string{
	".git",
	".muxi",
	"secrets.enc",
	".key",
	".env",
	".env.*",
	"node_modules",
	"__pycache__",
	"*.pyc",
	".DS_Store",
	"*.log",
	"*.tmp",
	".vscode",
	".idea",
}

// IncludedPatterns are patterns to include in bundles
var IncludedPatterns = []string{
	"formation.yaml",
	"README.md",
	"README",
	"LICENSE",
	"agents/*.yaml",
	"mcps/*.yaml",
	"a2a/*.yaml",
	"sops/*.md",
	"triggers/*.yaml",
	"knowledge/*.md",
}

// CreateBundle creates a ZIP bundle of a formation
func CreateBundle(formationDir string) (*BundleInfo, error) {
	// Create temp file for the bundle
	tmpFile, err := os.CreateTemp("", "muxi-bundle-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	zipWriter := zip.NewWriter(tmpFile)

	info := &BundleInfo{
		Path:     tmpPath,
		Warnings: []string{},
	}

	// Check for secrets.enc and warn
	if _, err := os.Stat(filepath.Join(formationDir, "secrets.enc")); err == nil {
		info.Warnings = append(info.Warnings, "secrets.enc found - will NOT be included in bundle")
	}

	// Walk the directory and add files
	err = filepath.Walk(formationDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(formationDir, path)
		if err != nil {
			return err
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Check if should be excluded
		if shouldExclude(relPath, fileInfo.IsDir()) {
			if fileInfo.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories (they're created implicitly)
		if fileInfo.IsDir() {
			return nil
		}

		// Check if should be included
		if !shouldInclude(relPath) {
			return nil
		}

		// Add file to zip
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", relPath, err)
		}
		defer file.Close()

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return fmt.Errorf("failed to add %s to bundle: %w", relPath, err)
		}

		if _, err := io.Copy(writer, file); err != nil {
			return fmt.Errorf("failed to write %s: %w", relPath, err)
		}

		info.FileCount++
		return nil
	})

	if err != nil {
		zipWriter.Close()
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to finalize bundle: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to close bundle: %w", err)
	}

	// Get final size
	stat, err := os.Stat(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to stat bundle: %w", err)
	}
	info.Size = stat.Size()

	return info, nil
}

// ExtractBundle extracts a ZIP bundle to a directory
func ExtractBundle(zipPath, destDir string) (int, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open bundle: %w", err)
	}
	defer reader.Close()

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

	fileCount := 0

	for _, file := range reader.File {
		destPath := filepath.Join(destDir, file.Name)

		// Security check - prevent path traversal
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fileCount, fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return fileCount, fmt.Errorf("failed to create directory %s: %w", file.Name, err)
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fileCount, fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Extract file
		srcFile, err := file.Open()
		if err != nil {
			return fileCount, fmt.Errorf("failed to open %s in bundle: %w", file.Name, err)
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			srcFile.Close()
			return fileCount, fmt.Errorf("failed to create %s: %w", file.Name, err)
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return fileCount, fmt.Errorf("failed to extract %s: %w", file.Name, err)
		}

		fileCount++
	}

	return fileCount, nil
}

// shouldExclude checks if a path should be excluded
func shouldExclude(path string, isDir bool) bool {
	name := filepath.Base(path)

	for _, pattern := range ExcludedPatterns {
		// Direct name match
		if name == pattern {
			return true
		}

		// Glob match
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}

		// Path prefix match for directories
		if isDir && strings.HasPrefix(path, pattern) {
			return true
		}
	}

	return false
}

// shouldInclude checks if a path should be included
func shouldInclude(path string) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)
	name := filepath.Base(path)

	// Root files
	rootIncludes := []string{"formation.yaml", "README.md", "README", "LICENSE", "secrets"}
	for _, include := range rootIncludes {
		if path == include || name == include {
			return true
		}
	}

	// Directory patterns - include all yaml/md files in these dirs
	dirPatterns := map[string][]string{
		"agents/":    {"*.yaml", "*.yml"},
		"mcps/":      {"*.yaml", "*.yml"},
		"a2a/":       {"*.yaml", "*.yml"},
		"sops/":      {"*.md", "*.yaml", "*.yml"},
		"triggers/":  {"*.yaml", "*.yml"},
		"knowledge/": {"*.md", "*.txt", "*.yaml", "*.yml"},
	}

	for dirPrefix, patterns := range dirPatterns {
		if strings.HasPrefix(path, dirPrefix) {
			for _, pattern := range patterns {
				if matched, _ := filepath.Match(pattern, name); matched {
					return true
				}
			}
		}
	}

	return false
}

// CleanupBundle removes a temporary bundle file
func CleanupBundle(path string) {
	os.Remove(path)
}
