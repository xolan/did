package stats

import (
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
)

// Helper function to create test times with specific dates
func makeTime(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.Local)
}

// Helper function to create an entry
func makeEntry(timestamp time.Time, durationMinutes int, description string) entry.Entry {
	return entry.Entry{
		Timestamp:       timestamp,
		DurationMinutes: durationMinutes,
		Description:     description,
		RawInput:        description,
	}
}

// Helper function to create a deleted entry
func makeDeletedEntry(timestamp time.Time, durationMinutes int, description string) entry.Entry {
	deletedAt := timestamp.Add(time.Hour)
	return entry.Entry{
		Timestamp:       timestamp,
		DurationMinutes: durationMinutes,
		Description:     description,
		RawInput:        description,
		DeletedAt:       &deletedAt,
	}
}

func TestCalculateStatistics_EmptyEntries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	stats := CalculateStatistics([]entry.Entry{}, start, end)

	if stats.TotalMinutes != 0 {
		t.Errorf("TotalMinutes = %d, expected 0", stats.TotalMinutes)
	}
	if stats.AverageMinutesPerDay != 0 {
		t.Errorf("AverageMinutesPerDay = %f, expected 0", stats.AverageMinutesPerDay)
	}
	if stats.EntryCount != 0 {
		t.Errorf("EntryCount = %d, expected 0", stats.EntryCount)
	}
	if stats.DaysWithEntries != 0 {
		t.Errorf("DaysWithEntries = %d, expected 0", stats.DaysWithEntries)
	}
}

func TestCalculateStatistics_SingleEntry(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		makeEntry(makeTime(2024, time.January, 16, 10, 0, 0), 120, "worked on feature"),
	}

	stats := CalculateStatistics(entries, start, end)

	if stats.TotalMinutes != 120 {
		t.Errorf("TotalMinutes = %d, expected 120", stats.TotalMinutes)
	}
	if stats.EntryCount != 1 {
		t.Errorf("EntryCount = %d, expected 1", stats.EntryCount)
	}
	if stats.DaysWithEntries != 1 {
		t.Errorf("DaysWithEntries = %d, expected 1", stats.DaysWithEntries)
	}

	// 7 days in range (Jan 15-21), 120 minutes total
	expectedAvg := 120.0 / 7.0
	if stats.AverageMinutesPerDay != expectedAvg {
		t.Errorf("AverageMinutesPerDay = %f, expected %f", stats.AverageMinutesPerDay, expectedAvg)
	}
}

func TestCalculateStatistics_MultipleEntries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		makeEntry(makeTime(2024, time.January, 15, 9, 0, 0), 120, "worked on feature A"),
		makeEntry(makeTime(2024, time.January, 15, 14, 0, 0), 90, "worked on feature B"),
		makeEntry(makeTime(2024, time.January, 17, 10, 0, 0), 180, "meeting"),
		makeEntry(makeTime(2024, time.January, 19, 15, 0, 0), 60, "code review"),
	}

	stats := CalculateStatistics(entries, start, end)

	expectedTotal := 120 + 90 + 180 + 60
	if stats.TotalMinutes != expectedTotal {
		t.Errorf("TotalMinutes = %d, expected %d", stats.TotalMinutes, expectedTotal)
	}
	if stats.EntryCount != 4 {
		t.Errorf("EntryCount = %d, expected 4", stats.EntryCount)
	}
	if stats.DaysWithEntries != 3 {
		t.Errorf("DaysWithEntries = %d, expected 3 (Jan 15, 17, 19)", stats.DaysWithEntries)
	}

	// 7 days in range (Jan 15-21), 450 minutes total
	expectedAvg := 450.0 / 7.0
	if stats.AverageMinutesPerDay != expectedAvg {
		t.Errorf("AverageMinutesPerDay = %f, expected %f", stats.AverageMinutesPerDay, expectedAvg)
	}
}

func TestCalculateStatistics_SingleDay(t *testing.T) {
	// Test average calculation for a single day range
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 15, 23, 59, 59)

	entries := []entry.Entry{
		makeEntry(makeTime(2024, time.January, 15, 9, 0, 0), 120, "morning work"),
		makeEntry(makeTime(2024, time.January, 15, 14, 0, 0), 180, "afternoon work"),
	}

	stats := CalculateStatistics(entries, start, end)

	expectedTotal := 120 + 180
	if stats.TotalMinutes != expectedTotal {
		t.Errorf("TotalMinutes = %d, expected %d", stats.TotalMinutes, expectedTotal)
	}
	if stats.EntryCount != 2 {
		t.Errorf("EntryCount = %d, expected 2", stats.EntryCount)
	}
	if stats.DaysWithEntries != 1 {
		t.Errorf("DaysWithEntries = %d, expected 1", stats.DaysWithEntries)
	}

	// 1 day in range, 300 minutes total
	expectedAvg := 300.0
	if stats.AverageMinutesPerDay != expectedAvg {
		t.Errorf("AverageMinutesPerDay = %f, expected %f", stats.AverageMinutesPerDay, expectedAvg)
	}
}

