package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
)

// Status symbols
const (
	SymbolSuccess    = "✓"
	SymbolError      = "✗"
	SymbolWarning    = "⚠"
	SymbolInfo       = "ℹ"
	SymbolSkipped    = "⊘"
	SymbolInProgress = "●"
)

// Color helpers
var (
	green  = color.New(color.FgGreen, color.Bold)
	red    = color.New(color.FgRed, color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)
	blue   = color.New(color.FgBlue)
	cyan   = color.New(color.FgCyan)
	dimmed = color.New(color.Faint)
	bold   = color.New(color.Bold)
	gold   = color.RGB(216, 137, 62) // Brand color #d8893e
)

// Success prints a success message with ✓ symbol (full line green, bold)
// Use this for important final success messages
func Success(message string) {
	green.Printf("%s %s\n", SymbolSuccess, message)
}

// Step prints a progress step with ✓ symbol (only icon colored, message normal)
// Use this for progress updates where you want less emphasis
func Step(message string) {
	fmt.Print("  ")
	green.Printf("%s ", SymbolSuccess)
	fmt.Println(message)
}

// Error prints an error message with ✗ symbol (red, bold)
func Error(message string) {
	fmt.Print("  ")
	red.Printf("%s %s\n", SymbolError, message)
}

// Warning prints a warning message with ⚠ symbol (yellow, bold)
func Warning(message string) {
	yellow.Printf("%s %s\n", SymbolWarning, message)
}

// Info prints an info message with ℹ symbol (blue)
func Info(message string) {
	blue.Printf("%s %s\n", SymbolInfo, message)
}

// Skipped prints a skipped message with ⊘ symbol (cyan)
func Skipped(message string) {
	cyan.Printf("%s %s\n", SymbolSkipped, message)
}

// InProgress prints an in-progress message with ● symbol (blue)
func InProgress(message string) {
	blue.Printf("%s %s\n", SymbolInProgress, message)
}

// Spinner represents an animated spinner
type Spinner struct {
	message string
	frames  []string
	stop    chan struct{}
	done    sync.WaitGroup
	mu      sync.Mutex
	padding int // number of blank lines below spinner
}

// NewSpinner creates a new spinner with a message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stop:    make(chan struct{}),
		padding: 0,
	}
}

