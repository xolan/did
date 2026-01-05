package timer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xolan/did/internal/osutil"
)

// Helper to create a temporary timer file path
func createTempTimerPath(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, TimerFile)
}

func TestSaveTimerState(t *testing.T) {
	tests := []struct {
		name  string
		state TimerState
	}{
		{
			name: "basic timer state",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 10, 30, 0, 0, time.Local),
				Description: "working on feature X",
			},
		},
		{
			name: "timer with project",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 14, 0, 0, 0, time.Local),
				Description: "code review",
				Project:     "acme",
			},
		},
		{
			name: "timer with tags",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 16, 0, 0, 0, time.Local),
				Description: "bug fix",
				Tags:        []string{"bugfix", "urgent"},
			},
		},
		{
			name: "timer with project and tags",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
				Description: "feature implementation",
				Project:     "client-project",
				Tags:        []string{"feature", "backend"},
			},
		},
		{
			name: "timer with special characters in description",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 11, 0, 0, 0, time.Local),
				Description: "fix bug #123 \"critical\" issue",
			},
		},
		{
			name: "timer with empty tags slice",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 12, 0, 0, 0, time.Local),
				Description: "meeting",
				Tags:        []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempTimerPath(t)

			err := SaveTimerState(tmpFile, tt.state)
			if err != nil {
				t.Fatalf("SaveTimerState() returned unexpected error: %v", err)
			}

			// Verify file exists
			if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
				t.Fatalf("SaveTimerState() did not create file")
			}

			// Verify file is valid JSON
			data, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read saved file: %v", err)
			}

			var loaded TimerState
			if err := json.Unmarshal(data, &loaded); err != nil {
				t.Fatalf("Saved file contains invalid JSON: %v", err)
			}

			// Verify saved data matches input
			if loaded.Description != tt.state.Description {
				t.Errorf("Saved description = %q, expected %q", loaded.Description, tt.state.Description)
			}
			if loaded.Project != tt.state.Project {
				t.Errorf("Saved project = %q, expected %q", loaded.Project, tt.state.Project)
			}
			if !loaded.StartedAt.Equal(tt.state.StartedAt) {
				t.Errorf("Saved StartedAt = %v, expected %v", loaded.StartedAt, tt.state.StartedAt)
			}
			if len(loaded.Tags) != len(tt.state.Tags) {
				t.Errorf("Saved tags length = %d, expected %d", len(loaded.Tags), len(tt.state.Tags))
			}
			for i, tag := range tt.state.Tags {
				if i < len(loaded.Tags) && loaded.Tags[i] != tag {
					t.Errorf("Saved tags[%d] = %q, expected %q", i, loaded.Tags[i], tag)
				}
			}
		})
	}
}

func TestSaveTimerState_CreatesFile(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Verify file doesn't exist
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Fatalf("Test setup error: file should not exist")
	}

	testState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "test timer",
	}

	err := SaveTimerState(tmpFile, testState)
	if err != nil {
		t.Fatalf("SaveTimerState() returned unexpected error: %v", err)
	}

	// Verify file now exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Errorf("SaveTimerState() did not create file")
	}
}

func TestSaveTimerState_Overwrites(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Save first state
	firstState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
		Description: "first timer",
		Project:     "old-project",
	}
	if err := SaveTimerState(tmpFile, firstState); err != nil {
		t.Fatalf("SaveTimerState() first call failed: %v", err)
	}

	// Save second state (should overwrite)
	secondState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "second timer",
		Project:     "new-project",
	}
	if err := SaveTimerState(tmpFile, secondState); err != nil {
		t.Fatalf("SaveTimerState() second call failed: %v", err)
	}

	// Load and verify second state is present
	loaded, err := LoadTimerState(tmpFile)
	if err != nil {
		t.Fatalf("LoadTimerState() returned unexpected error: %v", err)
	}

	if loaded.Description != secondState.Description {
		t.Errorf("Loaded description = %q, expected %q", loaded.Description, secondState.Description)
	}
	if loaded.Project != secondState.Project {
		t.Errorf("Loaded project = %q, expected %q", loaded.Project, secondState.Project)
	}
}

func TestSaveTimerState_FilePermissions(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	testState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "test",
	}

	err := SaveTimerState(tmpFile, testState)
	if err != nil {
		t.Fatalf("SaveTimerState() returned unexpected error: %v", err)
	}

	// Check file permissions (should be 0644)
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	expectedPerm := os.FileMode(0644)
	actualPerm := info.Mode().Perm()
	if actualPerm != expectedPerm {
		t.Errorf("File permissions = %o, expected %o", actualPerm, expectedPerm)
	}
}

