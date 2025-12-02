package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/muxi-ai/cli/pkg/registry"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// loginCmd handles muxi login
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the registry",
	Long:  "Authenticate with the MUXI registry using GitHub OAuth",
	RunE:  runLogin,
}

// logoutCmd handles muxi logout
var logoutCmd = &cobra.Command{
	Use:   "logout [registry]",
	Short: "Remove registry credentials",
	Long:  "Remove stored credentials for a registry",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLogout,
}

// pushCmd handles muxi push
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Publish formation to registry",
	Long:  "Publish the current formation to the MUXI registry",
	RunE:  runPush,
}

// pullCmd handles muxi pull
var pullCmd = &cobra.Command{
	Use:   "pull <@user/formation[:version]>",
	Short: "Download formation from registry",
	Long:  "Download a formation from the MUXI registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runPull,
}

// searchCmd handles muxi search
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for formations",
	Long:  "Search for formations in the MUXI registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

// showCmd handles muxi show
var showCmd = &cobra.Command{
	Use:   "show <@user/formation[:version]>",
	Short: "Display formation details",
	Long:  "Display detailed information about a formation",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

// registryCmd is the parent command for registry subcommands
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Registry commands",
	Long:  "Commands for interacting with the MUXI registry",
}

// Subcommand aliases for muxi registry *
var (
	registryLoginCmd = &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the registry",
		RunE:  runLogin,
	}
	registryLogoutCmd = &cobra.Command{
		Use:   "logout [registry]",
		Short: "Remove registry credentials",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLogout,
	}
	registryPushCmd = &cobra.Command{
		Use:   "push",
		Short: "Publish formation to registry",
		RunE:  runPush,
	}
	registryPullCmd = &cobra.Command{
		Use:   "pull <@user/formation[:version]>",
		Short: "Download formation from registry",
		Args:  cobra.ExactArgs(1),
		RunE:  runPull,
	}
	registrySearchCmd = &cobra.Command{
		Use:   "search <query>",
		Short: "Search for formations",
		Args:  cobra.ExactArgs(1),
		RunE:  runSearch,
	}
	registryShowCmd = &cobra.Command{
		Use:   "show <@user/formation[:version]>",
		Short: "Display formation details",
		Args:  cobra.ExactArgs(1),
		RunE:  runShow,
	}
	registryListCmd = &cobra.Command{
		Use:   "list",
		Short: "List configured registries",
		RunE:  runRegistryList,
	}
	registryAddCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a registry",
		RunE:  runRegistryAdd,
	}
	registryRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove a registry",
		RunE:  runRegistryRemove,
	}
	registryDefaultCmd = &cobra.Command{
		Use:   "default",
		Short: "Set default registry",
		RunE:  runRegistryDefault,
	}
)

func init() {
	// Login flags
	loginCmd.Flags().String("registry", "", "Registry to authenticate with")

	// Logout flags
	logoutCmd.Flags().String("registry", "", "Registry to logout from")

	// Push flags
	pushCmd.Flags().String("org", "", "Publish to organization")
	pushCmd.Flags().Bool("dry-run", false, "Show what would be published")
	pushCmd.Flags().String("registry", "", "Registry to publish to")

	// Pull flags
	pullCmd.Flags().StringP("output", "o", "", "Output directory")
	pullCmd.Flags().Bool("force", false, "Overwrite existing directory")
	pullCmd.Flags().String("registry", "", "Registry to pull from")

	// Search flags
	searchCmd.Flags().String("sort", "trending", "Sort by: trending, downloads, stars, recent")
	searchCmd.Flags().Int("limit", 20, "Maximum results (1-100)")
	searchCmd.Flags().String("registry", "", "Registry to search")

	// Show flags
	showCmd.Flags().Bool("versions", false, "Show all versions")
	showCmd.Flags().String("registry", "", "Registry to query")

	// Register top-level commands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(showCmd)

	// Register registry parent command with subcommands as aliases
	rootCmd.AddCommand(registryCmd)

	// Add flags to registry subcommands (same as top-level)
	registryLoginCmd.Flags().String("registry", "", "Registry to authenticate with")
	registryPushCmd.Flags().String("org", "", "Publish to organization")
	registryPushCmd.Flags().Bool("dry-run", false, "Show what would be published")
	registryPushCmd.Flags().String("registry", "", "Registry to publish to")
	registryPullCmd.Flags().StringP("output", "o", "", "Output directory")
	registryPullCmd.Flags().Bool("force", false, "Overwrite existing directory")
	registryPullCmd.Flags().String("registry", "", "Registry to pull from")
	registrySearchCmd.Flags().String("sort", "trending", "Sort by: trending, downloads, stars, recent")
	registrySearchCmd.Flags().Int("limit", 20, "Maximum results (1-100)")
	registrySearchCmd.Flags().String("registry", "", "Registry to search")
	registryShowCmd.Flags().Bool("versions", false, "Show all versions")
	registryShowCmd.Flags().String("registry", "", "Registry to query")

	// Register subcommands under registry
	registryCmd.AddCommand(registryLoginCmd)
	registryCmd.AddCommand(registryLogoutCmd)
	registryCmd.AddCommand(registryPushCmd)
	registryCmd.AddCommand(registryPullCmd)
	registryCmd.AddCommand(registrySearchCmd)
	registryCmd.AddCommand(registryShowCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registryDefaultCmd)
}

