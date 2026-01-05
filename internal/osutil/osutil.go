// Package osutil provides abstractions for OS-level operations to enable testing.
package osutil

import "os"

// PathProvider abstracts OS-level operations for path resolution.
// Used to enable testing of error paths in GetStoragePath, GetTimerPath, and GetConfigPath.
type PathProvider interface {
	UserConfigDir() (string, error)
	MkdirAll(path string, perm os.FileMode) error
}

// DefaultPathProvider uses real OS functions.
type DefaultPathProvider struct{}

// UserConfigDir returns the default root directory for user-specific configuration data.
func (DefaultPathProvider) UserConfigDir() (string, error) {
	return os.UserConfigDir()
}

// MkdirAll creates a directory named path, along with any necessary parents.
func (DefaultPathProvider) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Provider is the package-level path provider instance.
// In production, this is DefaultPathProvider. Tests can replace it.
var Provider PathProvider = DefaultPathProvider{}

// SetProvider sets a custom provider (for testing).
func SetProvider(p PathProvider) {
	Provider = p
}

// ResetProvider resets to the default provider.
func ResetProvider() {
	Provider = DefaultPathProvider{}
}
