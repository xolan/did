package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/timer"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the current timer",
	Long: `Show the status of the current timer if one is running.

Displays the timer description, elapsed time, start time, and any project/tags.
If no timer is running, displays a message indicating that.

Examples:
  did status`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}

// showStatus displays the current timer status
func showStatus() {
	// Get timer path
	timerPath, err := timer.GetTimerPath()
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to determine timer location")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Check that your home directory is accessible")
		deps.Exit(1)
		return
	}

	// Load timer state
	state, err := timer.LoadTimerState(timerPath)
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to load timer state")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return
	}

	// Check if timer is running
	if state == nil {
		_, _ = fmt.Fprintln(deps.Stdout, "No timer running")
		_, _ = fmt.Fprintln(deps.Stdout, "Start a timer with: did start <description>")
		return
	}

	// Calculate elapsed time
	elapsed := time.Since(state.StartedAt)
	elapsedFormatted := formatElapsedTime(elapsed)

	// Format start time in human-readable format
	startTime := state.StartedAt.Format("3:04 PM")
	startDate := state.StartedAt.Format("Mon Jan 2")

	// Check if started today
	now := time.Now()
	isToday := state.StartedAt.Year() == now.Year() &&
		state.StartedAt.Month() == now.Month() &&
		state.StartedAt.Day() == now.Day()

	var startTimeFormatted string
	if isToday {
		startTimeFormatted = fmt.Sprintf("today at %s", startTime)
	} else {
		startTimeFormatted = fmt.Sprintf("%s at %s", startDate, startTime)
	}

	// Display timer status
	formattedDesc := formatEntryForLog(state.Description, state.Project, state.Tags)
	_, _ = fmt.Fprintln(deps.Stdout, "Timer running:")
	_, _ = fmt.Fprintf(deps.Stdout, "  %s\n", formattedDesc)
	_, _ = fmt.Fprintf(deps.Stdout, "  Started: %s\n", startTimeFormatted)
	_, _ = fmt.Fprintf(deps.Stdout, "  Elapsed: %s\n", elapsedFormatted)
}
