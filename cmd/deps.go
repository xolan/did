package cmd

import (
	"io"
	"os"

	"github.com/xolan/did/internal/storage"
)

// Deps holds external dependencies for CLI commands, enabling testability.
type Deps struct {
	Stdout      io.Writer
	Stderr      io.Writer
	Stdin       io.Reader
	Exit        func(code int)
	StoragePath func() (string, error)
}

// DefaultDeps returns the default production dependencies.
func DefaultDeps() *Deps {
	return &Deps{
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin,
		Exit:        os.Exit,
		StoragePath: storage.GetStoragePath,
	}
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