// runLogin handles the login command
func runLogin(cmd *cobra.Command, args []string) error {
	registryFlag, _ := cmd.Flags().GetString("registry")

	client, err := registry.NewClient(registryFlag)
	if err != nil {
		return err
	}

	registryName := registry.GetDefaultRegistry()
	if registryFlag != "" {
		registryName = registryFlag
	}

	// Check if already logged in
	if client.IsAuthenticated() {
		username, _ := registry.GetUsername(registryName)
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  Already logged in to %s as %s", registryName, username))
		fmt.Println()
		fmt.Println("  To log out, run: muxi logout")
		return nil
	}

	fmt.Println()
	ui.Bold("Authenticate with Registry")
	fmt.Println()

	// Try browser callback first
	token, username, err := tryBrowserAuth(client)
	if err != nil {
		// Fallback to manual paste
		token, err = manualAuth(client)
		if err != nil {
			return err
		}
		username = ""
	}

	// If no username from callback, try to validate token
	if username == "" {
		fmt.Println()
		fmt.Print("  Validating token... ")
		user, err := client.ValidateToken(token)
		if err == nil && user != nil {
			username = user.Username
		}
		fmt.Println("OK")
	}

	// Save token
	if err := registry.SetToken(registryName, token, username); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println()
	if username != "" {
		ui.Success(fmt.Sprintf("Logged in as %s", username))
	} else {
		ui.Success("Logged in successfully")
	}

	// Show where token is saved
	path, _ := registry.GetRegistriesPath()
	ui.Dimmed(fmt.Sprintf("  Token saved to %s", path))

	return nil
}

// authResult holds token and username from auth callback
type authResult struct {
	Token    string
	Username string
}

// tryBrowserAuth attempts browser-based authentication
func tryBrowserAuth(client *registry.Client) (string, string, error) {
	// Find available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", fmt.Errorf("failed to start local server")
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// Channel to receive auth result
	resultChan := make(chan authResult, 1)
	errChan := make(chan error, 1)

	// Start local server
	server := &http.Server{}
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			errChan <- fmt.Errorf("no token received")
			w.Write([]byte("Authentication failed. Please try again."))
			return
		}
		username := r.URL.Query().Get("username")
		resultChan <- authResult{Token: token, Username: username}
		w.Write([]byte(`
			<html>
			<body style="font-family: sans-serif; text-align: center; padding: 50px;">
			<h1>✓ Authentication Successful</h1>
			<p>You can close this window and return to the terminal.</p>
			</body>
			</html>
		`))
	})

	go func() {
		server.Serve(listener)
	}()

	// Open browser
	authURL := client.GetAuthURL(port)
	fmt.Printf("  Opening browser to authenticate...\n")
	fmt.Printf("  %s\n", authURL)
	fmt.Println()

	if err := openBrowser(authURL); err != nil {
		listener.Close()
		return "", "", fmt.Errorf("failed to open browser")
	}

	fmt.Println("  Waiting for authentication...")

	// Wait for token or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	select {
	case result := <-resultChan:
		server.Shutdown(context.Background())
		return result.Token, result.Username, nil
	case err := <-errChan:
		server.Shutdown(context.Background())
		return "", "", err
	case <-ctx.Done():
		server.Shutdown(context.Background())
		return "", "", fmt.Errorf("authentication timed out")
	}
}

