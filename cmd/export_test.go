package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
)

// ExportOutput represents the structure of the JSON export
type ExportOutput struct {
	Metadata struct {
		ExportTimestamp time.Time              `json:"export_timestamp"`
		TotalEntries    int                    `json:"total_entries"`
		FilterCriteria  map[string]interface{} `json:"filter_criteria"`
	} `json:"metadata"`
	Entries []entry.Entry `json:"entries"`
}

// Helper function to create test entries for export testing
func createExportTestEntries(t *testing.T, storagePath string) []entry.Entry {
	t.Helper()

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -7), // 7 days ago
			Description:     "Code review for feature X",
			DurationMinutes: 60,
			RawInput:        "Code review for feature X for 1h",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -5), // 5 days ago
			Description:     "Bug fix in authentication",
			DurationMinutes: 90,
			RawInput:        "Bug fix in authentication for 1h30m",
			Project:         "client",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -3), // 3 days ago
			Description:     "Team meeting to discuss roadmap",
			DurationMinutes: 45,
			RawInput:        "Team meeting to discuss roadmap for 45m",
			Project:         "",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	return entries
}

func TestExportJSON_ValidOutput(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	exportJSON(exportJSONCmd)

	output := stdout.String()

	// Verify output is valid JSON
	var result ExportOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify we got 3 entries
	if len(result.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(result.Entries))
	}

	// Verify metadata
	if result.Metadata.TotalEntries != 3 {
		t.Errorf("Expected total_entries=3, got %d", result.Metadata.TotalEntries)
	}

	// Verify export_timestamp is recent (within last minute)
	if time.Since(result.Metadata.ExportTimestamp) > time.Minute {
		t.Errorf("Export timestamp is not recent: %v", result.Metadata.ExportTimestamp)
	}

	// Verify filter_criteria exists (should be empty for no filters)
	if result.Metadata.FilterCriteria == nil {
		t.Error("Expected filter_criteria to be initialized")
	}
}

func TestExportJSON_MetadataStructure(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	exportJSON(exportJSONCmd)

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check metadata object exists
	metadata, ok := result["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'metadata' object in JSON output")
	}

	// Check required metadata fields
	if _, ok := metadata["export_timestamp"]; !ok {
		t.Error("Expected 'export_timestamp' in metadata")
	}

	if _, ok := metadata["total_entries"]; !ok {
		t.Error("Expected 'total_entries' in metadata")
	}

	if _, ok := metadata["filter_criteria"]; !ok {
		t.Error("Expected 'filter_criteria' in metadata")
	}

	// Check entries array exists
	entries, ok := result["entries"].([]interface{})
	if !ok {
		t.Fatal("Expected 'entries' array in JSON output")
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestExportJSON_EntryFormat(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	testEntry := entry.Entry{
		Timestamp:       now,
		Description:     "Test entry",
		DurationMinutes: 120,
		RawInput:        "Test entry for 2h",
		Project:         "testproject",
		Tags:            []string{"tag1", "tag2"},
	}

	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(result.Entries))
	}

	exportedEntry := result.Entries[0]

	// Verify all entry fields are present and correct
	if exportedEntry.Description != testEntry.Description {
		t.Errorf("Expected description %q, got %q", testEntry.Description, exportedEntry.Description)
	}

	if exportedEntry.DurationMinutes != testEntry.DurationMinutes {
		t.Errorf("Expected duration %d, got %d", testEntry.DurationMinutes, exportedEntry.DurationMinutes)
	}

	if exportedEntry.RawInput != testEntry.RawInput {
		t.Errorf("Expected raw_input %q, got %q", testEntry.RawInput, exportedEntry.RawInput)
	}

	if exportedEntry.Project != testEntry.Project {
		t.Errorf("Expected project %q, got %q", testEntry.Project, exportedEntry.Project)
	}

	if len(exportedEntry.Tags) != len(testEntry.Tags) {
		t.Errorf("Expected %d tags, got %d", len(testEntry.Tags), len(exportedEntry.Tags))
	}

	for i, tag := range testEntry.Tags {
		if exportedEntry.Tags[i] != tag {
			t.Errorf("Expected tag[%d]=%q, got %q", i, tag, exportedEntry.Tags[i])
		}
	}

	// Verify timestamp is preserved (within 1 second tolerance for JSON serialization)
	timeDiff := exportedEntry.Timestamp.Sub(testEntry.Timestamp)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("Timestamp mismatch: expected %v, got %v", testEntry.Timestamp, exportedEntry.Timestamp)
	}
}

func TestExportJSON_EmptyStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	// Don't create any entries - storage file doesn't exist

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

	exportJSON(exportJSONCmd)

	output := stdout.String()

	// Verify output is valid JSON
	var result ExportOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify entries array is empty
	if len(result.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(result.Entries))
	}

	// Verify total_entries is 0
	if result.Metadata.TotalEntries != 0 {
		t.Errorf("Expected total_entries=0, got %d", result.Metadata.TotalEntries)
	}

	// Verify metadata is still present
	if result.Metadata.FilterCriteria == nil {
		t.Error("Expected filter_criteria to be initialized")
	}
}

func TestExportJSON_CorruptedEntriesHandled(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a valid entry
	validEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "Valid entry",
		DurationMinutes: 60,
		RawInput:        "Valid entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, validEntry); err != nil {
		t.Fatalf("Failed to create valid entry: %v", err)
	}

	// Append corrupted line to the file
	f, err := os.OpenFile(storagePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open storage file: %v", err)
	}
	_, err = f.WriteString("this is not valid json\n")
	if closeErr := f.Close(); closeErr != nil {
		t.Fatalf("Failed to close storage file: %v", closeErr)
	}
	if err != nil {
		t.Fatalf("Failed to write corrupted line: %v", err)
	}

	// Create another valid entry
	validEntry2 := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "Another valid entry",
		DurationMinutes: 30,
		RawInput:        "Another valid entry for 30m",
	}
	if err := storage.AppendEntry(storagePath, validEntry2); err != nil {
		t.Fatalf("Failed to create second valid entry: %v", err)
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

	exportJSON(exportJSONCmd)

	// Verify warning is shown on stderr
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Warning:") {
		t.Error("Expected warning about corrupted line on stderr")
	}
	if !strings.Contains(stderrOutput, "corrupted line") {
		t.Error("Expected 'corrupted line' message on stderr")
	}

	// Verify JSON output is still valid
	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, stdout.String())
	}

	// Verify only valid entries are exported
	if len(result.Entries) != 2 {
		t.Errorf("Expected 2 valid entries, got %d", len(result.Entries))
	}

	if result.Metadata.TotalEntries != 2 {
		t.Errorf("Expected total_entries=2, got %d", result.Metadata.TotalEntries)
	}
}

func TestExportJSON_StoragePathError(t *testing.T) {
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

	exportJSON(exportJSONCmd)

	if !exitCalled {
		t.Error("Expected exit to be called for storage path error")
	}
	if !strings.Contains(stderr.String(), "Failed to determine storage location") {
		t.Errorf("Expected storage path error, got: %s", stderr.String())
	}
}

func TestExportJSON_ReadEntriesError(t *testing.T) {
	// Use a path to a directory (not a file) to cause read error
	tmpDir := t.TempDir()

	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return tmpDir, nil // path to directory, not file
		},
	}
	SetDeps(d)
	defer ResetDeps()

	exportJSON(exportJSONCmd)

	if !exitCalled {
		t.Error("Expected exit to be called for read entries error")
	}
	if !strings.Contains(stderr.String(), "Failed to read entries") {
		t.Errorf("Expected read entries error, got: %s", stderr.String())
	}
}

