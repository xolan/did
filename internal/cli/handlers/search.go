package handlers

import (
	"fmt"
	"strings"

	"github.com/xolan/did/internal/cli"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/service"
)

// Search searches entries by keyword
func Search(deps *cli.Deps, keyword string, dateRange *service.DateRangeSpec, f *filter.Filter) {
	result, err := deps.Services.Search.Search(keyword, dateRange, f)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	// Display warnings about corrupted lines
	if len(result.Warnings) > 0 {
		_, _ = fmt.Fprintf(deps.Stderr, "Warning: Found %d corrupted line(s) in storage file:\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			_, _ = fmt.Fprintln(deps.Stderr, cli.FormatCorruptionWarning(warning))
		}
		_, _ = fmt.Fprintln(deps.Stderr)
	}

	if len(result.Entries) == 0 {
		if keyword != "" {
			_, _ = fmt.Fprintf(deps.Stdout, "No entries found matching '%s'\n", keyword)
		} else {
			_, _ = fmt.Fprintln(deps.Stdout, "No entries found")
		}
		return
	}

	// Build header
	header := fmt.Sprintf("Search results for '%s'", keyword)
	if keyword == "" {
		header = "All entries"
	}

	_, _ = fmt.Fprintf(deps.Stdout, "%s (%d %s):\n", header, result.Total, cli.Pluralize("result", result.Total))
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))

	// Calculate max index width for alignment
	maxIndex := 0
	for _, ie := range result.Entries {
		if ie.ActiveIndex > maxIndex {
			maxIndex = ie.ActiveIndex
		}
	}
	maxIndexWidth := len(fmt.Sprintf("%d", maxIndex))

	for _, ie := range result.Entries {
		e := ie.Entry
		_, _ = fmt.Fprintf(deps.Stdout, "[%*d] %s %s  %s (%s)\n",
			maxIndexWidth,
			ie.ActiveIndex,
			e.Timestamp.Format("2006-01-02"),
			e.Timestamp.Format("15:04"),
			cli.FormatEntryForLog(e.Description, e.Project, e.Tags),
			cli.FormatDuration(e.DurationMinutes))
	}
}
