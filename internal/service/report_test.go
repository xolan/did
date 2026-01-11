package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xolan/did/internal/config"
)

func TestNewReportService(t *testing.T) {
	svc := NewReportService("/tmp/entries.jsonl", config.DefaultConfig())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestReportService_ByProject(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	// Create some entries first
	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 @acme for 1h")
	_, _ = entrySvc.Create("task 2 @acme for 30m")
	_, _ = entrySvc.Create("task 3 @other for 45m")

	svc := NewReportService(storagePath, config.DefaultConfig())

	report, err := svc.ByProject("acme", DateRangeSpec{Type: DateRangeToday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalMinutes != 90 { // 60 + 30
		t.Errorf("expected 90 minutes, got %d", report.TotalMinutes)
	}
	if report.EntryCount != 2 {
		t.Errorf("expected 2 entries, got %d", report.EntryCount)
	}
	if len(report.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(report.Groups))
	}
	if report.Groups[0].Name != "acme" {
		t.Errorf("expected group name 'acme', got %q", report.Groups[0].Name)
	}
}

func TestReportService_ByTags(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 #urgent for 1h")
	_, _ = entrySvc.Create("task 2 #urgent #important for 30m")
	_, _ = entrySvc.Create("task 3 #other for 45m")

	svc := NewReportService(storagePath, config.DefaultConfig())

	report, err := svc.ByTags([]string{"urgent"}, DateRangeSpec{Type: DateRangeToday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalMinutes != 90 { // 60 + 30
		t.Errorf("expected 90 minutes, got %d", report.TotalMinutes)
	}
	if report.EntryCount != 2 {
		t.Errorf("expected 2 entries, got %d", report.EntryCount)
	}
}

func TestReportService_GroupByProject(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 @acme for 1h")
	_, _ = entrySvc.Create("task 2 @other for 30m")
	_, _ = entrySvc.Create("task 3 @acme for 45m")
	_, _ = entrySvc.Create("task 4 for 15m") // No project

	svc := NewReportService(storagePath, config.DefaultConfig())

	report, err := svc.GroupByProject(DateRangeSpec{Type: DateRangeToday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalMinutes != 150 { // 60 + 30 + 45 + 15
		t.Errorf("expected 150 minutes, got %d", report.TotalMinutes)
	}
	if report.EntryCount != 4 {
		t.Errorf("expected 4 entries, got %d", report.EntryCount)
	}
	if len(report.Groups) != 3 { // acme, other, (no project)
		t.Errorf("expected 3 groups, got %d", len(report.Groups))
	}

	// Should be sorted by total minutes descending
	// acme: 105m, other: 30m, (no project): 15m
	if report.Groups[0].Name != "acme" {
		t.Errorf("expected first group 'acme', got %q", report.Groups[0].Name)
	}
	if report.Groups[0].TotalMinutes != 105 {
		t.Errorf("expected acme total 105, got %d", report.Groups[0].TotalMinutes)
	}
}

func TestReportService_GroupByTag(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 #urgent for 1h")
	_, _ = entrySvc.Create("task 2 #urgent #important for 30m")
	_, _ = entrySvc.Create("task 3 #important for 45m")
	_, _ = entrySvc.Create("task 4 for 15m") // No tags

	svc := NewReportService(storagePath, config.DefaultConfig())

	report, err := svc.GroupByTag(DateRangeSpec{Type: DateRangeToday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total time should be sum of all entries
	if report.TotalMinutes != 150 { // 60 + 30 + 45 + 15
		t.Errorf("expected 150 minutes, got %d", report.TotalMinutes)
	}

	// Should have groups for urgent, important, and (no tags)
	if len(report.Groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(report.Groups))
	}

	// Note: urgent appears in 2 entries (60+30=90), important in 2 (30+45=75)
	// Should be sorted by total descending
	if report.Groups[0].Name != "urgent" {
		t.Errorf("expected first group 'urgent', got %q", report.Groups[0].Name)
	}
}

func TestReportService_DateRanges(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewReportService(storagePath, config.DefaultConfig())

	// Test all date range types
	ranges := []DateRangeSpec{
		{Type: DateRangeToday},
		{Type: DateRangeYesterday},
		{Type: DateRangeThisWeek},
		{Type: DateRangePrevWeek},
		{Type: DateRangeThisMonth},
		{Type: DateRangePrevMonth},
		{Type: DateRangeLast, LastDays: 7},
		{Type: DateRangeCustom, From: time.Now().AddDate(0, 0, -7), To: time.Now()},
	}

	for _, spec := range ranges {
		report, err := svc.GroupByProject(spec)
		if err != nil {
			t.Errorf("unexpected error for range type %d: %v", spec.Type, err)
		}
		if report == nil {
			t.Errorf("expected report for range type %d", spec.Type)
		}
	}
}

func TestReportService_EmptyStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewReportService(storagePath, config.DefaultConfig())

	report, err := svc.GroupByProject(DateRangeSpec{Type: DateRangeToday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalMinutes != 0 {
		t.Errorf("expected 0 minutes, got %d", report.TotalMinutes)
	}
	if report.EntryCount != 0 {
		t.Errorf("expected 0 entries, got %d", report.EntryCount)
	}
}

func TestReportService_ByProject_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewReportService(storagePath, config.DefaultConfig())
	_, err := svc.ByProject("acme", DateRangeSpec{Type: DateRangeToday})
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestReportService_ByTags_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewReportService(storagePath, config.DefaultConfig())
	_, err := svc.ByTags([]string{"tag"}, DateRangeSpec{Type: DateRangeToday})
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestReportService_GroupByProject_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewReportService(storagePath, config.DefaultConfig())
	_, err := svc.GroupByProject(DateRangeSpec{Type: DateRangeToday})
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestReportService_GroupByTag_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewReportService(storagePath, config.DefaultConfig())
	_, err := svc.GroupByTag(DateRangeSpec{Type: DateRangeToday})
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestReportService_resolveDateRange(t *testing.T) {
	svc := NewReportService("/tmp/test.jsonl", config.DefaultConfig())

	// Test default case
	start, end, period := svc.resolveDateRange(DateRangeSpec{Type: DateRange(999)})
	if period != "today" {
		t.Errorf("expected 'today' for unknown type, got %q", period)
	}
	if start.IsZero() || end.IsZero() {
		t.Error("expected non-zero times for default case")
	}
}
