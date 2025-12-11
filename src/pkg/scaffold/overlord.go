package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
)

// ConfigureOverlord runs the overlord configuration wizard
func ConfigureOverlord() error {
	// Must be in formation directory
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation\n\nOr create a new one:\n  muxi new formation",
		)
		os.Exit(1)
	}

	// Show banner
	ui.Banner(`╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure Overlord                                  MUXI │
│──────────────────────────────────────────────────────────────│
│ The Overlord orchestrates agents, creates and routes tasks,  │
│ manages conversations, and handles clarifications.           │
╰──────────────────────────────────────────────────────────────╯`)

	// Step 1: What to configure
	options := []wizard.SelectOption{
		{Value: "persona", Label: "Persona (identity and communication style)"},
		{Value: "response", Label: "Response options (format, streaming, progress)"},
		{Value: "workflow", Label: "Workflow behavior (routing, decomposition, timeouts)"},
		{Value: "clarification", Label: "Clarification settings (question style, limits)"},
	}

	choice, err := wizard.PromptSelect("What would you like to configure?", options, 0)
	if err != nil {
		return err
	}

	switch choice {
	case "persona":
		return configureOverlordPersona(ctx.RootDir)
	case "response":
		return configureOverlordResponse(ctx.RootDir)
	case "workflow":
		return configureOverlordWorkflow(ctx.RootDir)
	case "clarification":
		return configureOverlordClarification(ctx.RootDir)
	}

	return nil
}

// configureOverlordPersona handles Flow 1: Persona
func configureOverlordPersona(rootDir string) error {
	fmt.Println()
	ui.Bold("Overlord Persona")
	fmt.Println()
	ui.Dimmed("  The persona defines how the Overlord communicates with users.")
	fmt.Println()

	// How to set persona
	options := []wizard.SelectOption{
		{Value: "editor", Label: "Enter text directly (opens $EDITOR)"},
		{Value: "file", Label: "Load from file"},
	}

	method, err := wizard.PromptSelect("  How would you like to set the persona?", options, 0)
	if err != nil {
		return err
	}

	var persona string

	if method == "editor" {
		// Open editor with empty content (like git commit)
		persona, err = editInEditor("")
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		if persona == "" {
			ui.PromptSkipped("  Persona (empty, no changes made)")
			return nil
		}
		ui.PromptSuccess("  Persona", "updated via editor")
	} else {
		// Load from file
		ui.Dimmed("  Path to persona file (markdown or text)")
		filePath, err := wizard.PromptString("  Path", "", validateFileExists)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		persona = string(content)
		ui.PromptSuccess("  Persona", fmt.Sprintf("loaded from %s", filePath))
	}

	// Update formation.yaml
	return updateOverlordPersonaInFormation(rootDir, persona)
}

// configureOverlordResponse handles Flow 2: Response Options
func configureOverlordResponse(rootDir string) error {
	fmt.Println()
	ui.Bold("Response Options")
	fmt.Println()

	// Get current values
	currentFormat := getCurrentOverlordValue(rootDir, "response", "format")
	currentStreaming := getCurrentOverlordValue(rootDir, "response", "streaming")
	currentProgress := getCurrentOverlordValue(rootDir, "response", "progress")

	// Format selection
	ui.Dimmed("  Default output format for responses")
	formatOptions := []wizard.SelectOption{
		{Value: "markdown", Label: addCurrentIndicator("Markdown (rich formatting, default)", currentFormat == "markdown" || currentFormat == "")},
		{Value: "text", Label: addCurrentIndicator("Plain text", currentFormat == "text")},
		{Value: "html", Label: addCurrentIndicator("HTML (web rendering)", currentFormat == "html")},
	}

	defaultFormatIdx := 0
	if currentFormat == "text" {
		defaultFormatIdx = 1
	} else if currentFormat == "html" {
		defaultFormatIdx = 2
	}

	format, err := wizard.PromptSelect("  Format", formatOptions, defaultFormatIdx)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Format", format)

	// Streaming
	ui.Dimmed("  Stream responses as they're generated")
	streamingDefault := currentStreaming != "false"
	streaming, err := wizard.PromptConfirm("  Enable streaming?", streamingDefault)
	if err != nil {
		return err
	}
	if streaming {
		ui.PromptSuccess("  Streaming", "enabled")
	} else {
		ui.PromptSkipped("  Streaming")
	}

	// Progress events
	ui.Dimmed("  Show status updates during long operations")
	progressDefault := currentProgress != "false"
	progress, err := wizard.PromptConfirm("  Enable progress events?", progressDefault)
	if err != nil {
		return err
	}
	if progress {
		ui.PromptSuccess("  Progress", "enabled")
	} else {
		ui.PromptSkipped("  Progress")
	}

	// Update formation.yaml
	return updateOverlordResponseInFormation(rootDir, format, streaming, progress)
}

