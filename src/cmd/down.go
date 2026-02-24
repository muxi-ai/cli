package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:     "down [formation-id]",
	Short:   "Stop formation running in local development mode",
	GroupID: "formation",
	Long: `Stop a formation that was started with 'muxi up'.

If no formation ID is provided, reads it from formation.afs in 
the current directory.

The server continues running after the formation stops.
To stop the server: muxi-server stop`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDown,
}

func init() {
	rootCmd.AddCommand(downCmd)
	downCmd.Flags().Int("port", 7890, "Local server port")
}

// DevStopRequest is the request body for POST /rpc/dev/stop
type DevStopRequest struct {
	FormationID string `json:"formation_id"`
}

// DevStopResponse is the response from POST /rpc/dev/stop
type DevStopResponse struct {
	Success     bool   `json:"success"`
	FormationID string `json:"formation_id"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

func runDown(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetInt("port")

	// Get formation ID
	var formationID string
	if len(args) > 0 {
		formationID = args[0]
	} else {
		// Try to read from formation.afs in current directory
		ctx, err := context.DetectFormation()
		if err != nil {
			ui.ErrorBlock(
				"Cannot determine formation ID",
				"No formation ID provided and not in a formation directory.",
				"muxi down my-formation   # specify formation ID\ncd my-formation && muxi down   # or run from formation dir",
			)
			os.Exit(1)
		}
		formationID = ctx.ID
		if formationID == "" {
			formationID = filepath.Base(ctx.RootDir)
		}
	}

	// Check if server is running
	serverAddr := fmt.Sprintf("localhost:%d", port)
	if !isLocalServerRunning(serverAddr) {
		ui.ErrorBlock(
			"Server not running",
			fmt.Sprintf("muxi-server is not running on %s", serverAddr),
			"Nothing to stop - server is not running.",
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

	// Override base URL with local server
	client.BaseURL = fmt.Sprintf("http://%s", serverAddr)

	// POST to /rpc/dev/stop
	fmt.Println()
	spinner := ui.NewSpinner(fmt.Sprintf("Stopping %s...", formationID))
	spinner.Start()

	reqBody := DevStopRequest{FormationID: formationID}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		spinner.StopWithError("Failed to create request")
		return err
	}

	resp, err := client.Post("/rpc/dev/stop", bytes.NewReader(jsonBody), "application/json")
	if err != nil {
		spinner.StopWithError("Failed to connect to server")
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	var result DevStopResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		spinner.StopWithError("Failed to parse response")
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		spinner.StopWithError("Failed to stop")
		fmt.Println()
		ui.Error(result.Error)
		fmt.Println()
		os.Exit(1)
	}

	spinner.StopWithSuccess(fmt.Sprintf("Stopped %s", result.FormationID))

	// Clear draft mode in .muxi if set
	if ctx, err := context.DetectFormation(); err == nil {
		dotMuxi, _ := formation.LoadDotMuxi(ctx.RootDir)
		if dotMuxi.Draft {
			dotMuxi.Draft = false
			if err := formation.SaveDotMuxi(ctx.RootDir, dotMuxi); err == nil {
				ui.Dimmed("  Draft mode disabled")
			}
		}
	}

	// Success message
	fmt.Println()
	ui.Success("Formation stopped")
	fmt.Println()
	fmt.Printf("  Server is still running. To stop it:\n")
	fmt.Printf("    %s\n", ui.CyanText("muxi-server stop"))
	fmt.Println()

	return nil
}
