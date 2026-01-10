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

// Helper function to create test entries with various dates
func createTestEntries(t *testing.T, storagePath string) []entry.Entry {
	t.Helper()

	// Create entries spanning multiple days with different descriptions
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -7), // 7 days ago
			Description:     "Code review for feature X",
			DurationMinutes: 60,
			RawInput:        "Code review for feature X for 1h",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -5), // 5 days ago
			Description:     "Bug fix in authentication",
			DurationMinutes: 90,
			RawInput:        "Bug fix in authentication for 1h30m",
			Project:         "client",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -3), // 3 days ago
			Description:     "Team meeting to discuss roadmap",
			DurationMinutes: 45,
			RawInput:        "Team meeting to discuss roadmap for 45m",
			Project:         "",
			Tags:            []string{},
		},
		{
			Timestamp:       now.AddDate(0, 0, -2), // 2 days ago
			Description:     "Code refactoring in API module",
			DurationMinutes: 120,
			RawInput:        "Code refactoring in API module for 2h",
			Project:         "acme",
			Tags:            []string{"refactor"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -1), // 1 day ago
			Description:     "Review pull requests",
			DurationMinutes: 30,
			RawInput:        "Review pull requests for 30m",
			Project:         "client",
			Tags:            []string{"review"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	return entries
}

func TestSearchEntries_BasicKeywordSearch(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

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

	// Search for "review" - should find 2 entries
	searchEntries(searchCmd, []string{"review"})

	output := stdout.String()
	if !strings.Contains(output, "Search results for 'review'") {
		t.Errorf("Expected search header, got: %s", output)
	}
	if !strings.Contains(output, "Code review for feature X") {
		t.Error("Expected to find 'Code review for feature X' in results")
	}
	if !strings.Contains(output, "Review pull requests") {
		t.Error("Expected to find 'Review pull requests' in results")
	}
	// Should show total (1h + 30m = 1h 30m = 90 minutes)
	if !strings.Contains(output, "Total: 1h 30m") {
		t.Errorf("Expected 'Total: 1h 30m', got: %s", output)
	}
	if !strings.Contains(output, "2 entrys") {
		t.Errorf("Expected '2 entrys', got: %s", output)
	}
}

func TestSearchEntries_CaseInsensitiveMatching(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

	tests := []struct {
		name         string
		keyword      string
		expectedHits int
	}{
		{"lowercase", "code", 2}, // "Code review" and "Code refactoring"
		{"uppercase", "CODE", 2},
		{"mixed case", "CoDe", 2},
		{"uppercase review", "REVIEW", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			searchEntries(searchCmd, []string{tt.keyword})

			output := stdout.String()
			if !strings.Contains(output, fmt.Sprintf("Search results for '%s'", tt.keyword)) {
				t.Errorf("Expected search header with keyword '%s', got: %s", tt.keyword, output)
			}
			if !strings.Contains(output, fmt.Sprintf("%d entrys", tt.expectedHits)) {
				t.Errorf("Expected '%d entrys', got: %s", tt.expectedHits, output)
			}
		})
	}
}

func TestSearchEntries_NoResults(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

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

	// Search for something that doesn't exist
	searchEntries(searchCmd, []string{"nonexistent"})

	output := stdout.String()
	if !strings.Contains(output, "No entries found matching 'nonexistent'") {
		t.Errorf("Expected 'No entries found' message, got: %s", output)
	}
}

func TestSearchEntries_WithFromDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	testEntries := createTestEntries(t, storagePath)

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

	// Set --from flag to 3 days ago (should find entries from 3, 2, 1 days ago)
	threeDaysAgo := testEntries[2].Timestamp.Format("2006-01-02")
	_ = searchCmd.Flags().Set("from", threeDaysAgo)
	defer func() { _ = searchCmd.Flags().Set("from", "") }()

	// Search for any entry (empty keyword would match all, but we'll search for a common term)
	searchEntries(searchCmd, []string{"meeting"})

	output := stdout.String()
	if !strings.Contains(output, "Team meeting to discuss roadmap") {
		t.Errorf("Expected to find meeting entry, got: %s", output)
	}
	// Should show date filter in header
	if !strings.Contains(output, "Search results for 'meeting'") {
		t.Errorf("Expected search header with date filter, got: %s", output)
	}
}

