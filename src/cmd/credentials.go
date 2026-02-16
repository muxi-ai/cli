package cmd

import (
	"fmt"
	"strings"

	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
	"github.com/spf13/cobra"
)

var credentialsCmd = &cobra.Command{
	Use:     "credentials",
	Aliases: []string{"creds"},
	Short:   "Manage user credentials",
	GroupID: "formation",
	Long: `Manage user credentials for MCP services.

Credentials are stored encrypted and associated with a user ID.
They enable MCP servers to authenticate on behalf of users.`,
}

var credentialsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List user credentials",
	Long:    `List all credentials for a user. Shows metadata only - secrets are never exposed.`,
	RunE:    runCredentialsList,
}

var credentialsShowCmd = &cobra.Command{
	Use:   "show <credential-id>",
	Short: "Show credential details",
	Long:  `Show metadata for a specific credential. Secrets are never exposed.`,
	Args:  RequireArgs(1),
	RunE:  runCredentialsShow,
}

var credentialsAddCmd = &cobra.Command{
	Use:   "add [service]",
	Short: "Add a credential",
	Long: `Add a new credential for a user interactively.

If service is not provided, you'll be prompted to select from available services.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCredentialsAdd,
}

var credentialsDeleteCmd = &cobra.Command{
	Use:     "delete <credential-id>",
	Aliases: []string{"rm"},
	Short:   "Delete a credential",
	Long:    `Delete a credential for a user.`,
	Args:    RequireArgs(1),
	RunE:    runCredentialsDelete,
}

var credentialsServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List available credential services",
	Long:  `List MCP servers configured to use user credentials.`,
	RunE:  runCredentialsServices,
}

func init() {
	rootCmd.AddCommand(credentialsCmd)

	credentialsCmd.AddCommand(credentialsListCmd)
	credentialsCmd.AddCommand(credentialsShowCmd)
	credentialsCmd.AddCommand(credentialsAddCmd)
	credentialsCmd.AddCommand(credentialsDeleteCmd)
	credentialsCmd.AddCommand(credentialsServicesCmd)

	// Common flags
	formation.AddCommonFlags(credentialsListCmd)
	formation.AddCommonFlags(credentialsShowCmd)
	formation.AddCommonFlags(credentialsAddCmd)
	formation.AddCommonFlags(credentialsDeleteCmd)
	formation.AddCommonFlags(credentialsServicesCmd)
}

func runCredentialsList(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()

	resp, err := client.ListCredentials(userID)
	if err != nil {
		return err
	}

	if len(resp.Credentials) == 0 {
		ui.Dimmed("  No credentials stored")
		fmt.Println()
		fmt.Printf("  Add a credential: %s\n", ui.Command("muxi credentials add -u "+userID))
		fmt.Println()
		return nil
	}

	fmt.Printf("  %-15s %-12s %-20s %s\n", "ID", "SERVICE", "NAME", "PREVIEW")
	fmt.Printf("  %-15s %-12s %-20s %s\n", "───────────────", "────────────", "────────────────────", "─────────────────")

	for _, cred := range resp.Credentials {
		name := cred.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		fmt.Printf("  %-15s %-12s %-20s %s\n",
			truncate(cred.CredentialID, 15),
			cred.Service,
			name,
			cred.CredentialPreview)
	}

	fmt.Println()
	return nil
}

func runCredentialsShow(cmd *cobra.Command, args []string) error {
	credentialID := args[0]

	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()

	cred, err := client.GetCredential(credentialID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ui.ErrorBlock("Credential not found", fmt.Sprintf("Credential '%s' does not exist.", credentialID), "")
			return nil
		}
		return err
	}

	fmt.Printf("  Credential: %s\n", cred.CredentialID)
	fmt.Println()
	fmt.Printf("  Service:    %s\n", cred.Service)
	fmt.Printf("  Name:       %s\n", cred.Name)
	fmt.Printf("  Preview:    %s\n", cred.CredentialPreview)
	fmt.Println()
	fmt.Printf("  Created:    %s\n", cred.CreatedAt.Format("2006-01-02 15:04:05"))
	if !cred.UpdatedAt.IsZero() {
		fmt.Printf("  Updated:    %s\n", cred.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	return nil
}

func runCredentialsAdd(cmd *cobra.Command, args []string) error {
	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()

	var service string

	// If service provided as arg, use it
	if len(args) > 0 {
		service = normalizeService(args[0])
	} else {
		// Fetch available services
		servicesResp, err := client.ListCredentialServices()
		if err != nil {
			// If services endpoint fails, fall back to manual entry
			service, err = promptForService("")
			if err != nil {
				return err
			}
		} else if len(servicesResp.Services) == 0 {
			// No services configured, prompt for manual entry
			service, err = promptForService("")
			if err != nil {
				return err
			}
		} else {
			// Build select options from available services
			var options []wizard.SelectOption
			for _, svc := range servicesResp.Services {
				options = append(options, wizard.SelectOption{
					Value:       svc.Service,
					Label:       svc.Service,
					Description: svc.Description,
				})
			}
			// Add "Other" option
			options = append(options, wizard.SelectOption{
				Value: "__other__",
				Label: "Other (enter custom service)",
			})

			selected, err := wizard.PromptSelect("Service", options, 0)
			if err != nil {
				return err
			}

			if selected == "__other__" {
				service, err = promptForService("")
				if err != nil {
					return err
				}
			} else {
				service = selected
			}
		}
	}

	// Prompt for name
	name, err := wizard.PromptString("Name (e.g., company account, myusername)", "", nil)
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)

	// Prompt for token
	token, err := wizard.PromptPassword("Token", false)
	if err != nil {
		return err
	}

	// Create credential
	req := &formation.CreateCredentialRequest{
		Service: service,
		Name:    name,
		Credential: map[string]interface{}{
			"token": token,
		},
	}

	resp, err := client.CreateCredential(userID, req)
	if err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Credential added")
	fmt.Println()
	fmt.Printf("  ID:       %s\n", resp.CredentialID)
	fmt.Printf("  Service:  %s\n", resp.Service)
	fmt.Printf("  Name:     %s\n", resp.Name)
	fmt.Printf("  Preview:  %s\n", resp.CredentialPreview)
	fmt.Println()

	return nil
}

func runCredentialsDelete(cmd *cobra.Command, args []string) error {
	credentialID := args[0]

	client, userID, err := formation.ClientAndUserFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()

	// Confirm deletion
	confirm, err := wizard.PromptConfirm(fmt.Sprintf("Delete credential '%s'?", credentialID), false)
	if err != nil {
		return err
	}

	if !confirm {
		ui.Dimmed("  Cancelled")
		fmt.Println()
		return nil
	}

	_, err = client.DeleteCredential(credentialID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ui.ErrorBlock("Credential not found", fmt.Sprintf("Credential '%s' does not exist.", credentialID), "")
			return nil
		}
		return err
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Credential '%s' deleted", credentialID))
	fmt.Println()

	return nil
}

func runCredentialsServices(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()

	resp, err := client.ListCredentialServices()
	if err != nil {
		return err
	}

	if len(resp.Services) == 0 {
		ui.Dimmed("  No credential services configured")
		fmt.Println()
		return nil
	}

	fmt.Printf("  %-15s %-20s %s\n", "SERVICE", "SERVER ID", "DESCRIPTION")
	fmt.Printf("  %-15s %-20s %s\n", "───────────────", "────────────────────", "───────────────────────────")

	for _, svc := range resp.Services {
		desc := svc.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fmt.Printf("  %-15s %-20s %s\n", svc.Service, svc.ServerID, desc)
	}

	fmt.Println()
	return nil
}

// promptForService prompts for a service name with validation
func promptForService(defaultValue string) (string, error) {
	for {
		service, err := wizard.PromptString("Service name", defaultValue, nil)
		if err != nil {
			return "", err
		}
		service = strings.TrimSpace(service)

		if service == "" {
			fmt.Printf("  %s Service name cannot be empty\n\n", ui.RedText("✗"))
			continue
		}

		if strings.Contains(service, " ") {
			fmt.Printf("  %s Service name cannot contain spaces\n\n", ui.RedText("✗"))
			continue
		}

		// Normalize: lowercase
		return normalizeService(service), nil
	}
}

// normalizeService normalizes a service name (lowercase, no spaces)
func normalizeService(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
