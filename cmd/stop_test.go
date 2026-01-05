package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timer"
)

func TestCalculateDurationMinutes(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected int
	}{
		{"0 seconds", 0, 1},                          // minimum is 1 minute
		{"30 seconds", 30 * time.Second, 1},          // rounds to 1 minute (minimum)
		{"44 seconds", 44 * time.Second, 1},          // rounds down to 1 minute (minimum)
		{"45 seconds", 45 * time.Second, 1},          // rounds up to 1 minute
		{"1 minute", 1 * time.Minute, 1},             // exactly 1 minute
		{"1 minute 29 seconds", 89 * time.Second, 1}, // rounds down to 1 minute
		{"1 minute 30 seconds", 90 * time.Second, 2}, // rounds up to 2 minutes
		{"2 minutes", 2 * time.Minute, 2},            // exactly 2 minutes
		{"30 minutes", 30 * time.Minute, 30},         // 30 minutes
		{"59 minutes 29 seconds", 59*time.Minute + 29*time.Second, 59}, // rounds down
		{"59 minutes 30 seconds", 59*time.Minute + 30*time.Second, 60}, // rounds up to 1 hour
		{"1 hour", 60 * time.Minute, 60},                               // exactly 1 hour
		{"1 hour 30 minutes", 90 * time.Minute, 90},                    // 1.5 hours
		{"2 hours 14 seconds", 2*time.Hour + 14*time.Second, 120},      // rounds down
		{"2 hours 31 seconds", 2*time.Hour + 31*time.Second, 121},      // rounds up
		{"10 hours", 10 * time.Hour, 600},                              // 10 hours
		{"24 hours", 24 * time.Hour, 1440},                             // 24 hours
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDurationMinutes(tt.duration)
			if result != tt.expected {
				t.Errorf("calculateDurationMinutes(%v) = %d, expected %d", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestStopTimer_Success(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Start a timer 90 minutes ago
	timerPath, _ := timer.GetTimerPath()
	startTime := time.Now().Add(-90 * time.Minute)
	state := timer.TimerState{
		StartedAt:   startTime,
		Description: "fixing authentication bug",
		Project:     "",
		Tags:        []string{},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
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

	stopTimer()

	output := stdout.String()
	if !strings.Contains(output, "Stopped:") {
		t.Errorf("Expected 'Stopped:' in output, got: %s", output)
	}
	if !strings.Contains(output, "fixing authentication bug") {
		t.Errorf("Expected description in output, got: %s", output)
	}
	if !strings.Contains(output, "1h 30m") {
		t.Errorf("Expected duration '1h 30m' in output, got: %s", output)
	}

	// Verify entry was created in storage
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Description != "fixing authentication bug" {
		t.Errorf("Expected description 'fixing authentication bug', got: %s", entries[0].Description)
	}
	if entries[0].DurationMinutes != 90 {
		t.Errorf("Expected duration 90 minutes, got: %d", entries[0].DurationMinutes)
	}

	// Verify timer was cleared
	clearedState, _ := timer.LoadTimerState(timerPath)
	if clearedState != nil {
		t.Error("Expected timer to be cleared after stop")
	}
}

func TestStopTimer_WithProjectAndTags(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Start a timer with project and tags
	timerPath, _ := timer.GetTimerPath()
	startTime := time.Now().Add(-2 * time.Hour)
	state := timer.TimerState{
		StartedAt:   startTime,
		Description: "code review",
		Project:     "acme",
		Tags:        []string{"review", "urgent"},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
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

	stopTimer()

	output := stdout.String()
	if !strings.Contains(output, "Stopped:") {
		t.Errorf("Expected 'Stopped:' in output, got: %s", output)
	}
	if !strings.Contains(output, "code review") {
		t.Errorf("Expected description in output, got: %s", output)
	}
	if !strings.Contains(output, "@acme") {
		t.Errorf("Expected @acme in output, got: %s", output)
	}
	if !strings.Contains(output, "#review") {
		t.Errorf("Expected #review in output, got: %s", output)
	}
	if !strings.Contains(output, "#urgent") {
		t.Errorf("Expected #urgent in output, got: %s", output)
	}
	if !strings.Contains(output, "2h") {
		t.Errorf("Expected duration '2h' in output, got: %s", output)
	}

	// Verify entry was created with project and tags
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Description != "code review" {
		t.Errorf("Expected description 'code review', got: %s", entries[0].Description)
	}
	if entries[0].Project != "acme" {
		t.Errorf("Expected project 'acme', got: %s", entries[0].Project)
	}
	if len(entries[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got: %d", len(entries[0].Tags))
	}
	if !contains(entries[0].Tags, "review") || !contains(entries[0].Tags, "urgent") {
		t.Errorf("Expected tags 'review' and 'urgent', got: %v", entries[0].Tags)
	}
	if entries[0].DurationMinutes != 120 {
		t.Errorf("Expected duration 120 minutes, got: %d", entries[0].DurationMinutes)
	}

	// Verify timer was cleared
	timerState, _ := timer.LoadTimerState(timerPath)
	if timerState != nil {
		t.Error("Expected timer to be cleared after stop")
	}
}

func TestStopTimer_DurationRounding(t *testing.T) {
	tests := []struct {
		name            string
		elapsed         time.Duration
		expectedMinutes int
		expectedOutput  string
	}{
		{"29 seconds", 29 * time.Second, 1, "1m"},
		{"44 seconds", 44 * time.Second, 1, "1m"},
		{"45 seconds", 45 * time.Second, 1, "1m"},
		{"1 minute 29 seconds", 89 * time.Second, 1, "1m"},
		{"1 minute 30 seconds", 90 * time.Second, 2, "2m"},
		{"29 minutes 29 seconds", 29*time.Minute + 29*time.Second, 29, "29m"},
		{"29 minutes 30 seconds", 29*time.Minute + 30*time.Second, 30, "30m"},
		{"59 minutes 29 seconds", 59*time.Minute + 29*time.Second, 59, "59m"},
		{"59 minutes 30 seconds", 59*time.Minute + 30*time.Second, 60, "1h"},
		{"1 hour 29 seconds", 60*time.Minute + 29*time.Second, 60, "1h"},
		{"1 hour 30 seconds", 60*time.Minute + 30*time.Second, 61, "1h 1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTimerTest(t)
			defer cleanup()

			tmpDir := t.TempDir()
			storagePath := filepath.Join(tmpDir, "entries.jsonl")

			// Create a timer with specific elapsed time
			timerPath, _ := timer.GetTimerPath()
			startTime := time.Now().Add(-tt.elapsed)
			state := timer.TimerState{
				StartedAt:   startTime,
				Description: "test task",
				Project:     "",
				Tags:        []string{},
			}
			if err := timer.SaveTimerState(timerPath, state); err != nil {
				t.Fatalf("Failed to create timer: %v", err)
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

			stopTimer()

			// Verify output contains expected duration
			output := stdout.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected duration '%s' in output, got: %s", tt.expectedOutput, output)
			}

			// Verify entry has correct duration
			entries, _ := storage.ReadEntries(storagePath)
			if len(entries) != 1 {
				t.Fatalf("Expected 1 entry, got %d", len(entries))
			}
			if entries[0].DurationMinutes != tt.expectedMinutes {
				t.Errorf("Expected duration %d minutes, got: %d", tt.expectedMinutes, entries[0].DurationMinutes)
			}
		})
	}
}

func TestStopTimer_MinimumDuration(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a timer that just started (0 elapsed time)
	timerPath, _ := timer.GetTimerPath()
	state := timer.TimerState{
		StartedAt:   time.Now(),
		Description: "quick task",
		Project:     "",
		Tags:        []string{},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
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

	stopTimer()

	// Verify minimum duration is 1 minute
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].DurationMinutes < 1 {
		t.Errorf("Expected minimum duration of 1 minute, got: %d", entries[0].DurationMinutes)
	}

	// Verify output shows at least 1 minute
	output := stdout.String()
	if !strings.Contains(output, "1m") {
		t.Errorf("Expected '1m' in output for minimum duration, got: %s", output)
	}
}

func TestStopTimer_NoTimerRunning(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return storagePath, nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	stopTimer()

	if !exitCalled {
		t.Error("Expected exit to be called when no timer is running")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "No timer is running") {
		t.Errorf("Expected 'No timer is running' error, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "did start") {
		t.Errorf("Expected hint about 'did start', got: %s", errOutput)
	}

	// Verify no entry was created
	entries, err := storage.ReadEntries(storagePath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries when no timer running, got %d", len(entries))
	}
}

func TestStopTimer_TimerClearedAfterStop(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a timer
	timerPath, _ := timer.GetTimerPath()
	state := timer.TimerState{
		StartedAt:   time.Now().Add(-30 * time.Minute),
		Description: "test task",
		Project:     "test",
		Tags:        []string{"test"},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	// Verify timer exists before stopping
	isRunning, _ := timer.IsTimerRunning(timerPath)
	if !isRunning {
		t.Fatal("Timer should be running before stop")
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

	stopTimer()

	// Verify timer is cleared after stopping
	isRunning, _ = timer.IsTimerRunning(timerPath)
	if isRunning {
		t.Error("Timer should not be running after stop")
	}

	// Verify timer file is removed
	loadedState, _ := timer.LoadTimerState(timerPath)
	if loadedState != nil {
		t.Error("Timer state should be nil after stop")
	}
}

func TestStopCommand_Run(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a timer
	timerPath, _ := timer.GetTimerPath()
	state := timer.TimerState{
		StartedAt:   time.Now().Add(-45 * time.Minute),
		Description: "test task",
		Project:     "",
		Tags:        []string{},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
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

	// Call the stop command's Run function directly
	stopCmd.Run(stopCmd, []string{})

	if !strings.Contains(stdout.String(), "Stopped:") {
		t.Errorf("Expected 'Stopped:', got: %s", stdout.String())
	}
}

func TestStopTimer_EntryRawInput(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a timer
	timerPath, _ := timer.GetTimerPath()
	state := timer.TimerState{
		StartedAt:   time.Now().Add(-90 * time.Minute),
		Description: "API development",
		Project:     "client",
		Tags:        []string{"backend", "api"},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
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

	stopTimer()

	// Verify entry has proper RawInput field
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	// RawInput should be formatted as "description for duration"
	expectedRawInput := "API development for 1h 30m"
	if entries[0].RawInput != expectedRawInput {
		t.Errorf("Expected RawInput '%s', got: '%s'", expectedRawInput, entries[0].RawInput)
	}
}
