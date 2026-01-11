package handlers

import (
	"strings"
	"testing"
)

func TestShowWeeklyStats_Empty(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ShowWeeklyStats(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Statistics for this week") {
		t.Errorf("expected 'Statistics for this week' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Total time:") {
		t.Errorf("expected 'Total time:' in output, got %q", stdout.String())
	}
}

func TestShowWeeklyStats_WithEntries(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Create some entries
	CreateEntry(deps, "task 1 @acme for 1h")
	CreateEntry(deps, "task 2 #urgent for 30m")
	stdout.Reset()

	ShowWeeklyStats(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "this week") {
		t.Errorf("expected 'this week' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Total time:") {
		t.Errorf("expected 'Total time:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Total entries:") {
		t.Errorf("expected 'Total entries:' in output, got %q", stdout.String())
	}
}

func TestShowMonthlyStats_Empty(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ShowMonthlyStats(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Statistics for this month") {
		t.Errorf("expected 'Statistics for this month' in output, got %q", stdout.String())
	}
}

func TestShowMonthlyStats_WithEntries(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Create some entries
	CreateEntry(deps, "task 1 for 2h")
	CreateEntry(deps, "task 2 for 45m")
	stdout.Reset()

	ShowMonthlyStats(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "this month") {
		t.Errorf("expected 'this month' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Total entries:") {
		t.Errorf("expected 'Total entries:' in output, got %q", stdout.String())
	}
}

func TestDisplayStats_WithComparison(t *testing.T) {
	deps, stdout, _, _ := setupTestDeps(t)

	// Test displayStats with comparison
	displayStats(deps, 120, 5, 3, 40.0, "this week", "+30m vs last week", nil, nil)

	if !strings.Contains(stdout.String(), "Statistics for this week") {
		t.Errorf("expected period in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Comparison:") {
		t.Errorf("expected 'Comparison:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "+30m vs last week") {
		t.Errorf("expected comparison text in output, got %q", stdout.String())
	}
}

func TestDisplayStats_NoComparison(t *testing.T) {
	deps, stdout, _, _ := setupTestDeps(t)

	// Test displayStats without comparison
	displayStats(deps, 60, 2, 1, 60.0, "today", "", nil, nil)

	if strings.Contains(stdout.String(), "Comparison:") {
		t.Errorf("did not expect 'Comparison:' in output, got %q", stdout.String())
	}
}

func TestShowWeeklyStats_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	ShowWeeklyStats(deps)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestShowMonthlyStats_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	ShowMonthlyStats(deps)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}
