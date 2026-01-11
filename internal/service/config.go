package service

import (
	"fmt"
	"os"

	"github.com/xolan/did/internal/config"
)

// ConfigService provides operations for managing configuration
type ConfigService struct {
	configPath string
	config     config.Config
}

// NewConfigService creates a new ConfigService
func NewConfigService(configPath string, cfg config.Config) *ConfigService {
	return &ConfigService{
		configPath: configPath,
		config:     cfg,
	}
}

// Get returns the current configuration
func (s *ConfigService) Get() config.Config {
	return s.config
}

// GetPath returns the path to the config file
func (s *ConfigService) GetPath() string {
	return s.configPath
}

// Exists checks if the config file exists
func (s *ConfigService) Exists() bool {
	_, err := os.Stat(s.configPath)
	return err == nil
}

// Update updates the configuration with new values
func (s *ConfigService) Update(cfg config.Config) error {
	// Normalize and validate
	cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Write the config file
	if err := s.writeConfig(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Update in-memory config
	s.config = cfg

	return nil
}

// Init creates a sample config file
func (s *ConfigService) Init() error {
	// Check if file already exists
	if s.Exists() {
		return fmt.Errorf("config file already exists at %s", s.configPath)
	}

	// Write sample config
	sample := config.GenerateSampleConfig()
	if err := os.WriteFile(s.configPath, []byte(sample), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Reload reloads the configuration from disk
func (s *ConfigService) Reload() error {
	cfg, err := config.LoadOrDefault(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	s.config = cfg
	return nil
}

// writeConfig writes the config to the config file in TOML format
func (s *ConfigService) writeConfig(cfg config.Config) error {
	content := fmt.Sprintf(`# did configuration file

# Week start day: "monday" or "sunday"
week_start_day = %q

# Timezone: IANA timezone name (e.g., "America/New_York") or "Local"
timezone = %q
`, cfg.WeekStartDay, cfg.Timezone)

	return os.WriteFile(s.configPath, []byte(content), 0644)
}
