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

func TestPurgeDeleted_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entries
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.Add(-2 * time.Hour),
			Description:     "entry one",
			DurationMinutes: 60,
			RawInput:        "entry one for 1h",
		},
		{
			Timestamp:       now.Add(-1 * time.Hour),
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

	// Soft delete the second entry
	_, err := storage.SoftDeleteEntry(storagePath, 1)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
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

	purgeYesFlag = true
	defer func() { purgeYesFlag = false }()

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "Purged 1 entry") {
		t.Errorf("Expected 'Purged 1 entry' in output, got: %s", output)
	}

	// Verify deleted entry was permanently removed
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 1 {
		t.Errorf("Expected 1 total entry, got %d", len(allEntries))
	}
	if allEntries[0].Description != "entry one" {
		t.Errorf("Expected 'entry one' to remain, got: %s", allEntries[0].Description)
	}
}

func TestPurgeDeleted_NoDeletedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry (but don't delete it)
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "active entry",
		DurationMinutes: 60,
		RawInput:        "active entry for 1h",
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

	purgeYesFlag = true
	defer func() { purgeYesFlag = false }()

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "No deleted entries to purge") {
		t.Errorf("Expected 'No deleted entries to purge' in output, got: %s", output)
	}

	// Verify entry remains
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 1 {
		t.Errorf("Expected 1 entry to remain, got %d", len(allEntries))
	}
}

func TestPurgeDeleted_EmptyStorage(t *testing.T) {
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

	purgeYesFlag = true
	defer func() { purgeYesFlag = false }()

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "No deleted entries to purge") {
		t.Errorf("Expected 'No deleted entries to purge' in output, got: %s", output)
	}
}

func TestPurgeDeleted_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entries
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.Add(-3 * time.Hour),
			Description:     "entry one",
			DurationMinutes: 60,
			RawInput:        "entry one for 1h",
		},
		{
			Timestamp:       now.Add(-2 * time.Hour),
			Description:     "entry two",
			DurationMinutes: 30,
			RawInput:        "entry two for 30m",
		},
		{
			Timestamp:       now.Add(-1 * time.Hour),
			Description:     "entry three",
			DurationMinutes: 45,
			RawInput:        "entry three for 45m",
		},
	}
	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Soft delete entries two and three
	_, err := storage.SoftDeleteEntry(storagePath, 1)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
	}
	_, err = storage.SoftDeleteEntry(storagePath, 2)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
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

	purgeYesFlag = true
	defer func() { purgeYesFlag = false }()

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "Purged 2 entries") {
		t.Errorf("Expected 'Purged 2 entries' in output, got: %s", output)
	}

	// Verify only entry one remains
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 1 {
		t.Errorf("Expected 1 entry to remain, got %d", len(allEntries))
	}
	if allEntries[0].Description != "entry one" {
		t.Errorf("Expected 'entry one' to remain, got: %s", allEntries[0].Description)
	}
}

func TestPurgeDeleted_AllEntriesDeleted(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entries
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.Add(-2 * time.Hour),
			Description:     "entry one",
			DurationMinutes: 60,
			RawInput:        "entry one for 1h",
		},
		{
			Timestamp:       now.Add(-1 * time.Hour),
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

	// Soft delete all entries
	_, err := storage.SoftDeleteEntry(storagePath, 0)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
	}
	_, err = storage.SoftDeleteEntry(storagePath, 1)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
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

	purgeYesFlag = true
	defer func() { purgeYesFlag = false }()

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "Purged 2 entries") {
		t.Errorf("Expected 'Purged 2 entries' in output, got: %s", output)
	}

	// Verify all entries were removed
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 0 {
		t.Errorf("Expected 0 entries to remain, got %d", len(allEntries))
	}
}

func TestPurgeDeleted_ConfirmationYes(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create and delete test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	_, err := storage.SoftDeleteEntry(storagePath, 0)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
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

	purgeYesFlag = false

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "Purged 1 entry") {
		t.Errorf("Expected 'Purged 1 entry' in output, got: %s", output)
	}

	// Verify entry was purged
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 0 {
		t.Errorf("Expected 0 entries to remain, got %d", len(allEntries))
	}
}

