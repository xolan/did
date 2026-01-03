package storage

import (
	"fmt"
	"os"
)

const (
	// BackupSuffix is the file extension for backup files
	BackupSuffix = ".bak"
	// MaxBackupCount is the maximum number of backup files to keep
	MaxBackupCount = 3
)

// GetBackupPath returns the path to a backup file with the given rotation number.
// The rotation number n should be between 1 and MaxBackupCount (inclusive).
// Backup files are named with the format: entries.jsonl.bak.N where N is the rotation number.
// Lower numbers are more recent (e.g., .bak.1 is the most recent backup).
// If storagePath is empty, uses GetStoragePath() to get the default storage location.
func GetBackupPath(n int) (string, error) {
	return GetBackupPathForStorage("", n)
}

// GetBackupPathForStorage returns the backup path for a specific storage file.
// If storagePath is empty, uses GetStoragePath() to get the default storage location.
func GetBackupPathForStorage(storagePath string, n int) (string, error) {
	if storagePath == "" {
		var err error
		storagePath, err = GetStoragePath()
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s%s.%d", storagePath, BackupSuffix, n), nil
}

// rotateBackups shifts existing backup files to make room for a new backup.
// It renames .bak.1 -> .bak.2, .bak.2 -> .bak.3, and deletes the oldest .bak.3
// if it exists. This ensures only MaxBackupCount backups are kept.
// Returns an error if any file operation fails (except for missing files, which are OK).
func rotateBackups(storagePath string) error {
	// Delete the oldest backup (.bak.3) if it exists to make room
	oldestPath, err := GetBackupPathForStorage(storagePath, MaxBackupCount)
	if err != nil {
		return err
	}
	if err := os.Remove(oldestPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Rotate backups from MaxBackupCount-1 down to 1
	// This shifts .bak.2 -> .bak.3, then .bak.1 -> .bak.2
	for i := MaxBackupCount - 1; i >= 1; i-- {
		currentPath, err := GetBackupPathForStorage(storagePath, i)
		if err != nil {
			return err
		}

		nextPath, err := GetBackupPathForStorage(storagePath, i+1)
		if err != nil {
			return err
		}

		// Rename the file if it exists
		if err := os.Rename(currentPath, nextPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

// CreateBackup creates a backup of the storage file before making destructive modifications.
// It rotates existing backups and copies the current storage file to .bak.1.
// If the storage file doesn't exist, no backup is created and no error is returned.
// Returns an error if backup rotation or file copying fails.
func CreateBackup(storagePath string) error {
	// Check if storage file exists
	if _, err := os.Stat(storagePath); err != nil {
		if os.IsNotExist(err) {
			// No file to backup, return without error
			return nil
		}
		return err
	}

	// Rotate existing backups to make room for new backup
	if err := rotateBackups(storagePath); err != nil {
		return err
	}

	// Get path for the new backup (.bak.1)
	backupPath, err := GetBackupPathForStorage(storagePath, 1)
	if err != nil {
		return err
	}

	// Copy current storage file to .bak.1
	sourceFile, err := os.Open(storagePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the file contents
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}

	return nil
}

// BackupInfo contains information about a backup file
type BackupInfo struct {
	Number int    // The backup number (1, 2, or 3)
	Path   string // The full path to the backup file
}

// ListBackups returns available backup files sorted by recency.
// .bak.1 is the most recent backup, .bak.2 is older, .bak.3 is oldest.
// Returns an empty slice if no backups exist.
func ListBackups() ([]BackupInfo, error) {
	var backups []BackupInfo

	// Check each backup number from 1 to MaxBackupCount
	for i := 1; i <= MaxBackupCount; i++ {
		backupPath, err := GetBackupPath(i)
		if err != nil {
			return nil, err
		}

		// Check if the backup file exists
		if _, err := os.Stat(backupPath); err == nil {
			backups = append(backups, BackupInfo{
				Number: i,
				Path:   backupPath,
			})
		}
	}

	return backups, nil
}

// RestoreBackup restores a backup file to the main storage file.
// backupNum specifies which backup to restore (1 is most recent, 3 is oldest).
// Creates a backup of the current state before restoring for safety.
// Returns an error if the backup number is invalid or the backup file doesn't exist.
func RestoreBackup(backupNum int) error {
	// Validate backup number
	if backupNum < 1 || backupNum > MaxBackupCount {
		return fmt.Errorf("invalid backup number %d, must be between 1 and %d", backupNum, MaxBackupCount)
	}

	// Get the backup path
	backupPath, err := GetBackupPath(backupNum)
	if err != nil {
		return err
	}

	// Check if backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup %d does not exist", backupNum)
		}
		return err
	}

	// Get the main storage path
	storagePath, err := GetStoragePath()
	if err != nil {
		return err
	}

	// Create a backup of the current state before restoring (safety measure)
	if err := CreateBackup(storagePath); err != nil {
		return err
	}

	// Copy the backup file to the main storage file
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(storagePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the file contents
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}

	return nil
}
