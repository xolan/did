package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/xolan/did/internal/entry"
)

func writeEntriesToFile(file *os.File, entries []entry.Entry) error {
	for _, e := range entries {
		line, _ := json.Marshal(e)
		if _, err := file.WriteString(string(line) + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func writeEntriesToTempFile(file *os.File, tmpFile string, entries []entry.Entry) error {
	for _, e := range entries {
		line, _ := json.Marshal(e)
		if _, err := file.WriteString(string(line) + "\n"); err != nil {
			_ = file.Close()
			_ = os.Remove(tmpFile)
			return err
		}
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	return nil
}

func validateStorageScanAndRead(file *os.File, filepath string, health *StorageHealth) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		health.TotalLines++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	result, err := ReadEntriesWithWarnings(filepath)
	if err != nil {
		return err
	}

	health.ValidEntries = len(result.Entries)
	health.CorruptedEntries = len(result.Warnings)
	health.Warnings = result.Warnings
	return nil
}

func getBackupPathWithError(storagePath string, n int) (string, error) {
	return GetBackupPathForStorage(storagePath, n)
}