func TestSearchEntries_WithToDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	testEntries := createTestEntries(t, storagePath)

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

	// Set --to flag to 4 days ago (should find entries from 7, 5 days ago, not 3, 2, 1)
	fourDaysAgo := testEntries[1].Timestamp.AddDate(0, 0, 1).Format("2006-01-02")
	_ = searchCmd.Flags().Set("to", fourDaysAgo)
	defer func() { _ = searchCmd.Flags().Set("to", "") }()

	// Search for "fix"
	searchEntries(searchCmd, []string{"fix"})

	output := stdout.String()
	if !strings.Contains(output, "Bug fix in authentication") {
		t.Errorf("Expected to find bug fix entry, got: %s", output)
	}
	if strings.Contains(output, "Code refactoring") {
		t.Error("Should not find entries after --to date")
	}
}

func TestSearchEntries_WithDateRange(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	testEntries := createTestEntries(t, storagePath)

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

	// Set --from to 6 days ago and --to to 2 days ago (should find entries at 5, 3, 2 days ago)
	sixDaysAgo := testEntries[1].Timestamp.AddDate(0, 0, -1).Format("2006-01-02")
	twoDaysAgo := testEntries[3].Timestamp.Format("2006-01-02")
	_ = searchCmd.Flags().Set("from", sixDaysAgo)
	_ = searchCmd.Flags().Set("to", twoDaysAgo)
	defer func() {
		_ = searchCmd.Flags().Set("from", "")
		_ = searchCmd.Flags().Set("to", "")
	}()

	// Search for entries with "Code" (should find "Code refactoring")
	searchEntries(searchCmd, []string{"Code"})

	output := stdout.String()
	if !strings.Contains(output, "Code refactoring") {
		t.Errorf("Expected to find 'Code refactoring', got: %s", output)
	}
	// Should show date range in header
	if !strings.Contains(output, "Search results for 'Code'") {
		t.Errorf("Expected search header with date range, got: %s", output)
	}
}

func TestSearchEntries_WithLastDaysFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

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

	// Set --last 4 (last 4 days - should find entries at 3, 2, 1 days ago)
	_ = searchCmd.Flags().Set("last", "4")
	defer func() { _ = searchCmd.Flags().Set("last", "0") }()

	// Search for entries
	searchEntries(searchCmd, []string{"Code"})

	output := stdout.String()
	// Should show "last N days" in header
	if !strings.Contains(output, "last 4 days") {
		t.Errorf("Expected 'last 4 days' in header, got: %s", output)
	}
	if !strings.Contains(output, "Code refactoring") {
		t.Error("Expected to find code refactoring entry (2 days ago)")
	}
}

func TestSearchEntries_InvalidFromDate(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	_ = searchCmd.Flags().Set("from", "invalid-date")
	defer func() { _ = searchCmd.Flags().Set("from", "") }()

	searchEntries(searchCmd, []string{"test"})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid --from date")
	}
	if !strings.Contains(stderr.String(), "Invalid --from date") {
		t.Errorf("Expected invalid date error, got: %s", stderr.String())
	}
}

func TestSearchEntries_InvalidToDate(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	_ = searchCmd.Flags().Set("to", "invalid-date")
	defer func() { _ = searchCmd.Flags().Set("to", "") }()

	searchEntries(searchCmd, []string{"test"})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid --to date")
	}
	if !strings.Contains(stderr.String(), "Invalid --to date") {
		t.Errorf("Expected invalid date error, got: %s", stderr.String())
	}
}

func TestSearchEntries_ConflictingFlags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Set both --last and --from (should error)
	_ = searchCmd.Flags().Set("last", "7")
	_ = searchCmd.Flags().Set("from", "2024-01-01")
	defer func() {
		_ = searchCmd.Flags().Set("last", "0")
		_ = searchCmd.Flags().Set("from", "")
	}()

	searchEntries(searchCmd, []string{"test"})

	if !exitCalled {
		t.Error("Expected exit to be called for conflicting flags")
	}
	if !strings.Contains(stderr.String(), "Cannot use --last with --from or --to") {
		t.Errorf("Expected conflicting flags error, got: %s", stderr.String())
	}
}

