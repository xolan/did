package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/timer"
)

func TestFormatElapsedTime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"0 minutes", 0, "0m"},
		{"1 minute", 1 * time.Minute, "1m"},
		{"30 minutes", 30 * time.Minute, "30m"},
		{"59 minutes", 59 * time.Minute, "59m"},
		{"exactly 1 hour", 60 * time.Minute, "1h"},
		{"exactly 2 hours", 120 * time.Minute, "2h"},
		{"1 hour 30 minutes", 90 * time.Minute, "1h 30m"},
		{"2 hours 15 minutes", 135 * time.Minute, "2h 15m"},
		{"10 hours", 600 * time.Minute, "10h"},
		{"10 hours 5 minutes", 605 * time.Minute, "10h 5m"},
		{"24 hours", 1440 * time.Minute, "24h"},
		{"with seconds (rounds down)", 90*time.Minute + 45*time.Second, "1h 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatElapsedTime(tt.duration)
			if result != tt.expected {
				t.Errorf("formatElapsedTime(%v) = %q, expected %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// setupTimerTest sets up a test environment with a clean timer state.
// It backs up any existing timer and restores it after the test.
func setupTimerTest(t *testing.T) func() {
	t.Helper()

	// Get the real timer path
	timerPath, err := timer.GetTimerPath()
	if err != nil {
		t.Fatalf("Failed to get timer path: %v", err)
	}

	// Backup existing timer if it exists
	var backupData []byte
	if data, err := os.ReadFile(timerPath); err == nil {
		backupData = data
	}

	// Clear any existing timer
	_ = timer.ClearTimerState(timerPath)

	// Return cleanup function
	return func() {
		// Clear test timer
		_ = timer.ClearTimerState(timerPath)

		// Restore backup if it existed
		if backupData != nil {
			_ = os.WriteFile(timerPath, backupData, 0644)
		}
	}
}

func TestStartTimer_Success(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	startTimer([]string{"fixing", "authentication", "bug"})

	output := stdout.String()
	if !strings.Contains(output, "Timer started:") {
		t.Errorf("Expected 'Timer started:' in output, got: %s", output)
	}
	if !strings.Contains(output, "fixing authentication bug") {
		t.Errorf("Expected description in output, got: %s", output)
	}

	// Verify timer state was saved
	timerPath, _ := timer.GetTimerPath()
	state, err := timer.LoadTimerState(timerPath)
	if err != nil {
		t.Fatalf("Failed to load timer state: %v", err)
	}
	if state == nil {
		t.Fatal("Expected timer state to be saved, got nil")
	}
	if state.Description != "fixing authentication bug" {
		t.Errorf("Expected description 'fixing authentication bug', got: %s", state.Description)
	}
	if state.Project != "" {
		t.Errorf("Expected no project, got: %s", state.Project)
	}
	if len(state.Tags) != 0 {
		t.Errorf("Expected no tags, got: %v", state.Tags)
	}
}

func TestStartTimer_WithProjectAndTags(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	startTimer([]string{"code", "review", "@acme", "#review", "#urgent"})

	output := stdout.String()
	if !strings.Contains(output, "Timer started:") {
		t.Errorf("Expected 'Timer started:' in output, got: %s", output)
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

	// Verify timer state was saved with project and tags
	timerPath, _ := timer.GetTimerPath()
	state, err := timer.LoadTimerState(timerPath)
	if err != nil {
		t.Fatalf("Failed to load timer state: %v", err)
	}
	if state == nil {
		t.Fatal("Expected timer state to be saved, got nil")
	}
	if state.Description != "code review" {
		t.Errorf("Expected description 'code review', got: %s", state.Description)
	}
	if state.Project != "acme" {
		t.Errorf("Expected project 'acme', got: %s", state.Project)
	}
	if len(state.Tags) != 2 {
		t.Errorf("Expected 2 tags, got: %d", len(state.Tags))
	}
	if !contains(state.Tags, "review") || !contains(state.Tags, "urgent") {
		t.Errorf("Expected tags 'review' and 'urgent', got: %v", state.Tags)
	}
}

func TestStartTimer_AlreadyRunning_NoForce(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create an existing timer
	timerPath, _ := timer.GetTimerPath()
	existingState := timer.TimerState{
		StartedAt:   time.Now().Add(-30 * time.Minute),
		Description: "existing task",
		Project:     "oldproject",
		Tags:        []string{"old"},
	}
	if err := timer.SaveTimerState(timerPath, existingState); err != nil {
		t.Fatalf("Failed to create existing timer: %v", err)
	}

	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	forceFlag = false

	startTimer([]string{"new", "task"})

	if !exitCalled {
		t.Error("Expected exit to be called when timer is already running")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Warning: A timer is already running") {
		t.Errorf("Expected warning message, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "existing task") {
		t.Errorf("Expected existing task description in warning, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "@oldproject") {
		t.Errorf("Expected existing project in warning, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "#old") {
		t.Errorf("Expected existing tag in warning, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "ago") {
		t.Errorf("Expected elapsed time in warning, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "did stop") {
		t.Errorf("Expected 'did stop' suggestion, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "--force") {
		t.Errorf("Expected '--force' suggestion, got: %s", errOutput)
	}

	// Verify existing timer was not overwritten
	state, _ := timer.LoadTimerState(timerPath)
	if state.Description != "existing task" {
		t.Errorf("Expected existing timer to be preserved, got: %s", state.Description)
	}
}

func TestStartTimer_AlreadyRunning_WithForce(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create an existing timer
	timerPath, _ := timer.GetTimerPath()
	existingState := timer.TimerState{
		StartedAt:   time.Now().Add(-30 * time.Minute),
		Description: "existing task",
		Project:     "oldproject",
		Tags:        []string{"old"},
	}
	if err := timer.SaveTimerState(timerPath, existingState); err != nil {
		t.Fatalf("Failed to create existing timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	forceFlag = true
	defer func() { forceFlag = false }()

	startTimer([]string{"new", "task", "@newproject"})

	output := stdout.String()
	if !strings.Contains(output, "Timer started:") {
		t.Errorf("Expected 'Timer started:' in output, got: %s", output)
	}
	if !strings.Contains(output, "new task") {
		t.Errorf("Expected new description in output, got: %s", output)
	}
	if !strings.Contains(output, "(Previous timer was overwritten)") {
		t.Errorf("Expected overwrite message, got: %s", output)
	}

	// Verify existing timer was overwritten
	state, err := timer.LoadTimerState(timerPath)
	if err != nil {
		t.Fatalf("Failed to load timer state: %v", err)
	}
	if state.Description != "new task" {
		t.Errorf("Expected new description 'new task', got: %s", state.Description)
	}
	if state.Project != "newproject" {
		t.Errorf("Expected new project 'newproject', got: %s", state.Project)
	}
}

func TestStartTimer_EmptyDescription(t *testing.T) {
	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	startTimer([]string{})

	if !exitCalled {
		t.Error("Expected exit to be called for empty description")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Description cannot be empty") {
		t.Errorf("Expected empty description error, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "Usage:") {
		t.Errorf("Expected usage hint, got: %s", errOutput)
	}
}

func TestStartTimer_OnlyProjectAndTags(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	startTimer([]string{"@project", "#tag"})

	if !exitCalled {
		t.Error("Expected exit to be called when only project/tags provided")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Description cannot be empty (only project/tags provided)") {
		t.Errorf("Expected empty description error, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "Include a description") {
		t.Errorf("Expected hint to include description, got: %s", errOutput)
	}
}

func TestStartCommand_Run(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Call the start command's Run function directly
	startCmd.Run(startCmd, []string{"test", "task"})

	if !strings.Contains(stdout.String(), "Timer started:") {
		t.Errorf("Expected 'Timer started:', got: %s", stdout.String())
	}
}

func TestStartTimer_WhitespaceHandling(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	startTimer([]string{"  ", "test", "  ", "task", "  "})

	output := stdout.String()
	if !strings.Contains(output, "test task") {
		t.Errorf("Expected whitespace-trimmed description in output, got: %s", output)
	}

	// Verify saved state
	timerPath, _ := timer.GetTimerPath()
	state, _ := timer.LoadTimerState(timerPath)
	if state.Description != "test task" {
		t.Errorf("Expected description 'test task', got: %s", state.Description)
	}
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
