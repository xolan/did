package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
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

// Validate checks if the Config values are valid and returns helpful error messages.
// Validates that:
// - week_start_day is either "monday" or "sunday" (case-insensitive)
// - timezone is a valid IANA timezone name (e.g., "America/New_York") or "Local"
func (c *Config) Validate() error {
	// Normalize week_start_day to lowercase for comparison
	weekStartDay := strings.ToLower(strings.TrimSpace(c.WeekStartDay))
	if weekStartDay != "monday" && weekStartDay != "sunday" {
		return fmt.Errorf("invalid week_start_day: must be 'monday' or 'sunday', got '%s'", c.WeekStartDay)
	}
	// Normalize the value in the config
	c.WeekStartDay = weekStartDay

	// Validate timezone
	if c.Timezone != "" && c.Timezone != "Local" {
		// Try to load the timezone to validate it exists
		_, err := time.LoadLocation(c.Timezone)
		if err != nil {
			return fmt.Errorf("invalid timezone: '%s' is not a valid IANA timezone (e.g., 'America/New_York', 'Europe/London')", c.Timezone)
		}
	}

	return nil
}

// Load reads and parses the TOML config file at the given path.
// Returns an error if the file cannot be read or parsed, or if validation fails.
// The returned Config is validated and normalized (e.g., week_start_day is lowercase).
// Empty fields in the config file are replaced with default values.
func Load(path string) (Config, error) {
	// Start with default config
	cfg := DefaultConfig()

	// Read the TOML file, which will overwrite only the fields present
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the loaded config
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// LoadOrDefault attempts to load the config from the given path.
// If the file doesn't exist, returns DefaultConfig() instead.
// Returns an error if the file exists but cannot be parsed or is invalid.
func LoadOrDefault(path string) (Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return default config
			return DefaultConfig(), nil
		}
		// Some other error occurred while checking the file
		return Config{}, err
	}

	// File exists, try to load it
	return Load(path)
}

// GenerateSampleConfig returns a sample TOML configuration file content
// with all options commented out and documented with explanations and examples.
func GenerateSampleConfig() string {
	return `# did configuration file
# This file is optional - did works perfectly without any configuration.
# All settings have sensible defaults. Uncomment and modify only the
# settings you want to customize.

# ============================================================================
# Week Start Day
# ============================================================================
# Defines which day starts the week for weekly views (w, lw commands)
# and statistics (stats command).
#
# Valid values: "monday", "sunday"
# Default: "monday" (ISO 8601 standard)
#
# Examples:
#   week_start_day = "monday"    # Week starts Monday (default)
#   week_start_day = "sunday"    # Week starts Sunday (US convention)
#
# week_start_day = "monday"

# ============================================================================
# Timezone
# ============================================================================
# Defines the timezone for time operations and display.
# Uses IANA timezone names (e.g., "America/New_York", "Europe/London").
#
# Valid values: Any IANA timezone name or "Local" for system timezone
# Default: "Local" (uses your system's timezone)
#
# Examples:
#   timezone = "Local"              # Use system timezone (default)
#   timezone = "America/New_York"   # Eastern Time
#   timezone = "Europe/London"      # British Time
#   timezone = "Asia/Tokyo"         # Japan Time
#   timezone = "UTC"                # Coordinated Universal Time
#
# To see available timezones, check:
#   https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
#
# timezone = "Local"

# ============================================================================
# Default Output Format
# ============================================================================
# Defines the default output format for entry listings.
# This setting is reserved for future use when custom output formats
# are implemented.
#
# Default: "" (uses the built-in default format)
#
# Examples:
#   default_output_format = ""      # Use default format (default)
#
# default_output_format = ""
`
}
