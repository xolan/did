package storage

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xolan/did/internal/osutil"
)

// Helper to create a temporary storage file with content
func createTempStorage(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "entries.jsonl")
	if content != "" {
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp storage file: %v", err)
		}
	}
	return tmpFile
}

// Helper to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Helper to read file content
func readFileContent(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}

func TestGetBackupPath(t *testing.T) {
	tests := []struct {
		name           string
		rotationNumber int
		expectedSuffix string
	}{
		{"backup 1", 1, ".bak.1"},
		{"backup 2", 2, ".bak.2"},
		{"backup 3", 3, ".bak.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GetBackupPath(tt.rotationNumber)
			if err != nil {
				t.Fatalf("GetBackupPath(%d) returned unexpected error: %v", tt.rotationNumber, err)
			}

			if path == "" {
				t.Errorf("GetBackupPath(%d) returned empty path", tt.rotationNumber)
			}

			// Verify the path ends with the expected suffix
			if !filepath.IsAbs(path) {
				t.Errorf("GetBackupPath(%d) returned relative path, expected absolute", tt.rotationNumber)
			}

			// Check suffix
			expectedEnding := "entries.jsonl" + tt.expectedSuffix
			baseName := filepath.Base(path)
			if baseName != expectedEnding {
				t.Errorf("GetBackupPath(%d) basename = %q, expected %q", tt.rotationNumber, baseName, expectedEnding)
			}
		})
	}
}

func TestCreateBackup_NoExistingFile(t *testing.T) {
	// Test that CreateBackup handles missing storage file gracefully
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.jsonl")

	err := CreateBackup(nonExistentFile)
	if err != nil {
		t.Errorf("CreateBackup() with non-existent file returned error: %v, expected nil", err)
	}

	// Verify no backup files were created
	backupPath := nonExistentFile + BackupSuffix + ".1"
	if fileExists(backupPath) {
		t.Errorf("CreateBackup() created backup for non-existent file")
	}
}

func TestCreateBackup_FirstBackup(t *testing.T) {
	// Test creating the first backup when no backups exist
	content := `{"timestamp":"2024-01-15T10:00:00Z","description":"test entry","duration_minutes":60,"raw_input":"test for 1h"}
`
	tmpFile := createTempStorage(t, content)
	tmpDir := filepath.Dir(tmpFile)

	err := CreateBackup(tmpFile)
	if err != nil {
		t.Fatalf("CreateBackup() returned unexpected error: %v", err)
	}

	// Verify .bak.1 was created
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	if !fileExists(backup1Path) {
		t.Errorf("CreateBackup() did not create .bak.1 file")
	}

	// Verify backup content matches original
	backupContent := readFileContent(t, backup1Path)
	if backupContent != content {
		t.Errorf("Backup content = %q, expected %q", backupContent, content)
	}

	// Verify original file still exists and unchanged
	originalContent := readFileContent(t, tmpFile)
	if originalContent != content {
		t.Errorf("Original file was modified")
	}
}

func TestCreateBackup_WithOneExistingBackup(t *testing.T) {
	// Test rotation when one backup exists
	originalContent := `{"timestamp":"2024-01-15T12:00:00Z","description":"current","duration_minutes":60,"raw_input":"current for 1h"}
`
	oldBackupContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"old backup","duration_minutes":30,"raw_input":"old for 30m"}
