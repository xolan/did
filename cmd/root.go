package cmd

import (
	"fmt"
	"os"
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
  did <description> for <duration>    Log a new entry (e.g., did feature X for 2h)
  did                                 List today's entries
  did y                               List yesterday's entries
  did w                               List this week's entries
  did lw                              List last week's entries

Duration format: Yh (hours), Ym (minutes), or YhYm (combined)
Examples: 2h, 30m, 1h30m`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// No args: list today's entries
			listEntries("today", timeutil.Today)
			return
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
		listEntries("yesterday", timeutil.Yesterday)
	},
}

// wCmd represents the this week command
var wCmd = &cobra.Command{
	Use:   "w",
	Short: "List this week's entries",
	Long:  `List all time tracking entries logged this week (Monday-Sunday).`,
	Run: func(cmd *cobra.Command, args []string) {
		listEntries("this week", timeutil.ThisWeek)
	},
}

// lwCmd represents the last week command
var lwCmd = &cobra.Command{
	Use:   "lw",
	Short: "List last week's entries",
	Long:  `List all time tracking entries logged last week (Monday-Sunday).`,
	Run: func(cmd *cobra.Command, args []string) {
		listEntries("last week", timeutil.LastWeek)
	},
}

func init() {
	rootCmd.AddCommand(yCmd)
	rootCmd.AddCommand(wCmd)
	rootCmd.AddCommand(lwCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// createEntry parses arguments and creates a new time tracking entry
func createEntry(args []string) {
	// Join all arguments to form the raw input
	rawInput := strings.Join(args, " ")

	// Parse the input: expected format "<description> for <duration>"
	// Find the last "for" in the input to extract duration
	lastForIdx := strings.LastIndex(strings.ToLower(rawInput), " for ")
	if lastForIdx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid format. Use: did <description> for <duration>")
		fmt.Fprintln(os.Stderr, "Example: did feature X for 2h")
		os.Exit(1)
	}

	description := strings.TrimSpace(rawInput[:lastForIdx])
	durationStr := strings.TrimSpace(rawInput[lastForIdx+5:]) // +5 for " for "

	if description == "" {
		fmt.Fprintln(os.Stderr, "Error: Description cannot be empty")
		os.Exit(1)
	}

	// Parse the duration
	minutes, err := entry.ParseDuration(durationStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create the entry
	e := entry.Entry{
		Timestamp:       time.Now(),
		Description:     description,
		DurationMinutes: minutes,
		RawInput:        rawInput,
	}

	// Get storage path
	storagePath, err := storage.GetStoragePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get storage path: %v\n", err)
		os.Exit(1)
	}

	// Append the entry to storage
	if err := storage.AppendEntry(storagePath, e); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to save entry: %v\n", err)
		os.Exit(1)
	}

	// Display success message
	fmt.Printf("Logged: %s (%s)\n", description, formatDuration(minutes))
}

// listEntries reads and displays entries filtered by the given time range
func listEntries(period string, timeRangeFunc func() (time.Time, time.Time)) {
	storagePath, err := storage.GetStoragePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get storage path: %v\n", err)
		os.Exit(1)
	}

	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read entries: %v\n", err)
		os.Exit(1)
	}

	start, end := timeRangeFunc()

	// Filter entries by time range
	var filtered []entry.Entry
	for _, e := range entries {
		if timeutil.IsInRange(e.Timestamp, start, end) {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("No entries found for %s\n", period)
		return
	}

	// Calculate total duration
	totalMinutes := 0
	for _, e := range filtered {
		totalMinutes += e.DurationMinutes
	}

	// Display entries
	fmt.Printf("Entries for %s:\n", period)
	fmt.Println(strings.Repeat("-", 50))
	for _, e := range filtered {
		fmt.Printf("  %s  %s (%s)\n",
			e.Timestamp.Format("15:04"),
			e.Description,
			formatDuration(e.DurationMinutes))
	}
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Total: %s\n", formatDuration(totalMinutes))
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
