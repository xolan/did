package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report [@project | #tag]",
	Short: "Generate detailed reports showing time spent grouped by project or tag",
	Long: `Generate detailed reports showing time spent grouped by project or tag.
Useful for client billing, team stand-ups, or personal review.

Report Types:

  Single Project/Tag Reports:
    Show all entries for a specific project or tag with totals.

    did report @project            Show all entries for a specific project
    did report #tag                Show all entries with a specific tag
    did report --project acme      Alternative syntax for project filter
    did report --tag review        Alternative syntax for tag filter

  Grouped Reports:
    Show hours grouped by all projects or tags.

    did report --by project        Show hours grouped by all projects
    did report --by tag            Show hours grouped by all tags

Date Filtering:
  Use --from and --to to filter by date range
  Use --last to filter by relative days (e.g., 'last 7 days')

Examples:

  Single Project/Tag Reports:
    did report @acme                     Show all entries for project 'acme'
    did report #review                   Show all entries tagged 'review'
    did report --project acme            Alternative syntax
    did report --tag review              Alternative syntax
    did report @acme --last 7            Project report for last 7 days
    did report #bugfix --from 2024-01-01 Tag report from specific date

  Grouped Reports:
    did report --by project              Show hours by all projects
    did report --by tag                  Show hours by all tags
    did report --by project --last 30    Project breakdown for last 30 days
    did report --by tag --from 2024-01-01 --to 2024-01-31    Tag breakdown for date range`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		args = parseShorthandFilters(cmd, args)
		runReport(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Add --by flag for grouping mode
	reportCmd.Flags().String("by", "", "Group by 'project' or 'tag'")

	// Date filtering flags
	reportCmd.Flags().String("from", "", "Start date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	reportCmd.Flags().String("to", "", "End date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	reportCmd.Flags().Int("last", 0, "Filter by last N days (e.g., --last 7 for last 7 days)")

	// Note: --project and --tag flags are inherited from root command's PersistentFlags
}

// runReport handles the report command logic
func runReport(cmd *cobra.Command, args []string) {
	// Get flag values
	groupBy, _ := cmd.Flags().GetString("by")
	projectFilter, _ := cmd.Root().PersistentFlags().GetString("project")
	tagFilters, _ := cmd.Root().PersistentFlags().GetStringSlice("tag")

	// Validate --by flag value if provided
	if groupBy != "" && groupBy != "project" && groupBy != "tag" {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Invalid --by value. Must be 'project' or 'tag'")
		_, _ = fmt.Fprintln(deps.Stderr, "Usage:")
		_, _ = fmt.Fprintln(deps.Stderr, "  did report --by project")
		_, _ = fmt.Fprintln(deps.Stderr, "  did report --by tag")
		deps.Exit(1)
		return
	}

	// Validate flag combinations
	if groupBy != "" && (projectFilter != "" || len(tagFilters) > 0) {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Cannot use --by with --project or --tag filters")
		_, _ = fmt.Fprintln(deps.Stderr, "Use either:")
		_, _ = fmt.Fprintln(deps.Stderr, "  did report --by project     (grouped report)")
		_, _ = fmt.Fprintln(deps.Stderr, "  did report @project         (single project report)")
		deps.Exit(1)
		return
	}

	// Determine report mode
	if groupBy == "project" {
		// Grouped by project report (subtask 3.1)
		runGroupByProjectReport(cmd)
		return
	}

	if groupBy == "tag" {
		// Grouped by tag report (subtask 3.2)
		runGroupByTagReport(cmd)
		return
	}

	if projectFilter != "" {
		// Single project report (subtask 2.1)
		runSingleProjectReport(cmd, projectFilter)
		return
	}

	if len(tagFilters) > 0 {
		// Single tag report (subtask 2.2)
		runSingleTagReport(cmd, tagFilters)
		return
	}

	// No filters provided - show usage help
	_, _ = fmt.Fprintln(deps.Stderr, "Error: No filters specified")
	_, _ = fmt.Fprintln(deps.Stderr)
	_, _ = fmt.Fprintln(deps.Stderr, "Usage:")
	_, _ = fmt.Fprintln(deps.Stderr, "  did report @project              Show all entries for a project")
	_, _ = fmt.Fprintln(deps.Stderr, "  did report #tag                  Show all entries with a tag")
	_, _ = fmt.Fprintln(deps.Stderr, "  did report --by project          Show hours grouped by all projects")
	_, _ = fmt.Fprintln(deps.Stderr, "  did report --by tag              Show hours grouped by all tags")
	_, _ = fmt.Fprintln(deps.Stderr)
	_, _ = fmt.Fprintln(deps.Stderr, "Run 'did report --help' for more information")
	deps.Exit(1)
}

// runSingleProjectReport generates a report for a single project
func runSingleProjectReport(cmd *cobra.Command, projectFilter string) {
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

	// Filter out soft-deleted entries
	var activeEntries []entry.Entry
	for _, e := range result.Entries {
		if e.DeletedAt == nil {
			activeEntries = append(activeEntries, e)
		}
	}

	// Create filter with project
	f := filter.NewFilter("", projectFilter, nil)

	// Filter entries by project
	filtered := filter.FilterEntries(activeEntries, f)

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
			_, _ = fmt.Fprintf(deps.Stdout, "No entries found for project '@%s' in the specified date range\n", projectFilter)
		} else {
			_, _ = fmt.Fprintf(deps.Stdout, "No entries found for project '@%s'\n", projectFilter)
		}
		return
	}

	// Calculate total duration
	totalMinutes := 0
	for _, e := range filtered {
		totalMinutes += e.DurationMinutes
	}

	// Display results
	resultHeader := fmt.Sprintf("Report for project '@%s'", projectFilter)
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

// runSingleTagReport generates a report for one or more tags (ANDed together)
func runSingleTagReport(cmd *cobra.Command, tagFilters []string) {
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
				_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid --from date: %v\n", err)
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

	// Filter out soft-deleted entries
	var activeEntries []entry.Entry
	for _, e := range result.Entries {
		if e.DeletedAt == nil {
			activeEntries = append(activeEntries, e)
		}
	}

	// Create filter with tags (multiple tags are ANDed together)
	f := filter.NewFilter("", "", tagFilters)

	// Filter entries by tags
	filtered := filter.FilterEntries(activeEntries, f)

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

	// Format tag list for display
	var tagDisplay string
	if len(tagFilters) == 1 {
		tagDisplay = fmt.Sprintf("tag '#%s'", tagFilters[0])
	} else {
		// Multiple tags - format as '#tag1, #tag2'
		tagStrs := make([]string, len(tagFilters))
		for i, tag := range tagFilters {
			tagStrs[i] = "#" + tag
		}
		tagDisplay = fmt.Sprintf("tags '%s'", strings.Join(tagStrs, ", "))
	}

	// Check if any results found
	if len(filtered) == 0 {
		if hasDateFilter {
			_, _ = fmt.Fprintf(deps.Stdout, "No entries found for %s in the specified date range\n", tagDisplay)
		} else {
			_, _ = fmt.Fprintf(deps.Stdout, "No entries found for %s\n", tagDisplay)
		}
		return
	}

	// Calculate total duration
	totalMinutes := 0
	for _, e := range filtered {
		totalMinutes += e.DurationMinutes
	}

	// Display results
	resultHeader := fmt.Sprintf("Report for %s", tagDisplay)
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

// runGroupByProjectReport generates a report showing hours grouped by all projects
func runGroupByProjectReport(cmd *cobra.Command) {
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

	// Filter out soft-deleted entries
	var activeEntries []entry.Entry
	for _, e := range result.Entries {
		if e.DeletedAt == nil {
			activeEntries = append(activeEntries, e)
		}
	}

	// Apply date filtering if specified
	filtered := activeEntries
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
			_, _ = fmt.Fprintln(deps.Stdout, "No entries found in the specified date range")
		} else {
			_, _ = fmt.Fprintln(deps.Stdout, "No entries found")
		}
		return
	}

	// Group entries by project
	type ProjectGroup struct {
		Name         string
		TotalMinutes int
		EntryCount   int
	}

	projectGroups := make(map[string]*ProjectGroup)

	for _, e := range filtered {
		projectName := e.Project
		if projectName == "" {
			projectName = "(no project)"
		}

		if _, exists := projectGroups[projectName]; !exists {
			projectGroups[projectName] = &ProjectGroup{Name: projectName}
		}

		projectGroups[projectName].TotalMinutes += e.DurationMinutes
		projectGroups[projectName].EntryCount++
	}

	// Convert map to slice for sorting
	var groups []*ProjectGroup
	for _, group := range projectGroups {
		groups = append(groups, group)
	}

	// Sort by total time (descending)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].TotalMinutes > groups[j].TotalMinutes
	})

	// Calculate grand totals
	grandTotalMinutes := 0
	grandTotalEntries := 0
	for _, group := range groups {
		grandTotalMinutes += group.TotalMinutes
		grandTotalEntries += group.EntryCount
	}

	// Display results
	reportHeader := "Report grouped by project"
	if hasDateFilter {
		if lastDays > 0 {
			reportHeader += fmt.Sprintf(" (last %d %s)", lastDays, pluralize("day", lastDays))
		} else {
			reportHeader += fmt.Sprintf(" (%s)", formatDateRangeForDisplay(startDate, endDate))
		}
	}
	_, _ = fmt.Fprintf(deps.Stdout, "%s:\n", reportHeader)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 60))

	for _, group := range groups {
		// Format project name with special handling for "(no project)"
		projectDisplay := group.Name
		if group.Name == "(no project)" {
			projectDisplay = group.Name // Keep as is, no @ prefix
		} else {
			projectDisplay = "@" + group.Name
		}

		_, _ = fmt.Fprintf(deps.Stdout, "%-30s  %10s  (%d %s)\n",
			projectDisplay,
			formatDuration(group.TotalMinutes),
			group.EntryCount,
			pluralize("entry", group.EntryCount))
	}

	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 60))
	_, _ = fmt.Fprintf(deps.Stdout, "Grand Total: %s (%d %s across %d %s)\n",
		formatDuration(grandTotalMinutes),
		grandTotalEntries,
		pluralize("entry", grandTotalEntries),
		len(groups),
		pluralize("project", len(groups)))
}

