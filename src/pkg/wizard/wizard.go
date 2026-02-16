package wizard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"golang.org/x/term"
)

// Session-wide input history for arrow key navigation
var inputHistory []string

// PromptString prompts for a string input with optional validation
// Loops on validation errors until valid input is provided
// Returns the validated input (without showing success - caller should use ui.PromptSuccess)
// Supports arrow keys for history navigation (up/down) and line editing (left/right)
func PromptString(prompt, defaultValue string, validator func(string) error) (string, error) {
	// Try to use readline for better UX (arrow keys, history, etc.)
	// Falls back to basic input if readline initialization fails

	// For long prompts (>60 chars), put input on new line
	useTwoLines := len(prompt) > 60

	var promptText string
	if useTwoLines {
		// Print prompt on first line, cursor on second
		if defaultValue != "" {
			fmt.Printf("%s [%s]:\n", prompt, defaultValue)
			promptText = "  " // Indent input
		} else {
			fmt.Printf("%s:\n", prompt)
			promptText = "  " // Indent input
		}
	} else {
		// Traditional single-line prompt
		if defaultValue != "" {
			promptText = fmt.Sprintf("%s [%s]: ", prompt, defaultValue)
		} else {
			promptText = prompt + ": "
		}
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 promptText,
		HistoryLimit:           100,
		DisableAutoSaveHistory: true, // We'll manage history manually
	})

	if err != nil {
		// Fallback to basic input if readline fails
		return promptStringFallback(prompt, defaultValue, validator)
	}
	defer rl.Close()

	// Load session history into readline
	for _, h := range inputHistory {
		rl.SaveHistory(h)
	}

	for {
		line, err := rl.Readline()
		if err != nil {
			// Handle EOF or errors
			return "", err
		}

		input := strings.TrimSpace(line)

		// Use default if empty
		if input == "" && defaultValue != "" {
			input = defaultValue
		}

		// Validate if validator provided
		if validator != nil && input != "" {
			if err := validator(input); err != nil {
				// Show error and prompt again (readline handles the prompt)
				fmt.Printf("%s\n\n", err.Error())
				// Update prompt for retry
				rl.SetPrompt(promptText)
				continue
			}
		}

		// Save to session history (only valid inputs)
		if input != "" {
			inputHistory = append(inputHistory, input)
			// Keep history size reasonable
			if len(inputHistory) > 100 {
				inputHistory = inputHistory[len(inputHistory)-100:]
			}
		}

		// Clear the input line (so caller can replace with success message)
		if useTwoLines {
			// Clear input line and prompt line
			fmt.Print("\033[1A\033[2K") // Clear input line
			fmt.Print("\033[1A\033[2K") // Clear prompt line
		} else {
			// Clear single-line prompt
			fmt.Print("\033[1A\033[2K")
		}

		return input, nil
	}
}

// promptStringFallback is the original implementation without readline support
// Used as fallback when readline initialization fails
func promptStringFallback(prompt, defaultValue string, validator func(string) error) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// For long prompts (>60 chars), put input on new line
	useTwoLines := len(prompt) > 60

	for {
		if useTwoLines {
			// Print prompt on first line, cursor on second
			if defaultValue != "" {
				fmt.Printf("%s [%s]:\n  ", prompt, defaultValue)
			} else {
				fmt.Printf("%s:\n  ", prompt)
			}
		} else {
			// Traditional single-line prompt
			if defaultValue != "" {
				fmt.Printf("%s [%s]: ", prompt, defaultValue)
			} else {
				fmt.Printf("%s: ", prompt)
			}
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		// Use default if empty
		if input == "" && defaultValue != "" {
			input = defaultValue
		}

		// Validate if validator provided
		if validator != nil && input != "" {
			if err := validator(input); err != nil {
				if useTwoLines {
					// Clear input line, move up, clear prompt line, show error
					fmt.Print("\033[1A\033[2K") // Clear input line
					fmt.Print("\033[1A\033[2K") // Clear prompt line
					if defaultValue != "" {
						fmt.Printf("%s [%s]: %s\n", prompt, defaultValue, input)
					} else {
						fmt.Printf("%s: %s\n", prompt, input)
					}
					fmt.Printf("%s\n\n", err.Error())
				} else {
					// Move cursor up one line, clear it, and show error
					fmt.Print("\033[1A\033[2K")
					fmt.Printf("%s: %s\n", prompt, input)
					fmt.Printf("%s\n\n", err.Error())
				}
				continue
			}
		}

		if useTwoLines {
			// Clear input line and prompt line (so caller can replace with success)
			fmt.Print("\033[1A\033[2K") // Clear input line
			fmt.Print("\033[1A\033[2K") // Clear prompt line
		} else {
			// Move cursor up one line and clear it (so caller can replace with success)
			fmt.Print("\033[1A\033[2K")
		}

		return input, nil
	}
}

