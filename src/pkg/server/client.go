package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Version is set at build time
var Version = "dev"

// sdkTransport wraps http.RoundTripper to add X-Muxi-SDK header
type sdkTransport struct {
	base http.RoundTripper
}

func (t *sdkTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Muxi-SDK", "cli/"+Version)
	return t.base.RoundTrip(req)
}

// Client is an HTTP client for MUXI Server
type Client struct {
	BaseURL    string
	KeyID      string
	SecretKey  string
	HTTPClient *http.Client
}

// NewClient creates a new server client from a profile name
func NewClient(profile string) (*Client, error) {
	entry, err := GetProfile(profile)
	if err != nil {
		return nil, err
	}

	return NewClientFromEntry(entry), nil
}

// NewClientFromEntry creates a client from a profile entry
func NewClientFromEntry(entry *ProfileEntry) *Client {
	return &Client{
		BaseURL:   entry.URL,
		KeyID:     entry.KeyID,
		SecretKey: entry.SecretKey,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &sdkTransport{base: http.DefaultTransport},
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

// RestartFormation restarts a formation (non-streaming)
func (c *Client) RestartFormation(id string) error {
	_, err := c.restartFormationInternal(id, false, nil)
	return err
}

// RestartFormationStreaming restarts a formation with SSE progress
func (c *Client) RestartFormationStreaming(id string, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	return c.restartFormationInternal(id, true, callback)
}

// StartFormation starts a stopped formation (non-streaming)
func (c *Client) StartFormation(id string) error {
	_, err := c.startFormationInternal(id, false, nil)
	return err
}

// StartFormationStreaming starts a formation with SSE progress
func (c *Client) StartFormationStreaming(id string, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	return c.startFormationInternal(id, true, callback)
}

// startFormationInternal handles both streaming and non-streaming start
func (c *Client) startFormationInternal(id string, streaming bool, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	path := "/rpc/formations/" + id + "/start"

	req, err := http.NewRequest("POST", c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	if streaming {
		req.Header.Set("Accept", "text/event-stream")
	}

	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, "POST", path)
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 10 * time.Minute, Transport: &sdkTransport{base: http.DefaultTransport}}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to start: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, checkResponse(resp)
	}

	if !streaming {
		return nil, checkResponse(resp)
	}

	return parseSSEStream(resp.Body, callback)
}

// restartFormationInternal handles both streaming and non-streaming restart
func (c *Client) restartFormationInternal(id string, streaming bool, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	path := "/rpc/formations/" + id + "/restart"

	// Create request
	req, err := http.NewRequest("POST", c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	if streaming {
		req.Header.Set("Accept", "text/event-stream")
	}

	// Add auth
	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, "POST", path)
	req.Header.Set("Authorization", authHeader)

	// Use longer timeout for streaming
	client := &http.Client{Timeout: 10 * time.Minute, Transport: &sdkTransport{base: http.DefaultTransport}}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to restart: %w", err)
	}
	defer resp.Body.Close()

	// Check for error responses
	if resp.StatusCode >= 400 {
		return nil, checkResponse(resp)
	}

	// Non-streaming: just check response
	if !streaming {
		return nil, checkResponse(resp)
	}

	// Streaming: parse SSE events
	return parseSSEStream(resp.Body, callback)
}

// RollbackFormation rolls back a formation (non-streaming)
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

// RollbackFormationStreaming rolls back with SSE progress
func (c *Client) RollbackFormationStreaming(id string, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	path := "/rpc/formations/" + id + "/rollback"

	req, err := http.NewRequest("POST", c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/event-stream")
	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, "POST", path)
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 10 * time.Minute, Transport: &sdkTransport{base: http.DefaultTransport}}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to rollback: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, checkResponse(resp)
	}

	return parseSSEStream(resp.Body, callback)
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

// CancelUpdate cancels a running formation update
func (c *Client) CancelUpdate(id string) error {
	resp, err := c.Post("/rpc/formations/"+id+"/cancel-update", nil, "")
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

// StreamFormationLogs streams logs via SSE (follow mode)
func (c *Client) StreamFormationLogs(id, stream string, callback func(LogEvent) error) error {
	path := fmt.Sprintf("/rpc/formations/%s/logs?stream=%s&follow=true", id, stream)

	// Create request with SSE accept header
	req, err := http.NewRequest("GET", c.BaseURL+path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")

	// Add auth (use path without query params for signature)
	authPath := fmt.Sprintf("/rpc/formations/%s/logs", id)
	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, "GET", authPath)
	req.Header.Set("Authorization", authHeader)

	// No timeout for streaming
	client := &http.Client{Timeout: 0, Transport: &sdkTransport{base: http.DefaultTransport}}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return checkResponse(resp)
	}

	// Parse SSE stream
	reader := bufio.NewReader(resp.Body)
	var eventType string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			if eventType == "log" {
				var logEvent LogEvent
				if err := json.Unmarshal([]byte(data), &logEvent); err != nil {
					continue // Skip malformed events
				}
				if err := callback(logEvent); err != nil {
					return err
				}
			}
			// Reset for next event
			eventType = ""
		}
	}
}

// DeployFormation deploys a new formation (non-streaming)
func (c *Client) DeployFormation(id, bundlePath, version string) error {
	_, err := c.deployFormationInternal(id, bundlePath, version, false, nil)
	return err
}

// DeployFormationStreaming deploys a new formation with SSE progress
// The callback is called for each SSE event received
func (c *Client) DeployFormationStreaming(id, bundlePath, version string, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	return c.deployFormationInternal(id, bundlePath, version, true, callback)
}

