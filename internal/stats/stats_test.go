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

// Tests for CalculateProjectBreakdown

func TestCalculateProjectBreakdown_EmptyEntries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	breakdown := CalculateProjectBreakdown([]entry.Entry{}, start, end)

	if len(breakdown) != 0 {
		t.Errorf("Expected empty breakdown, got %d items", len(breakdown))
	}
}

func TestCalculateProjectBreakdown_SingleProject(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "worked on feature",
			RawInput:        "worked on feature @projectA for 2h",
			Project:         "projectA",
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 90,
			Description:     "more work",
			RawInput:        "more work @projectA for 1h30m",
			Project:         "projectA",
		},
	}

	breakdown := CalculateProjectBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(breakdown))
	}

	if breakdown[0].Project != "projectA" {
		t.Errorf("Project = %s, expected projectA", breakdown[0].Project)
	}
	if breakdown[0].TotalMinutes != 210 {
		t.Errorf("TotalMinutes = %d, expected 210", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 2 {
		t.Errorf("EntryCount = %d, expected 2", breakdown[0].EntryCount)
	}
}

func TestCalculateProjectBreakdown_MultipleProjects(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 180,
			Description:     "work",
			RawInput:        "work @projectA for 3h",
			Project:         "projectA",
		},
		{
			Timestamp:       makeTime(2024, time.January, 16, 14, 0, 0),
			DurationMinutes: 120,
			Description:     "work",
			RawInput:        "work @projectB for 2h",
			Project:         "projectB",
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 90,
			Description:     "work",
			RawInput:        "work @projectA for 1h30m",
			Project:         "projectA",
		},
		{
			Timestamp:       makeTime(2024, time.January, 18, 10, 0, 0),
			DurationMinutes: 60,
			Description:     "work",
			RawInput:        "work @projectC for 1h",
			Project:         "projectC",
		},
	}

	breakdown := CalculateProjectBreakdown(entries, start, end)

	if len(breakdown) != 3 {
		t.Fatalf("Expected 3 projects, got %d", len(breakdown))
	}

	// Check that results are sorted by total minutes descending
	// projectA: 180 + 90 = 270
	// projectB: 120
	// projectC: 60
	if breakdown[0].Project != "projectA" || breakdown[0].TotalMinutes != 270 {
		t.Errorf("First project should be projectA with 270 minutes, got %s with %d", breakdown[0].Project, breakdown[0].TotalMinutes)
	}
	if breakdown[1].Project != "projectB" || breakdown[1].TotalMinutes != 120 {
		t.Errorf("Second project should be projectB with 120 minutes, got %s with %d", breakdown[1].Project, breakdown[1].TotalMinutes)
	}
	if breakdown[2].Project != "projectC" || breakdown[2].TotalMinutes != 60 {
		t.Errorf("Third project should be projectC with 60 minutes, got %s with %d", breakdown[2].Project, breakdown[2].TotalMinutes)
	}
}

func TestCalculateProjectBreakdown_NoProject(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "work without project",
			RawInput:        "work without project for 2h",
			Project:         "", // No project
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 90,
			Description:     "more work",
			RawInput:        "more work for 1h30m",
			Project:         "", // No project
		},
	}

	breakdown := CalculateProjectBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 project group, got %d", len(breakdown))
	}

	if breakdown[0].Project != "(no project)" {
		t.Errorf("Project = %s, expected '(no project)'", breakdown[0].Project)
	}
	if breakdown[0].TotalMinutes != 210 {
		t.Errorf("TotalMinutes = %d, expected 210", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 2 {
		t.Errorf("EntryCount = %d, expected 2", breakdown[0].EntryCount)
	}
}

func TestCalculateProjectBreakdown_MixedProjectsAndNoProject(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 180,
			Description:     "work",
			RawInput:        "work @projectA for 3h",
			Project:         "projectA",
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "work without project",
			RawInput:        "work without project for 2h",
			Project:         "",
		},
	}

	breakdown := CalculateProjectBreakdown(entries, start, end)

	if len(breakdown) != 2 {
		t.Fatalf("Expected 2 project groups, got %d", len(breakdown))
	}

	// Should be sorted by minutes (projectA: 180, (no project): 120)
	if breakdown[0].Project != "projectA" {
		t.Errorf("First project = %s, expected 'projectA'", breakdown[0].Project)
	}
	if breakdown[1].Project != "(no project)" {
		t.Errorf("Second project = %s, expected '(no project)'", breakdown[1].Project)
	}
}

