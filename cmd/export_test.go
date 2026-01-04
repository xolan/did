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
	f.Close()
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
