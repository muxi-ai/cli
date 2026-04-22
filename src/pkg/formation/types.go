package formation

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// FlexTime handles both ISO 8601 strings and Unix timestamps
type FlexTime struct {
	time.Time
}

var apiTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999999999Z0700",
	"2006-01-02T15:04:05Z0700",
	"2006-01-02T15:04:05.999999999Z",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
}

func (ft *FlexTime) UnmarshalJSON(data []byte) error {
	// Try string first (ISO 8601)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			ft.Time = time.Time{}
			return nil
		}
		t, err := parseAPITimeString(s)
		if err != nil {
			return fmt.Errorf("cannot parse time string: %s", s)
		}
		ft.Time = t
		return nil
	}

	// Try Unix timestamp (seconds or milliseconds)
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		if n > 1e12 {
			// Milliseconds
			ft.Time = time.UnixMilli(n)
		} else {
			// Seconds
			ft.Time = time.Unix(n, 0)
		}
		return nil
	}

	// Try float (seconds with decimals)
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		ft.Time = time.Unix(int64(f), int64((f-float64(int64(f)))*1e9))
		return nil
	}

	return fmt.Errorf("cannot unmarshal time from: %s", string(data))
}

func parseAPITimeString(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	for _, layout := range apiTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time string: %s", s)
}

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

// HealthResponse from GET /health (plain JSON, no envelope)
type HealthResponse struct {
	Status      string `json:"status"`
	FormationID string `json:"formation_id,omitempty"`
	Version     string `json:"version,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
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
		DefaultRetryAttempts  int `json:"default_retry_attempts"`
		DefaultTimeoutSeconds int `json:"default_timeout_seconds"`
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
	Agents    []Agent `json:"agents"`
	AgentList []Agent `json:"agent_list"` // Alternative field name per spec
	Count     int     `json:"count"`
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
	Servers    []MCPServer `json:"servers"`
	MCPServers []MCPServer `json:"mcp_servers"` // Alternative field name
	Count      int         `json:"count"`
}

// MCPConfigResponse from GET /mcp
type MCPConfigResponse struct {
	Defaults struct {
		RetryAttempts  int `json:"retry_attempts"`
		TimeoutSeconds int `json:"timeout_seconds"`
	} `json:"defaults"`
	// Legacy fields for backward compatibility
	DefaultRetryAttempts  int `json:"default_retry_attempts,omitempty"`
	DefaultTimeoutSeconds int `json:"default_timeout_seconds,omitempty"`
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
	ID           string    `json:"session_id"`
	UserID       string    `json:"user_id,omitempty"`
	LastActivity *FlexTime `json:"last_activity,omitempty"`
	Active       bool      `json:"active,omitempty"`
	CreatedAt    *FlexTime `json:"created_at,omitempty"`
}

// SessionsListResponse from GET /sessions
type SessionsListResponse struct {
	Sessions []Session `json:"sessions"`
	Count    int       `json:"count"`
}

// Message represents a chat message
type Message struct {
	ID        string           `json:"id,omitempty"`
	Text      string           `json:"text"`
	Content   string           `json:"content,omitempty"` // alias for text
	Timestamp *FlexTime        `json:"timestamp,omitempty"`
	Metadata  *MessageMetadata `json:"metadata,omitempty"`
}

// MessageMetadata contains message metadata
type MessageMetadata struct {
	Role      string `json:"role"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
}

// GetContent returns text or content
func (m *Message) GetContent() string {
	if m.Text != "" {
		return m.Text
	}
	return m.Content
}

// GetRole returns role from metadata or empty
func (m *Message) GetRole() string {
	if m.Metadata != nil {
		return m.Metadata.Role
	}
	return ""
}

// GetAgent returns agent from metadata
func (m *Message) GetAgent() string {
	if m.Metadata != nil {
		return m.Metadata.AgentID
	}
	return ""
}

// SessionMessagesResponse from GET /sessions/{id}/messages
type SessionMessagesResponse struct {
	SessionID string    `json:"session_id"`
	Messages  []Message `json:"messages"`
	Count     int       `json:"count"`
}

// RequestItem represents a request in the list
type RequestItem struct {
	RequestID string    `json:"request_id"`
	Status    string    `json:"status"` // processing, completed, failed, cancelled
	Progress  int       `json:"progress,omitempty"`
	CreatedAt *FlexTime `json:"created_at,omitempty"`
}