func TestCalculateProjectBreakdown_SkipsDeletedEntries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "active",
			RawInput:        "active @projectA for 2h",
			Project:         "projectA",
			DeletedAt:       nil,
		},
		makeDeletedEntry(makeTime(2024, time.January, 17, 10, 0, 0), 180, "deleted @projectA for 3h"),
	}
	entries[1].Project = "projectA"

	breakdown := CalculateProjectBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(breakdown))
	}

	// Should only count active entry
	if breakdown[0].TotalMinutes != 120 {
		t.Errorf("TotalMinutes = %d, expected 120 (deleted entry should be excluded)", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 1 {
		t.Errorf("EntryCount = %d, expected 1", breakdown[0].EntryCount)
	}
}

func TestCalculateProjectBreakdown_EntriesOutsideRange(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 14, 23, 59, 59), // Before range
			DurationMinutes: 120,
			Description:     "before",
			RawInput:        "before @projectA for 2h",
			Project:         "projectA",
		},
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0), // In range
			DurationMinutes: 90,
			Description:     "in range",
			RawInput:        "in range @projectA for 1h30m",
			Project:         "projectA",
		},
		{
			Timestamp:       makeTime(2024, time.January, 22, 0, 0, 0), // After range
			DurationMinutes: 60,
			Description:     "after",
			RawInput:        "after @projectA for 1h",
			Project:         "projectA",
		},
	}

	breakdown := CalculateProjectBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(breakdown))
	}

	// Should only count entry in range
	if breakdown[0].TotalMinutes != 90 {
		t.Errorf("TotalMinutes = %d, expected 90 (only in-range entry)", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 1 {
		t.Errorf("EntryCount = %d, expected 1", breakdown[0].EntryCount)
	}
}

func TestCalculateProjectBreakdown_EntriesAtBoundaries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       start, // Exactly at start
			DurationMinutes: 60,
			Description:     "at start",
			RawInput:        "at start @projectA for 1h",
			Project:         "projectA",
		},
		{
			Timestamp:       end, // Exactly at end
			DurationMinutes: 90,
			Description:     "at end",
			RawInput:        "at end @projectA for 1h30m",
			Project:         "projectA",
		},
	}

	breakdown := CalculateProjectBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(breakdown))
	}

	// Both boundary entries should be included
	if breakdown[0].TotalMinutes != 150 {
		t.Errorf("TotalMinutes = %d, expected 150 (both boundaries)", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 2 {
		t.Errorf("EntryCount = %d, expected 2", breakdown[0].EntryCount)
	}
}

// Tests for CalculateTagBreakdown

func TestCalculateTagBreakdown_EmptyEntries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	breakdown := CalculateTagBreakdown([]entry.Entry{}, start, end)

	if len(breakdown) != 0 {
		t.Errorf("Expected empty breakdown, got %d items", len(breakdown))
	}
}