// deployFormationInternal handles both streaming and non-streaming deploy
func (c *Client) deployFormationInternal(id, bundlePath, version string, streaming bool, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	// Open bundle file
	file, err := os.Open(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bundle: %w", err)
	}
	defer file.Close()

	// Create request
	req, err := http.NewRequest("POST", c.BaseURL+"/rpc/formations", file)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("X-Formation-ID", id)
	if version != "" {
		req.Header.Set("X-Formation-Version", version)
	}
	if streaming {
		req.Header.Set("Accept", "text/event-stream")
	}

	// Add auth
	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, "POST", "/rpc/formations")
	req.Header.Set("Authorization", authHeader)

	// Use longer timeout for streaming (deployment can take minutes)
	client := &http.Client{Timeout: 10 * time.Minute, Transport: &sdkTransport{base: http.DefaultTransport}}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy: %w", err)
	}
	defer resp.Body.Close()

	// Check for error responses
	if resp.StatusCode >= 400 {
		return nil, checkResponse(resp)
	}

	// Non-streaming: just check response
	if !streaming {
		return nil, checkResponse(resp)
	}

	// Streaming: parse SSE events
	return parseSSEStream(resp.Body, callback)
}

// UpdateFormation updates an existing formation (non-streaming)
func (c *Client) UpdateFormation(id, bundlePath, version string) error {
	_, err := c.updateFormationInternal(id, bundlePath, version, false, nil)
	return err
}

// UpdateFormationStreaming updates a formation with SSE progress
func (c *Client) UpdateFormationStreaming(id, bundlePath, version string, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	return c.updateFormationInternal(id, bundlePath, version, true, callback)
}

// updateFormationInternal handles both streaming and non-streaming update
func (c *Client) updateFormationInternal(id, bundlePath, version string, streaming bool, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	// Open bundle file
	file, err := os.Open(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bundle: %w", err)
	}
	defer file.Close()

	path := "/rpc/formations/" + id

	// Create request
	req, err := http.NewRequest("PUT", c.BaseURL+path, file)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/gzip")
	if version != "" {
		req.Header.Set("X-Formation-Version", version)
	}
	if streaming {
		req.Header.Set("Accept", "text/event-stream")
	}

	// Add auth
	authHeader := BuildAuthHeader(c.KeyID, c.SecretKey, "PUT", path)
	req.Header.Set("Authorization", authHeader)

	// Use longer timeout for streaming
	client := &http.Client{Timeout: 10 * time.Minute, Transport: &sdkTransport{base: http.DefaultTransport}}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update: %w", err)
	}
	defer resp.Body.Close()

	// Check for error responses
	if resp.StatusCode >= 400 {
		return nil, checkResponse(resp)
	}

	// Non-streaming: just check response
	if !streaming {
		return nil, checkResponse(resp)
	}

	// Streaming: parse SSE events
	return parseSSEStream(resp.Body, callback)
}

// parseSSEStream parses Server-Sent Events from a response body
func parseSSEStream(body io.Reader, callback func(SSEEvent) error) (*DeployCompleteEvent, error) {
	reader := bufio.NewReader(body)
	var currentEvent SSEEvent

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading SSE stream: %w", err)
		}

		line = strings.TrimSpace(line)

		// Empty line = end of event
		if line == "" {
			if currentEvent.Event != "" && currentEvent.Data != "" {
				// Handle the event
				if callback != nil {
					if err := callback(currentEvent); err != nil {
						return nil, err
					}
				}

				// Check for terminal events
				switch currentEvent.Event {
				case "complete":
					var complete DeployCompleteEvent
					if err := json.Unmarshal([]byte(currentEvent.Data), &complete); err != nil {
						return nil, fmt.Errorf("failed to parse complete event: %w", err)
					}
					return &complete, nil
				case "error":
					var errEvent DeployErrorEvent
					if err := json.Unmarshal([]byte(currentEvent.Data), &errEvent); err != nil {
						return nil, fmt.Errorf("server error: %s", currentEvent.Data)
					}
					// Clean up error message - extract the useful part
					msg := cleanDeployErrorMessage(errEvent.Message)
					return nil, fmt.Errorf("%s", msg)
				}

				currentEvent = SSEEvent{}
			}
			continue
		}

		// Parse event type
		if strings.HasPrefix(line, "event:") {
			currentEvent.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		// Parse data
		if strings.HasPrefix(line, "data:") {
			currentEvent.Data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			continue
		}
	}

	return nil, fmt.Errorf("SSE stream ended without complete or error event")
}

// cleanDeployErrorMessage extracts the useful part of a deploy error message
func cleanDeployErrorMessage(msg string) string {
	// Remove common prefixes like "Formation failed health check after 5m0s: "
	prefixes := []string{
		"Formation failed health check after ",
		"Formation crashed during startup: ",
	}

	result := msg
	for _, prefix := range prefixes {
		if idx := strings.Index(result, prefix); idx != -1 {
			// Find the end of this prefix (after the colon and space)
			afterPrefix := result[idx+len(prefix):]
			if colonIdx := strings.Index(afterPrefix, ": "); colonIdx != -1 && colonIdx < 20 {
				// There's a duration like "5m0s: " - skip past it
				result = afterPrefix[colonIdx+2:]
			} else {
				result = afterPrefix
			}
		}
	}

	return strings.TrimSpace(result)
}

// DownloadFormation downloads a formation as a zip file
// Returns the path to the temporary zip file (caller must delete it)
// If includeDB is true, adds ?db=true to include SQLite database files
func (c *Client) DownloadFormation(id string, includeDB bool) (string, error) {
	path := "/rpc/formations/" + id + "/download"
	if includeDB {
		path += "?db=true"
	}

	resp, err := c.Get(path)
	if err != nil {
		return "", fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return "", err
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "muxi-download-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Copy response body to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to download: %w", err)
	}

	return tmpFile.Name(), nil
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
