package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/filter"
)

func TestNewSearchService(t *testing.T) {
	svc := NewSearchService("/tmp/entries.jsonl", config.DefaultConfig())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestSearchService_Search(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("fix login bug for 1h")
	_, _ = entrySvc.Create("code review for 30m")
	_, _ = entrySvc.Create("fix logout bug for 45m")

	svc := NewSearchService(storagePath, config.DefaultConfig())

	// Search for "bug"
	result, err := svc.Search("bug", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 results, got %d", result.Total)
	}
	if result.Query != "bug" {
		t.Errorf("expected query 'bug', got %q", result.Query)
	}

	// Results should be sorted by timestamp (most recent first)
	if len(result.Entries) >= 2 {
		if result.Entries[0].Entry.Timestamp.Before(result.Entries[1].Entry.Timestamp) {
			t.Error("expected results sorted by most recent first")
		}
	}
}

func TestSearchService_Search_EmptyKeyword(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 for 1h")
	_, _ = entrySvc.Create("task 2 for 30m")

	svc := NewSearchService(storagePath, config.DefaultConfig())

	// Empty keyword should return all entries
	result, err := svc.Search("", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 results, got %d", result.Total)
	}
}

func TestSearchService_Search_WithDateRange(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 for 1h")
	_, _ = entrySvc.Create("task 2 for 30m")

	svc := NewSearchService(storagePath, config.DefaultConfig())

	// Search with today's date range
	dateRange := DateRangeSpec{Type: DateRangeToday}
	result, err := svc.Search("", &dateRange, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 results, got %d", result.Total)
	}
}

func TestSearchService_Search_WithFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 @acme for 1h")
	_, _ = entrySvc.Create("task 2 @other for 30m")
	_, _ = entrySvc.Create("task 3 @acme #urgent for 45m")

	svc := NewSearchService(storagePath, config.DefaultConfig())

	// Search with project filter
	f := filter.NewFilter("", "acme", nil)
	result, err := svc.Search("task", nil, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 results, got %d", result.Total)
	}
}

func TestSearchService_Search_CombinedFilters(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("fix bug @acme #urgent for 1h")
	_, _ = entrySvc.Create("fix bug @acme for 30m")
	_, _ = entrySvc.Create("fix bug @other #urgent for 45m")

	svc := NewSearchService(storagePath, config.DefaultConfig())

	// Search for "bug" in project "acme" with tag "urgent"
	f := filter.NewFilter("", "acme", []string{"urgent"})
	result, err := svc.Search("bug", nil, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 result, got %d", result.Total)
	}
}

func TestSearchService_Search_NoResults(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 for 1h")

	svc := NewSearchService(storagePath, config.DefaultConfig())

	result, err := svc.Search("nonexistent", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 results, got %d", result.Total)
	}
}

func TestSearchService_Search_DateRanges(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewSearchService(storagePath, config.DefaultConfig())

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
		result, err := svc.Search("", &spec, nil)
		if err != nil {
			t.Errorf("unexpected error for range type %d: %v", spec.Type, err)
		}
		if result == nil {
			t.Errorf("expected result for range type %d", spec.Type)
		}
	}
}

func TestSearchService_Search_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewSearchService(storagePath, config.DefaultConfig())
	_, err := svc.Search("test", nil, nil)
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestSearchService_resolveDateRange_Default(t *testing.T) {
	svc := NewSearchService("/tmp/test.jsonl", config.DefaultConfig())

	// Test default case (unknown type)
	start, end, period := svc.resolveDateRange(DateRangeSpec{Type: DateRange(999)})
	if period != "all time" {
		t.Errorf("expected 'all time' for unknown type, got %q", period)
	}
	if !start.IsZero() {
		t.Error("expected zero start time for default case")
	}
	if end.IsZero() {
		t.Error("expected non-zero end time for default case")
	}
}
