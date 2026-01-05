package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// testDeps creates test dependencies with captured output
func testDeps(storagePath string) (*Deps, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
		Config: config.DefaultConfig(),
	}, stdout, stderr
}

// testDepsWithConfig creates test dependencies with a custom config and captured output
func testDepsWithConfig(storagePath string, cfg config.Config) (*Deps, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
		Config: cfg,
	}, stdout, stderr
}

// resetFilterFlags clears all persistent filter flags to avoid test contamination
// Note: StringSlice flags are difficult to reset cleanly in pflag, so we just mark them as unchanged
func resetFilterFlags(cmd *cobra.Command) {
	// Reset project flag
	_ = cmd.Root().PersistentFlags().Set("project", "")

	// For StringSlice tag flag, we need to get the current value and replace it
	// The pflag library accumulates StringSlice values, so we use Replace method if available
	// or manually clear by getting the slice pointer
	tagFlag := cmd.Root().PersistentFlags().Lookup("tag")
	if tagFlag != nil {
		// Cast to stringSliceValue to access Replace method
		if sliceVal, ok := tagFlag.Value.(interface{ Replace([]string) error }); ok {
			_ = sliceVal.Replace([]string{})
		}
		tagFlag.Changed = false
	}
}

// resetTimePeriodFlags clears all time period flags to avoid test contamination
func resetTimePeriodFlags(cmd *cobra.Command) {
	// Reset boolean flags
	_ = cmd.Flags().Set("yesterday", "false")
	_ = cmd.Flags().Set("this-week", "false")
	_ = cmd.Flags().Set("prev-week", "false")
	_ = cmd.Flags().Set("this-month", "false")
	_ = cmd.Flags().Set("prev-month", "false")
	// Reset int flag
	_ = cmd.Flags().Set("last", "0")
	// Reset string flags
	_ = cmd.Flags().Set("from", "")
	_ = cmd.Flags().Set("to", "")
	_ = cmd.Flags().Set("date", "")
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		count    int
		expected string
	}{
		{"singular entry", "entry", 1, "entry"},
		{"plural entries", "entry", 0, "entrys"},
		{"plural entries 2", "entry", 2, "entrys"},
		{"plural entries 10", "entry", 10, "entrys"},
		{"singular item", "item", 1, "item"},
		{"plural items", "item", 5, "items"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pluralize(tt.word, tt.count)
			if result != tt.expected {
				t.Errorf("pluralize(%q, %d) = %q, expected %q", tt.word, tt.count, result, tt.expected)
			}
		})
	}
}

func TestFormatCorruptionWarning(t *testing.T) {
	tests := []struct {
		name     string
		warning  storage.ParseWarning
		expected string
	}{
		{
			name: "short content",
			warning: storage.ParseWarning{
				LineNumber: 5,
				Content:    "invalid json",
				Error:      "unexpected end of JSON",
			},
			expected: "  Line 5: invalid json (error: unexpected end of JSON)",
		},
		{
			name: "exactly 50 chars",
			warning: storage.ParseWarning{
				LineNumber: 10,
				Content:    "12345678901234567890123456789012345678901234567890",
				Error:      "parse error",
			},
			expected: "  Line 10: 12345678901234567890123456789012345678901234567890 (error: parse error)",
		},
		{
			name: "over 50 chars truncated",
			warning: storage.ParseWarning{
				LineNumber: 1,
				Content:    "this is a very long line that exceeds fifty characters and should be truncated",
				Error:      "some error",
			},
			expected: "  Line 1: this is a very long line that exceeds fifty cha... (error: some error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCorruptionWarning(tt.warning)
			if result != tt.expected {
				t.Errorf("formatCorruptionWarning() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestCreateEntry_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"test", "task", "for", "2h"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Logged:") {
		t.Errorf("Expected 'Logged:' in output, got: %s", output)
	}
	if !strings.Contains(output, "test task") {
		t.Errorf("Expected 'test task' in output, got: %s", output)
	}

	// Verify entry was written
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestCreateEntry_MissingFor(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"test", "task", "2h"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Missing 'for <duration>'") {
		t.Errorf("Expected error about missing 'for', got: %s", stderr.String())
	}
}

func TestCreateEntry_EmptyDescription(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	// Input " for 2h" - note leading space so rawInput contains " for "
	createEntry([]string{"", "for", "2h"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Description cannot be empty") {
		t.Errorf("Expected error about empty description, got: %s", stderr.String())
	}
}

func TestCreateEntry_InvalidDuration(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"test", "for", "invalid"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Invalid duration") {
		t.Errorf("Expected error about invalid duration, got: %s", stderr.String())
	}
}

func TestListEntries_NoEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		now := time.Now()
		return now, now.Add(24 * time.Hour)
	})

	if !strings.Contains(stdout.String(), "No entries found") {
		t.Errorf("Expected 'No entries found', got: %s", stdout.String())
	}
}

func TestListEntries_WithEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test entry",
		DurationMinutes: 60,
		RawInput:        "test entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	})

	output := stdout.String()
	if !strings.Contains(output, "test entry") {
		t.Errorf("Expected 'test entry' in output, got: %s", output)
	}
	if !strings.Contains(output, "Total:") {
		t.Errorf("Expected 'Total:' in output, got: %s", output)
	}
}

func TestValidateStorage_Healthy(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create valid entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "valid entry",
		DurationMinutes: 60,
		RawInput:        "valid entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	validateStorage()

	output := stdout.String()
	if !strings.Contains(output, "Storage file is healthy") {
		t.Errorf("Expected 'Storage file is healthy', got: %s", output)
	}
	if !strings.Contains(output, "Valid entries:     1") {
		t.Errorf("Expected 'Valid entries:     1', got: %s", output)
	}
}

func TestValidateStorage_WithCorruption(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Write corrupted content
	content := `{"timestamp":"2024-01-15T10:00:00Z","description":"valid","duration_minutes":60,"raw_input":"valid for 1h"}
invalid json line
`
	if err := os.WriteFile(storagePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	validateStorage()

	output := stdout.String()
	if !strings.Contains(output, "Corrupted entries: 1") {
		t.Errorf("Expected 'Corrupted entries: 1', got: %s", output)
	}
	if !strings.Contains(stderr.String(), "corrupted line") {
		t.Errorf("Expected warning about corrupted lines in stderr, got: %s", stderr.String())
	}
}

func TestEditEntry_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original",
		DurationMinutes: 60,
		RawInput:        "original for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Create a mock command with flags
	_ = editCmd.Flags().Set("description", "updated")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}

	// Verify entry was updated
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "updated" {
		t.Errorf("Expected description 'updated', got: %s", entries[0].Description)
	}
}

func TestEditEntry_InvalidIndex(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("description", "updated")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"99"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "out of range") {
		t.Errorf("Expected 'out of range' error, got: %s", stderr.String())
	}
}

func TestEditEntry_NoFlags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	// Ensure flags are empty
	_ = editCmd.Flags().Set("description", "")
	_ = editCmd.Flags().Set("duration", "")

	editEntry(editCmd, []string{"1"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "At least one flag") {
		t.Errorf("Expected error about missing flags, got: %s", stderr.String())
	}
}

func TestEditEntry_InvalidIndexFormat(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("description", "test")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"abc"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Invalid index") {
		t.Errorf("Expected 'Invalid index' error, got: %s", stderr.String())
	}
}

func TestEditEntry_NoEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("description", "test")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "No entries found") {
		t.Errorf("Expected 'No entries found' error, got: %s", stderr.String())
	}
}

func TestEditEntry_DurationOnly(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original",
		DurationMinutes: 60,
		RawInput:        "original for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("duration", "2h")
	defer func() { _ = editCmd.Flags().Set("duration", "") }()

	editEntry(editCmd, []string{"1"})

	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}

	// Verify entry was updated
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].DurationMinutes != 120 {
		t.Errorf("Expected duration 120, got: %d", entries[0].DurationMinutes)
	}
}

func TestEditEntry_InvalidDuration(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("duration", "invalid")
	defer func() { _ = editCmd.Flags().Set("duration", "") }()

	editEntry(editCmd, []string{"1"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Invalid duration") {
		t.Errorf("Expected 'Invalid duration' error, got: %s", stderr.String())
	}
}

func TestEditEntry_BothFlags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original",
		DurationMinutes: 60,
		RawInput:        "original for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("description", "updated")
	_ = editCmd.Flags().Set("duration", "3h")
	defer func() {
		_ = editCmd.Flags().Set("description", "")
		_ = editCmd.Flags().Set("duration", "")
	}()

	editEntry(editCmd, []string{"1"})

	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}

	// Verify entry was updated
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "updated" || entries[0].DurationMinutes != 180 {
		t.Errorf("Expected updated desc and 180 minutes, got: %s, %d", entries[0].Description, entries[0].DurationMinutes)
	}
}

func TestCreateEntry_StoragePathError(t *testing.T) {
	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", fmt.Errorf("storage path error")
		},
	}
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"test", "for", "1h"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to determine storage location") {
		t.Errorf("Expected storage error, got: %s", stderr.String())
	}
}

func TestCreateEntry_AppendError(t *testing.T) {
	// Use a path that will fail to write
	storagePath := "/nonexistent/path/entries.jsonl"

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"test", "for", "1h"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to save entry") {
		t.Errorf("Expected save error, got: %s", stderr.String())
	}
}

func TestListEntries_StoragePathError(t *testing.T) {
	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", fmt.Errorf("storage path error")
		},
	}
	SetDeps(d)
	defer ResetDeps()

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		now := time.Now()
		return now, now.Add(24 * time.Hour)
	})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to determine storage location") {
		t.Errorf("Expected storage error, got: %s", stderr.String())
	}
}

func TestListEntries_WithCorruptedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Write file with corrupted line
	now := time.Now()
	content := fmt.Sprintf(`{"timestamp":"%s","description":"valid","duration_minutes":60,"raw_input":"valid for 1h"}
invalid json line
`, now.Format(time.RFC3339))
	if err := os.WriteFile(storagePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	})

	// Should show warning in stderr
	if !strings.Contains(stderr.String(), "corrupted line") {
		t.Errorf("Expected warning about corrupted lines, got stderr: %s", stderr.String())
	}
	// Should still show valid entries
	if !strings.Contains(stdout.String(), "valid") {
		t.Errorf("Expected 'valid' entry in output, got: %s", stdout.String())
	}
}

func TestValidateStorage_StoragePathError(t *testing.T) {
	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", fmt.Errorf("storage path error")
		},
	}
	SetDeps(d)
	defer ResetDeps()

	validateStorage()

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to get storage path") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}
}

func TestEditEntry_StoragePathError(t *testing.T) {
	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", fmt.Errorf("storage path error")
		},
	}
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("description", "test")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to determine storage location") {
		t.Errorf("Expected storage error, got: %s", stderr.String())
	}
}

func TestExecute(t *testing.T) {
	// Test Execute function - it just calls rootCmd.Execute()
	// We can't easily test this without side effects, but we can verify it doesn't panic
	// Reset args to avoid side effects from previous tests
	oldArgs := os.Args
	os.Args = []string{"did", "--help"}
	defer func() { os.Args = oldArgs }()

	// Execute should return nil for help
	err := Execute()
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}
}

