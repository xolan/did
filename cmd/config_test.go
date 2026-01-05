package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/xolan/did/internal/config"
)

// TestShowConfig_NoConfigFile tests showing config when no config file exists (uses defaults)
func TestShowConfig_NoConfigFile(t *testing.T) {
	// Get the real config path but ensure the config file doesn't exist
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
		// Remove the config file temporarily
		if err := os.Remove(configPath); err != nil {
			t.Fatalf("Failed to remove existing config for test: %v", err)
		}
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	showConfig()

	output := stdout.String()

	// Check that it indicates no config file exists
	if !strings.Contains(output, "No config file") {
		t.Errorf("Expected output to indicate no config file, got: %s", output)
	}

	// Check that it shows default settings
	if !strings.Contains(output, "monday") {
		t.Errorf("Expected output to show default week_start_day 'monday', got: %s", output)
	}
	if !strings.Contains(output, "Local") {
		t.Errorf("Expected output to show default timezone 'Local', got: %s", output)
	}
	if !strings.Contains(output, "(default)") {
		t.Errorf("Expected output to show '(default)' for output format, got: %s", output)
	}

	// Check that it shows the tip message
	if !strings.Contains(output, "Tip:") {
		t.Errorf("Expected output to contain tip message, got: %s", output)
	}
}

