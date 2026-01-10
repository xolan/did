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

func TestRestoreFromBackup_BackupNumberOutOfRange(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

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

	tests := []struct {
		name     string
		arg      string
		expected string
	}{
		{"zero", "0", "must be between 1 and 3"},
		{"negative", "-1", "must be between 1 and 3"},
		{"too high", "4", "must be between 1 and 3"},
		{"way too high", "100", "must be between 1 and 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCalled := false
			d, _, stderr := testDeps(storagePath)
			d.Exit = func(code int) { exitCalled = true }
			SetDeps(d)
			defer ResetDeps()

			restoreFromBackup([]string{tt.arg})

			if !exitCalled {
				t.Error("Expected exit to be called")
			}
			if !strings.Contains(stderr.String(), tt.expected) {
				t.Errorf("Expected '%s' error, got: %s", tt.expected, stderr.String())
			}
		})
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

// TestBackupIncludesSoftDeletedEntries verifies that backups include soft-deleted entries
func TestBackupIncludesSoftDeletedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create two entries
	entry1 := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "active entry",
		DurationMinutes: 60,
		RawInput:        "active for 1h",
	}
	entry2 := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "to be deleted",
		DurationMinutes: 30,
		RawInput:        "deleted for 30m",
	}
	if err := storage.AppendEntry(storagePath, entry1); err != nil {
		t.Fatalf("Failed to create entry1: %v", err)
	}
	if err := storage.AppendEntry(storagePath, entry2); err != nil {
		t.Fatalf("Failed to create entry2: %v", err)
	}

	// Soft-delete the second entry
	if _, err := storage.SoftDeleteEntry(storagePath, 1); err != nil {
		t.Fatalf("Failed to soft-delete entry: %v", err)
	}

	// Verify we have 2 total entries (1 active, 1 deleted)
	allEntries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(allEntries) != 2 {
		t.Fatalf("Expected 2 total entries, got %d", len(allEntries))
	}
	if allEntries[1].DeletedAt == nil {
		t.Fatal("Expected entry to be soft-deleted")
	}

	// Create a backup
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Read the backup file directly and verify it contains both entries
	backupPath := storagePath + ".bak.1"
	backupEntries, err := storage.ReadEntries(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	if len(backupEntries) != 2 {
		t.Errorf("Backup contains %d entries, expected 2 (including soft-deleted)", len(backupEntries))
	}

	// Verify the backup includes the soft-deleted entry with DeletedAt set
	if backupEntries[1].DeletedAt == nil {
		t.Error("Backup should include soft-deleted entry with DeletedAt timestamp")
	}
	if backupEntries[1].Description != "to be deleted" {
		t.Errorf("Backup entry[1] description = %q, expected 'to be deleted'", backupEntries[1].Description)
	}
}

// TestRestorePreservesSoftDeletedEntries verifies that restore preserves soft-deleted entries
func TestRestorePreservesSoftDeletedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create two entries
	entry1 := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "active entry",
		DurationMinutes: 60,
		RawInput:        "active for 1h",
	}
	entry2 := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "deleted entry",
		DurationMinutes: 30,
		RawInput:        "deleted for 30m",
	}
	if err := storage.AppendEntry(storagePath, entry1); err != nil {
		t.Fatalf("Failed to create entry1: %v", err)
	}
	if err := storage.AppendEntry(storagePath, entry2); err != nil {
		t.Fatalf("Failed to create entry2: %v", err)
	}

	// Soft-delete the second entry
	if _, err := storage.SoftDeleteEntry(storagePath, 1); err != nil {
		t.Fatalf("Failed to soft-delete entry: %v", err)
	}

	// Create a backup
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify the main file (purge all deleted entries)
	count, err := storage.PurgeDeletedEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to purge deleted entries: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected to purge 1 entry, purged %d", count)
	}

	// Verify only active entry remains
	activeEntries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(activeEntries) != 1 {
		t.Fatalf("Expected 1 entry after purge, got %d", len(activeEntries))
	}

	// Restore from backup
	d, stdout, _ := testDeps(storagePath)
	SetDeps(d)
	defer ResetDeps()

	restoreFromBackup([]string{})

	// Verify restore was successful
	if !strings.Contains(stdout.String(), "Successfully restored from backup 1") {
		t.Errorf("Expected success message, got: %s", stdout.String())
	}

	// Read all entries (including deleted)
	restoredEntries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read restored entries: %v", err)
	}

	// Should have 2 entries again (1 active, 1 soft-deleted)
	if len(restoredEntries) != 2 {
		t.Errorf("Restored storage has %d entries, expected 2", len(restoredEntries))
	}

	// Verify soft-deleted entry is restored with DeletedAt
	if restoredEntries[1].DeletedAt == nil {
		t.Error("Restored entry should have DeletedAt timestamp")
	}
	if restoredEntries[1].Description != "deleted entry" {
		t.Errorf("Restored entry description = %q, expected 'deleted entry'", restoredEntries[1].Description)
	}

	// Verify the deleted entry is excluded from active views
	activeOnly, err := storage.ReadActiveEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read active entries: %v", err)
	}
	if len(activeOnly) != 1 {
		t.Errorf("Active entries count = %d, expected 1", len(activeOnly))
	}
	if activeOnly[0].Description != "active entry" {
		t.Errorf("Active entry description = %q, expected 'active entry'", activeOnly[0].Description)
	}
}