func TestSearchEntries_TotalDurationCalculation(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with known durations
	entries := []entry.Entry{
		{
			Timestamp:       time.Now().AddDate(0, 0, -3),
			Description:     "Task one review",
			DurationMinutes: 60, // 1h
			RawInput:        "Task one review for 1h",
		},
		{
			Timestamp:       time.Now().AddDate(0, 0, -2),
			Description:     "Task two review",
			DurationMinutes: 90, // 1h 30m
			RawInput:        "Task two review for 1h30m",
		},
		{
			Timestamp:       time.Now().AddDate(0, 0, -1),
			Description:     "Task three review",
			DurationMinutes: 45, // 45m
			RawInput:        "Task three review for 45m",
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

	// Search for "review" - should find all 3 entries (195 minutes = 3h 15m)
	searchEntries(searchCmd, []string{"review"})

	output := stdout.String()
	if !strings.Contains(output, "Total: 3h 15m") {
		t.Errorf("Expected 'Total: 3h 15m', got: %s", output)
	}
	if !strings.Contains(output, "3 entrys") {
		t.Errorf("Expected '3 entrys', got: %s", output)
	}
}

func TestSearchEntries_StoragePathError(t *testing.T) {
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

	searchEntries(searchCmd, []string{"test"})

	if !exitCalled {
		t.Error("Expected exit to be called for storage path error")
	}
	if !strings.Contains(stderr.String(), "Failed to determine storage location") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}
}

func TestSearchEntries_ReadEntriesError(t *testing.T) {
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

	searchEntries(searchCmd, []string{"test"})

	if !exitCalled {
		t.Error("Expected exit to be called for read entries error")
	}
	if !strings.Contains(stderr.String(), "Failed to read entries") {
		t.Errorf("Expected read entries error, got: %s", stderr.String())
	}
}

func TestSearchEntries_NoResultsWithDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

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

	// Set date range where keyword doesn't match
	_ = searchCmd.Flags().Set("from", "2020-01-01")
	_ = searchCmd.Flags().Set("to", "2020-01-31")
	defer func() {
		_ = searchCmd.Flags().Set("from", "")
		_ = searchCmd.Flags().Set("to", "")
	}()

	searchEntries(searchCmd, []string{"nonexistent"})

	output := stdout.String()
	if !strings.Contains(output, "No entries found matching 'nonexistent' in the specified date range") {
		t.Errorf("Expected no results message with date range, got: %s", output)
	}
}

func TestSearchEntries_EmptyStorage(t *testing.T) {
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

	searchEntries(searchCmd, []string{"anything"})

	output := stdout.String()
	if !strings.Contains(output, "No entries found matching 'anything'") {
		t.Errorf("Expected 'No entries found' message, got: %s", output)
	}
}

func TestSearchEntries_ProjectAndTagsDisplayed(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

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

	// Search for "Code review" which has project and tags
	searchEntries(searchCmd, []string{"Code review"})

	output := stdout.String()
	// Should display project and tags in results
	if !strings.Contains(output, "@acme") {
		t.Error("Expected project '@acme' to be displayed in results")
	}
	if !strings.Contains(output, "#review") {
		t.Error("Expected tag '#review' to be displayed in results")
	}
}

func TestSearchCommand_Run(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createTestEntries(t, storagePath)

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

	// Call the search command's Run function directly
	searchCmd.Run(searchCmd, []string{"review"})

	output := stdout.String()
	if !strings.Contains(output, "Search results for 'review'") {
		t.Errorf("Expected search results, got: %s", output)
	}
}

// End-to-end integration tests combining search, filters, and date ranges