func TestListEntries_ReadEntriesError(t *testing.T) {
	// Use a directory instead of a file to cause read error
	tmpDir := t.TempDir()

	exitCalled := false
	d, _, stderr := testDeps(tmpDir)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		now := time.Now()
		return now, now.Add(24 * time.Hour)
	})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to read entries") {
		t.Errorf("Expected read error, got: %s", stderr.String())
	}
}

func TestValidateStorage_ValidateError(t *testing.T) {
	// Use a directory instead of a file to cause validation error
	tmpDir := t.TempDir()

	exitCalled := false
	d, _, stderr := testDeps(tmpDir)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	validateStorage()

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to validate storage") {
		t.Errorf("Expected validate error, got: %s", stderr.String())
	}
}

func TestEditEntry_UpdateError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	// Make the directory read-only to cause write error
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Skipf("Cannot change directory permissions: %v", err)
	}
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("description", "updated")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to save updated entry") {
		t.Errorf("Expected save error, got: %s", stderr.String())
	}
}

func TestEditEntry_ReadEntriesWithWarningsError(t *testing.T) {
	// Use a directory instead of a file to cause read error
	tmpDir := t.TempDir()

	exitCalled := false
	d, _, stderr := testDeps(tmpDir)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	_ = editCmd.Flags().Set("description", "test")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to read entries") {
		t.Errorf("Expected read error, got: %s", stderr.String())
	}
}

func TestRootCommand_NoArgs(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Call the root command's Run function directly with no args
	rootCmd.Run(rootCmd, []string{})

	// Should list today's entries (empty)
	if !strings.Contains(stdout.String(), "No entries found") {
		t.Errorf("Expected 'No entries found', got: %s", stdout.String())
	}
}

func TestRootCommand_WithArgs(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Call the root command's Run function with args
	rootCmd.Run(rootCmd, []string{"test", "task", "for", "1h"})

	if !strings.Contains(stdout.String(), "Logged:") {
		t.Errorf("Expected 'Logged:', got: %s", stdout.String())
	}
}

func TestYesterday_Flag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set yesterday flag
	_ = rootCmd.Flags().Set("yesterday", "true")

	rootCmd.Run(rootCmd, []string{})

	if !strings.Contains(stdout.String(), "No entries found for yesterday") {
		t.Errorf("Expected 'No entries found for yesterday', got: %s", stdout.String())
	}
}

func TestThisWeek_Flag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set this-week flag
	_ = rootCmd.Flags().Set("this-week", "true")

	rootCmd.Run(rootCmd, []string{})

	if !strings.Contains(stdout.String(), "No entries found for this week") {
		t.Errorf("Expected 'No entries found for this week' (with date range), got: %s", stdout.String())
	}
}

func TestPrevWeek_Flag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set prev-week flag
	_ = rootCmd.Flags().Set("prev-week", "true")

	rootCmd.Run(rootCmd, []string{})

	if !strings.Contains(stdout.String(), "No entries found for previous week") {
		t.Errorf("Expected 'No entries found for previous week', got: %s", stdout.String())
	}
}

// Test that '--this-week' shows correct date range with monday week start (default)
func TestThisWeek_DateRangeOutput_MondayStart(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Use Monday week start config (default)
	cfg := config.Config{
		WeekStartDay:        "monday",
		Timezone:            "Local",
		DefaultOutputFormat: "",
	}

	d, stdout, _ := testDepsWithConfig(storagePath, cfg)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Calculate expected date range
	now := time.Now()
	start := timeutil.StartOfWeekWithConfig(now, "monday")
	end := timeutil.EndOfWeekWithConfig(now, "monday")

	// Run command with --this-week flag
	_ = rootCmd.Flags().Set("this-week", "true")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Verify output contains date range
	startDate := start.Format("Jan 2")
	endDate := end.Format("Jan 2, 2006")

	if !strings.Contains(output, startDate) {
		t.Errorf("Expected output to contain start date '%s', got: %s", startDate, output)
	}
	if !strings.Contains(output, endDate) {
		t.Errorf("Expected output to contain end date '%s', got: %s", endDate, output)
	}

	// Verify the header shows "this week"
	if !strings.Contains(output, "this week") {
		t.Errorf("Expected output to contain 'this week', got: %s", output)
	}

	// Verify start is a Monday
	if start.Weekday() != time.Monday {
		t.Errorf("Expected week start to be Monday with monday config, got %s", start.Weekday())
	}

	// Verify end is a Sunday
	if end.Weekday() != time.Sunday {
		t.Errorf("Expected week end to be Sunday with monday config, got %s", end.Weekday())
	}
}

// Test that '--this-week' shows correct date range with sunday week start
func TestThisWeek_DateRangeOutput_SundayStart(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Use Sunday week start config
	cfg := config.Config{
		WeekStartDay:        "sunday",
		Timezone:            "Local",
		DefaultOutputFormat: "",
	}

	d, stdout, _ := testDepsWithConfig(storagePath, cfg)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Calculate expected date range
	now := time.Now()
	start := timeutil.StartOfWeekWithConfig(now, "sunday")
	end := timeutil.EndOfWeekWithConfig(now, "sunday")

	// Run command with --this-week flag
	_ = rootCmd.Flags().Set("this-week", "true")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Verify output contains date range
	startDate := start.Format("Jan 2")
	endDate := end.Format("Jan 2, 2006")

	if !strings.Contains(output, startDate) {
		t.Errorf("Expected output to contain start date '%s', got: %s", startDate, output)
	}
	if !strings.Contains(output, endDate) {
		t.Errorf("Expected output to contain end date '%s', got: %s", endDate, output)
	}

	// Verify the header shows "this week"
	if !strings.Contains(output, "this week") {
		t.Errorf("Expected output to contain 'this week', got: %s", output)
	}

	// Verify start is a Sunday
	if start.Weekday() != time.Sunday {
		t.Errorf("Expected week start to be Sunday with sunday config, got %s", start.Weekday())
	}

	// Verify end is a Saturday
	if end.Weekday() != time.Saturday {
		t.Errorf("Expected week end to be Saturday with sunday config, got %s", end.Weekday())
	}
}

// Test that '--prev-week' shows correct date range with monday week start (default)
func TestPrevWeek_DateRangeOutput_MondayStart(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Use Monday week start config (default)
	cfg := config.Config{
		WeekStartDay:        "monday",
		Timezone:            "Local",
		DefaultOutputFormat: "",
	}

	d, stdout, _ := testDepsWithConfig(storagePath, cfg)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Calculate expected date range
	lastWeek := time.Now().AddDate(0, 0, -7)
	start := timeutil.StartOfWeekWithConfig(lastWeek, "monday")
	end := timeutil.EndOfWeekWithConfig(lastWeek, "monday")

	// Run command with --prev-week flag
	_ = rootCmd.Flags().Set("prev-week", "true")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Verify output contains date range
	startDate := start.Format("Jan 2")
	// Handle year boundary cases where start and end may be in different years
	var endDate string
	if start.Year() == end.Year() {
		endDate = end.Format("Jan 2, 2006")
	} else {
		// Both dates should show full year if crossing year boundary
		startDate = start.Format("Jan 2, 2006")
		endDate = end.Format("Jan 2, 2006")
	}

	if !strings.Contains(output, startDate) {
		t.Errorf("Expected output to contain start date '%s', got: %s", startDate, output)
	}
	if !strings.Contains(output, endDate) {
		t.Errorf("Expected output to contain end date '%s', got: %s", endDate, output)
	}

	// Verify the header shows "previous week"
	if !strings.Contains(output, "previous week") {
		t.Errorf("Expected output to contain 'previous week', got: %s", output)
	}

	// Verify start is a Monday
	if start.Weekday() != time.Monday {
		t.Errorf("Expected week start to be Monday with monday config, got %s", start.Weekday())
	}

	// Verify end is a Sunday
	if end.Weekday() != time.Sunday {
		t.Errorf("Expected week end to be Sunday with monday config, got %s", end.Weekday())
	}
}

// Test that '--prev-week' shows correct date range with sunday week start
func TestPrevWeek_DateRangeOutput_SundayStart(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Use Sunday week start config
	cfg := config.Config{
		WeekStartDay:        "sunday",
		Timezone:            "Local",
		DefaultOutputFormat: "",
	}

	d, stdout, _ := testDepsWithConfig(storagePath, cfg)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Calculate expected date range
	lastWeek := time.Now().AddDate(0, 0, -7)
	start := timeutil.StartOfWeekWithConfig(lastWeek, "sunday")
	end := timeutil.EndOfWeekWithConfig(lastWeek, "sunday")

	// Run command with --prev-week flag
	_ = rootCmd.Flags().Set("prev-week", "true")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Verify output contains date range
	startDate := start.Format("Jan 2")
	// Handle year boundary cases where start and end may be in different years
	var endDate string
	if start.Year() == end.Year() {
		endDate = end.Format("Jan 2, 2006")
	} else {
		// Both dates should show full year if crossing year boundary
		startDate = start.Format("Jan 2, 2006")
		endDate = end.Format("Jan 2, 2006")
	}

	if !strings.Contains(output, startDate) {
		t.Errorf("Expected output to contain start date '%s', got: %s", startDate, output)
	}
	if !strings.Contains(output, endDate) {
		t.Errorf("Expected output to contain end date '%s', got: %s", endDate, output)
	}

	// Verify the header shows "previous week"
	if !strings.Contains(output, "previous week") {
		t.Errorf("Expected output to contain 'previous week', got: %s", output)
	}

	// Verify start is a Sunday
	if start.Weekday() != time.Sunday {
		t.Errorf("Expected week start to be Sunday with sunday config, got %s", start.Weekday())
	}

	// Verify end is a Saturday
	if end.Weekday() != time.Saturday {
		t.Errorf("Expected week end to be Saturday with sunday config, got %s", end.Weekday())
	}
}

func TestThisMonth_Flag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set this-month flag
	_ = rootCmd.Flags().Set("this-month", "true")

	rootCmd.Run(rootCmd, []string{})

	if !strings.Contains(stdout.String(), "No entries found for this month") {
		t.Errorf("Expected 'No entries found for this month' (with date range), got: %s", stdout.String())
	}
}

func TestPrevMonth_Flag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set prev-month flag
	_ = rootCmd.Flags().Set("prev-month", "true")

	rootCmd.Run(rootCmd, []string{})

	if !strings.Contains(stdout.String(), "No entries found for previous month") {
		t.Errorf("Expected 'No entries found for previous month', got: %s", stdout.String())
	}
}

// Test that '--this-month' shows correct date range
func TestThisMonth_DateRangeOutput(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Calculate expected date range
	now := time.Now()
	start := timeutil.StartOfMonth(now)
	end := timeutil.EndOfMonth(now)

	// Run command with --this-month flag
	_ = rootCmd.Flags().Set("this-month", "true")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Verify output contains date range
	startDate := start.Format("Jan 2")
	endDate := end.Format("Jan 2, 2006")

	if !strings.Contains(output, startDate) {
		t.Errorf("Expected output to contain start date '%s', got: %s", startDate, output)
	}
	if !strings.Contains(output, endDate) {
		t.Errorf("Expected output to contain end date '%s', got: %s", endDate, output)
	}

	// Verify the header shows "this month"
	if !strings.Contains(output, "this month") {
		t.Errorf("Expected output to contain 'this month', got: %s", output)
	}

	// Verify start is the first day of the month
	if start.Day() != 1 {
		t.Errorf("Expected month start to be day 1, got day %d", start.Day())
	}

	// Verify end is the last day of the month
	// Check that adding one day moves us to the next month
	nextDay := end.AddDate(0, 0, 1)
	if nextDay.Month() == end.Month() {
		t.Errorf("Expected end to be last day of month, but adding one day stays in same month")
	}
}

