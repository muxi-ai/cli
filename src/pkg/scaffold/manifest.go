package scaffold

import (
	"fmt"
	"os"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"gopkg.in/yaml.v3"
)

// AddComponentToFormation adds a component ID to the appropriate list in formation.yaml.
// section is one of: "agents", "mcp.servers", "a2a.outbound.services"
func AddComponentToFormation(rootDir, section, id string) error {
	formationPath, found := context.FindFormationFile(rootDir)
	if !found {
		return fmt.Errorf("formation file not found")
	}

	data, err := os.ReadFile(formationPath)
	if err != nil {
		return fmt.Errorf("failed to read formation file: %w", err)
	}

	content := string(data)

	switch section {
	case "agents":
		content = addToTopLevelList(content, "agents", id)
	case "mcp.servers":
		content = addToNestedList(content, "mcp", "servers", id)
	case "a2a.outbound.services":
		content = addToDeepNestedList(content, "a2a", "outbound", "services", id)
	default:
		return fmt.Errorf("unknown section: %s", section)
	}

	if err := os.WriteFile(formationPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write formation file: %w", err)
	}

	return nil
}

// addToTopLevelList adds an ID to a top-level list (e.g., agents:)
func addToTopLevelList(content, key, id string) string {
	lines := strings.Split(content, "\n")

	// Check if key already exists
	keyLine := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == key+":" || strings.HasPrefix(trimmed, key+":") {
			keyLine = i
			break
		}
	}

	entry := "  - " + id

	if keyLine == -1 {
		// Key doesn't exist -- append at end
		content = strings.TrimRight(content, "\n") + "\n\n" + key + ":\n" + entry + "\n"
		return content
	}

	// Key exists - check if ID already present in the list
	if isIDInList(lines, keyLine, id) {
		return content
	}

	// Find end of list to append
	insertAt := findListEnd(lines, keyLine)
	lines = insertLine(lines, insertAt, entry)
	return strings.Join(lines, "\n")
}

// addToNestedList adds an ID to a nested list (e.g., mcp.servers:)
func addToNestedList(content, parentKey, childKey, id string) string {
	lines := strings.Split(content, "\n")

	// Find parent key
	parentLine := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == parentKey+":" || strings.HasPrefix(trimmed, parentKey+":") {
			parentLine = i
			break
		}
	}

	if parentLine == -1 {
		// Parent doesn't exist -- append section
		content = strings.TrimRight(content, "\n") + "\n\n" + parentKey + ":\n  " + childKey + ":\n    - " + id + "\n"
		return content
	}

	// Find child key within parent
	childLine := -1
	for i := parentLine + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Check if we left parent section (non-indented non-empty line)
		if !strings.HasPrefix(lines[i], " ") && !strings.HasPrefix(lines[i], "\t") && trimmed != "" {
			break
		}
		if trimmed == childKey+":" || strings.HasPrefix(trimmed, childKey+":") {
			childLine = i
			break
		}
	}

	if childLine == -1 {
		// Child doesn't exist -- add under parent
		insertAt := findSectionEnd(lines, parentLine)
		lines = insertLine(lines, insertAt, "  "+childKey+":\n    - "+id)
		return strings.Join(lines, "\n")
	}

	if isIDInNestedList(lines, childLine, id, 4) {
		return content
	}

	insertAt := findListEnd(lines, childLine)
	lines = insertLine(lines, insertAt, "    - "+id)
	return strings.Join(lines, "\n")
}