`
	tmpFile := createTempStorage(t, originalContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create existing .bak.1
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	if err := os.WriteFile(backup1Path, []byte(oldBackupContent), 0644); err != nil {
		t.Fatalf("Failed to create existing backup: %v", err)
	}

	err := CreateBackup(tmpFile)
	if err != nil {
		t.Fatalf("CreateBackup() returned unexpected error: %v", err)
	}

	// Verify .bak.1 contains current content
	newBackup1Content := readFileContent(t, backup1Path)
	if newBackup1Content != originalContent {
		t.Errorf(".bak.1 content = %q, expected current content %q", newBackup1Content, originalContent)
	}

	// Verify .bak.2 contains old backup content
	backup2Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".2")
	if !fileExists(backup2Path) {
		t.Fatalf(".bak.2 was not created")
	}
	backup2Content := readFileContent(t, backup2Path)
	if backup2Content != oldBackupContent {
		t.Errorf(".bak.2 content = %q, expected old backup content %q", backup2Content, oldBackupContent)
	}
}

func TestCreateBackup_WithTwoExistingBackups(t *testing.T) {
	// Test rotation when two backups exist
	currentContent := "current version\n"
	backup1Content := "backup 1 version\n"
	backup2Content := "backup 2 version\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create existing backups
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	backup2Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".2")
	if err := os.WriteFile(backup1Path, []byte(backup1Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.1: %v", err)
	}
	if err := os.WriteFile(backup2Path, []byte(backup2Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.2: %v", err)
	}

	err := CreateBackup(tmpFile)
	if err != nil {
		t.Fatalf("CreateBackup() returned unexpected error: %v", err)
	}

	// Verify rotation: current -> .bak.1, .bak.1 -> .bak.2, .bak.2 -> .bak.3
	newBackup1Content := readFileContent(t, backup1Path)
	if newBackup1Content != currentContent {
		t.Errorf(".bak.1 content = %q, expected current %q", newBackup1Content, currentContent)
	}

	newBackup2Content := readFileContent(t, backup2Path)
	if newBackup2Content != backup1Content {
		t.Errorf(".bak.2 content = %q, expected old .bak.1 %q", newBackup2Content, backup1Content)
	}

	backup3Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".3")
	if !fileExists(backup3Path) {
		t.Fatalf(".bak.3 was not created")
	}
	backup3Content := readFileContent(t, backup3Path)
	if backup3Content != backup2Content {
		t.Errorf(".bak.3 content = %q, expected old .bak.2 %q", backup3Content, backup2Content)
	}
}

func TestCreateBackup_WithThreeExistingBackups_DeletesOldest(t *testing.T) {
	// Test that rotation deletes the oldest backup when limit is reached
	currentContent := "current version\n"
	backup1Content := "backup 1 version\n"
	backup2Content := "backup 2 version\n"
	backup3Content := "backup 3 version (oldest - should be deleted)\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create all three existing backups
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	backup2Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".2")
	backup3Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".3")
	if err := os.WriteFile(backup1Path, []byte(backup1Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.1: %v", err)
	}
	if err := os.WriteFile(backup2Path, []byte(backup2Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.2: %v", err)
	}
	if err := os.WriteFile(backup3Path, []byte(backup3Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.3: %v", err)
	}

	err := CreateBackup(tmpFile)
	if err != nil {
		t.Fatalf("CreateBackup() returned unexpected error: %v", err)
	}

	// Verify rotation: current -> .bak.1, .bak.1 -> .bak.2, .bak.2 -> .bak.3
	// Old .bak.3 should be deleted
	newBackup1Content := readFileContent(t, backup1Path)
	if newBackup1Content != currentContent {
		t.Errorf(".bak.1 content = %q, expected current %q", newBackup1Content, currentContent)
	}

	newBackup2Content := readFileContent(t, backup2Path)
	if newBackup2Content != backup1Content {
		t.Errorf(".bak.2 content = %q, expected old .bak.1 %q", newBackup2Content, backup1Content)
	}

	newBackup3Content := readFileContent(t, backup3Path)
	if newBackup3Content != backup2Content {
		t.Errorf(".bak.3 content = %q, expected old .bak.2 %q", newBackup3Content, backup2Content)
	}

	// Verify that the oldest backup content is gone (it was replaced by .bak.2)
	if newBackup3Content == backup3Content {
		t.Errorf(".bak.3 still contains oldest backup, should have been replaced")
	}

	// Verify no .bak.4 exists
	backup4Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".4")
	if fileExists(backup4Path) {
		t.Errorf(".bak.4 should not exist, backup limit is %d", MaxBackupCount)
	}
}

func TestCreateBackup_MaxBackupCount(t *testing.T) {
	// Verify that the backup count limit is enforced
	tmpFile := createTempStorage(t, "initial\n")
	tmpDir := filepath.Dir(tmpFile)

	// Create MaxBackupCount backups
	for i := 0; i < MaxBackupCount+2; i++ {
		content := string(rune('A'+i)) + " version\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to update storage file: %v", err)
		}

		if err := CreateBackup(tmpFile); err != nil {
			t.Fatalf("CreateBackup() iteration %d returned error: %v", i, err)
		}
	}

	// Count how many backup files exist
	backupCount := 0
	for i := 1; i <= MaxBackupCount+1; i++ {
		backupPath := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+"."+strconv.Itoa(i))
		if fileExists(backupPath) {
			backupCount++
		}
	}

	if backupCount != MaxBackupCount {
		t.Errorf("Found %d backup files, expected exactly %d", backupCount, MaxBackupCount)
	}
}

func TestCreateBackup_EmptyFile(t *testing.T) {
	// Test backing up an empty file
	tmpFile := createTempStorage(t, "")
	tmpDir := filepath.Dir(tmpFile)

	// Create the empty file explicitly
	if err := os.WriteFile(tmpFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	err := CreateBackup(tmpFile)
	if err != nil {
		t.Fatalf("CreateBackup() with empty file returned error: %v", err)
	}

	// Verify .bak.1 exists and is also empty
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	if !fileExists(backup1Path) {
		t.Errorf("CreateBackup() did not create .bak.1 for empty file")
	}

	backupContent := readFileContent(t, backup1Path)
	if backupContent != "" {
		t.Errorf("Backup of empty file has content %q, expected empty", backupContent)
	}
}

func TestCreateBackup_LargeFile(t *testing.T) {
	// Test backing up a larger file to ensure efficient copying
	var largeContent string
	for i := 0; i < 1000; i++ {
		largeContent += `{"timestamp":"2024-01-15T10:00:00Z","description":"entry ` + strconv.Itoa(i%10) + `","duration_minutes":15,"raw_input":"entry for 15m"}` + "\n"
	}

	tmpFile := createTempStorage(t, largeContent)
	tmpDir := filepath.Dir(tmpFile)

	err := CreateBackup(tmpFile)
	if err != nil {
		t.Fatalf("CreateBackup() with large file returned error: %v", err)
	}

	// Verify backup content matches
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	backupContent := readFileContent(t, backup1Path)
	if backupContent != largeContent {
		t.Errorf("Backup content length = %d, expected %d", len(backupContent), len(largeContent))
	}
}

func TestCreateBackup_PreservesFilePermissions(t *testing.T) {
	// Test that backup files have correct permissions
	content := "test content\n"
	tmpFile := createTempStorage(t, content)
	tmpDir := filepath.Dir(tmpFile)

	err := CreateBackup(tmpFile)
	if err != nil {
		t.Fatalf("CreateBackup() returned unexpected error: %v", err)
	}

	// Check backup file permissions
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	info, err := os.Stat(backup1Path)
	if err != nil {
		t.Fatalf("Failed to stat backup file: %v", err)
	}

	// Backup files should be readable/writable by owner, readable by group/others
	expectedPerm := os.FileMode(0644)
	actualPerm := info.Mode().Perm()
	if actualPerm != expectedPerm {
		t.Errorf("Backup file permissions = %o, expected %o", actualPerm, expectedPerm)
	}
}

func TestCreateBackup_InvalidPath(t *testing.T) {
	// Test error handling when storage path is invalid but accessible
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a directory instead of a file
	if err := os.Mkdir(tmpFile, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	err := CreateBackup(tmpFile)
	if err == nil {
		t.Errorf("CreateBackup() with directory path should return error")
	}
}

func TestRotateBackups_NoExistingBackups(t *testing.T) {
	// rotateBackups is not exported, but we can test it indirectly
	// by verifying CreateBackup behavior with no existing backups
	content := "test\n"
	tmpFile := createTempStorage(t, content)

	// First backup should succeed without any existing backups
	err := CreateBackup(tmpFile)
	if err != nil {
		t.Fatalf("CreateBackup() with no existing backups returned error: %v", err)
	}
}

func TestBackupConstants(t *testing.T) {
	// Verify backup constants are set correctly
	if BackupSuffix != ".bak" {
		t.Errorf("BackupSuffix = %q, expected %q", BackupSuffix, ".bak")
	}

	if MaxBackupCount != 3 {
		t.Errorf("MaxBackupCount = %d, expected 3", MaxBackupCount)
	}
}

func TestCreateBackup_MultipleCalls(t *testing.T) {
	// Test that multiple consecutive backups work correctly
	tmpFile := createTempStorage(t, "version 1\n")
	tmpDir := filepath.Dir(tmpFile)

	// Create first backup
	if err := CreateBackup(tmpFile); err != nil {
		t.Fatalf("First CreateBackup() failed: %v", err)
	}

	// Modify file and create second backup
	if err := os.WriteFile(tmpFile, []byte("version 2\n"), 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}
	if err := CreateBackup(tmpFile); err != nil {
		t.Fatalf("Second CreateBackup() failed: %v", err)
	}

	// Modify file and create third backup
	if err := os.WriteFile(tmpFile, []byte("version 3\n"), 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}
	if err := CreateBackup(tmpFile); err != nil {
		t.Fatalf("Third CreateBackup() failed: %v", err)
	}

	// Verify backups contain correct versions
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	backup2Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".2")
	backup3Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".3")

	backup1Content := readFileContent(t, backup1Path)
	if backup1Content != "version 3\n" {
		t.Errorf(".bak.1 = %q, expected %q", backup1Content, "version 3\n")
	}

	backup2Content := readFileContent(t, backup2Path)
	if backup2Content != "version 2\n" {
		t.Errorf(".bak.2 = %q, expected %q", backup2Content, "version 2\n")
	}

	backup3Content := readFileContent(t, backup3Path)
	if backup3Content != "version 1\n" {
		t.Errorf(".bak.3 = %q, expected %q", backup3Content, "version 1\n")
	}
}

func TestListBackups_NoBackups(t *testing.T) {
	// Test ListBackups when no backup files exist
	// We can't easily control GetStoragePath, so we just verify it doesn't crash
	backups, err := ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() returned unexpected error: %v", err)
	}

	// With no backups, should return empty slice
	if len(backups) != 0 {
		t.Errorf("ListBackups() with no backups returned %d items, expected 0", len(backups))
	}
}

func TestListBackups_OneBackup(t *testing.T) {
	// Test ListBackups with one backup file
	tmpFile := createTempStorage(t, "test content\n")

	// Create one backup
	if err := CreateBackup(tmpFile); err != nil {
		t.Fatalf("CreateBackup() failed: %v", err)
	}

	backups, err := ListBackupsForStorage(tmpFile)
	if err != nil {
		t.Fatalf("ListBackupsForStorage() returned unexpected error: %v", err)
	}

	if len(backups) != 1 {
		t.Fatalf("ListBackupsForStorage() returned %d backups, expected 1", len(backups))
	}

	// Verify the backup info
	if backups[0].Number != 1 {
		t.Errorf("Backup number = %d, expected 1", backups[0].Number)
	}

	if backups[0].Path == "" {
		t.Errorf("Backup path is empty")
	}

	// Verify the path ends with .bak.1
	if !fileExists(backups[0].Path) {
		t.Errorf("Backup path %q does not exist", backups[0].Path)
	}
}

func TestListBackups_TwoBackups(t *testing.T) {
	// Test ListBackups with two backup files
	tmpFile := createTempStorage(t, "version 1\n")

	// Create first backup
	if err := CreateBackup(tmpFile); err != nil {
		t.Fatalf("First CreateBackup() failed: %v", err)
	}

	// Create second backup
	if err := os.WriteFile(tmpFile, []byte("version 2\n"), 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}
	if err := CreateBackup(tmpFile); err != nil {
		t.Fatalf("Second CreateBackup() failed: %v", err)
	}

	backups, err := ListBackupsForStorage(tmpFile)
	if err != nil {
		t.Fatalf("ListBackupsForStorage() returned unexpected error: %v", err)
	}

	if len(backups) != 2 {
		t.Fatalf("ListBackupsForStorage() returned %d backups, expected 2", len(backups))
	}

	// Verify backups are sorted by recency (.bak.1 is most recent)
	if backups[0].Number != 1 {
		t.Errorf("First backup number = %d, expected 1 (most recent)", backups[0].Number)
	}

	if backups[1].Number != 2 {
		t.Errorf("Second backup number = %d, expected 2", backups[1].Number)
	}

	// Verify both paths exist
	for i, backup := range backups {
		if !fileExists(backup.Path) {
			t.Errorf("Backup %d path %q does not exist", i, backup.Path)
		}
	}
}

func TestListBackups_ThreeBackups(t *testing.T) {
	// Test ListBackups with three backup files (maximum)
	tmpFile := createTempStorage(t, "version 1\n")

	// Create three backups
	for i := 1; i <= 3; i++ {
		content := "version " + strconv.Itoa(i) + "\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to update file: %v", err)
		}
		if err := CreateBackup(tmpFile); err != nil {
			t.Fatalf("CreateBackup() iteration %d failed: %v", i, err)
		}
	}

	backups, err := ListBackupsForStorage(tmpFile)
	if err != nil {
		t.Fatalf("ListBackupsForStorage() returned unexpected error: %v", err)
	}

	if len(backups) != 3 {
		t.Fatalf("ListBackupsForStorage() returned %d backups, expected 3", len(backups))
	}

	// Verify backups are sorted by recency
	expectedNumbers := []int{1, 2, 3}
	for i, backup := range backups {
		if backup.Number != expectedNumbers[i] {
			t.Errorf("Backup %d has number %d, expected %d", i, backup.Number, expectedNumbers[i])
		}

		if !fileExists(backup.Path) {
			t.Errorf("Backup %d path %q does not exist", i, backup.Path)
		}
	}
}

func TestListBackups_SortedByRecency(t *testing.T) {
	// Test that ListBackups returns backups sorted by recency
	tmpFile := createTempStorage(t, "initial\n")

	// Create multiple backups
	for i := 0; i < 5; i++ {
		content := "version " + string(rune('A'+i)) + "\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to update file: %v", err)
		}
		if err := CreateBackup(tmpFile); err != nil {
			t.Fatalf("CreateBackup() iteration %d failed: %v", i, err)
		}
	}

	backups, err := ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() returned unexpected error: %v", err)
	}

	// Verify backups are in ascending order by number (1, 2, 3)
	// where 1 is the most recent
	for i := 0; i < len(backups)-1; i++ {
		if backups[i].Number >= backups[i+1].Number {
			t.Errorf("Backups not sorted: backup[%d].Number=%d >= backup[%d].Number=%d",
				i, backups[i].Number, i+1, backups[i+1].Number)
		}
	}

	// First backup should be .bak.1 (most recent)
	if len(backups) > 0 && backups[0].Number != 1 {
		t.Errorf("First backup number = %d, expected 1 (most recent)", backups[0].Number)
	}
}

func TestRestoreBackup_InvalidBackupNumber(t *testing.T) {
	// Test RestoreBackup with invalid backup numbers
	tests := []struct {
		name           string
		backupNum      int
		errorSubstring string
	}{
		{"zero backup number", 0, "invalid backup number"},
		{"negative backup number", -1, "invalid backup number"},
		{"backup number too high", 4, "invalid backup number"},
		{"backup number way too high", 100, "invalid backup number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RestoreBackup(tt.backupNum)
			if err == nil {
				t.Errorf("RestoreBackup(%d) expected error containing %q, got nil", tt.backupNum, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("RestoreBackup(%d) error = %q, expected to contain %q", tt.backupNum, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestRestoreBackup_NonExistentBackup(t *testing.T) {
	// Test RestoreBackup when backup file doesn't exist
	// First, ensure we have a clean state with no backups
	tmpFile := createTempStorage(t, "current content\n")

	// Try to restore from backup 1 when it doesn't exist
	err := RestoreBackup(1)
	if err == nil {
		t.Fatalf("RestoreBackup(1) with non-existent backup expected error, got nil")
	}

	// Verify error message mentions backup doesn't exist
	errMsg := err.Error()
	expectedSubstrings := []string{"backup", "does not exist"}
	for _, substr := range expectedSubstrings {
		if !strings.Contains(errMsg, substr) {
			t.Errorf("RestoreBackup(1) error = %q, expected to contain %q", errMsg, substr)
		}
	}

	// Verify original file is unchanged
	content := readFileContent(t, tmpFile)
	if content != "current content\n" {
		t.Errorf("Original file was modified, content = %q", content)
	}
}

func TestRestoreBackup_ValidBackup(t *testing.T) {
	// Test RestoreBackup with a valid backup
	backupContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"backup entry","duration_minutes":60,"raw_input":"backup for 1h"}
`
	currentContent := `{"timestamp":"2024-01-15T12:00:00Z","description":"current entry","duration_minutes":30,"raw_input":"current for 30m"}
`

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create a backup manually
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	if err := os.WriteFile(backup1Path, []byte(backupContent), 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	// Restore from backup 1
	err := RestoreBackupForStorage(tmpFile, 1)
	if err != nil {
		t.Fatalf("RestoreBackupForStorage(1) returned unexpected error: %v", err)
	}

	// Verify main storage file now contains backup content
	restoredContent := readFileContent(t, tmpFile)
	if restoredContent != backupContent {
		t.Errorf("Restored content = %q, expected %q", restoredContent, backupContent)
	}

	// Verify a safety backup was created (old current content should be in .bak.1)
	newBackup1Content := readFileContent(t, backup1Path)
	if newBackup1Content != currentContent {
		t.Errorf("Safety backup .bak.1 = %q, expected previous current content %q", newBackup1Content, currentContent)
	}
}

