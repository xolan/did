package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
)

// Helper to create a temporary test file with entries
func createTempEntriesFile(t *testing.T, entries []entry.Entry) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_entries.jsonl")

	if len(entries) > 0 {
		for _, e := range entries {
			if err := storage.AppendEntry(tmpFile, e); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
		}
	}

	return tmpFile
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		{"30 minutes", 30, "30m"},
		{"1 hour", 60, "1h"},
		{"1 hour 30 minutes", 90, "1h 30m"},
		{"2 hours", 120, "2h"},
		{"2 hours 15 minutes", 135, "2h 15m"},
		{"3 hours 45 minutes", 225, "3h 45m"},
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

func TestShowEntryForDeletion(t *testing.T) {
	tests := []struct {
		name            string
		entry           entry.Entry
		expectedOutputs []string
	}{
		{
			name: "entry with hour duration",
			entry: entry.Entry{
				Timestamp:       time.Date(2024, time.January, 15, 10, 30, 0, 0, time.Local),
				Description:     "work on feature X",
				DurationMinutes: 120,
				RawInput:        "work on feature X for 2h",
			},
			expectedOutputs: []string{
				"Entry to delete:",
				"2024-01-15 10:30",
				"work on feature X",
				"(2h)",
			},
		},
		{
			name: "entry with mixed duration",
			entry: entry.Entry{
				Timestamp:       time.Date(2024, time.June, 20, 14, 15, 0, 0, time.Local),
				Description:     "code review",
				DurationMinutes: 45,
				RawInput:        "code review for 45m",
			},
			expectedOutputs: []string{
				"Entry to delete:",
				"2024-06-20 14:15",
				"code review",
				"(45m)",
			},
		},
		{
			name: "entry with special characters",
			entry: entry.Entry{
				Timestamp:       time.Date(2024, time.March, 10, 9, 0, 0, 0, time.Local),
				Description:     "fix bug #123 \"critical\" issue",
				DurationMinutes: 90,
				RawInput:        "fix bug #123 \"critical\" issue for 1h 30m",
			},
			expectedOutputs: []string{
				"Entry to delete:",
				"2024-03-10 09:00",
				"fix bug #123 \"critical\" issue",
				"(1h 30m)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			showEntryForDeletion(tt.entry)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			// Verify all expected strings are in output
			for _, expected := range tt.expectedOutputs {
				if !strings.Contains(output, expected) {
					t.Errorf("Output missing expected string %q\nGot: %s", expected, output)
				}
			}
		})
	}
}

func TestPromptConfirmation_Yes(t *testing.T) {
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
		{"yes spelled out", "yes\n", false},
		{"no spelled out", "no\n", false},
		{"y with spaces", "  y  \n", true},
		{"Y with spaces", "  Y  \n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace stdin with test input
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write test input
			_, _ = w.Write([]byte(tt.input))
			w.Close()

			// Capture stdout (the prompt)
			oldStdout := os.Stdout
			stdoutR, stdoutW, _ := os.Pipe()
			os.Stdout = stdoutW

			result := promptConfirmation()

			stdoutW.Close()
			os.Stdout = oldStdout
			os.Stdin = oldStdin

			// Read and verify prompt was displayed
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, stdoutR)
			prompt := buf.String()

			if !strings.Contains(prompt, "Delete this entry? [y/N]:") {
				t.Errorf("Expected prompt to contain 'Delete this entry? [y/N]:', got: %s", prompt)
			}

			if result != tt.expected {
				t.Errorf("promptConfirmation() with input %q = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPromptConfirmation_EmptyReader(t *testing.T) {
	// Test behavior when stdin is closed/empty
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Close() // Close immediately to simulate EOF

	// Capture stdout
	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	result := promptConfirmation()

	stdoutW.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	// Drain stdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	if result != false {
		t.Errorf("promptConfirmation() with closed stdin = %v, expected false", result)
	}
}

// Integration-style tests that verify the delete logic flow
// These test the deleteEntry function indirectly through its components

func TestDeleteEntry_IndexParsing(t *testing.T) {
	// Since deleteEntry calls os.Exit, we test by verifying the storage functions work correctly
	// The actual command-level testing would require subprocess testing or refactoring

	tests := []struct {
		name          string
		entries       []entry.Entry
		deleteIndex   int // 0-based for storage layer
		expectedError bool
	}{
		{
			name: "delete first entry",
			entries: []entry.Entry{
				{
					Timestamp:       time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
					Description:     "first entry",
					DurationMinutes: 30,
					RawInput:        "first entry for 30m",
				},
				{
					Timestamp:       time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
					Description:     "second entry",
					DurationMinutes: 60,
					RawInput:        "second entry for 1h",
				},
			},
			deleteIndex:   0,
			expectedError: false,
		},
		{
			name: "delete last entry",
			entries: []entry.Entry{
				{
					Timestamp:       time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
					Description:     "first entry",
					DurationMinutes: 30,
					RawInput:        "first entry for 30m",
				},
				{
					Timestamp:       time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
					Description:     "second entry",
					DurationMinutes: 60,
					RawInput:        "second entry for 1h",
				},
			},
			deleteIndex:   1,
			expectedError: false,
		},
		{
			name: "index out of bounds",
			entries: []entry.Entry{
				{
					Timestamp:       time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
					Description:     "only entry",
					DurationMinutes: 30,
					RawInput:        "only entry for 30m",
				},
			},
			deleteIndex:   5,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempEntriesFile(t, tt.entries)

			deleted, err := storage.DeleteEntry(tmpFile, tt.deleteIndex)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error for index %d, got nil", tt.deleteIndex)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if deleted.Description != tt.entries[tt.deleteIndex].Description {
					t.Errorf("Deleted entry description = %q, expected %q",
						deleted.Description, tt.entries[tt.deleteIndex].Description)
				}

				// Verify remaining entries
				remaining, err := storage.ReadEntries(tmpFile)
				if err != nil {
					t.Fatalf("Failed to read entries after delete: %v", err)
				}

				expectedCount := len(tt.entries) - 1
				if len(remaining) != expectedCount {
					t.Errorf("Expected %d remaining entries, got %d", expectedCount, len(remaining))
				}
			}
		})
	}
}

