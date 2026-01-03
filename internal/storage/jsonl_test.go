package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
)

// Helper to create a temporary test file
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_entries.jsonl")
	if content != "" {
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
	}
	return tmpFile
}

func TestAppendEntry(t *testing.T) {
	tests := []struct {
		name          string
		entry         entry.Entry
		existingLines int
	}{
		{
			name: "append to empty file",
			entry: entry.Entry{
				Timestamp:       time.Date(2024, time.January, 15, 10, 30, 0, 0, time.Local),
				Description:     "work on feature X",
				DurationMinutes: 120,
				RawInput:        "work on feature X for 2h",
			},
			existingLines: 0,
		},
		{
			name: "append to file with existing entries",
			entry: entry.Entry{
				Timestamp:       time.Date(2024, time.January, 15, 14, 0, 0, 0, time.Local),
				Description:     "code review",
				DurationMinutes: 30,
				RawInput:        "code review for 30m",
			},
			existingLines: 1,
		},
		{
			name: "append entry with special characters in description",
			entry: entry.Entry{
				Timestamp:       time.Date(2024, time.January, 15, 16, 0, 0, 0, time.Local),
				Description:     "fix bug #123 \"critical\" issue",
				DurationMinutes: 45,
				RawInput:        "fix bug #123 \"critical\" issue for 45m",
			},
			existingLines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var initialContent string
			if tt.existingLines > 0 {
				initialContent = `{"timestamp":"2024-01-15T09:00:00Z","description":"existing entry","duration_minutes":60,"raw_input":"existing entry for 1h"}` + "\n"
			}

			tmpFile := createTempFile(t, initialContent)

			err := AppendEntry(tmpFile, tt.entry)
			if err != nil {
				t.Fatalf("AppendEntry() returned unexpected error: %v", err)
			}

			// Verify file exists
			if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
				t.Fatalf("AppendEntry() did not create file")
			}

			// Read entries back
			entries, err := ReadEntries(tmpFile)
			if err != nil {
				t.Fatalf("ReadEntries() returned unexpected error: %v", err)
			}

			expectedCount := tt.existingLines + 1
			if len(entries) != expectedCount {
				t.Errorf("Expected %d entries, got %d", expectedCount, len(entries))
			}

			// Verify the appended entry
			lastEntry := entries[len(entries)-1]
			if lastEntry.Description != tt.entry.Description {
				t.Errorf("Appended entry description = %q, expected %q", lastEntry.Description, tt.entry.Description)
			}
			if lastEntry.DurationMinutes != tt.entry.DurationMinutes {
				t.Errorf("Appended entry duration = %d, expected %d", lastEntry.DurationMinutes, tt.entry.DurationMinutes)
			}
			if lastEntry.RawInput != tt.entry.RawInput {
				t.Errorf("Appended entry raw_input = %q, expected %q", lastEntry.RawInput, tt.entry.RawInput)
			}
		})
	}
}

func TestAppendEntry_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "new_entries.jsonl")

	// Verify file doesn't exist
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Fatalf("Test setup error: file should not exist")
	}

	testEntry := entry.Entry{
		Timestamp:       time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description:     "test entry",
		DurationMinutes: 60,
		RawInput:        "test entry for 1h",
	}

	err := AppendEntry(tmpFile, testEntry)
	if err != nil {
		t.Fatalf("AppendEntry() returned unexpected error: %v", err)
	}

	// Verify file now exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Errorf("AppendEntry() did not create file")
	}

	// Verify entry was written
	entries, err := ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntries() returned unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestAppendEntry_MultipleEntries(t *testing.T) {
	tmpFile := createTempFile(t, "")

	testEntries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
			Description:     "morning standup",
			DurationMinutes: 15,
			RawInput:        "morning standup for 15m",
		},
		{
			Timestamp:       time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
			Description:     "feature development",
			DurationMinutes: 120,
			RawInput:        "feature development for 2h",
		},
		{
			Timestamp:       time.Date(2024, time.January, 15, 14, 0, 0, 0, time.Local),
			Description:     "code review",
			DurationMinutes: 45,
			RawInput:        "code review for 45m",
		},
	}

	// Append all entries
	for _, e := range testEntries {
		if err := AppendEntry(tmpFile, e); err != nil {
			t.Fatalf("AppendEntry() returned unexpected error: %v", err)
		}
	}

	// Read back all entries
	entries, err := ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntries() returned unexpected error: %v", err)
	}

	if len(entries) != len(testEntries) {
		t.Errorf("Expected %d entries, got %d", len(testEntries), len(entries))
	}

	// Verify order is preserved
	for i, expected := range testEntries {
		if entries[i].Description != expected.Description {
			t.Errorf("Entry %d description = %q, expected %q", i, entries[i].Description, expected.Description)
		}
	}
}