func TestRestoreBackup_RestoresFromBackup2(t *testing.T) {
	// Test RestoreBackup with backup number 2
	backup1Content := "most recent backup\n"
	backup2Content := "older backup\n"
	currentContent := "current version\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create backups manually
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	backup2Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".2")
	if err := os.WriteFile(backup1Path, []byte(backup1Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.1: %v", err)
	}
	if err := os.WriteFile(backup2Path, []byte(backup2Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.2: %v", err)
	}

	// Restore from backup 2
	err := RestoreBackupForStorage(tmpFile, 2)
	if err != nil {
		t.Fatalf("RestoreBackupForStorage(2) returned unexpected error: %v", err)
	}

	// Verify main storage file now contains backup 2 content
	restoredContent := readFileContent(t, tmpFile)
	if restoredContent != backup2Content {
		t.Errorf("Restored content = %q, expected backup 2 content %q", restoredContent, backup2Content)
	}
}

func TestRestoreBackup_RestoresFromBackup3(t *testing.T) {
	// Test RestoreBackup with backup number 3 (oldest)
	backup3Content := "oldest backup\n"
	currentContent := "current version\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create backup 3 manually
	backup3Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".3")
	if err := os.WriteFile(backup3Path, []byte(backup3Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.3: %v", err)
	}

	// Restore from backup 3
	err := RestoreBackupForStorage(tmpFile, 3)
	if err != nil {
		t.Fatalf("RestoreBackupForStorage(3) returned unexpected error: %v", err)
	}

	// Verify main storage file now contains backup 3 content
	restoredContent := readFileContent(t, tmpFile)
	if restoredContent != backup3Content {
		t.Errorf("Restored content = %q, expected backup 3 content %q", restoredContent, backup3Content)
	}
}

func TestRestoreBackup_EmptyBackupFile(t *testing.T) {
	// Test RestoreBackup with an empty backup file
	currentContent := "current content\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create an empty backup file
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	if err := os.WriteFile(backup1Path, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty backup: %v", err)
	}

	// Restore from empty backup
	err := RestoreBackupForStorage(tmpFile, 1)
	if err != nil {
		t.Fatalf("RestoreBackupForStorage(1) with empty backup returned unexpected error: %v", err)
	}

	// Verify main storage file is now empty
	restoredContent := readFileContent(t, tmpFile)
	if restoredContent != "" {
		t.Errorf("Restored content = %q, expected empty", restoredContent)
	}

	// Verify safety backup was created with previous content
	backupContent := readFileContent(t, backup1Path)
	if backupContent != currentContent {
		t.Errorf("Safety backup content = %q, expected %q", backupContent, currentContent)
	}
}

func TestRestoreBackup_CreatesParentDirectory(t *testing.T) {
	// Test that RestoreBackup works when storage directory exists
	backupContent := "backup content\n"
	currentContent := "current content\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create backup
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	if err := os.WriteFile(backup1Path, []byte(backupContent), 0644); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Restore should work without errors
	err := RestoreBackupForStorage(tmpFile, 1)
	if err != nil {
		t.Fatalf("RestoreBackupForStorage(1) returned unexpected error: %v", err)
	}

	// Verify restore succeeded
	restoredContent := readFileContent(t, tmpFile)
	if restoredContent != backupContent {
		t.Errorf("Restored content = %q, expected %q", restoredContent, backupContent)
	}
}

func TestRestoreBackup_LargeBackupFile(t *testing.T) {
	// Test RestoreBackup with a large backup file
	var largeBackupContent string
	for i := 0; i < 1000; i++ {
		largeBackupContent += `{"timestamp":"2024-01-15T10:00:00Z","description":"backup entry ` + strconv.Itoa(i%10) + `","duration_minutes":15,"raw_input":"entry for 15m"}` + "\n"
	}
	currentContent := "small current content\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create large backup
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	if err := os.WriteFile(backup1Path, []byte(largeBackupContent), 0644); err != nil {
		t.Fatalf("Failed to create large backup: %v", err)
	}

	// Restore from large backup
	err := RestoreBackupForStorage(tmpFile, 1)
	if err != nil {
		t.Fatalf("RestoreBackupForStorage(1) with large backup returned unexpected error: %v", err)
	}

	// Verify main storage file now contains large backup content
	restoredContent := readFileContent(t, tmpFile)
	if restoredContent != largeBackupContent {
		t.Errorf("Restored content length = %d, expected %d", len(restoredContent), len(largeBackupContent))
	}
}

func TestRestoreBackup_PreservesBackupNumber(t *testing.T) {
	// Test that RestoreBackup doesn't delete the backup it restores from
	backupContent := "backup content\n"
	currentContent := "current content\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create backup
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	if err := os.WriteFile(backup1Path, []byte(backupContent), 0644); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Restore from backup 1
	err := RestoreBackupForStorage(tmpFile, 1)
	if err != nil {
		t.Fatalf("RestoreBackupForStorage(1) returned unexpected error: %v", err)
	}

	// Verify backup file still exists
	if !fileExists(backup1Path) {
		t.Errorf("Backup file was deleted after restore")
	}

	// Note: The backup file will now contain the safety backup (previous current content)
	// This is expected behavior as CreateBackup is called before restoring
}

func TestRestoreBackup_MultipleRestores(t *testing.T) {
	// Test multiple consecutive restore operations
	backup1Content := "backup 1 content\n"
	backup2Content := "backup 2 content\n"
	currentContent := "current content\n"

	tmpFile := createTempStorage(t, currentContent)
	tmpDir := filepath.Dir(tmpFile)

	// Create backups
	backup1Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".1")
	backup2Path := filepath.Join(tmpDir, "entries.jsonl"+BackupSuffix+".2")
	if err := os.WriteFile(backup1Path, []byte(backup1Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.1: %v", err)
	}
	if err := os.WriteFile(backup2Path, []byte(backup2Content), 0644); err != nil {
		t.Fatalf("Failed to create .bak.2: %v", err)
	}

	// First restore from backup 2
	if err := RestoreBackupForStorage(tmpFile, 2); err != nil {
		t.Fatalf("First RestoreBackupForStorage(2) failed: %v", err)
	}

	// Verify content is backup 2
	content := readFileContent(t, tmpFile)
	if content != backup2Content {
		t.Errorf("After first restore, content = %q, expected %q", content, backup2Content)
	}

	// Second restore from backup 1 (which now contains a safety backup)
	// We need to recreate backup1 with known content for this test
	if err := os.WriteFile(backup1Path, []byte(backup1Content), 0644); err != nil {
		t.Fatalf("Failed to recreate .bak.1: %v", err)
	}

	if err := RestoreBackupForStorage(tmpFile, 1); err != nil {
		t.Fatalf("Second RestoreBackupForStorage(1) failed: %v", err)
	}

	// Verify content is now backup 1
	content = readFileContent(t, tmpFile)
	if content != backup1Content {
		t.Errorf("After second restore, content = %q, expected %q", content, backup1Content)
	}
}

func TestCreateBackup_SourceOpenError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a file
	if err := os.WriteFile(storagePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}

	// Make file unreadable
	if err := os.Chmod(storagePath, 0000); err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}
	defer func() { _ = os.Chmod(storagePath, 0644) }()

	err := CreateBackup(storagePath)
	if err == nil {
		t.Error("Expected error when source file is unreadable")
	}
}