// configureOverlordWorkflow handles Flow 3: Workflow Behavior
func configureOverlordWorkflow(rootDir string) error {
	fmt.Println()
	ui.Bold("Workflow Configuration")
	fmt.Println()

	// Get current values
	currentRouting := getCurrentOverlordValue(rootDir, "workflow", "routing_strategy")
	currentDecomp := getCurrentOverlordValue(rootDir, "workflow", "auto_decomposition")
	currentThreshold := getCurrentOverlordValue(rootDir, "workflow", "plan_approval_threshold")
	currentComplexity := getCurrentOverlordValue(rootDir, "workflow", "complexity_method")
	currentParallel := getCurrentOverlordValue(rootDir, "workflow", "parallel_execution")
	currentMaxParallel := getCurrentOverlordValue(rootDir, "workflow", "max_parallel_tasks")
	currentAffinity := getCurrentOverlordValue(rootDir, "workflow", "enable_agent_affinity")
	currentTaskTimeout := getCurrentOverlordValue(rootDir, "workflow.timeouts", "task")
	currentWorkflowTimeout := getCurrentOverlordValue(rootDir, "workflow.timeouts", "workflow")
	currentErrorRecovery := getCurrentOverlordValue(rootDir, "workflow", "error_recovery")

	// Routing strategy
	ui.Dimmed("  How to select agents for tasks")
	routingOptions := []wizard.SelectOption{
		{Value: "capability", Label: addCurrentIndicator("Capability-based (match task to agent skills)", currentRouting == "capability" || currentRouting == "")},
		{Value: "load_balanced", Label: addCurrentIndicator("Load-balanced (distribute work evenly)", currentRouting == "load_balanced")},
		{Value: "round_robin", Label: addCurrentIndicator("Round-robin (rotate through agents)", currentRouting == "round_robin")},
		{Value: "priority", Label: addCurrentIndicator("Priority-based (prefer higher-priority agents)", currentRouting == "priority")},
	}

	defaultRoutingIdx := getIndexForValue([]string{"capability", "load_balanced", "round_robin", "priority"}, currentRouting, 0)
	routing, err := wizard.PromptSelect("  Routing", routingOptions, defaultRoutingIdx)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Routing", routing)

	// Auto-decomposition
	ui.Dimmed("  Automatically break complex requests into subtasks")
	decompDefault := currentDecomp != "false"
	autoDecomp, err := wizard.PromptConfirm("  Enable auto-decomposition?", decompDefault)
	if err != nil {
		return err
	}
	if autoDecomp {
		ui.PromptSuccess("  Auto-decomposition", "enabled")
	} else {
		ui.PromptSkipped("  Auto-decomposition")
	}

	// Plan approval threshold
	ui.Dimmed("  Complexity score (1-10) above which user approval is required")
	thresholdDefault := "7"
	if currentThreshold != "" {
		thresholdDefault = currentThreshold
	}
	threshold, err := wizard.PromptString("  Plan approval threshold", thresholdDefault, validateThreshold)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Plan approval threshold", threshold)

	// Complexity method
	ui.Dimmed("  How to calculate task complexity")
	complexityOptions := []wizard.SelectOption{
		{Value: "hybrid", Label: addCurrentIndicator("Hybrid (balanced approach)", currentComplexity == "hybrid" || currentComplexity == "")},
		{Value: "heuristic", Label: addCurrentIndicator("Heuristic (fast, rule-based)", currentComplexity == "heuristic")},
		{Value: "llm", Label: addCurrentIndicator("LLM (accurate, uses LLM)", currentComplexity == "llm")},
	}

	defaultComplexityIdx := getIndexForValue([]string{"hybrid", "heuristic", "llm"}, currentComplexity, 0)
	complexity, err := wizard.PromptSelect("  Complexity method", complexityOptions, defaultComplexityIdx)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Complexity method", complexity)

	// Parallel execution
	ui.Dimmed("  Run independent tasks simultaneously")
	parallelDefault := currentParallel != "false"
	parallel, err := wizard.PromptConfirm("  Enable parallel execution?", parallelDefault)
	if err != nil {
		return err
	}
	if parallel {
		ui.PromptSuccess("  Parallel execution", "enabled")
	} else {
		ui.PromptSkipped("  Parallel execution")
	}

	// Max parallel tasks
	var maxParallel string
	if parallel {
		ui.Dimmed("  Maximum tasks to run at once (1-20)")
		maxParallelDefault := "5"
		if currentMaxParallel != "" {
			maxParallelDefault = currentMaxParallel
		}
		maxParallel, err = wizard.PromptString("  Max parallel tasks", maxParallelDefault, validateMaxParallel)
		if err != nil {
			return err
		}
		ui.PromptSuccess("  Max parallel tasks", maxParallel)
	}

	// Agent affinity
	ui.Dimmed("  Prefer agents that succeeded on similar tasks before")
	affinityDefault := currentAffinity != "false"
	affinity, err := wizard.PromptConfirm("  Enable agent affinity?", affinityDefault)
	if err != nil {
		return err
	}
	if affinity {
		ui.PromptSuccess("  Agent affinity", "enabled")
	} else {
		ui.PromptSkipped("  Agent affinity")
	}

	// Timeouts & Error Handling
	fmt.Println()
	ui.Bold("Timeouts & Error Handling")
	fmt.Println()

	// Task timeout
	ui.Dimmed("  Maximum time for a single task (seconds)")
	taskTimeoutDefault := "300"
	if currentTaskTimeout != "" {
		taskTimeoutDefault = currentTaskTimeout
	}
	taskTimeout, err := wizard.PromptString("  Task timeout", taskTimeoutDefault, validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Task timeout", taskTimeout+"s")

	// Workflow timeout
	ui.Dimmed("  Maximum time for an entire workflow (seconds)")
	workflowTimeoutDefault := "3600"
	if currentWorkflowTimeout != "" {
		workflowTimeoutDefault = currentWorkflowTimeout
	}
	workflowTimeout, err := wizard.PromptString("  Workflow timeout", workflowTimeoutDefault, validatePositiveInt)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Workflow timeout", workflowTimeout+"s")

	// Error recovery
	ui.Dimmed("  What to do when a task fails")
	errorOptions := []wizard.SelectOption{
		{Value: "retry_with_backoff", Label: addCurrentIndicator("Retry with backoff (retry with increasing delays)", currentErrorRecovery == "retry_with_backoff" || currentErrorRecovery == "")},
		{Value: "retry_with_alternate", Label: addCurrentIndicator("Retry with alternate (try a different agent)", currentErrorRecovery == "retry_with_alternate")},
		{Value: "fail_fast", Label: addCurrentIndicator("Fail fast (stop immediately)", currentErrorRecovery == "fail_fast")},
		{Value: "skip_and_continue", Label: addCurrentIndicator("Skip and continue (return partial results)", currentErrorRecovery == "skip_and_continue")},
	}

	defaultErrorIdx := getIndexForValue([]string{"retry_with_backoff", "retry_with_alternate", "fail_fast", "skip_and_continue"}, currentErrorRecovery, 0)
	errorRecovery, err := wizard.PromptSelect("  Error recovery", errorOptions, defaultErrorIdx)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Error recovery", errorRecovery)

	// Update formation.yaml
	return updateOverlordWorkflowInFormation(rootDir, routing, autoDecomp, threshold, complexity, parallel, maxParallel, affinity, taskTimeout, workflowTimeout, errorRecovery)
}

// configureOverlordClarification handles Flow 4: Clarification Settings
func configureOverlordClarification(rootDir string) error {
	fmt.Println()
	ui.Bold("Clarification Settings")
	fmt.Println()
	ui.Dimmed("  The Overlord asks clarifying questions for ambiguous requests.")
	fmt.Println()

	// Get current values
	currentStyle := getCurrentOverlordValue(rootDir, "clarification", "style")
	currentDirect := getCurrentOverlordValue(rootDir, "clarification.max_rounds", "direct")
	currentBrainstorm := getCurrentOverlordValue(rootDir, "clarification.max_rounds", "brainstorm")
	currentPlanning := getCurrentOverlordValue(rootDir, "clarification.max_rounds", "planning")
	currentExecution := getCurrentOverlordValue(rootDir, "clarification.max_rounds", "execution")

	// Style
	ui.Dimmed("  Communication style for clarifying questions")
	styleOptions := []wizard.SelectOption{
		{Value: "conversational", Label: addCurrentIndicator("Conversational (friendly, natural)", currentStyle == "conversational" || currentStyle == "")},
		{Value: "formal", Label: addCurrentIndicator("Formal (professional, structured)", currentStyle == "formal")},
		{Value: "brief", Label: addCurrentIndicator("Brief (minimal, direct)", currentStyle == "brief")},
	}

	defaultStyleIdx := getIndexForValue([]string{"conversational", "formal", "brief"}, currentStyle, 0)
	style, err := wizard.PromptSelect("  Style", styleOptions, defaultStyleIdx)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Style", style)

	// Max rounds
	fmt.Println()
	ui.Bold("Clarification Limits")
	fmt.Println()
	ui.Dimmed("  Maximum questions per conversation cycle:")
	fmt.Println()

	// Direct
	ui.Dimmed("  Direct mode (quick disambiguation)")
	directDefault := "3"
	if currentDirect != "" {
		directDefault = currentDirect
	}
	direct, err := wizard.PromptString("  Max rounds", directDefault, validateMaxRounds)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Direct max rounds", direct)

	// Brainstorm
	ui.Dimmed("  Brainstorm mode (creative exploration)")
	brainstormDefault := "10"
	if currentBrainstorm != "" {
		brainstormDefault = currentBrainstorm
	}
	brainstorm, err := wizard.PromptString("  Max rounds", brainstormDefault, validateMaxRounds)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Brainstorm max rounds", brainstorm)

	// Planning
	ui.Dimmed("  Planning mode (requirements gathering)")
	planningDefault := "7"
	if currentPlanning != "" {
		planningDefault = currentPlanning
	}
	planning, err := wizard.PromptString("  Max rounds", planningDefault, validateMaxRounds)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Planning max rounds", planning)

	// Execution
	ui.Dimmed("  Execution mode (parameter clarification)")
	executionDefault := "3"
	if currentExecution != "" {
		executionDefault = currentExecution
	}
	execution, err := wizard.PromptString("  Max rounds", executionDefault, validateMaxRounds)
	if err != nil {
		return err
	}
	ui.PromptSuccess("  Execution max rounds", execution)

	// Update formation.yaml
	return updateOverlordClarificationInFormation(rootDir, style, direct, brainstorm, planning, execution)
}

// Helper functions

func editInEditor(initialContent string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "muxi-persona-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	// Write initial content
	if _, err := tmpFile.WriteString(initialContent); err != nil {
		return "", err
	}
	tmpFile.Close()

	// Open editor
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Read result
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}

func validateFileExists(input string) error {
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", input)
	}
	return nil
}

