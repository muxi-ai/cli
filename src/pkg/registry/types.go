package registry

import "time"

// Formation represents a formation in the registry
type Formation struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Stars       int       `json:"stars"`
	Downloads   int       `json:"downloads"`
	Size        int64     `json:"size"`
	Components  Components `json:"components"`
	RegistryURL string    `json:"registry_url"`
	GitHubURL   string    `json:"github_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Components describes formation components
type Components struct {
	Agents   int `json:"agents"`
	MCPs     int `json:"mcps"`
	SOPs     int `json:"sops"`
	Triggers int `json:"triggers"`
	A2A      int `json:"a2a"`
}

// Version represents a formation version
type Version struct {
	Version   string    `json:"version"`
	Size      int64     `json:"size"`
	Downloads int       `json:"downloads"`
	CreatedAt time.Time `json:"created_at"`
}

// SearchResult represents search response
type SearchResult struct {
	Results []SearchItem `json:"results"`
	Total   int          `json:"total"`
	Query   string       `json:"query"`
}

// SearchItem represents a formation in search results
type SearchItem struct {
	Name        string `json:"name"`
	User        string `json:"user"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Downloads   int    `json:"downloads"`
	Stars       int    `json:"stars"`
	URL         string `json:"url"`
}

// PublishResult represents publish response
type PublishResult struct {
	Formation   string `json:"formation"`
	Version     string `json:"version"`
	RegistryURL string `json:"registry_url"`
	GitHubURL   string `json:"github_url"`
}

// PullInfo contains download information
type PullInfo struct {
	Formation   string `json:"formation"`
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	Size        int64  `json:"size"`
}

// UserInfo represents authenticated user
type UserInfo struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// RegistryEntry stores credentials for a registry
type RegistryEntry struct {
	Token     string    `yaml:"token"`
	Username  string    `yaml:"username"`
	CreatedAt time.Time `yaml:"created_at"`
}

// RegistriesConfig is the structure of ~/.muxi/cli/registries.yaml
type RegistriesConfig struct {
	Version         string                   `yaml:"version"`
	DefaultRegistry string                   `yaml:"default_registry"`
	Registries      map[string]RegistryEntry `yaml:"registries"`
}

// FormationRef represents a parsed formation reference (@user/name:version)
type FormationRef struct {
	Owner   string
	Name    string
	Version string
	Full    string
}

// DefaultRegistryURL is the default registry
const DefaultRegistryURL = "registry.muxi.org"