func TestReadEntries_Empty(t *testing.T) {
	// Test with non-existent file
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.jsonl")

	entries, err := ReadEntries(nonExistentFile)
	if err != nil {
		t.Errorf("ReadEntries() returned unexpected error for non-existent file: %v", err)
	}
	if entries == nil {
		t.Errorf("ReadEntries() returned nil, expected empty slice")
	}
	if len(entries) != 0 {
		t.Errorf("ReadEntries() returned %d entries, expected 0", len(entries))
	}
}

func TestReadEntries_EmptyFile(t *testing.T) {
	// Test with existing but empty file
	tmpFile := createTempFile(t, "")

	// Manually create empty file
	if err := os.WriteFile(tmpFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	entries, err := ReadEntries(tmpFile)
	if err != nil {
		t.Errorf("ReadEntries() returned unexpected error for empty file: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("ReadEntries() returned %d entries, expected 0", len(entries))
	}
}

func TestReadEntries_Malformed(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		expectedCount  int
		expectedDesc   string
	}{
		{
			name: "skip malformed line, return valid entry",
			fileContent: `{"timestamp":"2024-01-15T10:00:00Z","description":"valid entry","duration_minutes":60,"raw_input":"valid entry for 1h"}
invalid json line
{"timestamp":"2024-01-15T11:00:00Z","description":"another valid","duration_minutes":30,"raw_input":"another valid for 30m"}
`,
			expectedCount: 2,
			expectedDesc:  "valid entry",
		},
		{
			name: "skip empty lines",
			fileContent: `{"timestamp":"2024-01-15T10:00:00Z","description":"valid entry","duration_minutes":60,"raw_input":"valid entry for 1h"}

{"timestamp":"2024-01-15T11:00:00Z","description":"second valid","duration_minutes":30,"raw_input":"second valid for 30m"}
`,
			expectedCount: 2,
			expectedDesc:  "valid entry",
		},
		{
			name: "skip truncated JSON",
			fileContent: `{"timestamp":"2024-01-15T10:00:00Z","description":"valid entry","duration_minutes":60,"raw_input":"valid entry for 1h"}
{"timestamp":"2024-01-15T11:00:00Z","description":"trun
{"timestamp":"2024-01-15T12:00:00Z","description":"third valid","duration_minutes":45,"raw_input":"third valid for 45m"}
`,
			expectedCount: 2,
			expectedDesc:  "valid entry",
		},
		{
			name: "all lines malformed",
			fileContent: `invalid line 1
not json at all
{incomplete json
`,
			expectedCount: 0,
			expectedDesc:  "",
		},
		{
			name: "valid entry with extra whitespace",
			fileContent: `  {"timestamp":"2024-01-15T10:00:00Z","description":"with whitespace","duration_minutes":60,"raw_input":"with whitespace for 1h"}
`,
			expectedCount: 1, // Go's json.Unmarshal handles leading whitespace correctly
			expectedDesc:  "with whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempFile(t, tt.fileContent)

			entries, err := ReadEntries(tmpFile)
			if err != nil {
				t.Errorf("ReadEntries() returned unexpected error: %v", err)
			}

			if len(entries) != tt.expectedCount {
				t.Errorf("ReadEntries() returned %d entries, expected %d", len(entries), tt.expectedCount)
			}

			if tt.expectedCount > 0 && entries[0].Description != tt.expectedDesc {
				t.Errorf("First entry description = %q, expected %q", entries[0].Description, tt.expectedDesc)
			}
		})
	}
}