// PromptPassword prompts for a password with masked input (shows **** as you type)
// Returns the input (without showing success - caller should use ui.PromptSuccess if needed)
func PromptPassword(prompt string, allowEmpty bool) (string, error) {
	fmt.Printf("%s: ", prompt)

	// Put terminal in raw mode to read character by character
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to hidden input if raw mode fails
		return promptPasswordFallback(prompt, allowEmpty)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	var password []byte
	buf := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			term.Restore(int(os.Stdin.Fd()), oldState)
			return "", err
		}

		switch buf[0] {
		case 13, 10: // Enter
			term.Restore(int(os.Stdin.Fd()), oldState)
			fmt.Println() // New line after input
			input := strings.TrimSpace(string(password))
			if !allowEmpty && input == "" {
				return "", fmt.Errorf("input cannot be empty")
			}
			return input, nil
		case 127, 8: // Backspace
			if len(password) > 0 {
				password = password[:len(password)-1]
				fmt.Print("\b \b") // Erase last *
			}
		case 3: // Ctrl+C
			term.Restore(int(os.Stdin.Fd()), oldState)
			fmt.Println()
			return "", fmt.Errorf("cancelled")
		default:
			if buf[0] >= 32 && buf[0] <= 126 { // Printable ASCII
				password = append(password, buf[0])
				fmt.Print("*")
			}
		}
	}
}

// promptPasswordFallback uses standard hidden input as fallback
func promptPasswordFallback(prompt string, allowEmpty bool) (string, error) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()
	input := strings.TrimSpace(string(password))
	if !allowEmpty && input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}
	return input, nil
}

// PromptConfirm prompts for yes/no confirmation
func PromptConfirm(prompt string, defaultYes bool) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	if defaultYes {
		fmt.Printf("%s [Y/n]: ", prompt)
	} else {
		fmt.Printf("%s [y/N]: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.ToLower(strings.TrimSpace(input))

	// Empty input uses default
	if input == "" {
		return defaultYes, nil
	}

	return input == "y" || input == "yes", nil
}

// SelectOption represents a selectable option
type SelectOption struct {
	Value       string // Internal value (e.g., "specialist")
	Label       string // Display label (e.g., "Specialist")
	Description string // Optional description (e.g., "Domain expert with specific skills")
}

// PromptSelect shows an interactive selection menu with arrow keys
// Returns the selected option's value
func PromptSelect(prompt string, options []SelectOption, defaultIndex int) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options provided")
	}

	if defaultIndex < 0 || defaultIndex >= len(options) {
		defaultIndex = 0
	}

	fmt.Printf("%s (↑↓ to select, Enter to confirm):\n", prompt)

	selected := defaultIndex

	// Disable input buffering and echo
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to numbered selection if terminal control fails
		return promptSelectFallback(options, defaultIndex)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Initial display
	displayOptions(options, selected)

	// Read keys
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			term.Restore(int(os.Stdin.Fd()), oldState)
			return "", err
		}

		if n == 1 {
			switch buf[0] {
			case 13: // Enter
				// Clear the menu
				clearLines(len(options) + 1)
				term.Restore(int(os.Stdin.Fd()), oldState)
				return options[selected].Value, nil
			case 3: // Ctrl+C
				term.Restore(int(os.Stdin.Fd()), oldState)
				return "", fmt.Errorf("cancelled")
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			// Arrow keys (ESC [ A/B)
			switch buf[2] {
			case 65: // Up arrow
				if selected > 0 {
					selected--
					clearLines(len(options))
					displayOptions(options, selected)
				}
			case 66: // Down arrow
				if selected < len(options)-1 {
					selected++
					clearLines(len(options))
					displayOptions(options, selected)
				}
			}
		}
	}
}

// displayOptions displays the selection menu
func displayOptions(options []SelectOption, selected int) {
	green := color.New(color.FgGreen, color.Bold)

	for i, opt := range options {
		symbol := "◯"
		if i == selected {
			symbol = "◉"
			// Selected option in green bold
			if opt.Description != "" {
				fmt.Print("  ")
				green.Printf("%s %s - %s", symbol, opt.Label, opt.Description)
				fmt.Print("\r\n")
			} else {
				fmt.Print("  ")
				green.Printf("%s %s", symbol, opt.Label)
				fmt.Print("\r\n")
			}
		} else {
			// Unselected option in normal color
			if opt.Description != "" {
				fmt.Printf("  %s %s - %s\r\n", symbol, opt.Label, opt.Description)
			} else {
				fmt.Printf("  %s %s\r\n", symbol, opt.Label)
			}
		}
	}
}

// clearLines clears n lines up from current cursor position
func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\033[1A\033[2K\r") // Move up, clear line, move to column 0
	}
}

// promptSelectFallback shows numbered selection as fallback
func promptSelectFallback(options []SelectOption, defaultIndex int) (string, error) {
	fmt.Println("\nSelect an option:")
	for i, opt := range options {
		defaultMarker := ""
		if i == defaultIndex {
			defaultMarker = " (default)"
		}
		if opt.Description != "" {
			fmt.Printf("  %d. %s - %s%s\n", i+1, opt.Label, opt.Description, defaultMarker)
		} else {
			fmt.Printf("  %d. %s%s\n", i+1, opt.Label, defaultMarker)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect [1-%d] or press Enter for default: ", len(options))

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	// Empty input uses default
	if input == "" {
		return options[defaultIndex].Value, nil
	}

	// Parse number
	var choice int
	_, err = fmt.Sscanf(input, "%d", &choice)
	if err != nil || choice < 1 || choice > len(options) {
		return "", fmt.Errorf("invalid selection")
	}

	return options[choice-1].Value, nil
}
