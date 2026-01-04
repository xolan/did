package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
)

func TestHandleFromCommand_ValidDates(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entries in date range
	entries := []entry.Entry{
		{
			Timestamp:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			Description:     "entry in range",
			DurationMinutes: 60,
			RawInput:        "entry in range for 1h",
		},
		{
			Timestamp:       time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC),
			Description:     "another entry",
			DurationMinutes: 120,
			RawInput:        "another entry for 2h",
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

	// Test ISO format: from 2024-01-01 to 2024-01-31
	handleFromCommand(fromCmd, []string{"2024-01-01", "to", "2024-01-31"})

	output := stdout.String()
	if !strings.Contains(output, "entry in range") {
		t.Errorf("Expected 'entry in range' in output, got: %s", output)
	}
	if !strings.Contains(output, "another entry") {
		t.Errorf("Expected 'another entry' in output, got: %s", output)
	}
}

func TestHandleFromCommand_ValidDatesEuropeanFormat(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Description:     "test entry",
		DurationMinutes: 60,
		RawInput:        "test entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	// Test European format: from 01/01/2024 to 31/01/2024
	handleFromCommand(fromCmd, []string{"01/01/2024", "to", "31/01/2024"})

	output := stdout.String()
	if !strings.Contains(output, "test entry") {
		t.Errorf("Expected 'test entry' in output, got: %s", output)
	}
}

func TestHandleFromCommand_SingleDay(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry on specific day
	testEntry := entry.Entry{
		Timestamp:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Description:     "single day entry",
		DurationMinutes: 60,
		RawInput:        "single day entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	// Test single day range: from 2024-01-15 to 2024-01-15
	handleFromCommand(fromCmd, []string{"2024-01-15", "to", "2024-01-15"})

	output := stdout.String()
	if !strings.Contains(output, "single day entry") {
		t.Errorf("Expected 'single day entry' in output, got: %s", output)
	}
	// Should show single date format in period description
	if !strings.Contains(output, "Mon, Jan 15, 2024") {
		t.Errorf("Expected single date format 'Mon, Jan 15, 2024' in output, got: %s", output)
	}
}

func TestHandleFromCommand_MissingToKeyword(t *testing.T) {
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

	// Missing 'to' keyword
	handleFromCommand(fromCmd, []string{"2024-01-01", "2024-01-31"})

	if !exitCalled {
		t.Error("Expected exit to be called for missing 'to' keyword")
	}
	if !strings.Contains(stderr.String(), "Missing 'to' keyword") {
		t.Errorf("Expected 'Missing 'to' keyword' error, got: %s", stderr.String())
	}
}

func TestHandleFromCommand_MissingStartDate(t *testing.T) {
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

	// 'to' keyword is first argument (missing start date)
	handleFromCommand(fromCmd, []string{"to", "2024-01-31"})

	if !exitCalled {
		t.Error("Expected exit to be called for missing start date")
	}
	if !strings.Contains(stderr.String(), "Missing start date") {
		t.Errorf("Expected 'Missing start date' error, got: %s", stderr.String())
	}
}

func TestHandleFromCommand_MissingEndDate(t *testing.T) {
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

	// 'to' keyword is last argument (missing end date)
	handleFromCommand(fromCmd, []string{"2024-01-01", "to"})

	if !exitCalled {
		t.Error("Expected exit to be called for missing end date")
	}
	if !strings.Contains(stderr.String(), "Missing end date") {
		t.Errorf("Expected 'Missing end date' error, got: %s", stderr.String())
	}
}

func TestHandleFromCommand_InvalidStartDate(t *testing.T) {
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

	// Invalid start date format
	handleFromCommand(fromCmd, []string{"invalid-date", "to", "2024-01-31"})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid start date")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error message for invalid start date, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "Supported date formats:") {
		t.Errorf("Expected format guidance in error, got: %s", errOutput)
	}
}

func TestHandleFromCommand_InvalidEndDate(t *testing.T) {
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

	// Invalid end date format
	handleFromCommand(fromCmd, []string{"2024-01-01", "to", "invalid-date"})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid end date")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error message for invalid end date, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "Supported date formats:") {
		t.Errorf("Expected format guidance in error, got: %s", errOutput)
	}
}

func TestHandleFromCommand_StartAfterEnd(t *testing.T) {
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

	// Start date after end date
	handleFromCommand(fromCmd, []string{"2024-01-31", "to", "2024-01-01"})

	if !exitCalled {
		t.Error("Expected exit to be called when start date is after end date")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Start date") || !strings.Contains(errOutput, "after end date") {
		t.Errorf("Expected start/end date validation error, got: %s", errOutput)
	}
}

func TestHandleFromCommand_PartialStartDate(t *testing.T) {
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

	// Partial date (year-month only)
	handleFromCommand(fromCmd, []string{"2024-01", "to", "2024-01-31"})

	if !exitCalled {
		t.Error("Expected exit to be called for partial date")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error for partial date, got: %s", errOutput)
	}
	// Should show helpful error about missing day
	if !strings.Contains(errOutput, "missing day") {
		t.Errorf("Expected 'missing day' guidance in error, got: %s", errOutput)
	}
}

func TestHandleFromCommand_EmptyDate(t *testing.T) {
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

	// Empty start date (space only, will be trimmed)
	handleFromCommand(fromCmd, []string{"", "to", "2024-01-31"})

	if !exitCalled {
		t.Error("Expected exit to be called for empty date")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error for empty date, got: %s", errOutput)
	}
}

func TestFromCommand_Run(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Description:     "test entry",
		DurationMinutes: 60,
		RawInput:        "test entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, testEntry); err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
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

	// Call the from command's Run function directly
	fromCmd.Run(fromCmd, []string{"2024-01-01", "to", "2024-01-31"})

	output := stdout.String()
	if !strings.Contains(output, "test entry") {
		t.Errorf("Expected 'test entry' in output, got: %s", output)
	}
}
