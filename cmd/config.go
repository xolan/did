package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display or manage configuration settings",
	Long: `Display the current effective configuration settings for did.

Shows the configuration file location, whether it exists, and all current settings.
Configuration values are merged from the config file with sensible defaults.

By default, did works without any configuration file. All settings have defaults:
  - week_start_day: monday
  - timezone: Local (system timezone)
  - default_output_format: (empty, uses default format)

Examples:

  Display current configuration:
    did config                       Show all current settings

Configuration file location:
  ~/.config/did/config.toml          Linux/macOS
  %APPDATA%\did\config.toml          Windows

To customize settings, create a config.toml file at the location shown above.`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

// showConfig displays the current effective configuration
func showConfig() {
	// Get config file path
	configPath, err := config.GetConfigPath()
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to determine config file location")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: Check that your home directory is accessible")
		deps.Exit(1)
		return
	}

	// Check if config file exists
	fileExists := false
	if _, err := os.Stat(configPath); err == nil {
		fileExists = true
	}

	// Load config (will use defaults if file doesn't exist)
	cfg, err := config.LoadOrDefault(configPath)
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to load configuration")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr)
		_, _ = fmt.Fprintf(deps.Stderr, "Hint: Check that your config file is valid TOML format: %s\n", configPath)
		_, _ = fmt.Fprintln(deps.Stderr, "Valid week_start_day values: monday, sunday")
		_, _ = fmt.Fprintln(deps.Stderr, "Valid timezone examples: Local, America/New_York, Europe/London, Asia/Tokyo")
		deps.Exit(1)
		return
	}

	// Display header
	_, _ = fmt.Fprintln(deps.Stdout, "Configuration for did")
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 60))
	_, _ = fmt.Fprintln(deps.Stdout)

	// Display config file location and status
	_, _ = fmt.Fprintf(deps.Stdout, "Config file:     %s\n", configPath)
	if fileExists {
		_, _ = fmt.Fprintln(deps.Stdout, "Status:          File exists (using custom configuration)")
	} else {
		_, _ = fmt.Fprintln(deps.Stdout, "Status:          No config file (using defaults)")
	}
	_, _ = fmt.Fprintln(deps.Stdout)

	// Display current settings
	_, _ = fmt.Fprintln(deps.Stdout, "Current Settings:")
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 60))
	_, _ = fmt.Fprintf(deps.Stdout, "Week Start Day:  %s\n", cfg.WeekStartDay)
	_, _ = fmt.Fprintf(deps.Stdout, "Timezone:        %s\n", cfg.Timezone)

	// Display default_output_format with special handling for empty value
	if cfg.DefaultOutputFormat == "" {
		_, _ = fmt.Fprintln(deps.Stdout, "Output Format:   (default)")
	} else {
		_, _ = fmt.Fprintf(deps.Stdout, "Output Format:   %s\n", cfg.DefaultOutputFormat)
	}

	_, _ = fmt.Fprintln(deps.Stdout)

	// Display helpful information if using defaults
	if !fileExists {
		_, _ = fmt.Fprintln(deps.Stdout, "Tip: Create a config.toml file at the above location to customize settings.")
		_, _ = fmt.Fprintln(deps.Stdout, "     See documentation for available options and examples.")
		_, _ = fmt.Fprintln(deps.Stdout)
	}
}
