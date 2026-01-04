package stats

import (
	"sort"
	"time"

	"github.com/xolan/did/internal/entry"
)

// Statistics contains aggregated statistics for a set of entries
type Statistics struct {
	TotalMinutes         int
	AverageMinutesPerDay float64
	EntryCount           int
	DaysWithEntries      int
}

// ProjectBreakdown contains statistics for a single project
type ProjectBreakdown struct {
	Project      string
	TotalMinutes int
	EntryCount   int
}

// TagBreakdown contains statistics for a single tag
type TagBreakdown struct {
	Tag          string
	TotalMinutes int
	EntryCount   int
}

// CalculateStatistics computes statistics for entries within the given date range
func CalculateStatistics(entries []entry.Entry, start, end time.Time) Statistics {
	stats := Statistics{}

	if len(entries) == 0 {
		return stats
	}

	// Track which days have entries
	daysWithEntries := make(map[string]bool)

	for _, e := range entries {
		// Skip deleted entries
		if e.DeletedAt != nil {
			continue
		}

		// Check if entry is within the date range
		if (e.Timestamp.Equal(start) || e.Timestamp.After(start)) &&
			(e.Timestamp.Equal(end) || e.Timestamp.Before(end)) {
			stats.TotalMinutes += e.DurationMinutes
			stats.EntryCount++

			// Track the day this entry was logged
			dayKey := e.Timestamp.Format("2006-01-02")
			daysWithEntries[dayKey] = true
		}
	}

	stats.DaysWithEntries = len(daysWithEntries)

	// Calculate average minutes per day based on the total days in the range
	totalDays := int(end.Sub(start).Hours()/24) + 1
	if totalDays > 0 {
		stats.AverageMinutesPerDay = float64(stats.TotalMinutes) / float64(totalDays)
	}

	return stats
}

// CalculateProjectBreakdown groups entries by project and returns breakdown sorted by total minutes
func CalculateProjectBreakdown(entries []entry.Entry, start, end time.Time) []ProjectBreakdown {
	if len(entries) == 0 {
		return []ProjectBreakdown{}
	}

	// Group entries by project
	projectMap := make(map[string]*ProjectBreakdown)

	for _, e := range entries {
		// Skip deleted entries
		if e.DeletedAt != nil {
			continue
		}

		// Check if entry is within the date range
		if (e.Timestamp.Equal(start) || e.Timestamp.After(start)) &&
			(e.Timestamp.Equal(end) || e.Timestamp.Before(end)) {

			// Determine project name
			projectName := e.Project
			if projectName == "" {
				projectName = "(no project)"
			}

			// Initialize project breakdown if not exists
			if _, exists := projectMap[projectName]; !exists {
				projectMap[projectName] = &ProjectBreakdown{
					Project: projectName,
				}
			}

			// Accumulate totals
			projectMap[projectName].TotalMinutes += e.DurationMinutes
			projectMap[projectName].EntryCount++
		}
	}

	// Convert map to slice
	var breakdowns []ProjectBreakdown
	for _, breakdown := range projectMap {
		breakdowns = append(breakdowns, *breakdown)
	}

	// Sort by total minutes descending
	sort.Slice(breakdowns, func(i, j int) bool {
		return breakdowns[i].TotalMinutes > breakdowns[j].TotalMinutes
	})

	return breakdowns
}

// CalculateTagBreakdown groups entries by tag and returns breakdown sorted by total minutes
// Entries with multiple tags will contribute to each tag.
func CalculateTagBreakdown(entries []entry.Entry, start, end time.Time) []TagBreakdown {
	if len(entries) == 0 {
		return []TagBreakdown{}
	}

	// Group entries by tag
	tagMap := make(map[string]*TagBreakdown)

	for _, e := range entries {
		// Skip deleted entries
		if e.DeletedAt != nil {
			continue
		}

		// Check if entry is within the date range
		if (e.Timestamp.Equal(start) || e.Timestamp.After(start)) &&
			(e.Timestamp.Equal(end) || e.Timestamp.Before(end)) {

			// If entry has no tags, add to "(no tags)" group
			if len(e.Tags) == 0 {
				tagName := "(no tags)"
				if _, exists := tagMap[tagName]; !exists {
					tagMap[tagName] = &TagBreakdown{
						Tag: tagName,
					}
				}
				tagMap[tagName].TotalMinutes += e.DurationMinutes
				tagMap[tagName].EntryCount++
			} else {
				// Entry has tags - add to each tag group
				for _, tag := range e.Tags {
					if _, exists := tagMap[tag]; !exists {
						tagMap[tag] = &TagBreakdown{
							Tag: tag,
						}
					}
					tagMap[tag].TotalMinutes += e.DurationMinutes
					tagMap[tag].EntryCount++
				}
			}
		}
	}

	// Convert map to slice
	var breakdowns []TagBreakdown
	for _, breakdown := range tagMap {
		breakdowns = append(breakdowns, *breakdown)
	}

	// Sort by total minutes descending
	sort.Slice(breakdowns, func(i, j int) bool {
		return breakdowns[i].TotalMinutes > breakdowns[j].TotalMinutes
	})

	return breakdowns
}
