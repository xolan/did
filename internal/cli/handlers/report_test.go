package handlers

import (
	"strings"
	"testing"

	"github.com/xolan/did/internal/service"
)

func TestReportByProject(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task 1 @acme for 1h")
	CreateEntry(deps, "task 2 @acme for 30m")
	CreateEntry(deps, "task 3 @other for 45m")
	stdout.Reset()

	ReportByProject(deps, "acme", service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Report for @acme") {
		t.Errorf("expected 'Report for @acme' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Total time:") {
		t.Errorf("expected 'Total time:' in output, got %q", stdout.String())
	}
}

func TestReportByProject_Empty(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ReportByProject(deps, "nonexistent", service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	// Should still show report with 0 time
	if !strings.Contains(stdout.String(), "Report for @nonexistent") {
		t.Errorf("expected 'Report for @nonexistent' in output, got %q", stdout.String())
	}
}

func TestReportByTags(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task 1 #urgent for 1h")
	CreateEntry(deps, "task 2 #urgent #bug for 30m")
	stdout.Reset()

	ReportByTags(deps, []string{"urgent"}, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Report for #urgent") {
		t.Errorf("expected 'Report for #urgent' in output, got %q", stdout.String())
	}
}

func TestReportByTags_MultipleTags(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task #urgent #bug for 1h")
	stdout.Reset()

	ReportByTags(deps, []string{"urgent", "bug"}, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "#urgent #bug") {
		t.Errorf("expected '#urgent #bug' in output, got %q", stdout.String())
	}
}

func TestReportGroupByProject_Empty(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ReportGroupByProject(deps, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "No entries found") {
		t.Errorf("expected 'No entries found' in output, got %q", stdout.String())
	}
}

func TestReportGroupByProject_WithEntries(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task 1 @acme for 1h")
	CreateEntry(deps, "task 2 @other for 30m")
	CreateEntry(deps, "task 3 for 45m") // no project
	stdout.Reset()

	ReportGroupByProject(deps, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Time by project") {
		t.Errorf("expected 'Time by project' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "@acme") {
		t.Errorf("expected '@acme' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "@other") {
		t.Errorf("expected '@other' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "(no project)") {
		t.Errorf("expected '(no project)' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Total") {
		t.Errorf("expected 'Total' in output, got %q", stdout.String())
	}
}

func TestReportGroupByTag_Empty(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ReportGroupByTag(deps, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "No entries found") {
		t.Errorf("expected 'No entries found' in output, got %q", stdout.String())
	}
}

func TestReportGroupByTag_WithEntries(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	CreateEntry(deps, "task 1 #urgent for 1h")
	CreateEntry(deps, "task 2 #bug for 30m")
	CreateEntry(deps, "task 3 for 45m") // no tags
	stdout.Reset()

	ReportGroupByTag(deps, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Time by tag") {
		t.Errorf("expected 'Time by tag' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "#urgent") {
		t.Errorf("expected '#urgent' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "#bug") {
		t.Errorf("expected '#bug' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "(no tags)") {
		t.Errorf("expected '(no tags)' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Total") {
		t.Errorf("expected 'Total' in output, got %q", stdout.String())
	}
}

func TestReportByProject_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	ReportByProject(deps, "test", service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestReportByTags_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	ReportByTags(deps, []string{"test"}, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestReportGroupByProject_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	ReportGroupByProject(deps, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestReportGroupByTag_StorageError(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenDeps(t)

	ReportGroupByTag(deps, service.DateRangeSpec{Type: service.DateRangeToday})

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}
