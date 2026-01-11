package handlers

import (
	"strings"
	"testing"
)

func TestShowConfig(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ShowConfig(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Configuration:") {
		t.Errorf("expected 'Configuration:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Config file:") {
		t.Errorf("expected 'Config file:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "week_start_day:") {
		t.Errorf("expected 'week_start_day:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "timezone:") {
		t.Errorf("expected 'timezone:' in output, got %q", stdout.String())
	}
}

func TestShowConfig_NoFile(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	ShowConfig(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Using defaults") {
		t.Errorf("expected 'Using defaults' in output, got %q", stdout.String())
	}
}

func TestShowConfig_WithFile(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	// Initialize config file first
	InitConfig(deps)
	stdout.Reset()

	ShowConfig(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "File exists") {
		t.Errorf("expected 'File exists' in output, got %q", stdout.String())
	}
}

func TestInitConfig(t *testing.T) {
	deps, stdout, _, exitCode := setupTestDeps(t)

	InitConfig(deps)

	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	if !strings.Contains(stdout.String(), "Created config file:") {
		t.Errorf("expected 'Created config file:' in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Edit this file") {
		t.Errorf("expected 'Edit this file' in output, got %q", stdout.String())
	}
}

func TestInitConfig_Error(t *testing.T) {
	deps, _, stderr, exitCode := setupBrokenConfigDeps(t)

	InitConfig(deps)

	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected 'Error:' in stderr, got %q", stderr.String())
	}
}