func validateThreshold(input string) error {
	var val int
	_, err := fmt.Sscanf(input, "%d", &val)
	if err != nil || val < 1 || val > 10 {
		return fmt.Errorf("threshold must be between 1 and 10")
	}
	return nil
}

func validateMaxParallel(input string) error {
	var val int
	_, err := fmt.Sscanf(input, "%d", &val)
	if err != nil || val < 1 || val > 20 {
		return fmt.Errorf("max parallel tasks must be between 1 and 20")
	}
	return nil
}

func validateMaxRounds(input string) error {
	var val int
	_, err := fmt.Sscanf(input, "%d", &val)
	if err != nil || val < 1 || val > 32 {
		return fmt.Errorf("max rounds must be between 1 and 32")
	}
	return nil
}

func addCurrentIndicator(label string, isCurrent bool) string {
	if isCurrent {
		return label + " [current]"
	}
	return label
}

func getIndexForValue(values []string, current string, defaultIdx int) int {
	for i, v := range values {
		if v == current {
			return i
		}
	}
	return defaultIdx
}

func getCurrentOverlordPersona(rootDir string) string {
	formationFile, _ := context.FindFormationFile(rootDir)
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return ""
	}

	contentStr := string(content)

	// Look for persona under overlord
	// This is a simple extraction - persona is multi-line
	personaPattern := regexp.MustCompile(`(?s)overlord:\s*\n\s*persona:\s*\|?\s*\n((?:\s{4,}.*\n?)*)`)
	matches := personaPattern.FindStringSubmatch(contentStr)
	if len(matches) > 1 {
		// Remove leading indentation
		lines := strings.Split(matches[1], "\n")
		var result []string
		for _, line := range lines {
			trimmed := strings.TrimPrefix(line, "    ")
			if trimmed != "" || len(result) > 0 {
				result = append(result, trimmed)
			}
		}
		return strings.TrimSpace(strings.Join(result, "\n"))
	}

	return ""
}

