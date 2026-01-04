package cmd

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		{"zero minutes", 0, "0m"},
		{"single minute", 1, "1m"},
		{"30 minutes", 30, "30m"},
		{"59 minutes", 59, "59m"},
		{"exactly 1 hour", 60, "1h"},
		{"exactly 2 hours", 120, "2h"},
		{"1 hour 30 minutes", 90, "1h 30m"},
		{"2 hours 15 minutes", 135, "2h 15m"},
		{"10 hours", 600, "10h"},
		{"10 hours 5 minutes", 605, "10h 5m"},
		{"24 hours", 1440, "24h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.minutes)
			if result != tt.expected {
				t.Errorf("formatDuration(%d) = %q, expected %q", tt.minutes, result, tt.expected)
			}
		})
	}
}

func TestDeleteEntry_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entries
	entries := []entry.Entry{
		{
			Timestamp:       time.Now(),
			Description:     "entry one",
			DurationMinutes: 60,
			RawInput:        "entry one for 1h",
		},
		{
			Timestamp:       time.Now(),
			Description:     "entry two",
			DurationMinutes: 30,
			RawInput:        "entry two for 30m",
		},
	}
	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader("y\n"),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	yesFlag = true
	defer func() { yesFlag = false }()

	deleteEntry("1")

	output := stdout.String()
	if !strings.Contains(output, "Deleted:") {
		t.Errorf("Expected 'Deleted:' in output, got: %s", output)
	}

	// Verify entry was soft deleted (marked as deleted, not removed)
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 2 {
		t.Errorf("Expected 2 total entries (with deleted), got %d", len(allEntries))
	}

	// Verify only one active entry remains
	activeEntries, _ := storage.ReadActiveEntries(storagePath)
	if len(activeEntries) != 1 {
		t.Errorf("Expected 1 active entry, got %d", len(activeEntries))
	}
	if activeEntries[0].Description != "entry two" {
		t.Errorf("Expected 'entry two' to remain active, got: %s", activeEntries[0].Description)
	}
}

func TestDeleteEntry_InvalidIndex(t *testing.T) {
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

	deleteEntry("abc")

	if !exitCalled {
		t.Error("Expected exit to be called for invalid index")
	}
	if !strings.Contains(stderr.String(), "must be a number") {
		t.Errorf("Expected 'must be a number' error, got: %s", stderr.String())
	}
}

func TestDeleteEntry_IndexOutOfRange(t *testing.T) {
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

	deleteEntry("99")

	if !exitCalled {
		t.Error("Expected exit to be called for out of range index")
	}
	if !strings.Contains(stderr.String(), "out of range") {
		t.Errorf("Expected 'out of range' error, got: %s", stderr.String())
	}
}

func TestDeleteEntry_NegativeIndex(t *testing.T) {
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

	deleteEntry("0")

	if !exitCalled {
		t.Error("Expected exit to be called for zero index")
	}
	if !strings.Contains(stderr.String(), "must be 1 or greater") {
		t.Errorf("Expected 'must be 1 or greater' error, got: %s", stderr.String())
	}
}

func TestDeleteEntry_NoEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

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

	deleteEntry("1")

	if !exitCalled {
		t.Error("Expected exit to be called when no entries")
	}
	if !strings.Contains(stderr.String(), "No entries to delete") {
		t.Errorf("Expected 'No entries to delete' error, got: %s", stderr.String())
	}
}

func TestDeleteEntry_ConfirmationNo(t *testing.T) {
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

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader("n\n"),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	yesFlag = false

	deleteEntry("1")

	if !strings.Contains(stdout.String(), "Deletion cancelled") {
		t.Errorf("Expected 'Deletion cancelled', got: %s", stdout.String())
	}

	// Verify entry was NOT deleted (should still be active)
	activeEntries, _ := storage.ReadActiveEntries(storagePath)
	if len(activeEntries) != 1 {
		t.Errorf("Expected entry to still be active, got %d active entries", len(activeEntries))
	}
}

