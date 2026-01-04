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

// Helper function to create test entries for report testing
func createReportTestEntries(t *testing.T, storagePath string) []entry.Entry {
	t.Helper()

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
			Timestamp:       now.AddDate(0, 0, -6), // 6 days ago
			Description:     "Bug fix in authentication",
			DurationMinutes: 90,
			RawInput:        "Bug fix in authentication for 1h30m",
			Project:         "acme",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -5), // 5 days ago
			Description:     "Team meeting to discuss roadmap",
			DurationMinutes: 45,
			RawInput:        "Team meeting to discuss roadmap for 45m",
			Project:         "",
			Tags:            []string{},
		},
		{
			Timestamp:       now.AddDate(0, 0, -4), // 4 days ago
			Description:     "Design review session",
			DurationMinutes: 120,
			RawInput:        "Design review session for 2h",
			Project:         "client",
			Tags:            []string{"review", "design"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -3), // 3 days ago
			Description:     "Code refactoring in API module",
			DurationMinutes: 150,
			RawInput:        "Code refactoring in API module for 2h30m",
			Project:         "acme",
			Tags:            []string{"refactor"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -2), // 2 days ago
			Description:     "Client meeting",
			DurationMinutes: 30,
			RawInput:        "Client meeting for 30m",
			Project:         "client",
			Tags:            []string{},
		},
		{
			Timestamp:       now.AddDate(0, 0, -1), // 1 day ago
			Description:     "Testing new features",
			DurationMinutes: 75,
			RawInput:        "Testing new features for 1h15m",
			Project:         "",
			Tags:            []string{"testing"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	return entries
}

// Tests for single project report (did report @project)

func TestReport_SingleProject_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report @acme
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show header with project name
	if !strings.Contains(output, "Report for project '@acme'") {
		t.Errorf("Expected report header for @acme, got: %s", output)
	}
	// Should find 3 acme entries
	if !strings.Contains(output, "Code review for feature X") {
		t.Error("Expected to find code review entry")
	}
	if !strings.Contains(output, "Bug fix in authentication") {
		t.Error("Expected to find bug fix entry")
	}
	if !strings.Contains(output, "Code refactoring in API module") {
		t.Error("Expected to find refactoring entry")
	}
	// Should show total (60 + 90 + 150 = 300 minutes = 5h)
	if !strings.Contains(output, "Total: 5h") {
		t.Errorf("Expected 'Total: 5h', got: %s", output)
	}
	if !strings.Contains(output, "3 entrys") {
		t.Errorf("Expected '3 entrys', got: %s", output)
	}

	resetFilterFlags(reportCmd)
}

func TestReport_SingleProject_WithDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report @acme --last 5
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	_ = reportCmd.Flags().Set("last", "5")
	defer func() { _ = reportCmd.Flags().Set("last", "0") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show date filter in header
	if !strings.Contains(output, "last 5 days") {
		t.Errorf("Expected 'last 5 days' in header, got: %s", output)
	}
	// Should find only entries within last 5 days (refactoring from 3 days ago)
	if !strings.Contains(output, "Code refactoring in API module") {
		t.Error("Expected to find refactoring entry (3 days ago)")
	}
	// Should NOT find entries older than 5 days
	if strings.Contains(output, "Code review for feature X") {
		t.Error("Should not find code review entry (7 days ago)")
	}
	// Total should be just the refactoring entry (150 minutes = 2h 30m)
	if !strings.Contains(output, "Total: 2h 30m") {
		t.Errorf("Expected 'Total: 2h 30m', got: %s", output)
	}

	resetFilterFlags(reportCmd)
}

func TestReport_SingleProject_NoResults(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report @nonexistent
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "nonexistent")
	runReport(reportCmd, []string{})

	output := stdout.String()
	if !strings.Contains(output, "No entries found for project '@nonexistent'") {
		t.Errorf("Expected no results message, got: %s", output)
	}

	resetFilterFlags(reportCmd)
}

// Tests for single tag report (did report #tag)

func TestReport_SingleTag_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report #review
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("tag", "review")
	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show header with tag name
	if !strings.Contains(output, "Report for tag '#review'") {
		t.Errorf("Expected report header for #review, got: %s", output)
	}
	// Should find 2 review entries
	if !strings.Contains(output, "Code review for feature X") {
		t.Error("Expected to find code review entry")
	}
	if !strings.Contains(output, "Design review session") {
		t.Error("Expected to find design review entry")
	}
	// Should show total (60 + 120 = 180 minutes = 3h)
	if !strings.Contains(output, "Total: 3h") {
		t.Errorf("Expected 'Total: 3h', got: %s", output)
	}
	if !strings.Contains(output, "2 entrys") {
		t.Errorf("Expected '2 entrys', got: %s", output)
	}

	resetFilterFlags(reportCmd)
}