// runGroupByTagReport generates a report showing hours grouped by all tags
func runGroupByTagReport(cmd *cobra.Command) {
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

	// Filter out soft-deleted entries
	var activeEntries []entry.Entry
	for _, e := range result.Entries {
		if e.DeletedAt == nil {
			activeEntries = append(activeEntries, e)
		}
	}

	// Apply date filtering if specified
	filtered := activeEntries
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
			_, _ = fmt.Fprintln(deps.Stdout, "No entries found in the specified date range")
		} else {
			_, _ = fmt.Fprintln(deps.Stdout, "No entries found")
		}
		return
	}

	// Group entries by tag
	// Note: Entries with multiple tags will contribute to each tag group
	type TagGroup struct {
		Name         string
		TotalMinutes int
		EntryCount   int
	}

	tagGroups := make(map[string]*TagGroup)

	for _, e := range filtered {
		// If entry has no tags, add to "(no tags)" group
		if len(e.Tags) == 0 {
			tagName := "(no tags)"
			if _, exists := tagGroups[tagName]; !exists {
				tagGroups[tagName] = &TagGroup{Name: tagName}
			}
			tagGroups[tagName].TotalMinutes += e.DurationMinutes
			tagGroups[tagName].EntryCount++
		} else {
			// Entry has tags - add to each tag group
			for _, tag := range e.Tags {
				if _, exists := tagGroups[tag]; !exists {
					tagGroups[tag] = &TagGroup{Name: tag}
				}
				tagGroups[tag].TotalMinutes += e.DurationMinutes
				tagGroups[tag].EntryCount++
			}
		}
	}

	// Convert map to slice for sorting
	var groups []*TagGroup
	for _, group := range tagGroups {
		groups = append(groups, group)
	}

	// Sort by total time (descending)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].TotalMinutes > groups[j].TotalMinutes
	})

	// Calculate grand totals
	// Note: Because entries with multiple tags contribute to multiple groups,
	// we need to count unique entries, not sum up entry counts from groups
	grandTotalMinutes := 0
	grandTotalEntries := len(filtered)
	for _, e := range filtered {
		grandTotalMinutes += e.DurationMinutes
	}

	// Display results
	reportHeader := "Report grouped by tag"
	if hasDateFilter {
		if lastDays > 0 {
			reportHeader += fmt.Sprintf(" (last %d %s)", lastDays, pluralize("day", lastDays))
		} else {
			reportHeader += fmt.Sprintf(" (%s)", formatDateRangeForDisplay(startDate, endDate))
		}
	}
	_, _ = fmt.Fprintf(deps.Stdout, "%s:\n", reportHeader)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 60))

	for _, group := range groups {
		// Format tag name with special handling for "(no tags)"
		tagDisplay := group.Name
		if group.Name == "(no tags)" {
			tagDisplay = group.Name // Keep as is, no # prefix
		} else {
			tagDisplay = "#" + group.Name
		}

		_, _ = fmt.Fprintf(deps.Stdout, "%-30s  %10s  (%d %s)\n",
			tagDisplay,
			formatDuration(group.TotalMinutes),
			group.EntryCount,
			pluralize("entry", group.EntryCount))
	}

	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 60))
	_, _ = fmt.Fprintf(deps.Stdout, "Grand Total: %s (%d %s across %d %s)\n",
		formatDuration(grandTotalMinutes),
		grandTotalEntries,
		pluralize("entry", grandTotalEntries),
		len(groups),
		pluralize("tag", len(groups)))
}
