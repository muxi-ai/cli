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