// manualAuth prompts user to paste token manually
func manualAuth(client *registry.Client) (string, error) {
	authURL := client.GetAuthURL(0)
	fmt.Printf("  Visit this URL to authenticate:\n")
	fmt.Printf("  %s\n", authURL)
	fmt.Println()

	fmt.Print("  Paste your token: ")
	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("no token provided")
	}

	return token, nil
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// runLogout handles the logout command
func runLogout(cmd *cobra.Command, args []string) error {
	registryFlag, _ := cmd.Flags().GetString("registry")

	var registryName string
	if registryFlag != "" {
		registryName = registryFlag
	} else if len(args) > 0 {
		registryName = args[0]
	} else {
		registryName = registry.GetDefaultRegistry()
	}

	if !registry.IsLoggedIn(registryName) {
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  Not logged in to %s", registryName))
		return nil
	}

	if err := registry.RemoveToken(registryName); err != nil {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Logged out from %s", registryName))

	return nil
}

// runPush handles the push command
func runPush(cmd *cobra.Command, args []string) error {
	org, _ := cmd.Flags().GetString("org")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	registryFlag, _ := cmd.Flags().GetString("registry")

	// Check we're in a formation directory
	formationPath := filepath.Join(".", "formation.yaml")
	if _, err := os.Stat(formationPath); os.IsNotExist(err) {
		return fmt.Errorf("formation.yaml not found - run this from a formation directory")
	}

	// Create client and check auth
	client, err := registry.NewClient(registryFlag)
	if err != nil {
		return err
	}

	if !client.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run: muxi login")
	}

	fmt.Println()
	ui.Bold("Publishing Formation")
	fmt.Println()

	// Validate formation
	fmt.Print("  Validating formation... ")
	formationData, err := os.ReadFile(formationPath)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("failed to read formation.yaml: %w", err)
	}

	var formation struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal(formationData, &formation); err != nil {
		fmt.Println()
		return fmt.Errorf("invalid formation.yaml: %w", err)
	}

	if formation.Name == "" {
		fmt.Println()
		return fmt.Errorf("formation.yaml missing 'name' field")
	}
	if formation.Version == "" {
		fmt.Println()
		return fmt.Errorf("formation.yaml missing 'version' field")
	}
	fmt.Println("OK")

	fmt.Printf("  ✓ Name: %s\n", formation.Name)
	fmt.Printf("  ✓ Version: %s\n", formation.Version)

	// Count components
	agents := countFiles("agents", "*.yaml")
	mcps := countFiles("mcps", "*.yaml")
	sops := countFiles("sops", "*.md")
	triggers := countFiles("triggers", "*.yaml")

	if agents+mcps+sops+triggers > 0 {
		parts := []string{}
		if agents > 0 {
			parts = append(parts, fmt.Sprintf("%d agents", agents))
		}
		if mcps > 0 {
			parts = append(parts, fmt.Sprintf("%d MCPs", mcps))
		}
		if sops > 0 {
			parts = append(parts, fmt.Sprintf("%d SOPs", sops))
		}
		if triggers > 0 {
			parts = append(parts, fmt.Sprintf("%d triggers", triggers))
		}
		fmt.Printf("  ✓ Components: %s\n", strings.Join(parts, ", "))
	}

	fmt.Println()

	// Create bundle
	fmt.Print("  Creating bundle... ")
	bundle, err := registry.CreateBundle(".")
	if err != nil {
		fmt.Println()
		return fmt.Errorf("failed to create bundle: %w", err)
	}
	defer registry.CleanupBundle(bundle.Path)
	fmt.Println("OK")

	fmt.Printf("  ✓ %d files (%s)\n", bundle.FileCount, registry.FormatSize(bundle.Size))

	// Show warnings
	for _, warning := range bundle.Warnings {
		fmt.Println()
		ui.Warning("  " + warning)
	}

	if dryRun {
		fmt.Println()
		ui.Dimmed("  Dry run - no changes made")
		return nil
	}

	fmt.Println()

	// Publish
	fmt.Print("  Publishing to registry... ")
	result, err := client.Publish(bundle.Path, org)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("publish failed: %w", err)
	}
	fmt.Println("OK")

	fmt.Println()
	ui.Success(fmt.Sprintf("Published %s v%s", result.Formation, result.Version))
	fmt.Println()
	fmt.Println("  View at:")
	fmt.Printf("    Registry: %s\n", result.RegistryURL)
	if result.GitHubURL != "" {
		fmt.Printf("    GitHub:   %s\n", result.GitHubURL)
	}
	fmt.Println()
	fmt.Println("  Share with:")
	fmt.Printf("    muxi pull %s\n", result.Formation)

	return nil
}

