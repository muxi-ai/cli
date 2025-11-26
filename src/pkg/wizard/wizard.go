package wizard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptString prompts for a string input with optional validation
// Loops on validation errors until valid input is provided
// Returns the validated input (without showing success - caller should use ui.PromptSuccess)
func PromptString(prompt, defaultValue string, validator func(string) error) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
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
				// Move cursor up one line, clear it, and show error
				fmt.Print("\033[1A\033[2K")
				fmt.Printf("%s: %s\n", prompt, input)
				fmt.Printf("%s\n\n", err.Error())
				continue
			}
		}

		// Move cursor up one line and clear it (so caller can replace with success)
		fmt.Print("\033[1A\033[2K")
		
		return input, nil
	}
}

// PromptPassword prompts for a password (hidden input)
// Returns the input (without showing success - caller should use ui.PromptSuccess if needed)
func PromptPassword(prompt string, allowEmpty bool) (string, error) {
	fmt.Println(prompt)
	fmt.Print("    ")

	// Read password without echo
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	fmt.Println() // New line after password input

	input := strings.TrimSpace(string(password))

	if !allowEmpty && input == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Note: We don't clear the prompt line for passwords since the prompt is multi-line
	// The caller will handle showing success/skip status on a new line
	
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
	for i, opt := range options {
		symbol := "◯"
		if i == selected {
			symbol = "◉"
		}

		if opt.Description != "" {
			fmt.Printf("  %s %s - %s\r\n", symbol, opt.Label, opt.Description)
		} else {
			fmt.Printf("  %s %s\r\n", symbol, opt.Label)
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
