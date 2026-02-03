package updates

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	VersionURL    = "https://raw.githubusercontent.com/muxi-ai/cli/main/src/.version"
	CacheDuration = 24 * time.Hour
	FetchTimeout  = 2 * time.Second
)

// VersionCache stores the cached version info
type VersionCache struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

// UpdateInfo contains update notification details
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
}

var currentVersion string

// SetCurrentVersion sets the current CLI version (called from main)
func SetCurrentVersion(version string) {
	currentVersion = version
}

// CheckCachedUpdate checks if an update is available based on cached data
// Returns nil if no update available or cache doesn't exist
func CheckCachedUpdate() *UpdateInfo {
	if currentVersion == "" || currentVersion == "dev" {
		return nil
	}

	cache, err := loadCache()
	if err != nil || cache == nil {
		return nil
	}

	if isNewerVersion(cache.LatestVersion, currentVersion) {
		return &UpdateInfo{
			CurrentVersion: currentVersion,
			LatestVersion:  cache.LatestVersion,
		}
	}

	return nil
}

// RefreshCache fetches the latest version and updates the cache
// This should be called asynchronously (fire-and-forget)
// On error, keeps existing cache intact
func RefreshCache() {
	if currentVersion == "" || currentVersion == "dev" {
		return
	}

	// Check if cache is still fresh
	cache, _ := loadCache()
	if cache != nil && time.Since(cache.CheckedAt) < CacheDuration {
		return // Cache is fresh, no need to refresh
	}

	// Fetch latest version
	latest, err := fetchLatestVersion()
	if err != nil {
		// On error, ensure cache file exists with current version
		if cache == nil {
			saveCache(&VersionCache{
				LatestVersion: currentVersion,
				CheckedAt:     time.Now(),
			})
		}
		return // Keep existing cache on error
	}

	// Update cache
	saveCache(&VersionCache{
		LatestVersion: latest,
		CheckedAt:     time.Now(),
	})
}

// fetchLatestVersion fetches the latest version from GitHub
func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: FetchTimeout}

	resp, err := client.Get(VersionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

// isNewerVersion checks if latest is newer than current
func isNewerVersion(latest, current string) bool {
	// Simple string comparison works for our version format (0.YYYYMMDD.N)
	// Both should be in same format
	return latest > current && latest != current
}

// cachePath returns the path to the version cache file
func cachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muxi", "cli", "version-cache.json")
}

// loadCache loads the cached version info
func loadCache() (*VersionCache, error) {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return nil, err
	}

	var cache VersionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// saveCache saves the version info to cache
func saveCache(cache *VersionCache) error {
	path := cachePath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