// RequestsListResponse from GET /requests
type RequestsListResponse struct {
	Requests []RequestItem `json:"requests"`
	Count    int           `json:"count"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp    *FlexTime `json:"timestamp,omitempty"`
	RequestID    string    `json:"request_id,omitempty"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type,omitempty"`
	ResourceID   string    `json:"resource_id,omitempty"`
	User         string    `json:"user,omitempty"`
	IP           string    `json:"ip,omitempty"`
	Result       string    `json:"result,omitempty"`
	StatusCode   int       `json:"status_code,omitempty"`
	Message      string    `json:"message,omitempty"`
}

// AuditLogResponse from GET /audit
type AuditLogResponse struct {
	Entries      []AuditEntry `json:"entries"`
	Count        int          `json:"count"`
	TotalEntries int          `json:"total_entries,omitempty"`
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"` // one_time, recurring
	Schedule     string            `json:"schedule,omitempty"`
	RunAt        time.Time         `json:"run_at,omitempty"`
	Message      string            `json:"message"`
	UserID       string            `json:"user_id"`
	SessionID    string            `json:"session_id,omitempty"`
	Enabled      bool              `json:"enabled"`
	Status       string            `json:"status,omitempty"`
	LastRun      time.Time         `json:"last_run,omitempty"`
	NextRun      time.Time         `json:"next_run,omitempty"`
	FailureCount int               `json:"failure_count,omitempty"`
	History      []ScheduledJobRun `json:"history,omitempty"`
}

// ScheduledJobRun represents a single execution of a scheduled job
type ScheduledJobRun struct {
	RunAt      time.Time `json:"run_at"`
	Status     string    `json:"status"` // completed, failed
	DurationMs int       `json:"duration_ms,omitempty"`
}

// SchedulerJobsResponse from GET /scheduler/jobs
type SchedulerJobsResponse struct {
	Jobs  []ScheduledJob `json:"jobs"`
	Count int            `json:"count"`
}

// SchedulerConfigResponse from GET /scheduler
type SchedulerConfigResponse struct {
	Enabled                bool   `json:"enabled"`
	Timezone               string `json:"timezone"`
	CheckIntervalMinutes   int    `json:"check_interval_minutes"`
	MaxConcurrentJobs      int    `json:"max_concurrent_jobs"`
	MaxFailuresBeforePause int    `json:"max_failures_before_pause"`
}

