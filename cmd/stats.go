package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/stats"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show summary statistics for time entries",
	Long: `Show aggregated statistics for your time tracking entries.

Display summary statistics including:
  - Total hours logged
  - Average daily hours
  - Number of entries
  - Breakdown by project and tag (when available)
  - Comparison to previous period

By default, statistics are shown for the current week (Monday-Sunday).
Use the --month flag to show statistics for the current month instead.

Examples:

  Default (current week):
    did stats                          Show statistics for this week

  Monthly statistics:
    did stats --month                  Show statistics for this month

The stats command provides insights into your productivity patterns and
time distribution, helping you understand where your time goes.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStats(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)

	// Add --month flag to switch from week to month view
	statsCmd.Flags().Bool("month", false, "Show statistics for current month instead of week")
}

// runStats handles the stats command logic
func runStats(cmd *cobra.Command, args []string) {
	// Get flag values
	showMonth, _ := cmd.Flags().GetBool("month")

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

	// Filter out soft-deleted entries
	var activeEntries []entry.Entry
	for _, e := range result.Entries {
		if e.DeletedAt == nil {
			activeEntries = append(activeEntries, e)
		}
	}

	// Determine the time period based on --month flag
	var start, end time.Time
	var prevStart, prevEnd time.Time
	var periodName string
	var comparisonPeriod string
	if showMonth {
		start, end = timeutil.ThisMonth()
		prevStart, prevEnd = timeutil.LastMonth()
		periodName = "this month"
		comparisonPeriod = "month"
	} else {
		// Use configured week_start_day for weekly statistics
		now := time.Now()
		start = timeutil.StartOfWeekWithConfig(now, deps.Config.WeekStartDay)
		end = timeutil.EndOfWeekWithConfig(now, deps.Config.WeekStartDay)

		lastWeek := now.AddDate(0, 0, -7)
		prevStart = timeutil.StartOfWeekWithConfig(lastWeek, deps.Config.WeekStartDay)
		prevEnd = timeutil.EndOfWeekWithConfig(lastWeek, deps.Config.WeekStartDay)

		periodName = "this week"
		comparisonPeriod = "week"
	}

	// Calculate statistics for current period
	statistics := stats.CalculateStatistics(activeEntries, start, end)

	// Calculate statistics for previous period for comparison
	previousStatistics := stats.CalculateStatistics(activeEntries, prevStart, prevEnd)

	// Display header
	_, _ = fmt.Fprintf(deps.Stdout, "Statistics for %s\n", periodName)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 60))
	_, _ = fmt.Fprintln(deps.Stdout)

	// Display statistics
	displayStatistics(statistics)

	// Display comparison to previous period
	diffMinutes := stats.CompareStatistics(statistics, previousStatistics)
	comparison := stats.FormatComparison(diffMinutes, comparisonPeriod)
	_, _ = fmt.Fprintf(deps.Stdout, "Comparison:      %s\n", comparison)
	_, _ = fmt.Fprintln(deps.Stdout)

	// Calculate and display project breakdown if projects exist
	projectBreakdown := stats.CalculateProjectBreakdown(activeEntries, start, end)
	if len(projectBreakdown) > 0 {
		displayProjectBreakdown(projectBreakdown)
	}

	// Calculate and display tag breakdown if tags exist
	tagBreakdown := stats.CalculateTagBreakdown(activeEntries, start, end)
	if len(tagBreakdown) > 0 {
		displayTagBreakdown(tagBreakdown)
	}
}

// displayStatistics formats and displays statistics to stdout
func displayStatistics(stats stats.Statistics) {
	// Display total hours
	_, _ = fmt.Fprintf(deps.Stdout, "Total Hours:     %s\n", formatDuration(stats.TotalMinutes))

	// Display average daily hours
	avgHours := stats.AverageMinutesPerDay / 60.0
	_, _ = fmt.Fprintf(deps.Stdout, "Average/Day:     %.1fh\n", avgHours)

	// Display entry count
	_, _ = fmt.Fprintf(deps.Stdout, "Entries:         %d %s\n", stats.EntryCount, pluralize("entry", stats.EntryCount))

	// Display days with entries (useful context)
	_, _ = fmt.Fprintf(deps.Stdout, "Days Tracked:    %d %s\n", stats.DaysWithEntries, pluralize("day", stats.DaysWithEntries))

	_, _ = fmt.Fprintln(deps.Stdout)
}

// displayProjectBreakdown formats and displays project breakdown to stdout
func displayProjectBreakdown(breakdowns []stats.ProjectBreakdown) {
	_, _ = fmt.Fprintln(deps.Stdout, "By Project:")
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 60))
	_, _ = fmt.Fprintln(deps.Stdout)

	for _, breakdown := range breakdowns {
		// Format project name with special handling for "(no project)"
		projectDisplay := breakdown.Project
		if breakdown.Project != "(no project)" {
			projectDisplay = "@" + breakdown.Project
		}

		_, _ = fmt.Fprintf(deps.Stdout, "  %-28s  %10s  (%d %s)\n",
			projectDisplay,
			formatDuration(breakdown.TotalMinutes),
			breakdown.EntryCount,
			pluralize("entry", breakdown.EntryCount))
	}

	_, _ = fmt.Fprintln(deps.Stdout)
}

// displayTagBreakdown formats and displays tag breakdown to stdout
func displayTagBreakdown(breakdowns []stats.TagBreakdown) {
	_, _ = fmt.Fprintln(deps.Stdout, "By Tag:")
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 60))
	_, _ = fmt.Fprintln(deps.Stdout)

	for _, breakdown := range breakdowns {
		// Format tag name with special handling for "(no tags)"
		tagDisplay := breakdown.Tag
		if breakdown.Tag != "(no tags)" {
			tagDisplay = "#" + breakdown.Tag
		}

		_, _ = fmt.Fprintf(deps.Stdout, "  %-28s  %10s  (%d %s)\n",
			tagDisplay,
			formatDuration(breakdown.TotalMinutes),
			breakdown.EntryCount,
			pluralize("entry", breakdown.EntryCount))
	}

	_, _ = fmt.Fprintln(deps.Stdout)
}