// TestBackupRestoreWithMixedEntries verifies backup/restore with mix of active and deleted entries
func TestBackupRestoreWithMixedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create four entries
	now := time.Now()
	entries := []entry.Entry{
		{Timestamp: now, Description: "entry 1", DurationMinutes: 60, RawInput: "entry 1 for 1h"},
		{Timestamp: now, Description: "entry 2", DurationMinutes: 30, RawInput: "entry 2 for 30m"},
		{Timestamp: now, Description: "entry 3", DurationMinutes: 45, RawInput: "entry 3 for 45m"},
		{Timestamp: now, Description: "entry 4", DurationMinutes: 90, RawInput: "entry 4 for 1h30m"},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}
	}

	// Soft-delete entries 1 and 3 (indices 1 and 3)
	if _, err := storage.SoftDeleteEntry(storagePath, 1); err != nil {
		t.Fatalf("Failed to soft-delete entry 1: %v", err)
	}
	if _, err := storage.SoftDeleteEntry(storagePath, 3); err != nil {
		t.Fatalf("Failed to soft-delete entry 3: %v", err)
	}

	// Verify we have 4 total entries (2 active, 2 deleted)
	allEntries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(allEntries) != 4 {
		t.Fatalf("Expected 4 total entries, got %d", len(allEntries))
	}

	deletedCount := 0
	for _, e := range allEntries {
		if e.DeletedAt != nil {
			deletedCount++
		}
	}
	if deletedCount != 2 {
		t.Fatalf("Expected 2 deleted entries, got %d", deletedCount)
	}

	// Create backup
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify storage (delete all entries)
	if err := storage.WriteEntries(storagePath, []entry.Entry{}); err != nil {
		t.Fatalf("Failed to clear storage: %v", err)
	}

	// Restore from backup
	if err := storage.RestoreBackupForStorage(storagePath, 1); err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Verify all entries are restored
	restoredEntries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read restored entries: %v", err)
	}
	if len(restoredEntries) != 4 {
		t.Errorf("Restored %d entries, expected 4", len(restoredEntries))
	}

	// Verify deleted entries are still deleted
	restoredDeletedCount := 0
	for _, e := range restoredEntries {
		if e.DeletedAt != nil {
			restoredDeletedCount++
		}
	}
	if restoredDeletedCount != 2 {
		t.Errorf("Restored storage has %d deleted entries, expected 2", restoredDeletedCount)
	}

	// Verify active entries view shows only 2 entries
	activeEntries, err := storage.ReadActiveEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read active entries: %v", err)
	}
	if len(activeEntries) != 2 {
		t.Errorf("Active entries count = %d, expected 2", len(activeEntries))
	}
}

// TestBackupFormatCompatibility verifies no breaking changes to backup format
func TestBackupFormatCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create an entry without DeletedAt (simulating old format)
	oldFormatEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "old format entry",
		DurationMinutes: 60,
		RawInput:        "old for 1h",
		// DeletedAt is nil (old format)
	}
	if err := storage.AppendEntry(storagePath, oldFormatEntry); err != nil {
		t.Fatalf("Failed to create old format entry: %v", err)
	}

	// Create backup of old format entry
	if err := storage.CreateBackup(storagePath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Add a new entry with soft delete capability
	newFormatEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "new format entry",
		DurationMinutes: 30,
		RawInput:        "new for 30m",
	}
	if err := storage.AppendEntry(storagePath, newFormatEntry); err != nil {
		t.Fatalf("Failed to create new format entry: %v", err)
	}

	// Soft-delete the new entry
	if _, err := storage.SoftDeleteEntry(storagePath, 1); err != nil {
		t.Fatalf("Failed to soft-delete entry: %v", err)
	}

	// Restore from backup (old format)
	if err := storage.RestoreBackupForStorage(storagePath, 1); err != nil {
		t.Fatalf("Failed to restore old format backup: %v", err)
	}

	// Verify old format entry is restored correctly
	restoredEntries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read restored entries: %v", err)
	}
	if len(restoredEntries) != 1 {
		t.Errorf("Expected 1 restored entry, got %d", len(restoredEntries))
	}
	if restoredEntries[0].Description != "old format entry" {
		t.Errorf("Restored entry description = %q, expected 'old format entry'", restoredEntries[0].Description)
	}
	if restoredEntries[0].DeletedAt != nil {
		t.Error("Old format entry should have nil DeletedAt")
	}

	// Verify we can still work with the restored old format entry
	activeEntries, err := storage.ReadActiveEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read active entries: %v", err)
	}
	if len(activeEntries) != 1 {
		t.Errorf("Active entries count = %d, expected 1", len(activeEntries))
	}
}
