package osutil

import (
	"errors"
	"os"
	"testing"
)

// MockPathProvider is a mock implementation for testing.
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

func TestDefaultPathProvider_UserConfigDir(t *testing.T) {
	p := DefaultPathProvider{}
	dir, err := p.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir returned error: %v", err)
	}
	if dir == "" {
		t.Error("UserConfigDir returned empty string")
	}
}

func TestDefaultPathProvider_MkdirAll(t *testing.T) {
	p := DefaultPathProvider{}
	tmpDir := t.TempDir()
	testDir := tmpDir + "/test/nested/dir"

	err := p.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Failed to stat created directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("MkdirAll did not create a directory")
	}
}

func TestSetProvider(t *testing.T) {
	// Save original provider
	original := Provider
	defer func() { Provider = original }()

	mock := &MockPathProvider{
		UserConfigDirFn: func() (string, error) {
			return "/mock/config", nil
		},
	}

	SetProvider(mock)

	if Provider != mock {
		t.Error("SetProvider did not set the provider")
	}

	dir, _ := Provider.UserConfigDir()
	if dir != "/mock/config" {
		t.Errorf("Expected /mock/config, got %s", dir)
	}
}

func TestResetProvider(t *testing.T) {
	// Save original provider
	original := Provider
	defer func() { Provider = original }()

	mock := &MockPathProvider{}
	SetProvider(mock)

	ResetProvider()

	_, ok := Provider.(DefaultPathProvider)
	if !ok {
		t.Error("ResetProvider did not reset to DefaultPathProvider")
	}
}

func TestMockPathProvider_Error(t *testing.T) {
	expectedErr := errors.New("mock error")
	mock := &MockPathProvider{
		UserConfigDirFn: func() (string, error) {
			return "", expectedErr
		},
		MkdirAllFn: func(path string, perm os.FileMode) error {
			return expectedErr
		},
	}

	_, err := mock.UserConfigDir()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	err = mock.MkdirAll("/test", 0755)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}