func TestLoadTimerState(t *testing.T) {
	tests := []struct {
		name          string
		state         TimerState
		expectedDesc  string
		expectedProj  string
		expectedTags  []string
		expectedEmpty bool
	}{
		{
			name: "basic timer",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 10, 30, 0, 0, time.Local),
				Description: "working on feature",
			},
			expectedDesc: "working on feature",
			expectedProj: "",
			expectedTags: nil,
		},
		{
			name: "timer with project",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 14, 0, 0, 0, time.Local),
				Description: "code review",
				Project:     "acme",
			},
			expectedDesc: "code review",
			expectedProj: "acme",
			expectedTags: nil,
		},
		{
			name: "timer with tags",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 16, 0, 0, 0, time.Local),
				Description: "bug fix",
				Tags:        []string{"bugfix", "urgent"},
			},
			expectedDesc: "bug fix",
			expectedProj: "",
			expectedTags: []string{"bugfix", "urgent"},
		},
		{
			name: "timer with project and tags",
			state: TimerState{
				StartedAt:   time.Date(2024, time.January, 15, 9, 0, 0, 0, time.Local),
				Description: "feature implementation",
				Project:     "client",
				Tags:        []string{"feature", "backend"},
			},
			expectedDesc: "feature implementation",
			expectedProj: "client",
			expectedTags: []string{"feature", "backend"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempTimerPath(t)

			// Save state first
			if err := SaveTimerState(tmpFile, tt.state); err != nil {
				t.Fatalf("SaveTimerState() failed: %v", err)
			}

			// Load state
			loaded, err := LoadTimerState(tmpFile)
			if err != nil {
				t.Fatalf("LoadTimerState() returned unexpected error: %v", err)
			}

			if loaded == nil {
				t.Fatalf("LoadTimerState() returned nil, expected state")
			}

			// Verify loaded data
			if loaded.Description != tt.expectedDesc {
				t.Errorf("Loaded description = %q, expected %q", loaded.Description, tt.expectedDesc)
			}
			if loaded.Project != tt.expectedProj {
				t.Errorf("Loaded project = %q, expected %q", loaded.Project, tt.expectedProj)
			}
			if !loaded.StartedAt.Equal(tt.state.StartedAt) {
				t.Errorf("Loaded StartedAt = %v, expected %v", loaded.StartedAt, tt.state.StartedAt)
			}

			// Verify tags
			if len(loaded.Tags) != len(tt.expectedTags) {
				t.Errorf("Loaded tags length = %d, expected %d", len(loaded.Tags), len(tt.expectedTags))
			}
			for i, tag := range tt.expectedTags {
				if i < len(loaded.Tags) && loaded.Tags[i] != tag {
					t.Errorf("Loaded tags[%d] = %q, expected %q", i, loaded.Tags[i], tag)
				}
			}
		})
	}
}

func TestLoadTimerState_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.json")

	state, err := LoadTimerState(nonExistentFile)
	if err != nil {
		t.Errorf("LoadTimerState() returned unexpected error for non-existent file: %v", err)
	}
	if state != nil {
		t.Errorf("LoadTimerState() returned %v, expected nil", state)
	}
}