// CreateSchedulerJobRequest for POST /scheduler/jobs
type CreateSchedulerJobRequest struct {
	Type     string `json:"type"`     // one_time, recurring
	Schedule string `json:"schedule"` // cron expr or ISO datetime
	Message  string `json:"message"`  // prompt to send
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

// MemoryContent represents the content of a memory entry
type MemoryContent struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

// Memory represents a user memory entry
type Memory struct {
	ID        string        `json:"id"`
	UserID    string        `json:"user_id,omitempty"`
	Content   MemoryContent `json:"content"`
	CreatedAt time.Time     `json:"created_at"`
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

// UserBufferResponse from GET /memory/buffer
type UserBufferResponse struct {
	UserID        string          `json:"user_id"`
	TotalMessages int             `json:"total_messages"`
	Sessions      []BufferSession `json:"sessions"`
	BufferSizeKB  float64         `json:"buffer_size_kb"`
}

// BufferSession represents a session in user's buffer
type BufferSession struct {
	SessionID    string    `json:"session_id"`
	MessageCount int       `json:"message_count"`
	LastActivity time.Time `json:"last_activity"`
}

// BufferStatsResponse from GET /memory/buffer/stats
type BufferStatsResponse struct {
	TotalEntries  int     `json:"total_entries"`
	TotalUsers    int     `json:"total_users"`
	TotalSessions int     `json:"total_sessions"`
	BufferSizeKB  float64 `json:"buffer_size_kb"`
	MaxSize       int     `json:"max_size"`
	Utilization   float64 `json:"utilization"`
}

// BufferClearedResponse from DELETE /memory/buffer
type BufferClearedResponse struct {
	Message         string `json:"message"`
	UserID          string `json:"user_id,omitempty"`
	MessagesCleared int    `json:"messages_cleared"`
	SessionsCleared int    `json:"sessions_cleared"`
}

// SessionBufferClearedResponse from DELETE /memory/buffer/{session_id}
type SessionBufferClearedResponse struct {
	Message         string `json:"message"`
	UserID          string `json:"user_id,omitempty"`
	SessionID       string `json:"session_id"`
	MessagesCleared int    `json:"messages_cleared"`
}

// LLMSettingsResponse from GET /llm/settings
type LLMSettingsResponse struct {
	APIKeys  map[string]string `json:"api_keys,omitempty"`
	Models   []LLMModelConfig  `json:"models,omitempty"`
	Settings LLMGlobalSettings `json:"settings,omitempty"`
}

// LLMModelConfig represents a model configuration
type LLMModelConfig struct {
	// Model type fields (one of these will be set)
	Text      string `json:"text,omitempty"`
	Streaming string `json:"streaming,omitempty"`
	Embedding string `json:"embedding,omitempty"`
	Documents string `json:"documents,omitempty"`
	Audio     string `json:"audio,omitempty"`
	Vision    string `json:"vision,omitempty"`

	Settings map[string]interface{} `json:"settings,omitempty"`
}

// LLMGlobalSettings represents global LLM settings
type LLMGlobalSettings struct {
	Temperature    float64                `json:"temperature,omitempty"`
	MaxTokens      int                    `json:"max_tokens,omitempty"`
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"`
	MaxRetries     int                    `json:"max_retries,omitempty"`
	FallbackModel  string                 `json:"fallback_model,omitempty"`
	Caching        map[string]interface{} `json:"caching,omitempty"`
}

// AsyncSettingsResponse from GET /async
type AsyncSettingsResponse struct {
	ThresholdSeconds int    `json:"threshold_seconds"`
	EnableEstimation bool   `json:"enable_estimation"`
	WebhookURL       string `json:"webhook_url,omitempty"`
	WebhookRetries   int    `json:"webhook_retries"`
	WebhookTimeout   int    `json:"webhook_timeout"`
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
	System       LoggingSystemConfig       `json:"system"`
	Conversation LoggingConversationConfig `json:"conversation"`
}

// LoggingSystemConfig represents system logging configuration
type LoggingSystemConfig struct {
	Level       string `json:"level"`
	Destination string `json:"destination"`
}

// LoggingConversationConfig represents conversation logging configuration
type LoggingConversationConfig struct {
	Enabled bool            `json:"enabled"`
	Streams []LoggingStream `json:"streams"`
}

// LoggingStream represents a logging stream in the config
type LoggingStream struct {
	Transport   string                 `json:"transport"` // stdout, file, stream
	Destination string                 `json:"destination,omitempty"`
	Level       string                 `json:"level"`
	Format      string                 `json:"format"` // text, jsonl
	Protocol    string                 `json:"protocol,omitempty"`
	Auth        map[string]interface{} `json:"auth,omitempty"`
	Events      []string               `json:"events,omitempty"`
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
	Soul       string `json:"soul,omitempty"`
	SystemNote string `json:"system_note,omitempty"`

	Clarification struct {
		Style              string `json:"style,omitempty"`
		PersistLearnedInfo bool   `json:"persist_learned_info,omitempty"`
		MaxRounds          struct {
			Direct     int `json:"direct,omitempty"`
			Brainstorm int `json:"brainstorm,omitempty"`
			Planning   int `json:"planning,omitempty"`
			Execution  int `json:"execution,omitempty"`
		} `json:"max_rounds,omitempty"`
	} `json:"clarification,omitempty"`

	Workflow struct {
		RoutingStrategy       string `json:"routing_strategy,omitempty"`
		AutoDecomposition     bool   `json:"auto_decomposition,omitempty"`
		PlanApprovalThreshold int    `json:"plan_approval_threshold,omitempty"`
		ComplexityMethod      string `json:"complexity_method,omitempty"`
		ParallelExecution     bool   `json:"parallel_execution,omitempty"`
		MaxParallelTasks      int    `json:"max_parallel_tasks,omitempty"`
		EnableAgentAffinity   bool   `json:"enable_agent_affinity,omitempty"`
		ErrorRecovery         string `json:"error_recovery,omitempty"`
		Timeouts              struct {
			TaskTimeout     int `json:"task_timeout,omitempty"`
			WorkflowTimeout int `json:"workflow_timeout,omitempty"`
		} `json:"timeouts,omitempty"`
	} `json:"workflow,omitempty"`

	Response struct {
		Format    string `json:"format,omitempty"`
		Streaming bool   `json:"streaming,omitempty"`
		Progress  bool   `json:"progress,omitempty"`
	} `json:"response,omitempty"`

	LLM struct {
		Provider    string  `json:"provider,omitempty"`
		Model       string  `json:"model,omitempty"`
		Temperature float64 `json:"temperature,omitempty"`
		MaxTokens   int     `json:"max_tokens,omitempty"`
	} `json:"llm,omitempty"`

	Caching struct {
		Enabled bool `json:"enabled,omitempty"`
		TTL     int  `json:"ttl,omitempty"`
	} `json:"caching,omitempty"`
}

// ChatFile represents a file attachment for chat
type ChatFile struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`      // Base64 encoded
	ContentType string `json:"content_type"` // MIME type
	Size        int64  `json:"size,omitempty"`
}

// ChatRequest for POST /chat
type ChatRequest struct {
	Message          string     `json:"message"`
	UserID           string     `json:"user_id,omitempty"`
	SessionID        string     `json:"session_id,omitempty"`
	GroupID          string     `json:"group_id,omitempty"`
	Stream           bool       `json:"stream,omitempty"`
	UseAsync         *bool      `json:"use_async,omitempty"` // nil=auto, true=force async, false=force sync
	WebhookURL       string     `json:"webhook_url,omitempty"`
	ThresholdSeconds int        `json:"threshold_seconds,omitempty"`
	Files            []ChatFile `json:"files,omitempty"`
}

// AudioChatRequest for POST /audiochat (voice notes)
type AudioChatRequest struct {
	Files     []ChatFile `json:"files"`
	UserID    string     `json:"user_id,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
	AgentName string     `json:"agent_name,omitempty"`
	Stream    bool       `json:"stream,omitempty"`
}

// ChatResponse from POST /chat (non-streaming)
type ChatResponse struct {
	RequestID string          `json:"request_id"`
	SessionID string          `json:"session_id"`
	Response  json.RawMessage `json:"response"`
	Agent     string          `json:"agent,omitempty"`
	Model     string          `json:"model,omitempty"`
	Usage     struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// GetResponseText extracts the text content from the response field
func (r *ChatResponse) GetResponseText() string {
	if r.Response == nil {
		return ""
	}

	// Try as plain string
	var str string
	if json.Unmarshal(r.Response, &str) == nil {
		return str
	}

	// Try as object with content field
	var obj struct {
		Content string `json:"content"`
	}
	if json.Unmarshal(r.Response, &obj) == nil && obj.Content != "" {
		return obj.Content
	}

	// Try as array of content parts
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(r.Response, &parts) == nil && len(parts) > 0 {
		var texts []string
		for _, p := range parts {
			if p.Text != "" {
				texts = append(texts, p.Text)
			}
		}
		return strings.Join(texts, "\n")
	}

	return string(r.Response)
}

// GetResponseArtifacts extracts artifacts from the response field
func (r *ChatResponse) GetResponseArtifacts() []Artifact {
	if r.Response == nil {
		return nil
	}

	var obj struct {
		Artifacts []Artifact `json:"artifacts"`
	}
	if json.Unmarshal(r.Response, &obj) == nil && len(obj.Artifacts) > 0 {
		return obj.Artifacts
	}

	return nil
}

// Artifact represents a file artifact generated by the formation
type Artifact struct {
	Type     string           `json:"type"`   // text, document, image, data
	Format   string           `json:"format"` // pdf, png, csv, etc.
	Filename string           `json:"filename"`
	Content  *string          `json:"content"`  // Raw text (text-type only)
	DataURL  *string          `json:"data_url"` // Base64 data URL (binary)
	Metadata ArtifactMetadata `json:"metadata"`
}

// ArtifactMetadata contains metadata about an artifact
type ArtifactMetadata struct {
	SizeBytes int    `json:"size_bytes,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Pages     int    `json:"pages,omitempty"`
	Lines     int    `json:"lines,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

// TriggerRequest for POST /triggers/{name}
type TriggerRequest struct {
	Data      json.RawMessage `json:"data"`
	SessionID string          `json:"session_id,omitempty"`
	UseAsync  bool            `json:"use_async"`
}

// TriggerResponse from POST /triggers/{name}
// Async: status="processing", Content empty
// Sync: status="completed", Content has response
type TriggerResponse struct {
	RequestID string `json:"-"`                 // From envelope request.id (not in data)
	Status    string `json:"status"`            // "processing" or "completed"
	Content   string `json:"content,omitempty"` // Response content (sync only)
}

// LogStreamEvent from GET /logs/stream (SSE)
type LogStreamEvent struct {
	Timestamp int64                  `json:"timestamp"` // Unix ms
	Level     string                 `json:"level"`
	EventType string                 `json:"event_type,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	AgentID   string                 `json:"agent_id,omitempty"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// MCPTool represents an MCP tool
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Server      string                 `json:"server"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

// MCPToolsResponse from GET /mcp/tools
type MCPToolsResponse struct {
	Tools []MCPTool `json:"tools"`
	Count int       `json:"count"`
}

// MemoryBufferStatus represents a single buffer status
type MemoryBufferStatus struct {
	UserID       string `json:"user_id"`
	SessionCount int    `json:"session_count"`
	MessageCount int    `json:"message_count"`
	SizeBytes    int    `json:"size_bytes"`
}

// MemoryBuffersResponse from GET /memory/buffers (admin)
type MemoryBuffersResponse struct {
	Buffers []MemoryBufferStatus `json:"buffers"`
	Count   int                  `json:"count"`
}

// AsyncJob represents an async job
type AsyncJob struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// AsyncJobsResponse from GET /async/jobs
type AsyncJobsResponse struct {
	Jobs  []AsyncJob `json:"jobs"`
	Count int        `json:"count"`
}

// AsyncJobDetailResponse from GET /async/jobs/{job_id}
type AsyncJobDetailResponse struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Status    string                 `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	Result    map[string]interface{} `json:"result,omitempty"`
}

// RequestStatusResponse from GET /requests/{request_id}
type RequestStatusResponse struct {
	RequestID   string    `json:"request_id"`
	Status      string    `json:"status"` // processing, completed, failed
	Progress    string    `json:"progress,omitempty"`
	Error       string    `json:"error,omitempty"`
	CompletedAt *FlexTime `json:"completed_at,omitempty"`
}

// SchedulerJobDetail from GET /scheduler/jobs/{job_id}
type SchedulerJobDetail struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Schedule     string    `json:"schedule,omitempty"`
	RunAt        time.Time `json:"run_at,omitempty"`
	Message      string    `json:"message"`
	UserID       string    `json:"user_id"`
	Enabled      bool      `json:"enabled"`
	Status       string    `json:"status,omitempty"`
	NextRun      time.Time `json:"next_run,omitempty"`
	LastRun      time.Time `json:"last_run,omitempty"`
	FailureCount int       `json:"failure_count"`
}

var schedulerCronParser = cron.NewParser(
	cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

type schedulerJobWire struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	Schedule       string            `json:"schedule,omitempty"`
	RunAt          *FlexTime         `json:"run_at,omitempty"`
	Message        string            `json:"message,omitempty"`
	UserID         string            `json:"user_id,omitempty"`
	SessionID      string            `json:"session_id,omitempty"`
	Enabled        *bool             `json:"enabled,omitempty"`
	Status         string            `json:"status,omitempty"`
	LastRun        *FlexTime         `json:"last_run,omitempty"`
	NextRun        *FlexTime         `json:"next_run,omitempty"`
	FailureCount   int               `json:"failure_count,omitempty"`
	History        []ScheduledJobRun `json:"history,omitempty"`
	IsRecurring    *bool             `json:"is_recurring,omitempty"`
	CronExpression string            `json:"cron_expression,omitempty"`
	ScheduledFor   *FlexTime         `json:"scheduled_for,omitempty"`
	OriginalPrompt string            `json:"original_prompt,omitempty"`
	LastRunAt      *FlexTime         `json:"last_run_at,omitempty"`
	LastRunStatus  string            `json:"last_run_status,omitempty"`
	TotalRuns      int               `json:"total_runs,omitempty"`
	TotalFailures  int               `json:"total_failures,omitempty"`
}

type schedulerJobFields struct {
	ID           string
	Type         string
	Schedule     string
	RunAt        time.Time
	Message      string
	UserID       string
	SessionID    string
	Enabled      bool
	Status       string
	LastRun      time.Time
	NextRun      time.Time
	FailureCount int
	History      []ScheduledJobRun
}

func (j *ScheduledJob) UnmarshalJSON(data []byte) error {
	fields, err := parseSchedulerJobFields(data)
	if err != nil {
		return err
	}

	*j = ScheduledJob{
		ID:           fields.ID,
		Type:         fields.Type,
		Schedule:     fields.Schedule,
		RunAt:        fields.RunAt,
		Message:      fields.Message,
		UserID:       fields.UserID,
		SessionID:    fields.SessionID,
		Enabled:      fields.Enabled,
		Status:       fields.Status,
		LastRun:      fields.LastRun,
		NextRun:      fields.NextRun,
		FailureCount: fields.FailureCount,
		History:      fields.History,
	}

	return nil
}

func (j *SchedulerJobDetail) UnmarshalJSON(data []byte) error {
	fields, err := parseSchedulerJobFields(data)
	if err != nil {
		return err
	}

	*j = SchedulerJobDetail{
		ID:           fields.ID,
		Type:         fields.Type,
		Schedule:     fields.Schedule,
		RunAt:        fields.RunAt,
		Message:      fields.Message,
		UserID:       fields.UserID,
		Enabled:      fields.Enabled,
		Status:       fields.Status,
		NextRun:      fields.NextRun,
		LastRun:      fields.LastRun,
		FailureCount: fields.FailureCount,
	}

	return nil
}

func parseSchedulerJobFields(data []byte) (schedulerJobFields, error) {
	var wire schedulerJobWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return schedulerJobFields{}, err
	}

	jobType := normalizeSchedulerJobType(wire.Type, wire.IsRecurring)
	runAt := flexTimeValue(wire.RunAt)
	if runAt.IsZero() {
		runAt = flexTimeValue(wire.ScheduledFor)
	}

	schedule := strings.TrimSpace(wire.Schedule)
	if schedule == "" {
		switch {
		case wire.CronExpression != "":
			schedule = wire.CronExpression
		case !runAt.IsZero():
			schedule = runAt.Format(time.RFC3339Nano)
		}
	}

	if runAt.IsZero() && jobType == "one_time" && schedule != "" {
		if parsed, err := parseAPITimeString(schedule); err == nil {
			runAt = parsed
		}
	}

	lastRun := flexTimeValue(wire.LastRun)
	if lastRun.IsZero() {
		lastRun = flexTimeValue(wire.LastRunAt)
	}

	nextRun := flexTimeValue(wire.NextRun)
	if nextRun.IsZero() {
		nextRun = computeSchedulerNextRun(jobType, schedule, runAt)
	}

	message := wire.Message
	if message == "" {
		message = wire.OriginalPrompt
	}

	failureCount := wire.FailureCount
	if failureCount == 0 {
		failureCount = wire.TotalFailures
	}

	status := normalizeSchedulerJobStatus(wire.Status, wire.Enabled)

	return schedulerJobFields{
		ID:           wire.ID,
		Type:         jobType,
		Schedule:     schedule,
		RunAt:        runAt,
		Message:      message,
		UserID:       wire.UserID,
		SessionID:    wire.SessionID,
		Enabled:      schedulerJobEnabled(status, wire.Enabled),
		Status:       status,
		LastRun:      lastRun,
		NextRun:      nextRun,
		FailureCount: failureCount,
		History:      wire.History,
	}, nil
}

func flexTimeValue(value *FlexTime) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.Time
}

