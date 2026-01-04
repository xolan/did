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

func TestHandleLastCommand_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entries in the past 7 days
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -3), // 3 days ago
			Description:     "recent entry",
			DurationMinutes: 60,
			RawInput:        "recent entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -6), // 6 days ago
			Description:     "older entry",
			DurationMinutes: 120,
			RawInput:        "older entry for 2h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -10), // 10 days ago (outside range)
			Description:     "very old entry",
			DurationMinutes: 30,
			RawInput:        "very old entry for 30m",
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

	// Test: last 7 days
	handleLastCommand(lastCmd, []string{"7", "days"})

	output := stdout.String()
	if !strings.Contains(output, "recent entry") {
		t.Errorf("Expected 'recent entry' in output, got: %s", output)
	}
	if !strings.Contains(output, "older entry") {
		t.Errorf("Expected 'older entry' in output, got: %s", output)
	}
	// Should NOT include the 10-day-old entry
	if strings.Contains(output, "very old entry") {
		t.Errorf("Did not expect 'very old entry' in 7-day range, got: %s", output)
	}
}

func TestHandleLastCommand_ThirtyDays(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entries
	now := time.Now()
	entries := []entry.Entry{
		{
			Timestamp:       now.AddDate(0, 0, -15), // 15 days ago
			Description:     "mid-month entry",
			DurationMinutes: 60,
			RawInput:        "mid-month entry for 1h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -29), // 29 days ago
			Description:     "almost 30 days",
			DurationMinutes: 120,
			RawInput:        "almost 30 days for 2h",
		},
		{
			Timestamp:       now.AddDate(0, 0, -35), // 35 days ago (outside range)
			Description:     "outside range",
			DurationMinutes: 30,
			RawInput:        "outside range for 30m",
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

	// Test: last 30 days
	handleLastCommand(lastCmd, []string{"30", "days"})

	output := stdout.String()
	if !strings.Contains(output, "mid-month entry") {
		t.Errorf("Expected 'mid-month entry' in output, got: %s", output)
	}
	if !strings.Contains(output, "almost 30 days") {
		t.Errorf("Expected 'almost 30 days' in output, got: %s", output)
	}
	// Should NOT include the 35-day-old entry
	if strings.Contains(output, "outside range") {
		t.Errorf("Did not expect 'outside range' in 30-day range, got: %s", output)
	}
}

func TestHandleLastCommand_OneDay(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entries
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local)
	yesterday := today.AddDate(0, 0, -1)

	entries := []entry.Entry{
		{
			Timestamp:       today,
			Description:     "today entry",
			DurationMinutes: 60,
			RawInput:        "today entry for 1h",
		},
		{
			Timestamp:       yesterday,
			Description:     "yesterday entry",
			DurationMinutes: 30,
			RawInput:        "yesterday entry for 30m",
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

	// Test: last 1 day (singular form)
	handleLastCommand(lastCmd, []string{"1", "day"})

	output := stdout.String()
	if !strings.Contains(output, "today entry") {
		t.Errorf("Expected 'today entry' in output, got: %s", output)
	}
	// Should NOT include yesterday's entry
	if strings.Contains(output, "yesterday entry") {
		t.Errorf("Did not expect 'yesterday entry' in 1-day range, got: %s", output)
	}
}

func TestHandleLastCommand_InvalidNumber(t *testing.T) {
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

	// Invalid number (not a number)
	handleLastCommand(lastCmd, []string{"abc", "days"})

	if !exitCalled {
		t.Error("Expected exit to be called for invalid number")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error message for invalid number, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "Examples:") {
		t.Errorf("Expected examples in error output, got: %s", errOutput)
	}
}

func TestHandleLastCommand_WrongKeyword(t *testing.T) {
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

	// Wrong keyword (weeks instead of days)
	handleLastCommand(lastCmd, []string{"7", "weeks"})

	if !exitCalled {
		t.Error("Expected exit to be called for wrong keyword")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error message for wrong keyword, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "Examples:") {
		t.Errorf("Expected examples in error output, got: %s", errOutput)
	}
}

func TestHandleLastCommand_NegativeNumber(t *testing.T) {
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

	// Negative number
	handleLastCommand(lastCmd, []string{"-7", "days"})

	if !exitCalled {
		t.Error("Expected exit to be called for negative number")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error message for negative number, got: %s", errOutput)
	}
}

func TestHandleLastCommand_ZeroDays(t *testing.T) {
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

	// Zero days
	handleLastCommand(lastCmd, []string{"0", "days"})

	if !exitCalled {
		t.Error("Expected exit to be called for zero days")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error message for zero days, got: %s", errOutput)
	}
}

func TestHandleLastCommand_MissingKeyword(t *testing.T) {
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

	// Missing 'days' keyword (only number)
	handleLastCommand(lastCmd, []string{"7"})

	if !exitCalled {
		t.Error("Expected exit to be called for missing keyword")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error message for missing keyword, got: %s", errOutput)
	}
}

func TestHandleLastCommand_ExtraWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	now := time.Now()
	testEntry := entry.Entry{
		Timestamp:       now,
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

	// Test with extra spaces in args (should be handled by string joining)
	handleLastCommand(lastCmd, []string{"7", "days"})

	output := stdout.String()
	// Should successfully parse and show entry
	if !strings.Contains(output, "test entry") {
		t.Errorf("Expected 'test entry' in output, got: %s", output)
	}
}

func TestHandleLastCommand_LargeNumber(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry from 300 days ago
	oldEntry := entry.Entry{
		Timestamp:       time.Now().AddDate(0, 0, -300),
		Description:     "very old entry",
		DurationMinutes: 60,
		RawInput:        "very old entry for 1h",
	}
	if err := storage.AppendEntry(storagePath, oldEntry); err != nil {
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

	// Test with large number (365 days = 1 year)
	handleLastCommand(lastCmd, []string{"365", "days"})

	output := stdout.String()
	// Should successfully parse and show entry from 300 days ago
	if !strings.Contains(output, "very old entry") {
		t.Errorf("Expected 'very old entry' in output, got: %s", output)
	}
}

func TestLastCommand_Run(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create test entry
	testEntry := entry.Entry{
		Timestamp:       time.Now(),
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

	// Call the last command's Run function directly
	lastCmd.Run(lastCmd, []string{"7", "days"})

	output := stdout.String()
	if !strings.Contains(output, "test entry") {
		t.Errorf("Expected 'test entry' in output, got: %s", output)
	}
}

func TestHandleLastCommand_EmptyArgs(t *testing.T) {
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

	// Empty args (edge case - should fail)
	handleLastCommand(lastCmd, []string{})

	if !exitCalled {
		t.Error("Expected exit to be called for empty args")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("Expected error message for empty args, got: %s", errOutput)
	}
}
