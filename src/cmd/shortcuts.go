package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"

	"github.com/spf13/cobra"
)

// Shortcut commands - work from inside a formation directory without specifying ID

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current formation details (shortcut for 'formation get')",
	Long: `Get details of the current formation.

Must be run from inside a formation directory.
This is a shortcut for 'muxi formation get <id>'.`,
	Args: cobra.NoArgs,
	RunE: runShortcutGet,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop current formation (shortcut for 'formation stop')",
	Long: `Stop the current formation.

Must be run from inside a formation directory.
This is a shortcut for 'muxi formation stop <id>'.`,
	Args: cobra.NoArgs,
	RunE: runShortcutStop,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start current formation (shortcut for 'formation start')",
	Long: `Start the current formation (must be stopped).

Must be run from inside a formation directory.
This is a shortcut for 'muxi formation start <id>'.`,
	Args: cobra.NoArgs,
	RunE: runShortcutStart,
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart current formation (shortcut for 'formation restart')",
	Long: `Restart the current formation.

Must be run from inside a formation directory.
This is a shortcut for 'muxi formation restart <id>'.`,
	Args: cobra.NoArgs,
	RunE: runShortcutRestart,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete current formation (shortcut for 'formation delete')",
	Long: `Delete the current formation from the server.

Must be run from inside a formation directory.
This is a shortcut for 'muxi formation delete <id>'.`,
	Args: cobra.NoArgs,
	RunE: runShortcutDelete,
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback current formation (shortcut for 'formation rollback')",
	Long: `Rollback the current formation to its previous version.

Must be run from inside a formation directory.
This is a shortcut for 'muxi formation rollback <id>'.`,
	Args: cobra.NoArgs,
	RunE: runShortcutRollback,
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View current formation logs (shortcut for 'formation logs')",
	Long: `View logs for the current formation.

Must be run from inside a formation directory.
This is a shortcut for 'muxi formation logs <id>'.`,
	Args: cobra.NoArgs,
	RunE: runShortcutLogs,
}

func init() {
	// Register shortcut commands
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(logsCmd)

	// Flags for shortcuts (mirror formation command flags)
	getCmd.Flags().String("profile", "", "Server profile to use")
	getCmd.Flags().BoolP("verbose", "v", false, "Show internal details (port, pid)")

	stopCmd.Flags().String("profile", "", "Server profile to use")
	stopCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	startCmd.Flags().String("profile", "", "Server profile to use")

	restartCmd.Flags().String("profile", "", "Server profile to use")
	restartCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	deleteCmd.Flags().String("profile", "", "Server profile to use")
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	rollbackCmd.Flags().String("profile", "", "Server profile to use")
	rollbackCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	logsCmd.Flags().String("profile", "", "Server profile to use")
	logsCmd.Flags().IntP("lines", "n", 100, "Number of lines to show")
	logsCmd.Flags().String("stream", "", "Filter by stream (stdout, stderr)")
	logsCmd.Flags().BoolP("follow", "f", false, "Stream new logs (like tail -f)")
}

// requireFormationContext returns the formation context or exits with error
func requireFormationContext(command string) (*context.FormationContext, error) {
	ctx, err := context.DetectFormation()
	if err != nil {
		return nil, fmt.Errorf("%s not in a formation directory\n\nRun this command from inside a formation directory, or use:\n  %s <formation-id>", ui.RedText("✗"), ui.CyanText("muxi formation "+command))
	}
	if ctx.ID == "" {
		return nil, fmt.Errorf("%s formation.yaml is missing 'id' field", ui.RedText("✗"))
	}
	return ctx, nil
}

// confirmAction prompts for confirmation, returns true if confirmed
func confirmAction(action, formationID string) bool {
	fmt.Printf("\n  %s formation '%s'? [y/N]: ", action, formationID)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func runShortcutGet(cmd *cobra.Command, args []string) error {
	ctx, err := requireFormationContext("get")
	if err != nil {
		return err
	}

	// Reuse the formation get command logic
	profile, _ := cmd.Flags().GetString("profile")
	verbose, _ := cmd.Flags().GetBool("verbose")

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	f, err := client.GetFormation(ctx.ID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not deployed",
				fmt.Sprintf("Formation '%s' is not deployed to the server.", ctx.ID),
				ui.Command("muxi deploy"),
			)
			os.Exit(1)
		}
		return err
	}

	// Display (reuse logic from formation.go)
	return displayFormationDetails(f, verbose, profile)
}

