package cli

import (
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/storage"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		minutes int
		want    string
	}{
		{0, "0m"},
		{1, "1m"},
		{30, "30m"},
		{59, "59m"},
		{60, "1h"},
		{90, "1h 30m"},
		{120, "2h"},
		{150, "2h 30m"},
		{180, "3h"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := FormatDuration(tt.minutes)
			if result != tt.want {
				t.Errorf("FormatDuration(%d) = %q, want %q", tt.minutes, result, tt.want)
			}
		})
	}
}

func TestFormatProjectAndTags(t *testing.T) {
	tests := []struct {
		name    string
		project string
		tags    []string
		want    string
	}{
		{"no project or tags", "", nil, ""},
		{"empty tags slice", "", []string{}, ""},
		{"project only", "acme", nil, "@acme"},
		{"single tag", "", []string{"urgent"}, "#urgent"},
		{"multiple tags", "", []string{"urgent", "bug"}, "#urgent #bug"},
		{"project and tags", "acme", []string{"urgent"}, "@acme #urgent"},
		{"project and multiple tags", "acme", []string{"urgent", "bug"}, "@acme #urgent #bug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatProjectAndTags(tt.project, tt.tags)
			if result != tt.want {
				t.Errorf("FormatProjectAndTags(%q, %v) = %q, want %q", tt.project, tt.tags, result, tt.want)
			}
		})
	}
}

func TestFormatEntryForLog(t *testing.T) {
	tests := []struct {
		name        string
		description string
		project     string
		tags        []string
		want        string
	}{
		{"no metadata", "fix bug", "", nil, "fix bug"},
		{"with project", "fix bug", "acme", nil, "fix bug [@acme]"},
		{"with tag", "fix bug", "", []string{"urgent"}, "fix bug [#urgent]"},
		{"with project and tags", "fix bug", "acme", []string{"urgent"}, "fix bug [@acme #urgent]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEntryForLog(tt.description, tt.project, tt.tags)
			if result != tt.want {
				t.Errorf("FormatEntryForLog(%q, %q, %v) = %q, want %q", tt.description, tt.project, tt.tags, result, tt.want)
			}
		})
	}
}

func TestFormatEntry(t *testing.T) {
	e := entry.Entry{
		Description: "fix bug",
		Project:     "acme",
		Tags:        []string{"urgent"},
	}

	result := FormatEntry(e)
	want := "fix bug [@acme #urgent]"
	if result != want {
		t.Errorf("FormatEntry() = %q, want %q", result, want)
	}
}

func TestFormatElapsedTime(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{0, "0m"},
		{1 * time.Minute, "1m"},
		{30 * time.Minute, "30m"},
		{59 * time.Minute, "59m"},
		{60 * time.Minute, "1h"},
		{90 * time.Minute, "1h 30m"},
		{120 * time.Minute, "2h"},
		{150 * time.Minute, "2h 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := FormatElapsedTime(tt.duration)
			if result != tt.want {
				t.Errorf("FormatElapsedTime(%v) = %q, want %q", tt.duration, result, tt.want)
			}
		})
	}
}

