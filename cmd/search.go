package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/storage"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search for entries by keyword",
	Long: `Search for time tracking entries containing a specific keyword in their description.

The search is case-insensitive and searches across all entries in the storage file.

Examples:
  did search meeting          Search for entries containing 'meeting'
  did search bug              Search for entries containing 'bug'
  did search "code review"    Search for entries containing 'code review'`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchEntries(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

// searchEntries handles the search command logic
func searchEntries(cmd *cobra.Command, args []string) {
	// Get the keyword from arguments
	keyword := args[0]

	// Get storage path
	storagePath, err := deps.StoragePath()
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to determine storage location")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Check that your home directory is accessible")
		deps.Exit(1)
		return
	}

	// Read all entries from storage
	result, err := storage.ReadEntriesWithWarnings(storagePath)
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to read entries from storage")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintf(deps.Stderr, "Hint: Check that file exists and is readable: %s\n", storagePath)
		deps.Exit(1)
		return
	}

	// Display warnings about corrupted lines to stderr
	if len(result.Warnings) > 0 {
		_, _ = fmt.Fprintf(deps.Stderr, "Warning: Found %d corrupted line(s) in storage file:\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			_, _ = fmt.Fprintln(deps.Stderr, formatCorruptionWarning(warning))
		}
		_, _ = fmt.Fprintln(deps.Stderr)
	}

	// Create filter with keyword
	f := filter.NewFilter(keyword, "", nil)

	// Filter entries by keyword
	filtered := filter.FilterEntries(result.Entries, f)

	// Check if any results found
	if len(filtered) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No entries found matching '%s'\n", keyword)
		return
	}

	// Calculate total duration
	totalMinutes := 0
	for _, e := range filtered {
		totalMinutes += e.DurationMinutes
	}

	// Display results
	_, _ = fmt.Fprintf(deps.Stdout, "Search results for '%s':\n", keyword)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))

	// Calculate width for right-aligned indices
	maxIndexWidth := len(fmt.Sprintf("%d", len(filtered)))

	for i, e := range filtered {
		_, _ = fmt.Fprintf(deps.Stdout, "[%*d] %s %s  %s (%s)\n",
			maxIndexWidth,
			i+1, // 1-based index for user reference
			e.Timestamp.Format("2006-01-02"),
			e.Timestamp.Format("15:04"),
			formatEntryForLog(e.Description, e.Project, e.Tags),
			formatDuration(e.DurationMinutes))
	}
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "Total: %s (%d %s)\n", formatDuration(totalMinutes), len(filtered), pluralize("entry", len(filtered)))
}