func runShortcutStop(cmd *cobra.Command, args []string) error {
	ctx, err := requireFormationContext("stop")
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force && !confirmAction("Stop", ctx.ID) {
		fmt.Println("  Cancelled.")
		return nil
	}

	profile, _ := cmd.Flags().GetString("profile")
	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Check if formation exists and status
	f, err := client.GetFormation(ctx.ID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not deployed",
				fmt.Sprintf("Formation '%s' is not deployed to the server.", ctx.ID),
				ui.Command("muxi deploy"),
			)
			os.Exit(1)
		}
		return err
	}

	if f.Status == "stopped" {
		fmt.Println()
		ui.Warning(fmt.Sprintf("Formation '%s' is already stopped", ctx.ID))
		fmt.Println()
		return nil
	}

	spinner := ui.NewSpinner("Stopping formation...")
	spinner.Start()

	err = client.StopFormation(ctx.ID)
	if err != nil {
		spinner.StopWithError("Stop failed")
		return err
	}

	spinner.StopWithSuccess("Stopped formation")

	fmt.Println()
	ui.Success(fmt.Sprintf("Stopped %s", ctx.ID))
	fmt.Println()

	return nil
}

func runShortcutStart(cmd *cobra.Command, args []string) error {
	ctx, err := requireFormationContext("start")
	if err != nil {
		return err
	}

	profile, _ := cmd.Flags().GetString("profile")

	return runFormationStartWithID(ctx.ID, profile)
}

func runShortcutRestart(cmd *cobra.Command, args []string) error {
	ctx, err := requireFormationContext("restart")
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force && !confirmAction("Restart", ctx.ID) {
		fmt.Println("  Cancelled.")
		return nil
	}

	profile, _ := cmd.Flags().GetString("profile")

	// Delegate to existing restart logic with streaming
	return runFormationRestartWithID(ctx.ID, profile)
}

func runShortcutDelete(cmd *cobra.Command, args []string) error {
	ctx, err := requireFormationContext("delete")
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force && !confirmAction("Delete", ctx.ID) {
		fmt.Println("  Cancelled.")
		return nil
	}

	profile, _ := cmd.Flags().GetString("profile")
	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Check if formation exists
	_, err = client.GetFormation(ctx.ID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not deployed",
				fmt.Sprintf("Formation '%s' is not deployed to the server.", ctx.ID),
				"",
			)
			os.Exit(1)
		}
		return err
	}

	spinner := ui.NewSpinner("Deleting formation...")
	spinner.Start()

	err = client.DeleteFormation(ctx.ID)
	if err != nil {
		spinner.StopWithError("Delete failed")
		return err
	}

	spinner.StopWithSuccess("Deleted formation")

	fmt.Println()
	ui.Success(fmt.Sprintf("Deleted %s", ctx.ID))
	fmt.Println()

	return nil
}

func runShortcutRollback(cmd *cobra.Command, args []string) error {
	ctx, err := requireFormationContext("rollback")
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force && !confirmAction("Rollback", ctx.ID) {
		fmt.Println("  Cancelled.")
		return nil
	}

	profile, _ := cmd.Flags().GetString("profile")

	return runFormationRollbackWithID(ctx.ID, profile)
}

func runShortcutLogs(cmd *cobra.Command, args []string) error {
	ctx, err := requireFormationContext("logs")
	if err != nil {
		return err
	}

	profile, _ := cmd.Flags().GetString("profile")
	lines, _ := cmd.Flags().GetInt("lines")
	stream, _ := cmd.Flags().GetString("stream")
	follow, _ := cmd.Flags().GetBool("follow")

	if stream == "" {
		stream = "all"
	}

	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	// Check if formation exists
	_, err = client.GetFormation(ctx.ID)
	if err != nil {
		if err.Error() == "not found" {
			ui.ErrorBlock(
				"Formation not deployed",
				fmt.Sprintf("Formation '%s' is not deployed to the server.", ctx.ID),
				ui.Command("muxi deploy"),
			)
			os.Exit(1)
		}
		return err
	}

	// Follow mode
	if follow {
		return streamFormationLogs(client, ctx.ID, stream)
	}

	// Non-follow mode
	resp, err := client.GetFormationLogs(ctx.ID, lines, stream)
	if err != nil {
		return err
	}

	var allLogs []string
	if stream == "all" || stream == "stdout" {
		allLogs = append(allLogs, resp.Logs.Stdout...)
	}
	if stream == "all" || stream == "stderr" {
		allLogs = append(allLogs, resp.Logs.Stderr...)
	}

	if len(allLogs) == 0 {
		fmt.Println()
		ui.Dimmed("  No logs available")
		fmt.Println()
		return nil
	}

	for _, line := range allLogs {
		fmt.Println(line)
	}

	return nil
}
