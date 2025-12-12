package formation

import (
	"encoding/json"
	"time"
)

// APIResponse is the standard envelope for all Formation API responses
type APIResponse struct {
	Object    string          `json:"object"`
	Timestamp int64           `json:"timestamp"`
	Type      string          `json:"type"`
	Request   RequestInfo     `json:"request"`
	Success   bool            `json:"success"`
	Error     *APIError       `json:"error,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// RequestInfo contains request tracking information
type RequestInfo struct {
	ID             string `json:"id"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// APIError represents an error response
type APIError struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// HealthResponse from GET /health
type HealthResponse struct {
	Status      string `json:"status"`
	FormationID string `json:"formation_id,omitempty"`
	Version     string `json:"version,omitempty"`
}

// StatusResponse from GET /status
type StatusResponse struct {
	Formation struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
	} `json:"formation"`
	Server struct {
		Version       string `json:"version"`
		UptimeSeconds int64  `json:"uptime_seconds"`
	} `json:"server"`
	Agents struct {
		Count  int `json:"count"`
		Active int `json:"active"`
	} `json:"agents"`
	MCPServers struct {
		Count  int `json:"count"`
		Active int `json:"active"`
	} `json:"mcp_servers"`
	Stats struct {
		Running struct {
			Seconds int64 `json:"seconds"`
			Since   int64 `json:"since"`
		} `json:"running"`
		Memory struct {
			WorkingMemoryMB float64 `json:"working_memory_mb"`
			MemoryUsageMB   float64 `json:"memory_usage_mb"`
		} `json:"memory"`
		Requests struct {
			Total  int `json:"total"`
			Active int `json:"active"`
		} `json:"requests"`
		BufferSize int     `json:"buffer_size"`
		CPUPercent float64 `json:"cpu_percent"`
	} `json:"stats"`
}

// ConfigResponse from GET /config
type ConfigResponse struct {
	FormationID   string `json:"formation_id"`
	Version       string `json:"version"`
	Description   string `json:"description"`
	SchemaVersion string `json:"schema_version"`
	Agents        struct {
		Total    int    `json:"total"`
		Resource string `json:"resource"`
	} `json:"agents"`
	Secrets struct {
		Total    int    `json:"total"`
		Resource string `json:"resource"`
	} `json:"secrets"`
	MCP struct {
		DefaultRetryAttempts  int    `json:"default_retry_attempts"`
		DefaultTimeoutSeconds int    `json:"default_timeout_seconds"`
		Servers               struct {
			Total    int    `json:"total"`
			Resource string `json:"resource"`
		} `json:"servers"`
	} `json:"mcp"`
	Overlord struct {
		Resource string `json:"resource"`
	} `json:"overlord"`
	LLM struct {
		Resource string `json:"resource"`
	} `json:"llm"`
	Memory struct {
		Resource string `json:"resource"`
	} `json:"memory"`
	Async struct {
		Resource string `json:"resource"`
	} `json:"async"`
	Scheduler struct {
		Resource string `json:"resource"`
	} `json:"scheduler"`
	A2A struct {
		Resource string `json:"resource"`
	} `json:"a2a"`
	Logging struct {
		Resource string `json:"resource"`
	} `json:"logging"`
}

// Agent represents an agent in the formation
type Agent struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Description string   `json:"description,omitempty"`
	Model       string   `json:"model,omitempty"`
	Provider    string   `json:"provider,omitempty"`
	Enabled     bool     `json:"enabled"`
	Status      string   `json:"status,omitempty"`
	Tools       []string `json:"tools,omitempty"`
	MCPServers  []string `json:"mcp_servers,omitempty"`
}

// AgentListResponse from GET /agents
type AgentListResponse struct {
	Agents []Agent `json:"agents"`
	Count  int     `json:"count"`
}

// MCPServer represents an MCP server
type MCPServer struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"` // command, http, sse
	Status      string   `json:"status"`
	Enabled     bool     `json:"enabled"`
	Tools       []string `json:"tools,omitempty"`
	ToolsCount  int      `json:"tools_count,omitempty"`
	Description string   `json:"description,omitempty"`
}

