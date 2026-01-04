package stats

import (
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