func normalizeSchedulerJobType(jobType string, isRecurring *bool) string {
	switch strings.ToLower(strings.TrimSpace(jobType)) {
	case "once", "one_time":
		return "one_time"
	case "recurring":
		return "recurring"
	}

	if isRecurring != nil {
		if *isRecurring {
			return "recurring"
		}
		return "one_time"
	}

	return strings.TrimSpace(jobType)
}

func normalizeSchedulerJobStatus(status string, enabled *bool) string {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "" {
		return status
	}

	if enabled != nil {
		if *enabled {
			return "ACTIVE"
		}
		return "DISABLED"
	}

	return ""
}

func schedulerJobEnabled(status string, enabled *bool) bool {
	if enabled != nil {
		return *enabled
	}
	return status == "ACTIVE"
}

func computeSchedulerNextRun(jobType, schedule string, runAt time.Time) time.Time {
	if jobType == "one_time" {
		return runAt
	}

	if jobType != "recurring" || schedule == "" {
		return time.Time{}
	}

	parsed, err := schedulerCronParser.Parse(schedule)
	if err != nil {
		return time.Time{}
	}

	return parsed.Next(time.Now())
}

// UserResolveRequest for POST /users/resolve
type UserResolveRequest struct {
	Identifier string `json:"identifier"`
	CreateUser bool   `json:"create_user,omitempty"`
}

