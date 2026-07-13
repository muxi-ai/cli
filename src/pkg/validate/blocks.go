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

// Knowledge source validation. URL schemes mirror the runtime's
// REMOTE_KNOWLEDGE_SUPPORTED_SCHEMES and agent_tree.regenerate mirrors its
// _AGENT_TREE_REGENERATE_MODES. Per-source 'retrieval' mode is deliberately
// NOT validated here until the retrieval-mode consolidation decision lands.

var remoteKnowledgeSchemes = map[string]bool{
	"http": true, "https": true, "s3": true, "gs": true, "az": true,
	"rsync": true, "rsync+ssh": true, "ftp": true, "sftp": true, "file": true,
}

var agentTreeRegenerateModes = []string{"manual", "on-source-change", "on-formation-load"}

// validateKnowledgeSources checks an agent's knowledge.sources entries:
// path XOR url, supported remote URL schemes, and the agent_tree block
// (regenerate enum; local sources only).
func validateKnowledgeSources(agent map[string]interface{}, relPath string, result *Result) {
	knowledge, ok := agent["knowledge"].(map[string]interface{})
	if !ok {
		return
	}
	sources, ok := knowledge["sources"].([]interface{})
	if !ok {
		return
	}

	for i, raw := range sources {
		field := fmt.Sprintf("knowledge.sources[%d]", i)
		source, ok := raw.(map[string]interface{})
		if !ok {
			result.Errors = append(result.Errors, Issue{
				File:    relPath,
				Field:   field,
				Message: "knowledge source must be a mapping",
			})
			continue
		}

		_, hasPath := source["path"]
		url, hasURL := source["url"].(string)
		if hasPath && hasURL {
			result.Errors = append(result.Errors, Issue{
				File:    relPath,
				Field:   field,
				Message: "knowledge source must declare either 'path' or 'url', not both",
			})
			continue
		}
		if !hasPath && !hasURL {
			result.Errors = append(result.Errors, Issue{
				File:    relPath,
				Field:   field,
				Message: "knowledge source missing required field: 'path' or 'url'",
			})
			continue
		}

		if hasURL {
			scheme := ""
			if idx := strings.Index(url, "://"); idx > 0 {
				scheme = strings.ToLower(url[:idx])
			}
			if !remoteKnowledgeSchemes[scheme] {
				supported := make([]string, 0, len(remoteKnowledgeSchemes))
				for s := range remoteKnowledgeSchemes {
					supported = append(supported, s)
				}
				sort.Strings(supported)
				result.Errors = append(result.Errors, Issue{
					File:    relPath,
					Field:   field + ".url",
					Message: fmt.Sprintf("unsupported URL scheme '%s' - supported schemes: %s", scheme, strings.Join(supported, ", ")),
				})
			}
			if _, hasTree := source["agent_tree"]; hasTree {
				result.Errors = append(result.Errors, Issue{
					File:    relPath,
					Field:   field + ".agent_tree",
					Message: "'agent_tree' is not supported on remote (url) sources - sync the source locally first",
				})
			}
			continue
		}

		if rawTree, hasTree := source["agent_tree"]; hasTree {
			tree, ok := rawTree.(map[string]interface{})
			if !ok {
				result.Errors = append(result.Errors, Issue{
					File:    relPath,
					Field:   field + ".agent_tree",
					Message: "'agent_tree' must be a mapping",
				})
				continue
			}
			for key := range tree {
				if key != "regenerate" {
					result.Errors = append(result.Errors, Issue{
						File:    relPath,
						Field:   field + ".agent_tree",
						Message: fmt.Sprintf("agent_tree setting '%s' is not recognized (allowed: regenerate)", key),
					})
				}
			}
			if regenerate, ok := tree["regenerate"]; ok {
				mode, isString := regenerate.(string)
				valid := false
				for _, m := range agentTreeRegenerateModes {
					if isString && mode == m {
						valid = true
						break
					}
				}
				if !valid {
					result.Errors = append(result.Errors, Issue{
						File:    relPath,
						Field:   field + ".agent_tree.regenerate",
						Message: fmt.Sprintf("agent_tree 'regenerate' must be one of: %s", strings.Join(agentTreeRegenerateModes, ", ")),
					})
				}
			}
		}
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
