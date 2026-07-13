package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"gopkg.in/yaml.v3"
)

// The rules in this file mirror the runtime's fail-fast formation validation
// (runtime formation/config/validation.py). The runtime stays the source of
// truth for semantics; the CLI checks key sets, types, and enums so mistakes
// surface before a deploy instead of at formation load.

// validateToolFilters checks MCP tool filtering. 'allow'/'deny' are the
// canonical keys with 'whitelist'/'blacklist' as accepted aliases; declaring
// a list under both spellings, both an allow-list and a deny-list, or any
// unknown key is an error.
func validateToolFilters(tools map[string]interface{}, relPath string, result *Result) {
	known := map[string]bool{"allow": true, "deny": true, "whitelist": true, "blacklist": true}
	var unknown []string
	for key := range tools {
		if !known[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		result.Errors = append(result.Errors, Issue{
			File:    relPath,
			Field:   "tools",
			Message: fmt.Sprintf("unknown key(s) %v in 'tools' - only 'allow' and 'deny' ('whitelist'/'blacklist' aliases) are supported", unknown),
		})
	}

	for canonical, alias := range map[string]string{"allow": "whitelist", "deny": "blacklist"} {
		_, hasCanonical := tools[canonical]
		_, hasAlias := tools[alias]
		if hasCanonical && hasAlias {
			result.Errors = append(result.Errors, Issue{
				File:    relPath,
				Field:   "tools",
				Message: fmt.Sprintf("'tools.%s' and 'tools.%s' are the same list - declare only one spelling", canonical, alias),
			})
		}
	}

	_, hasAllow := tools["allow"]
	_, hasWhitelist := tools["whitelist"]
	_, hasDeny := tools["deny"]
	_, hasBlacklist := tools["blacklist"]
	if (hasAllow || hasWhitelist) && (hasDeny || hasBlacklist) {
		result.Errors = append(result.Errors, Issue{
			File:    relPath,
			Field:   "tools",
			Message: "'tools.allow' and 'tools.deny' are mutually exclusive - declare only one",
		})
	}
}

// validateTuning checks the top-level 'tuning' block: a bare boolean
// shorthand or a mapping with only the runtime's supported keys.
func validateTuning(formation map[string]interface{}, fileName string, result *Result) {
	raw, present := formation["tuning"]
	if !present {
		return
	}

	if _, ok := raw.(bool); ok {
		return
	}

	tuning, ok := raw.(map[string]interface{})
	if !ok {
		result.Errors = append(result.Errors, Issue{
			File:    fileName,
			Field:   "tuning",
			Message: "'tuning' must be a boolean or a mapping",
		})
		return
	}

	allowed := map[string]bool{"active": true, "interval_hours": true, "auto_apply": true}
	var unknown []string
	for key := range tuning {
		if !allowed[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		result.Errors = append(result.Errors, Issue{
			File:    fileName,
			Field:   "tuning",
			Message: fmt.Sprintf("unknown key(s) %v in 'tuning' - supported keys are [active auto_apply interval_hours]", unknown),
		})
	}

	for _, boolKey := range []string{"active", "auto_apply"} {
		if value, ok := tuning[boolKey]; ok {
			if _, isBool := value.(bool); !isBool {
				result.Errors = append(result.Errors, Issue{
					File:    fileName,
					Field:   "tuning." + boolKey,
					Message: fmt.Sprintf("'tuning.%s' must be a boolean", boolKey),
				})
			}
		}
	}

	if value, ok := tuning["interval_hours"]; ok {
		if !isPositiveNumber(value) {
			result.Errors = append(result.Errors, Issue{
				File:    fileName,
				Field:   "tuning.interval_hours",
				Message: "'tuning.interval_hours' must be a positive number",
			})
		}
	}
}

func isPositiveNumber(value interface{}) bool {
	switch n := value.(type) {
	case int:
		return n > 0
	case float64:
		return n > 0
	default:
		return false
	}
}

// A2A auth validation. Types and per-type required fields mirror the
// runtime's inbound/outbound checks; 'openid' is inbound-only (JWT/JWKS
// validation) and 'oauth2' is outbound-only (client_credentials).

var a2aAuthRequiredFields = map[string][]string{
	"api_key": {"key"},
	"bearer":  {"token"},
	"basic":   {"username", "password"},
	"custom":  {"headers"},
	"hmac":    {"secret"},
	"oauth2":  {"client_id", "client_secret", "token_url"},
	"openid":  {"issuer"},
}

var a2aOutboundAuthTypes = []string{"api_key", "bearer", "basic", "custom", "hmac", "oauth2", "none"}
var a2aInboundAuthTypes = []string{"api_key", "bearer", "basic", "custom", "hmac", "openid", "none"}

// validateA2AAuthBlock checks one 'auth' block against the valid types and
// required fields for its direction ("inbound" or "outbound").
func validateA2AAuthBlock(raw interface{}, direction, file, field string, result *Result) {
	auth, ok := raw.(map[string]interface{})
	if !ok {
		result.Errors = append(result.Errors, Issue{
			File:    file,
			Field:   field,
			Message: "auth must be a mapping",
		})
		return
	}

	authType := "none"
	if t, ok := auth["type"].(string); ok {
		authType = t
	}

	validTypes := a2aOutboundAuthTypes
	if direction == "inbound" {
		validTypes = a2aInboundAuthTypes
	}

	valid := false
	for _, t := range validTypes {
		if authType == t {
			valid = true
			break
		}
	}
	if !valid {
		hint := ""
		if direction == "outbound" && authType == "openid" {
			hint = " ('openid' is inbound-only; use 'oauth2' for outbound services)"
		}
		if direction == "inbound" && authType == "oauth2" {
			hint = " ('oauth2' is outbound-only; use 'openid' to validate inbound JWTs)"
		}
		result.Errors = append(result.Errors, Issue{
			File:    file,
			Field:   field,
			Message: fmt.Sprintf("auth type '%s' is invalid%s - valid types: %s", authType, hint, strings.Join(validTypes, ", ")),
		})
		return
	}

	for _, required := range a2aAuthRequiredFields[authType] {
		if _, ok := auth[required]; !ok {
			result.Errors = append(result.Errors, Issue{
				File:    file,
				Field:   field,
				Message: fmt.Sprintf("%s auth requires '%s' field", authType, required),
			})
		}
	}

	if authType == "openid" {
		if issuer, ok := auth["issuer"].(string); ok {
			if !strings.HasPrefix(issuer, "http://") && !strings.HasPrefix(issuer, "https://") {
				result.Errors = append(result.Errors, Issue{
					File:    file,
					Field:   field + ".issuer",
					Message: "openid auth 'issuer' must be a URL",
				})
			}
		}
	}
}

// validateA2A checks the formation-level a2a.inbound.auth block.
func validateA2A(formation map[string]interface{}, fileName string, result *Result) {
	a2a, ok := formation["a2a"].(map[string]interface{})
	if !ok {
		return
	}
	inbound, ok := a2a["inbound"].(map[string]interface{})
	if !ok {
		return
	}
	if auth, ok := inbound["auth"]; ok {
		validateA2AAuthBlock(auth, "inbound", fileName, "a2a.inbound.auth", result)
	}
}

// validateA2AServiceFiles checks the auth blocks of outbound A2A service
// files in the a2a/ directory.
func validateA2AServiceFiles(rootDir string, result *Result) {
	a2aDir := filepath.Join(rootDir, "a2a")
	if info, err := os.Stat(a2aDir); err != nil || !info.IsDir() {
		return
	}

	entries, err := os.ReadDir(a2aDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !context.HasConfigExtension(entry.Name()) {
			continue
		}

		relPath := filepath.Join("a2a", entry.Name())
		data, err := os.ReadFile(filepath.Join(a2aDir, entry.Name()))
		if err != nil {
			result.Errors = append(result.Errors, Issue{
				File:    relPath,
				Message: fmt.Sprintf("failed to read: %v", err),
			})
			continue
		}

		var service map[string]interface{}
		if err := yaml.Unmarshal(data, &service); err != nil {
			for _, e := range splitYAMLErrors(err) {
				result.Errors = append(result.Errors, Issue{
					File:    relPath,
					Message: e,
				})
			}
			continue
		}

		if auth, ok := service["auth"]; ok {
			validateA2AAuthBlock(auth, "outbound", relPath, "auth", result)
		}
	}
}