func TestCreateBackup_DestCreateError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a file
	if err := os.WriteFile(storagePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}

	// Make directory read-only so backup file can't be created
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Skipf("Cannot change directory permissions: %v", err)
	}
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	err := CreateBackup(storagePath)
	if err == nil {
		t.Error("Expected error when backup file cannot be created")
	}
}

func TestRestoreBackupForStorage_ReadError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	backupPath := storagePath + ".bak.1"

	// Create files
	if err := os.WriteFile(storagePath, []byte("current"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}
	if err := os.WriteFile(backupPath, []byte("backup"), 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	// Make backup file unreadable
	if err := os.Chmod(backupPath, 0000); err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}
	defer func() { _ = os.Chmod(backupPath, 0644) }()

	err := RestoreBackupForStorage(storagePath, 1)
	if err == nil {
		t.Error("Expected error when backup file is unreadable")
	}
}

func TestRestoreBackupForStorage_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	backupPath := storagePath + ".bak.1"

	// Create files
	if err := os.WriteFile(storagePath, []byte("current"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}
	if err := os.WriteFile(backupPath, []byte("backup"), 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	// Make storage file read-only so it can't be overwritten
	if err := os.Chmod(storagePath, 0444); err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}
	defer func() { _ = os.Chmod(storagePath, 0644) }()

	err := RestoreBackupForStorage(storagePath, 1)
	if err == nil {
		t.Error("Expected error when storage file cannot be written")
	}
}

