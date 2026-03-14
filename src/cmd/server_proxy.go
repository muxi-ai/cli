package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var serverProxyCmd = &cobra.Command{
	Use:                "server",
	Short:              "Proxy to muxi-server (the server daemon)",
	Long:               "Runs muxi-server with the provided arguments. Use 'muxi remote' to manage deployed formations.",
	DisableFlagParsing: true,
	RunE:               runServerProxy,
}

func init() {
	rootCmd.AddCommand(serverProxyCmd)
}

func runServerProxy(cmd *cobra.Command, args []string) error {
	binary, err := exec.LookPath("muxi-server")
	if err != nil {
		fmt.Println()
		ui.ErrorBlock(
			"muxi-server not found",
			"The muxi-server binary is not installed or not in your PATH.",
			"Install muxi-server: pip install muxi-server\n\n  To manage deployed formations, use:\n    muxi remote [command]",
		)
		os.Exit(1)
	}

	proc := exec.Command(binary, args...)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	// Forward signals to the child process
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := proc.Start(); err != nil {
		return fmt.Errorf("failed to start muxi-server: %w", err)
	}

	go func() {
		sig := <-sigChan
		if proc.Process != nil {
			proc.Process.Signal(sig)
		}
	}()

	if err := proc.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}