func TestExportJSON_OutputCanBeRedirected(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	exportJSON(exportJSONCmd)

	// Simulate writing to file
	outputFile := filepath.Join(tmpDir, "backup.json")
	if err := os.WriteFile(outputFile, stdout.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write output to file: %v", err)
	}

	// Verify file can be read back and parsed
	fileContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result ExportOutput
	if err := json.Unmarshal(fileContent, &result); err != nil {
		t.Fatalf("Failed to parse JSON from file: %v", err)
	}

	// Verify data integrity after file roundtrip
	if len(result.Entries) != 3 {
		t.Errorf("Expected 3 entries in file, got %d", len(result.Entries))
	}

	if result.Metadata.TotalEntries != 3 {
		t.Errorf("Expected total_entries=3 in file, got %d", result.Metadata.TotalEntries)
	}
}

func TestExportJSON_PrettyPrintedOutput(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	testEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "Test entry",
		DurationMinutes: 60,
		RawInput:        "Test entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	exportJSON(exportJSONCmd)

	output := stdout.String()

	// Verify output is pretty-printed (contains indentation)
	if !strings.Contains(output, "  ") {
		t.Error("Expected pretty-printed JSON with indentation")
	}

	// Verify output contains newlines (not compact)
	lines := strings.Split(output, "\n")
	if len(lines) < 5 {
		t.Errorf("Expected multi-line output, got %d lines", len(lines))
	}

	// Verify key JSON structure elements are on separate lines
	if !strings.Contains(output, "\"metadata\":") {
		t.Error("Expected 'metadata' key in output")
	}
	if !strings.Contains(output, "\"entries\":") {
		t.Error("Expected 'entries' key in output")
	}
}

func TestExportCommand_Exists(t *testing.T) {
	// Verify export command is registered
	if exportCmd == nil {
		t.Fatal("exportCmd should be defined")
	}

	if exportCmd.Use != "export" {
		t.Errorf("Expected export command Use='export', got %q", exportCmd.Use)
	}

	// Verify json subcommand exists
	if exportJSONCmd == nil {
		t.Fatal("exportJSONCmd should be defined")
	}

	if exportJSONCmd.Use != "json" {
		t.Errorf("Expected json command Use='json', got %q", exportJSONCmd.Use)
	}
}

// Date filtering tests

