package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

var rootCmd = &cobra.Command{
	Use:   "did",
	Short: "A time tracking CLI application",
	Long: `did is a CLI tool for logging work activities with time durations.

Usage:
  did <description> for <duration>              Log a new entry (e.g., did feature X for 2h)
  did                                           List today's entries
  did y                                         List yesterday's entries
  did w                                         List this week's entries
  did lw                                        List last week's entries
  did edit <index> --description 'text'         Edit entry description
  did edit <index> --duration 2h                Edit entry duration
  did delete <index>                            Delete an entry (with confirmation)
  did validate                                  Check storage file health
  did restore [n]                               Restore from backup (default: most recent)
  did search <keyword>                          Search entries by keyword

Date Query Commands:
  did 2024-01-15                                List entries for a specific date
  did from 2024-01-01 to 2024-01-31             List entries for a date range
  did last 7 days                               List entries from the past 7 days

Filter Options:
  --project <name>                              Filter entries by project
  --tag <name>                                  Filter entries by tag (can be repeated)
  @project                                      Shorthand for --project
  #tag                                          Shorthand for --tag

Filter Examples:
  did --project acme                            List today's entries for project 'acme'
  did @acme                                     Same as above (shorthand syntax)
  did w --tag bugfix                            List this week's entries tagged 'bugfix'
  did #bugfix                                   List today's entries tagged 'bugfix'
  did y @client #urgent                         Yesterday's entries for project 'client' tagged 'urgent'
  did --project acme --tag review --tag urgent  Multiple filters combined

Duration format: Yh (hours), Ym (minutes), or YhYm (combined)
Examples: 2h, 30m, 1h30m

Date formats: YYYY-MM-DD or DD/MM/YYYY
Examples: 2024-01-15 or 15/01/2024

Projects and Tags:
  Optionally categorize entries with @project and #tags in descriptions.
  did fix login bug @acme for 1h                Assign entry to project 'acme'
  did code review #review for 30m               Add tag 'review' to entry
  did API work @client #backend #api for 2h     Combine project with multiple tags`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		args = parseShorthandFilters(cmd, args)

		if len(args) == 0 {
			// No args: list today's entries
			listEntries("today", timeutil.Today)
			return
		}

		// Check if this is a date query (single date argument) vs entry creation
		// Entry creation requires 'for' keyword
		rawInput := strings.Join(args, " ")
		hasForKeyword := strings.Contains(strings.ToLower(rawInput), " for ")

		// If no 'for' keyword and single argument that looks like a date, list entries for that date
		if !hasForKeyword && len(args) == 1 {
			if date, err := timeutil.ParseDate(args[0]); err == nil {
				// Successfully parsed as a date - list entries for that day
				endDate := timeutil.EndOfDay(date)
				periodDesc := formatDateRangeForDisplay(date, endDate)
				listEntriesForRange(periodDesc, date, endDate)
				return
			} else {
				// Single argument without 'for' keyword failed to parse as date
				// Show helpful error message about date format OR entry creation
				_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
				_, _ = fmt.Fprintln(deps.Stderr)
				_, _ = fmt.Fprintln(deps.Stderr, "If querying a date, use format YYYY-MM-DD or DD/MM/YYYY:")
				_, _ = fmt.Fprintln(deps.Stderr, "  did 2024-01-15")
				_, _ = fmt.Fprintln(deps.Stderr, "  did 15/01/2024")
				_, _ = fmt.Fprintln(deps.Stderr)
				_, _ = fmt.Fprintln(deps.Stderr, "If creating an entry, include 'for <duration>':")
				_, _ = fmt.Fprintln(deps.Stderr, "  did <description> for <duration>")
				_, _ = fmt.Fprintln(deps.Stderr, "  did feature X for 2h")
				deps.Exit(1)
				return
			}
		}

		// With args: create a new entry
		createEntry(args)
	},
}

// yCmd represents the yesterday command
var yCmd = &cobra.Command{
	Use:   "y",
	Short: "List yesterday's entries",
	Long:  `List all time tracking entries logged yesterday.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		_ = parseShorthandFilters(cmd, args)
		listEntries("yesterday", timeutil.Yesterday)
	},
}

// wCmd represents the this week command
var wCmd = &cobra.Command{
	Use:   "w",
	Short: "List this week's entries",
	Long:  `List all time tracking entries logged this week (Monday-Sunday).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		_ = parseShorthandFilters(cmd, args)
		listEntries("this week", timeutil.ThisWeek)
	},
}

