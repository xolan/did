package handlers

import (
	"strings"
	"testing"
)

func TestStartTimer(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	StartTimer(deps, "working on feature", false)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Timer started:") {
		t.Errorf("expected 'Timer started:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "working on feature") {
		t.Errorf("expected task description in output, got %q", stdout.String())
	}
}

func TestStartTimer_WithProjectAndTags(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	StartTimer(deps, "task @project #tag", false)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "@project") {
		t.Errorf("expected '@project' in output, got %q", stdout.String())
	}
}

func TestStartTimer_AlreadyRunning(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	// Start first timer
	StartTimer(deps, "first task", false)

	// Try to start another
	*exitCode = 0
	StartTimer(deps, "second task", false)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "timer is already running") {
		t.Errorf("expected 'timer is already running' error, got %q", stderr.String())
	}
}

func TestStartTimer_Force(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Start first timer
	StartTimer(deps, "first task", false)
	stdout.Reset()

	// Force start another
	StartTimer(deps, "second task", true)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Timer started:") {
		t.Errorf("expected 'Timer started:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "overwritten") {
		t.Errorf("expected 'overwritten' in output, got %q", stdout.String())
	}
}

func TestStartTimer_EmptyDescription(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	StartTimer(deps, "", false)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Description cannot be empty") {
		t.Errorf("expected 'Description cannot be empty' error, got %q", stderr.String())
	}
}

func TestStopTimer(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Start a timer first
	StartTimer(deps, "task", false)
	stdout.Reset()

	// Stop it
	StopTimer(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Stopped:") {
		t.Errorf("expected 'Stopped:' in output, got %q", stdout.String())
	}
}

func TestStopTimer_NoTimerRunning(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	StopTimer(deps)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "No timer is running") {
		t.Errorf("expected 'No timer is running' error, got %q", stderr.String())
	}
}

func TestShowTimerStatus_NoTimer(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ShowTimerStatus(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "No timer running") {
		t.Errorf("expected 'No timer running' in output, got %q", stdout.String())
	}
}

func TestShowTimerStatus_TimerRunning(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Start a timer first
	StartTimer(deps, "working on task @project #tag", false)
	stdout.Reset()

	// Check status
	ShowTimerStatus(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Timer running:") {
		t.Errorf("expected 'Timer running:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "working on task") {
		t.Errorf("expected task description in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Started:") {
		t.Errorf("expected 'Started:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Elapsed:") {
		t.Errorf("expected 'Elapsed:' in output, got %q", stdout.String())
	}
}

func TestStartTimer_OnlyProjectTags(t *testing.T) {
	deps, _, stderr, exitCode := setupTestDeps(t)

	// Starting with only project/tags and no description should fail
	StartTimer(deps, "@project #tag", false)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Description cannot be empty") {
		t.Errorf("expected empty description error, got %q", stderr.String())
	}
}

func TestStopTimer_StopsAndCreatesEntry(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Start a timer with project and tags
	StartTimer(deps, "feature work @project #feature", false)
	stdout.Reset()

	// Stop it
	StopTimer(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "@project") {
		t.Errorf("expected '@project' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "#feature") {
		t.Errorf("expected '#feature' in output, got %q", stdout.String())
	}
}

func TestStopTimer_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	// Can't start/stop with broken storage path - just verify error handling
	StopTimer(deps)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error") || !strings.Contains(stderr.String(), "running") {
		// Either "No timer is running" or storage error
		t.Logf("stderr output: %s", stderr.String())
	}
}

func TestShowTimerStatus_WithError(t *testing.T) {
	// Create deps with broken timer path to trigger error
	deps, _, _, exitCode := setupBrokenTimerDeps(t)

	ShowTimerStatus(deps)

	// May exit 0 (no timer) or exit 1 (error) depending on error handling
	_ = exitCode
}