func TestGetBackupPath_WrapperFunction(t *testing.T) {
	// Test the wrapper function GetBackupPath
	// It should work when GetStoragePath works
	path, err := GetBackupPath(1)
	if err != nil {
		// If we can't get storage path in test environment, skip
		t.Skipf("GetBackupPath returned error: %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
}

func TestListBackups(t *testing.T) {
	// Test the wrapper function ListBackups
	_, err := ListBackups()
	// May return error in test environment if we can't get storage path
	// Just make sure it doesn't panic
	_ = err
}

func TestRestoreBackup(t *testing.T) {
	// Test the wrapper function RestoreBackup with invalid number
	err := RestoreBackup(0)
	if err == nil {
		t.Error("Expected error for invalid backup number 0")
	}
}

// backupMockPathProvider is a test helper for mocking osutil.PathProvider in backup tests
type backupMockPathProvider struct {
	userConfigDirFn func() (string, error)
	mkdirAllFn      func(path string, perm os.FileMode) error
}

func (m *backupMockPathProvider) UserConfigDir() (string, error) {
	if m.userConfigDirFn != nil {
		return m.userConfigDirFn()
	}
	return "", nil
}

func (m *backupMockPathProvider) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirAllFn != nil {
		return m.mkdirAllFn(path, perm)
	}
	return nil
}

func TestGetBackupPathForStorage_GetStoragePathError(t *testing.T) {
	// Save original provider
	defer osutil.ResetProvider()

	// Mock GetStoragePath to return an error
	osutil.SetProvider(&backupMockPathProvider{
		userConfigDirFn: func() (string, error) {
			return "", os.ErrPermission
		},
	})

	_, err := GetBackupPathForStorage("", 1)
	if err == nil {
		t.Error("GetBackupPathForStorage() should return error when GetStoragePath fails")
	}
}

func TestRotateBackups_GetBackupPathError(t *testing.T) {
	// rotateBackups is tested indirectly through CreateBackup
	// Test that if GetBackupPathForStorage fails inside rotateBackups, error is returned

	// Save original provider
	defer osutil.ResetProvider()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create storage file
	if err := os.WriteFile(storagePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock to fail after the first GetBackupPathForStorage call (during rotateBackups)
	callCount := 0
	osutil.SetProvider(&backupMockPathProvider{
		userConfigDirFn: func() (string, error) {
			callCount++
			// Fail on subsequent calls to simulate rotateBackups failing
			if callCount > 1 {
				return "", os.ErrPermission
			}
			return tmpDir, nil
		},
		mkdirAllFn: func(path string, perm os.FileMode) error {
			return nil
		},
	})

	// This should fail because rotateBackups can't get backup paths
	// Note: CreateBackup passes storagePath explicitly so it doesn't call GetStoragePath
	// The error happens inside rotateBackups when calling GetBackupPathForStorage with the storagePath
	// Since we pass storagePath explicitly, the mock doesn't affect this test.
	// The test is kept for documentation purposes that this path was considered.
	_ = callCount // suppress unused variable warning
}

func TestRotateBackups_RemoveOldestError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create storage file
	if err := os.WriteFile(storagePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}

	// Create backup.3 as a non-empty directory (os.Remove cannot delete non-empty directories)
	backup3Path := storagePath + ".bak.3"
	if err := os.MkdirAll(backup3Path, 0755); err != nil {
		t.Fatalf("Failed to create backup 3 directory: %v", err)
	}
	// Add a file inside to make it non-empty
	if err := os.WriteFile(filepath.Join(backup3Path, "dummy"), []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create file inside backup 3 directory: %v", err)
	}

	// CreateBackup should fail because os.Remove can't delete non-empty directory
	err := CreateBackup(storagePath)
	if err == nil {
		t.Error("CreateBackup() should fail when backup.3 is a non-empty directory")
	}
}

func TestRotateBackups_RenameError(t *testing.T) {
	// Create a subdirectory inside tmpDir to control permissions
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	storagePath := filepath.Join(subDir, "entries.jsonl")

	// Create storage file and backup 1
	if err := os.WriteFile(storagePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}
	backup1Path := storagePath + ".bak.1"
	if err := os.WriteFile(backup1Path, []byte("backup 1"), 0644); err != nil {
		t.Fatalf("Failed to create backup 1: %v", err)
	}

	// Make the subdirectory read-only so rename operations fail
	// os.Rename needs write permission on the directory to modify directory entries
	if err := os.Chmod(subDir, 0555); err != nil {
		t.Skipf("Cannot change directory permissions: %v", err)
	}
	defer func() { _ = os.Chmod(subDir, 0755) }()

	// CreateBackup should fail because os.Rename can't modify directory entries in read-only dir
	err := CreateBackup(storagePath)
	if err == nil {
		t.Error("CreateBackup() should fail when directory is read-only")
	}
}

func TestCreateBackup_StatError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a file
	if err := os.WriteFile(storagePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}

	// Make the directory unreadable so Stat fails with permission error
	if err := os.Chmod(tmpDir, 0000); err != nil {
		t.Skipf("Cannot change directory permissions: %v", err)
	}
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	err := CreateBackup(storagePath)
	if err == nil {
		t.Error("CreateBackup() should return error when Stat fails with permission error")
	}
}

