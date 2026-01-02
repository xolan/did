package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/xolan/did/internal/entry"
)

const (
	// AppName is the application name used for config directory
	AppName = "did"
	// EntriesFile is the name of the JSON Lines storage file
	EntriesFile = "entries.jsonl"
)

// GetStoragePath returns the path to the entries storage file.
// Uses os.UserConfigDir() for cross-platform XDG-compliant config directory.
// Creates the config directory if it doesn't exist.
func GetStoragePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, AppName)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, EntriesFile), nil
}

// AppendEntry appends a single entry to the JSON Lines storage file.
// Creates the file if it doesn't exist.
// Uses O_APPEND for atomic append operations.
func AppendEntry(filepath string, e entry.Entry) error {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	line, err := json.Marshal(e)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(line) + "\n")
	return err
}

// ReadEntries reads all entries from the JSON Lines storage file.
// Returns an empty slice if the file doesn't exist (graceful handling).
// Skips malformed lines for fault tolerance.
func ReadEntries(filepath string) ([]entry.Entry, error) {
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return []entry.Entry{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var entries []entry.Entry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e entry.Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue // Skip malformed lines
		}
		entries = append(entries, e)
	}

	return entries, scanner.Err()
}