func getCurrentOverlordValue(rootDir, section, key string) string {
	formationFile, _ := context.FindFormationFile(rootDir)
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return ""
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Navigate to the correct section
	inOverlord := false
	inSection := false
	sectionDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track overlord section
		if trimmed == "overlord:" {
			inOverlord = true
			continue
		}

		// Exit overlord if we hit another top-level key
		if inOverlord && len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") {
			inOverlord = false
			continue
		}

		if !inOverlord {
			continue
		}

		// Handle nested sections like "workflow.timeouts"
		sectionParts := strings.Split(section, ".")

		if len(sectionParts) == 1 {
			// Simple section like "response"
			if trimmed == section+":" {
				inSection = true
				sectionDepth = len(line) - len(strings.TrimLeft(line, " "))
				continue
			}
		} else {
			// Nested section like "workflow.timeouts"
			if trimmed == sectionParts[0]+":" {
				inSection = false // Reset for next level
				sectionDepth = len(line) - len(strings.TrimLeft(line, " "))
				continue
			}
			if trimmed == sectionParts[1]+":" {
				inSection = true
				sectionDepth = len(line) - len(strings.TrimLeft(line, " "))
				continue
			}
		}

		// Exit section if we hit a sibling or parent section
		if inSection {
			currentDepth := len(line) - len(strings.TrimLeft(line, " "))
			if currentDepth <= sectionDepth && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				if !strings.HasPrefix(trimmed, key+":") {
					inSection = false
					continue
				}
			}
		}

		// Look for key in current section
		if inSection && strings.HasPrefix(trimmed, key+":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) > 1 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, "\"'")
				return val
			}
		}
	}

	return ""
}