// MCPListResponse from GET /mcp/servers
type MCPListResponse struct {
	Servers []MCPServer `json:"servers"`
	Count   int         `json:"count"`
}

// MCPConfigResponse from GET /mcp
type MCPConfigResponse struct {
	DefaultRetryAttempts  int `json:"default_retry_attempts"`
	DefaultTimeoutSeconds int `json:"default_timeout_seconds"`
	Servers               struct {
		Total    int    `json:"total"`
		Resource string `json:"resource"`
	} `json:"servers"`
}

// SecretsListResponse from GET /secrets
type SecretsListResponse struct {
	Secrets map[string]string `json:"secrets"` // Key -> masked value
	Count   int               `json:"count"`
}

// TriggersListResponse from GET /triggers
// Note: API returns trigger names as strings, not objects
type TriggersListResponse struct {
	Triggers []string `json:"triggers"`
	Count    int      `json:"count"`
}

// TriggerDetail from GET /triggers/{name}
type TriggerDetail struct {
	Name       string   `json:"name"`
	Content    string   `json:"content"`
	DataFields []string `json:"data_fields,omitempty"`
}

// SOP represents a Standard Operating Procedure
type SOP struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"` // template, guide
	Steps       int      `json:"steps,omitempty"`
	Agents      []string `json:"agents,omitempty"`
	Content     string   `json:"content,omitempty"` // Full content when fetching single SOP
}

// SOPsListResponse from GET /sops
type SOPsListResponse struct {
	SOPs  []SOP `json:"sops"`
	Count int   `json:"count"`
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	MessageCount int       `json:"message_count"`
	LastActivity time.Time `json:"last_activity"`
	Status       string    `json:"status"` // active, inactive
	CreatedAt    time.Time `json:"created_at"`
}

// SessionsListResponse from GET /sessions
type SessionsListResponse struct {
	Sessions []Session `json:"sessions"`
	Count    int       `json:"count"`
}

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Agent     string    `json:"agent,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionMessagesResponse from GET /sessions/{id}/messages
type SessionMessagesResponse struct {
	SessionID string    `json:"session_id"`
	Messages  []Message `json:"messages"`
	Count     int       `json:"count"`
}

// Job represents an async job
type Job struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"` // pending, processing, completed, failed, cancelled
	Progress  int       `json:"progress,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Result    string    `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// JobsListResponse from GET /jobs/{user_id}
type JobsListResponse struct {
	Jobs  []Job `json:"jobs"`
	Count int   `json:"count"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	User      string    `json:"user,omitempty"`
	Details   string    `json:"details,omitempty"`
}

// AuditLogResponse from GET /audit
type AuditLogResponse struct {
	Entries []AuditEntry `json:"entries"`
	Count   int          `json:"count"`
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // one_time, recurring
	Schedule  string    `json:"schedule,omitempty"`
	RunAt     time.Time `json:"run_at,omitempty"`
	Message   string    `json:"message"`
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id,omitempty"`
	Enabled   bool      `json:"enabled"`
	LastRun   time.Time `json:"last_run,omitempty"`
	NextRun   time.Time `json:"next_run,omitempty"`
}

// SchedulerJobsResponse from GET /scheduler/jobs
type SchedulerJobsResponse struct {
	Jobs  []ScheduledJob `json:"jobs"`
	Count int            `json:"count"`
}

// SchedulerConfigResponse from GET /scheduler
type SchedulerConfigResponse struct {
	Enabled               bool   `json:"enabled"`
	Timezone              string `json:"timezone"`
	CheckIntervalMinutes  int    `json:"check_interval_minutes"`
	MaxConcurrentJobs     int    `json:"max_concurrent_jobs"`
	MaxFailuresBeforePause int   `json:"max_failures_before_pause"`
}