// Test that '--prev-month' shows correct date range
func TestPrevMonth_DateRangeOutput(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Calculate expected date range
	lastMonth := time.Now().AddDate(0, -1, 0)
	start := timeutil.StartOfMonth(lastMonth)
	end := timeutil.EndOfMonth(lastMonth)

	// Run command with --prev-month flag
	_ = rootCmd.Flags().Set("prev-month", "true")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Verify output contains date range
	startDate := start.Format("Jan 2")
	// Handle year boundary cases where start and end may be in different years
	var endDate string
	if start.Year() == end.Year() {
		endDate = end.Format("Jan 2, 2006")
	} else {
		// Both dates should show full year if crossing year boundary
		startDate = start.Format("Jan 2, 2006")
		endDate = end.Format("Jan 2, 2006")
	}

	if !strings.Contains(output, startDate) {
		t.Errorf("Expected output to contain start date '%s', got: %s", startDate, output)
	}
	if !strings.Contains(output, endDate) {
		t.Errorf("Expected output to contain end date '%s', got: %s", endDate, output)
	}

	// Verify the header shows "previous month"
	if !strings.Contains(output, "previous month") {
		t.Errorf("Expected output to contain 'previous month', got: %s", output)
	}

	// Verify start is the first day of the month
	if start.Day() != 1 {
		t.Errorf("Expected month start to be day 1, got day %d", start.Day())
	}

	// Verify end is the last day of the month
	// Check that adding one day moves us to the next month
	nextDay := end.AddDate(0, 0, 1)
	if nextDay.Month() == end.Month() {
		t.Errorf("Expected end to be last day of month, but adding one day stays in same month")
	}
}

func TestYesterday_WithProjectFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for yesterday with different projects
	yesterday := time.Now().AddDate(0, 0, -1)
	entries := []entry.Entry{
		{
			Timestamp:       yesterday,
			Description:     "work on acme",
			DurationMinutes: 60,
			RawInput:        "work on acme @acme for 1h",
			Project:         "acme",
			Tags:            []string{},
		},
		{
			Timestamp:       yesterday,
			Description:     "work on client",
			DurationMinutes: 30,
			RawInput:        "work on client @client for 30m",
			Project:         "client",
			Tags:            []string{},
		},
		{
			Timestamp:       yesterday,
			Description:     "no project work",
			DurationMinutes: 45,
			RawInput:        "no project work for 45m",
			Project:         "",
			Tags:            []string{},
		},
	}

	// Write entries to storage
	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set project filter flag and yesterday flag
	_ = rootCmd.PersistentFlags().Set("project", "acme")
	_ = rootCmd.Flags().Set("yesterday", "true")

	// Run root command with flags
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show filtered results
	if !strings.Contains(output, "work on acme") {
		t.Errorf("Expected 'work on acme' in output, got: %s", output)
	}

	// Should NOT show other projects
	if strings.Contains(output, "work on client") {
		t.Errorf("Should not show 'work on client' (different project), got: %s", output)
	}

	if strings.Contains(output, "no project work") {
		t.Errorf("Should not show 'no project work' (no project), got: %s", output)
	}

	// Should show filter in period description
	if !strings.Contains(output, "yesterday (@acme)") {
		t.Errorf("Expected 'yesterday (@acme)' in output, got: %s", output)
	}

	// Total should reflect only filtered entries (1h)
	if !strings.Contains(output, "Total: 1h") {
		t.Errorf("Expected 'Total: 1h' (filtered), got: %s", output)
	}
}

func TestYesterday_WithShorthandProjectFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	entries := []entry.Entry{
		{
			Timestamp:       yesterday,
			Description:     "work on acme",
			DurationMinutes: 60,
			RawInput:        "work on acme @acme for 1h",
			Project:         "acme",
			Tags:            []string{},
		},
		{
			Timestamp:       yesterday,
			Description:     "work on other",
			DurationMinutes: 30,
			RawInput:        "work on other for 30m",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set yesterday flag and use shorthand @acme syntax
	_ = rootCmd.Flags().Set("yesterday", "true")
	rootCmd.Run(rootCmd, []string{"@acme"})

	output := stdout.String()

	// Should show only acme project
	if !strings.Contains(output, "work on acme") {
		t.Errorf("Expected 'work on acme' in output, got: %s", output)
	}

	if strings.Contains(output, "work on other") {
		t.Errorf("Should not show 'work on other', got: %s", output)
	}

	if !strings.Contains(output, "yesterday (@acme)") {
		t.Errorf("Expected 'yesterday (@acme)' in output, got: %s", output)
	}
}

func TestThisWeek_WithTagFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for this week with different tags
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "fix bug",
			DurationMinutes: 120,
			RawInput:        "fix bug #bugfix for 2h",
			Project:         "",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now,
			Description:     "code review",
			DurationMinutes: 30,
			RawInput:        "code review #review for 30m",
			Project:         "",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now,
			Description:     "meeting",
			DurationMinutes: 60,
			RawInput:        "meeting for 1h",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set tag filter flag and this-week flag
	_ = rootCmd.PersistentFlags().Set("tag", "bugfix")
	_ = rootCmd.Flags().Set("this-week", "true")

	// Run root command with flags
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show filtered results
	if !strings.Contains(output, "fix bug") {
		t.Errorf("Expected 'fix bug' in output, got: %s", output)
	}

	// Should NOT show other tags
	if strings.Contains(output, "code review") {
		t.Errorf("Should not show 'code review' (different tag), got: %s", output)
	}

	if strings.Contains(output, "meeting") {
		t.Errorf("Should not show 'meeting' (no tag), got: %s", output)
	}

	// Should show filter in period description (with date range)
	if !strings.Contains(output, "this week") || !strings.Contains(output, "(#bugfix)") {
		t.Errorf("Expected 'this week' with date range and '(#bugfix)' in output, got: %s", output)
	}

	// Total should reflect only filtered entries (2h)
	if !strings.Contains(output, "Total: 2h") {
		t.Errorf("Expected 'Total: 2h' (filtered), got: %s", output)
	}
}

func TestThisWeek_WithShorthandTagFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for this week
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "fix bug",
			DurationMinutes: 120,
			RawInput:        "fix bug #bugfix for 2h",
			Project:         "",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now,
			Description:     "other work",
			DurationMinutes: 30,
			RawInput:        "other work for 30m",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set this-week flag and use shorthand #bugfix syntax
	_ = rootCmd.Flags().Set("this-week", "true")
	rootCmd.Run(rootCmd, []string{"#bugfix"})

	output := stdout.String()

	// Should show only bugfix tag
	if !strings.Contains(output, "fix bug") {
		t.Errorf("Expected 'fix bug' in output, got: %s", output)
	}

	if strings.Contains(output, "other work") {
		t.Errorf("Should not show 'other work', got: %s", output)
	}

	// Check that the filter is applied (period description should mention #bugfix)
	if !strings.Contains(output, "#bugfix") || !strings.Contains(output, "this week") {
		t.Errorf("Expected period description to mention 'this week' and '#bugfix', got: %s", output)
	}
}

