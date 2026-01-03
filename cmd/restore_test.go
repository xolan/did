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

func TestRestoreFromBackup_NoBackups(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	d, stdout, _ := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	restoreFromBackup([]string{})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stdout.String(), "No backups available") {
		t.Errorf("Expected 'No backups available', got: %s", stdout.String())
	}
}

func TestRestoreFromBackup_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create initial entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original",
		DurationMinutes: 60,
		RawInput:        "original for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	// Create a backup
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify the main file
	modifiedEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "modified",
		DurationMinutes: 120,
		RawInput:        "modified for 2h",
	}
	if err := storage.WriteEntries(storagePath, []entry.Entry{modifiedEntry}); err != nil {
		t.Fatalf("Failed to modify storage: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	restoreFromBackup([]string{})

	output := stdout.String()
	if !strings.Contains(output, "Available backups:") {
		t.Errorf("Expected 'Available backups:', got: %s", output)
	}
	if !strings.Contains(output, "Successfully restored from backup 1") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify content was restored
	entries, _ := storage.ReadEntries(storagePath)
	if len(entries) != 1 || entries[0].Description != "original" {
		t.Errorf("Expected restored entry with 'original', got: %v", entries)
	}
}

func TestRestoreFromBackup_InvalidNumber(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entry and backup
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	restoreFromBackup([]string{"invalid"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Invalid backup number") {
		t.Errorf("Expected 'Invalid backup number' error, got: %s", stderr.String())
	}
}

func TestRestoreFromBackup_BackupNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entry and backup
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	restoreFromBackup([]string{"3"})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Backup 3 does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %s", stderr.String())
	}
}

func TestRestoreFromBackup_SpecificBackupNumber(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entry
	entry1 := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "first",
		DurationMinutes: 60,
		RawInput:        "first for 1h",
	}
	if err := storage.AppendEntry(storagePath, entry1); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	// Create first backup
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create first backup: %v", err)
	}

	// Modify and create second backup
	entry2 := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "second",
		DurationMinutes: 120,
		RawInput:        "second for 2h",
	}
	if err := storage.WriteEntries(storagePath, []entry.Entry{entry2}); err != nil {
		t.Fatalf("Failed to modify storage: %v", err)
	}
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create second backup: %v", err)
	}

	// Modify main file again
	entry3 := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "third",
		DurationMinutes: 180,
		RawInput:        "third for 3h",
	}
	if err := storage.WriteEntries(storagePath, []entry.Entry{entry3}); err != nil {
		t.Fatalf("Failed to modify storage: %v", err)
	}

	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	// Restore from backup 2 (which should have "first" entry)
	restoreFromBackup([]string{"2"})

	output := stdout.String()
	if !strings.Contains(output, "Successfully restored from backup 2") {
		t.Errorf("Expected success for backup 2, got: %s", output)
	}

	// Verify content was restored from backup 2
	entries, _ := storage.ReadEntries(storagePath)
	if len(entries) != 1 || entries[0].Description != "first" {
		t.Errorf("Expected 'first' from backup 2, got: %v", entries)
	}
}

func TestRestoreFromBackup_ListBackupsError(t *testing.T) {
	tmpDir := t.TempDir()
	// Use a path that makes it impossible to list backups
	storagePath := filepath.Join(tmpDir, "nonexistent", "entries.jsonl")

	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Create the directory structure
	_ = os.MkdirAll(filepath.Dir(storagePath), 0755)

	restoreFromBackup([]string{})

	// No backups exist, so should exit with "No backups available"
	if !exitCalled {
		t.Error("Expected exit to be called")
	}
}

func TestRestoreFromBackup_StoragePathError(t *testing.T) {
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

	restoreFromBackup([]string{})

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to get storage path") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}
}

func TestRestoreFromBackup_RestoreError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entry and backup
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Corrupt the backup file to cause restore error
	backupPath := storagePath + ".bak.1"
	// Make the backup file a directory to cause read error
	if err := os.Remove(backupPath); err != nil {
		t.Fatalf("Failed to remove backup: %v", err)
	}
	if err := os.Mkdir(backupPath, 0755); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}

	exitCalled := false
	d, _, stderr := testDeps(storagePath)
	d.Exit = func(code int) { exitCalled = true }
	SetDeps(d)
	defer ResetDeps()

	restoreFromBackup([]string{})

	// Should fail because backup is a directory, not a file
	if !exitCalled {
		t.Error("Expected exit to be called")
	}
	if !strings.Contains(stderr.String(), "Failed to restore backup") {
		t.Errorf("Expected restore error, got: %s", stderr.String())
	}
}

func TestRestoreCommand_Run(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entry and backup
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "original",
		DurationMinutes: 60,
		RawInput:        "original for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
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

	// Call the restore command's Run function directly
	restoreCmd.Run(restoreCmd, []string{})

	if !strings.Contains(stdout.String(), "Successfully restored from backup") {
		t.Errorf("Expected 'Successfully restored from backup', got: %s", stdout.String())
	}
}
