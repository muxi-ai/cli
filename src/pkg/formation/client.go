package formation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the Formation API
type Client struct {
	BaseURL    string // e.g., http://localhost:7890/api/my-formation/v1
	AdminKey   string
	ClientKey  string
	HTTPClient *http.Client
}

// NewClient creates a new Formation API client
func NewClient(baseURL, adminKey, clientKey string) *Client {
	return &Client{
		BaseURL:   baseURL,
		AdminKey:  adminKey,
		ClientKey: clientKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Do performs an HTTP request with the appropriate auth header
func (c *Client) Do(method, path string, body io.Reader, useAdminKey bool) (*http.Response, error) {
	url := c.BaseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Set auth header based on endpoint type
	if useAdminKey {
		if c.AdminKey == "" {
			return nil, fmt.Errorf("admin key required but not set")
		}
		req.Header.Set("X-MUXI-ADMIN-KEY", c.AdminKey)
	} else {
		if c.ClientKey == "" {
			return nil, fmt.Errorf("client key required but not set")
		}
		req.Header.Set("X-MUXI-CLIENT-KEY", c.ClientKey)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.HTTPClient.Do(req)
}

// DoWithUserID performs a request with X-User-ID header (for client endpoints)
func (c *Client) DoWithUserID(method, path string, body io.Reader, userID string) (*http.Response, error) {
	url := c.BaseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if c.ClientKey == "" {
		return nil, fmt.Errorf("client key required but not set")
	}
	req.Header.Set("X-MUXI-CLIENT-KEY", c.ClientKey)
	req.Header.Set("X-User-ID", userID)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.HTTPClient.Do(req)
}

// Get performs an authenticated GET request (admin key)
func (c *Client) Get(path string) (*http.Response, error) {
	return c.Do("GET", path, nil, true)
}

// GetClient performs a GET request with client key
func (c *Client) GetClient(path string) (*http.Response, error) {
	return c.Do("GET", path, nil, false)
}

// GetWithUser performs a GET request with client key and user ID
func (c *Client) GetWithUser(path, userID string) (*http.Response, error) {
	return c.DoWithUserID("GET", path, nil, userID)
}

// Post performs an authenticated POST request (admin key)
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	return c.Do("POST", path, reader, true)
}

// PostClient performs a POST request with client key
func (c *Client) PostClient(path string, body interface{}) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	return c.Do("POST", path, reader, false)
}

// PostWithUser performs a POST request with client key and user ID
func (c *Client) PostWithUser(path string, body interface{}, userID string) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	return c.DoWithUserID("POST", path, reader, userID)
}

// Patch performs an authenticated PATCH request (admin key)
func (c *Client) Patch(path string, body interface{}) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	return c.Do("PATCH", path, reader, true)
}

// Put performs an authenticated PUT request (admin key)
func (c *Client) Put(path string, body interface{}) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	return c.Do("PUT", path, reader, true)
}

// Delete performs an authenticated DELETE request (admin key)
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.Do("DELETE", path, nil, true)
}

// DeleteWithUser performs a DELETE request with client key and user ID
func (c *Client) DeleteWithUser(path, userID string) (*http.Response, error) {
	return c.DoWithUserID("DELETE", path, nil, userID)
}

// Health checks formation health (no auth required)
func (c *Client) Health() (*HealthResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to formation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("formation returned: %s", resp.Status)
	}

	// Formation API wraps everything in APIResponse envelope
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("health check failed")
	}

	var result HealthResponse
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse health data: %w", err)
	}

	return &result, nil
}

// GetStatus gets formation status
func (c *Client) GetStatus() (*StatusResponse, error) {
	resp, err := c.Get("/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[StatusResponse](resp)
}

// GetConfig gets full formation config
func (c *Client) GetConfig() (*ConfigResponse, error) {
	resp, err := c.Get("/config")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[ConfigResponse](resp)
}

// GetAgents lists all agents
func (c *Client) GetAgents() (*AgentListResponse, error) {
	resp, err := c.Get("/agents")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[AgentListResponse](resp)
}

// GetMCPServers lists all MCP servers
func (c *Client) GetMCPServers() (*MCPListResponse, error) {
	resp, err := c.Get("/mcp/servers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[MCPListResponse](resp)
}

// GetSecrets lists all secrets (keys only)
func (c *Client) GetSecrets() (*SecretsListResponse, error) {
	resp, err := c.Get("/secrets")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SecretsListResponse](resp)
}

// GetTriggers lists all triggers
func (c *Client) GetTriggers() (*TriggersListResponse, error) {
	resp, err := c.GetClient("/triggers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[TriggersListResponse](resp)
}

// GetSOPs lists all SOPs
func (c *Client) GetSOPs() (*SOPsListResponse, error) {
	resp, err := c.GetClient("/sops")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SOPsListResponse](resp)
}

// GetSOP gets a single SOP by name
func (c *Client) GetSOP(name string) (*SOP, error) {
	resp, err := c.GetClient("/sops/" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SOP](resp)
}

// GetSessions lists sessions for a user
func (c *Client) GetSessions(userID string) (*SessionsListResponse, error) {
	resp, err := c.GetWithUser("/sessions", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SessionsListResponse](resp)
}

// GetSessionMessages gets messages for a session
func (c *Client) GetSessionMessages(sessionID, userID string) (*SessionMessagesResponse, error) {
	resp, err := c.GetWithUser("/sessions/"+sessionID+"/messages", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SessionMessagesResponse](resp)
}

// DeleteSession deletes a session
func (c *Client) DeleteSession(sessionID, userID string) error {
	resp, err := c.DeleteWithUser("/sessions/"+sessionID, userID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// GetJobs lists jobs for a user
func (c *Client) GetJobs(userID string) (*JobsListResponse, error) {
	resp, err := c.GetWithUser("/jobs/"+userID, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[JobsListResponse](resp)
}

// CancelJob cancels a job
func (c *Client) CancelJob(userID, jobID string) error {
	resp, err := c.DeleteWithUser("/jobs/"+userID+"/"+jobID, userID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// GetAuditLog gets audit log entries
func (c *Client) GetAuditLog() (*AuditLogResponse, error) {
	resp, err := c.Get("/audit")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[AuditLogResponse](resp)
}

// ClearAuditLog clears the audit log
func (c *Client) ClearAuditLog() error {
	resp, err := c.Delete("/audit")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// parseResponse parses an API response into the target type
func parseResponse[T any](resp *http.Response) (*T, error) {
	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("request failed")
	}

	var result T
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}

	return &result, nil
}

// checkResponse checks if the HTTP response indicates an error
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	
	// Try to parse as API error
	var apiResp APIResponse
	if json.Unmarshal(body, &apiResp) == nil && apiResp.Error != nil {
		return fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized: invalid API key")
	case http.StatusForbidden:
		return fmt.Errorf("forbidden: insufficient permissions")
	case http.StatusNotFound:
		return fmt.Errorf("not found")
	default:
		return fmt.Errorf("request failed: %s", resp.Status)
	}
}