// NewSpinnerWithPadding creates a spinner with blank lines below for terminal margin
func NewSpinnerWithPadding(message string, padding int) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stop:    make(chan struct{}),
		padding: padding,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.done.Add(1)
	go func() {
		defer s.done.Done()
		i := 0
		first := true
		for {
			select {
			case <-s.stop:
				return
			default:
				s.mu.Lock()
				msg := s.message
				pad := s.padding
				s.mu.Unlock()

				// Print spinner
				fmt.Printf("\r\033[K  %s %s", blue.Sprint(s.frames[i%len(s.frames)]), msg)

				// On first iteration, print padding lines below and move cursor back up
				if first && pad > 0 {
					for j := 0; j < pad; j++ {
						fmt.Println() // Print blank lines
					}
					fmt.Printf("\033[%dA", pad) // Move cursor up
					first = false
				}

				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// UpdateMessage updates the spinner message while running
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	close(s.stop)
	s.done.Wait()
	fmt.Print("\r\033[K") // Clear line
	// Clear padding lines below
	for i := 0; i < s.padding; i++ {
		fmt.Print("\n\033[K")
	}
	if s.padding > 0 {
		fmt.Printf("\033[%dA", s.padding) // Move back up
	}
}

// StopWithSuccess stops spinner and shows success message
func (s *Spinner) StopWithSuccess(message string) {
	close(s.stop)
	s.done.Wait()
	fmt.Print("\r\033[K") // Clear line
	Step(message)
	// Move past padding lines (they'll be overwritten by next output)
	if s.padding > 0 {
		for i := 0; i < s.padding; i++ {
			fmt.Print("\033[K\n") // Clear each padding line
		}
	}
}

// StopWithError stops spinner and shows error message
func (s *Spinner) StopWithError(message string) {
	close(s.stop)
	s.done.Wait()
	fmt.Print("\r\033[K") // Clear line
	Error(message)
	// Move past padding lines
	if s.padding > 0 {
		for i := 0; i < s.padding; i++ {
			fmt.Print("\033[K\n") // Clear each padding line
		}
	}
}

// Dimmed prints dimmed/faint text (~80% opacity)
func Dimmed(message string) {
	dimmed.Println(message)
}

// Gold prints in brand color (golden/orange)
func Gold(message string) {
	gold.Println(message)
}

// Bold prints bold text
func Bold(message string) {
	bold.Println(message)
}

// Text helpers that return colored strings (for inline use)

// RedText returns red colored text
func RedText(s string) string {
	return red.Sprint(s)
}

// YellowText returns yellow colored text
func YellowText(s string) string {
	return yellow.Sprint(s)
}

// GreenText returns green colored text
func GreenText(s string) string {
	return green.Sprint(s)
}

// CyanText returns cyan colored text
func CyanText(s string) string {
	return cyan.Sprint(s)
}

// BoldText returns bold text
func BoldText(s string) string {
	return bold.Sprint(s)
}

// Command returns a command formatted for display (cyan/blue)
func Command(s string) string {
	return cyan.Sprint(s)
}

// DimmedText returns dimmed text
func DimmedText(s string) string {
	return dimmed.Sprint(s)
}

func GoldText(s string) string {
	return gold.Sprint(s)
}

// PromptSuccess shows a prompt with successful input (Next.js style)
// Example: ✓ Formation name: my-bot
// Supports leading spaces in prompt for indentation: "  Temperature" -> "  ✓ Temperature: 0.7"
func PromptSuccess(prompt, value string) {
	indent := ""
	trimmed := strings.TrimLeft(prompt, " ")
	if len(trimmed) < len(prompt) {
		indent = prompt[:len(prompt)-len(trimmed)]
		prompt = trimmed
	}
	fmt.Print(indent)
	green.Printf("%s ", SymbolSuccess)
	fmt.Printf("%s: ", prompt)
	bold.Println(value)
}

// PromptError shows a prompt with invalid input and error message
// Example:
// ✗ Formation name: My-Bot
//
//   Names must be lowercase
func PromptError(prompt, value string, err error) {
	red.Printf("%s ", SymbolError)
	fmt.Printf("%s: %s\n\n", prompt, value)

	// Print error message indented
	lines := strings.Split(err.Error(), "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Printf("  %s\n", line)
		} else {
			fmt.Println()
		}
	}
	fmt.Println()
}

// PromptSkipped shows a prompt that was skipped
// Example: ⊘ Description: skipped
// Supports leading spaces in prompt for indentation
func PromptSkipped(prompt string) {
	indent := ""
	trimmed := strings.TrimLeft(prompt, " ")
	if len(trimmed) < len(prompt) {
		indent = prompt[:len(prompt)-len(trimmed)]
		prompt = trimmed
	}
	fmt.Print(indent)
	cyan.Printf("%s ", SymbolSkipped)
	fmt.Printf("%s: ", prompt)
	dimmed.Println("skipped")
}

// Section prints a section header
func Section(title string) {
	fmt.Printf("\n%s\n", title)
}

// ErrorBlock prints a formatted error block with title and details
// Example:
// ✗ FORMATION NOT FOUND
//
//   Formation 'my-bot' does not exist
//
//   Check available formations:
//     muxi server list
func ErrorBlock(title string, details string, suggestion string) {
	red.Printf("%s %s\n\n", SymbolError, title)

	if details != "" {
		lines := strings.Split(details, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Printf("  %s\n", line)
			} else {
				fmt.Println()
			}
		}
	}

	if suggestion != "" {
		fmt.Println()
		dimmed.Println("  " + strings.ReplaceAll(suggestion, "\n", "\n  "))
	}
}

// SuccessBlock prints a formatted success block with title and next steps
func SuccessBlock(title string, nextSteps string) {
	green.Printf("%s %s\n", SymbolSuccess, title)

	if nextSteps != "" {
		fmt.Println()
		dimmed.Println(nextSteps)
	}
}

// List prints a bullet list
func List(items []string) {
	for _, item := range items {
		dimmed.Print("  • ")
		fmt.Println(item)
	}
}

// StatusList prints a list with status icons
func StatusList(items []StatusItem) {
	for _, item := range items {
		var symbol string
		var printer *color.Color

		switch item.Status {
		case "success":
			symbol = SymbolSuccess
			printer = green
		case "error":
			symbol = SymbolError
			printer = red
		case "warning":
			symbol = SymbolWarning
			printer = yellow
		case "skipped":
			symbol = SymbolSkipped
			printer = cyan
		default:
			symbol = SymbolInfo
			printer = blue
		}

		printer.Printf("  %s ", symbol)
		fmt.Printf("%s", item.Text)

		if item.Detail != "" {
			dimmed.Printf(" (%s)", item.Detail)
		}

		fmt.Println()
	}
}

// StatusItem represents an item in a status list
type StatusItem struct {
	Text   string // Main text
	Detail string // Optional detail (shown dimmed in parentheses)
	Status string // "success", "error", "warning", "skipped", "info"
}

