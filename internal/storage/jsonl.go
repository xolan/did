package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xolan/did/internal/entry"
)

const (
	// AppName is the application name used for config directory
	AppName = "did"
	// EntriesFile is the name of the JSON Lines storage file
	EntriesFile = "entries.jsonl"
)

// ParseWarning represents a warning about a corrupted or malformed entry
type ParseWarning struct {
	LineNumber int    // Line number in the file (1-indexed)
	Content    string // Raw content of the corrupted line
	Error      string // Description of the parsing error
}

// ReadResult contains the results of reading entries from storage,
// including both successfully parsed entries and any warnings about
// corrupted or malformed lines.
type ReadResult struct {
	Entries  []entry.Entry  // Successfully parsed entries
	Warnings []ParseWarning // Warnings about corrupted lines
}

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
	defer func() { _ = file.Close() }()

	line, err := json.Marshal(e)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(line) + "\n")
	return err
}

// ReadEntriesWithWarnings reads all entries from the JSON Lines storage file
// and returns both successfully parsed entries and warnings about any corrupted lines.
// Returns an empty ReadResult if the file doesn't exist (graceful handling).
// Collects detailed warnings for each malformed line including line number, content, and error.
func ReadEntriesWithWarnings(filepath string) (ReadResult, error) {
	result := ReadResult{
		Entries:  []entry.Entry{},
		Warnings: []ParseWarning{},
	}

	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		lineContent := scanner.Text()

		var e entry.Entry
		if err := json.Unmarshal([]byte(lineContent), &e); err != nil {
			// Record warning for corrupted line
			result.Warnings = append(result.Warnings, ParseWarning{
				LineNumber: lineNumber,
				Content:    lineContent,
				Error:      err.Error(),
			})
			continue
		}
		result.Entries = append(result.Entries, e)
	}

	if err := scanner.Err(); err != nil {
		return result, err
	}

	return result, nil
}

// ReadEntries reads all entries from the JSON Lines storage file.
// Returns an empty slice if the file doesn't exist (graceful handling).
// Skips malformed lines for fault tolerance.
// This function is maintained for backward compatibility and internally calls ReadEntriesWithWarnings.
func ReadEntries(filepath string) ([]entry.Entry, error) {
	result, err := ReadEntriesWithWarnings(filepath)
	return result.Entries, err
}

// ReadActiveEntries reads all non-deleted entries from the JSON Lines storage file.
// Returns only entries where DeletedAt is nil.
// Returns an empty slice if the file doesn't exist (graceful handling).
// Skips malformed lines for fault tolerance.
func ReadActiveEntries(filepath string) ([]entry.Entry, error) {
	entries, err := ReadEntries(filepath)
	if err != nil {
		return nil, err
	}

	// Filter out deleted entries
	active := make([]entry.Entry, 0, len(entries))
	for _, e := range entries {
		if e.DeletedAt == nil {
			active = append(active, e)
		}
	}

	return active, nil
}

// WriteEntries writes all entries to the JSON Lines storage file.
// Overwrites the file if it exists. Creates the file with 0644 permissions.
// This is used for operations that modify existing entries (e.g., delete, update).
func WriteEntries(filepath string, entries []entry.Entry) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	for _, e := range entries {
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		if _, err := file.WriteString(string(line) + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// SoftDeleteEntry marks an entry as deleted by setting its DeletedAt timestamp.
// Index is 0-based. Returns an error if the index is out of bounds.
// The entry remains in the file but is marked as deleted.
func SoftDeleteEntry(filepath string, index int) (entry.Entry, error) {
	entries, err := ReadEntries(filepath)
	if err != nil {
		return entry.Entry{}, err
	}

	if index < 0 || index >= len(entries) {
		return entry.Entry{}, fmt.Errorf("index %d out of bounds (0-%d)", index, len(entries)-1)
	}

	// Set DeletedAt to current time
	now := time.Now()
	entries[index].DeletedAt = &now

	deleted := entries[index]

	if err := WriteEntries(filepath, entries); err != nil {
		return entry.Entry{}, err
	}

	return deleted, nil
}

// DeleteEntry deletes the entry at the given index and returns it.
// Index is 0-based. Returns an error if the index is out of bounds.
// Rewrites the entire file without the deleted entry.
func DeleteEntry(filepath string, index int) (entry.Entry, error) {
	entries, err := ReadEntries(filepath)
	if err != nil {
		return entry.Entry{}, err
	}

	if index < 0 || index >= len(entries) {
		return entry.Entry{}, fmt.Errorf("index %d out of bounds (0-%d)", index, len(entries)-1)
	}

	deleted := entries[index]

	// Remove the entry by creating a new slice without it
	newEntries := append(entries[:index], entries[index+1:]...)

	if err := WriteEntries(filepath, newEntries); err != nil {
		return entry.Entry{}, err
	}

	return deleted, nil
}

// UpdateEntry updates an entry at a specific index by rewriting the JSONL file.
// Uses 0-based indexing internally (caller handles 1-based conversion).
// Returns error if index is out of range.
// Uses atomic write pattern (write to temp file, then rename) for safety.
func UpdateEntry(filepath string, index int, e entry.Entry) error {
	// Read all entries
	entries, err := ReadEntries(filepath)
	if err != nil {
		return err
	}

	// Validate index
	if index < 0 || index >= len(entries) {
		return os.ErrInvalid
	}

	// Update the entry at the specified index
	entries[index] = e

	// Write to temporary file
	tmpFile := filepath + ".tmp"
	file, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// Write all entries to temp file
	for _, entry := range entries {
		line, err := json.Marshal(entry)
		if err != nil {
			_ = file.Close()
			_ = os.Remove(tmpFile)
			return err
		}
		if _, err := file.WriteString(string(line) + "\n"); err != nil {
			_ = file.Close()
			_ = os.Remove(tmpFile)
			return err
		}
	}

	// Close temp file before rename
	if err := file.Close(); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}

	// Atomic rename
	return os.Rename(tmpFile, filepath)
}

// StorageHealth contains information about the health status of the storage file.
// It provides metrics on total lines, valid entries, corrupted entries, and detailed
// warnings about each corruption.
type StorageHealth struct {
	TotalLines       int            // Total number of lines in the storage file
	ValidEntries     int            // Number of successfully parsed entries
	CorruptedEntries int            // Number of corrupted/malformed lines
	Warnings         []ParseWarning // Detailed information about each corrupted line
}

// ValidateStorage analyzes the storage file and returns health status information.
// Returns metrics on total lines, valid entries, corrupted entries, and details
// about each corruption. Returns empty health status if file doesn't exist.
func ValidateStorage(filepath string) (StorageHealth, error) {
	health := StorageHealth{
		TotalLines:       0,
		ValidEntries:     0,
		CorruptedEntries: 0,
		Warnings:         []ParseWarning{},
	}

	// Check if file exists
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return health, nil
		}
		return health, err
	}
	defer func() { _ = file.Close() }()

	// Count total lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		health.TotalLines++
	}

	if err := scanner.Err(); err != nil {
		return health, err
	}

	// Get entries and warnings
	result, err := ReadEntriesWithWarnings(filepath)
	if err != nil {
		return health, err
	}

	health.ValidEntries = len(result.Entries)
	health.CorruptedEntries = len(result.Warnings)
	health.Warnings = result.Warnings

	return health, nil
}
