package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// exportCmd represents the export parent command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export time entries to various formats",
	Long: `Export time entries to various formats for programmatic use, backup, or migration.

Available formats:
  json    Export entries as JSON
  csv     Export entries as CSV

Examples:
  did export json                Export all entries as JSON
  did export json > backup.json  Export to file
  did export csv                 Export all entries as CSV
  did export csv > entries.csv   Export to file`,
}

// exportJSONCmd represents the export json command
var exportJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Export time entries as JSON",
	Long: `Export all time entries to JSON format.

Output includes metadata (export timestamp, total entries, filter criteria)
and an array of entry objects.

Date Filtering:
  Use --from and --to to filter by date range
  Use --last to filter by relative days (e.g., 'last 7 days')

Project and Tag Filtering:
  Use --project to filter by project
  Use --tag to filter by tags (can be repeated)
  Use @project shorthand for --project
  Use #tag shorthand for --tag

Examples:
  did export json                          Export all entries as JSON
  did export json > backup.json            Export to file
  did export json --from 2024-01-01        Export from a specific date
  did export json --from 2024-01-01 --to 2024-01-31    Export within date range
  did export json --last 7                 Export last 7 days
  did export json --project acme           Export entries for project 'acme'
  did export json --tag review             Export entries tagged 'review'
  did export json @acme #review            Export using shorthand syntax
  did export json --last 30 --project acme Export last 30 days for project`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		_ = parseShorthandFilters(cmd, args)
		exportJSON(cmd)
	},
}

// exportCSVCmd represents the export csv command
var exportCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Export time entries as CSV",
	Long: `Export all time entries to CSV format.

Output is in standard CSV format with headers.

Date Filtering:
  Use --from and --to to filter by date range
  Use --last to filter by relative days (e.g., 'last 7 days')

Project and Tag Filtering:
  Use --project to filter by project
  Use --tag to filter by tags (can be repeated)
  Use @project shorthand for --project
  Use #tag shorthand for --tag

Examples:
  did export csv                           Export all entries as CSV
  did export csv > entries.csv             Export to file
  did export csv --from 2024-01-01         Export from a specific date
  did export csv --from 2024-01-01 --to 2024-01-31     Export within date range
  did export csv --last 7                  Export last 7 days
  did export csv --project acme            Export entries for project 'acme'
  did export csv --tag review              Export entries tagged 'review'
  did export csv @acme #review             Export using shorthand syntax
  did export csv --last 30 --project acme  Export last 30 days for project`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		_ = parseShorthandFilters(cmd, args)
		exportCSV(cmd)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(exportJSONCmd)
	exportCmd.AddCommand(exportCSVCmd)

	// Date filtering flags for JSON export
	exportJSONCmd.Flags().String("from", "", "Start date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	exportJSONCmd.Flags().String("to", "", "End date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	exportJSONCmd.Flags().Int("last", 0, "Filter by last N days (e.g., --last 7 for last 7 days)")

	// Date filtering flags for CSV export
	exportCSVCmd.Flags().String("from", "", "Start date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	exportCSVCmd.Flags().String("to", "", "End date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	exportCSVCmd.Flags().Int("last", 0, "Filter by last N days (e.g., --last 7 for last 7 days)")

	// Note: --project and --tag flags are inherited from root command's PersistentFlags
}

// exportJSON handles the export json command logic
func exportJSON(cmd *cobra.Command) {
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

	// Apply date filtering if specified
	entries := result.Entries
	if hasDateFilter {
		filtered := make([]entry.Entry, 0)
		for _, e := range entries {
			if timeutil.IsInRange(e.Timestamp, startDate, endDate) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	// Get project and tag filter flags from root persistent flags
	projectFilter, _ := cmd.Root().PersistentFlags().GetString("project")
	tagFilters, _ := cmd.Root().PersistentFlags().GetStringSlice("tag")

	// Apply project and tag filters if specified
	f := filter.NewFilter("", projectFilter, tagFilters)
	if !f.IsEmpty() {
		entries = filter.FilterEntries(entries, f)
	}

	// Create output structure with metadata
	output := struct {
		Metadata struct {
			ExportTimestamp time.Time              `json:"export_timestamp"`
			TotalEntries    int                    `json:"total_entries"`
			FilterCriteria  map[string]interface{} `json:"filter_criteria"`
		} `json:"metadata"`
		Entries []entry.Entry `json:"entries"`
	}{}

	output.Metadata.ExportTimestamp = time.Now()
	output.Metadata.TotalEntries = len(entries)
	output.Metadata.FilterCriteria = make(map[string]interface{})

	// Add date filter criteria to metadata if applicable
	if hasDateFilter {
		if lastDays > 0 {
			output.Metadata.FilterCriteria["last_days"] = lastDays
		} else {
			if fromStr != "" {
				output.Metadata.FilterCriteria["from"] = startDate.Format("2006-01-02")
			}
			if toStr != "" {
				output.Metadata.FilterCriteria["to"] = endDate.Format("2006-01-02")
			}
		}
	}

	// Add project and tag filter criteria to metadata if applicable
	if projectFilter != "" {
		output.Metadata.FilterCriteria["project"] = projectFilter
	}
	if len(tagFilters) > 0 {
		output.Metadata.FilterCriteria["tags"] = tagFilters
	}

	output.Entries = entries

	// Encode to JSON with pretty printing
	encoder := json.NewEncoder(deps.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to encode JSON output")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return
	}
}

// exportCSV handles the export csv command logic
func exportCSV(cmd *cobra.Command) {
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

	// Apply date filtering if specified
	entries := result.Entries
	if hasDateFilter {
		filtered := make([]entry.Entry, 0)
		for _, e := range entries {
			if timeutil.IsInRange(e.Timestamp, startDate, endDate) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	// Get project and tag filter flags from root persistent flags
	projectFilter, _ := cmd.Root().PersistentFlags().GetString("project")
	tagFilters, _ := cmd.Root().PersistentFlags().GetStringSlice("tag")

	// Apply project and tag filters if specified
	f := filter.NewFilter("", projectFilter, tagFilters)
	if !f.IsEmpty() {
		entries = filter.FilterEntries(entries, f)
	}

	// Create CSV writer
	writer := csv.NewWriter(deps.Stdout)
	defer writer.Flush()

	// Write CSV headers
	headers := []string{"date", "description", "duration_minutes", "duration_hours", "project", "tags"}
	if err := writer.Write(headers); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to write CSV headers")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return
	}

	// Write each entry as a CSV row
	for _, e := range entries {
		// Format date as YYYY-MM-DD
		date := e.Timestamp.Format("2006-01-02")

		// Format duration in hours as decimal
		durationHours := strconv.FormatFloat(float64(e.DurationMinutes)/60.0, 'f', 2, 64)

		// Format tags as semicolon-separated string
		tagsStr := strings.Join(e.Tags, ";")

		// Create row
		row := []string{
			date,
			e.Description,
			strconv.Itoa(e.DurationMinutes),
			durationHours,
			e.Project,
			tagsStr,
		}

		// Write row
		if err := writer.Write(row); err != nil {
			_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to write CSV row")
			_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
			deps.Exit(1)
			return
		}
	}

	// Ensure all buffered data is written
	writer.Flush()
	if err := writer.Error(); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to flush CSV output")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return
	}
}
