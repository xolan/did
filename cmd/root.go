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

var rootCmd = &cobra.Command{
	Use:   "did",
	Short: "A time tracking CLI application",
	Long: `did is a CLI tool for logging work activities with time durations.

Usage:
  did <description> for <duration>    Log a new entry (e.g., did feature X for 2h)
  did                                 List today's entries (default)

Time Period Flags (mutually exclusive):
  -y, --yesterday                     List yesterday's entries
  -w, --this-week                     List current week's entries
      --prev-week                     List previous week's entries
  -m, --this-month                    List current month's entries
      --prev-month                    List previous month's entries
  -l, --last <n>                      List entries from last N days
      --from <date> --to <date>       List entries in date range
  -d, --date <date>                   List entries for a specific date

Filter Options:
  --project <name>                    Filter entries by project
  --tag <name>                        Filter entries by tag (can be repeated)
  @project                            Shorthand for --project
  #tag                                Shorthand for --tag

Examples:
  did feature X for 2h                Log a new entry
  did                                 List today's entries
  did -y                              List yesterday's entries
  did -w                              List this week's entries
  did --prev-week                     List last week's entries
  did -m                              List this month's entries
  did --prev-month                    List last month's entries
  did -l 7                            List last 7 days
  did --from 2024-01-01 --to 2024-01-31   List entries in date range
  did -d 2024-01-15                   List entries for specific date
  did -w @acme                        This week's entries for project 'acme'
  did -l 30 #bugfix                   Last 30 days tagged 'bugfix'
  did --prev-week @client #urgent     Last week's entries with filters

Other Commands:
  did edit <index> --description 'text'   Edit entry description
  did edit <index> --duration 2h          Edit entry duration
  did delete <index>                      Delete an entry (with confirmation)
  did undo                                Restore the most recently deleted entry
  did purge                               Permanently remove all soft-deleted entries
  did validate                            Check storage file health
  did restore [n]                         Restore from backup (default: most recent)
  did search <keyword>                    Search entries by keyword
  did export json|csv                     Export entries to JSON or CSV
  did report @project|#tag|--by <type>    Generate reports
  did stats [--month]                     Show statistics

Timer Mode:
  did start <description>             Start a timer for a task
  did stop                            Stop the timer and create an entry
  did status                          Show current timer status

Duration format: Yh (hours), Ym (minutes), or YhYm (combined)
Examples: 2h, 30m, 1h30m

Date formats: YYYY-MM-DD or DD/MM/YYYY
Examples: 2024-01-15 or 15/01/2024

Projects and Tags:
  Optionally categorize entries with @project and #tags in descriptions.
  did fix login bug @acme for 1h      Assign entry to project 'acme'
  did code review #review for 30m     Add tag 'review' to entry
  did API work @client #backend for 2h    Combine project with multiple tags`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		args = parseShorthandFilters(cmd, args)

		// Check for time period flags and handle listing
		if handled := handleTimePeriodFlags(cmd, args); handled {
			return
		}

		// No time period flags - check for entry creation or default listing
		if len(args) == 0 {
			// No args and no time flags: list today's entries
			listEntries(cmd, "today", timeutil.Today)
			return
		}

		// With args: create a new entry
		createEntry(args)
	},
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit <index>",
	Short: "Edit an existing entry",
	Long: `Edit the description or duration of an existing time tracking entry.

Usage:
  did edit <index> --description 'new text'    Update entry description
  did edit <index> --duration 2h               Update entry duration
  did edit <index> --description 'text' --duration 2h    Update both

The index refers to the entry number shown in list output (starting from 1).
At least one flag (--description or --duration) is required.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		editEntry(cmd, args)
	},
}

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check storage file health",
	Long:  `Validate the storage file and report on its health status, including any corrupted entries.`,
	Run: func(cmd *cobra.Command, args []string) {
		validateStorage()
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(undoCmd)
	rootCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)

	// Add persistent filter flags (apply to all commands)
	rootCmd.PersistentFlags().String("project", "", "Filter entries by project")
	rootCmd.PersistentFlags().StringSlice("tag", []string{}, "Filter entries by tag (can be repeated)")

	// Add time period flags to root command
	rootCmd.Flags().BoolP("yesterday", "y", false, "List yesterday's entries")
	rootCmd.Flags().BoolP("this-week", "w", false, "List current week's entries")
	rootCmd.Flags().Bool("prev-week", false, "List previous week's entries")
	rootCmd.Flags().BoolP("this-month", "m", false, "List current month's entries")
	rootCmd.Flags().Bool("prev-month", false, "List previous month's entries")
	rootCmd.Flags().IntP("last", "l", 0, "List entries from last N days")
	rootCmd.Flags().String("from", "", "Start date for date range (YYYY-MM-DD or DD/MM/YYYY)")
	rootCmd.Flags().String("to", "", "End date for date range (YYYY-MM-DD or DD/MM/YYYY)")
	rootCmd.Flags().StringP("date", "d", "", "List entries for a specific date (YYYY-MM-DD or DD/MM/YYYY)")

	// Add flags to edit command
	editCmd.Flags().String("description", "", "New description for the entry")
	editCmd.Flags().String("duration", "", "New duration for the entry (e.g., 2h, 30m)")
}

// handleTimePeriodFlags checks for time period flags and lists entries accordingly.
// Returns true if a time period flag was handled, false otherwise.
func handleTimePeriodFlags(cmd *cobra.Command, args []string) bool {
	// Get all time period flag values
	yesterday, _ := cmd.Flags().GetBool("yesterday")
	thisWeek, _ := cmd.Flags().GetBool("this-week")
	prevWeek, _ := cmd.Flags().GetBool("prev-week")
	thisMonth, _ := cmd.Flags().GetBool("this-month")
	prevMonth, _ := cmd.Flags().GetBool("prev-month")
	lastDays, _ := cmd.Flags().GetInt("last")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	dateStr, _ := cmd.Flags().GetString("date")

	// Count how many time period options are set
	count := 0
	if yesterday {
		count++
	}
	if thisWeek {
		count++
	}
	if prevWeek {
		count++
	}
	if thisMonth {
		count++
	}
	if prevMonth {
		count++
	}
	if lastDays > 0 {
		count++
	}
	if fromStr != "" || toStr != "" {
		count++
	}
	if dateStr != "" {
		count++
	}

	// Check for mutual exclusivity
	if count > 1 {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Time period flags are mutually exclusive")
		_, _ = fmt.Fprintln(deps.Stderr, "Use only one of: --yesterday, --this-week, --prev-week, --this-month, --prev-month, --last, --from/--to, --date")
		deps.Exit(1)
		return true
	}

	// If no time period flags, return false to continue normal processing
	if count == 0 {
		// Check if this looks like shorthand filters only (no 'for' keyword)
		// In this case, treat as listing command
		if len(args) > 0 {
			rawInput := strings.Join(args, " ")
			if !strings.Contains(strings.ToLower(rawInput), " for ") {
				// No 'for' keyword - likely shorthand filters for listing
				listEntries(cmd, "today", timeutil.Today)
				return true
			}
		}
		return false
	}

	// Check if args contain entry creation (has 'for' keyword) - time flags shouldn't be used with entry creation
	if len(args) > 0 {
		rawInput := strings.Join(args, " ")
		if strings.Contains(strings.ToLower(rawInput), " for ") {
			_, _ = fmt.Fprintln(deps.Stderr, "Error: Time period flags cannot be used when creating entries")
			_, _ = fmt.Fprintln(deps.Stderr, "To create an entry: did <description> for <duration>")
			_, _ = fmt.Fprintln(deps.Stderr, "To list entries: did [time-flag] [@project] [#tag]")
			deps.Exit(1)
			return true
		}
	}

	// Handle each time period flag
	if yesterday {
		listEntries(cmd, "yesterday", timeutil.Yesterday)
		return true
	}

	if thisWeek {
		now := time.Now()
		start := timeutil.StartOfWeekWithConfig(now, deps.Config.WeekStartDay)
		end := timeutil.EndOfWeekWithConfig(now, deps.Config.WeekStartDay)
		dateRange := formatDateRangeForDisplay(start, end)
		period := fmt.Sprintf("this week (%s)", dateRange)
		listEntriesForRange(cmd, period, start, end)
		return true
	}

	if prevWeek {
		lastWeek := time.Now().AddDate(0, 0, -7)
		start := timeutil.StartOfWeekWithConfig(lastWeek, deps.Config.WeekStartDay)
		end := timeutil.EndOfWeekWithConfig(lastWeek, deps.Config.WeekStartDay)
		dateRange := formatDateRangeForDisplay(start, end)
		period := fmt.Sprintf("previous week (%s)", dateRange)
		listEntriesForRange(cmd, period, start, end)
		return true
	}

	if thisMonth {
		now := time.Now()
		start := timeutil.StartOfMonth(now)
		end := timeutil.EndOfMonth(now)
		dateRange := formatDateRangeForDisplay(start, end)
		period := fmt.Sprintf("this month (%s)", dateRange)
		listEntriesForRange(cmd, period, start, end)
		return true
	}

	if prevMonth {
		lastMonth := time.Now().AddDate(0, -1, 0)
		start := timeutil.StartOfMonth(lastMonth)
		end := timeutil.EndOfMonth(lastMonth)
		dateRange := formatDateRangeForDisplay(start, end)
		period := fmt.Sprintf("previous month (%s)", dateRange)
		listEntriesForRange(cmd, period, start, end)
		return true
	}

	if lastDays > 0 {
		now := time.Now()
		end := timeutil.EndOfDay(now)
		start := timeutil.StartOfDay(now.AddDate(0, 0, -(lastDays - 1)))
		dateRange := formatDateRangeForDisplay(start, end)
		period := fmt.Sprintf("last %d %s (%s)", lastDays, pluralize("day", lastDays), dateRange)
		listEntriesForRange(cmd, period, start, end)
		return true
	}

	if fromStr != "" || toStr != "" {
		var startDate, endDate time.Time
		var err error

		// Parse from date
		if fromStr != "" {
			startDate, err = timeutil.ParseDate(fromStr)
			if err != nil {
				_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid --from date: %v\n", err)
				_, _ = fmt.Fprintln(deps.Stderr, "Use format YYYY-MM-DD or DD/MM/YYYY")
				deps.Exit(1)
				return true
			}
		} else {
			// No from date: use beginning of time
			startDate = time.Time{}
		}

		// Parse to date
		if toStr != "" {
			toDate, err := timeutil.ParseDate(toStr)
			if err != nil {
				_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid --to date: %v\n", err)
				_, _ = fmt.Fprintln(deps.Stderr, "Use format YYYY-MM-DD or DD/MM/YYYY")
				deps.Exit(1)
				return true
			}
			endDate = timeutil.EndOfDay(toDate)
		} else {
			// No to date: use end of today
			endDate = timeutil.EndOfDay(time.Now())
		}

		// Validate that start is not after end
		if !startDate.IsZero() && startDate.After(endDate) {
			_, _ = fmt.Fprintf(deps.Stderr, "Error: --from date (%s) is after --to date (%s)\n",
				startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
			deps.Exit(1)
			return true
		}

		period := formatDateRangeForDisplay(startDate, endDate)
		listEntriesForRange(cmd, period, startDate, endDate)
		return true
	}

	if dateStr != "" {
		date, err := timeutil.ParseDate(dateStr)
		if err != nil {
			_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid --date value: %v\n", err)
			_, _ = fmt.Fprintln(deps.Stderr, "Use format YYYY-MM-DD or DD/MM/YYYY")
			deps.Exit(1)
			return true
		}
		endDate := timeutil.EndOfDay(date)
		period := formatDateRangeForDisplay(date, endDate)
		listEntriesForRange(cmd, period, date, endDate)
		return true
	}

	return false
}

// SetVersionInfo sets the version information for the CLI
func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(
		"did version {{.Version}}\n" +
			"commit: " + commit + "\n" +
			"built: " + date + "\n",
	)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// parseShorthandFilters parses @project and #tag shorthand syntax from args.
// It sets the corresponding flags for filtering, but does NOT modify the args
// so that project/tags can be parsed later for entry creation.
// Example: ["@acme", "#bugfix"] -> flags set, args returned unchanged
func parseShorthandFilters(cmd *cobra.Command, args []string) []string {
	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			project := strings.TrimPrefix(arg, "@")
			if project != "" {
				_ = cmd.Root().PersistentFlags().Set("project", project)
			}
		} else if strings.HasPrefix(arg, "#") {
			tag := strings.TrimPrefix(arg, "#")
			if tag != "" {
				_ = cmd.Root().PersistentFlags().Set("tag", tag)
			}
		}
	}
	return args
}

// createEntry parses arguments and creates a new time tracking entry
func createEntry(args []string) {
	// Join all arguments to form the raw input
	rawInput := strings.Join(args, " ")

	// Parse the input: expected format "<description> for <duration>"
	// Find the last "for" in the input to extract duration
	lastForIdx := strings.LastIndex(strings.ToLower(rawInput), " for ")
	if lastForIdx == -1 {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Invalid format. Missing 'for <duration>'")
		_, _ = fmt.Fprintln(deps.Stderr, "Usage: did <description> for <duration>")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: did feature X for 2h")
		deps.Exit(1)
		return
	}

	description := strings.TrimSpace(rawInput[:lastForIdx])
	durationStr := strings.TrimSpace(rawInput[lastForIdx+5:]) // +5 for " for "

	if description == "" {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Description cannot be empty")
		deps.Exit(1)
		return
	}

	// Parse project and tags from description
	cleanDesc, project, tags := entry.ParseProjectAndTags(description)

	// Check that cleaned description is not empty (in case it was only @project/#tags)
	if cleanDesc == "" {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Description cannot be empty (only project/tags provided)")
		deps.Exit(1)
		return
	}

	// Parse the duration
	minutes, err := entry.ParseDuration(durationStr)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid duration '%s'\n", durationStr)
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Use format like '2h' (hours) or '30m' (minutes), max 24h")
		deps.Exit(1)
		return
	}

	// Create the entry
	e := entry.Entry{
		Timestamp:       time.Now(),
		Description:     cleanDesc,
		DurationMinutes: minutes,
		RawInput:        rawInput,
		Project:         project,
		Tags:            tags,
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

	// Append the entry to storage
	if err := storage.AppendEntry(storagePath, e); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to save entry to storage")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintf(deps.Stderr, "Hint: Check that directory exists and is writable: %s\n", storagePath)
		deps.Exit(1)
		return
	}

	// Display success message
	_, _ = fmt.Fprintf(deps.Stdout, "Logged: %s (%s)\n", description, formatDuration(minutes))
}

// listEntries reads and displays entries filtered by the given time range.
// This function accepts a function that returns start/end times.
func listEntries(cmd *cobra.Command, period string, timeRangeFunc func() (time.Time, time.Time)) {
	start, end := timeRangeFunc()
	listEntriesForRange(cmd, period, start, end)
}

// listEntriesForRange reads and displays entries filtered by explicit start/end times and optional filters
func listEntriesForRange(cmd *cobra.Command, period string, start, end time.Time) {
	storagePath, err := deps.StoragePath()
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to determine storage location")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Check that your home directory is accessible")
		deps.Exit(1)
		return
	}

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

	type indexedEntry struct {
		entry.Entry
		activeIndex int
	}

	var activeEntries []indexedEntry
	activeIdx := 0
	for _, e := range result.Entries {
		if e.DeletedAt == nil {
			activeIdx++
			activeEntries = append(activeEntries, indexedEntry{Entry: e, activeIndex: activeIdx})
		}
	}

	var filtered []indexedEntry
	for _, ie := range activeEntries {
		if timeutil.IsInRange(ie.Timestamp, start, end) {
			filtered = append(filtered, ie)
		}
	}

	projectFilter, _ := cmd.Root().PersistentFlags().GetString("project")
	tagFilters, _ := cmd.Root().PersistentFlags().GetStringSlice("tag")

	f := filter.NewFilter("", projectFilter, tagFilters)
	if !f.IsEmpty() {
		var projectTagFiltered []indexedEntry
		for _, ie := range filtered {
			if f.Matches(ie.Entry) {
				projectTagFiltered = append(projectTagFiltered, ie)
			}
		}
		filtered = projectTagFiltered
		period = buildPeriodWithFilters(period, projectFilter, tagFilters)
	}

	if len(filtered) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No entries found for %s\n", period)
		return
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.Before(filtered[j].Timestamp)
	})

	totalMinutes := 0
	for _, ie := range filtered {
		totalMinutes += ie.DurationMinutes
	}

	_, _ = fmt.Fprintf(deps.Stdout, "Entries for %s:\n", period)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))

	maxIndex := activeEntries[len(activeEntries)-1].activeIndex
	maxIndexWidth := len(fmt.Sprintf("%d", maxIndex))

	entriesForDateCheck := make([]entry.Entry, len(filtered))
	for i, ie := range filtered {
		entriesForDateCheck[i] = ie.Entry
	}
	showDate := spansMultipleDays(entriesForDateCheck)

	for _, ie := range filtered {
		if showDate {
			_, _ = fmt.Fprintf(deps.Stdout, "[%*d] %s %s  %s (%s)\n",
				maxIndexWidth,
				ie.activeIndex,
				ie.Timestamp.Format("2006-01-02"),
				ie.Timestamp.Format("15:04"),
				formatEntryForLog(ie.Description, ie.Project, ie.Tags),
				formatDuration(ie.DurationMinutes))
		} else {
			_, _ = fmt.Fprintf(deps.Stdout, "[%*d] %s  %s (%s)\n",
				maxIndexWidth,
				ie.activeIndex,
				ie.Timestamp.Format("15:04"),
				formatEntryForLog(ie.Description, ie.Project, ie.Tags),
				formatDuration(ie.DurationMinutes))
		}
	}
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "Total: %s\n", formatDuration(totalMinutes))
}

// formatDateRangeForDisplay formats a date range for human-readable display.
// Used for custom date range queries to generate appropriate period descriptions.
func formatDateRangeForDisplay(start, end time.Time) string {
	// If same day, show single date
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		return start.Format("Mon, Jan 2, 2006")
	}

	// If same year, don't repeat the year
	if start.Year() == end.Year() {
		return fmt.Sprintf("%s - %s",
			start.Format("Jan 2"),
			end.Format("Jan 2, 2006"))
	}

	// Different years, show both
	return fmt.Sprintf("%s - %s",
		start.Format("Jan 2, 2006"),
		end.Format("Jan 2, 2006"))
}

// buildPeriodWithFilters appends filter information to the period description.
// Example: "today" -> "today (@acme #bugfix)"
func buildPeriodWithFilters(period, project string, tags []string) string {
	if project == "" && len(tags) == 0 {
		return period
	}

	var filters []string
	if project != "" {
		filters = append(filters, "@"+project)
	}
	for _, tag := range tags {
		filters = append(filters, "#"+tag)
	}

	return fmt.Sprintf("%s (%s)", period, strings.Join(filters, " "))
}

// formatCorruptionWarning formats a ParseWarning into a human-readable string
// with line number, truncated content (max 50 chars), and error description.
func formatCorruptionWarning(warning storage.ParseWarning) string {
	// Truncate content if too long (max 50 chars)
	content := warning.Content
	if len(content) > 50 {
		content = content[:47] + "..."
	}
	return fmt.Sprintf("  Line %d: %s (error: %s)", warning.LineNumber, content, warning.Error)
}

// validateStorage checks the storage file health and reports status
func validateStorage() {
	storagePath, err := deps.StoragePath()
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to get storage path: %v\n", err)
		deps.Exit(1)
		return
	}

	health, err := storage.ValidateStorage(storagePath)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to validate storage: %v\n", err)
		deps.Exit(1)
		return
	}

	// Display storage path
	_, _ = fmt.Fprintf(deps.Stdout, "Storage file: %s\n", storagePath)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))

	// Display health metrics
	_, _ = fmt.Fprintf(deps.Stdout, "Total lines:       %d\n", health.TotalLines)
	_, _ = fmt.Fprintf(deps.Stdout, "Valid entries:     %d\n", health.ValidEntries)
	_, _ = fmt.Fprintf(deps.Stdout, "Corrupted entries: %d\n", health.CorruptedEntries)

	// Display corrupted line details if any
	if len(health.Warnings) > 0 {
		_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))
		_, _ = fmt.Fprintln(deps.Stdout, "Corrupted lines:")
		for _, warning := range health.Warnings {
			_, _ = fmt.Fprintln(deps.Stdout, formatCorruptionWarning(warning))
		}
	}

	// Overall status message
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))
	if health.CorruptedEntries == 0 {
		_, _ = fmt.Fprintln(deps.Stdout, "Status: ✓ Storage file is healthy")
	} else {
		_, _ = fmt.Fprintf(deps.Stderr, "Status: ⚠ Storage file has %d corrupted line(s)\n", health.CorruptedEntries)
	}
}

// formatDuration formats minutes as a human-readable string
func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

// formatProjectAndTags formats project and tags for display.
// Returns format like: "@project" or "#tag1 #tag2" or "@project #tag1 #tag2"
// Returns empty string if no project or tags.
func formatProjectAndTags(project string, tags []string) string {
	if project == "" && len(tags) == 0 {
		return ""
	}

	var parts []string
	if project != "" {
		parts = append(parts, "@"+project)
	}
	for _, tag := range tags {
		parts = append(parts, "#"+tag)
	}

	return strings.Join(parts, " ")
}

// formatEntryForLog formats a description with optional project and tags for display.
// Returns format like: "description" or "description [@project]" or "description [#tag1 #tag2]"
// or "description [@project #tag1 #tag2]"
func formatEntryForLog(description, project string, tags []string) string {
	metadata := formatProjectAndTags(project, tags)
	if metadata == "" {
		return description
	}
	return fmt.Sprintf("%s [%s]", description, metadata)
}

// editEntry modifies an existing time tracking entry
func editEntry(cmd *cobra.Command, args []string) {
	// Parse the index argument (1-based from user)
	var userIndex int
	if _, err := fmt.Sscanf(args[0], "%d", &userIndex); err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid index '%s'. Index must be a number\n", args[0])
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: List entries with 'did' to see available indices")
		deps.Exit(1)
		return
	}

	// Get flag values
	newDescription, _ := cmd.Flags().GetString("description")
	newDuration, _ := cmd.Flags().GetString("duration")

	// Check that at least one flag is provided
	if newDescription == "" && newDuration == "" {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: At least one flag (--description or --duration) is required")
		_, _ = fmt.Fprintln(deps.Stderr, "Usage:")
		_, _ = fmt.Fprintln(deps.Stderr, "  did edit <index> --description 'new text'")
		_, _ = fmt.Fprintln(deps.Stderr, "  did edit <index> --duration 2h")
		_, _ = fmt.Fprintln(deps.Stderr, "  did edit <index> --description 'new text' --duration 2h")
		deps.Exit(1)
		return
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

	// Read all entries
	result, err := storage.ReadEntriesWithWarnings(storagePath)
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to read entries from storage")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintf(deps.Stderr, "Hint: Check that file exists and is readable: %s\n", storagePath)
		deps.Exit(1)
		return
	}

	allEntries := result.Entries

	// Filter to active entries only and create index mapping
	// Users should only be able to edit active (non-deleted) entries
	var activeEntries []entry.Entry
	var storageIndices []int // Maps active entry index to storage index
	for i, e := range allEntries {
		if e.DeletedAt == nil {
			activeEntries = append(activeEntries, e)
			storageIndices = append(storageIndices, i)
		}
	}

	// Check if any active entries exist
	if len(activeEntries) == 0 {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: No entries found to edit")
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Create an entry first with 'did <description> for <duration>'")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: did feature X for 2h")
		deps.Exit(1)
		return
	}

	// Convert 1-based user index to 0-based active entry index
	activeIndex := userIndex - 1

	// Validate index is in range of active entries
	if activeIndex < 0 || activeIndex >= len(activeEntries) {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Index %d is out of range\n", userIndex)
		_, _ = fmt.Fprintf(deps.Stderr, "Valid range: 1-%d (%d %s available)\n", len(activeEntries), len(activeEntries), pluralize("entry", len(activeEntries)))
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: List entries with 'did' to see all indices")
		deps.Exit(1)
		return
	}

	// Get the entry to modify (from active entries)
	e := activeEntries[activeIndex]

	// Get the actual storage index for this entry
	storageIndex := storageIndices[activeIndex]

	// Update description if provided
	if newDescription != "" {
		// Parse project and tags from new description
		cleanDesc, project, tags := entry.ParseProjectAndTags(newDescription)

		// Check that cleaned description is not empty (in case it was only @project/#tags)
		if cleanDesc == "" {
			_, _ = fmt.Fprintln(deps.Stderr, "Error: Description cannot be empty (only project/tags provided)")
			deps.Exit(1)
			return
		}

		e.Description = cleanDesc
		e.Project = project
		e.Tags = tags
	}

	// Update duration if provided
	if newDuration != "" {
		minutes, err := entry.ParseDuration(newDuration)
		if err != nil {
			_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid duration '%s'\n", newDuration)
			_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
			_, _ = fmt.Fprintln(deps.Stderr, "Hint: Use format like '2h' (hours) or '30m' (minutes), max 24h")
			deps.Exit(1)
			return
		}
		e.DurationMinutes = minutes
	}

	// Update RawInput field to reflect changes
	// Include project/tags in raw input reconstruction
	descWithMeta := e.Description
	if e.Project != "" || len(e.Tags) > 0 {
		descWithMeta = fmt.Sprintf("%s %s", e.Description, formatProjectAndTags(e.Project, e.Tags))
	}
	if newDescription != "" && newDuration != "" {
		// Both updated
		e.RawInput = fmt.Sprintf("%s for %s", descWithMeta, newDuration)
	} else if newDescription != "" {
		// Only description updated - reconstruct with existing duration
		e.RawInput = fmt.Sprintf("%s for %s", descWithMeta, formatDuration(e.DurationMinutes))
	} else if newDuration != "" {
		// Only duration updated - reconstruct with existing description and project/tags
		e.RawInput = fmt.Sprintf("%s for %s", descWithMeta, newDuration)
	}

	// Preserve original timestamp (already unchanged in e)

	// Save the updated entry
	if err := storage.UpdateEntry(storagePath, storageIndex, e); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to save updated entry to storage")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintf(deps.Stderr, "Hint: Check that file is writable: %s\n", storagePath)
		deps.Exit(1)
		return
	}

	// Display success message with project/tags
	_, _ = fmt.Fprintf(deps.Stdout, "Updated entry %d: %s (%s)\n", userIndex, formatEntryForLog(e.Description, e.Project, e.Tags), formatDuration(e.DurationMinutes))
}

// pluralize returns the singular or plural form of a word based on count
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func spansMultipleDays(entries []entry.Entry) bool {
	if len(entries) < 2 {
		return false
	}
	firstDay := entries[0].Timestamp.Format("2006-01-02")
	for _, e := range entries[1:] {
		if e.Timestamp.Format("2006-01-02") != firstDay {
			return true
		}
	}
	return false
}