// UserResolveResponse from POST /users/resolve
type UserResolveResponse struct {
	Identifier     string `json:"identifier"`
	MuxiUserID     string `json:"muxi_user_id"`
	InternalUserID int    `json:"internal_user_id"`
}

// FormationInfoResponse from GET /formation
type FormationInfoResponse struct {
	FormationID string `json:"formation_id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
}

// OverlordSoulResponse from GET /overlord/soul
type OverlordSoulResponse struct {
	Soul string `json:"soul"`
}

// SecretResponse from GET /secrets/{key}
type SecretResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"` // masked
}

// SessionDetailResponse from GET /sessions/{session_id}
type SessionDetailResponse struct {
	SessionID    string                 `json:"session_id"`
	UserID       string                 `json:"user_id"`
	CreatedAt    *FlexTime              `json:"created_at,omitempty"`
	LastActivity *FlexTime              `json:"last_activity,omitempty"`
	MessageCount int                    `json:"message_count,omitempty"`
	Active       bool                   `json:"active,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LoggingDestinationsResponse from GET /logging/destinations
type LoggingDestinationsResponse struct {
	System       LoggingSystemConfig             `json:"system"`
	Conversation LoggingConversationDestinations `json:"conversation"`
}

// LoggingConversationDestinations represents conversation logging destinations
type LoggingConversationDestinations struct {
	Destinations []LoggingDestination `json:"destinations"`
	Count        int                  `json:"count"`
}

// BulkIdentifiersRequest for POST /users/identifiers
type BulkIdentifiersRequest struct {
	MuxiUserID  string        `json:"muxi_user_id,omitempty"`
	Identifiers []interface{} `json:"identifiers"` // Can be strings or objects
}

// BulkIdentifiersResponse from POST /users/identifiers
type BulkIdentifiersResponse struct {
	MuxiUserID            string   `json:"muxi_user_id"`
	InternalUserID        int      `json:"internal_user_id"`
	IdentifiersAssociated int      `json:"identifiers_associated"`
	NewIdentifiers        []string `json:"new_identifiers"`
}

// Credential represents a stored user credential
type Credential struct {
	CredentialID      string    `json:"credential_id"`
	Service           string    `json:"service"`
	Name              string    `json:"name"`
	CredentialPreview string    `json:"credential_preview"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

// CredentialsListResponse from GET /credentials
type CredentialsListResponse struct {
	Credentials []Credential `json:"credentials"`
	Count       int          `json:"count"`
}

// CredentialService represents an available credential service
type CredentialService struct {
	Service     string `json:"service"`
	ServerID    string `json:"server_id"`
	Description string `json:"description"`
}

// CredentialServicesResponse from GET /credentials/services
type CredentialServicesResponse struct {
	Services []CredentialService `json:"services"`
	Count    int                 `json:"count"`
}

// CreateCredentialRequest for POST /credentials
type CreateCredentialRequest struct {
	Service    string                 `json:"service"`
	Name       string                 `json:"name,omitempty"`
	Credential map[string]interface{} `json:"credential"`
}

// CreateCredentialResponse from POST /credentials
type CreateCredentialResponse struct {
	CredentialID      string    `json:"credential_id"`
	Service           string    `json:"service"`
	Name              string    `json:"name"`
	CredentialPreview string    `json:"credential_preview"`
	CreatedAt         time.Time `json:"created_at"`
}

// DeleteCredentialResponse from DELETE /credentials/{id}
type DeleteCredentialResponse struct {
	CredentialID string `json:"credential_id"`
	Deleted      bool   `json:"deleted"`
}
