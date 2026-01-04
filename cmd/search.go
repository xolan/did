package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search for entries by keyword",
	Long: `Search for time tracking entries containing a specific keyword in their description.

The search is case-insensitive and searches across all entries in the storage file.

Date Filtering:
  Use --from and --to to filter by date range
  Use --last to filter by relative days (e.g., 'last 7 days')

Examples:
  did search meeting                      Search for entries containing 'meeting'
  did search bug                          Search for entries containing 'bug'
  did search "code review"                Search for entries containing 'code review'
  did search meeting --from 2024-01-01    Search from a specific date
  did search bug --from 2024-01-01 --to 2024-01-31    Search within date range
  did search review --last 7              Search in the last 7 days`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchEntries(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Date filtering flags
	searchCmd.Flags().String("from", "", "Start date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	searchCmd.Flags().String("to", "", "End date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	searchCmd.Flags().Int("last", 0, "Filter by last N days (e.g., --last 7 for last 7 days)")
}

// searchEntries handles the search command logic
func searchEntries(cmd *cobra.Command, args []string) {
	// Get the keyword from arguments
	keyword := args[0]

	// Parse date filtering flags
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	lastDays, _ := cmd.Flags().GetInt("last")

	// Validate flag combinations
	if lastDays > 0 && (fromStr != "" || toStr != "") {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Cannot use --last with --from or --to")
		_, _ = fmt.Fprintln(deps.Stderr, "Use either --last N or --from/--to, not both")
		deps.Exit(1)
		return
	}

	// Parse date range
	var startDate, endDate time.Time
	var hasDateFilter bool

	if lastDays > 0 {
		// Use relative days
		now := time.Now()
		endDate = timeutil.EndOfDay(now)
		startDate = timeutil.StartOfDay(now.AddDate(0, 0, -(lastDays - 1)))
		hasDateFilter = true
	} else if fromStr != "" || toStr != "" {
		// Use explicit date range
		hasDateFilter = true

		// Parse from date
		if fromStr != "" {
			var err error
			startDate, err = timeutil.ParseDate(fromStr)
			if err != nil {
				_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid --from date: %v\n", err)
				deps.Exit(1)
				return
			}
		} else {
			// No from date: use the beginning of time
			startDate = time.Time{}
		}

		// Parse to date
		if toStr != "" {
			var err error
			toDate, err := timeutil.ParseDate(toStr)
			if err != nil {
				_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid --to date: %v\n", err)
				deps.Exit(1)
				return
			}
			endDate = timeutil.EndOfDay(toDate)
		} else {
			// No to date: use now
			endDate = timeutil.EndOfDay(time.Now())
		}
	}

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

	// Apply date filtering if specified
	if hasDateFilter {
		dateFiltered := make([]entry.Entry, 0)
		for _, e := range filtered {
			if timeutil.IsInRange(e.Timestamp, startDate, endDate) {
				dateFiltered = append(dateFiltered, e)
			}
		}
		filtered = dateFiltered
	}

	// Check if any results found
	if len(filtered) == 0 {
		if hasDateFilter {
			_, _ = fmt.Fprintf(deps.Stdout, "No entries found matching '%s' in the specified date range\n", keyword)
		} else {
			_, _ = fmt.Fprintf(deps.Stdout, "No entries found matching '%s'\n", keyword)
		}
		return
	}

	// Calculate total duration
	totalMinutes := 0
	for _, e := range filtered {
		totalMinutes += e.DurationMinutes
	}

	// Display results
	resultHeader := fmt.Sprintf("Search results for '%s'", keyword)
	if hasDateFilter {
		if lastDays > 0 {
			resultHeader += fmt.Sprintf(" (last %d %s)", lastDays, pluralize("day", lastDays))
		} else {
			resultHeader += fmt.Sprintf(" (%s)", formatDateRangeForDisplay(startDate, endDate))
		}
	}
	_, _ = fmt.Fprintf(deps.Stdout, "%s:\n", resultHeader)
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