func TestReport_SingleTag_MultipleTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report #review #design (ANDed together)
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("tag", "review")
	_ = reportCmd.Root().PersistentFlags().Set("tag", "design")
	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show header with both tags
	if !strings.Contains(output, "Report for tags '#review, #design'") {
		t.Errorf("Expected report header with both tags, got: %s", output)
	}
	// Should find only entry with BOTH tags
	if !strings.Contains(output, "Design review session") {
		t.Error("Expected to find design review entry (has both tags)")
	}
	// Should NOT find entries with only one tag
	if strings.Contains(output, "Code review for feature X") {
		t.Error("Should not find code review (only has review tag, not design)")
	}
	// Total should be just the design review (120 minutes = 2h)
	if !strings.Contains(output, "Total: 2h") {
		t.Errorf("Expected 'Total: 2h', got: %s", output)
	}

	resetFilterFlags(reportCmd)
}

func TestReport_SingleTag_WithDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report #review --last 5
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("tag", "review")
	_ = reportCmd.Flags().Set("last", "5")
	defer func() { _ = reportCmd.Flags().Set("last", "0") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show date filter in header
	if !strings.Contains(output, "last 5 days") {
		t.Errorf("Expected 'last 5 days' in header, got: %s", output)
	}
	// Should find design review (4 days ago)
	if !strings.Contains(output, "Design review session") {
		t.Error("Expected to find design review entry")
	}
	// Should NOT find code review (7 days ago, outside range)
	if strings.Contains(output, "Code review for feature X") {
		t.Error("Should not find code review entry (too old)")
	}

	resetFilterFlags(reportCmd)
}

func TestReport_SingleTag_NoResults(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report #nonexistent
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("tag", "nonexistent")
	runReport(reportCmd, []string{})

	output := stdout.String()
	if !strings.Contains(output, "No entries found for tag '#nonexistent'") {
		t.Errorf("Expected no results message, got: %s", output)
	}

	resetFilterFlags(reportCmd)
}

// Tests for grouped by project report (did report --by project)

func TestReport_GroupByProject_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by project
	_ = reportCmd.Flags().Set("by", "project")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show header
	if !strings.Contains(output, "Report grouped by project") {
		t.Errorf("Expected grouped report header, got: %s", output)
	}
	// Should show all projects
	if !strings.Contains(output, "@acme") {
		t.Error("Expected to find @acme project")
	}
	if !strings.Contains(output, "@client") {
		t.Error("Expected to find @client project")
	}
	if !strings.Contains(output, "(no project)") {
		t.Error("Expected to find (no project) group")
	}
	// Should show project totals (acme: 5h, client: 2h30m, no project: 2h)
	if !strings.Contains(output, "5h") {
		t.Error("Expected to find 5h total for acme")
	}
	// Should show grand total (300 + 150 + 120 = 570 minutes = 9h 30m)
	if !strings.Contains(output, "Grand Total: 9h 30m") {
		t.Errorf("Expected 'Grand Total: 9h 30m', got: %s", output)
	}
	if !strings.Contains(output, "7 entrys") {
		t.Errorf("Expected '7 entrys' in grand total, got: %s", output)
	}
	if !strings.Contains(output, "3 projects") {
		t.Errorf("Expected '3 projects' in grand total, got: %s", output)
	}
}

func TestReport_GroupByProject_Sorted(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by project
	_ = reportCmd.Flags().Set("by", "project")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Projects should be sorted by time (descending)
	// acme: 300 minutes, client: 150 minutes, (no project): 120 minutes
	acmePos := strings.Index(output, "@acme")
	clientPos := strings.Index(output, "@client")
	noProjectPos := strings.Index(output, "(no project)")

	if acmePos < 0 || clientPos < 0 || noProjectPos < 0 {
		t.Fatal("Expected all projects to be in output")
	}

	// acme should come before client, client before (no project)
	if acmePos > clientPos {
		t.Error("Expected @acme to appear before @client (sorted by time)")
	}
	if clientPos > noProjectPos {
		t.Error("Expected @client to appear before (no project) (sorted by time)")
	}
}

