package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/timer"
)

func TestNewTimerService(t *testing.T) {
	svc := NewTimerService("/tmp/timer.json", "/tmp/entries.jsonl", config.DefaultConfig())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestTimerService_Start(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start a timer
	state, existing, err := svc.Start("working on task", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state == nil {
		t.Fatal("expected state, got nil")
	}
	if existing != nil {
		t.Error("expected no existing timer")
	}
	if state.Description != "working on task" {
		t.Errorf("expected description 'working on task', got %q", state.Description)
	}
}

func TestTimerService_Start_WithProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	state, _, err := svc.Start("task @project #tag1 #tag2", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Description != "task" {
		t.Errorf("expected 'task', got %q", state.Description)
	}
	if state.Project != "project" {
		t.Errorf("expected project 'project', got %q", state.Project)
	}
	if len(state.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(state.Tags))
	}
}

func TestTimerService_Start_AlreadyRunning(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start first timer
	_, _, _ = svc.Start("task 1", false)

	// Try to start another without force
	_, existing, err := svc.Start("task 2", false)
	if err != ErrTimerAlreadyRunning {
		t.Errorf("expected ErrTimerAlreadyRunning, got %v", err)
	}
	if existing == nil {
		t.Error("expected existing timer state")
	}
}

func TestTimerService_Start_Force(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start first timer
	_, _, _ = svc.Start("task 1", false)

	// Force start another
	state, existing, err := svc.Start("task 2", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if existing == nil {
		t.Error("expected existing timer state")
	}
	if state.Description != "task 2" {
		t.Errorf("expected 'task 2', got %q", state.Description)
	}
}

func TestTimerService_Start_EmptyDescription(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	_, _, err := svc.Start("", false)
	if err != ErrEmptyDescription {
		t.Errorf("expected ErrEmptyDescription, got %v", err)
	}

	_, _, err = svc.Start("   ", false)
	if err != ErrEmptyDescription {
		t.Errorf("expected ErrEmptyDescription for whitespace, got %v", err)
	}

	_, _, err = svc.Start("@project #tag", false)
	if err != ErrEmptyDescription {
		t.Errorf("expected ErrEmptyDescription for only project/tags, got %v", err)
	}
}

func TestTimerService_Stop(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start a timer
	_, _, _ = svc.Start("task", false)

	// Wait a tiny bit (simulates some work)
	time.Sleep(10 * time.Millisecond)

	// Stop the timer
	entry, state, err := svc.Stop()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if state == nil {
		t.Fatal("expected state, got nil")
	}
	if entry.Description != "task" {
		t.Errorf("expected 'task', got %q", entry.Description)
	}
	if entry.DurationMinutes < 1 {
		t.Error("expected at least 1 minute duration")
	}
}

func TestTimerService_Stop_NoTimer(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	_, _, err := svc.Stop()
	if err != ErrNoTimerRunning {
		t.Errorf("expected ErrNoTimerRunning, got %v", err)
	}
}

func TestTimerService_Cancel(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start a timer
	_, _, _ = svc.Start("task", false)

	// Cancel it
	state, err := svc.Cancel()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state == nil {
		t.Fatal("expected state, got nil")
	}
	if state.Description != "task" {
		t.Errorf("expected 'task', got %q", state.Description)
	}

	// Verify timer is no longer running
	running, _ := svc.IsRunning()
	if running {
		t.Error("timer should not be running after cancel")
	}
}

func TestTimerService_Cancel_NoTimer(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	_, err := svc.Cancel()
	if err != ErrNoTimerRunning {
		t.Errorf("expected ErrNoTimerRunning, got %v", err)
	}
}

func TestTimerService_Status(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// No timer running
	status, err := svc.Status()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Running {
		t.Error("expected not running")
	}
	if status.State != nil {
		t.Error("expected nil state")
	}

	// Start a timer
	_, _, _ = svc.Start("task", false)

	// Check status
	status, err = svc.Status()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Running {
		t.Error("expected running")
	}
	if status.State == nil {
		t.Fatal("expected state")
	}
	if status.State.Description != "task" {
		t.Errorf("expected 'task', got %q", status.State.Description)
	}
	if status.ElapsedTime < 0 {
		t.Error("expected non-negative elapsed time")
	}
}

func TestTimerService_IsRunning(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Initially not running
	running, err := svc.IsRunning()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if running {
		t.Error("expected not running")
	}

	// Start timer
	_, _, _ = svc.Start("task", false)

	running, err = svc.IsRunning()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !running {
		t.Error("expected running")
	}

	// Stop timer
	_, _, _ = svc.Stop()

	running, err = svc.IsRunning()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if running {
		t.Error("expected not running after stop")
	}
}

func TestCalculateDurationMinutes(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     int
	}{
		{0, 1},                // Minimum 1 minute
		{30 * time.Second, 1}, // Round up to 1
		{59 * time.Second, 1}, // Round to 1
		{1 * time.Minute, 1},
		{90 * time.Second, 2}, // Round 1.5 to 2
		{5 * time.Minute, 5},
		{5*time.Minute + 30*time.Second, 6}, // Round 5.5 to 6
		{1 * time.Hour, 60},
		{90 * time.Minute, 90},
	}

	for _, tt := range tests {
		t.Run(tt.duration.String(), func(t *testing.T) {
			result := calculateDurationMinutes(tt.duration)
			if result != tt.want {
				t.Errorf("calculateDurationMinutes(%v) = %d, want %d", tt.duration, result, tt.want)
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{" hello", "hello"},
		{"hello ", "hello"},
		{" hello ", "hello"},
		{"  hello  ", "hello"},
		{"\thello\t", "hello"},
		{"\nhello\n", "hello"},
		{"\r\nhello\r\n", "hello"},
		{" \t\n\rhello \t\n\r", "hello"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trimSpace(tt.input)
			if result != tt.want {
				t.Errorf("trimSpace(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func TestTimerService_Stop_WithProjectAndTags(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start a timer with project and tags
	_, _, _ = svc.Start("task @project #tag", false)

	// Stop the timer
	entry, _, err := svc.Stop()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Project != "project" {
		t.Errorf("expected project 'project', got %q", entry.Project)
	}
	if len(entry.Tags) != 1 || entry.Tags[0] != "tag" {
		t.Errorf("expected tags [tag], got %v", entry.Tags)
	}
}

func TestTimerService_Start_SaveError(t *testing.T) {
	// Use invalid path that can't be written to
	svc := NewTimerService("/nonexistent/dir/timer.json", "/tmp/entries.jsonl", config.DefaultConfig())
	_, _, err := svc.Start("task", false)
	if err == nil {
		t.Error("expected error for invalid timer path")
	}
}

func TestTimerService_Stop_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create directory where storage file should be
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start a timer
	_, _, _ = svc.Start("task", false)

	// Stop should fail because storage is a directory
	_, _, err := svc.Stop()
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestTimerService_Cancel_ClearError(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start a timer
	_, _, _ = svc.Start("task", false)

	// Make directory read-only to cause clear (delete) error
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	// Cancel should fail because timer file can't be deleted
	_, err := svc.Cancel()
	if err == nil {
		t.Error("expected error when directory is read-only")
	}
}

func TestTimerService_Start_CheckStatusError(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "subdir", "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create a file where the directory should be
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir"), []byte("file"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Start should fail because timer path is invalid
	_, _, err := svc.Start("task", false)
	if err == nil {
		t.Error("expected error for invalid timer path")
	}
}

func TestTimerService_Stop_LoadTimerState(t *testing.T) {
	tmpDir := t.TempDir()
	timerPath := filepath.Join(tmpDir, "timer.json")
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Manually create a timer state
	state := timer.TimerState{
		StartedAt:   time.Now().Add(-time.Hour),
		Description: "manual task",
		Project:     "test",
		Tags:        []string{"manual"},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatal(err)
	}

	svc := NewTimerService(timerPath, storagePath, config.DefaultConfig())

	// Stop should work
	entry, returnedState, err := svc.Stop()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Description != "manual task" {
		t.Errorf("expected 'manual task', got %q", entry.Description)
	}
	if returnedState.Description != "manual task" {
		t.Errorf("expected state description 'manual task', got %q", returnedState.Description)
	}
	// Duration should be approximately 60 minutes (1 hour)
	if entry.DurationMinutes < 55 || entry.DurationMinutes > 65 {
		t.Errorf("expected ~60 minutes, got %d", entry.DurationMinutes)
	}
}
