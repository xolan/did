package cli

import (
	"io"
	"os"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timer"
)

// Deps contains all dependencies for CLI operations
type Deps struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
	Exit   func(code int)

	// Services
	Services *service.Services

	// Raw paths (for backward compatibility and direct access)
	StoragePath func() (string, error)
	TimerPath   func() (string, error)
	Config      config.Config
}

// DefaultDeps creates a new Deps with default values
func DefaultDeps() *Deps {
	cfg := config.DefaultConfig()
	configPath, err := config.GetConfigPath()
	if err == nil {
		if loadedCfg, err := config.LoadOrDefault(configPath); err == nil {
			cfg = loadedCfg
		}
	}

	services, _ := service.NewServices()

	return &Deps{
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin,
		Exit:        os.Exit,
		Services:    services,
		StoragePath: storage.GetStoragePath,
		TimerPath:   timer.GetTimerPath,
		Config:      cfg,
	}
}

// NewDeps creates a new Deps with the given services
func NewDeps(services *service.Services, cfg config.Config) *Deps {
	return &Deps{
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin,
		Exit:        os.Exit,
		Services:    services,
		StoragePath: storage.GetStoragePath,
		TimerPath:   timer.GetTimerPath,
		Config:      cfg,
	}
}

// Global deps instance for CLI
var deps = DefaultDeps()

// SetDeps sets the global deps (for testing)
func SetDeps(d *Deps) {
	deps = d
}

// ResetDeps resets to default deps
func ResetDeps() {
	deps = DefaultDeps()
}

// GetDeps returns the current deps
func GetDeps() *Deps {
	return deps
}
