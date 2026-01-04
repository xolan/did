package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/timeutil"
)

// fromCmd represents the date range query command
var fromCmd = &cobra.Command{
	Use:   "from <start-date> to <end-date>",
	Short: "List entries for a custom date range",
	Long: `List time tracking entries for a custom date range.

Supported date formats:
  - YYYY-MM-DD (e.g., 2024-01-15)
  - DD/MM/YYYY (e.g., 15/01/2024)

Examples:
  did from 2024-01-01 to 2024-01-31    # List entries for January 2024
  did from 15/01/2024 to 31/01/2024    # Same range with European format
  did from 2024-01-15 to 2024-01-15    # List entries for a single day`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleFromCommand(args)
	},
}

func init() {
	rootCmd.AddCommand(fromCmd)
}

// handleFromCommand processes the 'from' command arguments and lists entries
func handleFromCommand(args []string) {
	// Parse arguments: expect "from START to END" or just "START to END"
	// Since the command is "did from ...", the first arg is already after "from"

	// Find "to" keyword in arguments
	toIndex := -1
	for i, arg := range args {
		if strings.ToLower(arg) == "to" {
			toIndex = i
			break
		}
	}

	if toIndex == -1 {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Missing 'to' keyword in date range")
		_, _ = fmt.Fprintln(deps.Stderr, "Usage: did from <start-date> to <end-date>")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: did from 2024-01-01 to 2024-01-31")
		deps.Exit(1)
		return
	}

	if toIndex == 0 {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Missing start date")
		_, _ = fmt.Fprintln(deps.Stderr, "Usage: did from <start-date> to <end-date>")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: did from 2024-01-01 to 2024-01-31")
		deps.Exit(1)
		return
	}

	if toIndex >= len(args)-1 {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Missing end date")
		_, _ = fmt.Fprintln(deps.Stderr, "Usage: did from <start-date> to <end-date>")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: did from 2024-01-01 to 2024-01-31")
		deps.Exit(1)
		return
	}

	// Extract start and end date strings
	startDateStr := strings.Join(args[:toIndex], " ")
	endDateStr := strings.Join(args[toIndex+1:], " ")

	// Trim any whitespace
	startDateStr = strings.TrimSpace(startDateStr)
	endDateStr = strings.TrimSpace(endDateStr)

	// Parse start date
	startDate, err := timeutil.ParseDate(startDateStr)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid start date '%s'\n", startDateStr)
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Use format YYYY-MM-DD or DD/MM/YYYY")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: 2024-01-15 or 15/01/2024")
		deps.Exit(1)
		return
	}

	// Parse end date
	endDate, err := timeutil.ParseDate(endDateStr)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid end date '%s'\n", endDateStr)
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Use format YYYY-MM-DD or DD/MM/YYYY")
		_, _ = fmt.Fprintln(deps.Stderr, "Example: 2024-01-31 or 31/01/2024")
		deps.Exit(1)
		return
	}

	// Validate that start date is not after end date
	if startDate.After(endDate) {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Start date (%s) is after end date (%s)\n",
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"))
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Ensure start date comes before or equals end date")
		deps.Exit(1)
		return
	}

	// Get end of the end date to include the full day
	endDate = timeutil.EndOfDay(endDate)

	// Format period description for display
	periodDesc := formatDateRangeForDisplay(startDate, endDate)

	// List entries using the custom time range
	listEntriesForRange(periodDesc, startDate, endDate)
}
