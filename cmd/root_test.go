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
	}, stdout, stderr
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

	listEntries("today", func() (time.Time, time.Time) {
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

	listEntries("today", func() (time.Time, time.Time) {
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

	listEntries("today", func() (time.Time, time.Time) {
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

	listEntries("today", func() (time.Time, time.Time) {
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

	listEntries("today", func() (time.Time, time.Time) {
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

	// Call the root command's Run function with args
	rootCmd.Run(rootCmd, []string{"test", "task", "for", "1h"})

	if !strings.Contains(stdout.String(), "Logged:") {
		t.Errorf("Expected 'Logged:', got: %s", stdout.String())
	}
}

func TestYesterday_Command(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	yCmd.Run(yCmd, []string{})

	if !strings.Contains(stdout.String(), "No entries found for yesterday") {
		t.Errorf("Expected 'No entries found for yesterday', got: %s", stdout.String())
	}
}

func TestThisWeek_Command(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	wCmd.Run(wCmd, []string{})

	if !strings.Contains(stdout.String(), "No entries found for this week") {
		t.Errorf("Expected 'No entries found for this week', got: %s", stdout.String())
	}
}

func TestLastWeek_Command(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	lwCmd.Run(lwCmd, []string{})

	if !strings.Contains(stdout.String(), "No entries found for last week") {
		t.Errorf("Expected 'No entries found for last week', got: %s", stdout.String())
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