// Formation update functions

func updateOverlordPersonaInFormation(rootDir, persona string) error {
	formationFile, _ := context.FindFormationFile(rootDir)
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	// Indent persona lines
	personaLines := strings.Split(persona, "\n")
	var indentedPersona strings.Builder
	for _, line := range personaLines {
		indentedPersona.WriteString("    " + line + "\n")
	}

	// Build overlord persona YAML
	var overlordYAML strings.Builder
	overlordYAML.WriteString("  persona: |\n")
	overlordYAML.WriteString(indentedPersona.String())

	return updateOverlordSectionInFormation(formationFile, content, "persona", overlordYAML.String())
}

func updateOverlordResponseInFormation(rootDir, format string, streaming, progress bool) error {
	formationFile, _ := context.FindFormationFile(rootDir)
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	// Build response YAML
	var responseYAML strings.Builder
	responseYAML.WriteString("  response:\n")
	responseYAML.WriteString(fmt.Sprintf("    format: \"%s\"\n", format))
	responseYAML.WriteString(fmt.Sprintf("    streaming: %t\n", streaming))
	responseYAML.WriteString(fmt.Sprintf("    progress: %t\n", progress))

	return updateOverlordSectionInFormation(formationFile, content, "response", responseYAML.String())
}

func updateOverlordWorkflowInFormation(rootDir, routing string, autoDecomp bool, threshold, complexity string, parallel bool, maxParallel string, affinity bool, taskTimeout, workflowTimeout, errorRecovery string) error {
	formationFile, _ := context.FindFormationFile(rootDir)
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	// Build workflow YAML
	var workflowYAML strings.Builder
	workflowYAML.WriteString("  workflow:\n")
	workflowYAML.WriteString(fmt.Sprintf("    routing_strategy: \"%s\"\n", routing))
	workflowYAML.WriteString(fmt.Sprintf("    auto_decomposition: %t\n", autoDecomp))
	workflowYAML.WriteString(fmt.Sprintf("    plan_approval_threshold: %s\n", threshold))
	workflowYAML.WriteString(fmt.Sprintf("    complexity_method: \"%s\"\n", complexity))
	workflowYAML.WriteString(fmt.Sprintf("    parallel_execution: %t\n", parallel))
	if parallel && maxParallel != "" {
		workflowYAML.WriteString(fmt.Sprintf("    max_parallel_tasks: %s\n", maxParallel))
	}
	workflowYAML.WriteString(fmt.Sprintf("    enable_agent_affinity: %t\n", affinity))
	workflowYAML.WriteString(fmt.Sprintf("    error_recovery: \"%s\"\n", errorRecovery))
	workflowYAML.WriteString("    timeouts:\n")
	workflowYAML.WriteString(fmt.Sprintf("      task: %s\n", taskTimeout))
	workflowYAML.WriteString(fmt.Sprintf("      workflow: %s\n", workflowTimeout))

	return updateOverlordSectionInFormation(formationFile, content, "workflow", workflowYAML.String())
}

