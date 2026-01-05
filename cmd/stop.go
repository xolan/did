package cmd

import (
	"fmt"
	"math"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timer"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the current timer and create an entry",
	Long: `Stop the currently running timer and create a time tracking entry.

The timer duration is calculated from when the timer was started until now.
Duration is rounded to the nearest minute with a minimum of 1 minute.

Examples:
  did stop`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		stopTimer()
	},
}

// stopTimer stops the current timer and creates an entry
func stopTimer() {
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
		_, _ = fmt.Fprintln(deps.Stderr, "Error: No timer is running")
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Start a timer with 'did start <description>'")
		deps.Exit(1)
		return
	}

	// Calculate duration from StartedAt to now
	elapsed := time.Since(state.StartedAt)
	durationMinutes := calculateDurationMinutes(elapsed)

	// Get storage path
	storagePath, err := deps.StoragePath()
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to get storage path")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return
	}

	// Create entry with the timer data
	now := time.Now()
	e := entry.Entry{
		Timestamp:       now,
		Description:     state.Description,
		DurationMinutes: durationMinutes,
		RawInput:        fmt.Sprintf("%s for %s", state.Description, formatDuration(durationMinutes)),
		Project:         state.Project,
		Tags:            state.Tags,
	}

	// Append entry to storage
	if err := storage.AppendEntry(storagePath, e); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to save entry")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return
	}

	// Clear timer state after successful entry creation
	if err := timer.ClearTimerState(timerPath); err != nil {
		// Entry was saved successfully, but we couldn't clear the timer
		// This is not critical - warn but don't exit with error
		_, _ = fmt.Fprintln(deps.Stderr, "Warning: Entry saved but failed to clear timer state")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
	}

	// Display success message
	formattedDesc := formatEntryForLog(state.Description, state.Project, state.Tags)
	_, _ = fmt.Fprintf(deps.Stdout, "Stopped: %s (%s)\n", formattedDesc, formatDuration(durationMinutes))
}

// calculateDurationMinutes calculates duration in minutes from a time.Duration.
// Rounds to the nearest minute with a minimum of 1 minute.
func calculateDurationMinutes(d time.Duration) int {
	// Convert to minutes and round to nearest minute
	minutes := int(math.Round(d.Minutes()))

	// Minimum duration is 1 minute
	if minutes < 1 {
		return 1
	}

	return minutes
}