func TestCalculateStatistics_MonthRange(t *testing.T) {
	// Test average calculation for a month range (31 days)
	start := makeTime(2024, time.January, 1, 0, 0, 0)
	end := makeTime(2024, time.January, 31, 23, 59, 59)

	entries := []entry.Entry{
		makeEntry(makeTime(2024, time.January, 5, 10, 0, 0), 480, "full day work"),
		makeEntry(makeTime(2024, time.January, 10, 10, 0, 0), 480, "full day work"),
		makeEntry(makeTime(2024, time.January, 15, 10, 0, 0), 480, "full day work"),
	}

	stats := CalculateStatistics(entries, start, end)

	expectedTotal := 480 * 3
	if stats.TotalMinutes != expectedTotal {
		t.Errorf("TotalMinutes = %d, expected %d", stats.TotalMinutes, expectedTotal)
	}
	if stats.EntryCount != 3 {
		t.Errorf("EntryCount = %d, expected 3", stats.EntryCount)
	}
	if stats.DaysWithEntries != 3 {
		t.Errorf("DaysWithEntries = %d, expected 3", stats.DaysWithEntries)
	}

	// 31 days in January, 1440 minutes total
	expectedAvg := 1440.0 / 31.0
	if stats.AverageMinutesPerDay != expectedAvg {
		t.Errorf("AverageMinutesPerDay = %f, expected %f", stats.AverageMinutesPerDay, expectedAvg)
	}
}

func TestCalculateStatistics_SkipsDeletedEntries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		makeEntry(makeTime(2024, time.January, 15, 9, 0, 0), 120, "active entry"),
		makeDeletedEntry(makeTime(2024, time.January, 16, 10, 0, 0), 180, "deleted entry"),
		makeEntry(makeTime(2024, time.January, 17, 11, 0, 0), 90, "another active entry"),
	}

	stats := CalculateStatistics(entries, start, end)

	// Should only count the non-deleted entries
	expectedTotal := 120 + 90
	if stats.TotalMinutes != expectedTotal {
		t.Errorf("TotalMinutes = %d, expected %d (deleted entries should be skipped)", stats.TotalMinutes, expectedTotal)
	}
	if stats.EntryCount != 2 {
		t.Errorf("EntryCount = %d, expected 2", stats.EntryCount)
	}
	if stats.DaysWithEntries != 2 {
		t.Errorf("DaysWithEntries = %d, expected 2", stats.DaysWithEntries)
	}
}

func TestCalculateStatistics_OnlyDeletedEntries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		makeDeletedEntry(makeTime(2024, time.January, 15, 9, 0, 0), 120, "deleted entry 1"),
		makeDeletedEntry(makeTime(2024, time.January, 16, 10, 0, 0), 180, "deleted entry 2"),
	}

	stats := CalculateStatistics(entries, start, end)

	// Should be empty since all entries are deleted
	if stats.TotalMinutes != 0 {
		t.Errorf("TotalMinutes = %d, expected 0", stats.TotalMinutes)
	}
	if stats.EntryCount != 0 {
		t.Errorf("EntryCount = %d, expected 0", stats.EntryCount)
	}
	if stats.DaysWithEntries != 0 {
		t.Errorf("DaysWithEntries = %d, expected 0", stats.DaysWithEntries)
	}
}

func TestCalculateStatistics_EntriesOutsideRange(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		makeEntry(makeTime(2024, time.January, 14, 23, 59, 59), 120, "before range"),
		makeEntry(makeTime(2024, time.January, 16, 10, 0, 0), 180, "in range"),
		makeEntry(makeTime(2024, time.January, 22, 0, 0, 0), 90, "after range"),
	}

	stats := CalculateStatistics(entries, start, end)

	// Should only count the entry within range
	if stats.TotalMinutes != 180 {
		t.Errorf("TotalMinutes = %d, expected 180", stats.TotalMinutes)
	}
	if stats.EntryCount != 1 {
		t.Errorf("EntryCount = %d, expected 1", stats.EntryCount)
	}
	if stats.DaysWithEntries != 1 {
		t.Errorf("DaysWithEntries = %d, expected 1", stats.DaysWithEntries)
	}
}

func TestCalculateStatistics_EntriesAtBoundaries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		makeEntry(start, 120, "exactly at start"),
		makeEntry(end, 180, "exactly at end"),
		makeEntry(makeTime(2024, time.January, 18, 12, 0, 0), 90, "in middle"),
	}

	stats := CalculateStatistics(entries, start, end)

	// All entries should be included (boundaries are inclusive)
	expectedTotal := 120 + 180 + 90
	if stats.TotalMinutes != expectedTotal {
		t.Errorf("TotalMinutes = %d, expected %d", stats.TotalMinutes, expectedTotal)
	}
	if stats.EntryCount != 3 {
		t.Errorf("EntryCount = %d, expected 3", stats.EntryCount)
	}
	if stats.DaysWithEntries != 3 {
		t.Errorf("DaysWithEntries = %d, expected 3", stats.DaysWithEntries)
	}
}