// runPull handles the pull command
func runPull(cmd *cobra.Command, args []string) error {
	ref := args[0]
	outputDir, _ := cmd.Flags().GetString("output")
	force, _ := cmd.Flags().GetBool("force")
	registryFlag, _ := cmd.Flags().GetString("registry")

	client, err := registry.NewClient(registryFlag)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  Fetching %s...\n", ref)

	// Get pull info
	info, err := client.GetPullInfo(ref)
	if err != nil {
		return err
	}

	fmt.Printf("  ✓ Found v%s (%s)\n", info.Version, registry.FormatSize(info.Size))

	// Determine output directory
	if outputDir == "" {
		parsed, _ := registry.ParseFormationRef(ref)
		outputDir = parsed.Name
	}

	// Check if directory exists
	if _, err := os.Stat(outputDir); err == nil {
		if !force {
			return fmt.Errorf("directory %s already exists - use --force to overwrite", outputDir)
		}
		// Remove existing directory
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	fmt.Println()

	// Download
	fmt.Print("  Downloading... ")
	tmpFile, err := os.CreateTemp("", "muxi-pull-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := client.DownloadFormation(info.DownloadURL, tmpPath); err != nil {
		return err
	}
	fmt.Println("OK")

	// Extract
	fmt.Printf("  Extracting to %s/... ", outputDir)
	fileCount, err := registry.ExtractBundle(tmpPath, outputDir)
	if err != nil {
		return err
	}
	fmt.Println("OK")

	fmt.Printf("  ✓ %d files extracted\n", fileCount)

	fmt.Println()
	ui.Success(fmt.Sprintf("Downloaded %s v%s", info.Formation, info.Version))
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    cd %s\n", outputDir)
	fmt.Println("    muxi secrets setup    # Configure required secrets")

	return nil
}

// runSearch handles the search command
func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	sort, _ := cmd.Flags().GetString("sort")
	registryFlag, _ := cmd.Flags().GetString("registry")

	client, err := registry.NewClient(registryFlag)
	if err != nil {
		return err
	}

	// Fetch up to 100 results for pagination
	result, err := client.Search(query, sort, 100)
	if err != nil {
		return err
	}

	if len(result.Results) == 0 {
		fmt.Println()
		ui.Dimmed("  No formations found")
		return nil
	}

	// Paginate results
	pageSize := 10
	totalResults := len(result.Results)
	totalPages := (totalResults + pageSize - 1) / pageSize

	// Single page - show without pagination
	if totalPages == 1 {
		fmt.Println()
		fmt.Printf("  Found %d formations:\n", totalResults)
		fmt.Println()

		for _, f := range result.Results {
			fmt.Printf("  @%s/%s", f.User, f.Name)
			fmt.Printf("        ★ %d   ↓ %d\n", f.Stars, f.Downloads)
			if f.Description != "" {
				ui.Dimmed(fmt.Sprintf("    %s", f.Description))
			}
			if f.Version != "" {
				ui.Dimmed(fmt.Sprintf("    v%s", f.Version))
			}
			fmt.Println()
		}

		fmt.Println("  Pull with: muxi pull @user/formation")
		return nil
	}

	// Multiple pages - interactive pagination
	currentPage := 0
	reader := bufio.NewReader(os.Stdin)

	for {
		// Clear screen and show current page
		fmt.Print("\033[H\033[2J")

		start := currentPage * pageSize
		end := start + pageSize
		if end > totalResults {
			end = totalResults
		}

		// Header
		fmt.Println()
		if result.Total > totalResults {
			fmt.Printf("  Search: \"%s\" - Showing %d-%d of %d+ results\n", query, start+1, end, result.Total)
		} else {
			fmt.Printf("  Search: \"%s\" - Showing %d-%d of %d results\n", query, start+1, end, totalResults)
		}
		fmt.Println()

		// Display current page
		for i := start; i < end; i++ {
			f := result.Results[i]
			fmt.Printf("  @%s/%s", f.User, f.Name)
			fmt.Printf("        ★ %d   ↓ %d\n", f.Stars, f.Downloads)

			if f.Description != "" {
				ui.Dimmed(fmt.Sprintf("    %s", f.Description))
			}
			if f.Version != "" {
				ui.Dimmed(fmt.Sprintf("    v%s", f.Version))
			}
			fmt.Println()
		}

		// Navigation
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  Page %d of %d", currentPage+1, totalPages))
		fmt.Println()

		// Build navigation options
		var navOptions []string
		if currentPage > 0 {
			navOptions = append(navOptions, "[p]revious")
		}
		if currentPage < totalPages-1 {
			navOptions = append(navOptions, "[n]ext")
		}
		navOptions = append(navOptions, "[q]uit")

		fmt.Printf("  %s: ", strings.Join(navOptions, " | "))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "n", "next":
			if currentPage < totalPages-1 {
				currentPage++
			}
		case "p", "previous", "prev":
			if currentPage > 0 {
				currentPage--
			}
		case "q", "quit", "":
			fmt.Println()
			fmt.Println("  Pull with: muxi pull @user/formation")
			return nil
		}
	}
}

