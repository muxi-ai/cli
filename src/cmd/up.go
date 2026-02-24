package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:     "up",
	Short:   "Start formation in local development mode",
	GroupID: "formation",
	Long: `Start a formation directly from source directory for local development.

Run this command from inside a formation directory. The formation runs
without the full deploy cycle (no compression, upload, or staging).

The server must be running locally. If not, start it with:
  muxi-server start

The formation will be accessible at:
  http://localhost:{server-port}/draft/{formation-id}`,
	Args: cobra.NoArgs,
	RunE: runUp,
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().Int("port", 7890, "Local server port")
}

// DevRunRequest is the request body for POST /rpc/dev/run
type DevRunRequest struct {
	Path string `json:"path"`
}

// DevRunResponse is the response from POST /rpc/dev/run
type DevRunResponse struct {
	Success     bool   `json:"success"`
	FormationID string `json:"formation_id"`
	Port        int    `json:"port"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

func runUp(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetInt("port")

	// Must be in a formation directory
	ctx, err := context.DetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in a formation directory",
			"Run this command from inside a formation directory.",
			"cd my-formation && muxi up",
		)
		os.Exit(1)
	}

	absPath := ctx.RootDir
	formationID := ctx.ID
	if formationID == "" {
		formationID = filepath.Base(absPath)
	}

	// Check if server is running
	serverAddr := fmt.Sprintf("localhost:%d", port)
	if !isLocalServerRunning(serverAddr) {
		ui.ErrorBlock(
			"Server not running",
			fmt.Sprintf("muxi-server is not running on %s", serverAddr),
			"muxi-server start        # in another terminal\nmuxi-server start &      # run in background",
		)
		os.Exit(1)
	}

	// Get server client for HMAC auth
	client, err := server.NewClient("")
	if err != nil {
		ui.ErrorBlock(
			"No server profile configured",
			"Run 'muxi server add' to configure a server profile.",
			"muxi server add",
		)
		os.Exit(1)
	}

	// Override base URL and increase timeout for dev startup (MCP servers can take a while)
	client.BaseURL = fmt.Sprintf("http://%s", serverAddr)
	client.HTTPClient.Timeout = 5 * time.Minute

	// POST to /rpc/dev/run
	fmt.Println()
	spinner := ui.NewSpinner(fmt.Sprintf("Starting %s...", formationID))
	spinner.Start()

	reqBody := DevRunRequest{Path: absPath}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		spinner.StopWithError("Failed to create request")
		return err
	}

	resp, err := client.Post("/rpc/dev/run", bytes.NewReader(jsonBody), "application/json")
	if err != nil {
		spinner.StopWithError("Failed to connect to server")
		fmt.Println()
		ui.Error(fmt.Sprintf("Could not reach muxi-server at %s", serverAddr))
		ui.Dimmed("  The server may be busy starting MCP servers or processing another request.")
		ui.Dimmed("  Check the server logs for details.")
		fmt.Println()
		return nil
	}
	defer resp.Body.Close()

	// Read raw body for better error handling
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		spinner.StopWithError("Failed to read response")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-success HTTP status codes
	if resp.StatusCode != http.StatusOK {
		spinner.StopWithError("Failed to start")
		fmt.Println()

		// Try to parse structured error
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
			Detail  string `json:"detail"`
		}
		if json.Unmarshal(rawBody, &errResp) == nil {
			msg := errResp.Error
			if msg == "" {
				msg = errResp.Message
			}
			if msg == "" {
				msg = errResp.Detail
			}
			if msg != "" {
				ui.Error(msg)
				fmt.Println()
				return nil
			}
		}

		ui.Error(fmt.Sprintf("Server returned %d: %s", resp.StatusCode, string(rawBody)))
		fmt.Println()
		return nil
	}

	var result DevRunResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		spinner.StopWithError("Failed to parse response")
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		spinner.StopWithError("Failed to start")
		fmt.Println()
		ui.Error(result.Error)
		fmt.Println()
		os.Exit(1)
	}

	spinner.StopWithSuccess(fmt.Sprintf("Started %s", result.FormationID))

	// Success message
	fmt.Println()
	ui.Success(fmt.Sprintf("Formation running on port %d", result.Port))
	fmt.Println()
	fmt.Printf("  Draft URL: http://%s/draft/%s\n", serverAddr, result.FormationID)
	fmt.Println()
	fmt.Printf("  To stop:   %s\n", ui.CyanText("muxi down"))
	fmt.Println()

	return nil
}

// isLocalServerRunning checks if the local muxi-server is running
func isLocalServerRunning(addr string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/health", addr))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