// UserIdentifier represents a user identifier mapping
type UserIdentifier struct {
	Identifier string    `json:"identifier"`
	UserID     string    `json:"user_id"`
	Type       string    `json:"type,omitempty"` // email, phone, external_id
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

// UserIdentifiersResponse from GET /users/identifiers
type UserIdentifiersResponse struct {
	Identifiers []UserIdentifier `json:"identifiers"`
	Count       int              `json:"count"`
}

// Memory represents a user memory entry
type Memory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	Type      string    `json:"type,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// MemoriesListResponse from GET /memories
type MemoriesListResponse struct {
	Memories []Memory `json:"memories"`
	Count    int      `json:"count"`
}

// MemoryConfigResponse from GET /memory
type MemoryConfigResponse struct {
	Buffer struct {
		Size         int     `json:"size"`
		Multiplier   float64 `json:"multiplier"`
		VectorSearch bool    `json:"vector_search"`
	} `json:"buffer"`
	Working struct {
		MaxMemoryMB     int `json:"max_memory_mb"`
		FIFOIntervalMin int `json:"fifo_interval_min"`
	} `json:"working"`
}

// MemoryBufferResponse from GET /memory/buffer
type MemoryBufferResponse struct {
	UserID       string `json:"user_id"`
	MessageCount int    `json:"message_count"`
	SizeBytes    int    `json:"size_bytes"`
}

// LLMSettingsResponse from GET /llm/settings
type LLMSettingsResponse struct {
	Temperature    float64 `json:"temperature"`
	MaxTokens      int     `json:"max_tokens"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	DefaultModel   string  `json:"default_model,omitempty"`
	DefaultProvider string  `json:"default_provider,omitempty"`
}

// AsyncSettingsResponse from GET /async
type AsyncSettingsResponse struct {
	ThresholdSeconds      int `json:"threshold_seconds"`
	WebhookTimeoutSeconds int `json:"webhook_timeout_seconds"`
	WebhookRetryAttempts  int `json:"webhook_retry_attempts"`
	WebhookMaxPayloadMB   int `json:"webhook_max_payload_mb"`
}

// A2AConfigResponse from GET /a2a
type A2AConfigResponse struct {
	Inbound struct {
		Enabled bool `json:"enabled"`
	} `json:"inbound"`
	Outbound struct {
		Enabled               bool     `json:"enabled"`
		DefaultRetryAttempts  int      `json:"default_retry_attempts"`
		DefaultTimeoutSeconds int      `json:"default_timeout_seconds"`
		AllowedFormations     []string `json:"allowed_formations,omitempty"`
	} `json:"outbound"`
}

// LoggingConfigResponse from GET /logging
type LoggingConfigResponse struct {
	Destinations []LoggingDestination `json:"destinations"`
}

// LoggingDestination represents a logging destination
type LoggingDestination struct {
	ID          string `json:"id"`
	Transport   string `json:"transport"` // stdout, file, stream
	Destination string `json:"destination,omitempty"`
	Level       string `json:"level"`
	Format      string `json:"format"` // text, jsonl
	Enabled     bool   `json:"enabled"`
}

// OverlordConfigResponse from GET /overlord
type OverlordConfigResponse struct {
	Persona    string `json:"persona,omitempty"`
	SystemNote string `json:"system_note,omitempty"`
	Routing    struct {
		Strategy string `json:"strategy,omitempty"`
		Fallback string `json:"fallback,omitempty"`
	} `json:"routing,omitempty"`
}

// ChatRequest for POST /chat
type ChatRequest struct {
	Message          string `json:"message"`
	UserID           string `json:"user_id,omitempty"`
	SessionID        string `json:"session_id,omitempty"`
	GroupID          string `json:"group_id,omitempty"`
	Stream           bool   `json:"stream,omitempty"`
	WebhookURL       string `json:"webhook_url,omitempty"`
	ThresholdSeconds int    `json:"threshold_seconds,omitempty"`
}

// ChatResponse from POST /chat (non-streaming)
type ChatResponse struct {
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
	Agent     string `json:"agent,omitempty"`
	Model     string `json:"model,omitempty"`
	Usage     struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// TriggerRequest for POST /triggers/{name}
type TriggerRequest struct {
	Data json.RawMessage `json:"data"`
}

// TriggerResponse from POST /triggers/{name}
type TriggerResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"` // completed, async
	Response  string `json:"response,omitempty"`
	JobID     string `json:"job_id,omitempty"`
}

// LogStreamEvent from GET /logs/stream (SSE)
type LogStreamEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	User      string    `json:"user,omitempty"`
	Agent     string    `json:"agent,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Session   string    `json:"session,omitempty"`
}
