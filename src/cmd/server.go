package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage server connections",
	Long:  `Add, list, and manage MUXI Server connections.`,
}

var serverAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a server connection",
	RunE:  runServerAdd,
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List server connections",
	RunE:  runServerList,
}

var serverSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set default server",
	RunE:  runServerSet,
}

var serverRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a server connection",
	RunE:  runServerRemove,
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show server status",
	RunE:  runServerStatus,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.AddCommand(serverAddCmd)
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverSetCmd)
	serverCmd.AddCommand(serverRemoveCmd)
	serverCmd.AddCommand(serverStatusCmd)

	// Flags
	serverStatusCmd.Flags().String("profile", "", "Server profile to use")
}

// runServerAdd handles muxi server add
func runServerAdd(cmd *cobra.Command, args []string) error {
	fmt.Println()

	// Server name
	name, err := wizard.PromptString("Server name", "localhost", nil)
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Check if exists
	config, err := server.LoadServers()
	if err != nil {
		return err
	}
	if _, exists := config.Servers[name]; exists {
		return fmt.Errorf("server '%s' already exists", name)
	}

	// Server URL
	defaultURL := "http://localhost:7890"
	url, err := wizard.PromptString("Server URL", defaultURL, nil)
	if err != nil {
		return err
	}
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")

	fmt.Println()
	ui.Dimmed("  Enter credentials from ~/.muxi-server/credentials.yaml on the server")
	fmt.Println()

	// Key ID
	keyID, err := wizard.PromptString("Key ID", "", nil)
	if err != nil {
		return err
	}
	keyID = strings.TrimSpace(keyID)

	// Secret Key (masked)
	secretKey, err := wizard.PromptPassword("Secret Key", false)
	if err != nil {
		return err
	}
	secretKey = strings.TrimSpace(secretKey)

	// Test connection
	fmt.Println()
	fmt.Print("  Testing connection... ")

	entry := server.ServerEntry{
		URL:       url,
		KeyID:     keyID,
		SecretKey: secretKey,
		AddedAt:   time.Now(),
	}

	client := server.NewClientFromEntry(&entry)
	if err := client.Ping(); err != nil {
		fmt.Println()
		return fmt.Errorf("connection failed: %w", err)
	}

	// Verify auth by calling status
	_, err = client.GetServerStatus()
	if err != nil {
		fmt.Println()
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Println("OK")

	// Save
	if err := server.AddServer(name, entry); err != nil {
		return err
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Server \"%s\" added", name))

	// Set as default?
	if config.Default == "" || len(config.Servers) == 0 {
		if err := server.SetDefaultServer(name); err != nil {
			return err
		}
		fmt.Printf("  Set as default server\n")
	} else {
		fmt.Println()
		setDefault, err := wizard.PromptConfirm("Set as default server?", false)
		if err != nil {
			return err
		}
		if setDefault {
			if err := server.SetDefaultServer(name); err != nil {
				return err
			}
			ui.Success(fmt.Sprintf("Default server set to: %s", name))
		}
	}

	fmt.Println()
	return nil
}

// runServerList handles muxi server list
func runServerList(cmd *cobra.Command, args []string) error {
	config, err := server.LoadServers()
	if err != nil {
		return err
	}

	if len(config.Servers) == 0 {
		fmt.Println()
		ui.Dimmed("  No servers configured")
		fmt.Println()
		fmt.Println("  Add a server: muxi server add")
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Printf("  %-15s %-35s %s\n", "NAME", "URL", "STATUS")
	fmt.Printf("  %-15s %-35s %s\n", "----", "---", "------")

	for name, entry := range config.Servers {
		// Check connectivity
		client := server.NewClientFromEntry(&entry)
		status := ui.GreenText("● online")
		if err := client.Ping(); err != nil {
			status = ui.DimmedText("○ offline")
		}

		displayName := name
		if name == config.Default {
			displayName = name + " *"
		}

		fmt.Printf("  %-15s %-35s %s\n", displayName, entry.URL, status)
	}

	fmt.Println()
	ui.Dimmed("  * = default")
	fmt.Println()

	return nil
}

// runServerSet handles muxi server set
func runServerSet(cmd *cobra.Command, args []string) error {
	config, err := server.LoadServers()
	if err != nil {
		return err
	}

	if len(config.Servers) == 0 {
		fmt.Println()
		ui.Dimmed("  No servers configured")
		fmt.Println()
		fmt.Println("  Add a server: muxi server add")
		fmt.Println()
		return nil
	}

	// Build options
	var options []wizard.SelectOption
	currentIndex := 0
	i := 0
	for name := range config.Servers {
		label := name
		if name == config.Default {
			label += " [current]"
			currentIndex = i
		}
		options = append(options, wizard.SelectOption{
			Value: name,
			Label: label,
		})
		i++
	}

	fmt.Println()
	selected, err := wizard.PromptSelect("Select default server", options, currentIndex)
	if err != nil {
		return err
	}

	if err := server.SetDefaultServer(selected); err != nil {
		return err
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Default server set to: %s", selected))
	fmt.Println()

	return nil
}

// runServerRemove handles muxi server remove
func runServerRemove(cmd *cobra.Command, args []string) error {
	config, err := server.LoadServers()
	if err != nil {
		return err
	}

	if len(config.Servers) == 0 {
		fmt.Println()
		ui.Dimmed("  No servers configured")
		fmt.Println()
		return nil
	}

	// Build options
	var options []wizard.SelectOption
	for name := range config.Servers {
		options = append(options, wizard.SelectOption{
			Value: name,
			Label: name,
		})
	}

	fmt.Println()
	selected, err := wizard.PromptSelect("Select server to remove", options, 0)
	if err != nil {
		return err
	}

	// Confirm
	confirm, err := wizard.PromptConfirm(fmt.Sprintf("Remove \"%s\"?", selected), false)
	if err != nil {
		return err
	}

	if !confirm {
		fmt.Println()
		ui.Dimmed("  Cancelled")
		fmt.Println()
		return nil
	}

	if err := server.RemoveServer(selected); err != nil {
		return err
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Server \"%s\" removed", selected))
	fmt.Println()

	return nil
}

// runServerStatus handles muxi server status
func runServerStatus(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Get server name for display
	serverName := profile
	if serverName == "" {
		serverName = server.GetDefaultServer()
	}

	status, err := client.GetServerStatus()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  Server: %s (%s)\n", serverName, client.BaseURL)
	fmt.Println()

	// Status indicator
	statusIcon := ui.GreenText("●")
	if status.Server.Status != "healthy" {
		statusIcon = ui.RedText("●")
	}
	fmt.Printf("  Status:     %s %s\n", statusIcon, status.Server.Status)
	fmt.Printf("  Version:    %s\n", status.Server.Version)
	fmt.Printf("  Uptime:     %s\n", formatDuration(status.Server.Uptime))
	fmt.Println()
	fmt.Printf("  Formations: %d total (%d running, %d stopped)\n",
		status.Formations.Total, status.Formations.Running, status.Formations.Stopped)
	fmt.Printf("  Port Pool:  %d-%d (%d available)\n",
		status.Ports.Start, status.Ports.End, status.Ports.Available)
	fmt.Println()

	if status.Runtime.Type != "" {
		fmt.Printf("  Runtime:    %s (%s)\n", status.Runtime.Type, status.Runtime.Version)
	}
	if status.Server.Platform != "" {
		fmt.Printf("  Platform:   %s\n", status.Server.Platform)
	}
	fmt.Println()

	return nil
}

// formatDuration formats seconds into human-readable duration
func formatDuration(seconds int64) string {
	d := time.Duration(seconds) * time.Second

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