func TestExportJSON_FromFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with different dates
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10), // 10 days ago (before range)
			Description:     "Old entry",
			DurationMinutes: 60,
			RawInput:        "Old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -3), // 3 days ago (in range)
			Description:     "Recent entry",
			DurationMinutes: 90,
			RawInput:        "Recent entry for 1h30m",
		},
		{
			Timestamp:       now.AddDate(0, 0, -1), // 1 day ago (in range)
			Description:     "Yesterday entry",
			DurationMinutes: 45,
			RawInput:        "Yesterday entry for 45m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Set --from flag to 5 days ago
	fromDate := now.AddDate(0, 0, -5).Format("2006-01-02")
	if err := exportJSONCmd.Flags().Set("from", fromDate); err != nil {
		t.Fatalf("Failed to set from flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries from last 5 days (not the 10-day-old entry)
	if len(result.Entries) != 2 {
		t.Errorf("Expected 2 entries (from last 5 days), got %d", len(result.Entries))
	}

	// Verify filter criteria in metadata
	if result.Metadata.FilterCriteria["from"] == nil {
		t.Error("Expected 'from' in filter_criteria")
	}
}

func TestExportJSON_ToFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with different dates
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10), // 10 days ago (in range)
			Description:     "Old entry",
			DurationMinutes: 60,
			RawInput:        "Old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -3), // 3 days ago (in range)
			Description:     "Recent entry",
			DurationMinutes: 90,
			RawInput:        "Recent entry for 1h30m",
		},
		{
			Timestamp:       now, // Today (after range)
			Description:     "Today entry",
			DurationMinutes: 45,
			RawInput:        "Today entry for 45m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Set --to flag to 5 days ago
	toDate := now.AddDate(0, 0, -5).Format("2006-01-02")
	if err := exportJSONCmd.Flags().Set("to", toDate); err != nil {
		t.Fatalf("Failed to set to flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries up to 5 days ago
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry (older than 5 days), got %d", len(result.Entries))
	}

	if result.Entries[0].Description != "Old entry" {
		t.Errorf("Expected 'Old entry', got %q", result.Entries[0].Description)
	}

	// Verify filter criteria in metadata
	if result.Metadata.FilterCriteria["to"] == nil {
		t.Error("Expected 'to' in filter_criteria")
	}
}

func TestExportJSON_FromAndToFlags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with different dates
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -15), // 15 days ago (before range)
			Description:     "Very old entry",
			DurationMinutes: 60,
			RawInput:        "Very old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -7), // 7 days ago (in range)
			Description:     "Week old entry",
			DurationMinutes: 90,
			RawInput:        "Week old entry for 1h30m",
		},
		{
			Timestamp:       now.AddDate(0, 0, -5), // 5 days ago (in range)
			Description:     "Five days old",
			DurationMinutes: 45,
			RawInput:        "Five days old for 45m",
		},
		{
			Timestamp:       now.AddDate(0, 0, -2), // 2 days ago (after range)
			Description:     "Recent entry",
			DurationMinutes: 30,
			RawInput:        "Recent entry for 30m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Set --from and --to flags for 10 days ago to 3 days ago
	fromDate := now.AddDate(0, 0, -10).Format("2006-01-02")
	toDate := now.AddDate(0, 0, -3).Format("2006-01-02")
	if err := exportJSONCmd.Flags().Set("from", fromDate); err != nil {
		t.Fatalf("Failed to set from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", toDate); err != nil {
		t.Fatalf("Failed to set to flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries in the 10-3 day range
	if len(result.Entries) != 2 {
		t.Errorf("Expected 2 entries in date range, got %d", len(result.Entries))
	}

	// Verify correct entries
	descriptions := []string{}
	for _, e := range result.Entries {
		descriptions = append(descriptions, e.Description)
	}

	expectedDescriptions := []string{"Week old entry", "Five days old"}
	for _, expected := range expectedDescriptions {
		found := false
		for _, desc := range descriptions {
			if desc == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected entry %q in results", expected)
		}
	}

	// Verify filter criteria in metadata
	if result.Metadata.FilterCriteria["from"] == nil {
		t.Error("Expected 'from' in filter_criteria")
	}
	if result.Metadata.FilterCriteria["to"] == nil {
		t.Error("Expected 'to' in filter_criteria")
	}
}

func TestExportJSON_LastFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with different dates
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10), // 10 days ago (outside range)
			Description:     "Old entry",
			DurationMinutes: 60,
			RawInput:        "Old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -5), // 5 days ago (in range)
			Description:     "Five days old",
			DurationMinutes: 90,
			RawInput:        "Five days old for 1h30m",
		},
		{
			Timestamp:       now.AddDate(0, 0, -2), // 2 days ago (in range)
			Description:     "Two days old",
			DurationMinutes: 45,
			RawInput:        "Two days old for 45m",
		},
		{
			Timestamp:       now, // Today (in range)
			Description:     "Today entry",
			DurationMinutes: 30,
			RawInput:        "Today entry for 30m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Set --last flag to 7 days
	if err := exportJSONCmd.Flags().Set("last", "7"); err != nil {
		t.Fatalf("Failed to set last flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries from last 7 days (3 entries)
	if len(result.Entries) != 3 {
		t.Errorf("Expected 3 entries from last 7 days, got %d", len(result.Entries))
	}

	// Verify filter criteria in metadata shows last_days
	if lastDays, ok := result.Metadata.FilterCriteria["last_days"].(float64); !ok || lastDays != 7 {
		t.Errorf("Expected last_days=7 in filter_criteria, got %v", result.Metadata.FilterCriteria["last_days"])
	}
}

func TestExportJSON_InvalidFromDate(t *testing.T) {
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

	// Set invalid --from date
	if err := exportJSONCmd.Flags().Set("from", "invalid-date"); err != nil {
		t.Fatalf("Failed to set from flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	if !exitCalled {
		t.Error("Expected exit to be called for invalid --from date")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Invalid --from date") {
		t.Errorf("Expected 'Invalid --from date' error, got: %s", stderrOutput)
	}
}

func TestExportJSON_InvalidToDate(t *testing.T) {
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

	// Set invalid --to date
	if err := exportJSONCmd.Flags().Set("to", "2024-13-45"); err != nil {
		t.Fatalf("Failed to set to flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	if !exitCalled {
		t.Error("Expected exit to be called for invalid --to date")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Invalid --to date") {
		t.Errorf("Expected 'Invalid --to date' error, got: %s", stderrOutput)
	}
}

func TestExportJSON_LastWithFromError(t *testing.T) {
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

	// Set both --last and --from (should error)
	if err := exportJSONCmd.Flags().Set("last", "7"); err != nil {
		t.Fatalf("Failed to set last flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("from", "2024-01-01"); err != nil {
		t.Fatalf("Failed to set from flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	if !exitCalled {
		t.Error("Expected exit to be called when using --last with --from")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Cannot use --last with --from") {
		t.Errorf("Expected error about conflicting flags, got: %s", stderrOutput)
	}
}

func TestExportJSON_LastWithToError(t *testing.T) {
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

	// Set both --last and --to (should error)
	if err := exportJSONCmd.Flags().Set("last", "7"); err != nil {
		t.Fatalf("Failed to set last flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", "2024-12-31"); err != nil {
		t.Fatalf("Failed to set to flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	if !exitCalled {
		t.Error("Expected exit to be called when using --last with --to")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Cannot use --last with") {
		t.Errorf("Expected error about conflicting flags, got: %s", stderrOutput)
	}
}

func TestExportJSON_LastWithBothFromAndToError(t *testing.T) {
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

	// Set --last with both --from and --to (should error)
	if err := exportJSONCmd.Flags().Set("last", "7"); err != nil {
		t.Fatalf("Failed to set last flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("from", "2024-01-01"); err != nil {
		t.Fatalf("Failed to set from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", "2024-12-31"); err != nil {
		t.Fatalf("Failed to set to flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	if !exitCalled {
		t.Error("Expected exit to be called when using --last with --from and --to")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Cannot use --last with") {
		t.Errorf("Expected error about conflicting flags, got: %s", stderrOutput)
	}
}

func TestExportJSON_FromOnlyIncludesUpToNow(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10), // 10 days ago (before from)
			Description:     "Old entry",
			DurationMinutes: 60,
			RawInput:        "Old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -3), // 3 days ago (after from)
			Description:     "Recent entry",
			DurationMinutes: 90,
			RawInput:        "Recent entry for 1h30m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Set only --from flag (should include up to now)
	fromDate := now.AddDate(0, 0, -5).Format("2006-01-02")
	if err := exportJSONCmd.Flags().Set("from", fromDate); err != nil {
		t.Fatalf("Failed to set from flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should include entry from 3 days ago but not 10 days ago
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(result.Entries))
	}

	if result.Entries[0].Description != "Recent entry" {
		t.Errorf("Expected 'Recent entry', got %q", result.Entries[0].Description)
	}
}

func TestExportJSON_ToOnlyIncludesFromBeginning(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, -6, 0), // 6 months ago (before to)
			Description:     "Very old entry",
			DurationMinutes: 60,
			RawInput:        "Very old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -2), // 2 days ago (after to)
			Description:     "Recent entry",
			DurationMinutes: 90,
			RawInput:        "Recent entry for 1h30m",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Set only --to flag (should include from beginning)
	toDate := now.AddDate(0, 0, -5).Format("2006-01-02")
	if err := exportJSONCmd.Flags().Set("to", toDate); err != nil {
		t.Fatalf("Failed to set to flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should include old entry but not recent entry
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(result.Entries))
	}

	if result.Entries[0].Description != "Very old entry" {
		t.Errorf("Expected 'Very old entry', got %q", result.Entries[0].Description)
	}
}

func TestExportJSON_EuropeanDateFormat(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Date(2024, 6, 15, 10, 0, 0, 0, time.Local),
		Description:     "Test entry",
		DurationMinutes: 60,
		RawInput:        "Test entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	// Use European date format (DD/MM/YYYY)
	if err := exportJSONCmd.Flags().Set("from", "01/06/2024"); err != nil { // June 1, 2024
		t.Fatalf("Failed to set from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", "30/06/2024"); err != nil { // June 30, 2024
		t.Fatalf("Failed to set to flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should include the entry
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry with European date format, got %d", len(result.Entries))
	}
}

func TestExportJSON_PartialDateError(t *testing.T) {
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

	// Set partial date (missing day)
	if err := exportJSONCmd.Flags().Set("from", "2024-01"); err != nil { // Year-Month only
		t.Fatalf("Failed to set from flag: %v", err)
	}
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	if !exitCalled {
		t.Error("Expected exit to be called for partial date")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Invalid --from date") {
		t.Errorf("Expected invalid date error, got: %s", stderrOutput)
	}
}

// Project and tag filtering tests

func TestExportJSON_ProjectFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)

	// Set --project flag
	if err := rootCmd.PersistentFlags().Set("project", "acme"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries for project 'acme' (1 entry)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry for project 'acme', got %d", len(result.Entries))
	}

	if result.Entries[0].Project != "acme" {
		t.Errorf("Expected project='acme', got %q", result.Entries[0].Project)
	}

	// Verify filter criteria in metadata
	if project, ok := result.Metadata.FilterCriteria["project"].(string); !ok || project != "acme" {
		t.Errorf("Expected project='acme' in filter_criteria, got %v", result.Metadata.FilterCriteria["project"])
	}
}

func TestExportJSON_ProjectShorthand(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)

	// Set project using -p shorthand
	if err := rootCmd.PersistentFlags().Set("project", "client"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries for project 'client' (1 entry)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry for project 'client', got %d", len(result.Entries))
	}

	if result.Entries[0].Project != "client" {
		t.Errorf("Expected project='client', got %q", result.Entries[0].Project)
	}

	// Verify filter criteria in metadata
	if project, ok := result.Metadata.FilterCriteria["project"].(string); !ok || project != "client" {
		t.Errorf("Expected project='client' in filter_criteria, got %v", result.Metadata.FilterCriteria["project"])
	}
}

func TestExportJSON_TagFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)

	// Set --tag flag
	if err := rootCmd.PersistentFlags().Set("tag", "review"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries with tag 'review' (1 entry)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry with tag 'review', got %d", len(result.Entries))
	}

	// Verify the entry has the review tag
	hasReviewTag := false
	for _, tag := range result.Entries[0].Tags {
		if tag == "review" {
			hasReviewTag = true
			break
		}
	}
	if !hasReviewTag {
		t.Errorf("Expected entry to have 'review' tag, got %v", result.Entries[0].Tags)
	}

	// Verify filter criteria in metadata
	tagsInterface := result.Metadata.FilterCriteria["tags"]
	if tagsInterface == nil {
		t.Error("Expected 'tags' in filter_criteria")
	} else {
		tagsSlice, ok := tagsInterface.([]interface{})
		if !ok {
			t.Errorf("Expected tags to be []interface{}, got %T", tagsInterface)
		} else if len(tagsSlice) != 1 {
			t.Errorf("Expected 1 tag in filter_criteria, got %d", len(tagsSlice))
		} else if tagStr, ok := tagsSlice[0].(string); !ok || tagStr != "review" {
			t.Errorf("Expected tag='review' in filter_criteria, got %v", tagsSlice[0])
		}
	}
}

func TestExportJSON_TagShorthand(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)

	// Set tag using -t shorthand
	if err := rootCmd.PersistentFlags().Set("tag", "bugfix"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries with tag 'bugfix' (1 entry)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry with tag 'bugfix', got %d", len(result.Entries))
	}

	// Verify the entry has the bugfix tag
	hasBugfixTag := false
	for _, tag := range result.Entries[0].Tags {
		if tag == "bugfix" {
			hasBugfixTag = true
			break
		}
	}
	if !hasBugfixTag {
		t.Errorf("Expected entry to have 'bugfix' tag, got %v", result.Entries[0].Tags)
	}

	// Verify filter criteria in metadata
	tagsInterface := result.Metadata.FilterCriteria["tags"]
	if tagsInterface == nil {
		t.Error("Expected 'tags' in filter_criteria")
	} else {
		tagsSlice, ok := tagsInterface.([]interface{})
		if !ok {
			t.Errorf("Expected tags to be []interface{}, got %T", tagsInterface)
		} else if len(tagsSlice) != 1 {
			t.Errorf("Expected 1 tag in filter_criteria, got %d", len(tagsSlice))
		} else if tagStr, ok := tagsSlice[0].(string); !ok || tagStr != "bugfix" {
			t.Errorf("Expected tag='bugfix' in filter_criteria, got %v", tagsSlice[0])
		}
	}
}

func TestExportJSON_MultipleTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with multiple tags
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -3),
			Description:     "API work with multiple tags",
			DurationMinutes: 120,
			RawInput:        "API work for 2h",
			Project:         "backend",
			Tags:            []string{"api", "review", "urgent"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -2),
			Description:     "Bug fix with api tag",
			DurationMinutes: 60,
			RawInput:        "Bug fix for 1h",
			Project:         "frontend",
			Tags:            []string{"api", "bugfix"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -1),
			Description:     "Meeting without api tag",
			DurationMinutes: 30,
			RawInput:        "Meeting for 30m",
			Project:         "",
			Tags:            []string{"meeting"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)

	// Set multiple --tag flags
	if err := rootCmd.PersistentFlags().Set("tag", "api"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("tag", "review"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries with BOTH 'api' AND 'review' tags (1 entry)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry with both 'api' and 'review' tags, got %d", len(result.Entries))
	}

	if result.Entries[0].Description != "API work with multiple tags" {
		t.Errorf("Expected 'API work with multiple tags', got %q", result.Entries[0].Description)
	}

	// Verify filter criteria in metadata
	tagsInterface := result.Metadata.FilterCriteria["tags"]
	if tagsInterface == nil {
		t.Error("Expected 'tags' in filter_criteria")
	} else {
		tagsSlice, ok := tagsInterface.([]interface{})
		if !ok {
			t.Errorf("Expected tags to be []interface{}, got %T", tagsInterface)
		} else if len(tagsSlice) != 2 {
			t.Errorf("Expected 2 tags in filter_criteria, got %d", len(tagsSlice))
		}
	}
}

func TestExportJSON_ProjectAndTagCombined(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with various project and tag combinations
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -5),
			Description:     "Code review for acme",
			DurationMinutes: 60,
			RawInput:        "Code review for 1h",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -4),
			Description:     "Bug fix for acme",
			DurationMinutes: 90,
			RawInput:        "Bug fix for 1h30m",
			Project:         "acme",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -3),
			Description:     "Code review for client",
			DurationMinutes: 45,
			RawInput:        "Code review for 45m",
			Project:         "client",
			Tags:            []string{"review"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)

	// Set both --project and --tag flags
	if err := rootCmd.PersistentFlags().Set("project", "acme"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("tag", "review"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include entries for project 'acme' with tag 'review' (1 entry)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry for project 'acme' with tag 'review', got %d", len(result.Entries))
	}

	if result.Entries[0].Description != "Code review for acme" {
		t.Errorf("Expected 'Code review for acme', got %q", result.Entries[0].Description)
	}

	// Verify both filters in metadata
	if project, ok := result.Metadata.FilterCriteria["project"].(string); !ok || project != "acme" {
		t.Errorf("Expected project='acme' in filter_criteria, got %v", result.Metadata.FilterCriteria["project"])
	}

	// Verify tags in filter criteria
	tagsInterface := result.Metadata.FilterCriteria["tags"]
	if tagsInterface == nil {
		t.Error("Expected 'tags' in filter_criteria")
	} else {
		// JSON unmarshals string slices as []interface{}
		tagsSlice, ok := tagsInterface.([]interface{})
		if !ok {
			t.Errorf("Expected tags to be []interface{}, got %T", tagsInterface)
		} else if len(tagsSlice) != 1 {
			t.Errorf("Expected 1 tag in filter_criteria, got %d", len(tagsSlice))
		} else if tagStr, ok := tagsSlice[0].(string); !ok || tagStr != "review" {
			t.Errorf("Expected tag='review' in filter_criteria, got %v", tagsSlice[0])
		}
	}
}

func TestExportJSON_ProjectWithDateFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create entries with different dates and projects
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10), // 10 days ago
			Description:     "Old acme work",
			DurationMinutes: 60,
			RawInput:        "Old acme work for 1h",
			Project:         "acme",
			Tags:            []string{},
		},
		{
			Timestamp:       now.AddDate(0, 0, -3), // 3 days ago
			Description:     "Recent acme work",
			DurationMinutes: 90,
			RawInput:        "Recent acme work for 1h30m",
			Project:         "acme",
			Tags:            []string{},
		},
		{
			Timestamp:       now.AddDate(0, 0, -2), // 2 days ago
			Description:     "Recent client work",
			DurationMinutes: 45,
			RawInput:        "Recent client work for 45m",
			Project:         "client",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)
	if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
		t.Fatalf("Failed to reset last flag: %v", err)
	}

	// Set --project and --last flags
	if err := rootCmd.PersistentFlags().Set("project", "acme"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "7"); err != nil {
		t.Fatalf("Failed to set last flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags
	defer func() {
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only include recent acme work (last 7 days, project acme)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry for project 'acme' in last 7 days, got %d", len(result.Entries))
		for i, e := range result.Entries {
			t.Logf("Entry %d: %s (project=%s)", i, e.Description, e.Project)
		}
	} else {
		if result.Entries[0].Description != "Recent acme work" {
			t.Errorf("Expected 'Recent acme work', got %q", result.Entries[0].Description)
		}
	}

	// Verify both filters in metadata
	if project, ok := result.Metadata.FilterCriteria["project"].(string); !ok || project != "acme" {
		t.Errorf("Expected project='acme' in filter_criteria, got %v", result.Metadata.FilterCriteria["project"])
	}

	if lastDays, ok := result.Metadata.FilterCriteria["last_days"].(float64); !ok || lastDays != 7 {
		t.Errorf("Expected last_days=7 in filter_criteria, got %v", result.Metadata.FilterCriteria["last_days"])
	}
}

func TestExportJSON_NoMatchingProject(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)

	// Set --project flag for non-existent project
	if err := rootCmd.PersistentFlags().Set("project", "nonexistent"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should return empty entries array
	if len(result.Entries) != 0 {
		t.Errorf("Expected 0 entries for non-existent project, got %d", len(result.Entries))
	}

	// Verify metadata still shows the filter
	if project, ok := result.Metadata.FilterCriteria["project"].(string); !ok || project != "nonexistent" {
		t.Errorf("Expected project='nonexistent' in filter_criteria, got %v", result.Metadata.FilterCriteria["project"])
	}

	if result.Metadata.TotalEntries != 0 {
		t.Errorf("Expected total_entries=0, got %d", result.Metadata.TotalEntries)
	}
}

func TestExportJSON_NoMatchingTag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	// Reset flags first to ensure clean state
	resetFilterFlags(exportJSONCmd)

	// Set --tag flag for non-existent tag
	if err := rootCmd.PersistentFlags().Set("tag", "nonexistent"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd) // Reset flags

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should return empty entries array
	if len(result.Entries) != 0 {
		t.Errorf("Expected 0 entries for non-existent tag, got %d", len(result.Entries))
	}

	// Verify metadata still shows the filter
	tagsInterface := result.Metadata.FilterCriteria["tags"]
	if tagsInterface == nil {
		t.Error("Expected 'tags' in filter_criteria")
	} else {
		tagsSlice, ok := tagsInterface.([]interface{})
		if !ok {
			t.Errorf("Expected tags to be []interface{}, got %T", tagsInterface)
		} else if len(tagsSlice) != 1 {
			t.Errorf("Expected 1 tag in filter_criteria, got %d", len(tagsSlice))
		} else if tagStr, ok := tagsSlice[0].(string); !ok || tagStr != "nonexistent" {
			t.Errorf("Expected tag='nonexistent' in filter_criteria, got %v", tagsSlice[0])
		}
	}

	if result.Metadata.TotalEntries != 0 {
		t.Errorf("Expected total_entries=0, got %d", result.Metadata.TotalEntries)
	}
}