func TestReadEntries_ValidFormat(t *testing.T) {
	fileContent := `{"timestamp":"2024-01-15T10:30:00Z","description":"work on feature X","duration_minutes":120,"raw_input":"work on feature X for 2h"}
{"timestamp":"2024-01-15T14:00:00Z","description":"code review","duration_minutes":30,"raw_input":"code review for 30m"}
{"timestamp":"2024-01-15T15:00:00Z","description":"fix bug #123","duration_minutes":45,"raw_input":"fix bug #123 for 45m"}
`
	tmpFile := createTempFile(t, fileContent)

	entries, err := ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntries() returned unexpected error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("ReadEntries() returned %d entries, expected 3", len(entries))
	}

	// Verify first entry
	if entries[0].Description != "work on feature X" {
		t.Errorf("Entry 0 description = %q, expected %q", entries[0].Description, "work on feature X")
	}
	if entries[0].DurationMinutes != 120 {
		t.Errorf("Entry 0 duration = %d, expected %d", entries[0].DurationMinutes, 120)
	}
	if entries[0].RawInput != "work on feature X for 2h" {
		t.Errorf("Entry 0 raw_input = %q, expected %q", entries[0].RawInput, "work on feature X for 2h")
	}

	// Verify second entry
	if entries[1].Description != "code review" {
		t.Errorf("Entry 1 description = %q, expected %q", entries[1].Description, "code review")
	}
	if entries[1].DurationMinutes != 30 {
		t.Errorf("Entry 1 duration = %d, expected %d", entries[1].DurationMinutes, 30)
	}

	// Verify third entry
	if entries[2].Description != "fix bug #123" {
		t.Errorf("Entry 2 description = %q, expected %q", entries[2].Description, "fix bug #123")
	}
	if entries[2].DurationMinutes != 45 {
		t.Errorf("Entry 2 duration = %d, expected %d", entries[2].DurationMinutes, 45)
	}
}

func TestReadEntries_PreservesTimestamp(t *testing.T) {
	// Use a specific timestamp to verify parsing
	fileContent := `{"timestamp":"2024-06-15T09:30:45Z","description":"test","duration_minutes":60,"raw_input":"test for 1h"}
`
	tmpFile := createTempFile(t, fileContent)

	entries, err := ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntries() returned unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("ReadEntries() returned %d entries, expected 1", len(entries))
	}

	expected := time.Date(2024, time.June, 15, 9, 30, 45, 0, time.UTC)
	if !entries[0].Timestamp.Equal(expected) {
		t.Errorf("Entry timestamp = %v, expected %v", entries[0].Timestamp, expected)
	}
}

func TestAppendEntry_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "permissions_test.jsonl")

	testEntry := entry.Entry{
		Timestamp:       time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description:     "test",
		DurationMinutes: 60,
		RawInput:        "test for 1h",
	}

	err := AppendEntry(tmpFile, testEntry)
	if err != nil {
		t.Fatalf("AppendEntry() returned unexpected error: %v", err)
	}

	// Check file permissions (should be 0644)
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	expectedPerm := os.FileMode(0644)
	actualPerm := info.Mode().Perm()
	if actualPerm != expectedPerm {
		t.Errorf("File permissions = %o, expected %o", actualPerm, expectedPerm)
	}
}

func TestReadEntries_LargeFile(t *testing.T) {
	var content string
	numEntries := 100

	for i := 0; i < numEntries; i++ {
		timestamp := time.Date(2024, time.January, 15, 9+i/4, (i%4)*15, 0, 0, time.UTC)
		line := `{"timestamp":"` + timestamp.Format(time.RFC3339) + `","description":"entry ` + string(rune('0'+i%10)) + `","duration_minutes":15,"raw_input":"entry for 15m"}` + "\n"
		content += line
	}

	tmpFile := createTempFile(t, content)

	entries, err := ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntries() returned unexpected error: %v", err)
	}

	if len(entries) != numEntries {
		t.Errorf("ReadEntries() returned %d entries, expected %d", len(entries), numEntries)
	}
}