func TestIntegration_SearchWithDateFilter(t *testing.T) {
	// Test: did search meeting --from 2024-01-01
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with specific dates
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			Description:     "Team meeting about Q1 goals",
			DurationMinutes: 60,
			RawInput:        "Team meeting about Q1 goals for 1h",
			Project:         "company",
			Tags:            []string{"planning"},
		},
		{
			Timestamp:       time.Date(2024, 2, 20, 14, 30, 0, 0, time.UTC),
			Description:     "Client meeting for project review",
			DurationMinutes: 90,
			RawInput:        "Client meeting for project review for 1h30m",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -1), // Yesterday
			Description:     "Daily standup meeting",
			DurationMinutes: 15,
			RawInput:        "Daily standup meeting for 15m",
			Project:         "",
			Tags:            []string{},
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

	// Search for "meeting" from 2024-01-01 onwards
	_ = searchCmd.Flags().Set("from", "2024-01-01")
	defer func() { _ = searchCmd.Flags().Set("from", "") }()

	searchEntries(searchCmd, []string{"meeting"})

	output := stdout.String()
	// Should find all 3 meetings (all are after 2024-01-01)
	if !strings.Contains(output, "Team meeting about Q1 goals") {
		t.Error("Expected to find January meeting")
	}
	if !strings.Contains(output, "Client meeting for project review") {
		t.Error("Expected to find February meeting")
	}
	if !strings.Contains(output, "Daily standup meeting") {
		t.Error("Expected to find recent meeting")
	}
	// Verify total duration (60 + 90 + 15 = 165 minutes = 2h 45m)
	if !strings.Contains(output, "Total: 2h 45m") {
		t.Errorf("Expected total '2h 45m', got: %s", output)
	}
	if !strings.Contains(output, "3 entrys") {
		t.Errorf("Expected 3 entries, got: %s", output)
	}
}

func TestIntegration_SearchWithDateRange(t *testing.T) {
	// Test: did search code --from 2024-01-01 --to 2024-01-31
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2023, 12, 20, 10, 0, 0, 0, time.UTC),
			Description:     "Code review for December feature",
			DurationMinutes: 45,
			RawInput:        "Code review for December feature for 45m",
			Project:         "legacy",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
			Description:     "Code refactoring in auth module",
			DurationMinutes: 120,
			RawInput:        "Code refactoring in auth module for 2h",
			Project:         "acme",
			Tags:            []string{"refactor"},
		},
		{
			Timestamp:       time.Date(2024, 2, 5, 9, 0, 0, 0, time.UTC),
			Description:     "Code cleanup after migration",
			DurationMinutes: 60,
			RawInput:        "Code cleanup after migration for 1h",
			Project:         "client",
			Tags:            []string{"maintenance"},
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

	// Search for "code" within January 2024
	_ = searchCmd.Flags().Set("from", "2024-01-01")
	_ = searchCmd.Flags().Set("to", "2024-01-31")
	defer func() {
		_ = searchCmd.Flags().Set("from", "")
		_ = searchCmd.Flags().Set("to", "")
	}()

	searchEntries(searchCmd, []string{"code"})

	output := stdout.String()
	// Should only find the January entry
	if !strings.Contains(output, "Code refactoring in auth module") {
		t.Error("Expected to find January code entry")
	}
	if strings.Contains(output, "December feature") {
		t.Error("Should not find December entry (before date range)")
	}
	if strings.Contains(output, "after migration") {
		t.Error("Should not find February entry (after date range)")
	}
	// Verify only one entry found (120 minutes = 2h)
	if !strings.Contains(output, "Total: 2h") {
		t.Errorf("Expected total '2h', got: %s", output)
	}
	if !strings.Contains(output, "1 entry") {
		t.Errorf("Expected 1 entry, got: %s", output)
	}
}

func TestIntegration_SearchWithLastDays(t *testing.T) {
	// Test: did search review --last 7
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10), // 10 days ago
			Description:     "Old code review task",
			DurationMinutes: 30,
			RawInput:        "Old code review task for 30m",
		},
		{
			Timestamp:       now.AddDate(0, 0, -5), // 5 days ago
			Description:     "Review pull request #123",
			DurationMinutes: 45,
			RawInput:        "Review pull request #123 for 45m",
			Project:         "acme",
		},
		{
			Timestamp:       now.AddDate(0, 0, -2), // 2 days ago
			Description:     "Peer review of design docs",
			DurationMinutes: 60,
			RawInput:        "Peer review of design docs for 1h",
			Tags:            []string{"documentation"},
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

	// Search for "review" in last 7 days
	_ = searchCmd.Flags().Set("last", "7")
	defer func() { _ = searchCmd.Flags().Set("last", "0") }()

	searchEntries(searchCmd, []string{"review"})

	output := stdout.String()
	// Should find entries from 5 and 2 days ago, but not 10 days ago
	if strings.Contains(output, "Old code review task") {
		t.Error("Should not find entry from 10 days ago")
	}
	if !strings.Contains(output, "Review pull request #123") {
		t.Error("Expected to find entry from 5 days ago")
	}
	if !strings.Contains(output, "Peer review of design docs") {
		t.Error("Expected to find entry from 2 days ago")
	}
	// Verify header shows "last 7 days"
	if !strings.Contains(output, "last 7 days") {
		t.Error("Expected header to show 'last 7 days'")
	}
	// Total: 45m + 1h = 105 minutes = 1h 45m
	if !strings.Contains(output, "Total: 1h 45m") {
		t.Errorf("Expected total '1h 45m', got: %s", output)
	}
}

