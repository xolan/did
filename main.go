package main

import (
	"os"

	"github.com/xolan/did/cmd"
)

// Version information injected by GoReleaser via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Validate configuration before executing commands
	// This ensures invalid config files show helpful error messages
	if !cmd.ValidateConfigOnStartup() {
		os.Exit(1)
	}

	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