func TestCalculateTagBreakdown_SingleTag(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "work",
			RawInput:        "work #development for 2h",
			Tags:            []string{"development"},
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 90,
			Description:     "more work",
			RawInput:        "more work #development for 1h30m",
			Tags:            []string{"development"},
		},
	}

	breakdown := CalculateTagBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(breakdown))
	}

	if breakdown[0].Tag != "development" {
		t.Errorf("Tag = %s, expected 'development'", breakdown[0].Tag)
	}
	if breakdown[0].TotalMinutes != 210 {
		t.Errorf("TotalMinutes = %d, expected 210", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 2 {
		t.Errorf("EntryCount = %d, expected 2", breakdown[0].EntryCount)
	}
}

func TestCalculateTagBreakdown_MultipleTags(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 180,
			Description:     "work",
			RawInput:        "work #development for 3h",
			Tags:            []string{"development"},
		},
		{
			Timestamp:       makeTime(2024, time.January, 16, 14, 0, 0),
			DurationMinutes: 120,
			Description:     "meeting",
			RawInput:        "meeting #meeting for 2h",
			Tags:            []string{"meeting"},
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 90,
			Description:     "work",
			RawInput:        "work #development for 1h30m",
			Tags:            []string{"development"},
		},
		{
			Timestamp:       makeTime(2024, time.January, 18, 10, 0, 0),
			DurationMinutes: 60,
			Description:     "review",
			RawInput:        "review #review for 1h",
			Tags:            []string{"review"},
		},
	}

	breakdown := CalculateTagBreakdown(entries, start, end)

	if len(breakdown) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(breakdown))
	}

	// Check that results are sorted by total minutes descending
	// development: 180 + 90 = 270
	// meeting: 120
	// review: 60
	if breakdown[0].Tag != "development" || breakdown[0].TotalMinutes != 270 {
		t.Errorf("First tag should be development with 270 minutes, got %s with %d", breakdown[0].Tag, breakdown[0].TotalMinutes)
	}
	if breakdown[1].Tag != "meeting" || breakdown[1].TotalMinutes != 120 {
		t.Errorf("Second tag should be meeting with 120 minutes, got %s with %d", breakdown[1].Tag, breakdown[1].TotalMinutes)
	}
	if breakdown[2].Tag != "review" || breakdown[2].TotalMinutes != 60 {
		t.Errorf("Third tag should be review with 60 minutes, got %s with %d", breakdown[2].Tag, breakdown[2].TotalMinutes)
	}
}

func TestCalculateTagBreakdown_NoTags(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "work without tags",
			RawInput:        "work without tags for 2h",
			Tags:            []string{},
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 90,
			Description:     "more work",
			RawInput:        "more work for 1h30m",
			Tags:            nil, // nil tags
		},
	}

	breakdown := CalculateTagBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 tag group, got %d", len(breakdown))
	}

	if breakdown[0].Tag != "(no tags)" {
		t.Errorf("Tag = %s, expected '(no tags)'", breakdown[0].Tag)
	}
	if breakdown[0].TotalMinutes != 210 {
		t.Errorf("TotalMinutes = %d, expected 210", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 2 {
		t.Errorf("EntryCount = %d, expected 2", breakdown[0].EntryCount)
	}
}

func TestCalculateTagBreakdown_MultipleTagsPerEntry(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "work",
			RawInput:        "work #development #backend for 2h",
			Tags:            []string{"development", "backend"},
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 90,
			Description:     "work",
			RawInput:        "work #development for 1h30m",
			Tags:            []string{"development"},
		},
	}

	breakdown := CalculateTagBreakdown(entries, start, end)

	// Should have 2 tags: development and backend
	if len(breakdown) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(breakdown))
	}

	// development: 120 (from first) + 90 (from second) = 210
	// backend: 120 (from first) = 120
	var developmentBreakdown, backendBreakdown *TagBreakdown
	for i := range breakdown {
		if breakdown[i].Tag == "development" {
			developmentBreakdown = &breakdown[i]
		} else if breakdown[i].Tag == "backend" {
			backendBreakdown = &breakdown[i]
		}
	}

	if developmentBreakdown == nil {
		t.Fatal("Expected 'development' tag in breakdown")
	}
	if developmentBreakdown.TotalMinutes != 210 {
		t.Errorf("development TotalMinutes = %d, expected 210", developmentBreakdown.TotalMinutes)
	}
	if developmentBreakdown.EntryCount != 2 {
		t.Errorf("development EntryCount = %d, expected 2 (both entries)", developmentBreakdown.EntryCount)
	}

	if backendBreakdown == nil {
		t.Fatal("Expected 'backend' tag in breakdown")
	}
	if backendBreakdown.TotalMinutes != 120 {
		t.Errorf("backend TotalMinutes = %d, expected 120", backendBreakdown.TotalMinutes)
	}
	if backendBreakdown.EntryCount != 1 {
		t.Errorf("backend EntryCount = %d, expected 1", backendBreakdown.EntryCount)
	}
}