func TestPurgeDeleted_ConfirmationNo(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create and delete test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	_, err := storage.SoftDeleteEntry(storagePath, 0)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
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

	purgeYesFlag = false

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "Purge cancelled") {
		t.Errorf("Expected 'Purge cancelled' in output, got: %s", output)
	}

	// Verify entry was NOT purged (still deleted but exists)
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 1 {
		t.Errorf("Expected 1 entry to remain (soft deleted), got %d", len(allEntries))
	}
}

func TestPurgeDeleted_StoragePathError(t *testing.T) {
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

	purgeYesFlag = true
	defer func() { purgeYesFlag = false }()

	purgeDeleted()

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to get storage path") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}
}

func TestPromptPurgeConfirmation(t *testing.T) {
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

			result := promptPurgeConfirmation()
			if result != tt.expected {
				t.Errorf("promptPurgeConfirmation() with input %q = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPromptPurgeConfirmation_ScannerFail(t *testing.T) {
	d := &Deps{
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
		Stdin:       eofReader{},
		Exit:        func(code int) {},
		StoragePath: func() (string, error) { return "", nil },
	}
	SetDeps(d)
	defer ResetDeps()

	result := promptPurgeConfirmation()
	if result != false {
		t.Error("Expected false when scanner fails")
	}
}

func TestPurgeCommand_Run(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create and delete test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	_, err := storage.SoftDeleteEntry(storagePath, 0)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
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

	// Set the --yes flag for the command
	_ = purgeCmd.Flags().Set("yes", "true")
	defer func() { _ = purgeCmd.Flags().Set("yes", "false") }()

	// Call the purge command's Run function directly
	purgeCmd.Run(purgeCmd, []string{})

	if !strings.Contains(stdout.String(), "Purged 1 entry") {
		t.Errorf("Expected 'Purged 1 entry', got: %s", stdout.String())
	}
}

func TestPurgeDeleted_ConfirmationUppercaseY(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create and delete test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	_, err := storage.SoftDeleteEntry(storagePath, 0)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader("Y\n"),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	purgeYesFlag = false

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "Purged 1 entry") {
		t.Errorf("Expected 'Purged 1 entry' in output (with uppercase Y), got: %s", output)
	}
}

func TestPurgeDeleted_ConfirmationEmptyInput(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create and delete test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	_, err := storage.SoftDeleteEntry(storagePath, 0)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader("\n"),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	purgeYesFlag = false

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "Purge cancelled") {
		t.Errorf("Expected 'Purge cancelled' with empty input, got: %s", output)
	}
}

func TestPurgeDeleted_PurgeEntriesError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create and delete test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test entry",
		DurationMinutes: 60,
		RawInput:        "test entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	_, err := storage.SoftDeleteEntry(storagePath, 0)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
	}

	// Make the storage file read-only to cause PurgeDeletedEntries to fail
	if err := os.Chmod(storagePath, 0444); err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	defer func() {
		_ = os.Chmod(storagePath, 0644)
	}()

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

	purgeYesFlag = true
	defer func() { purgeYesFlag = false }()

	purgeDeleted()

	if !exitCalled {
		t.Error("Expected exit to be called when purge fails")
	}
	if !strings.Contains(stderr.String(), "Failed to purge entries") {
		t.Errorf("Expected 'Failed to purge entries' error, got: %s", stderr.String())
	}
}

func TestPurgeDeleted_ScannerEOF(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create and delete test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	_, err := storage.SoftDeleteEntry(storagePath, 0)
	if err != nil {
		t.Fatalf("Failed to soft delete entry: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  eofReader{},
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	purgeYesFlag = false

	purgeDeleted()

	output := stdout.String()
	if !strings.Contains(output, "Purge cancelled") {
		t.Errorf("Expected 'Purge cancelled' with scanner EOF, got: %s", output)
	}

	// Verify entry was NOT purged
	allEntries, _ := storage.ReadEntries(storagePath)
	if len(allEntries) != 1 {
		t.Errorf("Expected 1 entry to remain (soft deleted), got %d", len(allEntries))
	}
}