// Integration tests combining multiple filter types

func TestExportJSON_Integration_ProjectTagAndDateRange(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create diverse test data
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -15),
			Description:     "Old acme review",
			DurationMinutes: 60,
			RawInput:        "Old acme review for 1h",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -7),
			Description:     "Recent acme review",
			DurationMinutes: 90,
			RawInput:        "Recent acme review for 1h30m",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -6),
			Description:     "Recent acme bugfix",
			DurationMinutes: 45,
			RawInput:        "Recent acme bugfix for 45m",
			Project:         "acme",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -5),
			Description:     "Recent client review",
			DurationMinutes: 30,
			RawInput:        "Recent client review for 30m",
			Project:         "client",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -2),
			Description:     "Very recent acme review",
			DurationMinutes: 120,
			RawInput:        "Very recent acme review for 2h",
			Project:         "acme",
			Tags:            []string{"review"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	// Reset flags first
	resetFilterFlags(exportJSONCmd)
	if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
		t.Fatalf("Failed to reset from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
		t.Fatalf("Failed to reset to flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
		t.Fatalf("Failed to reset last flag: %v", err)
	}

	// Set project='acme', tag='review', and date range (10 days ago to 3 days ago)
	if err := rootCmd.PersistentFlags().Set("project", "acme"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("tag", "review"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	fromDate := now.AddDate(0, 0, -10).Format("2006-01-02")
	toDate := now.AddDate(0, 0, -3).Format("2006-01-02")
	if err := exportJSONCmd.Flags().Set("from", fromDate); err != nil {
		t.Fatalf("Failed to set from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", toDate); err != nil {
		t.Fatalf("Failed to set to flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd)
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, stdout.String())
	}

	// Should only include "Recent acme review" (7 days ago, project=acme, tag=review, within date range)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry matching all filters, got %d", len(result.Entries))
		for i, e := range result.Entries {
			t.Logf("Entry %d: %s (project=%s, tags=%v, timestamp=%v)", i, e.Description, e.Project, e.Tags, e.Timestamp)
		}
	}

	if len(result.Entries) > 0 && result.Entries[0].Description != "Recent acme review" {
		t.Errorf("Expected 'Recent acme review', got %q", result.Entries[0].Description)
	}

	// Verify complete metadata structure
	if result.Metadata.TotalEntries != 1 {
		t.Errorf("Expected total_entries=1, got %d", result.Metadata.TotalEntries)
	}

	// Verify all filter criteria are present
	if project, ok := result.Metadata.FilterCriteria["project"].(string); !ok || project != "acme" {
		t.Errorf("Expected project='acme' in filter_criteria, got %v", result.Metadata.FilterCriteria["project"])
	}

	tagsInterface := result.Metadata.FilterCriteria["tags"]
	if tagsInterface == nil {
		t.Error("Expected 'tags' in filter_criteria")
	} else {
		tagsSlice, ok := tagsInterface.([]interface{})
		if !ok || len(tagsSlice) != 1 || tagsSlice[0] != "review" {
			t.Errorf("Expected tags=['review'] in filter_criteria, got %v", tagsInterface)
		}
	}

	if result.Metadata.FilterCriteria["from"] == nil {
		t.Error("Expected 'from' in filter_criteria")
	}

	if result.Metadata.FilterCriteria["to"] == nil {
		t.Error("Expected 'to' in filter_criteria")
	}

	// Verify timestamp is recent
	if time.Since(result.Metadata.ExportTimestamp) > time.Minute {
		t.Errorf("Export timestamp is not recent: %v", result.Metadata.ExportTimestamp)
	}
}

func TestExportJSON_Integration_MultipleTagsWithLast(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -20),
			Description:     "Old entry with both tags",
			DurationMinutes: 60,
			RawInput:        "Old entry for 1h",
			Project:         "backend",
			Tags:            []string{"api", "urgent"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -5),
			Description:     "Recent entry with api only",
			DurationMinutes: 90,
			RawInput:        "Recent entry for 1h30m",
			Project:         "backend",
			Tags:            []string{"api"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -3),
			Description:     "Recent entry with both tags",
			DurationMinutes: 120,
			RawInput:        "Recent entry for 2h",
			Project:         "backend",
			Tags:            []string{"api", "urgent"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -1),
			Description:     "Very recent with urgent only",
			DurationMinutes: 45,
			RawInput:        "Very recent for 45m",
			Project:         "backend",
			Tags:            []string{"urgent"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	resetFilterFlags(exportJSONCmd)
	if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
		t.Fatalf("Failed to reset from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
		t.Fatalf("Failed to reset to flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
		t.Fatalf("Failed to reset last flag: %v", err)
	}

	// Set multiple tags and last 7 days
	if err := rootCmd.PersistentFlags().Set("tag", "api"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("tag", "urgent"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "7"); err != nil {
		t.Fatalf("Failed to set last flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd)
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nStderr: %s\nStdout: %s", err, stderr.String(), stdout.String())
	}

	// Should only include entry with both tags in last 7 days
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry with both tags in last 7 days, got %d", len(result.Entries))
	}

	if len(result.Entries) > 0 && result.Entries[0].Description != "Recent entry with both tags" {
		t.Errorf("Expected 'Recent entry with both tags', got %q", result.Entries[0].Description)
	}

	// Verify metadata
	if result.Metadata.TotalEntries != 1 {
		t.Errorf("Expected total_entries=1, got %d", result.Metadata.TotalEntries)
	}

	tagsInterface := result.Metadata.FilterCriteria["tags"]
	if tagsSlice, ok := tagsInterface.([]interface{}); !ok || len(tagsSlice) != 2 {
		t.Errorf("Expected 2 tags in filter_criteria, got %v", tagsInterface)
	}

	if lastDays, ok := result.Metadata.FilterCriteria["last_days"].(float64); !ok || lastDays != 7 {
		t.Errorf("Expected last_days=7 in filter_criteria, got %v", result.Metadata.FilterCriteria["last_days"])
	}
}

func TestExportJSON_Integration_AllFiltersWithNoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -5),
			Description:     "Acme review entry",
			DurationMinutes: 60,
			RawInput:        "Acme review for 1h",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -3),
			Description:     "Client bugfix entry",
			DurationMinutes: 90,
			RawInput:        "Client bugfix for 1h30m",
			Project:         "client",
			Tags:            []string{"bugfix"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	resetFilterFlags(exportJSONCmd)
	if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
		t.Fatalf("Failed to reset from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
		t.Fatalf("Failed to reset to flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
		t.Fatalf("Failed to reset last flag: %v", err)
	}

	// Set filters that won't match any entry (project=acme, tag=bugfix)
	if err := rootCmd.PersistentFlags().Set("project", "acme"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("tag", "bugfix"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "7"); err != nil {
		t.Fatalf("Failed to set last flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd)
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nStderr: %s", err, stderr.String())
	}

	// Should return empty results
	if len(result.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(result.Entries))
	}

	if result.Metadata.TotalEntries != 0 {
		t.Errorf("Expected total_entries=0, got %d", result.Metadata.TotalEntries)
	}

	// Verify all filters are still in metadata
	if result.Metadata.FilterCriteria["project"] != "acme" {
		t.Errorf("Expected project='acme' in filter_criteria")
	}

	if result.Metadata.FilterCriteria["tags"] == nil {
		t.Error("Expected 'tags' in filter_criteria")
	}

	if result.Metadata.FilterCriteria["last_days"] == nil {
		t.Error("Expected 'last_days' in filter_criteria")
	}

	// Verify JSON structure is still valid with empty entries
	if result.Entries == nil {
		t.Error("Expected entries array to be initialized (not nil)")
	}
}

func TestExportJSON_Integration_ProjectTagAndFrom(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -30),
			Description:     "Very old matching entry",
			DurationMinutes: 60,
			RawInput:        "Very old for 1h",
			Project:         "frontend",
			Tags:            []string{"feature"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -8),
			Description:     "Recent matching entry",
			DurationMinutes: 90,
			RawInput:        "Recent for 1h30m",
			Project:         "frontend",
			Tags:            []string{"feature"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -7),
			Description:     "Recent non-matching project",
			DurationMinutes: 45,
			RawInput:        "Recent for 45m",
			Project:         "backend",
			Tags:            []string{"feature"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -6),
			Description:     "Recent non-matching tag",
			DurationMinutes: 30,
			RawInput:        "Recent for 30m",
			Project:         "frontend",
			Tags:            []string{"bugfix"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	resetFilterFlags(exportJSONCmd)
	if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
		t.Fatalf("Failed to reset from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
		t.Fatalf("Failed to reset to flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
		t.Fatalf("Failed to reset last flag: %v", err)
	}

	// Filter: project=frontend, tag=feature, from=10 days ago
	if err := rootCmd.PersistentFlags().Set("project", "frontend"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("tag", "feature"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	fromDate := now.AddDate(0, 0, -10).Format("2006-01-02")
	if err := exportJSONCmd.Flags().Set("from", fromDate); err != nil {
		t.Fatalf("Failed to set from flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd)
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nStderr: %s", err, stderr.String())
	}

	// Should only include "Recent matching entry" (within date range, matching project and tag)
	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(result.Entries))
		for i, e := range result.Entries {
			t.Logf("Entry %d: %s", i, e.Description)
		}
	}

	if len(result.Entries) > 0 && result.Entries[0].Description != "Recent matching entry" {
		t.Errorf("Expected 'Recent matching entry', got %q", result.Entries[0].Description)
	}

	// Verify complete JSON output structure
	if result.Metadata.TotalEntries != 1 {
		t.Errorf("Expected total_entries=1, got %d", result.Metadata.TotalEntries)
	}

	// Check all filter criteria are present
	if result.Metadata.FilterCriteria["project"] != "frontend" {
		t.Errorf("Expected project='frontend' in filter_criteria")
	}

	if result.Metadata.FilterCriteria["tags"] == nil {
		t.Error("Expected 'tags' in filter_criteria")
	}

	if result.Metadata.FilterCriteria["from"] == nil {
		t.Error("Expected 'from' in filter_criteria")
	}

	// Verify entries have complete fields
	if len(result.Entries) > 0 {
		e := result.Entries[0]
		if e.Description == "" {
			t.Error("Expected entry to have description")
		}
		if e.DurationMinutes == 0 {
			t.Error("Expected entry to have duration")
		}
		if e.RawInput == "" {
			t.Error("Expected entry to have raw_input")
		}
		if e.Project == "" {
			t.Error("Expected entry to have project")
		}
		if len(e.Tags) == 0 {
			t.Error("Expected entry to have tags")
		}
	}
}

func TestExportJSON_Integration_ComplexFilteringWithMultipleMatches(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -8),
			Description:     "Backend API feature 1",
			DurationMinutes: 120,
			RawInput:        "Backend API feature 1 for 2h",
			Project:         "backend",
			Tags:            []string{"api", "feature"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -6),
			Description:     "Backend API feature 2",
			DurationMinutes: 90,
			RawInput:        "Backend API feature 2 for 1h30m",
			Project:         "backend",
			Tags:            []string{"api", "feature"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -4),
			Description:     "Backend API feature 3",
			DurationMinutes: 60,
			RawInput:        "Backend API feature 3 for 1h",
			Project:         "backend",
			Tags:            []string{"api", "feature"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -3),
			Description:     "Frontend API feature",
			DurationMinutes: 45,
			RawInput:        "Frontend API feature for 45m",
			Project:         "frontend",
			Tags:            []string{"api", "feature"},
		},
		{
			Timestamp:       now.AddDate(0, 0, -2),
			Description:     "Backend API bugfix",
			DurationMinutes: 30,
			RawInput:        "Backend API bugfix for 30m",
			Project:         "backend",
			Tags:            []string{"api", "bugfix"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	resetFilterFlags(exportJSONCmd)
	if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
		t.Fatalf("Failed to reset from flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
		t.Fatalf("Failed to reset to flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
		t.Fatalf("Failed to reset last flag: %v", err)
	}

	// Filter: project=backend, tags=[api,feature], last 7 days
	if err := rootCmd.PersistentFlags().Set("project", "backend"); err != nil {
		t.Fatalf("Failed to set project flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("tag", "api"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("tag", "feature"); err != nil {
		t.Fatalf("Failed to set tag flag: %v", err)
	}
	if err := exportJSONCmd.Flags().Set("last", "7"); err != nil {
		t.Fatalf("Failed to set last flag: %v", err)
	}
	defer resetFilterFlags(exportJSONCmd)
	defer func() {
		if err := exportJSONCmd.Flags().Set("from", ""); err != nil {
			t.Errorf("Failed to reset from flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("to", ""); err != nil {
			t.Errorf("Failed to reset to flag: %v", err)
		}
		if err := exportJSONCmd.Flags().Set("last", "0"); err != nil {
			t.Errorf("Failed to reset last flag: %v", err)
		}
	}()

	exportJSON(exportJSONCmd)

	var result ExportOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nStderr: %s", err, stderr.String())
	}

	// Should include 2 backend entries with both api and feature tags from last 7 days
	if len(result.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result.Entries))
		for i, e := range result.Entries {
			t.Logf("Entry %d: %s (project=%s, tags=%v)", i, e.Description, e.Project, e.Tags)
		}
	}

	if result.Metadata.TotalEntries != 2 {
		t.Errorf("Expected total_entries=2, got %d", result.Metadata.TotalEntries)
	}

	// Verify entries are correctly ordered (most recent first)
	if len(result.Entries) == 2 {
		if result.Entries[0].Description != "Backend API feature 2" {
			t.Errorf("Expected first entry to be 'Backend API feature 2', got %q", result.Entries[0].Description)
		}
		if result.Entries[1].Description != "Backend API feature 3" {
			t.Errorf("Expected second entry to be 'Backend API feature 3', got %q", result.Entries[1].Description)
	}
	}

	// Verify complete metadata
	if result.Metadata.FilterCriteria["project"] != "backend" {
		t.Error("Expected 'project' in filter_criteria")
	}

	tagsInterface := result.Metadata.FilterCriteria["tags"]
	if tagsSlice, ok := tagsInterface.([]interface{}); !ok || len(tagsSlice) != 2 {
		t.Errorf("Expected 2 tags in filter_criteria, got %v", tagsInterface)
	}

	if lastDays, ok := result.Metadata.FilterCriteria["last_days"].(float64); !ok || lastDays != 7 {
		t.Errorf("Expected last_days=7 in filter_criteria, got %v", result.Metadata.FilterCriteria["last_days"])
	}

	// Verify all entries have complete data
	for i, e := range result.Entries {
		if e.Timestamp.IsZero() {
			t.Errorf("Entry %d has zero timestamp", i)
		}
		if e.Description == "" {
			t.Errorf("Entry %d has empty description", i)
		}
		if e.DurationMinutes == 0 {
			t.Errorf("Entry %d has zero duration", i)
		}
		if e.RawInput == "" {
			t.Errorf("Entry %d has empty raw_input", i)
		}
		if e.Project != "backend" {
			t.Errorf("Entry %d has wrong project: %s", i, e.Project)
		}
		if len(e.Tags) < 2 {
			t.Errorf("Entry %d has insufficient tags: %v", i, e.Tags)
		}
	}
}

// CSV Export Tests

func TestExportCSV_BasicOutput(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

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

	exportCSV(exportCSVCmd)

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header + 3 entries
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines (header + 3 entries), got %d", len(lines))
	}

	// Verify header
	expectedHeader := "date,description,duration_minutes,duration_hours,project,tags"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header:\n%s\nGot:\n%s", expectedHeader, lines[0])
	}

	// Verify data rows contain expected fields
	for i := 1; i < len(lines); i++ {
		fields := strings.Split(lines[i], ",")
		if len(fields) < 6 {
			t.Errorf("Line %d has insufficient fields: %d", i, len(fields))
		}
	}
}

func TestExportCSV_HeaderStructure(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a single entry
	now := time.Now()
	testEntry := entry.Entry{
		Timestamp:       now,
		Description:     "Test entry",
		DurationMinutes: 60,
		RawInput:        "Test entry for 1h",
		Project:         "test",
		Tags:            []string{"tag1"},
	}

	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	exportCSV(exportCSVCmd)

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check header
	if len(lines) < 1 {
		t.Fatal("No output generated")
	}

	header := lines[0]
	expectedColumns := []string{"date", "description", "duration_minutes", "duration_hours", "project", "tags"}

	for _, col := range expectedColumns {
		if !strings.Contains(header, col) {
			t.Errorf("Header missing column: %s", col)
		}
	}
}

func TestExportCSV_RFC4180_Escaping_Commas(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	testEntry := entry.Entry{
		Timestamp:       now,
		Description:     "Meeting with client, discussed features, and pricing",
		DurationMinutes: 60,
		RawInput:        "Meeting with client, discussed features, and pricing for 1h",
		Project:         "acme",
		Tags:            []string{"meeting"},
	}

	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Description with commas should be quoted in RFC 4180
	if !strings.Contains(output, `"Meeting with client, discussed features, and pricing"`) {
		t.Errorf("Description with commas should be quoted. Output:\n%s", output)
	}
}

func TestExportCSV_RFC4180_Escaping_Quotes(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	testEntry := entry.Entry{
		Timestamp:       now,
		Description:     `Client said "looks great" during review`,
		DurationMinutes: 30,
		RawInput:        `Client said "looks great" during review for 30m`,
		Project:         "acme",
		Tags:            []string{},
	}

	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Description with quotes should be escaped (doubled) and the field quoted
	if !strings.Contains(output, `"Client said ""looks great"" during review"`) {
		t.Errorf("Description with quotes should be properly escaped. Output:\n%s", output)
	}
}

func TestExportCSV_RFC4180_Escaping_Newlines(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	testEntry := entry.Entry{
		Timestamp:       now,
		Description:     "Multi-line description\nwith newline\ncharacters",
		DurationMinutes: 45,
		RawInput:        "Multi-line description for 45m",
		Project:         "test",
		Tags:            []string{},
	}

	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Description with newlines should be quoted and preserved
	if !strings.Contains(output, "\"Multi-line description\nwith newline\ncharacters\"") {
		t.Errorf("Description with newlines should be properly quoted. Output:\n%s", output)
	}
}

func TestExportCSV_EmptyResults(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create empty storage file
	if err := os.WriteFile(storagePath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty storage file: %v", err)
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

	exportCSV(exportCSVCmd)

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should only have header row
	if len(lines) != 1 {
		t.Errorf("Expected 1 line (header only), got %d", len(lines))
	}

	// Verify header is present
	expectedHeader := "date,description,duration_minutes,duration_hours,project,tags"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header:\n%s\nGot:\n%s", expectedHeader, lines[0])
	}
}

func TestExportCSV_DurationHoursCalculation(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	testCases := []struct {
		minutes       int
		expectedHours string
	}{
		{60, "1.00"},
		{90, "1.50"},
		{45, "0.75"},
		{30, "0.50"},
		{120, "2.00"},
		{75, "1.25"},
	}

	for _, tc := range testCases {
		testEntry := entry.Entry{
			Timestamp:       now,
			Description:     fmt.Sprintf("Entry with %d minutes", tc.minutes),
			DurationMinutes: tc.minutes,
			RawInput:        fmt.Sprintf("Entry for %dm", tc.minutes),
			Project:         "",
			Tags:            []string{},
		}

		if err := storage.AppendEntry(storagePath, testEntry); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
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

	exportCSV(exportCSVCmd)

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Skip header, check each data row
	for i, tc := range testCases {
		lineIdx := i + 1
		if lineIdx >= len(lines) {
			t.Fatalf("Missing line %d for test case %d minutes", lineIdx, tc.minutes)
		}

		if !strings.Contains(lines[lineIdx], tc.expectedHours) {
			t.Errorf("Expected duration_hours %s for %d minutes, line:\n%s",
				tc.expectedHours, tc.minutes, lines[lineIdx])
		}
	}
}

func TestExportCSV_CorruptedEntriesWarning(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Write one valid entry and one corrupted entry
	validEntry := entry.Entry{
		Timestamp:       time.Now(),
		Description:     "Valid entry",
		DurationMinutes: 60,
		RawInput:        "Valid entry for 1h",
		Project:         "",
		Tags:            []string{},
	}

	if err := storage.AppendEntry(storagePath, validEntry); err != nil {
		t.Fatalf("Failed to create valid entry: %v", err)
	}

	// Append corrupted line
	f, err := os.OpenFile(storagePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open storage file: %v", err)
	}
	_, _ = f.WriteString("{invalid json}\n")
	f.Close()

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

	exportCSV(exportCSVCmd)

	// Check that warning was written to stderr
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Warning") {
		t.Error("Expected warning about corrupted entry in stderr")
	}
	if !strings.Contains(stderrOutput, "corrupted") {
		t.Error("Expected 'corrupted' in stderr warning")
	}

	// Check that valid entry was still exported
	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Valid entry") {
		t.Error("Expected valid entry to be exported despite corruption")
	}
}

func TestExportCSV_FromFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10),
			Description:     "Old entry",
			DurationMinutes: 60,
			RawInput:        "Old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -3),
			Description:     "Recent entry",
			DurationMinutes: 60,
			RawInput:        "Recent entry for 1h",
		},
		{
			Timestamp:       now,
			Description:     "Today entry",
			DurationMinutes: 60,
			RawInput:        "Today entry for 1h",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Set --from flag to 5 days ago
	fromDate := now.AddDate(0, 0, -5).Format("2006-01-02")
	exportCSVCmd.Flags().Set("from", fromDate)
	defer exportCSVCmd.Flags().Set("from", "")

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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Should NOT include old entry
	if strings.Contains(output, "Old entry") {
		t.Error("Should not include entry from before --from date")
	}

	// Should include recent and today entries
	if !strings.Contains(output, "Recent entry") {
		t.Error("Should include recent entry")
	}
	if !strings.Contains(output, "Today entry") {
		t.Error("Should include today entry")
	}
}

func TestExportCSV_ToFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10),
			Description:     "Old entry",
			DurationMinutes: 60,
			RawInput:        "Old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -5),
			Description:     "Middle entry",
			DurationMinutes: 60,
			RawInput:        "Middle entry for 1h",
		},
		{
			Timestamp:       now,
			Description:     "Today entry",
			DurationMinutes: 60,
			RawInput:        "Today entry for 1h",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Set --to flag to 7 days ago
	toDate := now.AddDate(0, 0, -7).Format("2006-01-02")
	exportCSVCmd.Flags().Set("to", toDate)
	defer exportCSVCmd.Flags().Set("to", "")

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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Should include old entry
	if !strings.Contains(output, "Old entry") {
		t.Error("Should include entry from before --to date")
	}

	// Should NOT include middle or today entries
	if strings.Contains(output, "Middle entry") {
		t.Error("Should not include entry after --to date")
	}
	if strings.Contains(output, "Today entry") {
		t.Error("Should not include entry after --to date")
	}
}

