package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xolan/did/internal/app"
	"github.com/xolan/did/internal/osutil"
)

// Helper to create a temporary config file
func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.toml")
	// Always write the file, even if content is empty
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	return tmpFile
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify default week_start_day
	if cfg.WeekStartDay != "monday" {
		t.Errorf("DefaultConfig().WeekStartDay = %q, expected %q", cfg.WeekStartDay, "monday")
	}

	// Verify default timezone
	if cfg.Timezone != "Local" {
		t.Errorf("DefaultConfig().Timezone = %q, expected %q", cfg.Timezone, "Local")
	}

	// Verify default output format (empty string is default)
	if cfg.DefaultOutputFormat != "" {
		t.Errorf("DefaultConfig().DefaultOutputFormat = %q, expected %q", cfg.DefaultOutputFormat, "")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	tests := []struct {
		name              string
		configContent     string
		expectedWeekStart string
		expectedTimezone  string
		expectedOutputFmt string
	}{
		{
			name: "all fields set",
			configContent: `week_start_day = "sunday"
timezone = "America/New_York"
default_output_format = "json"`,
			expectedWeekStart: "sunday",
			expectedTimezone:  "America/New_York",
			expectedOutputFmt: "json",
		},
		{
			name: "monday week start",
			configContent: `week_start_day = "monday"
timezone = "Local"`,
			expectedWeekStart: "monday",
			expectedTimezone:  "Local",
			expectedOutputFmt: "",
		},
		{
			name: "different timezone",
			configContent: `week_start_day = "monday"
timezone = "Europe/London"`,
			expectedWeekStart: "monday",
			expectedTimezone:  "Europe/London",
			expectedOutputFmt: "",
		},
		{
			name: "mixed case week_start_day normalized",
			configContent: `week_start_day = "Sunday"
timezone = "Local"`,
			expectedWeekStart: "sunday",
			expectedTimezone:  "Local",
			expectedOutputFmt: "",
		},
		{
			name: "uppercase week_start_day normalized",
			configContent: `week_start_day = "MONDAY"
timezone = "Local"`,
			expectedWeekStart: "monday",
			expectedTimezone:  "Local",
			expectedOutputFmt: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempConfigFile(t, tt.configContent)

			cfg, err := Load(tmpFile)
			if err != nil {
				t.Fatalf("Load() returned unexpected error: %v", err)
			}

			if cfg.WeekStartDay != tt.expectedWeekStart {
				t.Errorf("WeekStartDay = %q, expected %q", cfg.WeekStartDay, tt.expectedWeekStart)
			}
			if cfg.Timezone != tt.expectedTimezone {
				t.Errorf("Timezone = %q, expected %q", cfg.Timezone, tt.expectedTimezone)
			}
			if cfg.DefaultOutputFormat != tt.expectedOutputFmt {
				t.Errorf("DefaultOutputFormat = %q, expected %q", cfg.DefaultOutputFormat, tt.expectedOutputFmt)
			}
		})
	}
}