func updateOverlordClarificationInFormation(rootDir, style, direct, brainstorm, planning, execution string) error {
	formationFile, _ := context.FindFormationFile(rootDir)
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return err
	}

	// Build clarification YAML
	var clarificationYAML strings.Builder
	clarificationYAML.WriteString("  clarification:\n")
	clarificationYAML.WriteString(fmt.Sprintf("    style: \"%s\"\n", style))
	clarificationYAML.WriteString("    max_rounds:\n")
	clarificationYAML.WriteString(fmt.Sprintf("      direct: %s\n", direct))
	clarificationYAML.WriteString(fmt.Sprintf("      brainstorm: %s\n", brainstorm))
	clarificationYAML.WriteString(fmt.Sprintf("      planning: %s\n", planning))
	clarificationYAML.WriteString(fmt.Sprintf("      execution: %s\n", execution))

	return updateOverlordSectionInFormation(formationFile, content, "clarification", clarificationYAML.String())
}

func updateOverlordSectionInFormation(formationFile string, content []byte, section, sectionYAML string) error {
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Find overlord section and the specific subsection
	overlordLineIdx := -1
	sectionStartIdx := -1
	sectionEndIdx := -1
	inOverlord := false
	inSection := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Found overlord: section
		if trimmed == "overlord:" {
			overlordLineIdx = i
			inOverlord = true
			continue
		}

		// Exit overlord section if we hit a new top-level key
		if inOverlord && len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") && trimmed != "" {
			inOverlord = false
			if inSection {
				sectionEndIdx = i
				inSection = false
			}
			continue
		}

		// Found the specific subsection (e.g., "  persona:", "  response:", etc.)
		if inOverlord && trimmed == section+":" && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
			sectionStartIdx = i
			inSection = true
			continue
		}

		// Handle persona which might be multi-line with |
		if inOverlord && section == "persona" && strings.HasPrefix(trimmed, "persona:") {
			sectionStartIdx = i
			inSection = true
			continue
		}

		// Exit subsection if we hit another subsection or end of overlord
		if inSection && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			sectionEndIdx = i
			inSection = false
		}
	}

	// If section exists, mark end at EOF if not found
	if inSection && sectionEndIdx == -1 {
		sectionEndIdx = len(lines)
	}

	var result []string

	if sectionStartIdx >= 0 && sectionEndIdx >= 0 {
		// Replace existing section
		result = append(result, lines[:sectionStartIdx]...)
		result = append(result, strings.TrimSuffix(sectionYAML, "\n"))
		result = append(result, lines[sectionEndIdx:]...)
	} else if overlordLineIdx >= 0 {
		// Overlord exists but subsection doesn't - add it
		for i, line := range lines {
			result = append(result, line)
			if i == overlordLineIdx {
				result = append(result, strings.TrimSuffix(sectionYAML, "\n"))
			}
		}
	} else {
		// No overlord section - add at end
		result = lines
		result = append(result, "")
		result = append(result, "overlord:")
		result = append(result, strings.TrimSuffix(sectionYAML, "\n"))
	}

	output := strings.Join(result, "\n")
	if err := os.WriteFile(formationFile, []byte(output), 0644); err != nil {
		return err
	}

	ui.Success("Updated formation.yaml with overlord configuration")
	return nil
}