func TestExportCSV_FromAndToFlags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -15),
			Description:     "Very old entry",
			DurationMinutes: 60,
			RawInput:        "Very old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -8),
			Description:     "In range entry",
			DurationMinutes: 60,
			RawInput:        "In range entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -2),
			Description:     "Recent entry",
			DurationMinutes: 60,
			RawInput:        "Recent entry for 1h",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Set date range: 10 days ago to 5 days ago
	fromDate := now.AddDate(0, 0, -10).Format("2006-01-02")
	toDate := now.AddDate(0, 0, -5).Format("2006-01-02")
	exportCSVCmd.Flags().Set("from", fromDate)
	exportCSVCmd.Flags().Set("to", toDate)
	defer func() {
		exportCSVCmd.Flags().Set("from", "")
		exportCSVCmd.Flags().Set("to", "")
	}()

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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Should only include entry in range
	if strings.Contains(output, "Very old entry") {
		t.Error("Should not include entry before --from date")
	}
	if !strings.Contains(output, "In range entry") {
		t.Error("Should include entry within date range")
	}
	if strings.Contains(output, "Recent entry") {
		t.Error("Should not include entry after --to date")
	}
}

func TestExportCSV_LastFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -10),
			Description:     "Old entry",
			DurationMinutes: 60,
			RawInput:        "Old entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -5),
			Description:     "Last week entry",
			DurationMinutes: 60,
			RawInput:        "Last week entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -2),
			Description:     "Recent entry",
			DurationMinutes: 60,
			RawInput:        "Recent entry for 1h",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Set --last flag to 7 days
	exportCSVCmd.Flags().Set("last", "7")
	defer exportCSVCmd.Flags().Set("last", "")

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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Should NOT include old entry (10 days ago)
	if strings.Contains(output, "Old entry") {
		t.Error("Should not include entry older than --last days")
	}

	// Should include entries within last 7 days
	if !strings.Contains(output, "Last week entry") {
		t.Error("Should include entry within --last days")
	}
	if !strings.Contains(output, "Recent entry") {
		t.Error("Should include recent entry within --last days")
	}
}