func TestGetStoragePath(t *testing.T) {
	// Test that GetStoragePath returns a valid path
	path, err := GetStoragePath()
	if err != nil {
		t.Fatalf("GetStoragePath() returned unexpected error: %v", err)
	}

	// Path should not be empty
	if path == "" {
		t.Errorf("GetStoragePath() returned empty path")
	}

	// Path should end with entries.jsonl
	if filepath.Base(path) != EntriesFile {
		t.Errorf("GetStoragePath() path base = %q, expected %q", filepath.Base(path), EntriesFile)
	}

	// Parent directory should exist
	parentDir := filepath.Dir(path)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Errorf("GetStoragePath() parent directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("GetStoragePath() parent is not a directory")
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are set correctly
	if AppName != "did" {
		t.Errorf("AppName = %q, expected %q", AppName, "did")
	}

	if EntriesFile != "entries.jsonl" {
		t.Errorf("EntriesFile = %q, expected %q", EntriesFile, "entries.jsonl")
	}
}

func TestReadEntries_UnicodeContent(t *testing.T) {
	// Test handling of unicode characters in descriptions
	fileContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"工作 on 功能 🚀","duration_minutes":60,"raw_input":"工作 for 1h"}
{"timestamp":"2024-01-15T11:00:00Z","description":"código review","duration_minutes":30,"raw_input":"código review for 30m"}
`
	tmpFile := createTempFile(t, fileContent)

	entries, err := ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntries() returned unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("ReadEntries() returned %d entries, expected 2", len(entries))
	}

	if entries[0].Description != "工作 on 功能 🚀" {
		t.Errorf("Entry 0 description = %q, expected %q", entries[0].Description, "工作 on 功能 🚀")
	}

	if entries[1].Description != "código review" {
		t.Errorf("Entry 1 description = %q, expected %q", entries[1].Description, "código review")
	}
}

func TestAppendEntry_UnicodeContent(t *testing.T) {
	tmpFile := createTempFile(t, "")

	testEntry := entry.Entry{
		Timestamp:       time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description:     "工作 on feature 🎉",
		DurationMinutes: 60,
		RawInput:        "工作 on feature 🎉 for 1h",
	}

	err := AppendEntry(tmpFile, testEntry)
	if err != nil {
		t.Fatalf("AppendEntry() returned unexpected error: %v", err)
	}

	// Read back and verify
	entries, err := ReadEntries(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntries() returned unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Description != testEntry.Description {
		t.Errorf("Entry description = %q, expected %q", entries[0].Description, testEntry.Description)
	}
}

func TestReadEntriesWithWarnings_MalformedJSON(t *testing.T) {
	tests := []struct {
		name             string
		fileContent      string
		expectedEntries  int
		expectedWarnings int
	}{
		{
			name: "malformed JSON line",
			fileContent: `{"timestamp":"2024-01-15T10:00:00Z","description":"valid entry","duration_minutes":60,"raw_input":"valid entry for 1h"}
{invalid json}
{"timestamp":"2024-01-15T11:00:00Z","description":"another valid","duration_minutes":30,"raw_input":"another valid for 30m"}
`,
			expectedEntries:  2,
			expectedWarnings: 1,
		},
		{
			name: "truncated JSON line",
			fileContent: `{"timestamp":"2024-01-15T10:00:00Z","description":"valid entry","duration_minutes":60,"raw_input":"valid entry for 1h"}
{"timestamp":"2024-01-15T11:00:00Z","description":"trun
{"timestamp":"2024-01-15T12:00:00Z","description":"third valid","duration_minutes":45,"raw_input":"third valid for 45m"}
`,
			expectedEntries:  2,
			expectedWarnings: 1,
		},
		{
			name: "empty line",
			fileContent: `{"timestamp":"2024-01-15T10:00:00Z","description":"valid entry","duration_minutes":60,"raw_input":"valid entry for 1h"}

{"timestamp":"2024-01-15T11:00:00Z","description":"second valid","duration_minutes":30,"raw_input":"second valid for 30m"}
`,
			expectedEntries:  2,
			expectedWarnings: 1,
		},
		{
			name: "mixed valid and invalid",
			fileContent: `{"timestamp":"2024-01-15T10:00:00Z","description":"valid 1","duration_minutes":60,"raw_input":"valid 1 for 1h"}
invalid line 1
{"timestamp":"2024-01-15T11:00:00Z","description":"valid 2","duration_minutes":30,"raw_input":"valid 2 for 30m"}
{broken json
{"timestamp":"2024-01-15T12:00:00Z","description":"valid 3","duration_minutes":45,"raw_input":"valid 3 for 45m"}

{"timestamp":"2024-01-15T13:00:00Z","description":"valid 4","duration_minutes":15,"raw_input":"valid 4 for 15m"}
`,
			expectedEntries:  4,
			expectedWarnings: 3,
		},
		{
			name: "all lines corrupted",
			fileContent: `invalid line 1
not json at all
{incomplete json
`,
			expectedEntries:  0,
			expectedWarnings: 3,
		},
		{
			name: "only valid entries",
			fileContent: `{"timestamp":"2024-01-15T10:00:00Z","description":"valid 1","duration_minutes":60,"raw_input":"valid 1 for 1h"}
{"timestamp":"2024-01-15T11:00:00Z","description":"valid 2","duration_minutes":30,"raw_input":"valid 2 for 30m"}
`,
			expectedEntries:  2,
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempFile(t, tt.fileContent)

			result, err := ReadEntriesWithWarnings(tmpFile)
			if err != nil {
				t.Fatalf("ReadEntriesWithWarnings() returned unexpected error: %v", err)
			}

			if len(result.Entries) != tt.expectedEntries {
				t.Errorf("ReadEntriesWithWarnings() returned %d entries, expected %d", len(result.Entries), tt.expectedEntries)
			}

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("ReadEntriesWithWarnings() returned %d warnings, expected %d", len(result.Warnings), tt.expectedWarnings)
			}
		})
	}
}

func TestReadEntriesWithWarnings_WarningDetails(t *testing.T) {
	fileContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"valid entry","duration_minutes":60,"raw_input":"valid entry for 1h"}
invalid json line
{"timestamp":"2024-01-15T11:00:00Z","description":"another valid","duration_minutes":30,"raw_input":"another valid for 30m"}
`
	tmpFile := createTempFile(t, fileContent)

	result, err := ReadEntriesWithWarnings(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntriesWithWarnings() returned unexpected error: %v", err)
	}

	if len(result.Warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(result.Warnings))
	}

	warning := result.Warnings[0]

	// Verify line number (1-indexed)
	if warning.LineNumber != 2 {
		t.Errorf("Warning line number = %d, expected 2", warning.LineNumber)
	}

	// Verify content
	if warning.Content != "invalid json line" {
		t.Errorf("Warning content = %q, expected %q", warning.Content, "invalid json line")
	}

	// Verify error message is not empty
	if warning.Error == "" {
		t.Errorf("Warning error message is empty")
	}
}

func TestReadEntriesWithWarnings_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.jsonl")

	result, err := ReadEntriesWithWarnings(nonExistentFile)
	if err != nil {
		t.Errorf("ReadEntriesWithWarnings() returned unexpected error for non-existent file: %v", err)
	}

	if result.Entries == nil {
		t.Errorf("ReadEntriesWithWarnings() returned nil entries, expected empty slice")
	}

	if len(result.Entries) != 0 {
		t.Errorf("ReadEntriesWithWarnings() returned %d entries, expected 0", len(result.Entries))
	}

	if result.Warnings == nil {
		t.Errorf("ReadEntriesWithWarnings() returned nil warnings, expected empty slice")
	}

	if len(result.Warnings) != 0 {
		t.Errorf("ReadEntriesWithWarnings() returned %d warnings, expected 0", len(result.Warnings))
	}
}

func TestReadEntriesWithWarnings_MultipleWarnings(t *testing.T) {
	fileContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"valid 1","duration_minutes":60,"raw_input":"valid 1 for 1h"}
corrupted line 1
corrupted line 2

corrupted line 3
{"timestamp":"2024-01-15T11:00:00Z","description":"valid 2","duration_minutes":30,"raw_input":"valid 2 for 30m"}
`
	tmpFile := createTempFile(t, fileContent)

	result, err := ReadEntriesWithWarnings(tmpFile)
	if err != nil {
		t.Fatalf("ReadEntriesWithWarnings() returned unexpected error: %v", err)
	}

	if len(result.Entries) != 2 {
		t.Errorf("Expected 2 valid entries, got %d", len(result.Entries))
	}

	if len(result.Warnings) != 4 {
		t.Errorf("Expected 4 warnings, got %d", len(result.Warnings))
	}

	// Verify line numbers are correct and in order
	expectedLineNumbers := []int{2, 3, 4, 5}
	for i, warning := range result.Warnings {
		if warning.LineNumber != expectedLineNumbers[i] {
			t.Errorf("Warning %d line number = %d, expected %d", i, warning.LineNumber, expectedLineNumbers[i])
		}
	}
}

func TestValidateStorage_HealthyFile(t *testing.T) {
	fileContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"valid 1","duration_minutes":60,"raw_input":"valid 1 for 1h"}
{"timestamp":"2024-01-15T11:00:00Z","description":"valid 2","duration_minutes":30,"raw_input":"valid 2 for 30m"}
{"timestamp":"2024-01-15T12:00:00Z","description":"valid 3","duration_minutes":45,"raw_input":"valid 3 for 45m"}
`
	tmpFile := createTempFile(t, fileContent)

	health, err := ValidateStorage(tmpFile)
	if err != nil {
		t.Fatalf("ValidateStorage() returned unexpected error: %v", err)
	}

	if health.TotalLines != 3 {
		t.Errorf("ValidateStorage() TotalLines = %d, expected 3", health.TotalLines)
	}

	if health.ValidEntries != 3 {
		t.Errorf("ValidateStorage() ValidEntries = %d, expected 3", health.ValidEntries)
	}

	if health.CorruptedEntries != 0 {
		t.Errorf("ValidateStorage() CorruptedEntries = %d, expected 0", health.CorruptedEntries)
	}

	if len(health.Warnings) != 0 {
		t.Errorf("ValidateStorage() returned %d warnings, expected 0", len(health.Warnings))
	}
}