func TestCalculateStatistics_MultipleSameDay(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		makeEntry(makeTime(2024, time.January, 16, 9, 0, 0), 120, "morning work"),
		makeEntry(makeTime(2024, time.January, 16, 14, 0, 0), 180, "afternoon work"),
		makeEntry(makeTime(2024, time.January, 16, 18, 0, 0), 90, "evening work"),
		makeEntry(makeTime(2024, time.January, 18, 10, 0, 0), 60, "different day"),
	}

	stats := CalculateStatistics(entries, start, end)

	expectedTotal := 120 + 180 + 90 + 60
	if stats.TotalMinutes != expectedTotal {
		t.Errorf("TotalMinutes = %d, expected %d", stats.TotalMinutes, expectedTotal)
	}
	if stats.EntryCount != 4 {
		t.Errorf("EntryCount = %d, expected 4", stats.EntryCount)
	}
	// Should count only 2 unique days (Jan 16 and Jan 18)
	if stats.DaysWithEntries != 2 {
		t.Errorf("DaysWithEntries = %d, expected 2", stats.DaysWithEntries)
	}
}

func TestCalculateStatistics_DifferentDaySpans(t *testing.T) {
	tests := []struct {
		name        string
		start       time.Time
		end         time.Time
		minutes     int
		expectedAvg float64
	}{
		{
			name:        "1 day span",
			start:       makeTime(2024, time.January, 15, 0, 0, 0),
			end:         makeTime(2024, time.January, 15, 23, 59, 59),
			minutes:     300,
			expectedAvg: 300.0,
		},
		{
			name:        "7 day span (week)",
			start:       makeTime(2024, time.January, 15, 0, 0, 0),
			end:         makeTime(2024, time.January, 21, 23, 59, 59),
			minutes:     700,
			expectedAvg: 100.0,
		},
		{
			name:        "30 day span",
			start:       makeTime(2024, time.January, 1, 0, 0, 0),
			end:         makeTime(2024, time.January, 30, 23, 59, 59),
			minutes:     900,
			expectedAvg: 30.0,
		},
		{
			name:        "31 day span (full month)",
			start:       makeTime(2024, time.January, 1, 0, 0, 0),
			end:         makeTime(2024, time.January, 31, 23, 59, 59),
			minutes:     930,
			expectedAvg: 30.0,
		},
		{
			name:        "28 day span (February non-leap)",
			start:       makeTime(2023, time.February, 1, 0, 0, 0),
			end:         makeTime(2023, time.February, 28, 23, 59, 59),
			minutes:     560,
			expectedAvg: 20.0,
		},
		{
			name:        "29 day span (February leap year)",
			start:       makeTime(2024, time.February, 1, 0, 0, 0),
			end:         makeTime(2024, time.February, 29, 23, 59, 59),
			minutes:     580,
			expectedAvg: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Place entry in the middle of the range
			entryTime := tt.start.Add(12 * time.Hour)
			entries := []entry.Entry{
				makeEntry(entryTime, tt.minutes, "test entry"),
			}

			stats := CalculateStatistics(entries, tt.start, tt.end)

			if stats.AverageMinutesPerDay != tt.expectedAvg {
				t.Errorf("AverageMinutesPerDay = %f, expected %f", stats.AverageMinutesPerDay, tt.expectedAvg)
			}
		})
	}
}

func TestCalculateStatistics_ZeroAverage(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	// No entries in the range
	entries := []entry.Entry{
		makeEntry(makeTime(2024, time.January, 10, 10, 0, 0), 120, "before range"),
	}

	stats := CalculateStatistics(entries, start, end)

	if stats.TotalMinutes != 0 {
		t.Errorf("TotalMinutes = %d, expected 0", stats.TotalMinutes)
	}
	if stats.AverageMinutesPerDay != 0 {
		t.Errorf("AverageMinutesPerDay = %f, expected 0", stats.AverageMinutesPerDay)
	}
}

func TestCalculateStatistics_LargeNumbers(t *testing.T) {
	start := makeTime(2024, time.January, 1, 0, 0, 0)
	end := makeTime(2024, time.December, 31, 23, 59, 59)

	// Create entries for each day with 8 hours of work
	var entries []entry.Entry
	for day := 1; day <= 365; day++ {
		timestamp := start.AddDate(0, 0, day-1).Add(9 * time.Hour)
		entries = append(entries, makeEntry(timestamp, 480, "daily work"))
	}

	stats := CalculateStatistics(entries, start, end)

	expectedTotal := 480 * 365
	if stats.TotalMinutes != expectedTotal {
		t.Errorf("TotalMinutes = %d, expected %d", stats.TotalMinutes, expectedTotal)
	}
	if stats.EntryCount != 365 {
		t.Errorf("EntryCount = %d, expected 365", stats.EntryCount)
	}
	if stats.DaysWithEntries != 365 {
		t.Errorf("DaysWithEntries = %d, expected 365", stats.DaysWithEntries)
	}

	// 366 days in 2024 (leap year), 480 * 365 minutes total
	expectedAvg := float64(480*365) / 366.0
	if stats.AverageMinutesPerDay != expectedAvg {
		t.Errorf("AverageMinutesPerDay = %f, expected %f", stats.AverageMinutesPerDay, expectedAvg)
	}
}
