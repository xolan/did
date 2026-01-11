package service

import (
	"path/filepath"
	"testing"

	"github.com/xolan/did/internal/config"
)

func TestNewServicesWithPaths(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	configPath := filepath.Join(tmpDir, "config.toml")

	services := NewServicesWithPaths(storagePath, timerPath, configPath, config.DefaultConfig())

	if services == nil {
		t.Fatal("expected non-nil services")
	}
	if services.Entry == nil {
		t.Error("expected non-nil Entry service")
	}
	if services.Timer == nil {
		t.Error("expected non-nil Timer service")
	}
	if services.Report == nil {
		t.Error("expected non-nil Report service")
	}
	if services.Search == nil {
		t.Error("expected non-nil Search service")
	}
	if services.Stats == nil {
		t.Error("expected non-nil Stats service")
	}
	if services.Config == nil {
		t.Error("expected non-nil Config service")
	}
}

func TestNewServices(t *testing.T) {
	// NewServices uses default paths which depend on HOME env var
	// This test verifies it doesn't panic and returns expected structure
	services, err := NewServices()
	if err != nil {
		// May fail in CI environments without HOME, that's okay
		t.Skipf("NewServices failed (may be expected in CI): %v", err)
	}

	if services == nil {
		t.Fatal("expected non-nil services")
	}
	if services.Entry == nil {
		t.Error("expected non-nil Entry service")
	}
	if services.Timer == nil {
		t.Error("expected non-nil Timer service")
	}
	if services.Report == nil {
		t.Error("expected non-nil Report service")
	}
	if services.Search == nil {
		t.Error("expected non-nil Search service")
	}
	if services.Stats == nil {
		t.Error("expected non-nil Stats service")
	}
	if services.Config == nil {
		t.Error("expected non-nil Config service")
	}
}

func TestServicesIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	configPath := filepath.Join(tmpDir, "config.toml")

	services := NewServicesWithPaths(storagePath, timerPath, configPath, config.DefaultConfig())

	// Test Entry service
	entry, err := services.Entry.Create("test task for 1h")
	if err != nil {
		t.Fatalf("Entry.Create failed: %v", err)
	}
	if entry.Description != "test task" {
		t.Errorf("expected description 'test task', got %q", entry.Description)
	}

	// Test Timer service
	state, _, err := services.Timer.Start("timer task", false)
	if err != nil {
		t.Fatalf("Timer.Start failed: %v", err)
	}
	if state.Description != "timer task" {
		t.Errorf("expected description 'timer task', got %q", state.Description)
	}
	_, _ = services.Timer.Cancel()

	// Test Search service
	result, err := services.Search.Search("test", nil, nil)
	if err != nil {
		t.Fatalf("Search.Search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 search result, got %d", result.Total)
	}

	// Test Report service
	report, err := services.Report.GroupByProject(DateRangeSpec{Type: DateRangeToday})
	if err != nil {
		t.Fatalf("Report.GroupByProject failed: %v", err)
	}
	if report.EntryCount != 1 {
		t.Errorf("expected 1 entry in report, got %d", report.EntryCount)
	}

	// Test Stats service
	stats, err := services.Stats.Weekly()
	if err != nil {
		t.Fatalf("Stats.Weekly failed: %v", err)
	}
	if stats.Statistics.EntryCount != 1 {
		t.Errorf("expected 1 entry in stats, got %d", stats.Statistics.EntryCount)
	}

	// Test Config service
	cfg := services.Config.Get()
	if cfg.WeekStartDay == "" {
		t.Error("expected non-empty WeekStartDay")
	}
}