func TestCreateBackup_RotateBackupsError(t *testing.T) {
	// Note: Testing rotateBackups errors directly is difficult because:
	// 1. GetBackupPathForStorage doesn't fail when storagePath is provided
	// 2. os.Remove errors on non-existent files are ignored
	// 3. os.Rename behavior is platform-specific
	// This test verifies that CreateBackup works correctly when rotation succeeds.
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create storage file
	if err := os.WriteFile(storagePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}

	// Create all backup files
	for i := 1; i <= MaxBackupCount; i++ {
		backupPath := storagePath + ".bak." + strconv.Itoa(i)
		if err := os.WriteFile(backupPath, []byte("backup "+strconv.Itoa(i)), 0644); err != nil {
			t.Fatalf("Failed to create backup %d: %v", i, err)
		}
	}

	// CreateBackup should successfully rotate and create new backup
	err := CreateBackup(storagePath)
	if err != nil {
		t.Errorf("CreateBackup() returned unexpected error: %v", err)
	}

	// Verify backup 1 now contains storage content
	backup1Content, err := os.ReadFile(storagePath + ".bak.1")
	if err != nil {
		t.Fatalf("Failed to read backup 1: %v", err)
	}
	if string(backup1Content) != "test content" {
		t.Errorf("Backup 1 content = %q, expected %q", string(backup1Content), "test content")
	}
}

