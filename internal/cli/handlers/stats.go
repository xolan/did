package handlers

import (
	"fmt"
	"strings"

	"github.com/xolan/did/internal/cli"
)

// ShowWeeklyStats shows weekly statistics
func ShowWeeklyStats(deps *cli.Deps) {
	result, err := deps.Services.Stats.Weekly()
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	displayStats(deps, result.Statistics.TotalMinutes, result.Statistics.EntryCount,
		result.Statistics.DaysWithEntries, result.Statistics.AverageMinutesPerDay,
		result.Period, result.Comparison, result.ProjectStats, result.TagStats)
}

// ShowMonthlyStats shows monthly statistics
func ShowMonthlyStats(deps *cli.Deps) {
	result, err := deps.Services.Stats.Monthly()
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	displayStats(deps, result.Statistics.TotalMinutes, result.Statistics.EntryCount,
		result.Statistics.DaysWithEntries, result.Statistics.AverageMinutesPerDay,
		result.Period, result.Comparison, result.ProjectStats, result.TagStats)
}

func displayStats(deps *cli.Deps, totalMinutes, entryCount, daysWithEntries int, avgPerDay float64,
	period, comparison string, projectStats interface{}, tagStats interface{}) {

	_, _ = fmt.Fprintf(deps.Stdout, "Statistics for %s:\n", period)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "Total time:      %s\n", cli.FormatDuration(totalMinutes))
	_, _ = fmt.Fprintf(deps.Stdout, "Total entries:   %d %s\n", entryCount, cli.Pluralize("entry", entryCount))
	_, _ = fmt.Fprintf(deps.Stdout, "Days with work:  %d %s\n", daysWithEntries, cli.Pluralize("day", daysWithEntries))
	_, _ = fmt.Fprintf(deps.Stdout, "Average per day: %s\n", cli.FormatDuration(int(avgPerDay)))

	if comparison != "" {
		_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))
		_, _ = fmt.Fprintf(deps.Stdout, "Comparison: %s\n", comparison)
	}
}