func TestCalculateTagBreakdown_MixedTagsAndNoTags(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 180,
			Description:     "work",
			RawInput:        "work #development for 3h",
			Tags:            []string{"development"},
		},
		{
			Timestamp:       makeTime(2024, time.January, 17, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "work without tags",
			RawInput:        "work without tags for 2h",
			Tags:            []string{},
		},
	}

	breakdown := CalculateTagBreakdown(entries, start, end)

	if len(breakdown) != 2 {
		t.Fatalf("Expected 2 tag groups, got %d", len(breakdown))
	}

	// Should be sorted by minutes (development: 180, (no tags): 120)
	if breakdown[0].Tag != "development" {
		t.Errorf("First tag = %s, expected 'development'", breakdown[0].Tag)
	}
	if breakdown[1].Tag != "(no tags)" {
		t.Errorf("Second tag = %s, expected '(no tags)'", breakdown[1].Tag)
	}
}

func TestCalculateTagBreakdown_SkipsDeletedEntries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0),
			DurationMinutes: 120,
			Description:     "active",
			RawInput:        "active #development for 2h",
			Tags:            []string{"development"},
			DeletedAt:       nil,
		},
		makeDeletedEntry(makeTime(2024, time.January, 17, 10, 0, 0), 180, "deleted #development for 3h"),
	}
	entries[1].Tags = []string{"development"}

	breakdown := CalculateTagBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(breakdown))
	}

	// Should only count active entry
	if breakdown[0].TotalMinutes != 120 {
		t.Errorf("TotalMinutes = %d, expected 120 (deleted entry should be excluded)", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 1 {
		t.Errorf("EntryCount = %d, expected 1", breakdown[0].EntryCount)
	}
}

func TestCalculateTagBreakdown_EntriesOutsideRange(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       makeTime(2024, time.January, 14, 23, 59, 59), // Before range
			DurationMinutes: 120,
			Description:     "before",
			RawInput:        "before #development for 2h",
			Tags:            []string{"development"},
		},
		{
			Timestamp:       makeTime(2024, time.January, 16, 10, 0, 0), // In range
			DurationMinutes: 90,
			Description:     "in range",
			RawInput:        "in range #development for 1h30m",
			Tags:            []string{"development"},
		},
		{
			Timestamp:       makeTime(2024, time.January, 22, 0, 0, 0), // After range
			DurationMinutes: 60,
			Description:     "after",
			RawInput:        "after #development for 1h",
			Tags:            []string{"development"},
		},
	}

	breakdown := CalculateTagBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(breakdown))
	}

	// Should only count entry in range
	if breakdown[0].TotalMinutes != 90 {
		t.Errorf("TotalMinutes = %d, expected 90 (only in-range entry)", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 1 {
		t.Errorf("EntryCount = %d, expected 1", breakdown[0].EntryCount)
	}
}

func TestCalculateTagBreakdown_EntriesAtBoundaries(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 21, 23, 59, 59)

	entries := []entry.Entry{
		{
			Timestamp:       start, // Exactly at start
			DurationMinutes: 60,
			Description:     "at start",
			RawInput:        "at start #development for 1h",
			Tags:            []string{"development"},
		},
		{
			Timestamp:       end, // Exactly at end
			DurationMinutes: 90,
			Description:     "at end",
			RawInput:        "at end #development for 1h30m",
			Tags:            []string{"development"},
		},
	}

	breakdown := CalculateTagBreakdown(entries, start, end)

	if len(breakdown) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(breakdown))
	}

	// Both boundary entries should be included
	if breakdown[0].TotalMinutes != 150 {
		t.Errorf("TotalMinutes = %d, expected 150 (both boundaries)", breakdown[0].TotalMinutes)
	}
	if breakdown[0].EntryCount != 2 {
		t.Errorf("EntryCount = %d, expected 2", breakdown[0].EntryCount)
	}
}
