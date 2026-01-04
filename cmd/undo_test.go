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

func TestUndoDelete_Success(t *testing.T) {
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

	undoDelete()

	output := stdout.String()
	if !strings.Contains(output, "Restored:") {
		t.Errorf("Expected 'Restored:' in output, got: %s", output)
	}
	if !strings.Contains(output, "entry two") {
		t.Errorf("Expected 'entry two' in output, got: %s", output)
	}

	// Verify entry was restored (both entries should be active now)
	activeEntries, _ := storage.ReadActiveEntries(storagePath)
	if len(activeEntries) != 2 {
		t.Errorf("Expected 2 active entries, got %d", len(activeEntries))
	}
}

func TestUndoDelete_NoDeletedEntries(t *testing.T) {
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

	undoDelete()

	if !exitCalled {
		t.Error("Expected exit to be called when no deleted entries")
	}
	if !strings.Contains(stderr.String(), "Hint: No entries to restore") {
		t.Errorf("Expected 'Hint: No entries to restore' in error, got: %s", stderr.String())
	}
}

func TestUndoDelete_EmptyStorage(t *testing.T) {
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

	undoDelete()

	if !exitCalled {
		t.Error("Expected exit to be called when storage is empty")
	}
	if !strings.Contains(stderr.String(), "Hint: No entries to restore") {
		t.Errorf("Expected 'Hint: No entries to restore' in error, got: %s", stderr.String())
	}
}

func TestUndoDelete_RestoresMostRecent(t *testing.T) {
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

	// Soft delete entry two first (earlier deletion)
	time.Sleep(10 * time.Millisecond)
	_, err := storage.SoftDeleteEntry(storagePath, 1)
	if err != nil {
		t.Fatalf("Failed to soft delete entry two: %v", err)
	}

	// Soft delete entry three second (later deletion - should be restored first)
	time.Sleep(10 * time.Millisecond)
	_, err = storage.SoftDeleteEntry(storagePath, 2)
	if err != nil {
		t.Fatalf("Failed to soft delete entry three: %v", err)
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

	undoDelete()

	output := stdout.String()
	if !strings.Contains(output, "entry three") {
		t.Errorf("Expected 'entry three' (most recently deleted) in output, got: %s", output)
	}

	// Verify entry three was restored, but entry two is still deleted
	activeEntries, _ := storage.ReadActiveEntries(storagePath)
	if len(activeEntries) != 2 {
		t.Errorf("Expected 2 active entries, got %d", len(activeEntries))
	}

	// Check that entry one and entry three are active
	foundEntryOne := false
	foundEntryThree := false
	for _, e := range activeEntries {
		if e.Description == "entry one" {
			foundEntryOne = true
		}
		if e.Description == "entry three" {
			foundEntryThree = true
		}
	}
	if !foundEntryOne || !foundEntryThree {
		t.Error("Expected entry one and entry three to be active")
	}
}

func TestUndoDelete_StoragePathError(t *testing.T) {
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

	undoDelete()

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to get storage path") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}
}

func TestUndoDelete_RestoreError(t *testing.T) {
	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			// Return a path that doesn't exist to cause restore error
			return "/nonexistent/path/entries.jsonl", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	undoDelete()

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	// Should fail on GetMostRecentlyDeleted since the path doesn't exist
	if stderr.String() == "" {
		t.Error("Expected error message in stderr")
	}
}

func TestUndoDelete_WithProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry with project and tags
	now := time.Now()
	testEntry := entry.Entry{
		Timestamp:       now,
		Description:     "implemented feature",
		Project:         "myproject",
		Tags:            []string{"bug", "urgent"},
		DurationMinutes: 120,
		RawInput:        "implemented feature +myproject #bug #urgent for 2h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	// Soft delete the entry
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

	undoDelete()

	output := stdout.String()
	if !strings.Contains(output, "implemented feature") {
		t.Errorf("Expected description in output, got: %s", output)
	}
	if !strings.Contains(output, "@myproject") {
		t.Errorf("Expected project in output, got: %s", output)
	}
	if !strings.Contains(output, "#bug") || !strings.Contains(output, "#urgent") {
		t.Errorf("Expected tags in output, got: %s", output)
	}
	if !strings.Contains(output, "2h") {
		t.Errorf("Expected duration in output, got: %s", output)
	}
}

func TestUndoCommand_Run(t *testing.T) {
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

	// Call the undo command's Run function directly
	undoCmd.Run(undoCmd, []string{})

	if !strings.Contains(stdout.String(), "Restored:") {
		t.Errorf("Expected 'Restored:', got: %s", stdout.String())
	}
}
