package handlers

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/cli"
	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/service"
)

func setupTestDeps(t *testing.T) (*cli.Deps, *bytes.Buffer, *bytes.Buffer, *int) {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	cfg := config.DefaultConfig()

	services := &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(filepath.Join(tmpDir, "config.yaml"), cfg),
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := 0

	deps := &cli.Deps{
		Stdout:   stdout,
		Stderr:   stderr,
		Stdin:    strings.NewReader(""),
		Exit:     func(code int) { exitCode = code },
		Services: services,
		Config:   cfg,
	}

	return deps, stdout, stderr, &exitCode
}

// setupBrokenDeps creates deps with a storage path that's a directory (causes errors)
func setupBrokenDeps(t *testing.T) (*cli.Deps, *bytes.Buffer, *bytes.Buffer, *int) {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	cfg := config.DefaultConfig()

	// Create a directory where the storage file should be
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	services := &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(filepath.Join(tmpDir, "config.yaml"), cfg),
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := 0

	deps := &cli.Deps{
		Stdout:   stdout,
		Stderr:   stderr,
		Stdin:    strings.NewReader(""),
		Exit:     func(code int) { exitCode = code },
		Services: services,
		Config:   cfg,
	}

	return deps, stdout, stderr, &exitCode
}

// setupBrokenTimerDeps creates deps with a timer path that's a directory (causes errors)
func setupBrokenTimerDeps(t *testing.T) (*cli.Deps, *bytes.Buffer, *bytes.Buffer, *int) {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	cfg := config.DefaultConfig()

	// Create a directory where the timer file should be
	if err := os.MkdirAll(timerPath, 0755); err != nil {
		t.Fatal(err)
	}

	services := &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(filepath.Join(tmpDir, "config.yaml"), cfg),
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := 0

	deps := &cli.Deps{
		Stdout:   stdout,
		Stderr:   stderr,
		Stdin:    strings.NewReader(""),
		Exit:     func(code int) { exitCode = code },
		Services: services,
		Config:   cfg,
	}

	return deps, stdout, stderr, &exitCode
}

// setupBrokenConfigDeps creates deps with a config path that's a directory (causes errors)
func setupBrokenConfigDeps(t *testing.T) (*cli.Deps, *bytes.Buffer, *bytes.Buffer, *int) {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfg := config.DefaultConfig()

	// Create a directory where the config file should be
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatal(err)
	}

	services := &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(configPath, cfg),
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := 0

	deps := &cli.Deps{
		Stdout:   stdout,
		Stderr:   stderr,
		Stdin:    strings.NewReader(""),
		Exit:     func(code int) { exitCode = code },
		Services: services,
		Config:   cfg,
	}

	return deps, stdout, stderr, &exitCode
}

// setupMultiDayDeps creates deps with entries spanning multiple days
func setupMultiDayDeps(t *testing.T) (*cli.Deps, *bytes.Buffer, *bytes.Buffer, *int) {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	cfg := config.DefaultConfig()

	// Create entries with different dates
	now := time.Now()
	today := now.Format("2006-01-02T15:04:05Z07:00")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02T15:04:05Z07:00")

	content := `{"description":"task today","duration_minutes":60,"timestamp":"` + today + `"}
{"description":"task yesterday","duration_minutes":30,"timestamp":"` + yesterday + `"}
`
	if err := os.WriteFile(storagePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	services := &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(filepath.Join(tmpDir, "config.yaml"), cfg),
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := 0

	deps := &cli.Deps{
		Stdout:   stdout,
		Stderr:   stderr,
		Stdin:    strings.NewReader(""),
		Exit:     func(code int) { exitCode = code },
		Services: services,
		Config:   cfg,
	}

	return deps, stdout, stderr, &exitCode
}

// setupCorruptedDeps creates deps with a storage file containing corrupted lines
func setupCorruptedDeps(t *testing.T) (*cli.Deps, *bytes.Buffer, *bytes.Buffer, *int) {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	cfg := config.DefaultConfig()

	// Use today's date for the valid entries so they show up in date range queries
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")

	// Create a file with a mix of valid and corrupted lines
	corruptedContent := `{"description":"valid entry","duration_minutes":60,"timestamp":"` + now + `","project":"test","tags":["tag1"]}
this is not valid json
{"description":"another valid","duration_minutes":30,"timestamp":"` + now + `"}
also corrupted line
`
	if err := os.WriteFile(storagePath, []byte(corruptedContent), 0644); err != nil {
		t.Fatal(err)
	}

	services := &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(filepath.Join(tmpDir, "config.yaml"), cfg),
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := 0

	deps := &cli.Deps{
		Stdout:   stdout,
		Stderr:   stderr,
		Stdin:    strings.NewReader(""),
		Exit:     func(code int) { exitCode = code },
		Services: services,
		Config:   cfg,
	}

	return deps, stdout, stderr, &exitCode
}

func TestCreateEntry(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "fix bug for 1h")

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Logged:") {
		t.Errorf("expected 'Logged:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "fix bug") {
		t.Errorf("expected 'fix bug' in output, got %q", stdout.String())
	}
}