func TestDeleteEntry_ConfirmationYes(t *testing.T) {
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

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader("y\n"),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	yesFlag = false

	deleteEntry("1")

	if !strings.Contains(stdout.String(), "Deleted:") {
		t.Errorf("Expected 'Deleted:', got: %s", stdout.String())
	}

	// Verify entry was soft deleted (exists but not active)
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 1 {
		t.Errorf("Expected 1 entry in storage (soft deleted), got %d entries", len(allEntries))
	}

	// Verify no active entries remain
	activeEntries, _ := storage.ReadActiveEntries(storagePath)
	if len(activeEntries) != 0 {
		t.Errorf("Expected 0 active entries, got %d active entries", len(activeEntries))
	}
}

func TestShowEntryForDeletion(t *testing.T) {
	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	testEntry := entry.Entry{
		Timestamp:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Description:     "test entry",
		DurationMinutes: 60,
		RawInput:        "test entry for 1h",
	}

	showEntryForDeletion(testEntry)

	output := stdout.String()
	if !strings.Contains(output, "Entry to delete:") {
		t.Errorf("Expected 'Entry to delete:', got: %s", output)
	}
	if !strings.Contains(output, "test entry") {
		t.Errorf("Expected entry description in output, got: %s", output)
	}
}

func TestPromptConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"lowercase y", "y\n", true},
		{"uppercase Y", "Y\n", true},
		{"lowercase n", "n\n", false},
		{"uppercase N", "N\n", false},
		{"empty input", "\n", false},
		{"random text", "maybe\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deps{
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
				Stdin:  strings.NewReader(tt.input),
				Exit:   func(code int) {},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			result := promptConfirmation()
			if result != tt.expected {
				t.Errorf("promptConfirmation() with input %q = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDeleteEntry_StoragePathError(t *testing.T) {
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

	deleteEntry("1")

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to get storage path") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}
}

func TestDeleteEntry_ReadEntriesError(t *testing.T) {
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

	deleteEntry("1")

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to read entries") {
		t.Errorf("Expected read error, got: %s", stderr.String())
	}
}

func TestDeleteEntry_DeleteError(t *testing.T) {
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

	// Make the file read-only to cause write error during delete
	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			// Return a path that won't work for writing
			return "/nonexistent/path/entries.jsonl", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	yesFlag = true
	defer func() { yesFlag = false }()

	deleteEntry("1")

	// Should fail on reading entries since the path doesn't exist
	if !exitCalled {
		t.Error("Expected exit to be called")
	}
}

// eofReader is an io.Reader that immediately returns EOF
type eofReader struct{}

func (e eofReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func TestPromptConfirmation_ScannerFail(t *testing.T) {
	d := &Deps{
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
		Stdin:       eofReader{},
		Exit:        func(code int) {},
		StoragePath: func() (string, error) { return "", nil },
	}
	SetDeps(d)
	defer ResetDeps()

	result := promptConfirmation()
	if result != false {
		t.Error("Expected false when scanner fails")
	}
}

func TestDeleteCommand_Run(t *testing.T) {
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

	yesFlag = true
	defer func() { yesFlag = false }()

	// Call the delete command's Run function directly
	deleteCmd.Run(deleteCmd, []string{"1"})

	if !strings.Contains(stdout.String(), "Deleted:") {
		t.Errorf("Expected 'Deleted:', got: %s", stdout.String())
	}
}

func TestEditCommand_Run(t *testing.T) {
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

	_ = editCmd.Flags().Set("description", "updated")
	defer func() { _ = editCmd.Flags().Set("description", "") }()

	// Call the edit command's Run function directly
	editCmd.Run(editCmd, []string{"1"})

	if !strings.Contains(stdout.String(), "Updated entry 1") {
		t.Errorf("Expected 'Updated entry 1', got: %s", stdout.String())
	}
}