func TestExportCSV_ProjectFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "Acme work",
			DurationMinutes: 60,
			RawInput:        "Acme work for 1h",
			Project:         "acme",
		},
		{
			Timestamp:       now,
			Description:     "Client work",
			DurationMinutes: 60,
			RawInput:        "Client work for 1h",
			Project:         "client",
		},
		{
			Timestamp:       now,
			Description:     "No project work",
			DurationMinutes: 60,
			RawInput:        "No project work for 1h",
			Project:         "",
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Reset flags first to ensure clean state
	resetFilterFlags(exportCSVCmd)

	// Set --project flag
	exportCSVCmd.Root().PersistentFlags().Set("project", "acme")
	defer resetFilterFlags(exportCSVCmd)

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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Should only include acme project
	if !strings.Contains(output, "Acme work") {
		t.Error("Should include acme project entry")
	}
	if strings.Contains(output, "Client work") {
		t.Error("Should not include client project entry")
	}
	if strings.Contains(output, "No project work") {
		t.Error("Should not include entry without project")
	}
}

func TestExportCSV_TagFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "Code review task",
			DurationMinutes: 60,
			RawInput:        "Code review task for 1h",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now,
			Description:     "Bug fix task",
			DurationMinutes: 60,
			RawInput:        "Bug fix task for 1h",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now,
			Description:     "No tags task",
			DurationMinutes: 60,
			RawInput:        "No tags task for 1h",
			Tags:            []string{},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Reset flags first to ensure clean state
	resetFilterFlags(exportCSVCmd)

	// Set --tag flag
	exportCSVCmd.Root().PersistentFlags().Set("tag", "review")
	defer resetFilterFlags(exportCSVCmd)

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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Should only include review tag entry
	if !strings.Contains(output, "Code review task") {
		t.Error("Should include entry with review tag")
	}
	if strings.Contains(output, "Bug fix task") {
		t.Error("Should not include entry with different tag")
	}
	if strings.Contains(output, "No tags task") {
		t.Error("Should not include entry without tags")
	}
}

