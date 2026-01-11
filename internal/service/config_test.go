package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xolan/did/internal/config"
)

func TestNewConfigService(t *testing.T) {
	svc := NewConfigService("/tmp/config.toml", config.DefaultConfig())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestConfigService_Get(t *testing.T) {
	cfg := config.DefaultConfig()
	svc := NewConfigService("/tmp/config.toml", cfg)

	result := svc.Get()
	if result.WeekStartDay != cfg.WeekStartDay {
		t.Errorf("expected WeekStartDay %q, got %q", cfg.WeekStartDay, result.WeekStartDay)
	}
	if result.Timezone != cfg.Timezone {
		t.Errorf("expected Timezone %q, got %q", cfg.Timezone, result.Timezone)
	}
}

func TestConfigService_GetPath(t *testing.T) {
	svc := NewConfigService("/tmp/test/config.toml", config.DefaultConfig())

	path := svc.GetPath()
	if path != "/tmp/test/config.toml" {
		t.Errorf("expected path '/tmp/test/config.toml', got %q", path)
	}
}

func TestConfigService_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	svc := NewConfigService(configPath, config.DefaultConfig())

	// File doesn't exist yet
	if svc.Exists() {
		t.Error("expected Exists() to return false")
	}

	// Create the file
	if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Now it exists
	if !svc.Exists() {
		t.Error("expected Exists() to return true")
	}
}

func TestConfigService_Update(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	svc := NewConfigService(configPath, config.DefaultConfig())

	// Update config
	newCfg := config.Config{
		WeekStartDay: "sunday",
		Timezone:     "America/New_York",
	}

	err := svc.Update(newCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify in-memory config was updated
	result := svc.Get()
	if result.WeekStartDay != "sunday" {
		t.Errorf("expected WeekStartDay 'sunday', got %q", result.WeekStartDay)
	}
	if result.Timezone != "America/New_York" {
		t.Errorf("expected Timezone 'America/New_York', got %q", result.Timezone)
	}

	// Verify file was written
	if !svc.Exists() {
		t.Error("expected config file to exist")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 {
		t.Error("expected non-empty config file")
	}
}

func TestConfigService_Update_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	svc := NewConfigService(configPath, config.DefaultConfig())

	// Invalid week start day
	invalidCfg := config.Config{
		WeekStartDay: "invalid",
		Timezone:     "Local",
	}

	err := svc.Update(invalidCfg)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestConfigService_Init(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	svc := NewConfigService(configPath, config.DefaultConfig())

	// Init should create a sample config
	err := svc.Init()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should exist
	if !svc.Exists() {
		t.Error("expected config file to exist after Init")
	}

	// File should have content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 {
		t.Error("expected non-empty config file after Init")
	}
}

func TestConfigService_Init_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create existing file
	if err := os.WriteFile(configPath, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewConfigService(configPath, config.DefaultConfig())

	// Init should fail if file already exists
	err := svc.Init()
	if err == nil {
		t.Error("expected error when config file already exists")
	}
}

func TestConfigService_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create initial config
	initialContent := `
week_start_day = "monday"
timezone = "Local"
`
	if err := os.WriteFile(configPath, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewConfigService(configPath, config.DefaultConfig())

	// Reload
	err := svc.Reload()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := svc.Get()
	if result.WeekStartDay != "monday" {
		t.Errorf("expected WeekStartDay 'monday', got %q", result.WeekStartDay)
	}
}

func TestConfigService_Reload_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create invalid config
	if err := os.WriteFile(configPath, []byte("invalid toml {{{"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewConfigService(configPath, config.DefaultConfig())

	err := svc.Reload()
	if err == nil {
		t.Error("expected error for invalid config file")
	}
}

func TestConfigService_writeConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	svc := NewConfigService(configPath, config.DefaultConfig())

	cfg := config.Config{
		WeekStartDay: "sunday",
		Timezone:     "Europe/London",
	}

	err := svc.writeConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "sunday") {
		t.Error("expected config to contain 'sunday'")
	}
	if !contains(contentStr, "Europe/London") {
		t.Error("expected config to contain 'Europe/London'")
	}
}

func TestConfigService_Update_WriteError(t *testing.T) {
	// Use a path that can't be written to
	svc := NewConfigService("/nonexistent/dir/config.toml", config.DefaultConfig())

	err := svc.Update(config.DefaultConfig())
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestConfigService_Init_WriteError(t *testing.T) {
	// Use a path that can't be written to
	svc := NewConfigService("/nonexistent/dir/config.toml", config.DefaultConfig())

	err := svc.Init()
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
