package handlers

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xolan/did/internal/cli"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/service"
)

// CreateEntry creates a new time tracking entry from raw input
func CreateEntry(deps *cli.Deps, rawInput string) {
	entry, err := deps.Services.Entry.Create(rawInput)
	if err != nil {
		switch err {
		case service.ErrMissingDuration:
			_, _ = fmt.Fprintln(deps.Stderr, "Error: Invalid format. Missing 'for <duration>'")
			_, _ = fmt.Fprintln(deps.Stderr, "Usage: did <description> for <duration>")
			_, _ = fmt.Fprintln(deps.Stderr, "Example: did feature X for 2h")
		case service.ErrEmptyDescription:
			_, _ = fmt.Fprintln(deps.Stderr, "Error: Description cannot be empty")
		default:
			_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		}
		deps.Exit(1)
		return
	}

	// Display success message
	desc := cli.FormatEntryForLog(entry.Description, entry.Project, entry.Tags)
	_, _ = fmt.Fprintf(deps.Stdout, "Logged: %s (%s)\n", desc, cli.FormatDuration(entry.DurationMinutes))
}

// ListEntries lists entries for the given date range and filter
func ListEntries(deps *cli.Deps, dateRange service.DateRangeSpec, f *filter.Filter) {
	result, err := deps.Services.Entry.List(dateRange, f)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	// Display warnings about corrupted lines
	if len(result.Warnings) > 0 {
		_, _ = fmt.Fprintf(deps.Stderr, "Warning: Found %d corrupted line(s) in storage file:\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			_, _ = fmt.Fprintln(deps.Stderr, cli.FormatCorruptionWarning(warning))
		}
		_, _ = fmt.Fprintln(deps.Stderr)
	}

	// Build period string with filters
	period := result.Period
	if f != nil && !f.IsEmpty() {
		period = cli.BuildPeriodWithFilters(period, f.Project, f.Tags)
	}

	if len(result.Entries) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No entries found for %s\n", period)
		return
	}

	// Display entries
	_, _ = fmt.Fprintf(deps.Stdout, "Entries for %s:\n", period)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))

	// Calculate max index width for alignment
	maxIndex := result.Entries[len(result.Entries)-1].ActiveIndex
	maxIndexWidth := len(fmt.Sprintf("%d", maxIndex))

	// Check if entries span multiple days
	showDate := cli.SpansMultipleDaysIndexed(result.Entries)

	for _, ie := range result.Entries {
		e := ie.Entry
		if showDate {
			_, _ = fmt.Fprintf(deps.Stdout, "[%*d] %s %s  %s (%s)\n",
				maxIndexWidth,
				ie.ActiveIndex,
				e.Timestamp.Format("2006-01-02"),
				e.Timestamp.Format("15:04"),
				cli.FormatEntryForLog(e.Description, e.Project, e.Tags),
				cli.FormatDuration(e.DurationMinutes))
		} else {
			_, _ = fmt.Fprintf(deps.Stdout, "[%*d] %s  %s (%s)\n",
				maxIndexWidth,
				ie.ActiveIndex,
				e.Timestamp.Format("15:04"),
				cli.FormatEntryForLog(e.Description, e.Project, e.Tags),
				cli.FormatDuration(e.DurationMinutes))
		}
	}
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "Total: %s\n", cli.FormatDuration(result.Total))
}