func TestDeleteEntry_EmptyFile(t *testing.T) {
	tmpFile := createTempEntriesFile(t, []entry.Entry{})

	_, err := storage.DeleteEntry(tmpFile, 0)
	if err == nil {
		t.Error("Expected error when deleting from empty file, got nil")
	}

	if !strings.Contains(err.Error(), "out of bounds") {
		t.Errorf("Error message should mention 'out of bounds', got: %v", err)
	}
}

func TestDeleteEntry_NegativeIndex(t *testing.T) {
	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
			Description:     "test entry",
			DurationMinutes: 30,
			RawInput:        "test entry for 30m",
		},
	}

	tmpFile := createTempEntriesFile(t, entries)

	_, err := storage.DeleteEntry(tmpFile, -1)
	if err == nil {
		t.Error("Expected error when deleting with negative index, got nil")
	}

	if !strings.Contains(err.Error(), "out of bounds") {
		t.Errorf("Error message should mention 'out of bounds', got: %v", err)
	}
}

func TestDeleteEntry_SingleEntry(t *testing.T) {
	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
			Description:     "only entry",
			DurationMinutes: 60,
			RawInput:        "only entry for 1h",
		},
	}

	tmpFile := createTempEntriesFile(t, entries)

	deleted, err := storage.DeleteEntry(tmpFile, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if deleted.Description != "only entry" {
		t.Errorf("Deleted entry description = %q, expected %q", deleted.Description, "only entry")
	}

	// Verify file is now empty
	remaining, err := storage.ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read entries after delete: %v", err)
	}

	if len(remaining) != 0 {
		t.Errorf("Expected 0 remaining entries, got %d", len(remaining))
	}
}

func TestDeleteEntry_MultipleEntriesReindexing(t *testing.T) {
	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
			Description:     "entry 1",
			DurationMinutes: 30,
			RawInput:        "entry 1 for 30m",
		},
		{
			Timestamp:       time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
			Description:     "entry 2",
			DurationMinutes: 60,
			RawInput:        "entry 2 for 1h",
		},
		{
			Timestamp:       time.Date(2024, time.January, 15, 11, 0, 0, 0, time.Local),
			Description:     "entry 3",
			DurationMinutes: 45,
			RawInput:        "entry 3 for 45m",
		},
		{
			Timestamp:       time.Date(2024, time.January, 15, 12, 0, 0, 0, time.Local),
			Description:     "entry 4",
			DurationMinutes: 90,
			RawInput:        "entry 4 for 1h 30m",
		},
	}

	tmpFile := createTempEntriesFile(t, entries)

	// Delete entry at index 1 (second entry - "entry 2")
	deleted, err := storage.DeleteEntry(tmpFile, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if deleted.Description != "entry 2" {
		t.Errorf("Deleted entry description = %q, expected %q", deleted.Description, "entry 2")
	}

	// Verify remaining entries are properly re-indexed
	remaining, err := storage.ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read entries after delete: %v", err)
	}

	if len(remaining) != 3 {
		t.Fatalf("Expected 3 remaining entries, got %d", len(remaining))
	}

	// Verify the order and content
	expectedDescriptions := []string{"entry 1", "entry 3", "entry 4"}
	for i, expected := range expectedDescriptions {
		if remaining[i].Description != expected {
			t.Errorf("Entry %d description = %q, expected %q", i, remaining[i].Description, expected)
		}
	}
}

func TestDeleteEntry_WithSpecialCharacters(t *testing.T) {
	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
			Description:     "fix bug #123 \"critical\" issue",
			DurationMinutes: 60,
			RawInput:        "fix bug #123 \"critical\" issue for 1h",
		},
		{
			Timestamp:       time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
			Description:     "工作 on 功能 🚀",
			DurationMinutes: 30,
			RawInput:        "工作 for 30m",
		},
	}

	tmpFile := createTempEntriesFile(t, entries)

	// Delete first entry with special characters
	deleted, err := storage.DeleteEntry(tmpFile, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if deleted.Description != "fix bug #123 \"critical\" issue" {
		t.Errorf("Deleted entry description = %q, expected %q",
			deleted.Description, "fix bug #123 \"critical\" issue")
	}

	// Verify remaining entry
	remaining, err := storage.ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read entries after delete: %v", err)
	}

	if len(remaining) != 1 {
		t.Fatalf("Expected 1 remaining entry, got %d", len(remaining))
	}

	if remaining[0].Description != "工作 on 功能 🚀" {
		t.Errorf("Remaining entry description = %q, expected %q",
			remaining[0].Description, "工作 on 功能 🚀")
	}
}