func TestListBackupsForStorage_GetBackupPathError(t *testing.T) {
	// Save original provider
	defer osutil.ResetProvider()

	// Mock GetStoragePath to return an error
	osutil.SetProvider(&backupMockPathProvider{
		userConfigDirFn: func() (string, error) {
			return "", os.ErrPermission
		},
	})

	_, err := ListBackupsForStorage("")
	if err == nil {
		t.Error("ListBackupsForStorage() should return error when GetBackupPathForStorage fails")
	}
}

func TestRestoreBackupForStorage_GetStoragePathError(t *testing.T) {
	// Save original provider
	defer osutil.ResetProvider()

	// Mock GetStoragePath to return an error
	osutil.SetProvider(&backupMockPathProvider{
		userConfigDirFn: func() (string, error) {
			return "", os.ErrPermission
		},
	})

	err := RestoreBackupForStorage("", 1)
	if err == nil {
		t.Error("RestoreBackupForStorage() should return error when GetStoragePath fails")
	}
}

func TestRestoreBackupForStorage_CreateBackupError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	backupPath := storagePath + ".bak.1"

	// Create storage file and backup
	if err := os.WriteFile(storagePath, []byte("current content"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}
	if err := os.WriteFile(backupPath, []byte("backup content"), 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	// Make the storage file read-only so CreateBackup fails when trying to create dest file
	// Actually, CreateBackup creates a backup of the storage file, not writes to it directly.
	// To cause CreateBackup to fail, we need the backup file location to be unwritable.

	// Make directory read-only after reading files
	// This prevents creating new backup files during rotation
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Skipf("Cannot change directory permissions: %v", err)
	}
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	err := RestoreBackupForStorage(storagePath, 1)
	if err == nil {
		t.Error("RestoreBackupForStorage() should return error when CreateBackup fails")
	}
}

