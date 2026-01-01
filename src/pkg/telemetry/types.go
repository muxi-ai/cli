package telemetry

import "time"

// Event represents the telemetry payload sent to the server
type Event struct {
	Module        string  `json:"module"`
	MachineID     string  `json:"machine_id"`
	Timestamp     string  `json:"ts"`
	Country       string  `json:"country"`
	SchemaVersion int     `json:"schema_version"`
	Payload       Payload `json:"payload"`
}

// Payload contains the CLI-specific telemetry data
type Payload struct {
	System         SystemInfo         `json:"system"`
	Registry       RegistryStats      `json:"registry"`
	Formations     FormationStats     `json:"formations"`
	Scaffolding    ScaffoldStats      `json:"scaffolding"`
	Usage          UsageStats         `json:"usage"`
	Infrastructure InfrastructureInfo `json:"infrastructure"`
	Help           map[string]int     `json:"help"`
}

// SystemInfo contains CLI version and platform info
type SystemInfo struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// RegistryStats tracks registry operations
type RegistryStats struct {
	Pulls  int `json:"pulls"`
	Pushes int `json:"pushes"`
}

// FormationStats tracks formation lifecycle
type FormationStats struct {
	Created  int `json:"created"`
	Deployed int `json:"deployed"`
}

// ScaffoldStats tracks scaffolding usage
type ScaffoldStats struct {
	Agents   int `json:"agents"`
	MCPs     int `json:"mcps"`
	SOPs     int `json:"sops"`
	Triggers int `json:"triggers"`
}

// UsageStats tracks interactive usage
type UsageStats struct {
	ChatSessions int `json:"chat_sessions"`
	LogsViewed   int `json:"logs_viewed"`
}

// InfrastructureInfo tracks configured resources (read at flush time)
type InfrastructureInfo struct {
	ProfilesConfigured   int `json:"profiles_configured"`
	FormationsConfigured int `json:"formations_configured"`
	RegistriesConfigured int `json:"registries_configured"`
}

// LocalState represents the persistent state stored in telemetry.json
type LocalState struct {
	LastFlush   time.Time      `json:"last_flush"`
	Registry    RegistryStats  `json:"registry"`
	Formations  FormationStats `json:"formations"`
	Scaffolding ScaffoldStats  `json:"scaffolding"`
	Usage       UsageStats     `json:"usage"`
	Help        map[string]int `json:"help"`
}

// NewLocalState creates a new empty local state
func NewLocalState() *LocalState {
	return &LocalState{
		LastFlush: time.Now(),
		Help:      make(map[string]int),
	}
}
