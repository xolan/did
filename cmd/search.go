package cmd

import (
	"github.com/spf13/cobra"
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
	// TODO: Implement search functionality in subtask 2.2
}
