package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
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

Examples:
  did export json                Export all entries as JSON
  did export json > backup.json  Export to file`,
	Run: func(cmd *cobra.Command, args []string) {
		exportJSON(cmd)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(exportJSONCmd)
}

// exportJSON handles the export json command logic
func exportJSON(cmd *cobra.Command) {
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

	// Create output structure with metadata
	output := struct {
		Metadata struct {
			ExportTimestamp time.Time `json:"export_timestamp"`
			TotalEntries    int       `json:"total_entries"`
			FilterCriteria  struct{}  `json:"filter_criteria"`
		} `json:"metadata"`
		Entries []entry.Entry `json:"entries"`
	}{}

	output.Metadata.ExportTimestamp = time.Now()
	output.Metadata.TotalEntries = len(result.Entries)
	output.Entries = result.Entries

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
