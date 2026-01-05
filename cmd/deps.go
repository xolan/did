package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timer"
)

// Deps holds external dependencies for CLI commands, enabling testability.
type Deps struct {
	Stdout      io.Writer
	Stderr      io.Writer
	Stdin       io.Reader
	Exit        func(code int)
	StoragePath func() (string, error)
	TimerPath   func() (string, error)
	Config      config.Config
}

// DefaultDeps returns the default production dependencies.
func DefaultDeps() *Deps {
	// Load config from file or use defaults
	// Note: We don't call os.Exit() here to allow tests to work.
	// Config validation happens in ValidateConfigOnStartup() which is called from main.
	cfg := config.DefaultConfig()
	configPath, err := config.GetConfigPath()
	if err == nil {
		// Try to load config from file
		if loadedCfg, err := config.LoadOrDefault(configPath); err == nil {
			cfg = loadedCfg
		}
		// If there's an error, we use default config.
		// Validation will happen in ValidateConfigOnStartup() for production.
	}

	return &Deps{
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin,
		Exit:        os.Exit,
		StoragePath: storage.GetStoragePath,
		TimerPath:   timer.GetTimerPath,
		Config:      cfg,
	}
}

// ValidateConfigOnStartup checks if the config file is valid and shows helpful
// error messages if not. This should be called from main() before executing commands.
// Returns true if config is valid or doesn't exist, false if invalid.
func ValidateConfigOnStartup() bool {
	configPath, err := config.GetConfigPath()
	if err != nil {
		// Fatal error getting config path
		_, _ = fmt.Fprintln(os.Stderr, "Error: Failed to determine config file location")
		_, _ = fmt.Fprintf(os.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(os.Stderr, "Hint: Check that your home directory is accessible")
		return false
	}

	// Try to load config
	_, err = config.LoadOrDefault(configPath)
	if err != nil {
		// Config file exists but is invalid - show helpful error
		_, _ = fmt.Fprintln(os.Stderr, "Error: Failed to load configuration")
		_, _ = fmt.Fprintf(os.Stderr, "Details: %v\n", err)
		_, _ = fmt.Fprintln(os.Stderr)
		_, _ = fmt.Fprintf(os.Stderr, "Config file: %s\n", configPath)
		_, _ = fmt.Fprintln(os.Stderr)
		_, _ = fmt.Fprintln(os.Stderr, "Hint: Check that your config file is valid TOML format.")
		_, _ = fmt.Fprintln(os.Stderr, "Valid week_start_day values: monday, sunday")
		_, _ = fmt.Fprintln(os.Stderr, "Valid timezone examples: Local, America/New_York, Europe/London, Asia/Tokyo")
		_, _ = fmt.Fprintln(os.Stderr)
		_, _ = fmt.Fprintln(os.Stderr, "To see current config: did config")
		_, _ = fmt.Fprintln(os.Stderr, "To create a fresh sample config: did config --init")
		return false
	}

	return true
}

// deps is the global dependencies instance used by commands.
// In production, this is DefaultDeps(). Tests can replace it.
var deps = DefaultDeps()

// SetDeps sets the global dependencies (for testing).
func SetDeps(d *Deps) {
	deps = d
}

// ResetDeps resets dependencies to defaults (for testing cleanup).
func ResetDeps() {
	deps = DefaultDeps()
}
