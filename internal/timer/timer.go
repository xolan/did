package timer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/xolan/did/internal/osutil"
)

const (
	// AppName is the application name used for config directory
	AppName = "did"
	// TimerFile is the name of the JSON timer state file
	TimerFile = "timer.json"
)

// TimerState represents the state of an active timer
type TimerState struct {
	StartedAt   time.Time `json:"started_at"`
	Description string    `json:"description"`
	Project     string    `json:"project,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

// GetTimerPath returns the path to the timer state file.
// Uses os.UserConfigDir() for cross-platform XDG-compliant config directory.
// Creates the config directory if it doesn't exist.
func GetTimerPath() (string, error) {
	configDir, err := osutil.Provider.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, AppName)

	// Create config directory if it doesn't exist
	if err := osutil.Provider.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, TimerFile), nil
}

// SaveTimerState writes the timer state to the timer file.
// Overwrites the file if it exists. Creates the file with 0644 permissions.
// Uses atomic write pattern (write to temp file, then rename) for safety.
func SaveTimerState(filepath string, state TimerState) error {
	// Marshal state to JSON
	// TimerState struct contains only JSON-safe types, so Marshal cannot fail
	data, _ := json.MarshalIndent(state, "", "  ")

	// Write to temporary file
	tmpFile := filepath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	return os.Rename(tmpFile, filepath)
}

// LoadTimerState reads the timer state from the timer file.
// Returns nil if the file doesn't exist (no active timer).
// Returns an error if the file exists but cannot be read or parsed.
func LoadTimerState(filepath string) (*TimerState, error) {
	// Check if file exists
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal JSON
	var state TimerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// ClearTimerState removes the timer state file.
// Returns nil if the file doesn't exist (idempotent operation).
func ClearTimerState(filepath string) error {
	err := os.Remove(filepath)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

// IsTimerRunning checks if an active timer exists.
// Returns true if a valid timer state file exists, false otherwise.
func IsTimerRunning(filepath string) (bool, error) {
	state, err := LoadTimerState(filepath)
	if err != nil {
		return false, err
	}
	return state != nil, nil
}