func TestLoad_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.toml")

	_, err := Load(nonExistentFile)
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
	}{
		{
			name:          "malformed TOML",
			configContent: `week_start_day = "monday`,
		},
		{
			name:          "invalid syntax",
			configContent: `this is not valid TOML at all`,
		},
		{
			name:          "missing quotes",
			configContent: `week_start_day = monday`,
		},
		{
			name: "unclosed brackets",
			configContent: `[section
week_start_day = "monday"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempConfigFile(t, tt.configContent)

			_, err := Load(tmpFile)
			if err == nil {
				t.Error("Load() should return error for invalid TOML")
			}
			if !strings.Contains(err.Error(), "failed to parse config file") {
				t.Errorf("Error message should mention parsing failure, got: %v", err)
			}
		})
	}
}

func TestLoad_InvalidWeekStartDay(t *testing.T) {
	tests := []struct {
		name           string
		weekStartDay   string
		errorSubstring string
	}{
		{
			name:           "invalid day",
			weekStartDay:   "tuesday",
			errorSubstring: "invalid week_start_day",
		},
		{
			name:           "empty string",
			weekStartDay:   "",
			errorSubstring: "invalid week_start_day",
		},
		{
			name:           "number",
			weekStartDay:   "1",
			errorSubstring: "invalid week_start_day",
		},
		{
			name:           "misspelled",
			weekStartDay:   "munday",
			errorSubstring: "invalid week_start_day",
		},
		{
			name:           "abbreviated",
			weekStartDay:   "mon",
			errorSubstring: "invalid week_start_day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configContent := `week_start_day = "` + tt.weekStartDay + `"
timezone = "Local"`
			tmpFile := createTempConfigFile(t, configContent)

			_, err := Load(tmpFile)
			if err == nil {
				t.Errorf("Load() should return error for invalid week_start_day: %q", tt.weekStartDay)
			}
			if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("Error should contain %q, got: %v", tt.errorSubstring, err)
			}
		})
	}
}

func TestLoad_InvalidTimezone(t *testing.T) {
	tests := []struct {
		name           string
		timezone       string
		errorSubstring string
	}{
		{
			name:           "invalid timezone",
			timezone:       "Invalid/Timezone",
			errorSubstring: "invalid timezone",
		},
		{
			name:           "non-existent location",
			timezone:       "Mars/Olympus",
			errorSubstring: "invalid timezone",
		},
		{
			name:           "random string",
			timezone:       "not_a_timezone",
			errorSubstring: "invalid timezone",
		},
		{
			name:           "number",
			timezone:       "123",
			errorSubstring: "invalid timezone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configContent := `week_start_day = "monday"
timezone = "` + tt.timezone + `"`
			tmpFile := createTempConfigFile(t, configContent)

			_, err := Load(tmpFile)
			if err == nil {
				t.Errorf("Load() should return error for invalid timezone: %q", tt.timezone)
			}
			if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("Error should contain %q, got: %v", tt.errorSubstring, err)
			}
		})
	}
}

func TestLoad_ValidTimezones(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
	}{
		{"Local timezone", "Local"},
		{"US Eastern", "America/New_York"},
		{"US Pacific", "America/Los_Angeles"},
		{"UK", "Europe/London"},
		{"Japan", "Asia/Tokyo"},
		{"Australia", "Australia/Sydney"},
		{"UTC", "UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configContent := `week_start_day = "monday"
timezone = "` + tt.timezone + `"`
			tmpFile := createTempConfigFile(t, configContent)

			cfg, err := Load(tmpFile)
			if err != nil {
				t.Fatalf("Load() returned unexpected error for valid timezone %q: %v", tt.timezone, err)
			}
			if cfg.Timezone != tt.timezone {
				t.Errorf("Timezone = %q, expected %q", cfg.Timezone, tt.timezone)
			}
		})
	}
}

func TestLoadOrDefault_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.toml")

	cfg, err := LoadOrDefault(nonExistentFile)
	if err != nil {
		t.Fatalf("LoadOrDefault() returned unexpected error for non-existent file: %v", err)
	}

	// Should return default config
	defaultCfg := DefaultConfig()
	if cfg.WeekStartDay != defaultCfg.WeekStartDay {
		t.Errorf("WeekStartDay = %q, expected default %q", cfg.WeekStartDay, defaultCfg.WeekStartDay)
	}
	if cfg.Timezone != defaultCfg.Timezone {
		t.Errorf("Timezone = %q, expected default %q", cfg.Timezone, defaultCfg.Timezone)
	}
	if cfg.DefaultOutputFormat != defaultCfg.DefaultOutputFormat {
		t.Errorf("DefaultOutputFormat = %q, expected default %q", cfg.DefaultOutputFormat, defaultCfg.DefaultOutputFormat)
	}
}

func TestLoadOrDefault_ExistingValidFile(t *testing.T) {
	configContent := `week_start_day = "sunday"
timezone = "America/New_York"`
	tmpFile := createTempConfigFile(t, configContent)

	cfg, err := LoadOrDefault(tmpFile)
	if err != nil {
		t.Fatalf("LoadOrDefault() returned unexpected error: %v", err)
	}

	// Should load from file, not use defaults
	if cfg.WeekStartDay != "sunday" {
		t.Errorf("WeekStartDay = %q, expected %q", cfg.WeekStartDay, "sunday")
	}
	if cfg.Timezone != "America/New_York" {
		t.Errorf("Timezone = %q, expected %q", cfg.Timezone, "America/New_York")
	}
}

func TestLoadOrDefault_ExistingInvalidFile(t *testing.T) {
	// Invalid config file should return error, not default
	configContent := `week_start_day = "tuesday"
timezone = "Local"`
	tmpFile := createTempConfigFile(t, configContent)

	_, err := LoadOrDefault(tmpFile)
	if err == nil {
		t.Error("LoadOrDefault() should return error for invalid config file")
	}
	if !strings.Contains(err.Error(), "invalid week_start_day") {
		t.Errorf("Error should mention invalid week_start_day, got: %v", err)
	}
}

func TestValidate_ValidConfigs(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantLower string // expected normalized week_start_day
	}{
		{
			name: "monday lowercase",
			config: Config{
				WeekStartDay: "monday",
				Timezone:     "Local",
			},
			wantLower: "monday",
		},
		{
			name: "sunday lowercase",
			config: Config{
				WeekStartDay: "sunday",
				Timezone:     "Local",
			},
			wantLower: "sunday",
		},
		{
			name: "Monday mixed case normalized to lowercase",
			config: Config{
				WeekStartDay: "Monday",
				Timezone:     "Local",
			},
			wantLower: "monday",
		},
		{
			name: "SUNDAY uppercase normalized to lowercase",
			config: Config{
				WeekStartDay: "SUNDAY",
				Timezone:     "Local",
			},
			wantLower: "sunday",
		},
		{
			name: "with whitespace trimmed",
			config: Config{
				WeekStartDay: "  monday  ",
				Timezone:     "Local",
			},
			wantLower: "monday",
		},
		{
			name: "valid timezone",
			config: Config{
				WeekStartDay: "monday",
				Timezone:     "America/New_York",
			},
			wantLower: "monday",
		},
		{
			name: "empty timezone is valid",
			config: Config{
				WeekStartDay: "monday",
				Timezone:     "",
			},
			wantLower: "monday",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.Normalize()
			err := tt.config.Validate()
			if err != nil {
				t.Errorf("Validate() returned unexpected error: %v", err)
			}
			if tt.config.WeekStartDay != tt.wantLower {
				t.Errorf("After Normalize()+Validate(), WeekStartDay = %q, expected %q", tt.config.WeekStartDay, tt.wantLower)
			}
		})
	}
}

func TestValidate_InvalidWeekStartDay(t *testing.T) {
	tests := []struct {
		name         string
		weekStartDay string
	}{
		{"empty string", ""},
		{"tuesday", "tuesday"},
		{"wednesday", "wednesday"},
		{"invalid", "invalid"},
		{"number", "1"},
		{"mon abbreviation", "mon"},
		{"sun abbreviation", "sun"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				WeekStartDay: tt.weekStartDay,
				Timezone:     "Local",
			}
			err := cfg.Validate()
			if err == nil {
				t.Errorf("Validate() should return error for week_start_day %q", tt.weekStartDay)
			}
			if !strings.Contains(err.Error(), "invalid week_start_day") {
				t.Errorf("Error should contain 'invalid week_start_day', got: %v", err)
			}
		})
	}
}

func TestValidate_InvalidTimezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
	}{
		{"invalid location", "Invalid/Timezone"},
		{"non-existent", "Mars/Olympus"},
		{"random string", "not_a_timezone"},
		{"number", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				WeekStartDay: "monday",
				Timezone:     tt.timezone,
			}
			err := cfg.Validate()
			if err == nil {
				t.Errorf("Validate() should return error for timezone %q", tt.timezone)
			}
			if !strings.Contains(err.Error(), "invalid timezone") {
				t.Errorf("Error should contain 'invalid timezone', got: %v", err)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned unexpected error: %v", err)
	}

	// Path should not be empty
	if path == "" {
		t.Error("GetConfigPath() returned empty path")
	}

	// Path should end with config.toml
	if filepath.Base(path) != ConfigFile {
		t.Errorf("GetConfigPath() path base = %q, expected %q", filepath.Base(path), ConfigFile)
	}

	// Parent directory should exist (created by GetConfigPath)
	parentDir := filepath.Dir(path)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Errorf("GetConfigPath() parent directory does not exist: %v", err)
	}
	if info != nil && !info.IsDir() {
		t.Error("GetConfigPath() parent is not a directory")
	}

	// Directory name should contain app name
	if !strings.Contains(parentDir, app.Name) {
		t.Errorf("GetConfigPath() parent directory should contain %q, got %q", app.Name, parentDir)
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are set correctly
	if app.Name != "did" {
		t.Errorf("app.Name = %q, expected %q", app.Name, "did")
	}

	if ConfigFile != "config.toml" {
		t.Errorf("ConfigFile = %q, expected %q", ConfigFile, "config.toml")
	}
}

func TestLoad_PartialConfig(t *testing.T) {
	defaultCfg := DefaultConfig()

	tests := []struct {
		name              string
		configContent     string
		expectedWeekStart string
		expectedTimezone  string
		expectedOutputFmt string
	}{
		{
			name:              "only week_start_day",
			configContent:     `week_start_day = "sunday"`,
			expectedWeekStart: "sunday",
			expectedTimezone:  defaultCfg.Timezone, // Should merge with default
			expectedOutputFmt: defaultCfg.DefaultOutputFormat,
		},
		{
			name:              "only timezone",
			configContent:     `timezone = "America/New_York"`,
			expectedWeekStart: defaultCfg.WeekStartDay, // Should merge with default
			expectedTimezone:  "America/New_York",
			expectedOutputFmt: defaultCfg.DefaultOutputFormat,
		},
		{
			name:              "only output format",
			configContent:     `default_output_format = "json"`,
			expectedWeekStart: defaultCfg.WeekStartDay, // Should merge with default
			expectedTimezone:  defaultCfg.Timezone,
			expectedOutputFmt: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempConfigFile(t, tt.configContent)

			cfg, err := Load(tmpFile)
			if err != nil {
				t.Fatalf("Load() returned unexpected error: %v", err)
			}

			if cfg.WeekStartDay != tt.expectedWeekStart {
				t.Errorf("WeekStartDay = %q, expected %q", cfg.WeekStartDay, tt.expectedWeekStart)
			}
			if cfg.Timezone != tt.expectedTimezone {
				t.Errorf("Timezone = %q, expected %q", cfg.Timezone, tt.expectedTimezone)
			}
			if cfg.DefaultOutputFormat != tt.expectedOutputFmt {
				t.Errorf("DefaultOutputFormat = %q, expected %q", cfg.DefaultOutputFormat, tt.expectedOutputFmt)
			}
		})
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpFile := createTempConfigFile(t, "")

	// Empty file should now merge with defaults and work correctly
	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Should have default values
	defaultCfg := DefaultConfig()
	if cfg.WeekStartDay != defaultCfg.WeekStartDay {
		t.Errorf("WeekStartDay = %q, expected %q", cfg.WeekStartDay, defaultCfg.WeekStartDay)
	}
	if cfg.Timezone != defaultCfg.Timezone {
		t.Errorf("Timezone = %q, expected %q", cfg.Timezone, defaultCfg.Timezone)
	}
	if cfg.DefaultOutputFormat != defaultCfg.DefaultOutputFormat {
		t.Errorf("DefaultOutputFormat = %q, expected %q", cfg.DefaultOutputFormat, defaultCfg.DefaultOutputFormat)
	}
}

func TestLoad_UnreadableFile(t *testing.T) {
	tmpFile := createTempConfigFile(t, `week_start_day = "monday"`)

	// Make file unreadable
	if err := os.Chmod(tmpFile, 0000); err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}
	defer func() { _ = os.Chmod(tmpFile, 0644) }()

	_, err := Load(tmpFile)
	if err == nil {
		t.Error("Load() should return error for unreadable file")
	}
}

func TestLoadOrDefault_PermissionError(t *testing.T) {
	tmpFile := createTempConfigFile(t, `week_start_day = "monday"`)

	// Make file unreadable
	if err := os.Chmod(tmpFile, 0000); err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}
	defer func() { _ = os.Chmod(tmpFile, 0644) }()

	// LoadOrDefault should return error for permission issues (not default config)
	_, err := LoadOrDefault(tmpFile)
	if err == nil {
		t.Error("LoadOrDefault() should return error for unreadable file")
	}
}

func TestValidate_NormalizesWeekStartDay(t *testing.T) {
	// Test that Validate modifies the Config in place to normalize week_start_day
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase monday unchanged", "monday", "monday"},
		{"lowercase sunday unchanged", "sunday", "sunday"},
		{"uppercase MONDAY normalized", "MONDAY", "monday"},
		{"uppercase SUNDAY normalized", "SUNDAY", "sunday"},
		{"mixed case Monday normalized", "Monday", "monday"},
		{"mixed case SuNdAy normalized", "SuNdAy", "sunday"},
		{"with leading space", " monday", "monday"},
		{"with trailing space", "sunday ", "sunday"},
		{"with both spaces", "  Monday  ", "monday"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				WeekStartDay: tt.input,
				Timezone:     "Local",
			}

			cfg.Normalize()
			err := cfg.Validate()
			if err != nil {
				t.Fatalf("Normalize()+Validate() returned unexpected error: %v", err)
			}

			if cfg.WeekStartDay != tt.expected {
				t.Errorf("After Normalize(), WeekStartDay = %q, expected %q", cfg.WeekStartDay, tt.expected)
			}
		})
	}
}

func TestGenerateSampleConfig(t *testing.T) {
	content := GenerateSampleConfig()

	// Verify content is not empty
	if content == "" {
		t.Error("GenerateSampleConfig() returned empty string")
	}

	// Verify it contains expected sections
	expectedStrings := []string{
		"# did configuration file",
		"week_start_day",
		"timezone",
		"default_output_format",
		"monday",
		"sunday",
		"Local",
		"America/New_York",
		"Europe/London",
		"Asia/Tokyo",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("GenerateSampleConfig() missing expected content: %q", expected)
		}
	}

	// Verify it's commented out by default (values should be in comments)
	if !strings.Contains(content, "# week_start_day") {
		t.Error("GenerateSampleConfig() week_start_day should be commented out")
	}
	if !strings.Contains(content, "# timezone") {
		t.Error("GenerateSampleConfig() timezone should be commented out")
	}
}

func TestGetConfigPath_UserConfigDirError(t *testing.T) {
	// Save original provider and ensure cleanup
	defer osutil.ResetProvider()

	// Mock UserConfigDir to return an error
	osutil.SetProvider(&mockPathProvider{
		userConfigDirFn: func() (string, error) {
			return "", os.ErrPermission
		},
	})

	_, err := GetConfigPath()
	if err == nil {
		t.Error("GetConfigPath() should return error when UserConfigDir fails")
	}
}

func TestGetConfigPath_MkdirAllError(t *testing.T) {
	// Save original provider
	original := osutil.Provider
	defer osutil.ResetProvider()

	tmpDir := t.TempDir()

	// Mock MkdirAll to return an error
	osutil.SetProvider(&mockPathProvider{
		userConfigDirFn: func() (string, error) {
			return tmpDir, nil
		},
		mkdirAllFn: func(path string, perm os.FileMode) error {
			return os.ErrPermission
		},
	})

	_, err := GetConfigPath()
	if err == nil {
		t.Error("GetConfigPath() should return error when MkdirAll fails")
	}

	// Reset
	osutil.Provider = original
}

func TestLoadOrDefault_StatError(t *testing.T) {
	// Create a directory structure where we can trigger a stat error
	// by making the parent directory unreadable
	tmpDir := t.TempDir()
	parentDir := filepath.Join(tmpDir, "parent")
	if err := os.Mkdir(parentDir, 0755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}

	configPath := filepath.Join(parentDir, "config.toml")

	// Make parent directory unreadable (this will cause stat to fail with permission error)
	if err := os.Chmod(parentDir, 0000); err != nil {
		t.Skipf("Cannot change directory permissions: %v", err)
	}
	defer func() { _ = os.Chmod(parentDir, 0755) }()

	// LoadOrDefault should return error (not default config) because stat fails
	_, err := LoadOrDefault(configPath)
	if err == nil {
		t.Error("LoadOrDefault() should return error when os.Stat fails with permission error")
	}
}

// mockPathProvider is a test helper for mocking osutil.PathProvider
type mockPathProvider struct {
	userConfigDirFn func() (string, error)
	mkdirAllFn      func(path string, perm os.FileMode) error
}

func (m *mockPathProvider) UserConfigDir() (string, error) {
	if m.userConfigDirFn != nil {
		return m.userConfigDirFn()
	}
	return "", nil
}

func (m *mockPathProvider) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirAllFn != nil {
		return m.mkdirAllFn(path, perm)
	}
	return nil
}