func TestPrevWeek_WithProjectAndTagFilters(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for last week
	lastWeekStart, _ := timeutil.LastWeek()
	entries := []entry.Entry{
		{
			Timestamp:       lastWeekStart.Add(24 * time.Hour),
			Description:     "urgent client work",
			DurationMinutes: 180,
			RawInput:        "urgent client work @client #urgent for 3h",
			Project:         "client",
			Tags:            []string{"urgent"},
		},
		{
			Timestamp:       lastWeekStart.Add(24 * time.Hour),
			Description:     "client review",
			DurationMinutes: 60,
			RawInput:        "client review @client #review for 1h",
			Project:         "client",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       lastWeekStart.Add(24 * time.Hour),
			Description:     "other urgent work",
			DurationMinutes: 90,
			RawInput:        "other urgent work @other #urgent for 1h30m",
			Project:         "other",
			Tags:            []string{"urgent"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set both project and tag filters and prev-week flag
	_ = rootCmd.PersistentFlags().Set("project", "client")
	_ = rootCmd.PersistentFlags().Set("tag", "urgent")
	_ = rootCmd.Flags().Set("prev-week", "true")

	// Run root command with flags
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show only entries matching BOTH filters
	if !strings.Contains(output, "urgent client work") {
		t.Errorf("Expected 'urgent client work' in output, got: %s", output)
	}

	// Should NOT show entries matching only one filter
	if strings.Contains(output, "client review") {
		t.Errorf("Should not show 'client review' (wrong tag), got: %s", output)
	}

	if strings.Contains(output, "other urgent work") {
		t.Errorf("Should not show 'other urgent work' (wrong project), got: %s", output)
	}

	// Should show both filters in period description
	if !strings.Contains(output, "@client") || !strings.Contains(output, "#urgent") || !strings.Contains(output, "previous week") {
		t.Errorf("Expected period description to mention 'previous week', '@client', and '#urgent', got: %s", output)
	}

	// Total should reflect only filtered entries (3h)
	if !strings.Contains(output, "Total: 3h") {
		t.Errorf("Expected 'Total: 3h' (filtered), got: %s", output)
	}
}

func TestPrevWeek_WithShorthandProjectAndTagFilters(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for last week
	lastWeekStart, _ := timeutil.LastWeek()
	entries := []entry.Entry{
		{
			Timestamp:       lastWeekStart.Add(24 * time.Hour),
			Description:     "urgent client work",
			DurationMinutes: 180,
			RawInput:        "urgent client work @client #urgent for 3h",
			Project:         "client",
			Tags:            []string{"urgent"},
		},
		{
			Timestamp:       lastWeekStart.Add(24 * time.Hour),
			Description:     "other work",
			DurationMinutes: 60,
			RawInput:        "other work for 1h",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set prev-week flag and use shorthand syntax for both project and tag
	_ = rootCmd.Flags().Set("prev-week", "true")
	rootCmd.Run(rootCmd, []string{"@client", "#urgent"})

	output := stdout.String()

	// Should show only entries matching both filters
	if !strings.Contains(output, "urgent client work") {
		t.Errorf("Expected 'urgent client work' in output, got: %s", output)
	}

	if strings.Contains(output, "other work") {
		t.Errorf("Should not show 'other work', got: %s", output)
	}

	// Check that filters are applied (period description should mention both filters)
	if !strings.Contains(output, "@client") || !strings.Contains(output, "#urgent") || !strings.Contains(output, "previous week") {
		t.Errorf("Expected period description to mention 'previous week', '@client', and '#urgent', got: %s", output)
	}
}

func TestYesterday_WithMultipleTagFilters(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	entries := []entry.Entry{
		{
			Timestamp:       yesterday,
			Description:     "urgent bugfix",
			DurationMinutes: 120,
			RawInput:        "urgent bugfix #bugfix #urgent for 2h",
			Project:         "",
			Tags:            []string{"bugfix", "urgent"},
		},
		{
			Timestamp:       yesterday,
			Description:     "regular bugfix",
			DurationMinutes: 60,
			RawInput:        "regular bugfix #bugfix for 1h",
			Project:         "",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       yesterday,
			Description:     "other urgent work",
			DurationMinutes: 30,
			RawInput:        "other urgent work #urgent for 30m",
			Project:         "",
			Tags:            []string{"urgent"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set yesterday flag and use shorthand syntax for multiple tags
	_ = rootCmd.Flags().Set("yesterday", "true")
	rootCmd.Run(rootCmd, []string{"#bugfix", "#urgent"})

	output := stdout.String()

	// Should show only entries with BOTH tags (AND logic)
	if !strings.Contains(output, "urgent bugfix") {
		t.Errorf("Expected 'urgent bugfix' in output, got: %s", output)
	}

	// Should NOT show entries with only one tag
	if strings.Contains(output, "regular bugfix") {
		t.Errorf("Should not show 'regular bugfix' (missing urgent tag), got: %s", output)
	}

	if strings.Contains(output, "other urgent work") {
		t.Errorf("Should not show 'other urgent work' (missing bugfix tag), got: %s", output)
	}

	// Check that filters are applied (period description should mention both tags)
	if !strings.Contains(output, "#bugfix") || !strings.Contains(output, "#urgent") || !strings.Contains(output, "yesterday") {
		t.Errorf("Expected period description to mention 'yesterday', '#bugfix', and '#urgent', got: %s", output)
	}
}

func TestValidate_Command(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a valid entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	validateCmd.Run(validateCmd, []string{})

	if !strings.Contains(stdout.String(), "Storage file is healthy") {
		t.Errorf("Expected 'Storage file is healthy', got: %s", stdout.String())
	}
}

func TestFormatProjectAndTags(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		tags     []string
		expected string
	}{
		{
			name:     "empty project and tags",
			project:  "",
			tags:     nil,
			expected: "",
		},
		{
			name:     "empty project and empty tags slice",
			project:  "",
			tags:     []string{},
			expected: "",
		},
		{
			name:     "project only",
			project:  "acme",
			tags:     nil,
			expected: "@acme",
		},
		{
			name:     "single tag only",
			project:  "",
			tags:     []string{"bugfix"},
			expected: "#bugfix",
		},
		{
			name:     "multiple tags only",
			project:  "",
			tags:     []string{"bugfix", "urgent"},
			expected: "#bugfix #urgent",
		},
		{
			name:     "project and single tag",
			project:  "acme",
			tags:     []string{"bugfix"},
			expected: "@acme #bugfix",
		},
		{
			name:     "project and multiple tags",
			project:  "acme",
			tags:     []string{"bugfix", "urgent", "frontend"},
			expected: "@acme #bugfix #urgent #frontend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatProjectAndTags(tt.project, tt.tags)
			if result != tt.expected {
				t.Errorf("formatProjectAndTags(%q, %v) = %q, expected %q", tt.project, tt.tags, result, tt.expected)
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
		expected    string
	}{
		{
			name:        "description only",
			description: "fix bug",
			project:     "",
			tags:        nil,
			expected:    "fix bug",
		},
		{
			name:        "description with project",
			description: "fix bug",
			project:     "acme",
			tags:        nil,
			expected:    "fix bug [@acme]",
		},
		{
			name:        "description with tags",
			description: "fix bug",
			project:     "",
			tags:        []string{"bugfix", "urgent"},
			expected:    "fix bug [#bugfix #urgent]",
		},
		{
			name:        "description with project and tags",
			description: "fix bug",
			project:     "acme",
			tags:        []string{"bugfix", "urgent"},
			expected:    "fix bug [@acme #bugfix #urgent]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEntryForLog(tt.description, tt.project, tt.tags)
			if result != tt.expected {
				t.Errorf("formatEntryForLog(%q, %q, %v) = %q, expected %q", tt.description, tt.project, tt.tags, result, tt.expected)
			}
		})
	}
}

func TestEditEntry_DescriptionWithProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry without project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original task",
		DurationMinutes: 60,
		RawInput:        "original task for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Edit with description containing @project and #tags
	_ = editCmd.Flags().Set("description", "fix bug @acme #bugfix #urgent")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	// Verify success message shows project/tags
	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}
	if !strings.Contains(output, "fix bug") {
		t.Errorf("Expected 'fix bug' in output, got: %s", output)
	}
	if !strings.Contains(output, "@acme") {
		t.Errorf("Expected '@acme' in output, got: %s", output)
	}
	if !strings.Contains(output, "#bugfix") {
		t.Errorf("Expected '#bugfix' in output, got: %s", output)
	}
	if !strings.Contains(output, "#urgent") {
		t.Errorf("Expected '#urgent' in output, got: %s", output)
	}

	// Verify entry was correctly updated
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "fix bug" {
		t.Errorf("Expected description 'fix bug', got: %s", entries[0].Description)
	}
	if entries[0].Project != "acme" {
		t.Errorf("Expected project 'acme', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 || entries[0].Tags[0] != "bugfix" || entries[0].Tags[1] != "urgent" {
		t.Errorf("Expected tags ['bugfix', 'urgent'], got: %v", entries[0].Tags)
	}
}

func TestEditEntry_DurationPreservesProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITH project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "fix bug",
		DurationMinutes: 60,
		RawInput:        "fix bug @acme #bugfix for 1h",
		Project:         "acme",
		Tags:            []string{"bugfix", "urgent"},
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Edit only duration
	_ = editCmd.Flags().Set("duration", "2h")
	defer func() { _ = editCmd.Flags().Set("duration", "") }()

	editEntry(editCmd, []string{"1"})

	// Verify success message shows project/tags (preserved)
	output := stdout.String()
	if !strings.Contains(output, "@acme") {
		t.Errorf("Expected '@acme' in output (preserved), got: %s", output)
	}
	if !strings.Contains(output, "#bugfix") {
		t.Errorf("Expected '#bugfix' in output (preserved), got: %s", output)
	}
	if !strings.Contains(output, "2h") {
		t.Errorf("Expected '2h' in output, got: %s", output)
	}

	// Verify project/tags are preserved in the entry
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Project != "acme" {
		t.Errorf("Expected project 'acme' preserved, got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 {
		t.Errorf("Expected 2 tags preserved, got: %v", entries[0].Tags)
	}
	if entries[0].DurationMinutes != 120 {
		t.Errorf("Expected duration 120, got: %d", entries[0].DurationMinutes)
	}
}

func TestEditEntry_RemoveProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITH project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "fix bug",
		DurationMinutes: 60,
		RawInput:        "fix bug @acme #bugfix for 1h",
		Project:         "acme",
		Tags:            []string{"bugfix"},
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, _, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Edit with description that has NO project/tags
	_ = editCmd.Flags().Set("description", "plain description")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	// Verify project/tags were removed
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "plain description" {
		t.Errorf("Expected description 'plain description', got: %s", entries[0].Description)
	}
	if entries[0].Project != "" {
		t.Errorf("Expected empty project, got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 0 {
		t.Errorf("Expected empty tags, got: %v", entries[0].Tags)
	}
}

func TestEditEntry_AddProjectAndTagsToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITHOUT project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "plain task",
		DurationMinutes: 60,
		RawInput:        "plain task for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Add project and tags via edit
	_ = editCmd.Flags().Set("description", "updated task @newproject #newtag")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	output := stdout.String()
	if !strings.Contains(output, "@newproject") {
		t.Errorf("Expected '@newproject' in output, got: %s", output)
	}
	if !strings.Contains(output, "#newtag") {
		t.Errorf("Expected '#newtag' in output, got: %s", output)
	}

	// Verify entry was updated correctly
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Project != "newproject" {
		t.Errorf("Expected project 'newproject', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 1 || entries[0].Tags[0] != "newtag" {
		t.Errorf("Expected tags ['newtag'], got: %v", entries[0].Tags)
	}
}

func TestEditEntry_EmptyDescriptionWithOnlyProjectTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original",
		DurationMinutes: 60,
		RawInput:        "original for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	// Try to edit with only @project/#tags (no actual description)
	_ = editCmd.Flags().Set("description", "@acme #bugfix")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if !exitCalled {
		t.Error("Expected exit to be called for empty description")
	}
	if !strings.Contains(stderr.String(), "Description cannot be empty") {
		t.Errorf("Expected empty description error, got: %s", stderr.String())
	}
}

// Integration tests for entry creation with project and tags

func TestCreateEntry_WithProject(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"fix", "bug", "@acme", "for", "2h"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Logged:") {
		t.Errorf("Expected 'Logged:' in output, got: %s", output)
	}
	if !strings.Contains(output, "fix bug") {
		t.Errorf("Expected 'fix bug' in output, got: %s", output)
	}
	if !strings.Contains(output, "@acme") {
		t.Errorf("Expected '@acme' in output, got: %s", output)
	}

	// Verify entry was written with correct project
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Description != "fix bug" {
		t.Errorf("Expected description 'fix bug', got: %s", entries[0].Description)
	}
	if entries[0].Project != "acme" {
		t.Errorf("Expected project 'acme', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 0 {
		t.Errorf("Expected no tags, got: %v", entries[0].Tags)
	}
	if entries[0].DurationMinutes != 120 {
		t.Errorf("Expected duration 120 minutes, got: %d", entries[0].DurationMinutes)
	}
}

func TestCreateEntry_WithTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"fix", "bug", "#bugfix", "#urgent", "for", "1h30m"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Logged:") {
		t.Errorf("Expected 'Logged:' in output, got: %s", output)
	}
	if !strings.Contains(output, "fix bug") {
		t.Errorf("Expected 'fix bug' in output, got: %s", output)
	}
	if !strings.Contains(output, "#bugfix") {
		t.Errorf("Expected '#bugfix' in output, got: %s", output)
	}
	if !strings.Contains(output, "#urgent") {
		t.Errorf("Expected '#urgent' in output, got: %s", output)
	}

	// Verify entry was written with correct tags
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Description != "fix bug" {
		t.Errorf("Expected description 'fix bug', got: %s", entries[0].Description)
	}
	if entries[0].Project != "" {
		t.Errorf("Expected empty project, got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got: %v", entries[0].Tags)
	} else {
		if entries[0].Tags[0] != "bugfix" || entries[0].Tags[1] != "urgent" {
			t.Errorf("Expected tags ['bugfix', 'urgent'], got: %v", entries[0].Tags)
		}
	}
	if entries[0].DurationMinutes != 90 {
		t.Errorf("Expected duration 90 minutes, got: %d", entries[0].DurationMinutes)
	}
}

func TestCreateEntry_WithProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"implement", "feature", "@clientco", "#feature", "#priority", "for", "3h"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Logged:") {
		t.Errorf("Expected 'Logged:' in output, got: %s", output)
	}
	if !strings.Contains(output, "implement feature") {
		t.Errorf("Expected 'implement feature' in output, got: %s", output)
	}
	if !strings.Contains(output, "@clientco") {
		t.Errorf("Expected '@clientco' in output, got: %s", output)
	}
	if !strings.Contains(output, "#feature") {
		t.Errorf("Expected '#feature' in output, got: %s", output)
	}
	if !strings.Contains(output, "#priority") {
		t.Errorf("Expected '#priority' in output, got: %s", output)
	}

	// Verify entry was written with correct project and tags
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Description != "implement feature" {
		t.Errorf("Expected description 'implement feature', got: %s", entries[0].Description)
	}
	if entries[0].Project != "clientco" {
		t.Errorf("Expected project 'clientco', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got: %v", entries[0].Tags)
	} else {
		if entries[0].Tags[0] != "feature" || entries[0].Tags[1] != "priority" {
			t.Errorf("Expected tags ['feature', 'priority'], got: %v", entries[0].Tags)
		}
	}
	if entries[0].DurationMinutes != 180 {
		t.Errorf("Expected duration 180 minutes, got: %d", entries[0].DurationMinutes)
	}
}

func TestCreateEntry_WithoutProjectOrTags_BackwardCompat(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"simple", "task", "for", "45m"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Logged:") {
		t.Errorf("Expected 'Logged:' in output, got: %s", output)
	}
	if !strings.Contains(output, "simple task") {
		t.Errorf("Expected 'simple task' in output, got: %s", output)
	}
	// Output should NOT contain project/tag brackets when neither is present
	if strings.Contains(output, "[@") || strings.Contains(output, "[#") {
		t.Errorf("Did not expect project/tag brackets in output, got: %s", output)
	}

	// Verify entry was written without project or tags
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Description != "simple task" {
		t.Errorf("Expected description 'simple task', got: %s", entries[0].Description)
	}
	if entries[0].Project != "" {
		t.Errorf("Expected empty project, got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 0 {
		t.Errorf("Expected no tags, got: %v", entries[0].Tags)
	}
	if entries[0].DurationMinutes != 45 {
		t.Errorf("Expected duration 45 minutes, got: %d", entries[0].DurationMinutes)
	}
}

func TestCreateEntry_VerifyJSONLStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, _, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"review", "code", "@acme", "#code-review", "for", "1h"})

	// Read raw JSONL file to verify correct JSON encoding
	content, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}

	jsonStr := string(content)
	// Verify project field is present in JSON
	if !strings.Contains(jsonStr, `"project":"acme"`) {
		t.Errorf("Expected JSON to contain project field, got: %s", jsonStr)
	}
	// Verify tags field is present in JSON
	if !strings.Contains(jsonStr, `"tags":["code-review"]`) {
		t.Errorf("Expected JSON to contain tags field, got: %s", jsonStr)
	}
	// Verify description is clean (without @project and #tags)
	if !strings.Contains(jsonStr, `"description":"review code"`) {
		t.Errorf("Expected JSON to contain clean description, got: %s", jsonStr)
	}
}