func TestCreateEntry_WithProjectAndTags(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "fix bug @acme #urgent for 1h")

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "@acme") {
		t.Errorf("expected '@acme' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "#urgent") {
		t.Errorf("expected '#urgent' in output, got %q", stdout.String())
	}
}

func TestCreateEntry_MissingDuration(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	CreateEntry(deps, "fix bug")

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Missing 'for <duration>'") {
		t.Errorf("expected duration error in stderr, got %q", stderr.String())
	}
}

func TestCreateEntry_EmptyDescription(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	// @project #tag with no description still triggers empty description error
	CreateEntry(deps, "@project #tag for 1h")

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Description cannot be empty") {
		t.Errorf("expected empty description error in stderr, got %q", stderr.String())
	}
}

func TestListEntries_Empty(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ListEntries(deps, service.DateRangeSpec{Type: service.DateRangeToday}, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "No entries found") {
		t.Errorf("expected 'No entries found', got %q", stdout.String())
	}
}

func TestListEntries_WithEntries(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Create some entries
	CreateEntry(deps, "task 1 for 1h")
	CreateEntry(deps, "task 2 for 30m")
	stdout.Reset()

	ListEntries(deps, service.DateRangeSpec{Type: service.DateRangeToday}, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "task 1") {
		t.Errorf("expected 'task 1' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "task 2") {
		t.Errorf("expected 'task 2' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Total:") {
		t.Errorf("expected 'Total:' in output, got %q", stdout.String())
	}
}

func TestListEntries_WithFilter(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task 1 @acme for 1h")
	CreateEntry(deps, "task 2 @other for 30m")
	stdout.Reset()

	f := &filter.Filter{Project: "acme"}
	ListEntries(deps, service.DateRangeSpec{Type: service.DateRangeToday}, f)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "task 1") {
		t.Errorf("expected 'task 1' in output, got %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "task 2") {
		t.Errorf("did not expect 'task 2' in output, got %q", stdout.String())
	}
}

func TestEditEntry(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "original task for 1h")
	stdout.Reset()

	EditEntry(deps, "1", "edited task", "")

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Updated entry") {
		t.Errorf("expected 'Updated entry' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "edited task") {
		t.Errorf("expected 'edited task' in output, got %q", stdout.String())
	}
}

func TestEditEntry_InvalidIndex(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	EditEntry(deps, "abc", "new desc", "")

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Invalid index") {
		t.Errorf("expected 'Invalid index' error, got %q", stderr.String())
	}
}

func TestEditEntry_NoChanges(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task for 1h")

	EditEntry(deps, "1", "", "")

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "At least one flag") {
		t.Errorf("expected 'At least one flag' error, got %q", stderr.String())
	}
}

func TestEditEntry_NoEntries(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	EditEntry(deps, "1", "new desc", "")

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "No entries found") {
		t.Errorf("expected 'No entries found' error, got %q", stderr.String())
	}
}

func TestDeleteEntry_InvalidIndex(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	DeleteEntry(deps, "abc", true)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Invalid index") {
		t.Errorf("expected 'Invalid index' error, got %q", stderr.String())
	}
}

func TestDeleteEntry_NegativeIndex(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	DeleteEntry(deps, "-1", true)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Index must be 1 or greater") {
		t.Errorf("expected index error, got %q", stderr.String())
	}
}

func TestDeleteEntry_NotFound(t *testing.T) {
	deps, _, _, exitCode := setupTestDeps(t)

	DeleteEntry(deps, "1", true)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	// Should fail because no entries exist
}

func TestDeleteEntry_WithConfirmation_Yes(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task to delete for 1h")
	stdout.Reset()

	// Simulate "y" input
	deps.Stdin = strings.NewReader("y\n")

	DeleteEntry(deps, "1", false)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Deleted:") {
		t.Errorf("expected 'Deleted:' in output, got %q", stdout.String())
	}
}

func TestDeleteEntry_WithConfirmation_No(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task to keep for 1h")
	stdout.Reset()

	// Simulate "n" input
	deps.Stdin = strings.NewReader("n\n")

	DeleteEntry(deps, "1", false)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "cancelled") {
		t.Errorf("expected 'cancelled' in output, got %q", stdout.String())
	}
}

func TestDeleteEntry_SkipConfirm(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task to delete for 1h")
	stdout.Reset()

	DeleteEntry(deps, "1", true)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Deleted:") {
		t.Errorf("expected 'Deleted:' in output, got %q", stdout.String())
	}
}

func TestRestoreEntry(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task for 1h")
	stdout.Reset()
	DeleteEntry(deps, "1", true)
	stdout.Reset()

	RestoreEntry(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Restored:") {
		t.Errorf("expected 'Restored:' in output, got %q", stdout.String())
	}
}