func TestReport_GroupByProject_WithDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by project --last 4
	_ = reportCmd.Flags().Set("by", "project")
	_ = reportCmd.Flags().Set("last", "4")
	defer func() {
		_ = reportCmd.Flags().Set("by", "")
		_ = reportCmd.Flags().Set("last", "0")
	}()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show date filter in header
	if !strings.Contains(output, "last 4 days") {
		t.Errorf("Expected 'last 4 days' in header, got: %s", output)
	}
	// Should only include entries from last 4 days
	// acme: 1 entry (refactoring, 150min), client: 2 entries (150min), (no project): 1 entry (75min)
	if !strings.Contains(output, "@acme") {
		t.Error("Expected to find @acme")
	}
	if !strings.Contains(output, "@client") {
		t.Error("Expected to find @client")
	}
	if !strings.Contains(output, "(no project)") {
		t.Error("Expected to find (no project)")
	}
}

func TestReport_GroupByProject_EmptyResults(t *testing.T) {
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

	// Test: did report --by project
	_ = reportCmd.Flags().Set("by", "project")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	if !strings.Contains(output, "No entries found") {
		t.Errorf("Expected no entries message, got: %s", output)
	}
}

// Tests for grouped by tag report (did report --by tag)

func TestReport_GroupByTag_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by tag
	_ = reportCmd.Flags().Set("by", "tag")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show header
	if !strings.Contains(output, "Report grouped by tag") {
		t.Errorf("Expected grouped by tag header, got: %s", output)
	}
	// Should show all tags
	if !strings.Contains(output, "#review") {
		t.Error("Expected to find #review tag")
	}
	if !strings.Contains(output, "#bugfix") {
		t.Error("Expected to find #bugfix tag")
	}
	if !strings.Contains(output, "#refactor") {
		t.Error("Expected to find #refactor tag")
	}
	if !strings.Contains(output, "#design") {
		t.Error("Expected to find #design tag")
	}
	if !strings.Contains(output, "#testing") {
		t.Error("Expected to find #testing tag")
	}
	if !strings.Contains(output, "(no tags)") {
		t.Error("Expected to find (no tags) group")
	}
	// Should show grand total (all 7 entries)
	if !strings.Contains(output, "Grand Total:") {
		t.Error("Expected grand total")
	}
	if !strings.Contains(output, "7 entrys") {
		t.Errorf("Expected '7 entrys' in grand total, got: %s", output)
	}
}

func TestReport_GroupByTag_MultipleTagsPerEntry(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by tag
	_ = reportCmd.Flags().Set("by", "tag")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Design review session has both 'review' and 'design' tags (120 minutes)
	// So both #review and #design should show that entry
	// This means the entry count in each group may sum to more than total entries
	// but grand total should show unique entry count

	// Check that grand total shows unique entries (7), not sum of group counts
	if !strings.Contains(output, "7 entrys") {
		t.Errorf("Expected grand total to show 7 unique entries, got: %s", output)
	}
}

func TestReport_GroupByTag_Sorted(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by tag
	_ = reportCmd.Flags().Set("by", "tag")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Tags should be sorted by total time (descending)
	// review: 180 minutes (60 + 120), refactor: 150, design: 120, bugfix: 90, testing: 75, (no tags): 75
	reviewPos := strings.Index(output, "#review")
	refactorPos := strings.Index(output, "#refactor")

	if reviewPos < 0 || refactorPos < 0 {
		t.Fatal("Expected review and refactor tags to be in output")
	}

	// review should come before refactor (sorted by time)
	if reviewPos > refactorPos {
		t.Error("Expected #review to appear before #refactor (sorted by time)")
	}
}

func TestReport_GroupByTag_WithDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by tag --last 4
	_ = reportCmd.Flags().Set("by", "tag")
	_ = reportCmd.Flags().Set("last", "4")
	defer func() {
		_ = reportCmd.Flags().Set("by", "")
		_ = reportCmd.Flags().Set("last", "0")
	}()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show date filter in header
	if !strings.Contains(output, "last 4 days") {
		t.Errorf("Expected 'last 4 days' in header, got: %s", output)
	}
	// Should only include entries from last 4 days
	// Should have: design, review, refactor, testing, (no tags)
	// Should NOT have: bugfix (from 6 days ago, outside range)
	if strings.Contains(output, "#bugfix") {
		t.Error("Should not find #bugfix tag (entry too old)")
	}
}

