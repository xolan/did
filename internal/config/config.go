package config

import (
	"os"
	"path/filepath"
)

const (
	// AppName is the application name used for config directory
	AppName = "did"
	// ConfigFile is the name of the TOML configuration file
	ConfigFile = "config.toml"
)

// Config represents the application configuration
type Config struct {
	// WeekStartDay defines which day starts the week (monday or sunday)
	WeekStartDay string `toml:"week_start_day"`
	// Timezone defines the timezone for time operations (IANA timezone name, e.g., "America/New_York")
	Timezone string `toml:"timezone"`
	// DefaultOutputFormat defines the default output format for entries
	DefaultOutputFormat string `toml:"default_output_format"`
}

// DefaultConfig returns a Config with sensible defaults that match current behavior.
// - week_start_day: "monday" (ISO 8601 standard, current behavior)
// - timezone: "Local" (use system local timezone)
// - default_output_format: "" (use current default formatting)
func DefaultConfig() Config {
	return Config{
		WeekStartDay:        "monday",
		Timezone:            "Local",
		DefaultOutputFormat: "",
	}
}

// GetConfigPath returns the path to the config file.
// Uses os.UserConfigDir() for cross-platform XDG-compliant config directory.
// Creates the config directory if it doesn't exist.
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, AppName)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, ConfigFile), nil
}
