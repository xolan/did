package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
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

Examples:
  did export json                Export all entries as JSON
  did export json > backup.json  Export to file`,
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

Examples:
  did export json                          Export all entries as JSON
  did export json > backup.json            Export to file
  did export json --from 2024-01-01        Export from a specific date
  did export json --from 2024-01-01 --to 2024-01-31    Export within date range
  did export json --last 7                 Export last 7 days`,
	Run: func(cmd *cobra.Command, args []string) {
		exportJSON(cmd)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(exportJSONCmd)

	// Date filtering flags
	exportJSONCmd.Flags().String("from", "", "Start date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	exportJSONCmd.Flags().String("to", "", "End date for filtering (YYYY-MM-DD or DD/MM/YYYY)")
	exportJSONCmd.Flags().Int("last", 0, "Filter by last N days (e.g., --last 7 for last 7 days)")
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
