package formation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
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

// DoWithUserID performs a request with X-Muxi-User-ID header (for client endpoints)
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
	req.Header.Set("X-Muxi-User-ID", userID)

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

// GetClientWithUser performs a GET request with client key and user ID
func (c *Client) GetClientWithUser(path, userID string) (*http.Response, error) {
	return c.DoWithUserID("GET", path, nil, userID)
}

// GetWithUser performs a GET request with client key and user ID
func (c *Client) GetWithUser(path, userID string) (*http.Response, error) {
	return c.DoWithUserID("GET", path, nil, userID)
}

// GetAdminWithUser performs a GET request with admin key and optional user ID
func (c *Client) GetAdminWithUser(path, userID string) (*http.Response, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.AdminKey == "" {
		return nil, fmt.Errorf("admin key required but not set")
	}
	req.Header.Set("X-MUXI-ADMIN-KEY", c.AdminKey)
	if userID != "" {
		req.Header.Set("X-Muxi-User-ID", userID)
	}
	return c.HTTPClient.Do(req)
}

// PostAdminWithUser performs a POST request with admin key and optional user ID
func (c *Client) PostAdminWithUser(path string, body interface{}, userID string) (*http.Response, error) {
	url := c.BaseURL + path
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, err
	}
	if c.AdminKey == "" {
		return nil, fmt.Errorf("admin key required but not set")
	}
	req.Header.Set("X-MUXI-ADMIN-KEY", c.AdminKey)
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-Muxi-User-ID", userID)
	}
	return c.HTTPClient.Do(req)
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