// runShow handles the show command
func runShow(cmd *cobra.Command, args []string) error {
	ref := args[0]
	showVersions, _ := cmd.Flags().GetBool("versions")
	registryFlag, _ := cmd.Flags().GetString("registry")

	client, err := registry.NewClient(registryFlag)
	if err != nil {
		return err
	}

	formation, err := client.GetFormation(ref, false)
	if err != nil {
		return err
	}

	fmt.Println()
	ui.Bold(fmt.Sprintf("@%s/%s v%s", formation.Owner, formation.Name, formation.Version))
	fmt.Println()

	if formation.Description != "" {
		fmt.Printf("  %s\n", formation.Description)
		fmt.Println()
	}

	// Stats
	fmt.Printf("  ★ %d stars   ↓ %d downloads   ⊞ %s\n",
		formation.Stars, formation.Downloads, registry.FormatSize(formation.Size))
	fmt.Println()

	// Components
	if formation.Components.Agents > 0 || formation.Components.MCPs > 0 ||
		formation.Components.SOPs > 0 || formation.Components.Triggers > 0 {
		fmt.Println("  Components:")
		if formation.Components.Agents > 0 {
			fmt.Printf("    • %d agents\n", formation.Components.Agents)
		}
		if formation.Components.MCPs > 0 {
			fmt.Printf("    • %d MCPs\n", formation.Components.MCPs)
		}
		if formation.Components.SOPs > 0 {
			fmt.Printf("    • %d SOPs\n", formation.Components.SOPs)
		}
		if formation.Components.Triggers > 0 {
			fmt.Printf("    • %d triggers\n", formation.Components.Triggers)
		}
		fmt.Println()
	}

	// Links
	fmt.Println("  Links:")
	fmt.Printf("    Registry: %s\n", formation.RegistryURL)
	if formation.GitHubURL != "" {
		fmt.Printf("    GitHub:   %s\n", formation.GitHubURL)
	}
	fmt.Println()

	// Dates
	fmt.Printf("  Published: %s\n", formation.CreatedAt.Format("Jan 2, 2006"))
	fmt.Printf("  Updated:   %s\n", formatTimeAgo(formation.UpdatedAt))
	fmt.Println()

	// Show versions if requested
	if showVersions {
		versions, err := client.GetVersions(ref)
		if err != nil {
			return err
		}

		fmt.Println("  Versions:")
		for _, v := range versions {
			fmt.Printf("    v%s (%s) - %s\n",
				v.Version,
				registry.FormatSize(v.Size),
				formatTimeAgo(v.CreatedAt))
		}
		fmt.Println()
	}

	// Pull command
	fmt.Printf("  Pull with: muxi pull @%s/%s\n", formation.Owner, formation.Name)

	return nil
}

// countFiles counts files matching pattern in a directory
func countFiles(dir, pattern string) int {
	files, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return 0
	}
	return len(files)
}

// formatTimeAgo formats a time as "X ago"
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// runRegistryList handles muxi registry list
func runRegistryList(cmd *cobra.Command, args []string) error {
	ui.InfoBanner("Configured Registries")

	config, err := registry.LoadRegistries()
	if err != nil {
		return fmt.Errorf("failed to load registries: %w", err)
	}

	if len(config.Registries) == 0 {
		fmt.Println()
		ui.Dimmed("  No registries configured")
		fmt.Println()
		fmt.Println("  Add a registry with: muxi registry add")
		fmt.Println("  Or login to default:  muxi login")
		return nil
	}

	fmt.Println()
	for name, entry := range config.Registries {
		// Mark default
		defaultMark := ""
		if name == config.DefaultRegistry {
			defaultMark = ui.DimmedText(" (default)")
		}

		// Status indicator
		status := ui.GreenText("●")
		if entry.Token == "" {
			status = ui.DimmedText("○")
		}

		fmt.Printf("  %s %s%s\n", status, name, defaultMark)

		if entry.Username != "" {
			ui.Dimmed(fmt.Sprintf("    Logged in as %s", entry.Username))
		} else {
			ui.Dimmed("    Not authenticated")
		}

		if !entry.CreatedAt.IsZero() {
			ui.Dimmed(fmt.Sprintf("    Added %s", formatTimeAgo(entry.CreatedAt)))
		}
		fmt.Println()
	}

	return nil
}