func TestExportCSV_TagsSemicolonSeparated(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Reset any persistent flags from previous tests
	resetFilterFlags(exportCSVCmd)

	now := time.Now()
	testEntry := entry.Entry{
		Timestamp:       now,
		Description:     "Multi-tag entry",
		DurationMinutes: 60,
		RawInput:        "Multi-tag entry for 1h",
		Project:         "",
		Tags:            []string{"tag1", "tag2", "tag3"},
	}

	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Tags should be semicolon-separated
	if !strings.Contains(output, "tag1;tag2;tag3") {
		t.Errorf("Expected semicolon-separated tags. Output:\n%s", output)
	}
}

func TestExportCSV_ProjectAndTagCombined(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now,
			Description:     "Acme review",
			DurationMinutes: 60,
			RawInput:        "Acme review for 1h",
			Project:         "acme",
			Tags:            []string{"review"},
		},
		{
			Timestamp:       now,
			Description:     "Acme bugfix",
			DurationMinutes: 60,
			RawInput:        "Acme bugfix for 1h",
			Project:         "acme",
			Tags:            []string{"bugfix"},
		},
		{
			Timestamp:       now,
			Description:     "Client review",
			DurationMinutes: 60,
			RawInput:        "Client review for 1h",
			Project:         "client",
			Tags:            []string{"review"},
		},
	}

	for _, e := range entries {
		if err := storage.AppendEntry(storagePath, e); err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Reset flags first to ensure clean state
	resetFilterFlags(exportCSVCmd)

	// Set both --project and --tag flags
	exportCSVCmd.Root().PersistentFlags().Set("project", "acme")
	exportCSVCmd.Root().PersistentFlags().Set("tag", "review")
	defer resetFilterFlags(exportCSVCmd)

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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Should only include entry matching both project and tag
	if !strings.Contains(output, "Acme review") {
		t.Error("Should include entry matching both project and tag")
	}
	if strings.Contains(output, "Acme bugfix") {
		t.Error("Should not include entry with wrong tag")
	}
	if strings.Contains(output, "Client review") {
		t.Error("Should not include entry with wrong project")
	}
}

func TestExportCSV_OutputCanBeRedirected(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	createExportTestEntries(t, storagePath)

	// Reset any persistent flags from previous tests
	resetFilterFlags(exportCSVCmd)

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

	exportCSV(exportCSVCmd)

	output := stdout.String()

	// Verify output can be written to a file (simulate redirection)
	outputFile := filepath.Join(tmpDir, "export.csv")
	if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
		t.Fatalf("Failed to write output to file: %v", err)
	}

	// Verify file contains expected content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "date,description,duration_minutes") {
		t.Error("Output file should contain CSV header")
	}
	if !strings.Contains(contentStr, "Code review") {
		t.Error("Output file should contain entry data")
	}
}
