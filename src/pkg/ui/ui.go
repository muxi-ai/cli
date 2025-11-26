package ui

import (
	"fmt"
	"strings"

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
)

// Success prints a success message with ✓ symbol (full line green, bold)
// Use this for important final success messages
func Success(message string) {
	green.Printf("%s %s\n", SymbolSuccess, message)
}

// Step prints a progress step with ✓ symbol (only icon colored, message normal)
// Use this for progress updates where you want less emphasis
func Step(message string) {
	green.Printf("%s ", SymbolSuccess)
	fmt.Println(message)
}

// Error prints an error message with ✗ symbol (red, bold)
func Error(message string) {
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

// Dimmed prints dimmed/faint text (~80% opacity)
func Dimmed(message string) {
	dimmed.Println(message)
}

// Bold prints bold text
func Bold(message string) {
	bold.Println(message)
}

// PromptSuccess shows a prompt with successful input (Next.js style)
// Example: ✓ Formation name: my-bot
func PromptSuccess(prompt, value string) {
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
func PromptSkipped(prompt string) {
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
//     muxi formation list
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

// InfoBanner displays an info message in a framed box
func InfoBanner(message string) {
	lines := strings.Split(message, "\n")
	
	// Find the longest line for width
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	
	// Top border
	fmt.Printf("╭%s╮\n", strings.Repeat("─", maxWidth+2))
	
	// Content lines
	for _, line := range lines {
		padding := maxWidth - len(line)
		fmt.Printf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}
	
	// Bottom border
	fmt.Printf("╰%s╯\n", strings.Repeat("─", maxWidth+2))
	fmt.Println()
}
