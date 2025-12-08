package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Manage server connections",
	GroupID: "server",
	Long:    `Add, list, and manage MUXI Server connections.`,
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

var serverPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Test server connectivity",
	RunE:  runServerPing,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.AddCommand(serverAddCmd)
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverRemoveCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverPingCmd)

	// Flags
	serverStatusCmd.Flags().String("profile", "", "Server profile to use")
	serverPingCmd.Flags().String("profile", "", "Server profile to use")
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
	if _, err := client.Ping(); err != nil {
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
		fmt.Printf("  Add a server: %s\n", ui.Command("muxi server add"))
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
		if _, err := client.Ping(); err != nil {
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

	// Check health first
	health, healthErr := client.Health()

	// Get status
	status, err := client.GetServerStatus()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  Server: %s (%s)\n", serverName, client.BaseURL)
	fmt.Println()

	// Status from health endpoint
	statusIcon := ui.GreenText("●")
	statusText := "healthy"
	if healthErr != nil || health == nil {
		statusIcon = ui.RedText("●")
		statusText = "unreachable"
	} else if health.Data.Status != "ok" {
		statusIcon = ui.RedText("●")
		statusText = health.Data.Status
	}
	fmt.Printf("  Status:     %s %s\n", statusIcon, statusText)
	fmt.Printf("  Version:    %s\n", status.Server.Version)
	fmt.Printf("  Uptime:     %s\n", formatDuration(status.Server.Uptime))
	fmt.Println()
	fmt.Printf("  Formations: %d total (%d running, %d stopped)\n",
		status.Formations.Total, status.Formations.Running, status.Formations.Stopped)
	if status.Ports.Range != "" {
		fmt.Printf("  Port Pool:  %s (%d available)\n", status.Ports.Range, status.Ports.Available)
	}
	fmt.Println()

	if status.Runtime.Type != "" {
		runtimeVersion := ""
		if len(status.Runtime.Versions) > 0 {
			runtimeVersion = status.Runtime.Versions[0]
		}
		if runtimeVersion != "" {
			fmt.Printf("  Runtime:    %s (%s)\n", status.Runtime.Type, runtimeVersion)
		} else {
			fmt.Printf("  Runtime:    %s\n", status.Runtime.Type)
		}
	}
	if status.Runtime.Platform != "" {
		fmt.Printf("  Platform:   %s\n", status.Runtime.Platform)
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

// formatLatency formats a duration with appropriate units
func formatLatency(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// runServerPing handles muxi server ping
func runServerPing(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("PING %s (%s)\n", serverName, client.BaseURL)

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	seq := 0
	successCount := 0
	var totalLatency time.Duration

	for {
		select {
		case <-sigChan:
			// Print summary
			fmt.Println()
			fmt.Printf("--- %s ping statistics ---\n", serverName)
			lossPercent := float64(seq-successCount) / float64(seq) * 100
			fmt.Printf("%d packets transmitted, %d received, %.0f%% packet loss\n",
				seq, successCount, lossPercent)
			if successCount > 0 {
				avgLatency := totalLatency / time.Duration(successCount)
				fmt.Printf("avg latency: %s\n", formatLatency(avgLatency))
			}
			return nil

		default:
			seq++
			start := time.Now()
			bytes, err := client.Ping()
			latency := time.Since(start)

			if err != nil {
				fmt.Printf("seq=%d: %s\n", seq, ui.RedText("unreachable"))
			} else {
				successCount++
				totalLatency += latency
				fmt.Printf("%d bytes from %s: seq=%d time=%s\n", bytes, serverName, seq, formatLatency(latency))
			}

			time.Sleep(1 * time.Second)
		}
	}
}