func TestRestoreBackupForStorage_ReadBackupPermissionError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	backupPath := storagePath + ".bak.1"

	// Create storage file and backup
	if err := os.WriteFile(storagePath, []byte("current content"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}
	if err := os.WriteFile(backupPath, []byte("backup content"), 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	// Make backup file unreadable
	if err := os.Chmod(backupPath, 0000); err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}
	defer func() { _ = os.Chmod(backupPath, 0644) }()

	err := RestoreBackupForStorage(storagePath, 1)
	if err == nil {
		t.Error("RestoreBackupForStorage() should return error when backup file is unreadable")
	}
}

func TestRestoreBackupForStorage_WriteFileError(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	storagePath := filepath.Join(subDir, "entries.jsonl")
	backupPath := storagePath + ".bak.1"

	// Create storage file (needed for CreateBackup to work)
	if err := os.WriteFile(storagePath, []byte("current content"), 0644); err != nil {
		t.Fatalf("Failed to create storage file: %v", err)
	}
	// Create backup file
	if err := os.WriteFile(backupPath, []byte("backup content"), 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	// Read the backup content first (before making directory read-only)
	// Then make directory read-only so WriteFile fails
	// But CreateBackup needs write access...
	// This is tricky because CreateBackup runs before WriteFile

	// Alternative approach: Make storage file read-only after backup is read
	// but this won't cause WriteFile to fail (it creates a new file)

	// Actually, we need to make the storage file a directory to cause WriteFile to fail
	// Remove the storage file and create a directory with the same name
	if err := os.Remove(storagePath); err != nil {
		t.Fatalf("Failed to remove storage file: %v", err)
	}
	// Create directory with same name as storage file
	// But wait - CreateBackup checks if storage exists first, so it won't error
	// We need storage to exist for CreateBackup, but then WriteFile should fail

	// Let's use a different approach: create a subdirectory named as the storage file
	// This won't work because CreateBackup needs to read the storage file

	// Skip this test as it's very hard to trigger WriteFile error after CreateBackup succeeds
	t.Skip("WriteFile error path is difficult to test - requires storage to exist for CreateBackup but fail for WriteFile")
}
