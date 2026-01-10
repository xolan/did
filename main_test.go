package main

import (
	"errors"
	"os"
	"testing"

	"github.com/xolan/did/internal/osutil"
)

// MockPathProvider for testing config validation failure
type MockPathProvider struct {
	UserConfigDirFn func() (string, error)
	MkdirAllFn      func(path string, perm os.FileMode) error
}

func (m *MockPathProvider) UserConfigDir() (string, error) {
	if m.UserConfigDirFn != nil {
		return m.UserConfigDirFn()
	}
	return "", nil
}

func (m *MockPathProvider) MkdirAll(path string, perm os.FileMode) error {
	if m.MkdirAllFn != nil {
		return m.MkdirAllFn(path, perm)
	}
	return nil
}

func TestRun_Success(t *testing.T) {
	// run() with valid config should return 0
	// The default config is always valid, so this should succeed
	// Note: This may show output, but that's expected
	code := run()
	if code != 0 {
		t.Errorf("Expected exit code 0, got %d", code)
	}
}

func TestRun_ConfigValidationFailure(t *testing.T) {
	// Save original provider
	original := osutil.Provider
	defer osutil.ResetProvider()

	// Mock to simulate config path error
	osutil.SetProvider(&MockPathProvider{
		UserConfigDirFn: func() (string, error) {
			return "", errors.New("permission denied")
		},
	})

	code := run()
	if code != 1 {
		t.Errorf("Expected exit code 1 for config validation failure, got %d", code)
	}

	// Reset for next test
	osutil.Provider = original
}

func TestRun_ExecuteError(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"did", "--unknownflag"}

	code := run()
	if code != 1 {
		t.Errorf("Expected exit code 1 for Execute error, got %d", code)
	}
}

func TestMain_CallsExitWithRunResult(t *testing.T) {
	originalExit := exitFunc
	originalArgs := os.Args
	defer func() {
		exitFunc = originalExit
		os.Args = originalArgs
	}()

	var capturedCode int
	exitFunc = func(code int) {
		capturedCode = code
	}
	os.Args = []string{"did"}

	main()

	if capturedCode != 0 {
		t.Errorf("Expected exit code 0, got %d", capturedCode)
	}
}
