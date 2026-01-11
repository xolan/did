// Package cli provides the CLI presentation layer for the did application.
// It handles command-line output formatting and user interaction.
package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/storage"
)

// FormatDuration formats minutes as a human-readable string
// Examples: "30m", "2h", "1h 30m"
func FormatDuration(minutes int) string {
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

// FormatProjectAndTags formats project and tags for display.
// Returns format like: "@project" or "#tag1 #tag2" or "@project #tag1 #tag2"
// Returns empty string if no project or tags.
func FormatProjectAndTags(project string, tags []string) string {
	if project == "" && len(tags) == 0 {
		return ""
	}

	var parts []string
	if project != "" {
		parts = append(parts, "@"+project)
	}
	for _, tag := range tags {
		parts = append(parts, "#"+tag)
	}

	return strings.Join(parts, " ")
}

// FormatEntryForLog formats a description with optional project and tags for display.
// Returns format like: "description" or "description [@project]" or "description [#tag1 #tag2]"
// or "description [@project #tag1 #tag2]"
func FormatEntryForLog(description, project string, tags []string) string {
	metadata := FormatProjectAndTags(project, tags)
	if metadata == "" {
		return description
	}
	return fmt.Sprintf("%s [%s]", description, metadata)
}

// FormatEntry formats an entry for display
func FormatEntry(e entry.Entry) string {
	return FormatEntryForLog(e.Description, e.Project, e.Tags)
}

// FormatElapsedTime formats a duration as human-readable elapsed time
// Examples: "5m", "1h 23m", "2h"
func FormatElapsedTime(d time.Duration) string {
	totalMinutes := int(d.Minutes())
	if totalMinutes < 60 {
		return fmt.Sprintf("%dm", totalMinutes)
	}
	hours := totalMinutes / 60
	mins := totalMinutes % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

// FormatDateRangeForDisplay formats a date range for human-readable display.
func FormatDateRangeForDisplay(start, end time.Time) string {
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		return start.Format("Mon, Jan 2, 2006")
	}
	if start.Year() == end.Year() {
		return fmt.Sprintf("%s - %s", start.Format("Jan 2"), end.Format("Jan 2, 2006"))
	}
	return fmt.Sprintf("%s - %s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
}

// FormatCorruptionWarning formats a ParseWarning into a human-readable string
func FormatCorruptionWarning(warning storage.ParseWarning) string {
	content := warning.Content
	if len(content) > 50 {
		content = content[:47] + "..."
	}
	return fmt.Sprintf("  Line %d: %s (error: %s)", warning.LineNumber, content, warning.Error)
}

// BuildPeriodWithFilters appends filter information to the period description.
// Example: "today" -> "today (@acme #bugfix)"
func BuildPeriodWithFilters(period, project string, tags []string) string {
	if project == "" && len(tags) == 0 {
		return period
	}

	var filters []string
	if project != "" {
		filters = append(filters, "@"+project)
	}
	for _, tag := range tags {
		filters = append(filters, "#"+tag)
	}

	return fmt.Sprintf("%s (%s)", period, strings.Join(filters, " "))
}

// Pluralize returns the singular or plural form of a word based on count
func Pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

// SpansMultipleDays checks if entries span multiple calendar days
func SpansMultipleDays(entries []entry.Entry) bool {
	if len(entries) < 2 {
		return false
	}
	firstDay := entries[0].Timestamp.Format("2006-01-02")
	for _, e := range entries[1:] {
		if e.Timestamp.Format("2006-01-02") != firstDay {
			return true
		}
	}
	return false
}

// SpansMultipleDaysIndexed checks if indexed entries span multiple calendar days
func SpansMultipleDaysIndexed(entries []service.IndexedEntry) bool {
	if len(entries) < 2 {
		return false
	}
	firstDay := entries[0].Entry.Timestamp.Format("2006-01-02")
	for _, ie := range entries[1:] {
		if ie.Entry.Timestamp.Format("2006-01-02") != firstDay {
			return true
		}
	}
	return false
}

// FormatTimerStartTime formats the timer start time for display
func FormatTimerStartTime(startedAt time.Time) string {
	now := time.Now()
	startTime := startedAt.Format("3:04 PM")

	isToday := startedAt.Year() == now.Year() &&
		startedAt.Month() == now.Month() &&
		startedAt.Day() == now.Day()

	if isToday {
		return fmt.Sprintf("today at %s", startTime)
	}
	return fmt.Sprintf("%s at %s", startedAt.Format("Mon Jan 2"), startTime)
}