// runRegistryAdd handles muxi registry add
func runRegistryAdd(cmd *cobra.Command, args []string) error {
	ui.InfoBanner("Add Registry")

	ui.Dimmed("  Enter the registry URL (e.g., registry.example.com)")
	url, err := wizard.PromptString("Registry URL", "", nil)
	if err != nil {
		return err
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("registry URL is required")
	}

	// Clean up URL
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")

	// Check if already exists
	config, err := registry.LoadRegistries()
	if err != nil {
		return fmt.Errorf("failed to load registries: %w", err)
	}

	if _, exists := config.Registries[url]; exists {
		fmt.Println()
		ui.Dimmed(fmt.Sprintf("  Registry %s already configured", url))
		return nil
	}

	// Add the registry
	if err := registry.SetToken(url, "", ""); err != nil {
		return fmt.Errorf("failed to add registry: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Added registry: %s", url))

	// Trigger login flow for the new registry
	fmt.Println()
	client, err := registry.NewClient(url)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Try browser callback first
	token, username, err := tryBrowserAuth(client)
	if err != nil {
		// Fallback to manual paste
		token, err = manualAuth(client)
		if err != nil {
			return err
		}
		username = ""
	}

	// If no username from callback, try to validate token
	if username == "" {
		fmt.Println()
		fmt.Print("  Validating token... ")
		user, _ := client.ValidateToken(token)
		if user != nil {
			username = user.Username
		}
		fmt.Println("OK")
	}

	// Save token
	if err := registry.SetToken(url, token, username); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println()
	if username != "" {
		ui.Success(fmt.Sprintf("Logged in as %s", username))
	} else {
		ui.Success("Logged in successfully")
	}

	return nil
}

// runRegistryRemove handles muxi registry remove
func runRegistryRemove(cmd *cobra.Command, args []string) error {
	ui.InfoBanner("Remove Registry")

	config, err := registry.LoadRegistries()
	if err != nil {
		return fmt.Errorf("failed to load registries: %w", err)
	}

	if len(config.Registries) == 0 {
		fmt.Println()
		ui.Dimmed("  No registries configured")
		return nil
	}

	// Build options
	var options []wizard.SelectOption
	for name := range config.Registries {
		label := name
		if name == config.DefaultRegistry {
			label += " (default)"
		}
		options = append(options, wizard.SelectOption{
			Value: name,
			Label: label,
		})
	}

	fmt.Println()
	selected, err := wizard.PromptSelect("Select registry to remove", options, 0)
	if err != nil {
		return err
	}

	// Confirm removal
	fmt.Println()
	fmt.Printf("  Remove %s? (y/N): ", selected)
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		ui.Dimmed("  Cancelled")
		return nil
	}

	// Remove the registry
	if err := registry.RemoveToken(selected); err != nil {
		return fmt.Errorf("failed to remove registry: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Removed registry: %s", selected))

	return nil
}

// runRegistryDefault handles muxi registry default
func runRegistryDefault(cmd *cobra.Command, args []string) error {
	ui.InfoBanner("Set Default Registry")

	config, err := registry.LoadRegistries()
	if err != nil {
		return fmt.Errorf("failed to load registries: %w", err)
	}

	if len(config.Registries) == 0 {
		fmt.Println()
		ui.Dimmed("  No registries configured")
		fmt.Println()
		fmt.Println("  Add a registry first: muxi registry add")
		return nil
	}

	// Build options with current default marked
	var options []wizard.SelectOption
	currentIndex := 0
	i := 0
	for name := range config.Registries {
		label := name
		if name == config.DefaultRegistry {
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
	selected, err := wizard.PromptSelect("Select default registry", options, currentIndex)
	if err != nil {
		return err
	}

	// Update default
	config.DefaultRegistry = selected
	if err := registry.SaveRegistries(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Default registry set to: %s", selected))

	return nil
}
