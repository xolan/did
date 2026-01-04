package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// Helper function to create test entries for stats testing
func createStatsTestEntries(t *testing.T, storagePath string) []entry.Entry {
	t.Helper()

	now := time.Now()
	// Get the start of this week to ensure entries are in current week
	startOfWeek, _ := timeutil.ThisWeek()

	entries := []entry.Entry{
		{
			Timestamp:       startOfWeek,
			Description:     "First entry this week",
			DurationMinutes: 120, // 2h
			RawInput:        "First entry this week for 2h",
		},
		{
			Timestamp:       startOfWeek.Add(24 * time.Hour), // Next day
			Description:     "Second entry",
			DurationMinutes: 90, // 1h 30m
			RawInput:        "Second entry for 1h30m",
		},
		{
			Timestamp:       startOfWeek.Add(48 * time.Hour), // Two days later
			Description:     "Third entry",
			DurationMinutes: 45, // 45m
			RawInput:        "Third entry for 45m",
		},
	}

	// Add an old entry outside current week
	entries = append(entries, entry.Entry{
		Timestamp:       now.AddDate(0, 0, -30), // 30 days ago
		Description:     "Old entry",
		DurationMinutes: 60,
		RawInput:        "Old entry for 1h",
	})

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	return entries
}

// Tests for default week statistics

func TestStats_DefaultWeek_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createStatsTestEntries(t, storagePath)

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats (default week)
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should show header for week
	if !strings.Contains(output, "Statistics for this week") {
		t.Errorf("Expected header 'Statistics for this week', got: %s", output)
	}
	// Should show total hours (120 + 90 + 45 = 255 minutes = 4h 15m)
	if !strings.Contains(output, "Total Hours:") {
		t.Error("Expected 'Total Hours:' label in output")
	}
	if !strings.Contains(output, "4h 15m") {
		t.Errorf("Expected '4h 15m' in output, got: %s", output)
	}
	// Should show average daily hours
	if !strings.Contains(output, "Average/Day:") {
		t.Error("Expected 'Average/Day:' label in output")
	}
	// Should show entries count
	if !strings.Contains(output, "Entries:") {
		t.Error("Expected 'Entries:' label in output")
	}
	if !strings.Contains(output, "3 entrys") {
		t.Errorf("Expected '3 entrys', got: %s", output)
	}
	// Should show days tracked
	if !strings.Contains(output, "Days Tracked:") {
		t.Error("Expected 'Days Tracked:' label in output")
	}
	if !strings.Contains(output, "3 days") {
		t.Errorf("Expected '3 days' tracked, got: %s", output)
	}
}

func TestStats_DefaultWeek_EmptyStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	// Don't create any entries

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats (with no entries)
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should show header
	if !strings.Contains(output, "Statistics for this week") {
		t.Errorf("Expected header, got: %s", output)
	}
	// Should show zero values
	if !strings.Contains(output, "Total Hours:     0m") {
		t.Errorf("Expected 'Total Hours: 0m', got: %s", output)
	}
	if !strings.Contains(output, "0 entrys") {
		t.Errorf("Expected '0 entrys', got: %s", output)
	}
	if !strings.Contains(output, "0 days") {
		t.Errorf("Expected '0 days', got: %s", output)
	}
}