func TestLoadTimerState_EmptyFile(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Create empty file
	if err := os.WriteFile(tmpFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	_, err := LoadTimerState(tmpFile)
	if err == nil {
		t.Error("LoadTimerState() should return error for empty file")
	}
}

func TestLoadTimerState_MalformedJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"invalid json", "not json at all"},
		{"truncated json", `{"started_at":"2024-01-15T10:00:00Z","descrip`},
		{"wrong type", `{"started_at":"not a timestamp","description":"test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempTimerPath(t)

			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			_, err := LoadTimerState(tmpFile)
			if err == nil {
				t.Errorf("LoadTimerState() should return error for malformed JSON")
			}
		})
	}
}

func TestLoadTimerState_PreservesTimestamp(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Use a specific timestamp
	testTime := time.Date(2024, time.June, 15, 9, 30, 45, 123456789, time.UTC)
	testState := TimerState{
		StartedAt:   testTime,
		Description: "test timer",
	}

	if err := SaveTimerState(tmpFile, testState); err != nil {
		t.Fatalf("SaveTimerState() failed: %v", err)
	}

	loaded, err := LoadTimerState(tmpFile)
	if err != nil {
		t.Fatalf("LoadTimerState() returned unexpected error: %v", err)
	}

	// Time should be equal (allowing for JSON serialization)
	if !loaded.StartedAt.Equal(testTime) {
		t.Errorf("Loaded StartedAt = %v, expected %v", loaded.StartedAt, testTime)
	}
}

func TestLoadTimerState_UnicodeContent(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	testState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "工作 on feature 🎉",
		Project:     "项目",
		Tags:        []string{"标签", "emoji🚀"},
	}

	if err := SaveTimerState(tmpFile, testState); err != nil {
		t.Fatalf("SaveTimerState() failed: %v", err)
	}

	loaded, err := LoadTimerState(tmpFile)
	if err != nil {
		t.Fatalf("LoadTimerState() returned unexpected error: %v", err)
	}

	if loaded.Description != testState.Description {
		t.Errorf("Loaded description = %q, expected %q", loaded.Description, testState.Description)
	}
	if loaded.Project != testState.Project {
		t.Errorf("Loaded project = %q, expected %q", loaded.Project, testState.Project)
	}
	if len(loaded.Tags) != len(testState.Tags) {
		t.Fatalf("Loaded tags length = %d, expected %d", len(loaded.Tags), len(testState.Tags))
	}
	for i, tag := range testState.Tags {
		if loaded.Tags[i] != tag {
			t.Errorf("Loaded tags[%d] = %q, expected %q", i, loaded.Tags[i], tag)
		}
	}
}

func TestClearTimerState(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Create a timer file
	testState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "test timer",
	}
	if err := SaveTimerState(tmpFile, testState); err != nil {
		t.Fatalf("SaveTimerState() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatalf("Test setup error: file should exist")
	}

	// Clear timer state
	err := ClearTimerState(tmpFile)
	if err != nil {
		t.Fatalf("ClearTimerState() returned unexpected error: %v", err)
	}

	// Verify file is removed
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Errorf("ClearTimerState() did not remove file")
	}
}

func TestClearTimerState_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.json")

	// Should not return error for non-existent file (idempotent)
	err := ClearTimerState(nonExistentFile)
	if err != nil {
		t.Errorf("ClearTimerState() returned unexpected error for non-existent file: %v", err)
	}
}

func TestClearTimerState_Idempotent(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Create and clear timer
	testState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "test timer",
	}
	if err := SaveTimerState(tmpFile, testState); err != nil {
		t.Fatalf("SaveTimerState() failed: %v", err)
	}

	// Clear first time
	if err := ClearTimerState(tmpFile); err != nil {
		t.Fatalf("ClearTimerState() first call failed: %v", err)
	}

	// Clear second time (should still succeed)
	if err := ClearTimerState(tmpFile); err != nil {
		t.Errorf("ClearTimerState() second call failed: %v", err)
	}
}

func TestIsTimerRunning_True(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Create a timer file
	testState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "test timer",
	}
	if err := SaveTimerState(tmpFile, testState); err != nil {
		t.Fatalf("SaveTimerState() failed: %v", err)
	}

	running, err := IsTimerRunning(tmpFile)
	if err != nil {
		t.Fatalf("IsTimerRunning() returned unexpected error: %v", err)
	}

	if !running {
		t.Errorf("IsTimerRunning() = false, expected true")
	}
}

func TestIsTimerRunning_False(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist.json")

	running, err := IsTimerRunning(nonExistentFile)
	if err != nil {
		t.Errorf("IsTimerRunning() returned unexpected error for non-existent file: %v", err)
	}

	if running {
		t.Errorf("IsTimerRunning() = true, expected false")
	}
}

func TestIsTimerRunning_MalformedFile(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Write malformed JSON
	if err := os.WriteFile(tmpFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := IsTimerRunning(tmpFile)
	if err == nil {
		t.Error("IsTimerRunning() should return error for malformed file")
	}
}

func TestIsTimerRunning_AfterClear(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Create timer
	testState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "test timer",
	}
	if err := SaveTimerState(tmpFile, testState); err != nil {
		t.Fatalf("SaveTimerState() failed: %v", err)
	}

	// Verify running
	running, err := IsTimerRunning(tmpFile)
	if err != nil {
		t.Fatalf("IsTimerRunning() returned unexpected error: %v", err)
	}
	if !running {
		t.Errorf("IsTimerRunning() = false before clear, expected true")
	}

	// Clear timer
	if err := ClearTimerState(tmpFile); err != nil {
		t.Fatalf("ClearTimerState() failed: %v", err)
	}

	// Verify not running
	running, err = IsTimerRunning(tmpFile)
	if err != nil {
		t.Fatalf("IsTimerRunning() returned unexpected error after clear: %v", err)
	}
	if running {
		t.Errorf("IsTimerRunning() = true after clear, expected false")
	}
}

func TestGetTimerPath(t *testing.T) {
	// Test that GetTimerPath returns a valid path
	path, err := GetTimerPath()
	if err != nil {
		t.Fatalf("GetTimerPath() returned unexpected error: %v", err)
	}

	// Path should not be empty
	if path == "" {
		t.Errorf("GetTimerPath() returned empty path")
	}

	// Path should end with timer.json
	if filepath.Base(path) != TimerFile {
		t.Errorf("GetTimerPath() path base = %q, expected %q", filepath.Base(path), TimerFile)
	}

	// Parent directory should exist (created by GetTimerPath)
	parentDir := filepath.Dir(path)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Errorf("GetTimerPath() parent directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("GetTimerPath() parent is not a directory")
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are set correctly
	if AppName != "did" {
		t.Errorf("AppName = %q, expected %q", AppName, "did")
	}

	if TimerFile != "timer.json" {
		t.Errorf("TimerFile = %q, expected %q", TimerFile, "timer.json")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	originalState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 30, 45, 0, time.UTC),
		Description: "complete feature implementation with tests and documentation",
		Project:     "my-awesome-project",
		Tags:        []string{"feature", "backend", "api", "urgent"},
	}

	// Save
	if err := SaveTimerState(tmpFile, originalState); err != nil {
		t.Fatalf("SaveTimerState() failed: %v", err)
	}

	// Load
	loadedState, err := LoadTimerState(tmpFile)
	if err != nil {
		t.Fatalf("LoadTimerState() failed: %v", err)
	}

	// Verify complete round-trip
	if loadedState == nil {
		t.Fatal("LoadTimerState() returned nil")
	}

	if !loadedState.StartedAt.Equal(originalState.StartedAt) {
		t.Errorf("Round-trip StartedAt = %v, expected %v", loadedState.StartedAt, originalState.StartedAt)
	}
	if loadedState.Description != originalState.Description {
		t.Errorf("Round-trip Description = %q, expected %q", loadedState.Description, originalState.Description)
	}
	if loadedState.Project != originalState.Project {
		t.Errorf("Round-trip Project = %q, expected %q", loadedState.Project, originalState.Project)
	}
	if len(loadedState.Tags) != len(originalState.Tags) {
		t.Fatalf("Round-trip Tags length = %d, expected %d", len(loadedState.Tags), len(originalState.Tags))
	}
	for i, tag := range originalState.Tags {
		if loadedState.Tags[i] != tag {
			t.Errorf("Round-trip Tags[%d] = %q, expected %q", i, loadedState.Tags[i], tag)
		}
	}
}

func TestLoadTimerState_PermissionError(t *testing.T) {
	tmpFile := createTempTimerPath(t)

	// Create a file
	testState := TimerState{
		StartedAt:   time.Date(2024, time.January, 15, 10, 0, 0, 0, time.Local),
		Description: "test",
	}
	if err := SaveTimerState(tmpFile, testState); err != nil {
		t.Fatalf("SaveTimerState() failed: %v", err)
	}

	// Make file unreadable
	if err := os.Chmod(tmpFile, 0000); err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}
	defer func() { _ = os.Chmod(tmpFile, 0644) }()

	_, err := LoadTimerState(tmpFile)
	if err == nil {
		t.Error("Expected error when reading unreadable file")
	}
}

func TestSaveTimerState_DirectoryError(t *testing.T) {
	// Use a path that can't be written to
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "subdir", "timer.json")

	// Don't create the subdir, so writing will fail
	testState := TimerState{
		StartedAt:   time.Now(),
		Description: "test",
	}

	err := SaveTimerState(invalidPath, testState)
	if err == nil {
		t.Error("Expected error when writing to non-existent directory")
	}
}

// mockPathProvider is a test helper for mocking osutil.PathProvider
type mockPathProvider struct {
	userConfigDirFn func() (string, error)
	mkdirAllFn      func(path string, perm os.FileMode) error
}

func (m *mockPathProvider) UserConfigDir() (string, error) {
	if m.userConfigDirFn != nil {
		return m.userConfigDirFn()
	}
	return "", nil
}

func (m *mockPathProvider) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirAllFn != nil {
		return m.mkdirAllFn(path, perm)
	}
	return nil
}

func TestGetTimerPath_UserConfigDirError(t *testing.T) {
	// Save original provider
	defer osutil.ResetProvider()

	// Mock UserConfigDir to return an error
	osutil.SetProvider(&mockPathProvider{
		userConfigDirFn: func() (string, error) {
			return "", os.ErrPermission
		},
	})

	_, err := GetTimerPath()
	if err == nil {
		t.Error("GetTimerPath() should return error when UserConfigDir fails")
	}
}

func TestGetTimerPath_MkdirAllError(t *testing.T) {
	// Save original provider
	defer osutil.ResetProvider()

	tmpDir := t.TempDir()

	// Mock MkdirAll to return an error
	osutil.SetProvider(&mockPathProvider{
		userConfigDirFn: func() (string, error) {
			return tmpDir, nil
		},
		mkdirAllFn: func(path string, perm os.FileMode) error {
			return os.ErrPermission
		},
	})

	_, err := GetTimerPath()
	if err == nil {
		t.Error("GetTimerPath() should return error when MkdirAll fails")
	}
}