func TestValidateStorage_CorruptedFile(t *testing.T) {
	fileContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"valid 1","duration_minutes":60,"raw_input":"valid 1 for 1h"}
invalid json line
{"timestamp":"2024-01-15T11:00:00Z","description":"valid 2","duration_minutes":30,"raw_input":"valid 2 for 30m"}
{broken
{"timestamp":"2024-01-15T12:00:00Z","description":"valid 3","duration_minutes":45,"raw_input":"valid 3 for 45m"}
`
	tmpFile := createTempFile(t, fileContent)

	health, err := ValidateStorage(tmpFile)
	if err != nil {
		t.Fatalf("ValidateStorage() returned unexpected error: %v", err)
	}

	if health.TotalLines != 5 {
		t.Errorf("ValidateStorage() TotalLines = %d, expected 5", health.TotalLines)
	}

	if health.ValidEntries != 3 {
		t.Errorf("ValidateStorage() ValidEntries = %d, expected 3", health.ValidEntries)
	}

	if health.CorruptedEntries != 2 {
		t.Errorf("ValidateStorage() CorruptedEntries = %d, expected 2", health.CorruptedEntries)
	}

	if len(health.Warnings) != 2 {
		t.Errorf("ValidateStorage() returned %d warnings, expected 2", len(health.Warnings))
	}
}

func TestValidateStorage_AllCorrupted(t *testing.T) {
	fileContent := `invalid line 1
not json at all
{incomplete json
`
	tmpFile := createTempFile(t, fileContent)

	health, err := ValidateStorage(tmpFile)
	if err != nil {
		t.Fatalf("ValidateStorage() returned unexpected error: %v", err)
	}

	if health.TotalLines != 3 {
		t.Errorf("ValidateStorage() TotalLines = %d, expected 3", health.TotalLines)
	}

	if health.ValidEntries != 0 {
		t.Errorf("ValidateStorage() ValidEntries = %d, expected 0", health.ValidEntries)
	}

	if health.CorruptedEntries != 3 {
		t.Errorf("ValidateStorage() CorruptedEntries = %d, expected 3", health.CorruptedEntries)
	}

	if len(health.Warnings) != 3 {
		t.Errorf("ValidateStorage() returned %d warnings, expected 3", len(health.Warnings))
	}
}

func TestValidateStorage_EmptyFile(t *testing.T) {
	tmpFile := createTempFile(t, "")

	// Manually create empty file
	if err := os.WriteFile(tmpFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	health, err := ValidateStorage(tmpFile)
	if err != nil {
		t.Fatalf("ValidateStorage() returned unexpected error: %v", err)
	}

	if health.TotalLines != 0 {
		t.Errorf("ValidateStorage() TotalLines = %d, expected 0", health.TotalLines)
	}

	if health.ValidEntries != 0 {
		t.Errorf("ValidateStorage() ValidEntries = %d, expected 0", health.ValidEntries)
	}

	if health.CorruptedEntries != 0 {
		t.Errorf("ValidateStorage() CorruptedEntries = %d, expected 0", health.CorruptedEntries)
	}
}

func TestValidateStorage_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.jsonl")

	health, err := ValidateStorage(nonExistentFile)
	if err != nil {
		t.Errorf("ValidateStorage() returned unexpected error for non-existent file: %v", err)
	}

	if health.TotalLines != 0 {
		t.Errorf("ValidateStorage() TotalLines = %d, expected 0", health.TotalLines)
	}

	if health.ValidEntries != 0 {
		t.Errorf("ValidateStorage() ValidEntries = %d, expected 0", health.ValidEntries)
	}

	if health.CorruptedEntries != 0 {
		t.Errorf("ValidateStorage() CorruptedEntries = %d, expected 0", health.CorruptedEntries)
	}
}

func TestValidateStorage_EmptyLines(t *testing.T) {
	fileContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"valid 1","duration_minutes":60,"raw_input":"valid 1 for 1h"}

{"timestamp":"2024-01-15T11:00:00Z","description":"valid 2","duration_minutes":30,"raw_input":"valid 2 for 30m"}

`
	tmpFile := createTempFile(t, fileContent)

	health, err := ValidateStorage(tmpFile)
	if err != nil {
		t.Fatalf("ValidateStorage() returned unexpected error: %v", err)
	}

	// Total lines includes empty lines
	if health.TotalLines != 4 {
		t.Errorf("ValidateStorage() TotalLines = %d, expected 4", health.TotalLines)
	}

	if health.ValidEntries != 2 {
		t.Errorf("ValidateStorage() ValidEntries = %d, expected 2", health.ValidEntries)
	}

	// Empty lines are treated as corrupted
	if health.CorruptedEntries != 2 {
		t.Errorf("ValidateStorage() CorruptedEntries = %d, expected 2", health.CorruptedEntries)
	}
}

func TestValidateStorage_WarningDetails(t *testing.T) {
	fileContent := `{"timestamp":"2024-01-15T10:00:00Z","description":"valid 1","duration_minutes":60,"raw_input":"valid 1 for 1h"}
corrupted line content
{"timestamp":"2024-01-15T11:00:00Z","description":"valid 2","duration_minutes":30,"raw_input":"valid 2 for 30m"}
`
	tmpFile := createTempFile(t, fileContent)

	health, err := ValidateStorage(tmpFile)
	if err != nil {
		t.Fatalf("ValidateStorage() returned unexpected error: %v", err)
	}

	if len(health.Warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(health.Warnings))
	}

	warning := health.Warnings[0]

	if warning.LineNumber != 2 {
		t.Errorf("Warning line number = %d, expected 2", warning.LineNumber)
	}

	if warning.Content != "corrupted line content" {
		t.Errorf("Warning content = %q, expected %q", warning.Content, "corrupted line content")
	}

	if warning.Error == "" {
		t.Errorf("Warning error message is empty")
	}
}
