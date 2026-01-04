package cmd

import (
	"github.com/spf13/cobra"
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

func init() {
	rootCmd.AddCommand(exportCmd)
}
