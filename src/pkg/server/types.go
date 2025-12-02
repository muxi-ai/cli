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

// HealthResponse from GET /health
type HealthResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Server struct {
			Status     string `json:"status"`
			Version    string `json:"version"`
			Uptime     int64  `json:"uptime"`
			Formations int    `json:"formations"`
		} `json:"server"`
		Formations []struct {
			ID      string `json:"id"`
			Status  string `json:"status"`
			Healthy bool   `json:"healthy"`
		} `json:"formations"`
	} `json:"data"`
}

// ServerStatusResponse from GET /rpc/server/status
type ServerStatusResponse struct {
	Server struct {
		ID       string `json:"id"`
		Status   string `json:"status"`
		Version  string `json:"version"`
		Uptime   int64  `json:"uptime"`
		Platform string `json:"platform"`
	} `json:"server"`
	Formations struct {
		Total   int `json:"total"`
		Running int `json:"running"`
		Stopped int `json:"stopped"`
	} `json:"formations"`
	Ports struct {
		Start     int `json:"start"`
		End       int `json:"end"`
		Available int `json:"available"`
	} `json:"ports"`
	Runtime struct {
		Type    string `json:"type"`
		Version string `json:"version"`
	} `json:"runtime"`
}

// ListFormationsResponse from GET /rpc/formations
type ListFormationsResponse struct {
	Formations []FormationListItem `json:"formations"`
	Total      int                 `json:"total"`
}

// FormationListItem is a formation in the list
type FormationListItem struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Port    int    `json:"port"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptime"`
	Healthy bool   `json:"healthy"`
}

// FormationDetail from GET /rpc/formations/{id}
type FormationDetail struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Port            int    `json:"port"`
	CurrentVersion  string `json:"current_version"`
	PreviousVersion string `json:"previous_version"`
	Uptime          int64  `json:"uptime"`
	PID             int    `json:"pid"`
	Healthy         bool   `json:"healthy"`
	RestartCount    int    `json:"restart_count"`
	DeployedAt      string `json:"deployed_at"`
	UpdatedAt       string `json:"updated_at"`
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
