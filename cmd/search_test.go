package cmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
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
