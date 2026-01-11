package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xolan/did/internal/config"
)

func TestNewStatsService(t *testing.T) {
	svc := NewStatsService("/tmp/entries.jsonl", config.DefaultConfig())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestStatsService_Weekly(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 @acme for 1h")
	_, _ = entrySvc.Create("task 2 @acme for 30m")
	_, _ = entrySvc.Create("task 3 #urgent for 45m")

	svc := NewStatsService(storagePath, config.DefaultConfig())

	result, err := svc.Weekly()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Period != "this week" {
		t.Errorf("expected period 'this week', got %q", result.Period)
	}
	if result.Statistics.TotalMinutes != 135 { // 60 + 30 + 45
		t.Errorf("expected 135 minutes, got %d", result.Statistics.TotalMinutes)
	}
	if result.Statistics.EntryCount != 3 {
		t.Errorf("expected 3 entries, got %d", result.Statistics.EntryCount)
	}

	// Should have project stats
	if len(result.ProjectStats) == 0 {
		t.Error("expected project stats")
	}
}

func TestStatsService_Monthly(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 for 1h")
	_, _ = entrySvc.Create("task 2 for 30m")

	svc := NewStatsService(storagePath, config.DefaultConfig())

	result, err := svc.Monthly()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Period != "this month" {
		t.Errorf("expected period 'this month', got %q", result.Period)
	}
	if result.Statistics.TotalMinutes != 90 {
		t.Errorf("expected 90 minutes, got %d", result.Statistics.TotalMinutes)
	}
}

func TestStatsService_ForDateRange(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 for 1h")
	_, _ = entrySvc.Create("task 2 for 30m")

	svc := NewStatsService(storagePath, config.DefaultConfig())

	result, err := svc.ForDateRange(DateRangeSpec{Type: DateRangeToday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Statistics.TotalMinutes != 90 {
		t.Errorf("expected 90 minutes, got %d", result.Statistics.TotalMinutes)
	}

	// ForDateRange doesn't include comparison
	if result.Comparison != "" {
		t.Errorf("expected empty comparison for ForDateRange, got %q", result.Comparison)
	}
}

func TestStatsService_ForDateRange_AllTypes(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewStatsService(storagePath, config.DefaultConfig())

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
		result, err := svc.ForDateRange(spec)
		if err != nil {
			t.Errorf("unexpected error for range type %d: %v", spec.Type, err)
		}
		if result == nil {
			t.Errorf("expected result for range type %d", spec.Type)
		}
	}
}

func TestStatsService_Weekly_Comparison(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task for 1h")

	svc := NewStatsService(storagePath, config.DefaultConfig())

	result, err := svc.Weekly()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have comparison string
	// (might be empty if equal to last week, but should not error)
	_ = result.Comparison
}

func TestStatsService_EmptyStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewStatsService(storagePath, config.DefaultConfig())

	result, err := svc.Weekly()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Statistics.TotalMinutes != 0 {
		t.Errorf("expected 0 minutes, got %d", result.Statistics.TotalMinutes)
	}
	if result.Statistics.EntryCount != 0 {
		t.Errorf("expected 0 entries, got %d", result.Statistics.EntryCount)
	}
}

func TestStatsService_resolveDateRange_Default(t *testing.T) {
	svc := NewStatsService("/tmp/test.jsonl", config.DefaultConfig())

	// Test default case
	start, end, period := svc.resolveDateRange(DateRangeSpec{Type: DateRange(999)})
	if period != "today" {
		t.Errorf("expected 'today' for unknown type, got %q", period)
	}
	if start.IsZero() || end.IsZero() {
		t.Error("expected non-zero times for default case")
	}
}

func TestStatsService_Weekly_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewStatsService(storagePath, config.DefaultConfig())
	_, err := svc.Weekly()
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestStatsService_Monthly_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewStatsService(storagePath, config.DefaultConfig())
	_, err := svc.Monthly()
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestStatsService_ForDateRange_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewStatsService(storagePath, config.DefaultConfig())
	_, err := svc.ForDateRange(DateRangeSpec{Type: DateRangeToday})
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestStatsService_ProjectAndTagBreakdown(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")

	entrySvc := NewEntryService(storagePath, config.DefaultConfig())
	_, _ = entrySvc.Create("task 1 @acme #urgent for 1h")
	_, _ = entrySvc.Create("task 2 @other #urgent for 30m")
	_, _ = entrySvc.Create("task 3 @acme for 45m")

	svc := NewStatsService(storagePath, config.DefaultConfig())

	result, err := svc.Weekly()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check project breakdown
	if len(result.ProjectStats) < 2 {
		t.Errorf("expected at least 2 projects, got %d", len(result.ProjectStats))
	}

	// Check tag breakdown
	if len(result.TagStats) == 0 {
		t.Error("expected tag stats")
	}
}