// EditEntry edits an existing entry
func EditEntry(deps *cli.Deps, indexStr, newDescription, newDuration string) {
	userIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid index '%s'. Index must be a number\n", indexStr)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: List entries with 'did' to see available indices")
		deps.Exit(1)
		return
	}

	entry, err := deps.Services.Entry.Edit(userIndex, newDescription, newDuration)
	if err != nil {
		switch err {
		case service.ErrNoChangesSpecified:
			_, _ = fmt.Fprintln(deps.Stderr, "Error: At least one flag (--description or --duration) is required")
			_, _ = fmt.Fprintln(deps.Stderr, "Usage:")
			_, _ = fmt.Fprintln(deps.Stderr, "  did edit <index> --description 'new text'")
			_, _ = fmt.Fprintln(deps.Stderr, "  did edit <index> --duration 2h")
		case service.ErrNoEntries:
			_, _ = fmt.Fprintln(deps.Stderr, "Error: No entries found to edit")
			_, _ = fmt.Fprintln(deps.Stderr, "Hint: Create an entry first with 'did <description> for <duration>'")
		default:
			_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
			_, _ = fmt.Fprintln(deps.Stderr, "Hint: List entries with 'did' to see all indices")
		}
		deps.Exit(1)
		return
	}

	// Display success message
	desc := cli.FormatEntryForLog(entry.Description, entry.Project, entry.Tags)
	_, _ = fmt.Fprintf(deps.Stdout, "Updated entry %d: %s (%s)\n", userIndex, desc, cli.FormatDuration(entry.DurationMinutes))
}

// DeleteEntry deletes an entry with optional confirmation
func DeleteEntry(deps *cli.Deps, indexStr string, skipConfirm bool) {
	userIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid index '%s'. Index must be a number\n", indexStr)
		deps.Exit(1)
		return
	}

	if userIndex < 1 {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Index must be 1 or greater (got %d)\n", userIndex)
		deps.Exit(1)
		return
	}

	// Get the entry first to show it
	ie, err := deps.Services.Entry.GetByIndex(userIndex)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	// Show the entry being deleted
	_, _ = fmt.Fprintln(deps.Stdout, "Entry to delete:")
	_, _ = fmt.Fprintf(deps.Stdout, "  %s  %s (%s)\n",
		ie.Entry.Timestamp.Format("2006-01-02 15:04"),
		ie.Entry.Description,
		cli.FormatDuration(ie.Entry.DurationMinutes))

	// Prompt for confirmation unless --yes flag is set
	if !skipConfirm {
		if !promptConfirmation(deps.Stdout, deps.Stdin) {
			_, _ = fmt.Fprintln(deps.Stdout, "Deletion cancelled")
			return
		}
	}

	// Delete the entry
	deletedEntry, err := deps.Services.Entry.Delete(userIndex)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	// Show success message
	_, _ = fmt.Fprintf(deps.Stdout, "Deleted: %s (%s)\n", deletedEntry.Description, cli.FormatDuration(deletedEntry.DurationMinutes))
	_, _ = fmt.Fprintln(deps.Stdout, "Tip: Use 'did undo' to recover this entry if needed")
}

// RestoreEntry restores the most recently deleted entry
func RestoreEntry(deps *cli.Deps) {
	restoredEntry, err := deps.Services.Entry.Restore()
	if err != nil {
		if err == service.ErrNoDeletedEntries {
			_, _ = fmt.Fprintln(deps.Stderr, "Error: no deleted entries found")
			_, _ = fmt.Fprintln(deps.Stderr, "Hint: No entries to restore. Delete an entry first with 'did delete <index>'")
		} else {
			_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		}
		deps.Exit(1)
		return
	}

	// Show success message with entry details
	desc := cli.FormatEntryForLog(restoredEntry.Description, restoredEntry.Project, restoredEntry.Tags)
	_, _ = fmt.Fprintf(deps.Stdout, "Restored: %s (%s)\n", desc, cli.FormatDuration(restoredEntry.DurationMinutes))
	_, _ = fmt.Fprintf(deps.Stdout, "  Timestamp: %s\n", restoredEntry.Timestamp.Format("2006-01-02 15:04"))
}

// promptConfirmation asks the user to confirm deletion
func promptConfirmation(stdout io.Writer, stdin io.Reader) bool {
	_, _ = fmt.Fprint(stdout, "Delete this entry? [y/N]: ")

	scanner := bufio.NewScanner(stdin)
	if !scanner.Scan() {
		return false
	}

	response := strings.TrimSpace(scanner.Text())
	return response == "y" || response == "Y"
}
