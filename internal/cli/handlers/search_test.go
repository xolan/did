package handlers

import (
	"strings"
	"testing"

	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/service"
)

func TestSearch_NoResults(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	Search(deps, "nonexistent", nil, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "No entries found matching") {
		t.Errorf("expected 'No entries found matching' in output, got %q", stdout.String())
	}
}

func TestSearch_EmptyKeyword_NoEntries(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	Search(deps, "", nil, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "No entries found") {
		t.Errorf("expected 'No entries found' in output, got %q", stdout.String())
	}
}

func TestSearch_WithResults(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Create some entries
	CreateEntry(deps, "fix authentication bug for 1h")
	CreateEntry(deps, "update documentation for 30m")
	CreateEntry(deps, "fix another bug for 45m")
	stdout.Reset()

	Search(deps, "fix", nil, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Search results for 'fix'") {
		t.Errorf("expected 'Search results for' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "authentication bug") {
		t.Errorf("expected 'authentication bug' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "another bug") {
		t.Errorf("expected 'another bug' in output, got %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "documentation") {
		t.Errorf("did not expect 'documentation' in output, got %q", stdout.String())
	}
}

func TestSearch_EmptyKeyword_WithEntries(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task 1 for 1h")
	CreateEntry(deps, "task 2 for 30m")
	stdout.Reset()

	Search(deps, "", nil, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "All entries") {
		t.Errorf("expected 'All entries' in output, got %q", stdout.String())
	}
}

func TestSearch_WithDateRange(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task for 1h")
	stdout.Reset()

	dateRange := &service.DateRangeSpec{Type: service.DateRangeToday}
	Search(deps, "task", dateRange, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "task") {
		t.Errorf("expected 'task' in output, got %q", stdout.String())
	}
}

func TestSearch_WithFilter(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task @acme for 1h")
	CreateEntry(deps, "task @other for 30m")
	stdout.Reset()

	f := &filter.Filter{Project: "acme"}
	Search(deps, "task", nil, f)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "@acme") {
		t.Errorf("expected '@acme' in output, got %q", stdout.String())
	}
}

func TestSearch_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	Search(deps, "test", nil, nil)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestSearch_CorruptedLines(t *testing.T) {
	deps, stdout, stderr, exitCode := setupCorruptedDeps(t)

	Search(deps, "valid", nil, nil)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	// Should show warning about corrupted lines
	if !strings.Contains(stderr.String(), "Warning:") {
		t.Errorf("expected 'Warning:' in stderr, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "corrupted") {
		t.Errorf("expected 'corrupted' in stderr, got %q", stderr.String())
	}
	// Should still show valid entries matching search
	if !strings.Contains(stdout.String(), "valid entry") {
		t.Errorf("expected 'valid entry' in stdout, got %q", stdout.String())
	}
}