func TestCreateEntry_OnlyProjectTagsError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	// Create entry with only @project/#tags (no actual description)
	createEntry([]string{"@acme", "#bugfix", "for", "1h"})

	if !exitCalled {
		t.Error("Expected exit to be called for empty description")
	}
	if !strings.Contains(stderr.String(), "Description cannot be empty") {
		t.Errorf("Expected empty description error, got: %s", stderr.String())
	}
}

func TestCreateEntry_ProjectAndTagsInMiddle(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Test with @project and #tags in the middle of description
	createEntry([]string{"fix", "@acme", "bug", "#bugfix", "in", "login", "for", "2h"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Logged:") {
		t.Errorf("Expected 'Logged:' in output, got: %s", output)
	}

	// Verify entry was written correctly
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	// Description should be cleaned of @project and #tags
	if entries[0].Description != "fix bug in login" {
		t.Errorf("Expected description 'fix bug in login', got: %s", entries[0].Description)
	}
	if entries[0].Project != "acme" {
		t.Errorf("Expected project 'acme', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 1 || entries[0].Tags[0] != "bugfix" {
		t.Errorf("Expected tags ['bugfix'], got: %v", entries[0].Tags)
	}
}

func TestCreateEntry_MultipleTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	createEntry([]string{"deploy", "app", "#deploy", "#production", "#release-v1", "for", "30m"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "#deploy") {
		t.Errorf("Expected '#deploy' in output, got: %s", output)
	}
	if !strings.Contains(output, "#production") {
		t.Errorf("Expected '#production' in output, got: %s", output)
	}
	if !strings.Contains(output, "#release-v1") {
		t.Errorf("Expected '#release-v1' in output, got: %s", output)
	}

	// Verify all tags were stored
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if len(entries[0].Tags) != 3 {
		t.Errorf("Expected 3 tags, got: %v", entries[0].Tags)
	} else {
		if entries[0].Tags[0] != "deploy" || entries[0].Tags[1] != "production" || entries[0].Tags[2] != "release-v1" {
			t.Errorf("Expected tags ['deploy', 'production', 'release-v1'], got: %v", entries[0].Tags)
		}
	}
}

// Integration tests for listing entries with project and tags

func TestListEntries_WithProject(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry with project
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "fix bug",
		DurationMinutes: 60,
		RawInput:        "fix bug @acme for 1h",
		Project:         "acme",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset filter flags to avoid contamination from other tests
	resetFilterFlags(rootCmd)

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	})

	output := stdout.String()
	// Verify output shows @project
	if !strings.Contains(output, "@acme") {
		t.Errorf("Expected '@acme' in list output, got: %s", output)
	}
	if !strings.Contains(output, "fix bug") {
		t.Errorf("Expected 'fix bug' in list output, got: %s", output)
	}
	// Verify format is correct (description [@project])
	if !strings.Contains(output, "fix bug [@acme]") {
		t.Errorf("Expected 'fix bug [@acme]' format in list output, got: %s", output)
	}
}

func TestListEntries_WithTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry with tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "implement feature",
		DurationMinutes: 120,
		RawInput:        "implement feature #feature #urgent for 2h",
		Tags:            []string{"feature", "urgent"},
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset filter flags to avoid contamination from other tests
	resetFilterFlags(rootCmd)

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	})

	output := stdout.String()
	// Verify output shows #tags
	if !strings.Contains(output, "#feature") {
		t.Errorf("Expected '#feature' in list output, got: %s", output)
	}
	if !strings.Contains(output, "#urgent") {
		t.Errorf("Expected '#urgent' in list output, got: %s", output)
	}
	if !strings.Contains(output, "implement feature") {
		t.Errorf("Expected 'implement feature' in list output, got: %s", output)
	}
	// Verify format is correct (description [#tag1 #tag2])
	if !strings.Contains(output, "implement feature [#feature #urgent]") {
		t.Errorf("Expected 'implement feature [#feature #urgent]' format in list output, got: %s", output)
	}
}

func TestListEntries_WithProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry with both project and tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "deploy app",
		DurationMinutes: 90,
		RawInput:        "deploy app @clientco #deploy #production for 1h30m",
		Project:         "clientco",
		Tags:            []string{"deploy", "production"},
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset filter flags to avoid contamination from other tests
	resetFilterFlags(rootCmd)

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	})

	output := stdout.String()
	// Verify output shows @project and #tags
	if !strings.Contains(output, "@clientco") {
		t.Errorf("Expected '@clientco' in list output, got: %s", output)
	}
	if !strings.Contains(output, "#deploy") {
		t.Errorf("Expected '#deploy' in list output, got: %s", output)
	}
	if !strings.Contains(output, "#production") {
		t.Errorf("Expected '#production' in list output, got: %s", output)
	}
	if !strings.Contains(output, "deploy app") {
		t.Errorf("Expected 'deploy app' in list output, got: %s", output)
	}
	// Verify format is correct (description [@project #tag1 #tag2])
	if !strings.Contains(output, "deploy app [@clientco #deploy #production]") {
		t.Errorf("Expected 'deploy app [@clientco #deploy #production]' format in list output, got: %s", output)
	}
}

func TestListEntries_BackwardCompatibility_NoProjectOrTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITHOUT project or tags (simulating old entries)
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "plain task",
		DurationMinutes: 45,
		RawInput:        "plain task for 45m",
		// No Project or Tags fields set (empty values)
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset filter flags to avoid contamination from other tests
	resetFilterFlags(rootCmd)

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	})

	output := stdout.String()
	// Verify description is shown without brackets
	if !strings.Contains(output, "plain task") {
		t.Errorf("Expected 'plain task' in list output, got: %s", output)
	}
	// Verify no project/tag brackets are present
	if strings.Contains(output, "[@") || strings.Contains(output, "[#") {
		t.Errorf("Did not expect project/tag brackets in list output for plain entries, got: %s", output)
	}
	// Verify the output shows proper formatting (should just be "plain task" without brackets)
	if strings.Contains(output, "plain task [") {
		t.Errorf("Did not expect 'plain task [' in list output, got: %s", output)
	}
}

func TestListEntries_MixedEntriesWithAndWithoutProjectTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	// Use fixed times within today to avoid edge cases when test runs early in the day
	todayBase := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	// Create multiple entries with different combinations
	testEntries := []entry.Entry{
		{
			Timestamp:       todayBase,
			Description:     "plain entry",
			DurationMinutes: 30,
			RawInput:        "plain entry for 30m",
		},
		{
			Timestamp:       todayBase.Add(1 * time.Hour),
			Description:     "with project",
			DurationMinutes: 60,
			RawInput:        "with project @acme for 1h",
			Project:         "acme",
		},
		{
			Timestamp:       todayBase.Add(2 * time.Hour),
			Description:     "with tags",
			DurationMinutes: 45,
			RawInput:        "with tags #tag1 #tag2 for 45m",
			Tags:            []string{"tag1", "tag2"},
		},
		{
			Timestamp:       todayBase.Add(3 * time.Hour),
			Description:     "with both",
			DurationMinutes: 90,
			RawInput:        "with both @proj #mytag for 1h30m",
			Project:         "proj",
			Tags:            []string{"mytag"},
		},
	}

	for _, e := range testEntries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset filter flags to avoid contamination from other tests
	resetFilterFlags(rootCmd)

	listEntries(rootCmd, "today", func() (time.Time, time.Time) {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	})

	output := stdout.String()

	// Verify plain entry shows without brackets
	if !strings.Contains(output, "plain entry") {
		t.Errorf("Expected 'plain entry' in output, got: %s", output)
	}
	// Plain entry should NOT have brackets following it in the format "plain entry ["
	// But it's hard to check this precisely, so we'll just verify others have correct format

	// Verify entry with project shows @project
	if !strings.Contains(output, "with project [@acme]") {
		t.Errorf("Expected 'with project [@acme]' in output, got: %s", output)
	}

	// Verify entry with tags shows #tags
	if !strings.Contains(output, "with tags [#tag1 #tag2]") {
		t.Errorf("Expected 'with tags [#tag1 #tag2]' in output, got: %s", output)
	}

	// Verify entry with both shows both
	if !strings.Contains(output, "with both [@proj #mytag]") {
		t.Errorf("Expected 'with both [@proj #mytag]' in output, got: %s", output)
	}

	// Verify total is shown
	if !strings.Contains(output, "Total:") {
		t.Errorf("Expected 'Total:' in output, got: %s", output)
	}
}

// Integration tests for editing entries with project and tags