// lwCmd represents the last week command
var lwCmd = &cobra.Command{
	Use:   "lw",
	Short: "List last week's entries",
	Long:  `List all time tracking entries logged last week (Monday-Sunday).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		_ = parseShorthandFilters(cmd, args)
		listEntries("last week", timeutil.LastWeek)
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
	rootCmd.AddCommand(yCmd)
	rootCmd.AddCommand(wCmd)
	rootCmd.AddCommand(lwCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(deleteCmd)

	// Add persistent filter flags (apply to all commands)
	rootCmd.PersistentFlags().String("project", "", "Filter entries by project")
	rootCmd.PersistentFlags().StringSlice("tag", []string{}, "Filter entries by tag (can be repeated)")

	// Add flags to edit command
	editCmd.Flags().String("description", "", "New description for the entry")
	editCmd.Flags().String("duration", "", "New duration for the entry (e.g., 2h, 30m)")
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
// It extracts shorthand filters, sets the corresponding flags, and returns the remaining args.
// Example: ["@acme", "#bugfix", "y"] -> flags set, returns ["y"]
func parseShorthandFilters(cmd *cobra.Command, args []string) []string {
	var remainingArgs []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			// Extract project (remove @ prefix)
			project := strings.TrimPrefix(arg, "@")
			if project != "" {
				_ = cmd.Root().PersistentFlags().Set("project", project)
			}
		} else if strings.HasPrefix(arg, "#") {
			// Extract tag (remove # prefix)
			tag := strings.TrimPrefix(arg, "#")
			if tag != "" {
				// cobra StringSlice flags append when Set is called multiple times
				_ = cmd.Root().PersistentFlags().Set("tag", tag)
			}
		} else {
			// Not a shorthand filter, keep in remaining args
			remainingArgs = append(remainingArgs, arg)
		}
	}

	return remainingArgs
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

	// Display success message with optional project and tags
	_, _ = fmt.Fprintf(deps.Stdout, "Logged: %s (%s)\n", formatEntryForLog(cleanDesc, project, tags), formatDuration(minutes))
}

// listEntries reads and displays entries filtered by the given time range
// This function accepts a function that returns start/end times for backward compatibility
func listEntries(period string, timeRangeFunc func() (time.Time, time.Time)) {
	start, end := timeRangeFunc()
	listEntriesForRange(period, start, end)
}

// listEntriesForRange reads and displays entries filtered by explicit start/end times
func listEntriesForRange(period string, start, end time.Time) {
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

	entries := result.Entries

	// Filter entries by time range
	var filtered []entry.Entry
	for _, e := range entries {
		if timeutil.IsInRange(e.Timestamp, start, end) {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No entries found for %s\n", period)
		return
	}

	// Calculate total duration
	totalMinutes := 0
	for _, e := range filtered {
		totalMinutes += e.DurationMinutes
	}

	// Display entries
	_, _ = fmt.Fprintf(deps.Stdout, "Entries for %s:\n", period)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))

	// Calculate width for right-aligned indices
	maxIndexWidth := len(fmt.Sprintf("%d", len(filtered)))

	for i, e := range filtered {
		_, _ = fmt.Fprintf(deps.Stdout, "[%*d] %s  %s (%s)\n",
			maxIndexWidth,
			i+1, // 1-based index for user reference
			e.Timestamp.Format("15:04"),
			formatEntryForLog(e.Description, e.Project, e.Tags),
			formatDuration(e.DurationMinutes))
	}
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "Total: %s\n", formatDuration(totalMinutes))
}

// formatDateRangeForDisplay formats a date range for human-readable display
// Used for custom date range queries to generate appropriate period descriptions
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

// formatProjectAndTags formats project and tags for display
// Returns format like: "@project" or "#tag1 #tag2" or "@project #tag1 #tag2"
// Returns empty string if no project or tags
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

// formatEntryForLog formats a description with optional project and tags for display
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

	entries := result.Entries

	// Check if any entries exist
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: No entries found to edit")
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Create an entry first with 'did <description> for <duration>'")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: did feature X for 2h")
		deps.Exit(1)
		return
	}

	// Convert 1-based user index to 0-based internal index
	internalIndex := userIndex - 1

	// Validate index is in range
	if internalIndex < 0 || internalIndex >= len(entries) {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Index %d is out of range\n", userIndex)
		_, _ = fmt.Fprintf(deps.Stderr, "Valid range: 1-%d (%d %s available)\n", len(entries), len(entries), pluralize("entry", len(entries)))
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: List entries with 'did' to see all indices")
		deps.Exit(1)
		return
	}

	// Get the entry to modify
	e := entries[internalIndex]

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
	if err := storage.UpdateEntry(storagePath, internalIndex, e); err != nil {
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
