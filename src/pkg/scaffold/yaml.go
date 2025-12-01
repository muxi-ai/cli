package scaffold

import (
	"bytes"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// marshalYAML marshals a yaml.Node with 2-space indentation
func marshalYAML(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// removeCommentedSection removes a commented-out config section from YAML content.
// It looks for patterns like "# key:" and removes that block including subsequent
// commented lines that are indented (part of the same section).
func removeCommentedSection(content string, key string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	// Pattern to match "# key:" with optional leading whitespace
	sectionPattern := regexp.MustCompile(`^#\s*` + regexp.QuoteMeta(key) + `:`)

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Check if this line starts a commented section we want to remove
		if sectionPattern.MatchString(line) {
			// Skip this line and all subsequent commented/indented lines
			i++
			for i < len(lines) {
				nextLine := lines[i]
				// Continue skipping if:
				// - Line is a comment starting with "#   " (indented comment, part of section)
				// - Line is empty
				// But stop if we hit a non-indented comment or non-comment line
				if strings.HasPrefix(nextLine, "#   ") || strings.HasPrefix(nextLine, "#\t") {
					i++
					continue
				}
				if nextLine == "" || nextLine == "#" {
					i++
					continue
				}
				// Stop at next section (comment or real)
				break
			}
			continue
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

// ensureBlankLineBeforeTopLevel ensures there's a blank line before top-level keys
func ensureBlankLineBeforeTopLevel(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	// Top-level keys that should have a blank line before them
	topLevelPattern := regexp.MustCompile(`^[a-z_]+:`)

	for i, line := range lines {
		// If this is a top-level key (not indented, ends with :)
		if topLevelPattern.MatchString(line) && i > 0 {
			// Check if previous line is not empty and not a comment
			prevLine := strings.TrimSpace(lines[i-1])
			if prevLine != "" && !strings.HasPrefix(prevLine, "#") {
				// Add blank line before this top-level key
				result = append(result, "")
			}
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// cleanupAdditionalConfigSection ensures the "Additional configuration" section
// is at the very end of the file, and removes any empty lines at the end.
func cleanupAdditionalConfigSection(content string) string {
	lines := strings.Split(content, "\n")

	// Find the "Additional configuration" header line
	headerPattern := regexp.MustCompile(`^#\s*─+\s*$`)
	additionalConfigPattern := regexp.MustCompile(`^#\s*Additional configuration`)

	var beforeAdditional []string
	var additionalSection []string
	var afterAdditional []string

	state := "before" // before, in_header, in_additional, after

	for _, line := range lines {
		switch state {
		case "before":
			if headerPattern.MatchString(line) {
				state = "in_header"
				additionalSection = append(additionalSection, line)
			} else {
				beforeAdditional = append(beforeAdditional, line)
			}

		case "in_header":
			additionalSection = append(additionalSection, line)
			if additionalConfigPattern.MatchString(line) {
				state = "in_additional"
			} else if !headerPattern.MatchString(line) && !strings.HasPrefix(line, "#") {
				// Not the right header block, put it back
				beforeAdditional = append(beforeAdditional, additionalSection...)
				additionalSection = nil
				state = "before"
			}

		case "in_additional":
			// Check if this is a non-comment line (real config added after)
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				state = "after"
				afterAdditional = append(afterAdditional, line)
			} else {
				additionalSection = append(additionalSection, line)
			}

		case "after":
			afterAdditional = append(afterAdditional, line)
		}
	}

	// If there's content after the additional section, move it before
	if len(afterAdditional) > 0 {
		// Remove trailing empty lines from beforeAdditional
		for len(beforeAdditional) > 0 && strings.TrimSpace(beforeAdditional[len(beforeAdditional)-1]) == "" {
			beforeAdditional = beforeAdditional[:len(beforeAdditional)-1]
		}

		// Remove leading empty lines from afterAdditional
		for len(afterAdditional) > 0 && strings.TrimSpace(afterAdditional[0]) == "" {
			afterAdditional = afterAdditional[1:]
		}

		// Rebuild: before + after (moved real config) + additional section
		result := append(beforeAdditional, afterAdditional...)
		if len(additionalSection) > 0 {
			result = append(result, additionalSection...)
		}

		// Clean up trailing empty lines and ensure single newline at end
		for len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "" {
			result = result[:len(result)-1]
		}

		return strings.Join(result, "\n") + "\n"
	}

	// No reorganization needed, just clean up trailing newlines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n") + "\n"
}