func TestRestoreEntry_NoDeletedEntries(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	RestoreEntry(deps)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "no deleted entries") {
		t.Errorf("expected 'no deleted entries' error, got %q", stderr.String())
	}
}

func TestListEntries_MultipleEntryWidthAlignment(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Create more than 9 entries to test index width formatting
	for i := 0; i < 12; i++ {
		CreateEntry(deps, "task for 1h")
	}
	stdout.Reset()

	ListEntries(deps, service.DateRangeSpec{Type: service.DateRangeToday}, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	// With 12 entries, indices should be right-aligned
	if !strings.Contains(stdout.String(), "[12]") {
		t.Errorf("expected '[12]' in output, got %q", stdout.String())
	}
}

func TestEditEntry_WithDuration(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "original task for 1h")
	stdout.Reset()

	EditEntry(deps, "1", "", "2h")

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Updated entry") {
		t.Errorf("expected 'Updated entry' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "2h") {
		t.Errorf("expected '2h' in output, got %q", stdout.String())
	}
}

func TestEditEntry_IndexOutOfRange(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task for 1h")

	EditEntry(deps, "99", "new desc", "")

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestDeleteEntry_ShowsEntryBeforeDelete(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "specific task for 45m")
	stdout.Reset()

	DeleteEntry(deps, "1", true)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Entry to delete:") {
		t.Errorf("expected 'Entry to delete:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "specific task") {
		t.Errorf("expected 'specific task' in output, got %q", stdout.String())
	}
}

func TestRestoreEntry_ShowsDetails(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task to restore @project #tag for 1h")
	stdout.Reset()
	DeleteEntry(deps, "1", true)
	stdout.Reset()

	RestoreEntry(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Timestamp:") {
		t.Errorf("expected 'Timestamp:' in output, got %q", stdout.String())
	}
}

// Error case tests using broken storage
func TestCreateEntry_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	CreateEntry(deps, "task for 1h")

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestListEntries_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	ListEntries(deps, service.DateRangeSpec{Type: service.DateRangeToday}, nil)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestRestoreEntry_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	RestoreEntry(deps)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	// Either no entries or storage error
	if !strings.Contains(stderr.String(), "Error") && !strings.Contains(stderr.String(), "error") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestListEntries_CorruptedLines(t *testing.T) {
	deps, stdout, stderr, exitCode := setupCorruptedDeps(t)

	// Use today's date range since the corrupted file uses today's timestamps
	ListEntries(deps, service.DateRangeSpec{Type: service.DateRangeToday}, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	// Should show warning about corrupted lines
	if !strings.Contains(stderr.String(), "Warning:") {
		t.Errorf("expected 'Warning:' in stderr, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "corrupted") {
		t.Errorf("expected 'corrupted' in stderr, got %q", stderr.String())
	}
	// Should still show valid entries
	if !strings.Contains(stdout.String(), "valid entry") {
		t.Errorf("expected 'valid entry' in stdout, got %q", stdout.String())
	}
}

func TestListEntries_MultiDayShowsDate(t *testing.T) {
	deps, stdout, _, exitCode := setupMultiDayDeps(t)

	// Use week range to include both days
	ListEntries(deps, service.DateRangeSpec{Type: service.DateRangeThisWeek}, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	// Should show date column when entries span multiple days
	// The format includes the date in "2006-01-02" format
	if !strings.Contains(stdout.String(), "task today") && !strings.Contains(stdout.String(), "task yesterday") {
		t.Errorf("expected entries in output, got %q", stdout.String())
	}
}

func TestPromptConfirmation_Yes(t *testing.T) {
	out := &bytes.Buffer{}
	stdin := strings.NewReader("y\n")

	result := promptConfirmation(out, stdin)
	if !result {
		t.Error("expected true for 'y' input")
	}
}

func TestPromptConfirmation_CapitalY(t *testing.T) {
	out := &bytes.Buffer{}
	stdin := strings.NewReader("Y\n")

	result := promptConfirmation(out, stdin)
	if !result {
		t.Error("expected true for 'Y' input")
	}
}

func TestPromptConfirmation_No(t *testing.T) {
	out := &bytes.Buffer{}
	stdin := strings.NewReader("n\n")

	result := promptConfirmation(out, stdin)
	if result {
		t.Error("expected false for 'n' input")
	}
}

func TestPromptConfirmation_Empty(t *testing.T) {
	out := &bytes.Buffer{}
	stdin := strings.NewReader("\n")

	result := promptConfirmation(out, stdin)
	if result {
		t.Error("expected false for empty input")
	}
}

func TestPromptConfirmation_EOF(t *testing.T) {
	out := &bytes.Buffer{}
	stdin := strings.NewReader("")

	result := promptConfirmation(out, stdin)
	if result {
		t.Error("expected false for EOF")
	}
}
