package formation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	// Health endpoint returns plain JSON (no envelope)
	var result HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse health response: %w", err)
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

// GetAgent gets a single agent by ID (returns raw config as map)
func (c *Client) GetAgent(agentID string) (map[string]interface{}, error) {
	resp, err := c.Get("/agents/" + agentID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := parseResponse[map[string]interface{}](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
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

// GetMCPServer gets a single MCP server by ID (returns raw config as map)
func (c *Client) GetMCPServer(serverID string) (map[string]interface{}, error) {
	resp, err := c.Get("/mcp/servers/" + serverID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := parseResponse[map[string]interface{}](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
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

// GetTrigger gets details for a specific trigger
func (c *Client) GetTrigger(name string) (*TriggerDetail, error) {
	resp, err := c.GetClient("/triggers/" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[TriggerDetail](resp)
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

// GetLLMSettings gets LLM settings
func (c *Client) GetLLMSettings() (*LLMSettingsResponse, error) {
	resp, err := c.Get("/llm/settings")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[LLMSettingsResponse](resp)
}

// GetMemoryConfig gets memory configuration
func (c *Client) GetMemoryConfig() (*MemoryConfigResponse, error) {
	resp, err := c.Get("/memory")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[MemoryConfigResponse](resp)
}

// GetOverlordConfig gets overlord configuration
func (c *Client) GetOverlordConfig() (*OverlordConfigResponse, error) {
	resp, err := c.Get("/overlord")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[OverlordConfigResponse](resp)
}

// TriggerTrigger fires a trigger with optional data
func (c *Client) TriggerTrigger(name string, data json.RawMessage, async bool) (*TriggerResponse, error) {
	path := "/triggers/" + name
	if async {
		path += "?async=true"
	}

	body := TriggerRequest{Data: data}
	resp, err := c.PostClient(path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[TriggerResponse](resp)
}

// GetSchedulerConfig gets scheduler configuration
func (c *Client) GetSchedulerConfig() (*SchedulerConfigResponse, error) {
	resp, err := c.Get("/scheduler")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SchedulerConfigResponse](resp)
}

// GetSchedulerJobs lists scheduler jobs
func (c *Client) GetSchedulerJobs() (*SchedulerJobsResponse, error) {
	resp, err := c.Get("/scheduler/jobs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SchedulerJobsResponse](resp)
}

// GetUserIdentifiers lists all user identifier mappings (admin)
func (c *Client) GetUserIdentifiers() (*UserIdentifiersResponse, error) {
	resp, err := c.Get("/users/identifiers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[UserIdentifiersResponse](resp)
}

// GetUserIdentifiersForUser lists identifier mappings for a specific user
func (c *Client) GetUserIdentifiersForUser(userID string) (*UserIdentifiersResponse, error) {
	resp, err := c.GetWithUser("/users/identifiers", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[UserIdentifiersResponse](resp)
}

// LinkUserIdentifier links an identifier to a user ID
func (c *Client) LinkUserIdentifier(identifier, userID, idType string) error {
	body := map[string]string{
		"identifier": identifier,
		"user_id":    userID,
		"type":       idType,
	}
	resp, err := c.Post("/users/identifiers", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// UnlinkUserIdentifier removes an identifier mapping
func (c *Client) UnlinkUserIdentifier(identifier string) error {
	resp, err := c.Delete("/users/identifiers/" + identifier)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// ResolveUserIdentifier resolves an identifier to a user
func (c *Client) ResolveUserIdentifier(identifier string) (*UserIdentifier, error) {
	resp, err := c.Get("/users/identifiers/" + identifier)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[UserIdentifier](resp)
}

// GetMemories lists memories for a user
func (c *Client) GetMemories(userID string) (*MemoriesListResponse, error) {
	resp, err := c.GetWithUser("/memories", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[MemoriesListResponse](resp)
}

// AddMemory adds a memory for a user
func (c *Client) AddMemory(userID, content string) (*Memory, error) {
	body := map[string]string{
		"content": content,
	}

	resp, err := c.PostWithUser("/memories", body, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[Memory](resp)
}

// DeleteMemory deletes a memory
func (c *Client) DeleteMemory(userID, memoryID string) error {
	resp, err := c.DeleteWithUser("/memories/"+memoryID, userID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// StreamLogs returns the response body for SSE log streaming (caller must close)
func (c *Client) StreamLogs(userID, level, agent, requestID string) (*http.Response, error) {
	path := "/logs/stream"
	params := []string{}
	if userID != "" {
		params = append(params, "user_id="+userID)
	}
	if level != "" {
		params = append(params, "level="+level)
	}
	if agent != "" {
		params = append(params, "agent="+agent)
	}
	if requestID != "" {
		params = append(params, "request_id="+requestID)
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	// Create request with no timeout for streaming
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.AdminKey == "" {
		return nil, fmt.Errorf("admin key required but not set")
	}
	req.Header.Set("X-MUXI-ADMIN-KEY", c.AdminKey)
	req.Header.Set("Accept", "text/event-stream")

	// Use a client with no timeout for streaming
	streamClient := &http.Client{}
	return streamClient.Do(req)
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
