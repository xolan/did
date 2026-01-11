package handlers

import (
	"fmt"

	"github.com/xolan/did/internal/cli"
	"github.com/xolan/did/internal/service"
)

// StartTimer starts a new timer
func StartTimer(deps *cli.Deps, description string, force bool) {
	state, existingTimer, err := deps.Services.Timer.Start(description, force)
	if err != nil {
		if err == service.ErrTimerAlreadyRunning && existingTimer != nil {
			_, _ = fmt.Fprintln(deps.Stderr, "Warning: A timer is already running")
			_, _ = fmt.Fprintf(deps.Stderr, "Current timer: %s\n",
				cli.FormatEntryForLog(existingTimer.Description, existingTimer.Project, existingTimer.Tags))
			_, _ = fmt.Fprintf(deps.Stderr, "Started: %s\n", cli.FormatTimerStartTime(existingTimer.StartedAt))
			_, _ = fmt.Fprintln(deps.Stderr)
			_, _ = fmt.Fprintln(deps.Stderr, "Options:")
			_, _ = fmt.Fprintln(deps.Stderr, "  - Stop the current timer with 'did stop'")
			_, _ = fmt.Fprintln(deps.Stderr, "  - Override with 'did start <description> --force'")
		} else if err == service.ErrEmptyDescription {
			_, _ = fmt.Fprintln(deps.Stderr, "Error: Description cannot be empty")
			_, _ = fmt.Fprintln(deps.Stderr, "Usage: did start <description>")
			_, _ = fmt.Fprintln(deps.Stderr, "Example: did start fixing authentication bug")
		} else {
			_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		}
		deps.Exit(1)
		return
	}

	// Display success message
	desc := cli.FormatEntryForLog(state.Description, state.Project, state.Tags)
	_, _ = fmt.Fprintf(deps.Stdout, "Timer started: %s\n", desc)
	if force && existingTimer != nil {
		_, _ = fmt.Fprintln(deps.Stdout, "(Previous timer was overwritten)")
	}
}

// StopTimer stops the current timer and creates an entry
func StopTimer(deps *cli.Deps) {
	entry, state, err := deps.Services.Timer.Stop()
	if err != nil {
		if err == service.ErrNoTimerRunning {
			_, _ = fmt.Fprintln(deps.Stderr, "Error: No timer is running")
			_, _ = fmt.Fprintln(deps.Stderr, "Hint: Start a timer with 'did start <description>'")
		} else {
			_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		}
		deps.Exit(1)
		return
	}

	// Display success message
	desc := cli.FormatEntryForLog(state.Description, state.Project, state.Tags)
	_, _ = fmt.Fprintf(deps.Stdout, "Stopped: %s (%s)\n", desc, cli.FormatDuration(entry.DurationMinutes))
}

// ShowTimerStatus shows the current timer status
func ShowTimerStatus(deps *cli.Deps) {
	status, err := deps.Services.Timer.Status()
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	if !status.Running || status.State == nil {
		_, _ = fmt.Fprintln(deps.Stdout, "No timer running")
		_, _ = fmt.Fprintln(deps.Stdout, "Start a timer with: did start <description>")
		return
	}

	state := status.State
	desc := cli.FormatEntryForLog(state.Description, state.Project, state.Tags)

	_, _ = fmt.Fprintln(deps.Stdout, "Timer running:")
	_, _ = fmt.Fprintf(deps.Stdout, "  %s\n", desc)
	_, _ = fmt.Fprintf(deps.Stdout, "  Started: %s\n", cli.FormatTimerStartTime(state.StartedAt))
	_, _ = fmt.Fprintf(deps.Stdout, "  Elapsed: %s\n", cli.FormatElapsedTime(status.ElapsedTime))
}