// PostClientWithUser performs a POST request with client key and user ID
func (c *Client) PostClientWithUser(path string, body interface{}, userID string) (*http.Response, error) {
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

// DeleteClientWithUser is an alias for DeleteWithUser (both use client key)
func (c *Client) DeleteClientWithUser(path, userID string) (*http.Response, error) {
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
func (c *Client) GetSessions(userID string, limit int) (*SessionsListResponse, error) {
	path := "/sessions"
	if limit > 0 {
		path = fmt.Sprintf("/sessions?limit=%d", limit)
	}
	resp, err := c.GetWithUser(path, userID)
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

// RestoreSession restores a session from messages
func (c *Client) RestoreSession(sessionID, userID string, messages []Message) error {
	payload := struct {
		Messages []Message `json:"messages"`
	}{Messages: messages}

	resp, err := c.PostWithUser("/sessions/"+sessionID+"/restore", payload, userID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// GetRequests lists requests for a user
func (c *Client) GetRequests(userID string) (*RequestsListResponse, error) {
	resp, err := c.GetWithUser("/requests", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[RequestsListResponse](resp)
}

// CancelRequest cancels a request
func (c *Client) CancelRequest(requestID, userID string) error {
	resp, err := c.DeleteWithUser("/requests/"+requestID, userID)
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
	resp, err := c.Delete("/audit?confirm=clear-audit-log")
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
func (c *Client) TriggerTrigger(name string, data json.RawMessage, async bool, userID string) (*TriggerResponse, error) {
	path := "/triggers/" + name

	body := TriggerRequest{
		Data:     data,
		UseAsync: async,
	}

	var resp *http.Response
	var err error
	if userID != "" {
		resp, err = c.PostClientWithUser(path, body, userID)
	} else {
		resp, err = c.PostClient(path, body)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response and extract request ID from envelope
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

	var triggerResp TriggerResponse
	if err := json.Unmarshal(apiResp.Data, &triggerResp); err != nil {
		return nil, fmt.Errorf("failed to parse trigger response: %w", err)
	}

	// Copy request ID from envelope
	triggerResp.RequestID = apiResp.Request.ID

	return &triggerResp, nil
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

// GetSchedulerJobs lists scheduler jobs (optionally filtered by user ID)
func (c *Client) GetSchedulerJobs(userID string) (*SchedulerJobsResponse, error) {
	resp, err := c.GetAdminWithUser("/scheduler/jobs", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SchedulerJobsResponse](resp)
}

// CreateSchedulerJob creates a new scheduled job
func (c *Client) CreateSchedulerJob(jobType, schedule, message, userID string) (*ScheduledJob, error) {
	body := CreateSchedulerJobRequest{
		Type:     jobType,
		Schedule: schedule,
		Message:  message,
	}

	resp, err := c.PostAdminWithUser("/scheduler/jobs", body, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[ScheduledJob](resp)
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
	resp, err := c.Get("/users/" + identifier)
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
func (c *Client) AddMemory(userID, memType, detail string) (*Memory, error) {
	body := map[string]interface{}{
		"content": map[string]string{
			"type":   memType,
			"detail": detail,
		},
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

// GetMCPConfig gets MCP configuration
func (c *Client) GetMCPConfig() (*MCPConfigResponse, error) {
	resp, err := c.Get("/mcp")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[MCPConfigResponse](resp)
}

// GetMCPTools lists all MCP tools
func (c *Client) GetMCPTools() (*MCPToolsResponse, error) {
	resp, err := c.Get("/mcp/tools")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[MCPToolsResponse](resp)
}

// GetUserBuffer gets buffer data for a user (GET /memory/buffer)
func (c *Client) GetUserBuffer(userID string) (*UserBufferResponse, error) {
	resp, err := c.GetWithUser("/memory/buffer", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[UserBufferResponse](resp)
}

// GetBufferStats gets aggregate buffer stats (GET /memory/stats)
func (c *Client) GetBufferStats() (*BufferStatsResponse, error) {
	resp, err := c.Get("/memory/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[BufferStatsResponse](resp)
}

// ClearUserBuffer clears buffer for a specific user (DELETE /memory/buffer with user header)
func (c *Client) ClearUserBuffer(userID string) (*BufferClearedResponse, error) {
	resp, err := c.DeleteWithUser("/memory/buffer", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[BufferClearedResponse](resp)
}

// ClearAllBuffers clears all buffers (DELETE /memory/buffer without user header, admin)
func (c *Client) ClearAllBuffers() (*BufferClearedResponse, error) {
	resp, err := c.Delete("/memory/buffer")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[BufferClearedResponse](resp)
}

// ClearSessionBuffer clears buffer for a specific session (DELETE /memory/buffer/{session_id})
func (c *Client) ClearSessionBuffer(sessionID, userID string) (*SessionBufferClearedResponse, error) {
	var resp *http.Response
	var err error
	if userID != "" {
		resp, err = c.DeleteWithUser("/memory/buffer/"+sessionID, userID)
	} else {
		resp, err = c.Delete("/memory/buffer/" + sessionID)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SessionBufferClearedResponse](resp)
}

// GetAsyncConfig gets async configuration
func (c *Client) GetAsyncConfig() (*AsyncSettingsResponse, error) {
	resp, err := c.Get("/async")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[AsyncSettingsResponse](resp)
}

// GetAsyncJobs lists async jobs (admin)
func (c *Client) GetAsyncJobs() (*AsyncJobsResponse, error) {
	resp, err := c.Get("/async/jobs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[AsyncJobsResponse](resp)
}

// GetAsyncJob gets async job details
func (c *Client) GetAsyncJob(jobID string) (*AsyncJobDetailResponse, error) {
	resp, err := c.Get("/async/jobs/" + jobID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[AsyncJobDetailResponse](resp)
}

// CancelAsyncJob cancels an async job
func (c *Client) CancelAsyncJob(jobID string) error {
	resp, err := c.Delete("/async/jobs/" + jobID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// GetSchedulerJob gets a specific scheduled job
func (c *Client) GetSchedulerJob(jobID string) (*SchedulerJobDetail, error) {
	resp, err := c.Get("/scheduler/jobs/" + jobID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SchedulerJobDetail](resp)
}

// DeleteSchedulerJob deletes a scheduled job
func (c *Client) DeleteSchedulerJob(jobID string) error {
	resp, err := c.Delete("/scheduler/jobs/" + jobID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// GetA2AConfig gets A2A configuration
func (c *Client) GetA2AConfig() (*A2AConfigResponse, error) {
	resp, err := c.Get("/a2a")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[A2AConfigResponse](resp)
}

// GetLoggingConfig gets logging configuration
func (c *Client) GetLoggingConfig() (*LoggingConfigResponse, error) {
	resp, err := c.Get("/logging")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[LoggingConfigResponse](resp)
}

// GetRequestStatus gets request status
func (c *Client) GetRequestStatus(requestID, userID string) (*RequestStatusResponse, error) {
	resp, err := c.GetWithUser("/requests/"+requestID, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[RequestStatusResponse](resp)
}

// ResolveUser resolves an identifier to a user, optionally creating
func (c *Client) ResolveUser(identifier string, createUser bool) (*UserResolveResponse, error) {
	body := UserResolveRequest{
		Identifier: identifier,
		CreateUser: createUser,
	}

	resp, err := c.PostClient("/users/resolve", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[UserResolveResponse](resp)
}

// Chat sends a chat message (non-streaming)
func (c *Client) Chat(req *ChatRequest) (*ChatResponse, error) {
	resp, err := c.PostClient("/chat", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[ChatResponse](resp)
}

// streamClient is used for SSE streaming requests
// Has dial timeout but no response timeout (streaming can run indefinitely)
var streamClient = &http.Client{
	Timeout: 0, // No timeout for streaming responses
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second, // Connection timeout
		}).DialContext,
		DisableCompression: true, // Important for SSE - don't buffer/decompress
	},
}

// ChatStream sends a chat message and returns SSE stream (caller must close)
func (c *Client) ChatStream(req *ChatRequest, userID string) (*http.Response, error) {
	url := c.BaseURL + "/chat"
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Create request without context timeout - we handle connection timeout via Transport
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")
	httpReq.Header.Set("X-MUXI-CLIENT-KEY", c.ClientKey)
	if userID != "" {
		httpReq.Header.Set("X-Muxi-User-ID", userID)
	}

	resp, err := streamClient.Do(httpReq)
	if err != nil {
		// Simplify connection refused errors
		if strings.Contains(err.Error(), "connection refused") {
			return nil, fmt.Errorf("formation not reachable - is it running?")
		}
		return nil, fmt.Errorf("connection failed: %v", err)
	}
	return resp, nil
}

// AudioChat sends audio for transcription and chat response (non-streaming)
func (c *Client) AudioChat(req *AudioChatRequest) (*ChatResponse, error) {
	resp, err := c.PostClient("/audiochat", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[ChatResponse](resp)
}

// AudioChatStream sends audio and returns SSE stream (caller must close)
func (c *Client) AudioChatStream(req *AudioChatRequest, userID string) (*http.Response, error) {
	url := c.BaseURL + "/audiochat"
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")
	httpReq.Header.Set("X-MUXI-CLIENT-KEY", c.ClientKey)
	if userID != "" {
		httpReq.Header.Set("X-Muxi-User-ID", userID)
	}

	return streamClient.Do(httpReq)
}

// GetFormationInfo gets basic formation info
func (c *Client) GetFormationInfo() (*FormationInfoResponse, error) {
	resp, err := c.Get("/formation")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[FormationInfoResponse](resp)
}

// GetOverlordPersona gets the overlord persona text
func (c *Client) GetOverlordPersona() (*OverlordPersonaResponse, error) {
	resp, err := c.Get("/overlord/persona")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[OverlordPersonaResponse](resp)
}

// GetSecret gets a single secret by key (masked)
func (c *Client) GetSecret(key string) (*SecretResponse, error) {
	resp, err := c.Get("/secrets/" + key)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SecretResponse](resp)
}

// SetSecret sets a secret value
func (c *Client) SetSecret(key, value string) error {
	body := map[string]string{"value": value}
	resp, err := c.Put("/secrets/"+key, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// DeleteSecret deletes a secret
func (c *Client) DeleteSecret(key string) error {
	resp, err := c.Delete("/secrets/" + key)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// GetSession gets details for a specific session
func (c *Client) GetSession(sessionID, userID string) (*SessionDetailResponse, error) {
	resp, err := c.GetWithUser("/sessions/"+sessionID, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[SessionDetailResponse](resp)
}

// GetLoggingDestinations lists logging destinations
func (c *Client) GetLoggingDestinations() (*LoggingDestinationsResponse, error) {
	resp, err := c.Get("/logging/destinations")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[LoggingDestinationsResponse](resp)
}

// BulkLinkIdentifiers associates multiple identifiers to a user
func (c *Client) BulkLinkIdentifiers(userID string, identifiers []string) (*BulkIdentifiersResponse, error) {
	// Convert strings to interface slice
	ids := make([]interface{}, len(identifiers))
	for i, id := range identifiers {
		ids[i] = id
	}

	body := BulkIdentifiersRequest{
		MuxiUserID:  userID,
		Identifiers: ids,
	}

	resp, err := c.Post("/users/identifiers", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[BulkIdentifiersResponse](resp)
}

// StreamEvents returns SSE stream for user events (caller must close)
func (c *Client) StreamEvents(userID string) (*http.Response, error) {
	url := c.BaseURL + "/events"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.ClientKey == "" {
		return nil, fmt.Errorf("client key required but not set")
	}
	req.Header.Set("X-MUXI-CLIENT-KEY", c.ClientKey)
	req.Header.Set("X-User-ID", userID)
	req.Header.Set("Accept", "text/event-stream")

	streamClient := &http.Client{}
	return streamClient.Do(req)
}

// StreamRequest returns SSE stream for a specific request (caller must close)
func (c *Client) StreamRequest(userID, sessionID, requestID string) (*http.Response, error) {
	url := c.BaseURL + "/stream/" + sessionID + "/" + requestID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.ClientKey == "" {
		return nil, fmt.Errorf("client key required but not set")
	}
	req.Header.Set("X-MUXI-CLIENT-KEY", c.ClientKey)
	req.Header.Set("X-User-ID", userID)
	req.Header.Set("Accept", "text/event-stream")

	streamClient := &http.Client{}
	return streamClient.Do(req)
}

// LogFilters for GET /logs/stream
type LogFilters struct {
	UserID    string
	SessionID string
	RequestID string
	AgentID   string
	Level     string
	EventType string
}

// StreamLogs returns the response body for SSE log streaming (caller must close)
func (c *Client) StreamLogs(filters LogFilters) (*http.Response, error) {
	return c.StreamLogsWithContext(context.Background(), filters)
}

// StreamLogsWithContext returns the response body for SSE log streaming with context support
func (c *Client) StreamLogsWithContext(ctx context.Context, filters LogFilters) (*http.Response, error) {
	path := "/logs/stream"
	params := []string{}
	if filters.UserID != "" {
		params = append(params, "user_id="+filters.UserID)
	}
	if filters.SessionID != "" {
		params = append(params, "session_id="+filters.SessionID)
	}
	if filters.RequestID != "" {
		params = append(params, "request_id="+filters.RequestID)
	}
	if filters.AgentID != "" {
		params = append(params, "agent_id="+filters.AgentID)
	}
	if filters.Level != "" {
		params = append(params, "level="+filters.Level)
	}
	if filters.EventType != "" {
		params = append(params, "event_type="+filters.EventType)
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	// Create request with context for cancellation
	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

// ListCredentialServices lists available credential services
func (c *Client) ListCredentialServices() (*CredentialServicesResponse, error) {
	resp, err := c.GetClient("/credentials/services")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[CredentialServicesResponse](resp)
}

// ListCredentials lists all credentials for a user
func (c *Client) ListCredentials(userID string) (*CredentialsListResponse, error) {
	resp, err := c.GetClientWithUser("/credentials", userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[CredentialsListResponse](resp)
}

// GetCredential gets a specific credential by ID
func (c *Client) GetCredential(credentialID, userID string) (*Credential, error) {
	resp, err := c.GetClientWithUser("/credentials/"+credentialID, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[Credential](resp)
}

// CreateCredential stores a new credential for a user
func (c *Client) CreateCredential(userID string, req *CreateCredentialRequest) (*CreateCredentialResponse, error) {
	resp, err := c.PostClientWithUser("/credentials", req, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[CreateCredentialResponse](resp)
}

// DeleteCredential deletes a credential
func (c *Client) DeleteCredential(credentialID, userID string) (*DeleteCredentialResponse, error) {
	resp, err := c.DeleteClientWithUser("/credentials/"+credentialID, userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[DeleteCredentialResponse](resp)
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
