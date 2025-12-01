package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Client is the registry API client
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new registry client
func NewClient(registry string) (*Client, error) {
	if registry == "" {
		registry = GetDefaultRegistry()
	}

	token, err := GetToken(registry)
	if err != nil {
		return nil, err
	}

	baseURL := "https://" + registry

	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// IsAuthenticated returns true if client has a token
func (c *Client) IsAuthenticated() bool {
	return c.Token != ""
}

// GetAuthURL returns the URL for browser authentication
func (c *Client) GetAuthURL(callbackPort int) string {
	if callbackPort > 0 {
		return fmt.Sprintf("%s/auth/cli/authorize?callback=http://localhost:%d/auth", c.BaseURL, callbackPort)
	}
	return fmt.Sprintf("%s/auth/cli/authorize", c.BaseURL)
}

// ValidateToken validates a token with the registry
func (c *Client) ValidateToken(token string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/auth/validate", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid token")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("validation failed: %s", resp.Status)
	}

	var user UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &user, nil
}

// GetFormation retrieves formation metadata
func (c *Client) GetFormation(ref string, trackDownload bool) (*Formation, error) {
	parsed, err := ParseFormationRef(ref)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/api/formations/@%s/%s", parsed.Owner, parsed.Name)
	if parsed.Version != "" {
		endpoint += "?version=" + parsed.Version
	}
	if trackDownload {
		if parsed.Version != "" {
			endpoint += "&pull=true"
		} else {
			endpoint += "?pull=true"
		}
	}

	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("formation %s not found", ref)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get formation: %s - %s", resp.Status, string(body))
	}

	var formation Formation
	if err := json.NewDecoder(resp.Body).Decode(&formation); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &formation, nil
}

// GetVersions retrieves all versions of a formation
func (c *Client) GetVersions(ref string) ([]Version, error) {
	parsed, err := ParseFormationRef(ref)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/api/formations/@%s/%s/versions", parsed.Owner, parsed.Name)

	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get versions: %s", resp.Status)
	}

	var versions []Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return versions, nil
}

// Search searches for formations
func (c *Client) Search(query, sort string, limit int) (*SearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	if sort != "" {
		params.Set("sort", sort)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	endpoint := "/api/search?" + params.Encode()

	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetPullInfo retrieves download information for a formation
func (c *Client) GetPullInfo(ref string) (*PullInfo, error) {
	parsed, err := ParseFormationRef(ref)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/api/formations/@%s/%s/pull", parsed.Owner, parsed.Name)
	if parsed.Version != "" {
		endpoint += "?version=" + parsed.Version
	}

	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("formation %s not found", ref)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get pull info: %s", resp.Status)
	}

	var info PullInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &info, nil
}

// DownloadFormation downloads a formation ZIP to a file
func (c *Client) DownloadFormation(downloadURL, destPath string) error {
	resp, err := c.HTTPClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Ensure directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// Publish uploads a formation bundle to the registry
func (c *Client) Publish(zipPath string, org string) (*PublishResult, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bundle: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("bundle", filepath.Base(zipPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to write bundle: %w", err)
	}

	if org != "" {
		if err := writer.WriteField("org", org); err != nil {
			return nil, fmt.Errorf("failed to add org field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close form: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/formations/publish", &body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("version already exists - bump version in formation.yaml")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("publish failed: %s - %s", resp.Status, string(respBody))
	}

	var result PublishResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ParseFormationRef parses a formation reference like @user/name:version
func ParseFormationRef(ref string) (*FormationRef, error) {
	// Remove leading @ if present
	ref = strings.TrimPrefix(ref, "@")

	// Pattern: user/name or user/name:version
	pattern := regexp.MustCompile(`^([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)(?::([a-zA-Z0-9._-]+))?$`)
	matches := pattern.FindStringSubmatch(ref)

	if matches == nil {
		return nil, fmt.Errorf("invalid formation reference: %s (expected @user/name or @user/name:version)", ref)
	}

	return &FormationRef{
		Owner:   matches[1],
		Name:    matches[2],
		Version: matches[3],
		Full:    "@" + ref,
	}, nil
}

// FormatSize formats bytes to human readable string
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
