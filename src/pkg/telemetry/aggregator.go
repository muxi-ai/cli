package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	flushInterval = 1 * time.Hour
)

// statePath returns the path to ~/.muxi/cli/telemetry.json
func statePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muxi", "cli", "telemetry.json")
}

// Load loads the local state from disk
func Load() *LocalState {
	path := statePath()

	data, err := os.ReadFile(path)
	if err != nil {
		return NewLocalState()
	}

	var state LocalState
	if err := json.Unmarshal(data, &state); err != nil {
		return NewLocalState()
	}

	if state.Help == nil {
		state.Help = make(map[string]int)
	}

	return &state
}

// Save persists the local state to disk
func (s *LocalState) Save() error {
	path := statePath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Reset clears all counters and updates last_flush
func (s *LocalState) Reset() {
	s.LastFlush = time.Now()
	s.Registry = RegistryStats{}
	s.Formations = FormationStats{}
	s.Scaffolding = ScaffoldStats{}
	s.Usage = UsageStats{}
	s.Help = make(map[string]int)
}

// FlushIfDue checks if 1h has passed and flushes if needed
func (s *LocalState) FlushIfDue() {
	if time.Since(s.LastFlush) < flushInterval {
		return
	}

	// Only send if telemetry is enabled
	if IsEnabled() {
		Send(s.buildEvent())
	}

	// Always reset after flush interval
	s.Reset()
}

// buildEvent creates the telemetry event from current state
func (s *LocalState) buildEvent() Event {
	return Event{
		Module:        "cli",
		MachineID:     GetMachineID(),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Country:       GetCountry(),
		SchemaVersion: 1,
		Payload: Payload{
			System:         getSystemInfo(),
			Registry:       s.Registry,
			Formations:     s.Formations,
			Scaffolding:    s.Scaffolding,
			Usage:          s.Usage,
			Infrastructure: getInfrastructureInfo(),
			Help:           s.Help,
		},
	}
}

// Increment methods

func (s *LocalState) IncrementPull() {
	s.Registry.Pulls++
}

func (s *LocalState) IncrementPush() {
	s.Registry.Pushes++
}

func (s *LocalState) IncrementFormationCreated() {
	s.Formations.Created++
}

func (s *LocalState) IncrementDeploy() {
	s.Formations.Deployed++
}

func (s *LocalState) IncrementScaffold(kind string) {
	switch kind {
	case "agent":
		s.Scaffolding.Agents++
	case "mcp":
		s.Scaffolding.MCPs++
	case "sop":
		s.Scaffolding.SOPs++
	case "trigger":
		s.Scaffolding.Triggers++
	}
}

func (s *LocalState) IncrementChat() {
	s.Usage.ChatSessions++
}

func (s *LocalState) IncrementLogs() {
	s.Usage.LogsViewed++
}

func (s *LocalState) IncrementHelp(cmd string) {
	if s.Help == nil {
		s.Help = make(map[string]int)
	}
	s.Help[cmd]++
}