// Tests for error handling

func TestReport_NoFiltersProvided(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report (no filters)
	resetFilterFlags(reportCmd)
	runReport(reportCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called when no filters provided")
	}
	if !strings.Contains(stderr.String(), "No filters specified") {
		t.Errorf("Expected 'No filters specified' error, got: %s", stderr.String())
	}

	resetFilterFlags(reportCmd)
}

func TestReport_InvalidByFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by invalid
	_ = reportCmd.Flags().Set("by", "invalid")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid --by value")
	}
	if !strings.Contains(stderr.String(), "Invalid --by value") {
		t.Errorf("Expected invalid --by error, got: %s", stderr.String())
	}
}

func TestReport_ConflictingFlags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by project @acme (conflicting)
	resetFilterFlags(reportCmd)
	_ = reportCmd.Flags().Set("by", "project")
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	defer func() {
		_ = reportCmd.Flags().Set("by", "")
	}()

	runReport(reportCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for conflicting flags")
	}
	if !strings.Contains(stderr.String(), "Cannot use --by with --project or --tag") {
		t.Errorf("Expected conflicting flags error, got: %s", stderr.String())
	}

	resetFilterFlags(reportCmd)
}

func TestReport_ConflictingDateFlags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report @acme --last 7 --from 2024-01-01 (conflicting)
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	_ = reportCmd.Flags().Set("last", "7")
	_ = reportCmd.Flags().Set("from", "2024-01-01")
	defer func() {
		_ = reportCmd.Flags().Set("last", "0")
		_ = reportCmd.Flags().Set("from", "")
	}()

	runReport(reportCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for conflicting date flags")
	}
	if !strings.Contains(stderr.String(), "Cannot use --last with --from or --to") {
		t.Errorf("Expected conflicting date flags error, got: %s", stderr.String())
	}

	resetFilterFlags(reportCmd)
}

func TestReport_InvalidFromDate(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report @acme --from invalid-date
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	_ = reportCmd.Flags().Set("from", "invalid-date")
	defer func() { _ = reportCmd.Flags().Set("from", "") }()

	runReport(reportCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid --from date")
	}
	if !strings.Contains(stderr.String(), "Invalid --from date") {
		t.Errorf("Expected invalid date error, got: %s", stderr.String())
	}

	resetFilterFlags(reportCmd)
}

func TestReport_InvalidToDate(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report @acme --to invalid-date
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	_ = reportCmd.Flags().Set("to", "invalid-date")
	defer func() { _ = reportCmd.Flags().Set("to", "") }()

	runReport(reportCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid --to date")
	}
	if !strings.Contains(stderr.String(), "Invalid --to date") {
		t.Errorf("Expected invalid date error, got: %s", stderr.String())
	}

	resetFilterFlags(reportCmd)
}

func TestReport_StoragePathError(t *testing.T) {
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

	// Test: did report @acme
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	runReport(reportCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for storage path error")
	}
	if !strings.Contains(stderr.String(), "Failed to determine storage location") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}

	resetFilterFlags(reportCmd)
}

func TestReport_ReadEntriesError(t *testing.T) {
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

	// Test: did report @acme
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	runReport(reportCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for read entries error")
	}
	if !strings.Contains(stderr.String(), "Failed to read entries") {
		t.Errorf("Expected read entries error, got: %s", stderr.String())
	}

	resetFilterFlags(reportCmd)
}

// Tests for output formatting

func TestReport_OutputFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report @acme
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should have separators (60 equals signs)
	if !strings.Contains(output, strings.Repeat("=", 60)) {
		t.Error("Expected separator line in output")
	}
	// Should have entry indices ([1], [2], etc.)
	if !strings.Contains(output, "[1]") {
		t.Error("Expected entry index [1] in output")
	}
	// Should have project and tag markers
	if !strings.Contains(output, "@acme") {
		t.Error("Expected project marker @acme in output")
	}
	if !strings.Contains(output, "#review") {
		t.Error("Expected tag marker #review in output")
	}
	// Should have duration in human-readable format (1h, 30m, etc.)
	if !strings.Contains(output, "1h") || !strings.Contains(output, "30m") {
		t.Error("Expected human-readable duration format")
	}

	resetFilterFlags(reportCmd)
}