// Indent adds indentation to text
func Indent(text string, level int) string {
	indent := strings.Repeat("  ", level)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// ProgressStep prints a multi-step progress indicator
// Example: [1/3] Validating formation...
func ProgressStep(current, total int, message string) {
	fmt.Printf("[%d/%d] %s\n", current, total, message)
}

// Confirm prints a confirmation prompt
func Confirm(prompt string, defaultYes bool) {
	if defaultYes {
		fmt.Printf("%s [Y/n]: ", prompt)
	} else {
		fmt.Printf("%s [y/N]: ", prompt)
	}
}

// Logo colors (gradient from light to dark gold)
var logoColors = []*color.Color{
	color.RGB(217, 170, 84),  // #d9aa54
	color.RGB(218, 158, 75),  // #da9e4b
	color.RGB(219, 150, 71),  // #db9647
	color.RGB(220, 143, 66),  // #dc8f42
	color.RGB(216, 137, 62),  // #d8893e
	color.RGB(191, 120, 64),  // #bf7840
}

// MUXIHeader prints the MUXI ASCII art logo with gradient colors and version info
func MUXIHeader(version, arch string) {
	logo := []string{
		"███╗   ███╗██╗   ██╗██╗  ██╗██╗",
		"████╗ ████║██║   ██║╚██╗██╔╝██║",
		"██╔████╔██║██║   ██║ ╚███╔╝ ██║",
		"██║╚██╔╝██║██║   ██║ ██╔██╗ ██║",
		"██║ ╚═╝ ██║╚██████╔╝██╔╝ ██╗██║",
		"╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝",
	}

	fmt.Println()
	for i, line := range logo {
		fmt.Print("  ")
		logoColors[i].Println(line)
	}
	fmt.Println()
	bold.Printf("  MUXI CLI %s (Apache 2.0 %s)\n", version, arch)
	fmt.Println()
	dimmed.Println("   * Documentation:  https://muxi.ai/docs")
	dimmed.Println("   * Support:        https://muxi.ai/support")
	fmt.Println()
	fmt.Println()
}

// Banner displays a pre-formatted banner with box drawing characters
// The banner string should already include all formatting (newlines, borders, etc.)
// "MUXI" text in banners is automatically colored gold (brand color)
func Banner(banner string) {
	// Color "MUXI" in gold within banners
	lines := strings.Split(banner, "\n")
	for _, line := range lines {
		if strings.Contains(line, "MUXI │") {
			// Split at MUXI and print with gold color
			parts := strings.SplitN(line, "MUXI", 2)
			fmt.Print(parts[0])
			gold.Print("MUXI")
			fmt.Println(parts[1])
		} else {
			fmt.Println(line)
		}
	}
	fmt.Println()
}

// FormationMCPBanner displays the formation-level MCP banner with red warning
func FormationMCPBanner() {
	fmt.Println("╭──────────────────────────────────────────────────────────────╮")
	fmt.Print("│ [+] Adding new MCP to formation                         ")
	gold.Print("MUXI")
	fmt.Println(" │")
	fmt.Println("│──────────────────────────────────────────────────────────────│")
	fmt.Println("│ MCPs (Model Context Protocol) are tools that agents use      │")
	fmt.Println("│ to interact with external services, APIs, and databases.     │")
	fmt.Println("│──────────────────────────────────────────────────────────────│")

	// Warning line in red
	fmt.Print("│ ")
	red.Print("⚠ Formation-level MCPs can be used by all agents.")
	fmt.Println("            │")

	fmt.Println("│                                                              │")
	fmt.Println("│ For tools specific to one agent, use:                        │")
	fmt.Println("│   $ muxi new mcp --agent <agent-id>                          │")
	fmt.Println("╰──────────────────────────────────────────────────────────────╯")
	fmt.Println()
}

// InfoBanner displays an info message in a framed box with fixed width (64 chars)
func InfoBanner(message string) {
	const frameWidth = 64  // Total frame width including borders
	const contentWidth = 60 // Content area width (frameWidth - 4 for borders and padding)

	lines := strings.Split(message, "\n")

	// Top border
	fmt.Printf("╭%s╮\n", strings.Repeat("─", frameWidth-2))

	// Content lines
	for _, line := range lines {
		if len(line) > contentWidth {
			// Truncate if too long (shouldn't happen with proper messages)
			line = line[:contentWidth]
		}
		padding := contentWidth - len(line)
		fmt.Printf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}

	// Bottom border
	fmt.Printf("╰%s╯\n", strings.Repeat("─", frameWidth-2))
	fmt.Println()
}

// RenderMarkdown renders markdown content with syntax highlighting and formatting
func RenderMarkdown(content string) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(76),
	)
	if err != nil {
		return content
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		return content
	}

	// Trim extra whitespace glamour adds
	return strings.TrimSpace(rendered)
}

// IndentString adds n spaces to the beginning of each line
func IndentString(s string, n int) string {
	indent := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// RenderYAML renders YAML content with syntax highlighting (no background)
func RenderYAML(content string) string {
	lexer := lexers.Get("yaml")
	if lexer == nil {
		return content
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return content
	}

	return buf.String()
}