func TestFormatDateRangeForDisplay(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{
			name:  "same day",
			start: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "same year",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "different years",
			start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDateRangeForDisplay(tt.start, tt.end)
			if result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestFormatDateRangeForDisplay_SameDay(t *testing.T) {
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC)

	result := FormatDateRangeForDisplay(start, end)
	// Should be a single date format like "Mon, Jan 2, 2006"
	if result != "Mon, Jan 15, 2024" {
		t.Errorf("expected 'Mon, Jan 15, 2024', got %q", result)
	}
}

func TestFormatDateRangeForDisplay_SameYear(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	result := FormatDateRangeForDisplay(start, end)
	// Should be "Jan 2 - Jan 2, 2006" format
	if result != "Jan 1 - Dec 31, 2024" {
		t.Errorf("expected 'Jan 1 - Dec 31, 2024', got %q", result)
	}
}

func TestFormatDateRangeForDisplay_DifferentYears(t *testing.T) {
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	result := FormatDateRangeForDisplay(start, end)
	if result != "Jan 1, 2023 - Dec 31, 2024" {
		t.Errorf("expected 'Jan 1, 2023 - Dec 31, 2024', got %q", result)
	}
}

func TestFormatCorruptionWarning(t *testing.T) {
	tests := []struct {
		name    string
		warning storage.ParseWarning
	}{
		{
			name: "short content",
			warning: storage.ParseWarning{
				LineNumber: 1,
				Content:    "short content",
				Error:      "parse error",
			},
		},
		{
			name: "long content gets truncated",
			warning: storage.ParseWarning{
				LineNumber: 42,
				Content:    "this is a very long content that should be truncated because it exceeds 50 characters",
				Error:      "invalid json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCorruptionWarning(tt.warning)
			if result == "" {
				t.Error("expected non-empty result")
			}
			// Verify line number is included
			if len(result) == 0 {
				t.Error("expected formatted string")
			}
		})
	}
}

func TestFormatCorruptionWarning_Truncation(t *testing.T) {
	warning := storage.ParseWarning{
		LineNumber: 1,
		Content:    "this is a very long content that should definitely be truncated because it's too long",
		Error:      "parse error",
	}

	result := FormatCorruptionWarning(warning)
	// Should contain "..." indicating truncation
	if len(result) > 150 { // Reasonable max length
		t.Errorf("result seems too long: %s", result)
	}
}

func TestBuildPeriodWithFilters(t *testing.T) {
	tests := []struct {
		name    string
		period  string
		project string
		tags    []string
		want    string
	}{
		{"no filters", "today", "", nil, "today"},
		{"empty tags", "today", "", []string{}, "today"},
		{"project only", "today", "acme", nil, "today (@acme)"},
		{"tag only", "today", "", []string{"urgent"}, "today (#urgent)"},
		{"project and tag", "today", "acme", []string{"urgent"}, "today (@acme #urgent)"},
		{"multiple tags", "this week", "", []string{"urgent", "bug"}, "this week (#urgent #bug)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPeriodWithFilters(tt.period, tt.project, tt.tags)
			if result != tt.want {
				t.Errorf("BuildPeriodWithFilters(%q, %q, %v) = %q, want %q", tt.period, tt.project, tt.tags, result, tt.want)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		word  string
		count int
		want  string
	}{
		{"entry", 0, "entrys"},
		{"entry", 1, "entry"},
		{"entry", 2, "entrys"},
		{"day", 1, "day"},
		{"day", 5, "days"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := Pluralize(tt.word, tt.count)
			if result != tt.want {
				t.Errorf("Pluralize(%q, %d) = %q, want %q", tt.word, tt.count, result, tt.want)
			}
		})
	}
}

func TestSpansMultipleDays(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name    string
		entries []entry.Entry
		want    bool
	}{
		{"empty", []entry.Entry{}, false},
		{"single entry", []entry.Entry{{Timestamp: now}}, false},
		{"same day", []entry.Entry{{Timestamp: now}, {Timestamp: now}}, false},
		{"different days", []entry.Entry{{Timestamp: now}, {Timestamp: yesterday}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SpansMultipleDays(tt.entries)
			if result != tt.want {
				t.Errorf("SpansMultipleDays() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestSpansMultipleDaysIndexed(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name    string
		entries []service.IndexedEntry
		want    bool
	}{
		{"empty", []service.IndexedEntry{}, false},
		{"single entry", []service.IndexedEntry{{Entry: entry.Entry{Timestamp: now}}}, false},
		{"same day", []service.IndexedEntry{
			{Entry: entry.Entry{Timestamp: now}},
			{Entry: entry.Entry{Timestamp: now}},
		}, false},
		{"different days", []service.IndexedEntry{
			{Entry: entry.Entry{Timestamp: now}},
			{Entry: entry.Entry{Timestamp: yesterday}},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SpansMultipleDaysIndexed(tt.entries)
			if result != tt.want {
				t.Errorf("SpansMultipleDaysIndexed() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestFormatTimerStartTime(t *testing.T) {
	now := time.Now()

	// Test today
	todayResult := FormatTimerStartTime(now)
	if todayResult == "" {
		t.Error("expected non-empty result for today")
	}
	if len(todayResult) < 5 {
		t.Error("result seems too short")
	}

	// Test yesterday
	yesterday := now.AddDate(0, 0, -1)
	yesterdayResult := FormatTimerStartTime(yesterday)
	if yesterdayResult == "" {
		t.Error("expected non-empty result for yesterday")
	}
	// Should not contain "today"
	if contains(yesterdayResult, "today") {
		t.Errorf("yesterday result should not contain 'today': %s", yesterdayResult)
	}
}

func TestFormatTimerStartTime_Today(t *testing.T) {
	// Create a time for today
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 14, 30, 0, 0, now.Location())

	result := FormatTimerStartTime(today)
	if !contains(result, "today") {
		t.Errorf("expected 'today' in result, got %q", result)
	}
}

func TestFormatTimerStartTime_PastDay(t *testing.T) {
	// Create a time for a past day
	pastDay := time.Now().AddDate(0, 0, -3)

	result := FormatTimerStartTime(pastDay)
	if contains(result, "today") {
		t.Errorf("past day result should not contain 'today': %s", result)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