func TestIntegration_RootCommandWithShorthandFilters_Today(t *testing.T) {
	// Test: did @client #urgent (today filtered)
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "Fix critical production bug",
			DurationMinutes: 120,
			RawInput:        "Fix critical production bug @client for 2h",
			Project:         "client",
			Tags:            []string{"urgent", "bugfix"},
		},
		{
			Timestamp:       today.Add(2 * time.Hour),
			Description:     "Deploy hotfix to production",
			DurationMinutes: 30,
			RawInput:        "Deploy hotfix to production @client for 30m",
			Project:         "client",
			Tags:            []string{"urgent", "deployment"},
		},
		{
			Timestamp:       today.Add(4 * time.Hour),
			Description:     "Regular maintenance work",
			DurationMinutes: 60,
			RawInput:        "Regular maintenance work @client for 1h",
			Project:         "client",
			Tags:            []string{"maintenance"},
		},
		{
			Timestamp:       today.Add(5 * time.Hour),
			Description:     "Urgent security patch",
			DurationMinutes: 45,
			RawInput:        "Urgent security patch @other for 45m",
			Project:         "other",
			Tags:            []string{"urgent"},
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

	// Reset filters first to ensure clean state
	resetFilterFlags(rootCmd)

	// Simulate: did @client #urgent
	// Parse shorthand filters and execute root command
	args := parseShorthandFilters(rootCmd, []string{"@client", "#urgent"})
	// Args are returned unchanged (for entry creation parsing), but flags are set for filtering
	if len(args) != 2 {
		t.Errorf("Expected 2 remaining args, got: %v", args)
	}

	// Execute the root command (which lists today's entries)
	listEntries(rootCmd, "today", timeutil.Today)

	output := stdout.String()
	// Should find 2 entries: both @client AND #urgent
	if !strings.Contains(output, "Fix critical production bug") {
		t.Error("Expected to find urgent client bug fix")
	}
	if !strings.Contains(output, "Deploy hotfix to production") {
		t.Error("Expected to find urgent client deployment")
	}
	// Should NOT find entries that don't have both filters
	if strings.Contains(output, "Regular maintenance work") {
		t.Error("Should not find maintenance entry (not urgent)")
	}
	if strings.Contains(output, "Urgent security patch") {
		t.Error("Should not find security patch (different project)")
	}
	// Verify filter info in header
	if !strings.Contains(output, "@client") {
		t.Error("Expected header to show @client filter")
	}
	if !strings.Contains(output, "#urgent") {
		t.Error("Expected header to show #urgent filter")
	}
	// Total: 2h + 30m = 150 minutes = 2h 30m
	if !strings.Contains(output, "Total: 2h 30m") {
		t.Errorf("Expected total '2h 30m', got: %s", output)
	}

	// Clean up
	resetFilterFlags(rootCmd)
}

