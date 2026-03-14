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

var profilesCmd = &cobra.Command{
	Use:     "profiles",
	Aliases: []string{"profile"},
	Short:   "Manage server profiles",
	GroupID: "remote",
	Long: `Add, list, and manage MUXI Server profiles.

Profiles are saved to ~/.muxi/cli/profiles.yaml and can be referenced by name
using the -p (--profile) flag in other commands.`,
}

var profilesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a server profile",
	Long: `Add a new MUXI Server profile interactively.

Prompts for server URL, name, and HMAC credentials. The server is verified
before being saved.`,
	RunE: runProfilesAdd,
}

var profilesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List server profiles",
	Long: `List all configured server profiles.

Shows profile name, URL, and online status. The default profile is marked
with an asterisk (*).`,
	RunE: runProfilesList,
}

var profilesRemoveCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "Remove a server profile",
	Long: `Remove a server profile interactively.

Select from a list of configured profiles to remove.`,
	RunE: runProfilesRemove,
}

var profilesStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show server status",
	Long: `Show detailed status for a server profile.

Displays server health, version, and deployed formations.`,
	RunE: runProfilesStatus,
}

var profilesPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Test server connectivity",
	Long: `Test connectivity to a server.

Sends a ping request and displays the response time.`,
	RunE: runProfilesPing,
}

func init() {
	rootCmd.AddCommand(profilesCmd)

	profilesCmd.AddCommand(profilesAddCmd)
	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesRemoveCmd)
	profilesCmd.AddCommand(profilesStatusCmd)
	profilesCmd.AddCommand(profilesPingCmd)

	// Flags
	profilesStatusCmd.Flags().String("profile", "", "Server profile to use")
	profilesPingCmd.Flags().String("profile", "", "Server profile to use")
}

// runProfilesAdd handles muxi profiles add
func runProfilesAdd(cmd *cobra.Command, args []string) error {
	fmt.Println()

	// Profile name
	name, err := wizard.PromptString("Profile name", "localhost", nil)
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check if exists
	config, err := server.LoadProfiles()
	if err != nil {
		return err
	}
	if _, exists := config.Profiles[name]; exists {
		return fmt.Errorf("profile '%s' already exists", name)
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
	ui.Dimmed("  Enter credentials from ~/.muxi/server/config.yaml on the server")
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

	entry := server.ProfileEntry{
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
	if err := server.AddProfile(name, entry); err != nil {
		return err
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Profile \"%s\" added", name))

	// Set as default?
	if config.DefaultProfile == "" || len(config.Profiles) == 0 {
		if err := server.SetDefaultProfile(name); err != nil {
			return err
		}
		fmt.Printf("  Set as default profile\n")
	} else {
		fmt.Println()
		setDefault, err := wizard.PromptConfirm("Set as default profile?", false)
		if err != nil {
			return err
		}
		if setDefault {
			if err := server.SetDefaultProfile(name); err != nil {
				return err
			}
			ui.Success(fmt.Sprintf("Default profile set to: %s", name))
		}
	}

	fmt.Println()
	return nil
}

// runProfilesList handles muxi profiles list
func runProfilesList(cmd *cobra.Command, args []string) error {
	config, err := server.LoadProfiles()
	if err != nil {
		return err
	}

	if len(config.Profiles) == 0 {
		fmt.Println()
		ui.Dimmed("  No profiles configured")
		fmt.Println()
		fmt.Printf("  Add a profile: %s\n", ui.Command("muxi profiles add"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Printf("  %-15s %-35s %s\n", "NAME", "URL", "STATUS")
	fmt.Printf("  %-15s %-35s %s\n", "----", "---", "------")

	for name, entry := range config.Profiles {
		// Check connectivity
		client := server.NewClientFromEntry(&entry)
		status := ui.GreenText("● online")
		if _, err := client.Ping(); err != nil {
			status = ui.DimmedText("○ offline")
		}

		displayName := name
		if name == config.DefaultProfile {
			displayName = name + " *"
		}

		fmt.Printf("  %-15s %-35s %s\n", displayName, entry.URL, status)
	}

	fmt.Println()
	ui.Dimmed("  * = default")
	fmt.Println()

	return nil
}

// runProfilesRemove handles muxi profiles remove
func runProfilesRemove(cmd *cobra.Command, args []string) error {
	config, err := server.LoadProfiles()
	if err != nil {
		return err
	}

	if len(config.Profiles) == 0 {
		fmt.Println()
		ui.Dimmed("  No profiles configured")
		fmt.Println()
		return nil
	}

	// Build options
	var options []wizard.SelectOption
	for name := range config.Profiles {
		options = append(options, wizard.SelectOption{
			Value: name,
			Label: name,
		})
	}

	fmt.Println()
	selected, err := wizard.PromptSelect("Select profile to remove", options, 0)
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

	if err := server.RemoveProfile(selected); err != nil {
		return err
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Profile \"%s\" removed", selected))
	fmt.Println()

	return nil
}

// runProfilesStatus handles muxi profiles status
func runProfilesStatus(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Get profile name for display
	profileName := profile
	if profileName == "" {
		profileName = server.GetDefaultProfile()
	}

	// Check health first
	health, healthErr := client.Health()

	// Get status
	status, err := client.GetServerStatus()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  Profile: %s (%s)\n", profileName, client.BaseURL)
	fmt.Println()

	// Status from health endpoint
	statusIcon := ui.GreenText("●")
	statusText := "healthy"
	if healthErr != nil || health == nil {
		statusIcon = ui.RedText("●")
		statusText = "unreachable"
	} else if health.GetStatus() != "ok" {
		statusIcon = ui.RedText("●")
		statusText = health.GetStatus()
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

// runProfilesPing handles muxi profiles ping
func runProfilesPing(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Get profile name for display
	profileName := profile
	if profileName == "" {
		profileName = server.GetDefaultProfile()
	}

	fmt.Printf("PING %s (%s)\n", profileName, client.BaseURL)

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
			fmt.Printf("--- %s ping statistics ---\n", profileName)
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
				fmt.Printf("%d bytes from %s: seq=%d time=%s\n", bytes, profileName, seq, formatLatency(latency))
			}

			time.Sleep(1 * time.Second)
		}
	}
}
