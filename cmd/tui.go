package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/tui"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	Long: `Launch the interactive Terminal User Interface for did.

The TUI provides a full-featured interface for managing your time entries
with keyboard navigation, multiple views, and real-time updates.

Views available:
  - Entries: Browse and manage time entries with filtering
  - Timer: Start, stop, and monitor timers
  - Stats: View weekly and monthly statistics
  - Search: Search entries by keyword
  - Config: View and manage configuration

Keyboard shortcuts:
  - Tab/Shift+Tab: Navigate between views
  - 1-5: Jump to specific view
  - j/k or arrows: Navigate within lists
  - ?: Show help
  - q: Quit`,
	Run: func(cmd *cobra.Command, args []string) {
		runTUI()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	// Add --tui flag to root command for quick access
	rootCmd.PersistentFlags().Bool("tui", false, "Launch interactive terminal UI")
}

// runTUI initializes and runs the TUI application
func runTUI() {
	// Initialize services
	services, err := service.NewServices()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing services: %v\n", err)
		os.Exit(1)
	}

	// Run the TUI
	if err := tui.Run(services); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// CheckTUIFlag checks if the --tui flag is set and runs the TUI if so.
// Returns true if the TUI was launched, false otherwise.
func CheckTUIFlag(cmd *cobra.Command) bool {
	tuiFlag, _ := cmd.Root().PersistentFlags().GetBool("tui")
	if tuiFlag {
		runTUI()
		return true
	}
	return false
}
