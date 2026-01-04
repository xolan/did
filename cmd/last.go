package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/timeutil"
)

// lastCmd represents the relative date range query command
var lastCmd = &cobra.Command{
	Use:   "last <n> days",
	Short: "List entries for the last N days",
	Long: `List time tracking entries for the last N days.

The range includes N complete days ending today (inclusive).

Examples:
  did last 7 days     # List entries from the past 7 days
  did last 30 days    # List entries from the past 30 days
  did last 1 day      # List entries for today only`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleLastCommand(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(lastCmd)
}

// handleLastCommand processes the 'last' command arguments and lists entries
func handleLastCommand(cmd *cobra.Command, args []string) {
	// Join all args to form the full expression
	// Need to prepend "last" since it's consumed by the command framework
	expression := "last " + strings.Join(args, " ")
	expression = strings.TrimSpace(expression)

	// Parse the relative date expression using the parser
	startDate, endDate, err := timeutil.ParseRelativeDays(expression)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr)
		_, _ = fmt.Fprintln(deps.Stderr, "Examples:")
		_, _ = fmt.Fprintln(deps.Stderr, "  did last 7 days    # Past week")
		_, _ = fmt.Fprintln(deps.Stderr, "  did last 30 days   # Past month")
		_, _ = fmt.Fprintln(deps.Stderr, "  did last 1 day     # Today only")
		deps.Exit(1)
		return
	}

	// Format period description for display
	periodDesc := formatDateRangeForDisplay(startDate, endDate)

	// List entries using the custom time range
	listEntriesForRange(cmd, periodDesc, startDate, endDate)
}