func TestEditEntry_Integration_DescriptionReParsesProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry without project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original task",
		DurationMinutes: 60,
		RawInput:        "original task for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Edit with description containing @project and #tags
	_ = editCmd.Flags().Set("description", "fix bug @acme #bugfix #urgent")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	// Verify success message shows project/tags
	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}
	if !strings.Contains(output, "fix bug") {
		t.Errorf("Expected 'fix bug' in output, got: %s", output)
	}
	if !strings.Contains(output, "@acme") {
		t.Errorf("Expected '@acme' in output, got: %s", output)
	}
	if !strings.Contains(output, "#bugfix") {
		t.Errorf("Expected '#bugfix' in output, got: %s", output)
	}
	if !strings.Contains(output, "#urgent") {
		t.Errorf("Expected '#urgent' in output, got: %s", output)
	}

	// Verify JSONL storage contains correct fields
	content, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}
	jsonStr := string(content)
	if !strings.Contains(jsonStr, `"description":"fix bug"`) {
		t.Errorf("Expected JSON to contain clean description, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"project":"acme"`) {
		t.Errorf("Expected JSON to contain project field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"tags":["bugfix","urgent"]`) {
		t.Errorf("Expected JSON to contain tags field, got: %s", jsonStr)
	}

	// Verify entry fields via storage API
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "fix bug" {
		t.Errorf("Expected description 'fix bug', got: %s", entries[0].Description)
	}
	if entries[0].Project != "acme" {
		t.Errorf("Expected project 'acme', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 || entries[0].Tags[0] != "bugfix" || entries[0].Tags[1] != "urgent" {
		t.Errorf("Expected tags ['bugfix', 'urgent'], got: %v", entries[0].Tags)
	}
}

func TestEditEntry_Integration_DurationPreservesProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITH project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "fix bug",
		DurationMinutes: 60,
		RawInput:        "fix bug @acme #bugfix #urgent for 1h",
		Project:         "acme",
		Tags:            []string{"bugfix", "urgent"},
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Edit only duration
	_ = editCmd.Flags().Set("duration", "3h")
	defer func() { _ = editCmd.Flags().Set("duration", "") }()

	editEntry(editCmd, []string{"1"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	// Verify success message shows preserved project/tags
	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}
	if !strings.Contains(output, "@acme") {
		t.Errorf("Expected '@acme' in output (preserved), got: %s", output)
	}
	if !strings.Contains(output, "#bugfix") {
		t.Errorf("Expected '#bugfix' in output (preserved), got: %s", output)
	}
	if !strings.Contains(output, "#urgent") {
		t.Errorf("Expected '#urgent' in output (preserved), got: %s", output)
	}
	if !strings.Contains(output, "3h") {
		t.Errorf("Expected '3h' in output, got: %s", output)
	}

	// Verify JSONL storage preserves project/tags
	content, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}
	jsonStr := string(content)
	if !strings.Contains(jsonStr, `"project":"acme"`) {
		t.Errorf("Expected JSON to preserve project field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"tags":["bugfix","urgent"]`) {
		t.Errorf("Expected JSON to preserve tags field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"duration_minutes":180`) {
		t.Errorf("Expected JSON to have updated duration, got: %s", jsonStr)
	}

	// Verify entry fields via storage API
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Project != "acme" {
		t.Errorf("Expected project 'acme' preserved, got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 || entries[0].Tags[0] != "bugfix" || entries[0].Tags[1] != "urgent" {
		t.Errorf("Expected tags ['bugfix', 'urgent'] preserved, got: %v", entries[0].Tags)
	}
	if entries[0].DurationMinutes != 180 {
		t.Errorf("Expected duration 180, got: %d", entries[0].DurationMinutes)
	}
}

func TestEditEntry_Integration_RemoveProjectAndTagsViaEdit(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITH project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "fix bug",
		DurationMinutes: 60,
		RawInput:        "fix bug @acme #bugfix for 1h",
		Project:         "acme",
		Tags:            []string{"bugfix"},
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Edit with description that has NO project/tags
	_ = editCmd.Flags().Set("description", "plain description")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	// Verify success message shows plain description without brackets
	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}
	if !strings.Contains(output, "plain description") {
		t.Errorf("Expected 'plain description' in output, got: %s", output)
	}
	// Should NOT have project/tag brackets
	if strings.Contains(output, "[@") || strings.Contains(output, "[#") {
		t.Errorf("Did not expect project/tag brackets after removal, got: %s", output)
	}

	// Verify JSONL storage has empty/missing project and tags (omitempty)
	content, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}
	jsonStr := string(content)
	if strings.Contains(jsonStr, `"project":`) {
		t.Errorf("Expected JSON to NOT contain project field (omitempty), got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"tags":`) {
		t.Errorf("Expected JSON to NOT contain tags field (omitempty), got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"description":"plain description"`) {
		t.Errorf("Expected JSON to contain new description, got: %s", jsonStr)
	}

	// Verify entry fields via storage API
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "plain description" {
		t.Errorf("Expected description 'plain description', got: %s", entries[0].Description)
	}
	if entries[0].Project != "" {
		t.Errorf("Expected empty project after removal, got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 0 {
		t.Errorf("Expected empty tags after removal, got: %v", entries[0].Tags)
	}
}

func TestEditEntry_Integration_AddProjectAndTagsViaEdit(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITHOUT project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "plain task",
		DurationMinutes: 45,
		RawInput:        "plain task for 45m",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Add project and tags via edit
	_ = editCmd.Flags().Set("description", "updated task @newproject #newtag #othertag")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	// Verify success message shows new project/tags
	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}
	if !strings.Contains(output, "@newproject") {
		t.Errorf("Expected '@newproject' in output, got: %s", output)
	}
	if !strings.Contains(output, "#newtag") {
		t.Errorf("Expected '#newtag' in output, got: %s", output)
	}
	if !strings.Contains(output, "#othertag") {
		t.Errorf("Expected '#othertag' in output, got: %s", output)
	}

	// Verify JSONL storage contains new project and tags
	content, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}
	jsonStr := string(content)
	if !strings.Contains(jsonStr, `"project":"newproject"`) {
		t.Errorf("Expected JSON to contain new project field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"tags":["newtag","othertag"]`) {
		t.Errorf("Expected JSON to contain new tags field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"description":"updated task"`) {
		t.Errorf("Expected JSON to contain clean description, got: %s", jsonStr)
	}

	// Verify entry fields via storage API
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "updated task" {
		t.Errorf("Expected description 'updated task', got: %s", entries[0].Description)
	}
	if entries[0].Project != "newproject" {
		t.Errorf("Expected project 'newproject', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 || entries[0].Tags[0] != "newtag" || entries[0].Tags[1] != "othertag" {
		t.Errorf("Expected tags ['newtag', 'othertag'], got: %v", entries[0].Tags)
	}
}

func TestEditEntry_Integration_ChangeProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITH project/tags
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "fix bug",
		DurationMinutes: 60,
		RawInput:        "fix bug @oldproject #oldtag for 1h",
		Project:         "oldproject",
		Tags:            []string{"oldtag"},
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Change to different project and tags
	_ = editCmd.Flags().Set("description", "implement feature @newclient #feature #priority")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	// Verify success message shows new project/tags
	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}
	if !strings.Contains(output, "@newclient") {
		t.Errorf("Expected '@newclient' in output, got: %s", output)
	}
	if !strings.Contains(output, "#feature") {
		t.Errorf("Expected '#feature' in output, got: %s", output)
	}
	if !strings.Contains(output, "#priority") {
		t.Errorf("Expected '#priority' in output, got: %s", output)
	}
	// Old values should not appear
	if strings.Contains(output, "@oldproject") {
		t.Errorf("Did not expect '@oldproject' in output after change, got: %s", output)
	}
	if strings.Contains(output, "#oldtag") {
		t.Errorf("Did not expect '#oldtag' in output after change, got: %s", output)
	}

	// Verify JSONL storage contains new values
	content, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}
	jsonStr := string(content)
	if !strings.Contains(jsonStr, `"project":"newclient"`) {
		t.Errorf("Expected JSON to contain new project, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"tags":["feature","priority"]`) {
		t.Errorf("Expected JSON to contain new tags, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"description":"implement feature"`) {
		t.Errorf("Expected JSON to contain new description, got: %s", jsonStr)
	}
	// Old values should not appear
	if strings.Contains(jsonStr, `"oldproject"`) {
		t.Errorf("Did not expect old project in JSON after change, got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"oldtag"`) {
		t.Errorf("Did not expect old tag in JSON after change, got: %s", jsonStr)
	}

	// Verify entry fields via storage API
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "implement feature" {
		t.Errorf("Expected description 'implement feature', got: %s", entries[0].Description)
	}
	if entries[0].Project != "newclient" {
		t.Errorf("Expected project 'newclient', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 || entries[0].Tags[0] != "feature" || entries[0].Tags[1] != "priority" {
		t.Errorf("Expected tags ['feature', 'priority'], got: %v", entries[0].Tags)
	}
}

func TestEditEntry_Integration_BackwardCompatibility_NoProjectOrTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry WITHOUT project/tags (simulating old entry)
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "old entry",
		DurationMinutes: 30,
		RawInput:        "old entry for 30m",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Edit description without adding project/tags
	_ = editCmd.Flags().Set("description", "updated entry")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	editEntry(editCmd, []string{"1"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	// Verify success message shows plain description without brackets
	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}
	if !strings.Contains(output, "updated entry") {
		t.Errorf("Expected 'updated entry' in output, got: %s", output)
	}
	// Should NOT have project/tag brackets
	if strings.Contains(output, "[@") || strings.Contains(output, "[#") {
		t.Errorf("Did not expect project/tag brackets for plain entry, got: %s", output)
	}

	// Verify JSONL storage does not have project/tags (omitempty)
	content, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}
	jsonStr := string(content)
	if strings.Contains(jsonStr, `"project":`) {
		t.Errorf("Expected JSON to NOT contain project field (omitempty), got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"tags":`) {
		t.Errorf("Expected JSON to NOT contain tags field (omitempty), got: %s", jsonStr)
	}

	// Verify entry fields via storage API
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "updated entry" {
		t.Errorf("Expected description 'updated entry', got: %s", entries[0].Description)
	}
	if entries[0].Project != "" {
		t.Errorf("Expected empty project, got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 0 {
		t.Errorf("Expected empty tags, got: %v", entries[0].Tags)
	}
}

func TestEditEntry_Integration_EditBothDescriptionAndDurationWithProjectTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original",
		DurationMinutes: 60,
		RawInput:        "original for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	d, stdout, stderr := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Edit both description and duration with project/tags
	_ = editCmd.Flags().Set("description", "new task @project #tag1 #tag2")
	_ = editCmd.Flags().Set("duration", "2h30m")
	defer func() {
		_ = editCmd.Flags().Set("description", "")
		_ = editCmd.Flags().Set("duration", "")
	}()

	editEntry(editCmd, []string{"1"})

	if stderr.Len() > 0 {
		t.Errorf("Unexpected stderr output: %s", stderr.String())
	}

	// Verify success message
	output := stdout.String()
	if !strings.Contains(output, "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", output)
	}
	if !strings.Contains(output, "new task") {
		t.Errorf("Expected 'new task' in output, got: %s", output)
	}
	if !strings.Contains(output, "@project") {
		t.Errorf("Expected '@project' in output, got: %s", output)
	}
	if !strings.Contains(output, "#tag1") {
		t.Errorf("Expected '#tag1' in output, got: %s", output)
	}
	if !strings.Contains(output, "#tag2") {
		t.Errorf("Expected '#tag2' in output, got: %s", output)
	}
	if !strings.Contains(output, "2h 30m") {
		t.Errorf("Expected '2h 30m' in output, got: %s", output)
	}

	// Verify JSONL storage
	content, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}
	jsonStr := string(content)
	if !strings.Contains(jsonStr, `"description":"new task"`) {
		t.Errorf("Expected JSON to contain new description, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"project":"project"`) {
		t.Errorf("Expected JSON to contain project field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"tags":["tag1","tag2"]`) {
		t.Errorf("Expected JSON to contain tags field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"duration_minutes":150`) {
		t.Errorf("Expected JSON to have updated duration (150 minutes), got: %s", jsonStr)
	}

	// Verify entry fields via storage API
	entries, _ := storage.ReadEntries(storagePath)
	if entries[0].Description != "new task" {
		t.Errorf("Expected description 'new task', got: %s", entries[0].Description)
	}
	if entries[0].Project != "project" {
		t.Errorf("Expected project 'project', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 || entries[0].Tags[0] != "tag1" || entries[0].Tags[1] != "tag2" {
		t.Errorf("Expected tags ['tag1', 'tag2'], got: %v", entries[0].Tags)
	}
	if entries[0].DurationMinutes != 150 {
		t.Errorf("Expected duration 150 minutes, got: %d", entries[0].DurationMinutes)
	}
}

func TestParseShorthandFilters(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedProject   string
		expectedTags      []string
		expectedRemaining []string
	}{
		{
			name:              "empty args",
			args:              []string{},
			expectedProject:   "",
			expectedTags:      []string{},
			expectedRemaining: []string{},
		},
		{
			name:              "single @project",
			args:              []string{"@acme"},
			expectedProject:   "acme",
			expectedTags:      []string{},
			expectedRemaining: []string{},
		},
		{
			name:              "single #tag",
			args:              []string{"#bugfix"},
			expectedProject:   "",
			expectedTags:      []string{"bugfix"},
			expectedRemaining: []string{},
		},
		{
			name:              "multiple #tags",
			args:              []string{"#bugfix", "#urgent"},
			expectedProject:   "",
			expectedTags:      []string{"bugfix", "urgent"},
			expectedRemaining: []string{},
		},
		{
			name:              "@project and #tag",
			args:              []string{"@acme", "#bugfix"},
			expectedProject:   "acme",
			expectedTags:      []string{"bugfix"},
			expectedRemaining: []string{},
		},
		{
			name:              "@project, multiple #tags",
			args:              []string{"@client", "#urgent", "#backend"},
			expectedProject:   "client",
			expectedTags:      []string{"urgent", "backend"},
			expectedRemaining: []string{},
		},
		{
			name:              "shorthand with non-shorthand args",
			args:              []string{"@acme", "y"},
			expectedProject:   "acme",
			expectedTags:      []string{},
			expectedRemaining: []string{"y"},
		},
		{
			name:              "mixed order",
			args:              []string{"y", "@client", "#urgent"},
			expectedProject:   "client",
			expectedTags:      []string{"urgent"},
			expectedRemaining: []string{"y"},
		},
		{
			name:              "non-shorthand only",
			args:              []string{"y", "w"},
			expectedProject:   "",
			expectedTags:      []string{},
			expectedRemaining: []string{"y", "w"},
		},
		{
			name:              "empty @ prefix",
			args:              []string{"@"},
			expectedProject:   "",
			expectedTags:      []string{},
			expectedRemaining: []string{},
		},
		{
			name:              "empty # prefix",
			args:              []string{"#"},
			expectedProject:   "",
			expectedTags:      []string{},
			expectedRemaining: []string{},
		},
		{
			name:              "multiple @projects (last wins)",
			args:              []string{"@project1", "@project2"},
			expectedProject:   "project2",
			expectedTags:      []string{},
			expectedRemaining: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command with persistent flags
			cmd := &cobra.Command{}
			cmd.PersistentFlags().String("project", "", "Filter entries by project")
			cmd.PersistentFlags().StringSlice("tag", []string{}, "Filter entries by tag")

			remaining := parseShorthandFilters(cmd, tt.args)

			// Check project flag
			project, _ := cmd.PersistentFlags().GetString("project")
			if project != tt.expectedProject {
				t.Errorf("Expected project %q, got %q", tt.expectedProject, project)
			}

			// Check tag flags
			tags, _ := cmd.PersistentFlags().GetStringSlice("tag")
			if len(tags) != len(tt.expectedTags) {
				t.Errorf("Expected %d tags, got %d", len(tt.expectedTags), len(tags))
			} else {
				for i, expectedTag := range tt.expectedTags {
					if tags[i] != expectedTag {
						t.Errorf("Expected tag[%d] %q, got %q", i, expectedTag, tags[i])
					}
				}
			}

			// Check remaining args
			if len(remaining) != len(tt.expectedRemaining) {
				t.Errorf("Expected %d remaining args, got %d", len(tt.expectedRemaining), len(remaining))
			} else {
				for i, expectedArg := range tt.expectedRemaining {
					if remaining[i] != expectedArg {
						t.Errorf("Expected remaining[%d] %q, got %q", i, expectedArg, remaining[i])
					}
				}
			}
		})
	}
}

func TestParseShorthandFilters_PreservesExistingFlags(t *testing.T) {
	// Test that shorthand syntax combines with existing --tag flags
	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("project", "", "Filter entries by project")
	cmd.PersistentFlags().StringSlice("tag", []string{}, "Filter entries by tag")

	// Set existing --tag flag value
	_ = cmd.PersistentFlags().Set("tag", "existing")

	// Parse shorthand with additional tag
	remaining := parseShorthandFilters(cmd, []string{"#new", "y"})

	// Check that both tags are present
	tags, _ := cmd.PersistentFlags().GetStringSlice("tag")
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags (existing + new), got %d: %v", len(tags), tags)
	}

	expectedTags := map[string]bool{"existing": true, "new": true}
	for _, tag := range tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %q", tag)
		}
	}

	// Check remaining args
	if len(remaining) != 1 || remaining[0] != "y" {
		t.Errorf("Expected remaining args ['y'], got %v", remaining)
	}
}

func TestRootCommand_WithProjectFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for today with different projects
	today := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "work on acme",
			DurationMinutes: 60,
			RawInput:        "work on acme @acme for 1h",
			Project:         "acme",
			Tags:            []string{},
		},
		{
			Timestamp:       today,
			Description:     "work on client",
			DurationMinutes: 30,
			RawInput:        "work on client @client for 30m",
			Project:         "client",
			Tags:            []string{},
		},
		{
			Timestamp:       today,
			Description:     "no project work",
			DurationMinutes: 45,
			RawInput:        "no project work for 45m",
			Project:         "",
			Tags:            []string{},
		},
	}

	// Write entries to storage
	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set project filter flag
	_ = rootCmd.PersistentFlags().Set("project", "acme")

	// Run root command with filter
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show filtered results
	if !strings.Contains(output, "work on acme") {
		t.Errorf("Expected 'work on acme' in output, got: %s", output)
	}

	// Should NOT show other projects
	if strings.Contains(output, "work on client") {
		t.Errorf("Should not show 'work on client' (different project), got: %s", output)
	}

	if strings.Contains(output, "no project work") {
		t.Errorf("Should not show 'no project work' (no project), got: %s", output)
	}

	// Should show filter in period description
	if !strings.Contains(output, "today (@acme)") {
		t.Errorf("Expected 'today (@acme)' in output, got: %s", output)
	}

	// Total should reflect only filtered entries (1h)
	if !strings.Contains(output, "Total: 1h") {
		t.Errorf("Expected 'Total: 1h' (filtered), got: %s", output)
	}
}

func TestRootCommand_WithShorthandProjectFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for today
	today := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "work on acme",
			DurationMinutes: 60,
			RawInput:        "work on acme @acme for 1h",
			Project:         "acme",
			Tags:            []string{},
		},
		{
			Timestamp:       today,
			Description:     "work on other",
			DurationMinutes: 30,
			RawInput:        "work on other for 30m",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Use shorthand @acme syntax
	rootCmd.Run(rootCmd, []string{"@acme"})

	output := stdout.String()

	// Should show only acme project
	if !strings.Contains(output, "work on acme") {
		t.Errorf("Expected 'work on acme' in output, got: %s", output)
	}

	if strings.Contains(output, "work on other") {
		t.Errorf("Should not show 'work on other', got: %s", output)
	}

	if !strings.Contains(output, "today (@acme)") {
		t.Errorf("Expected 'today (@acme)' in output, got: %s", output)
	}
}

