package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/timer"
)

var forceFlag bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start <description>",
	Short: "Start a timer for a task",
	Long: `Start a timer for a task with the given description.
The timer will run until you stop it with 'did stop'.

The description can include @project and #tags for categorization.
Timer state persists across terminal sessions.

Examples:
  did start fixing authentication bug
  did start code review @acme
  did start API work @client #backend #api`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startTimer(args)
	},
}

func init() {
	startCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "override existing timer if one is already running")
}

// startTimer starts a new timer with the given description
func startTimer(args []string) {
	// Join all arguments to form the description
	description := strings.Join(args, " ")

	// Trim whitespace
	description = strings.TrimSpace(description)

	// Check that description is not empty
	if description == "" {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Description cannot be empty")
		_, _ = fmt.Fprintln(deps.Stderr, "Usage: did start <description>")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: did start fixing authentication bug")
		deps.Exit(1)
		return
	}

	// Parse project and tags from description
	cleanDesc, project, tags := entry.ParseProjectAndTags(description)

	// Check that cleaned description is not empty (in case it was only @project/#tags)
	if cleanDesc == "" {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Description cannot be empty (only project/tags provided)")
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Include a description along with @project and #tags")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: did start fixing bug @acme #urgent")
		deps.Exit(1)
		return
	}

	// Get timer path
	timerPath, err := deps.TimerPath()
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to determine timer location")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Check that your home directory is accessible")
		deps.Exit(1)
		return
	}

	// Check if timer is already running
	isRunning, err := timer.IsTimerRunning(timerPath)
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to check timer status")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return
	}

	if isRunning && !forceFlag {
		existingTimer, err := loadExistingTimerState(timerPath)
		if err != nil {
			return
		}

		_, _ = fmt.Fprintln(deps.Stderr, "Warning: A timer is already running")
		_, _ = fmt.Fprintf(deps.Stderr, "Current timer: %s\n", formatEntryForLog(existingTimer.Description, existingTimer.Project, existingTimer.Tags))
		elapsed := time.Since(existingTimer.StartedAt)
		_, _ = fmt.Fprintf(deps.Stderr, "Started: %s ago\n", formatElapsedTime(elapsed))
		_, _ = fmt.Fprintln(deps.Stderr)
		_, _ = fmt.Fprintln(deps.Stderr, "Options:")
		_, _ = fmt.Fprintln(deps.Stderr, "  - Stop the current timer with 'did stop'")
		_, _ = fmt.Fprintln(deps.Stderr, "  - Override with 'did start <description> --force'")
		deps.Exit(1)
		return
	}

	// Create timer state
	state := timer.TimerState{
		StartedAt:   time.Now(),
		Description: cleanDesc,
		Project:     project,
		Tags:        tags,
	}

	// Save timer state
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to save timer state")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintf(deps.Stderr, "Hint: Check that directory is writable: %s\n", timerPath)
		deps.Exit(1)
		return
	}

	// Display success message
	_, _ = fmt.Fprintf(deps.Stdout, "Timer started: %s\n", formatEntryForLog(cleanDesc, project, tags))
	if forceFlag && isRunning {
		_, _ = fmt.Fprintln(deps.Stdout, "(Previous timer was overwritten)")
	}
}

// formatElapsedTime formats a duration as human-readable elapsed time
// Examples: "5m", "1h 23m", "2h"
func formatElapsedTime(d time.Duration) string {
	totalMinutes := int(d.Minutes())
	if totalMinutes < 60 {
		return fmt.Sprintf("%dm", totalMinutes)
	}
	hours := totalMinutes / 60
	mins := totalMinutes % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}
