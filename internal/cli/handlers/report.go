package handlers

import (
	"fmt"
	"strings"

	"github.com/xolan/did/internal/cli"
	"github.com/xolan/did/internal/service"
)

// ReportByProject shows a report for a specific project
func ReportByProject(deps *cli.Deps, project string, dateRange service.DateRangeSpec) {
	result, err := deps.Services.Report.ByProject(project, dateRange)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	_, _ = fmt.Fprintf(deps.Stdout, "Report for @%s (%s):\n", project, result.Period)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "Total time:    %s\n", cli.FormatDuration(result.TotalMinutes))
	_, _ = fmt.Fprintf(deps.Stdout, "Total entries: %d %s\n", result.EntryCount, cli.Pluralize("entry", result.EntryCount))
}

// ReportByTags shows a report for specific tags
func ReportByTags(deps *cli.Deps, tags []string, dateRange service.DateRangeSpec) {
	result, err := deps.Services.Report.ByTags(tags, dateRange)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	tagStr := "#" + strings.Join(tags, " #")
	_, _ = fmt.Fprintf(deps.Stdout, "Report for %s (%s):\n", tagStr, result.Period)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "Total time:    %s\n", cli.FormatDuration(result.TotalMinutes))
	_, _ = fmt.Fprintf(deps.Stdout, "Total entries: %d %s\n", result.EntryCount, cli.Pluralize("entry", result.EntryCount))
}

// ReportGroupByProject shows entries grouped by project
func ReportGroupByProject(deps *cli.Deps, dateRange service.DateRangeSpec) {
	result, err := deps.Services.Report.GroupByProject(dateRange)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	if len(result.Groups) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No entries found for %s\n", result.Period)
		return
	}

	_, _ = fmt.Fprintf(deps.Stdout, "Time by project (%s):\n", result.Period)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))

	for _, group := range result.Groups {
		projectDisplay := group.Name
		if group.Name != "(no project)" {
			projectDisplay = "@" + group.Name
		}
		_, _ = fmt.Fprintf(deps.Stdout, "  %-28s  %10s  (%d %s)\n",
			projectDisplay,
			cli.FormatDuration(group.TotalMinutes),
			group.EntryCount,
			cli.Pluralize("entry", group.EntryCount))
	}

	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "  %-28s  %10s  (%d %s)\n",
		"Total",
		cli.FormatDuration(result.TotalMinutes),
		result.EntryCount,
		cli.Pluralize("entry", result.EntryCount))
}

// ReportGroupByTag shows entries grouped by tag
func ReportGroupByTag(deps *cli.Deps, dateRange service.DateRangeSpec) {
	result, err := deps.Services.Report.GroupByTag(dateRange)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	if len(result.Groups) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No entries found for %s\n", result.Period)
		return
	}

	_, _ = fmt.Fprintf(deps.Stdout, "Time by tag (%s):\n", result.Period)
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))

	for _, group := range result.Groups {
		tagDisplay := group.Name
		if group.Name != "(no tags)" {
			tagDisplay = "#" + group.Name
		}
		_, _ = fmt.Fprintf(deps.Stdout, "  %-28s  %10s  (%d %s)\n",
			tagDisplay,
			cli.FormatDuration(group.TotalMinutes),
			group.EntryCount,
			cli.Pluralize("entry", group.EntryCount))
	}

	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "  %-28s  %10s  (%d %s)\n",
		"Total",
		cli.FormatDuration(result.TotalMinutes),
		result.EntryCount,
		cli.Pluralize("entry", result.EntryCount))
}