// addToDeepNestedList adds an ID to a2a.outbound.services
func addToDeepNestedList(content, key1, key2, key3, id string) string {
	lines := strings.Split(content, "\n")

	// Find key1 (a2a:)
	k1Line := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == key1+":" || strings.HasPrefix(trimmed, key1+":") {
			k1Line = i
			break
		}
	}

	if k1Line == -1 {
		content = strings.TrimRight(content, "\n") + "\n\n" + key1 + ":\n  " + key2 + ":\n    " + key3 + ":\n      - " + id + "\n"
		return content
	}

	// Find key2 (outbound:) within key1
	k2Line := -1
	for i := k1Line + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(lines[i], " ") && !strings.HasPrefix(lines[i], "\t") && trimmed != "" {
			break
		}
		if trimmed == key2+":" || strings.HasPrefix(trimmed, key2+":") {
			k2Line = i
			break
		}
	}

	if k2Line == -1 {
		insertAt := findSectionEnd(lines, k1Line)
		lines = insertLine(lines, insertAt, "  "+key2+":\n    "+key3+":\n      - "+id)
		return strings.Join(lines, "\n")
	}

	// Find key3 (services:) within key2
	k3Line := -1
	for i := k2Line + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := countIndent(lines[i])
		if indent <= countIndent(lines[k2Line]) && trimmed != "" {
			break
		}
		if trimmed == key3+":" || strings.HasPrefix(trimmed, key3+":") {
			k3Line = i
			break
		}
	}

	if k3Line == -1 {
		insertAt := findSectionEnd(lines, k2Line)
		lines = insertLine(lines, insertAt, "    "+key3+":\n      - "+id)
		return strings.Join(lines, "\n")
	}

	if isIDInNestedList(lines, k3Line, id, 6) {
		return content
	}

	insertAt := findListEnd(lines, k3Line)
	lines = insertLine(lines, insertAt, "      - "+id)
	return strings.Join(lines, "\n")
}

// isIDInList checks if an ID already exists in the list starting at keyLine
func isIDInList(lines []string, keyLine int, id string) bool {
	for i := keyLine + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(lines[i], " ") && !strings.HasPrefix(lines[i], "\t") {
			break
		}
		if trimmed == "- "+id || trimmed == "- \""+id+"\"" {
			return true
		}
		// Stop if we hit a non-list-item
		if !strings.HasPrefix(trimmed, "- ") {
			break
		}
	}
	return false
}

// isIDInNestedList checks if ID is in a list at a specific indent level
func isIDInNestedList(lines []string, keyLine int, id string, minIndent int) bool {
	for i := keyLine + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := countIndent(lines[i])
		if indent < minIndent && trimmed != "" {
			break
		}
		if trimmed == "- "+id || trimmed == "- \""+id+"\"" {
			return true
		}
	}
	return false
}

// findListEnd finds the line after the last list item under keyLine
func findListEnd(lines []string, keyLine int) int {
	lastItem := keyLine + 1
	for i := keyLine + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(lines[i], " ") && !strings.HasPrefix(lines[i], "\t") {
			return i
		}
		if strings.HasPrefix(trimmed, "- ") {
			lastItem = i + 1
		} else {
			// Non-list content at same or lesser indent = end
			keyIndent := countIndent(lines[keyLine])
			curIndent := countIndent(lines[i])
			if curIndent <= keyIndent {
				return i
			}
		}
	}
	return lastItem
}

// findSectionEnd finds line after a section's content
func findSectionEnd(lines []string, sectionLine int) int {
	sectionIndent := countIndent(lines[sectionLine])
	for i := sectionLine + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if countIndent(lines[i]) <= sectionIndent && trimmed != "" {
			return i
		}
	}
	return len(lines)
}

func countIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

func insertLine(lines []string, at int, newLine string) []string {
	if at >= len(lines) {
		return append(lines, newLine)
	}
	lines = append(lines[:at+1], lines[at:]...)
	lines[at] = newLine
	return lines
}

// RemoveComponentFromFormation removes a component ID from formation.yaml (for future use)
func RemoveComponentFromFormation(rootDir, section, id string) error {
	formationPath, found := context.FindFormationFile(rootDir)
	if !found {
		return fmt.Errorf("formation file not found")
	}

	data, err := os.ReadFile(formationPath)
	if err != nil {
		return fmt.Errorf("failed to read formation file: %w", err)
	}

	// Use yaml round-trip to remove the entry
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to parse formation file: %w", err)
	}

	// Simple string replacement approach for removing list entries
	content := string(data)
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "- "+id || trimmed == "- \""+id+"\"" {
			continue
		}
		result = append(result, line)
	}

	return os.WriteFile(formationPath, []byte(strings.Join(result, "\n")), 0644)
}