func TestStats_DefaultWeek_NoEntriesInPeriod(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create only old entries (outside current week)
	now := time.Now()
	oldEntry := entry.Entry{
		Timestamp:       now.AddDate(0, 0, -30), // 30 days ago
		Description:     "Old entry",
		DurationMinutes: 120,
		RawInput:        "Old entry for 2h",
	}
	if err := storage.AppendEntry(storagePath, oldEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats (with no entries in current week)
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should show zero values even though entries exist in storage
	if !strings.Contains(output, "Total Hours:     0m") {
		t.Errorf("Expected zero total hours, got: %s", output)
	}
	if !strings.Contains(output, "0 entrys") {
		t.Errorf("Expected zero entries, got: %s", output)
	}
}

// Tests for monthly statistics

func TestStats_Month_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries in current month
	startOfMonth, _ := timeutil.ThisMonth()
	entries := []entry.Entry{
		{
			Timestamp:       startOfMonth,
			Description:     "First entry this month",
			DurationMinutes: 180, // 3h
			RawInput:        "First entry this month for 3h",
		},
		{
			Timestamp:       startOfMonth.Add(24 * time.Hour),
			Description:     "Second entry",
			DurationMinutes: 120, // 2h
			RawInput:        "Second entry for 2h",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats --month
	_ = statsCmd.Flags().Set("month", "true")
	defer func() { _ = statsCmd.Flags().Set("month", "false") }()

	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should show header for month
	if !strings.Contains(output, "Statistics for this month") {
		t.Errorf("Expected header 'Statistics for this month', got: %s", output)
	}
	// Should show total hours (180 + 120 = 300 minutes = 5h)
	if !strings.Contains(output, "5h") {
		t.Errorf("Expected '5h' in output, got: %s", output)
	}
	// Should show entries count
	if !strings.Contains(output, "2 entrys") {
		t.Errorf("Expected '2 entrys', got: %s", output)
	}
	// Should show days tracked
	if !strings.Contains(output, "2 days") {
		t.Errorf("Expected '2 days' tracked, got: %s", output)
	}
}

func TestStats_Month_EmptyPeriod(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entry outside current month
	now := time.Now()
	oldEntry := entry.Entry{
		Timestamp:       now.AddDate(0, -2, 0), // 2 months ago
		Description:     "Old entry",
		DurationMinutes: 60,
		RawInput:        "Old entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, oldEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats --month (with no entries in current month)
	_ = statsCmd.Flags().Set("month", "true")
	defer func() { _ = statsCmd.Flags().Set("month", "false") }()

	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should show zero values
	if !strings.Contains(output, "Total Hours:     0m") {
		t.Errorf("Expected zero total hours, got: %s", output)
	}
	if !strings.Contains(output, "0 entrys") {
		t.Errorf("Expected zero entries, got: %s", output)
	}
}

// Tests for output formatting

func TestStats_OutputFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createStatsTestEntries(t, storagePath)

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should have separators (60 equals signs)
	if !strings.Contains(output, strings.Repeat("=", 60)) {
		t.Error("Expected separator line (60 equals) in output")
	}
	// Should have properly formatted labels with padding
	if !strings.Contains(output, "Total Hours:") {
		t.Error("Expected 'Total Hours:' label")
	}
	if !strings.Contains(output, "Average/Day:") {
		t.Error("Expected 'Average/Day:' label")
	}
	if !strings.Contains(output, "Entries:") {
		t.Error("Expected 'Entries:' label")
	}
	if !strings.Contains(output, "Days Tracked:") {
		t.Error("Expected 'Days Tracked:' label")
	}
}

func TestStats_DurationFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with specific durations to test formatting
	startOfWeek, _ := timeutil.ThisWeek()
	entries := []entry.Entry{
		{
			Timestamp:       startOfWeek,
			Description:     "Entry with hours and minutes",
			DurationMinutes: 135, // 2h 15m
			RawInput:        "Entry for 2h15m",
		},
		{
			Timestamp:       startOfWeek.Add(1 * time.Hour),
			Description:     "Entry with only hours",
			DurationMinutes: 180, // 3h
			RawInput:        "Entry for 3h",
		},
		{
			Timestamp:       startOfWeek.Add(2 * time.Hour),
			Description:     "Entry with only minutes",
			DurationMinutes: 45, // 45m
			RawInput:        "Entry for 45m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Total: 135 + 180 + 45 = 360 minutes = 6h
	if !strings.Contains(output, "6h") {
		t.Errorf("Expected '6h' in output, got: %s", output)
	}
	// Should show average as decimal (average calculation tested in internal/stats)
	if !strings.Contains(output, "Average/Day:") {
		t.Error("Expected average to be displayed")
	}
}

func TestStats_AverageCalculation(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries totaling 7 hours (420 minutes) over a week
	startOfWeek, _ := timeutil.ThisWeek()
	entries := []entry.Entry{
		{
			Timestamp:       startOfWeek,
			Description:     "Day 1",
			DurationMinutes: 60, // 1h
			RawInput:        "Day 1 for 1h",
		},
		{
			Timestamp:       startOfWeek.Add(24 * time.Hour),
			Description:     "Day 2",
			DurationMinutes: 120, // 2h
			RawInput:        "Day 2 for 2h",
		},
		{
			Timestamp:       startOfWeek.Add(48 * time.Hour),
			Description:     "Day 3",
			DurationMinutes: 240, // 4h
			RawInput:        "Day 3 for 4h",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Total: 420 minutes = 7h
	if !strings.Contains(output, "7h") {
		t.Errorf("Expected '7h' total, got: %s", output)
	}
	// Average should be displayed (calculation tested in stats package)
	if !strings.Contains(output, "Average/Day:") {
		t.Error("Expected average to be shown")
	}
	// Should show days tracked
	if !strings.Contains(output, "3 days") {
		t.Errorf("Expected '3 days' tracked, got: %s", output)
	}
}

// Tests for edge cases

func TestStats_SoftDeletedEntriesExcluded(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	startOfWeek, _ := timeutil.ThisWeek()
	entries := []entry.Entry{
		{
			Timestamp:       startOfWeek,
			Description:     "Active entry",
			DurationMinutes: 120, // 2h
			RawInput:        "Active entry for 2h",
			DeletedAt:       nil, // Not deleted
		},
		{
			Timestamp:       startOfWeek.Add(1 * time.Hour),
			Description:     "Deleted entry",
			DurationMinutes: 60, // 1h
			RawInput:        "Deleted entry for 1h",
			DeletedAt:       &now, // Soft deleted
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should only count active entry (2h)
	if !strings.Contains(output, "Total Hours:     2h") {
		t.Errorf("Expected 'Total Hours: 2h' (excluding deleted), got: %s", output)
	}
	// Should show 1 entry (not 2)
	if !strings.Contains(output, "1 entry") {
		t.Errorf("Expected '1 entry', got: %s", output)
	}
}

func TestStats_EntriesOutsideRangeExcluded(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	startOfWeek, _ := timeutil.ThisWeek()
	entries := []entry.Entry{
		{
			Timestamp:       startOfWeek,
			Description:     "This week entry",
			DurationMinutes: 120, // 2h
			RawInput:        "This week entry for 2h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -30), // 30 days ago
			Description:     "Old entry",
			DurationMinutes: 60, // 1h
			RawInput:        "Old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -60), // 60 days ago
			Description:     "Very old entry",
			DurationMinutes: 90, // 1h 30m
			RawInput:        "Very old entry for 1h30m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats (default week)
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should only count current week entry (2h)
	if !strings.Contains(output, "Total Hours:     2h") {
		t.Errorf("Expected 'Total Hours: 2h' (only current week), got: %s", output)
	}
	// Should show 1 entry
	if !strings.Contains(output, "1 entry") {
		t.Errorf("Expected '1 entry', got: %s", output)
	}
}

func TestStats_CorruptedStorageWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a valid entry
	startOfWeek, _ := timeutil.ThisWeek()
	validEntry := entry.Entry{
		Timestamp:       startOfWeek,
		Description:     "Valid entry",
		DurationMinutes: 60,
		RawInput:        "Valid entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, validEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	// Append a corrupted line manually
	f, err := os.OpenFile(storagePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open storage file: %v", err)
	}
	_, err = f.WriteString("this is not valid json\n")
	if closeErr := f.Close(); closeErr != nil {
		t.Fatalf("Failed to close storage file: %v", closeErr)
	}
	if err != nil {
		t.Fatalf("Failed to write corrupted line: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats
	runStats(statsCmd, []string{})

	stderrOutput := stderr.String()
	// Should show warning about corrupted line on stderr
	if !strings.Contains(stderrOutput, "Warning: Found") {
		t.Errorf("Expected corruption warning on stderr, got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "corrupted line") {
		t.Errorf("Expected 'corrupted line' in warning, got: %s", stderrOutput)
	}

	stdoutOutput := stdout.String()
	// Should still show statistics for valid entries
	if !strings.Contains(stdoutOutput, "Statistics for this week") {
		t.Error("Expected statistics to be shown despite corruption")
	}
	if !strings.Contains(stdoutOutput, "1h") {
		t.Errorf("Expected '1h' from valid entry, got: %s", stdoutOutput)
	}
}

// Tests for error handling

func TestStats_StoragePathError(t *testing.T) {
	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", fmt.Errorf("storage path error")
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats
	runStats(statsCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for storage path error")
	}
	if !strings.Contains(stderr.String(), "Failed to determine storage location") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}
}

func TestStats_ReadEntriesError(t *testing.T) {
	// Use a path to a directory (not a file) to cause read error
	tmpDir := t.TempDir()

	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return tmpDir, nil // path to directory, not file
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats
	runStats(statsCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for read entries error")
	}
	if !strings.Contains(stderr.String(), "Failed to read entries") {
		t.Errorf("Expected read entries error, got: %s", stderr.String())
	}
}

func TestStats_WeekVsMonthHeader(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test 1: Week header (default)
	runStats(statsCmd, []string{})
	output := stdout.String()
	if !strings.Contains(output, "Statistics for this week") {
		t.Errorf("Expected 'Statistics for this week', got: %s", output)
	}
	if strings.Contains(output, "Statistics for this month") {
		t.Error("Should not show month header for default week view")
	}

	// Reset stdout
	stdout.Reset()

	// Test 2: Month header (with --month flag)
	_ = statsCmd.Flags().Set("month", "true")
	defer func() { _ = statsCmd.Flags().Set("month", "false") }()

	runStats(statsCmd, []string{})
	output = stdout.String()
	if !strings.Contains(output, "Statistics for this month") {
		t.Errorf("Expected 'Statistics for this month', got: %s", output)
	}
	if strings.Contains(output, "Statistics for this week") {
		t.Error("Should not show week header for month view")
	}
}

func TestStats_Pluralization(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create exactly 1 entry
	startOfWeek, _ := timeutil.ThisWeek()
	singleEntry := entry.Entry{
		Timestamp:       startOfWeek,
		Description:     "Single entry",
		DurationMinutes: 60,
		RawInput:        "Single entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, singleEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Test: did stats
	runStats(statsCmd, []string{})

	output := stdout.String()
	// Should show "1 entry" (singular), not "1 entrys"
	if !strings.Contains(output, "1 entry") {
		t.Errorf("Expected '1 entry' (singular), got: %s", output)
	}
	// Should show "1 day" (singular), not "1 days"
	if !strings.Contains(output, "1 day") {
		t.Errorf("Expected '1 day' (singular), got: %s", output)
	}
}