func TestRootCommand_WithTagFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for today with different tags
	today := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "fix bug",
			DurationMinutes: 120,
			RawInput:        "fix bug #bugfix for 2h",
			Project:         "",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       today,
			Description:     "code review",
			DurationMinutes: 60,
			RawInput:        "code review #review for 1h",
			Project:         "",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       today,
			Description:     "untagged work",
			DurationMinutes: 45,
			RawInput:        "untagged work for 45m",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set tag filter flag
	_ = rootCmd.PersistentFlags().Set("tag", "bugfix")

	// Run root command with filter
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show bugfix entries
	if !strings.Contains(output, "fix bug") {
		t.Errorf("Expected 'fix bug' in output, got: %s", output)
	}

	// Should NOT show other tags
	if strings.Contains(output, "code review") {
		t.Errorf("Should not show 'code review' (different tag), got: %s", output)
	}

	if strings.Contains(output, "untagged work") {
		t.Errorf("Should not show 'untagged work' (no tags), got: %s", output)
	}

	// Should show filter in period description
	if !strings.Contains(output, "today (#bugfix)") {
		t.Errorf("Expected 'today (#bugfix)' in output, got: %s", output)
	}

	// Total should reflect only filtered entries (2h)
	if !strings.Contains(output, "Total: 2h") {
		t.Errorf("Expected 'Total: 2h' (filtered), got: %s", output)
	}
}

func TestRootCommand_WithShorthandTagFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for today
	today := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "urgent task",
			DurationMinutes: 90,
			RawInput:        "urgent task #urgent for 1h30m",
			Project:         "",
			Tags:            []string{"urgent"},
		},
		{
			Timestamp:       today,
			Description:     "regular task",
			DurationMinutes: 30,
			RawInput:        "regular task for 30m",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Use shorthand #urgent syntax
	rootCmd.Run(rootCmd, []string{"#urgent"})

	output := stdout.String()

	// Should show only urgent tag
	if !strings.Contains(output, "urgent task") {
		t.Errorf("Expected 'urgent task' in output, got: %s", output)
	}

	if strings.Contains(output, "regular task") {
		t.Errorf("Should not show 'regular task', got: %s", output)
	}

	if !strings.Contains(output, "today (#urgent)") {
		t.Errorf("Expected 'today (#urgent)' in output, got: %s", output)
	}
}

