package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client is an HTTP client for MUXI Server
type Client struct {
	BaseURL    string
	KeyID      string
	SecretKey  string
	HTTPClient *http.Client
}

// NewClient creates a new server client from a profile name
func NewClient(profile string) (*Client, error) {
	entry, err := GetServer(profile)
	if err != nil {
		return nil, err
	}

	return NewClientFromEntry(entry), nil
}

// NewClientFromEntry creates a client from a server entry
func NewClientFromEntry(entry *ServerEntry) *Client {
	return &Client{
		BaseURL:   entry.URL,
		KeyID:     entry.KeyID,
		SecretKey: entry.SecretKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Do performs an authenticated HTTP request
func (c *Client) Do(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	url := c.BaseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Add HMAC auth header
	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, method, path)
	req.Header.Set("Authorization", authHeader)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return c.HTTPClient.Do(req)
}

// Get performs an authenticated GET request
func (c *Client) Get(path string) (*http.Response, error) {
	return c.Do("GET", path, nil, "")
}

// Post performs an authenticated POST request
func (c *Client) Post(path string, body io.Reader, contentType string) (*http.Response, error) {
	return c.Do("POST", path, body, contentType)
}

// Put performs an authenticated PUT request
func (c *Client) Put(path string, body io.Reader, contentType string) (*http.Response, error) {
	return c.Do("PUT", path, body, contentType)
}

// Delete performs an authenticated DELETE request
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.Do("DELETE", path, nil, "")
}

// Ping tests connectivity to the server (unauthenticated)
// Returns the response body size in bytes
func (c *Client) Ping() (int64, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/ping")
	if err != nil {
		return 0, fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("server returned: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return int64(len(body)), nil
}

// Health checks server health (unauthenticated)
func (c *Client) Health() (*HealthResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %s", resp.Status)
	}

	var result HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetServerStatus gets server status (authenticated)
func (c *Client) GetServerStatus() (*ServerStatusResponse, error) {
	resp, err := c.Get("/rpc/server/status")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("%s", result.Message)
	}

	var status ServerStatusResponse
	if err := json.Unmarshal(result.Data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return &status, nil
}

// ListFormations lists all formations
func (c *Client) ListFormations() (*ListFormationsResponse, error) {
	resp, err := c.Get("/rpc/formations")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("%s", result.Message)
	}

	var formations ListFormationsResponse
	if err := json.Unmarshal(result.Data, &formations); err != nil {
		return nil, fmt.Errorf("failed to parse formations: %w", err)
	}

	return &formations, nil
}

// GetFormation gets a formation by ID
func (c *Client) GetFormation(id string) (*FormationDetail, error) {
	resp, err := c.Get("/rpc/formations/" + id)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("%s", result.Message)
	}

	var formation FormationDetail
	if err := json.Unmarshal(result.Data, &formation); err != nil {
		return nil, fmt.Errorf("failed to parse formation: %w", err)
	}

	return &formation, nil
}

// StopFormation stops a formation
func (c *Client) StopFormation(id string) error {
	resp, err := c.Post("/rpc/formations/"+id+"/stop", nil, "")
	if err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// RestartFormation restarts a formation
func (c *Client) RestartFormation(id string) error {
	resp, err := c.Post("/rpc/formations/"+id+"/restart", nil, "")
	if err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// RollbackFormation rolls back a formation to previous version
func (c *Client) RollbackFormation(id string) (*RollbackResponse, error) {
	resp, err := c.Post("/rpc/formations/"+id+"/rollback", nil, "")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var rollback RollbackResponse
	if err := json.Unmarshal(result.Data, &rollback); err != nil {
		return nil, fmt.Errorf("failed to parse rollback: %w", err)
	}

	return &rollback, nil
}

// DeleteFormation deletes a formation
func (c *Client) DeleteFormation(id string) error {
	resp, err := c.Delete("/rpc/formations/" + id)
	if err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// GetFormationLogs gets formation logs
func (c *Client) GetFormationLogs(id string, lines int, stream string) (*LogsResponse, error) {
	path := fmt.Sprintf("/rpc/formations/%s/logs?lines=%d&stream=%s", id, lines, stream)
	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var logs LogsResponse
	if err := json.Unmarshal(result.Data, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}

	return &logs, nil
}

// DeployFormation deploys a new formation
func (c *Client) DeployFormation(id, bundlePath, version string) error {
	// Open bundle file
	file, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to open bundle: %w", err)
	}
	defer file.Close()

	// Create request
	req, err := http.NewRequest("POST", c.BaseURL+"/rpc/formations", file)
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("X-Formation-ID", id)
	if version != "" {
		req.Header.Set("X-Formation-Version", version)
	}

	// Add auth
	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, "POST", "/rpc/formations")
	req.Header.Set("Authorization", authHeader)

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// UpdateFormation updates an existing formation
func (c *Client) UpdateFormation(id, bundlePath, version string) error {
	// Open bundle file
	file, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to open bundle: %w", err)
	}
	defer file.Close()

	path := "/rpc/formations/" + id

	// Create request
	req, err := http.NewRequest("PUT", c.BaseURL+path, file)
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/gzip")
	if version != "" {
		req.Header.Set("X-Formation-Version", version)
	}

	// Add auth
	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, "PUT", path)
	req.Header.Set("Authorization", authHeader)

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// checkResponse checks for common error responses
func checkResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed - check server credentials")
	case http.StatusForbidden:
		return fmt.Errorf("access denied")
	case http.StatusNotFound:
		return fmt.Errorf("not found")
	case http.StatusConflict:
		return fmt.Errorf("conflict - resource already exists")
	default:
		body, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return fmt.Errorf("%s", errResp.Message)
		}
		return fmt.Errorf("server error: %s", resp.Status)
	}
}