func TestIntegration_WeekCommandWithProjectAndTagFilters(t *testing.T) {
	// Test: did w --project acme --tag review
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	// Get start of current week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	thisWeek := time.Date(monday.Year(), monday.Month(), monday.Day(), 10, 0, 0, 0, monday.Location())

	entries := []entry.Entry{
		{
			Timestamp:       thisWeek,
			Description:     "Code review for new feature",
			DurationMinutes: 60,
			RawInput:        "Code review for new feature @acme for 1h",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       thisWeek.AddDate(0, 0, 1),
			Description:     "Design review meeting",
			DurationMinutes: 90,
			RawInput:        "Design review meeting @acme for 1h30m",
			Project:         "acme",
			Tags:            []string{"review", "meeting"},
		},
		{
			Timestamp:       thisWeek.AddDate(0, 0, 2),
			Description:     "Bug fix in payment module",
			DurationMinutes: 120,
			RawInput:        "Bug fix in payment module @acme for 2h",
			Project:         "acme",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       thisWeek.AddDate(0, 0, 3),
			Description:     "Architecture review session",
			DurationMinutes: 45,
			RawInput:        "Architecture review session @client for 45m",
			Project:         "client",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -10), // Last week
			Description:     "Old code review",
			DurationMinutes: 30,
			RawInput:        "Old code review @acme for 30m",
			Project:         "acme",
			Tags:            []string{"review"},
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

	// Reset filters first
	resetFilterFlags(rootCmd)

	// Set filters: --project acme --tag review
	_ = rootCmd.PersistentFlags().Set("project", "acme")
	_ = rootCmd.PersistentFlags().Set("tag", "review")

	// Execute the week command
	listEntries(rootCmd, "this week", timeutil.ThisWeek)

	output := stdout.String()
	// Should find 2 entries: this week AND @acme AND #review
	if !strings.Contains(output, "Code review for new feature") {
		t.Error("Expected to find code review for acme")
	}
	if !strings.Contains(output, "Design review meeting") {
		t.Error("Expected to find design review for acme")
	}
	// Should NOT find entries that don't match all filters
	if strings.Contains(output, "Bug fix in payment module") {
		t.Error("Should not find bug fix (no review tag)")
	}
	if strings.Contains(output, "Architecture review session") {
		t.Error("Should not find architecture review (different project)")
	}
	if strings.Contains(output, "Old code review") {
		t.Error("Should not find old review (not this week)")
	}
	// Verify filters in header
	if !strings.Contains(output, "@acme") {
		t.Error("Expected header to show @acme filter")
	}
	if !strings.Contains(output, "#review") {
		t.Error("Expected header to show #review filter")
	}
	// Total: 1h + 1h30m = 150 minutes = 2h 30m
	if !strings.Contains(output, "Total: 2h 30m") {
		t.Errorf("Expected total '2h 30m', got: %s", output)
	}

	// Clean up
	resetFilterFlags(rootCmd)
}

func TestIntegration_SearchWithProjectInKeyword(t *testing.T) {
	// Test: did search "bug @acme" - treats @acme as part of the search keyword
	// Note: Search command doesn't currently support project/tag filters,
	// so @acme is treated as a literal part of the search term
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entries := []entry.Entry{
		{
			Timestamp:       time.Now().AddDate(0, 0, -1),
			Description:     "Fix bug @acme in user authentication",
			DurationMinutes: 90,
			RawInput:        "Fix bug @acme in user authentication for 1h30m",
			Project:         "acme",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       time.Now().AddDate(0, 0, -2),
			Description:     "Investigate bug in payment flow",
			DurationMinutes: 60,
			RawInput:        "Investigate bug in payment flow @client for 1h",
			Project:         "client",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       time.Now().AddDate(0, 0, -3),
			Description:     "Refactor code in acme project",
			DurationMinutes: 120,
			RawInput:        "Refactor code in acme project for 2h",
			Project:         "acme",
			Tags:            []string{"refactor"},
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

	// Search for entries containing "bug" (case-insensitive)
	searchEntries(searchCmd, []string{"bug"})

	output := stdout.String()
	// Should find entries with "bug" in description
	if !strings.Contains(output, "Fix bug @acme in user authentication") {
		t.Error("Expected to find acme bug fix")
	}
	if !strings.Contains(output, "Investigate bug in payment flow") {
		t.Error("Expected to find client bug investigation")
	}
	// Should NOT find refactor entry (no "bug" in description)
	if strings.Contains(output, "Refactor code") {
		t.Error("Should not find refactor entry")
	}
	// Should display project information
	if !strings.Contains(output, "@acme") {
		t.Error("Expected to see @acme in results")
	}
	if !strings.Contains(output, "@client") {
		t.Error("Expected to see @client in results")
	}
	// Total: 1h30m + 1h = 150 minutes = 2h 30m
	if !strings.Contains(output, "Total: 2h 30m") {
		t.Errorf("Expected total '2h 30m', got: %s", output)
	}
}

func TestIntegration_CombinedFiltersNoResults(t *testing.T) {
	// Edge case: Filters that don't match any entries
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "Regular development work",
			DurationMinutes: 120,
			RawInput:        "Regular development work @acme for 2h",
			Project:         "acme",
			Tags:            []string{"development"},
		},
		{
			Timestamp:       today.Add(2 * time.Hour),
			Description:     "Code review session",
			DurationMinutes: 45,
			RawInput:        "Code review session @client for 45m",
			Project:         "client",
			Tags:            []string{"review"},
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

	// Reset and set filters that won't match anything
	resetFilterFlags(rootCmd)
	_ = rootCmd.PersistentFlags().Set("project", "nonexistent")
	_ = rootCmd.PersistentFlags().Set("tag", "urgent")

	// List today's entries with filters
	listEntries(rootCmd, "today", timeutil.Today)

	output := stdout.String()
	// Should show no entries found message
	if !strings.Contains(output, "No entries found") {
		t.Errorf("Expected 'No entries found' message, got: %s", output)
	}
	// Should still show filter info
	if !strings.Contains(output, "@nonexistent") {
		t.Error("Expected header to show @nonexistent filter")
	}
	if !strings.Contains(output, "#urgent") {
		t.Error("Expected header to show #urgent filter")
	}

	// Clean up
	resetFilterFlags(rootCmd)
}

func TestIntegration_MultipleTagFiltersANDLogic(t *testing.T) {
	// Test that multiple tag filters use AND logic (entry must have ALL tags)
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "Critical bug fix",
			DurationMinutes: 90,
			RawInput:        "Critical bug fix @acme for 1h30m",
			Project:         "acme",
			Tags:            []string{"urgent", "bugfix"},
		},
		{
			Timestamp:       today.Add(2 * time.Hour),
			Description:     "Regular bug fix",
			DurationMinutes: 60,
			RawInput:        "Regular bug fix @acme for 1h",
			Project:         "acme",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       today.Add(3 * time.Hour),
			Description:     "Urgent feature request",
			DurationMinutes: 120,
			RawInput:        "Urgent feature request @acme for 2h",
			Project:         "acme",
			Tags:            []string{"urgent", "feature"},
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

	// Reset and set multiple tag filters
	resetFilterFlags(rootCmd)
	_ = rootCmd.PersistentFlags().Set("project", "acme")
	_ = rootCmd.PersistentFlags().Set("tag", "urgent")
	_ = rootCmd.PersistentFlags().Set("tag", "bugfix")

	// List today's entries
	listEntries(rootCmd, "today", timeutil.Today)

	output := stdout.String()
	// Should only find the entry with BOTH urgent AND bugfix tags
	if !strings.Contains(output, "Critical bug fix") {
		t.Error("Expected to find critical bug fix (has both tags)")
	}
	// Should NOT find entries with only one of the tags
	if strings.Contains(output, "Regular bug fix") {
		t.Error("Should not find regular bug fix (missing urgent tag)")
	}
	if strings.Contains(output, "Urgent feature request") {
		t.Error("Should not find urgent feature (missing bugfix tag)")
	}
	// Verify both tags shown in header
	if !strings.Contains(output, "#urgent") {
		t.Error("Expected header to show #urgent filter")
	}
	if !strings.Contains(output, "#bugfix") {
		t.Error("Expected header to show #bugfix filter")
	}
	// Total: 1h30m
	if !strings.Contains(output, "Total: 1h 30m") {
		t.Errorf("Expected total '1h 30m', got: %s", output)
	}

	// Clean up
	resetFilterFlags(rootCmd)
}

func TestSearchEntries_CorruptedStorageWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a file with both valid and corrupted entries
	validEntry := `{"timestamp":"2024-01-15T10:00:00Z","description":"Valid entry","duration_minutes":60,"raw_input":"Valid entry for 1h"}`
	corruptedLine := `{invalid json}`
	content := validEntry + "\n" + corruptedLine + "\n"
	if err := os.WriteFile(storagePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
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

	searchEntries(searchCmd, []string{"Valid"})

	// Should show warning about corrupted line in stderr
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "corrupted line") {
		t.Errorf("Expected corruption warning in stderr, got: %s", stderrOutput)
	}

	// Should still show valid entry in stdout
	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Valid entry") {
		t.Errorf("Expected valid entry in stdout, got: %s", stdoutOutput)
	}
}