func TestReport_GroupedOutputFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createReportTestEntries(t, storagePath)

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

	// Test: did report --by project
	_ = reportCmd.Flags().Set("by", "project")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should have report header
	if !strings.Contains(output, "Report grouped by project") {
		t.Error("Expected grouped report header")
	}
	// Should have separators
	if !strings.Contains(output, strings.Repeat("=", 60)) {
		t.Error("Expected separator line in output")
	}
	// Should have grand total line
	if !strings.Contains(output, "Grand Total:") {
		t.Error("Expected grand total line")
	}
	// Should show project names with @ prefix
	if !strings.Contains(output, "@acme") {
		t.Error("Expected @acme in grouped output")
	}
	// Should show entry counts for each group
	if !strings.Contains(output, "entrys") {
		t.Error("Expected entry count in grouped output")
	}
}

// Tests for edge cases

func TestReport_EntriesWithNoProject(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "Work without project",
			DurationMinutes: 60,
			RawInput:        "Work without project for 1h",
			Project:         "",
			Tags:            []string{"general"},
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

	// Test: did report --by project
	_ = reportCmd.Flags().Set("by", "project")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show (no project) group
	if !strings.Contains(output, "(no project)") {
		t.Errorf("Expected '(no project)' group, got: %s", output)
	}
}

func TestReport_EntriesWithNoTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "Work without tags",
			DurationMinutes: 60,
			RawInput:        "Work without tags for 1h",
			Project:         "acme",
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

	// Test: did report --by tag
	_ = reportCmd.Flags().Set("by", "tag")
	defer func() { _ = reportCmd.Flags().Set("by", "") }()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should show (no tags) group
	if !strings.Contains(output, "(no tags)") {
		t.Errorf("Expected '(no tags)' group, got: %s", output)
	}
}

func TestReport_SoftDeletedEntriesExcluded(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "Active entry",
			DurationMinutes: 60,
			RawInput:        "Active entry for 1h",
			Project:         "acme",
			Tags:            []string{"active"},
			DeletedAt:       nil, // Not deleted
		},
		{
			Timestamp:       now.Add(-1 * time.Hour),
			Description:     "Deleted entry",
			DurationMinutes: 30,
			RawInput:        "Deleted entry for 30m",
			Project:         "acme",
			Tags:            []string{"deleted"},
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

	// Test: did report @acme
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should find active entry
	if !strings.Contains(output, "Active entry") {
		t.Error("Expected to find active entry")
	}
	// Should NOT find deleted entry
	if strings.Contains(output, "Deleted entry") {
		t.Error("Should not find soft-deleted entry")
	}
	// Total should be 1h (only active entry)
	if !strings.Contains(output, "Total: 1h") {
		t.Errorf("Expected 'Total: 1h', got: %s", output)
	}

	resetFilterFlags(reportCmd)
}

func TestReport_DateRangeFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			Description:     "January entry",
			DurationMinutes: 60,
			RawInput:        "January entry for 1h",
			Project:         "acme",
			Tags:            []string{"test"},
		},
		{
			Timestamp:       time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC),
			Description:     "February entry",
			DurationMinutes: 90,
			RawInput:        "February entry for 1h30m",
			Project:         "acme",
			Tags:            []string{"test"},
		},
		{
			Timestamp:       time.Date(2024, 3, 20, 10, 0, 0, 0, time.UTC),
			Description:     "March entry",
			DurationMinutes: 120,
			RawInput:        "March entry for 2h",
			Project:         "acme",
			Tags:            []string{"test"},
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

	// Test: did report @acme --from 2024-02-01 --to 2024-02-28
	resetFilterFlags(reportCmd)
	_ = reportCmd.Root().PersistentFlags().Set("project", "acme")
	_ = reportCmd.Flags().Set("from", "2024-02-01")
	_ = reportCmd.Flags().Set("to", "2024-02-28")
	defer func() {
		_ = reportCmd.Flags().Set("from", "")
		_ = reportCmd.Flags().Set("to", "")
	}()

	runReport(reportCmd, []string{})

	output := stdout.String()
	// Should only find February entry
	if !strings.Contains(output, "February entry") {
		t.Error("Expected to find February entry")
	}
	if strings.Contains(output, "January entry") {
		t.Error("Should not find January entry (before range)")
	}
	if strings.Contains(output, "March entry") {
		t.Error("Should not find March entry (after range)")
	}
	// Total should be 1h30m (only February entry)
	if !strings.Contains(output, "Total: 1h 30m") {
		t.Errorf("Expected 'Total: 1h 30m', got: %s", output)
	}

	resetFilterFlags(reportCmd)
}