// TestShowConfig_ValidConfigFile tests showing config when valid config file exists
func TestShowConfig_ValidConfigFile(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	// Create a custom config file
	customConfig := `week_start_day = "sunday"
timezone = "America/New_York"
default_output_format = "custom"
`
	if err := os.WriteFile(configPath, []byte(customConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	showConfig()

	output := stdout.String()

	// Check that it indicates file exists
	if !strings.Contains(output, "File exists") {
		t.Errorf("Expected output to indicate config file exists, got: %s", output)
	}

	// Check that it shows custom settings
	if !strings.Contains(output, "sunday") {
		t.Errorf("Expected output to show custom week_start_day 'sunday', got: %s", output)
	}
	if !strings.Contains(output, "America/New_York") {
		t.Errorf("Expected output to show custom timezone 'America/New_York', got: %s", output)
	}
	if !strings.Contains(output, "custom") {
		t.Errorf("Expected output to show custom output format 'custom', got: %s", output)
	}

	// Should not show the tip message when config exists
	if strings.Contains(output, "Tip:") {
		t.Errorf("Expected no tip message when config file exists, got: %s", output)
	}
}

// TestShowConfig_InvalidConfigFile tests showing config with invalid config file
func TestShowConfig_InvalidConfigFile(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	// Create an invalid config file (invalid week_start_day)
	invalidConfig := `week_start_day = "tuesday"
timezone = "Local"
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	showConfig()

	if !exitCalled {
		t.Error("Expected exit to be called for invalid config")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Failed to load configuration") {
		t.Errorf("Expected stderr to contain load error, got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "monday") || !strings.Contains(stderrOutput, "sunday") {
		t.Errorf("Expected stderr to show valid week_start_day values, got: %s", stderrOutput)
	}
}

// TestInitConfig_NoExistingFile tests creating sample config when no file exists
func TestInitConfig_NoExistingFile(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
		// Remove the config file temporarily
		if err := os.Remove(configPath); err != nil {
			t.Fatalf("Failed to remove existing config for test: %v", err)
		}
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	initConfig()

	output := stdout.String()

	// Check that success message is shown
	if !strings.Contains(output, "Created sample configuration file") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Check that next steps are shown
	if !strings.Contains(output, "Next steps:") {
		t.Errorf("Expected next steps message, got: %s", output)
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}

	// Verify config file has sample content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "week_start_day") {
		t.Errorf("Expected config file to contain week_start_day documentation, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "timezone") {
		t.Errorf("Expected config file to contain timezone documentation, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "default_output_format") {
		t.Errorf("Expected config file to contain default_output_format documentation, got: %s", contentStr)
	}
}

// TestInitConfig_OverwriteConfirmYes tests overwriting existing config with confirmation
func TestInitConfig_OverwriteConfirmYes(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	// Create an existing config file
	existingConfig := `week_start_day = "sunday"`
	if err := os.WriteFile(configPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to create existing config file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader("y\n"),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	initConfig()

	output := stdout.String()

	// Check that it prompts for confirmation
	if !strings.Contains(output, "already exists") {
		t.Errorf("Expected prompt for overwrite confirmation, got: %s", output)
	}

	// Check that success message is shown
	if !strings.Contains(output, "Created sample configuration file") {
		t.Errorf("Expected success message after overwrite, got: %s", output)
	}

	// Verify config file was overwritten with sample content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	// Check that the old config (which was just: week_start_day = "sunday") was replaced
	// The file should start with "# did configuration file" if it's the sample config
	if !strings.HasPrefix(contentStr, "# did configuration file") {
		t.Errorf("Expected sample config (should start with '# did configuration file'), got: %s", contentStr)
	}
	// Ensure the actual config line (not comment) is commented out in sample
	if !strings.Contains(contentStr, "# week_start_day = \"monday\"") {
		t.Errorf("Expected sample config with commented-out week_start_day, got: %s", contentStr)
	}
}

// TestInitConfig_OverwriteConfirmNo tests cancelling overwrite when user says no
func TestInitConfig_OverwriteConfirmNo(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	// Create an existing config file
	existingConfig := `week_start_day = "sunday"`
	if err := os.WriteFile(configPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to create existing config file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader("n\n"),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	initConfig()

	output := stdout.String()

	// Check that it prompts for confirmation
	if !strings.Contains(output, "already exists") {
		t.Errorf("Expected prompt for overwrite confirmation, got: %s", output)
	}

	// Check that cancellation message is shown
	if !strings.Contains(output, "Cancelled") {
		t.Errorf("Expected cancellation message, got: %s", output)
	}

	// Verify config file was NOT overwritten
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	if contentStr != existingConfig {
		t.Errorf("Expected original config to be preserved, got: %s", contentStr)
	}
}

// TestPromptOverwriteConfirmation tests the overwrite confirmation prompt with various inputs
func TestPromptOverwriteConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"lowercase y", "y\n", true},
		{"uppercase Y", "Y\n", true},
		{"lowercase n", "n\n", false},
		{"uppercase N", "N\n", false},
		{"empty input", "\n", false},
		{"random text", "maybe\n", false},
		{"y with spaces", "  y  \n", true},
		{"Y with spaces", "  Y  \n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deps{
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
				Stdin:  strings.NewReader(tt.input),
				Exit:   func(code int) {},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			result := promptOverwriteConfirmation()
			if result != tt.expected {
				t.Errorf("promptOverwriteConfirmation() with input %q = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPromptOverwriteConfirmation_ScannerFail tests prompt with scanner EOF
func TestPromptOverwriteConfirmation_ScannerFail(t *testing.T) {
	d := &Deps{
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
		Stdin:       &eofReaderConfig{},
		Exit:        func(code int) {},
		StoragePath: func() (string, error) { return "", nil },
	}
	SetDeps(d)
	defer ResetDeps()

	result := promptOverwriteConfirmation()
	if result != false {
		t.Error("Expected false when scanner fails to read")
	}
}

// eofReaderConfig is an io.Reader that immediately returns EOF (for testing scanner failure)
type eofReaderConfig struct{}

func (e eofReaderConfig) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// TestConfigCmd_Display tests the config command without flags (display mode)
func TestConfigCmd_Display(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
		// Remove the config file temporarily
		if err := os.Remove(configPath); err != nil {
			t.Fatalf("Failed to remove existing config for test: %v", err)
		}
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Reset the init flag to false (default)
	configInitFlag = false

	// Call the config command's Run function
	configCmd.Run(configCmd, []string{})

	output := stdout.String()

	// Verify it shows config display output
	if !strings.Contains(output, "Configuration for did") {
		t.Errorf("Expected config display output, got: %s", output)
	}
	if !strings.Contains(output, "No config file") {
		t.Errorf("Expected 'No config file' message, got: %s", output)
	}
}

// TestConfigCmd_Init tests the config command with --init flag
func TestConfigCmd_Init(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
		// Remove the config file temporarily
		if err := os.Remove(configPath); err != nil {
			t.Fatalf("Failed to remove existing config for test: %v", err)
		}
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Set the init flag to true
	configInitFlag = true
	defer func() { configInitFlag = false }()

	// Call the config command's Run function
	configCmd.Run(configCmd, []string{})

	output := stdout.String()

	// Verify it shows init output
	if !strings.Contains(output, "Created sample configuration file") {
		t.Errorf("Expected init success message, got: %s", output)
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created by --init flag")
	}
}

// TestShowConfig_GetConfigPathError tests showing config when GetConfigPath fails
func TestShowConfig_GetConfigPathError(t *testing.T) {
	// Note: This is hard to test because GetConfigPath uses os.UserConfigDir which
	// only fails when HOME is not set properly. We'll test what we can.
	// The error path in showConfig() is from line 58-64.
	// To test this, we'd need to mock config.GetConfigPath, but it's a package function.
	// This test documents the expected behavior.
	t.Skip("GetConfigPath error path requires mocking package-level function")
}

// TestInitConfig_GetConfigPathError tests initConfig when GetConfigPath fails
func TestInitConfig_GetConfigPathError(t *testing.T) {
	// Similar to above - requires mocking config.GetConfigPath
	t.Skip("GetConfigPath error path requires mocking package-level function")
}

// TestInitConfig_WriteFileError tests initConfig when WriteFile fails
func TestInitConfig_WriteFileError(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
		// Remove the config file temporarily
		if err := os.Remove(configPath); err != nil {
			t.Fatalf("Failed to remove existing config for test: %v", err)
		}
	}

	// Restore config after test
	defer func() {
		// Restore parent directory permissions
		parentDir := configPath[:len(configPath)-len("/config.toml")]
		_ = os.Chmod(parentDir, 0755)
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	// Get parent directory and make it read-only to cause WriteFile to fail
	parentDir := configPath[:len(configPath)-len("/config.toml")]

	// Create the parent directory if it doesn't exist
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatalf("Failed to create parent dir: %v", err)
	}

	// Make parent directory read-only
	if err := os.Chmod(parentDir, 0555); err != nil {
		t.Fatalf("Failed to change parent dir permissions: %v", err)
	}

	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	initConfig()

	if !exitCalled {
		t.Error("Expected exit to be called when WriteFile fails")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Failed to create config file") {
		t.Errorf("Expected stderr to contain write error, got: %s", stderrOutput)
	}
}

// TestShowConfig_PartialConfig tests showing config with partial config file (merged with defaults)
func TestShowConfig_PartialConfig(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadExistingConfig bool
	if content, err := os.ReadFile(configPath); err == nil {
		backupContent = content
		hadExistingConfig = true
	}

	// Restore config after test
	defer func() {
		if hadExistingConfig {
			_ = os.WriteFile(configPath, backupContent, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}()

	// Create a partial config file (only week_start_day)
	partialConfig := `week_start_day = "sunday"
`
	if err := os.WriteFile(configPath, []byte(partialConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	showConfig()

	output := stdout.String()

	// Check that it shows custom week_start_day
	if !strings.Contains(output, "sunday") {
		t.Errorf("Expected output to show custom week_start_day 'sunday', got: %s", output)
	}

	// Check that it shows default timezone (merged from defaults)
	if !strings.Contains(output, "Local") {
		t.Errorf("Expected output to show default timezone 'Local' (merged), got: %s", output)
	}

	// Check that it shows default output format
	if !strings.Contains(output, "(default)") {
		t.Errorf("Expected output to show '(default)' for output format, got: %s", output)
	}
}