func TestRootCommand_WithMultipleTagFilters(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for today with different tag combinations
	today := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "urgent bug",
			DurationMinutes: 120,
			RawInput:        "urgent bug #urgent #bugfix for 2h",
			Project:         "",
			Tags:            []string{"urgent", "bugfix"},
		},
		{
			Timestamp:       today,
			Description:     "bug only",
			DurationMinutes: 60,
			RawInput:        "bug only #bugfix for 1h",
			Project:         "",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       today,
			Description:     "urgent only",
			DurationMinutes: 45,
			RawInput:        "urgent only #urgent for 45m",
			Project:         "",
			Tags:            []string{"urgent"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set multiple tag filters (AND logic)
	_ = rootCmd.PersistentFlags().Set("tag", "urgent")
	tagFlag := rootCmd.PersistentFlags().Lookup("tag")
	_ = tagFlag.Value.Set("bugfix")

	// Run root command with multiple tag filters
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show only entries with ALL tags (AND logic)
	if !strings.Contains(output, "urgent bug") {
		t.Errorf("Expected 'urgent bug' (has both tags) in output, got: %s", output)
	}

	// Should NOT show entries with only one tag
	if strings.Contains(output, "bug only") {
		t.Errorf("Should not show 'bug only' (missing urgent tag), got: %s", output)
	}

	if strings.Contains(output, "urgent only") {
		t.Errorf("Should not show 'urgent only' (missing bugfix tag), got: %s", output)
	}

	// Should show both filters in period description
	if !strings.Contains(output, "#urgent") || !strings.Contains(output, "#bugfix") {
		t.Errorf("Expected both '#urgent' and '#bugfix' in output, got: %s", output)
	}

	// Total should reflect only filtered entries (2h)
	if !strings.Contains(output, "Total: 2h") {
		t.Errorf("Expected 'Total: 2h' (filtered), got: %s", output)
	}
}

func TestRootCommand_WithProjectAndTagFilters(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for today with different project and tag combinations
	today := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "acme urgent work",
			DurationMinutes: 90,
			RawInput:        "acme urgent work @acme #urgent for 1h30m",
			Project:         "acme",
			Tags:            []string{"urgent"},
		},
		{
			Timestamp:       today,
			Description:     "acme regular work",
			DurationMinutes: 60,
			RawInput:        "acme regular work @acme for 1h",
			Project:         "acme",
			Tags:            []string{},
		},
		{
			Timestamp:       today,
			Description:     "client urgent work",
			DurationMinutes: 45,
			RawInput:        "client urgent work @client #urgent for 45m",
			Project:         "client",
			Tags:            []string{"urgent"},
		},
		{
			Timestamp:       today,
			Description:     "untagged work",
			DurationMinutes: 30,
			RawInput:        "untagged work for 30m",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set both project and tag filters (AND logic)
	_ = rootCmd.PersistentFlags().Set("project", "acme")
	_ = rootCmd.PersistentFlags().Set("tag", "urgent")

	// Run root command with combined filters
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show only entries matching BOTH filters
	if !strings.Contains(output, "acme urgent work") {
		t.Errorf("Expected 'acme urgent work' (matches both filters) in output, got: %s", output)
	}

	// Should NOT show entries matching only one filter
	if strings.Contains(output, "acme regular work") {
		t.Errorf("Should not show 'acme regular work' (no urgent tag), got: %s", output)
	}

	if strings.Contains(output, "client urgent work") {
		t.Errorf("Should not show 'client urgent work' (different project), got: %s", output)
	}

	if strings.Contains(output, "untagged work") {
		t.Errorf("Should not show 'untagged work' (no project/tag), got: %s", output)
	}

	// Should show both filters in period description
	if !strings.Contains(output, "today (@acme #urgent)") {
		t.Errorf("Expected 'today (@acme #urgent)' in output, got: %s", output)
	}

	// Total should reflect only filtered entries (1h 30m)
	if !strings.Contains(output, "Total: 1h 30m") {
		t.Errorf("Expected 'Total: 1h 30m' (filtered), got: %s", output)
	}
}

func TestRootCommand_WithShorthandProjectAndTagFilters(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for today
	today := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "acme urgent",
			DurationMinutes: 90,
			RawInput:        "acme urgent @acme #urgent for 1h30m",
			Project:         "acme",
			Tags:            []string{"urgent"},
		},
		{
			Timestamp:       today,
			Description:     "acme only",
			DurationMinutes: 60,
			RawInput:        "acme only @acme for 1h",
			Project:         "acme",
			Tags:            []string{},
		},
		{
			Timestamp:       today,
			Description:     "urgent only",
			DurationMinutes: 45,
			RawInput:        "urgent only #urgent for 45m",
			Project:         "",
			Tags:            []string{"urgent"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags to avoid contamination from other tests
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Use shorthand @acme #urgent syntax
	rootCmd.Run(rootCmd, []string{"@acme", "#urgent"})

	output := stdout.String()

	// Should show only entries matching BOTH filters
	if !strings.Contains(output, "acme urgent") {
		t.Errorf("Expected 'acme urgent' (matches both filters) in output, got: %s", output)
	}

	if strings.Contains(output, "acme only") {
		t.Errorf("Should not show 'acme only' (no urgent tag), got: %s", output)
	}

	if strings.Contains(output, "urgent only") {
		t.Errorf("Should not show 'urgent only' (no acme project), got: %s", output)
	}

	if !strings.Contains(output, "today (@acme #urgent)") {
		t.Errorf("Expected 'today (@acme #urgent)' in output, got: %s", output)
	}
}

// Tests for new time period flags

func TestLastFlag_ValidDays(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for different days
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "today work",
			DurationMinutes: 60,
			RawInput:        "today work for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -3),
			Description:     "three days ago",
			DurationMinutes: 90,
			RawInput:        "three days ago for 1h30m",
		},
		{
			Timestamp:       now.AddDate(0, 0, -10),
			Description:     "ten days ago",
			DurationMinutes: 120,
			RawInput:        "ten days ago for 2h",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set --last 7 flag
	_ = rootCmd.Flags().Set("last", "7")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show entries from last 7 days
	if !strings.Contains(output, "today work") {
		t.Errorf("Expected 'today work' in output, got: %s", output)
	}
	if !strings.Contains(output, "three days ago") {
		t.Errorf("Expected 'three days ago' in output, got: %s", output)
	}
	// Should NOT show entry from 10 days ago
	if strings.Contains(output, "ten days ago") {
		t.Errorf("Should not show 'ten days ago' (beyond 7 days), got: %s", output)
	}
	// Header should mention "last 7 days"
	if !strings.Contains(output, "last 7 days") {
		t.Errorf("Expected header to mention 'last 7 days', got: %s", output)
	}
}

func TestDateFlag_SpecificDate(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for different days
	targetDate := time.Date(2024, 6, 15, 10, 0, 0, 0, time.Local)
	entries := []entry.Entry{
		{
			Timestamp:       targetDate,
			Description:     "work on june 15",
			DurationMinutes: 60,
			RawInput:        "work on june 15 for 1h",
		},
		{
			Timestamp:       targetDate.Add(2 * time.Hour),
			Description:     "more work june 15",
			DurationMinutes: 30,
			RawInput:        "more work june 15 for 30m",
		},
		{
			Timestamp:       targetDate.AddDate(0, 0, 1),
			Description:     "work on june 16",
			DurationMinutes: 120,
			RawInput:        "work on june 16 for 2h",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set --date flag
	_ = rootCmd.Flags().Set("date", "2024-06-15")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show entries from June 15 only
	if !strings.Contains(output, "work on june 15") {
		t.Errorf("Expected 'work on june 15' in output, got: %s", output)
	}
	if !strings.Contains(output, "more work june 15") {
		t.Errorf("Expected 'more work june 15' in output, got: %s", output)
	}
	// Should NOT show entry from June 16
	if strings.Contains(output, "work on june 16") {
		t.Errorf("Should not show 'work on june 16', got: %s", output)
	}
}

func TestFromToFlags_DateRange(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries for different days
	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, 6, 10, 10, 0, 0, 0, time.Local),
			Description:     "work on june 10",
			DurationMinutes: 60,
			RawInput:        "work on june 10 for 1h",
		},
		{
			Timestamp:       time.Date(2024, 6, 15, 10, 0, 0, 0, time.Local),
			Description:     "work on june 15",
			DurationMinutes: 90,
			RawInput:        "work on june 15 for 1h30m",
		},
		{
			Timestamp:       time.Date(2024, 6, 20, 10, 0, 0, 0, time.Local),
			Description:     "work on june 20",
			DurationMinutes: 120,
			RawInput:        "work on june 20 for 2h",
		},
		{
			Timestamp:       time.Date(2024, 6, 25, 10, 0, 0, 0, time.Local),
			Description:     "work on june 25",
			DurationMinutes: 30,
			RawInput:        "work on june 25 for 30m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Reset flags
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set --from and --to flags
	_ = rootCmd.Flags().Set("from", "2024-06-14")
	_ = rootCmd.Flags().Set("to", "2024-06-21")
	rootCmd.Run(rootCmd, []string{})

	output := stdout.String()

	// Should show entries in range only
	if !strings.Contains(output, "work on june 15") {
		t.Errorf("Expected 'work on june 15' in output, got: %s", output)
	}
	if !strings.Contains(output, "work on june 20") {
		t.Errorf("Expected 'work on june 20' in output, got: %s", output)
	}
	// Should NOT show entries outside range
	if strings.Contains(output, "work on june 10") {
		t.Errorf("Should not show 'work on june 10' (before range), got: %s", output)
	}
	if strings.Contains(output, "work on june 25") {
		t.Errorf("Should not show 'work on june 25' (after range), got: %s", output)
	}
}

func TestTimePeriodFlags_MutualExclusivity(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	// Reset flags
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set multiple time flags
	_ = rootCmd.Flags().Set("yesterday", "true")
	_ = rootCmd.Flags().Set("this-week", "true")
	rootCmd.Run(rootCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called due to mutual exclusivity")
	}
	if !strings.Contains(stderr.String(), "mutually exclusive") {
		t.Errorf("Expected mutual exclusivity error, got: %s", stderr.String())
	}
}

func TestTimePeriodFlags_CannotCreateEntryWithTimeFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	// Reset flags
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set time flag and try to create entry
	_ = rootCmd.Flags().Set("yesterday", "true")
	rootCmd.Run(rootCmd, []string{"test", "task", "for", "1h"})

	if !exitCalled {
		t.Error("Expected exit to be called - time flags cannot be used with entry creation")
	}
	if !strings.Contains(stderr.String(), "cannot be used when creating entries") {
		t.Errorf("Expected error about time flags with entry creation, got: %s", stderr.String())
	}
}

func TestDateFlag_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	// Reset flags
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set invalid date format
	_ = rootCmd.Flags().Set("date", "invalid-date")
	rootCmd.Run(rootCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid date")
	}
	if !strings.Contains(stderr.String(), "Invalid --date") {
		t.Errorf("Expected invalid date error, got: %s", stderr.String())
	}
}

func TestFromToFlags_FromAfterTo(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	// Reset flags
	resetTimePeriodFlags(rootCmd)
	resetFilterFlags(rootCmd)

	// Set from date after to date
	_ = rootCmd.Flags().Set("from", "2024-06-20")
	_ = rootCmd.Flags().Set("to", "2024-06-10")
	rootCmd.Run(rootCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for from > to")
	}
	if !strings.Contains(stderr.String(), "is after") {
		t.Errorf("Expected 'is after' error, got: %s", stderr.String())
	}
}
