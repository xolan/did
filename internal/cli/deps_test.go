package cli

import (
	"testing"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/service"
)

func TestNewDeps(t *testing.T) {
	cfg := config.DefaultConfig()
	services, err := service.NewServices()
	if err != nil {
		t.Fatalf("failed to create services: %v", err)
	}

	deps := NewDeps(services, cfg)
	if deps == nil {
		t.Fatal("expected non-nil deps")
	}
	if deps.Services != services {
		t.Error("expected services to match")
	}
	if deps.Stdout == nil {
		t.Error("expected non-nil Stdout")
	}
	if deps.Stderr == nil {
		t.Error("expected non-nil Stderr")
	}
	if deps.Stdin == nil {
		t.Error("expected non-nil Stdin")
	}
	if deps.Exit == nil {
		t.Error("expected non-nil Exit")
	}
	if deps.StoragePath == nil {
		t.Error("expected non-nil StoragePath")
	}
	if deps.TimerPath == nil {
		t.Error("expected non-nil TimerPath")
	}
}

func TestSetDeps(t *testing.T) {
	// Save original deps
	original := GetDeps()
	defer SetDeps(original)

	// Create new deps
	cfg := config.DefaultConfig()
	services, err := service.NewServices()
	if err != nil {
		t.Fatalf("failed to create services: %v", err)
	}
	newDeps := NewDeps(services, cfg)

	// Set new deps
	SetDeps(newDeps)

	// Verify it was set
	if GetDeps() != newDeps {
		t.Error("expected GetDeps to return the set deps")
	}
}

func TestResetDeps(t *testing.T) {
	// Save original
	original := GetDeps()

	// Set custom deps
	cfg := config.DefaultConfig()
	services, err := service.NewServices()
	if err != nil {
		t.Fatalf("failed to create services: %v", err)
	}
	customDeps := NewDeps(services, cfg)
	SetDeps(customDeps)

	// Reset
	ResetDeps()

	// Should be different from custom deps (new instance)
	current := GetDeps()
	if current == customDeps {
		t.Error("expected ResetDeps to create new deps")
	}
	if current.Stdout == nil {
		t.Error("expected reset deps to have non-nil Stdout")
	}

	// Restore original for other tests
	SetDeps(original)
}

func TestGetDeps(t *testing.T) {
	deps := GetDeps()
	if deps == nil {
		t.Fatal("expected non-nil deps from GetDeps")
	}
	if deps.Stdout == nil {
		t.Error("expected non-nil Stdout")
	}
}
