package server

import "encoding/json"

// APIResponse is the common response wrapper
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Message string          `json:"message,omitempty"`
	Code    int             `json:"code,omitempty"`
}

// HealthResponse from GET /health (actual server response)
type HealthResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Status     string `json:"status"` // "ok" when healthy
		Formations int    `json:"formations"`
		PortPool   struct {
			Allocated int `json:"allocated"`
			Available int `json:"available"`
			Total     int `json:"total"`
		} `json:"port_pool"`
	} `json:"data"`
}

// ServerStatusResponse from GET /rpc/server/status
type ServerStatusResponse struct {
	Server struct {
		ServerID string `json:"server_id"`
		Version  string `json:"version"`
		Uptime   int64  `json:"uptime"`
		Port     int    `json:"port"`
	} `json:"server"`
	Formations struct {
		Total   int `json:"total"`
		Running int `json:"running"`
		Stopped int `json:"stopped"`
		Healthy int `json:"healthy"`
	} `json:"formations"`
	Ports struct {
		Allocated int    `json:"allocated"`
		Available int    `json:"available"`
		Range     string `json:"range"`
	} `json:"ports"`
	Runtime struct {
		Type     string   `json:"type"`
		Platform string   `json:"platform"`
		Versions []string `json:"versions"`
	} `json:"runtime"`
}

// ListFormationsResponse from GET /rpc/formations
type ListFormationsResponse struct {
	Formations []FormationListItem `json:"formations"`
	Total      int                 `json:"total"`
}

// FormationListItem is a formation in the list
type FormationListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Port         int    `json:"port"`
	PID          int    `json:"pid"`
	Uptime       int64  `json:"uptime"` // seconds
	RestartCount int    `json:"restart_count"`
	Healthy      bool   `json:"healthy"`
}

// FormationDetail from GET /rpc/formations/{id}
type FormationDetail struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Port         int    `json:"port"`
	PID          int    `json:"pid"`
	Uptime       int64  `json:"uptime"` // seconds
	Healthy      bool   `json:"healthy"`
	RestartCount int    `json:"restart_count"`
	CreatedAt    string `json:"created_at"`
	DeployedAt   string `json:"deployed_at"`
	UpdatedAt    string `json:"updated_at"`
}

// DeployResponse from POST /rpc/formations
type DeployResponse struct {
	ID      string `json:"id"`
	Port    int    `json:"port"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// RollbackResponse from POST /rpc/formations/{id}/rollback
type RollbackResponse struct {
	ID              string `json:"id"`
	PreviousVersion string `json:"previous_version"`
	CurrentVersion  string `json:"current_version"`
}

// LogsResponse from GET /rpc/formations/{id}/logs
type LogsResponse struct {
	FormationID string   `json:"formation_id"`
	Lines       []string `json:"lines"`
	Stream      string   `json:"stream"`
}

// SSE Event Types for Deploy/Update streaming

// DeployProgressEvent is a progress update during deployment
type DeployProgressEvent struct {
	Stage       string `json:"stage"`        // extracting, validating, resolving_runtime, downloading_sif, pulling_runner, spawning, health_check
	Message     string `json:"message"`      // Human-readable message
	Progress    int    `json:"progress"`     // 0-100 for downloading_sif
	URL         string `json:"url"`          // Download URL for downloading_sif
	Version     string `json:"version"`      // Resolved version for resolving_runtime
	Attempt     int    `json:"attempt"`      // Current attempt for health_check
	MaxAttempts int    `json:"max_attempts"` // Max attempts for health_check
	StagingPort int    `json:"staging_port"` // For spawning_staging (update)
}

// DeployCompleteEvent is sent when deployment succeeds
type DeployCompleteEvent struct {
	FormationID     string `json:"formation_id"`
	Port            int    `json:"port"`
	Status          string `json:"status"`
	URL             string `json:"url"`
	HealthURL       string `json:"health_url"`
	PID             int    `json:"pid"`
	PreviousVersion string `json:"previous_version,omitempty"` // For updates
	NewVersion      string `json:"new_version,omitempty"`      // For updates
}

// DeployErrorEvent is sent when deployment fails
type DeployErrorEvent struct {
	Error          string `json:"error"`
	Message        string `json:"message"`
	Stage          string `json:"stage"`
	RollbackStatus string `json:"rollback_status,omitempty"` // For updates: not_needed, success, failed
}

// SSEEvent represents a parsed Server-Sent Event
type SSEEvent struct {
	Event string // "progress", "complete", "error"
	Data  string // JSON payload
}
