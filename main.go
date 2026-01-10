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

var exitFunc = os.Exit

func main() {
	exitFunc(run())
}

// run executes the main application logic and returns an exit code.
// This function is separated from main() to enable testing.
func run() int {
	// Validate configuration before executing commands
	// This ensures invalid config files show helpful error messages
	if !cmd.ValidateConfigOnStartup() {
		return 1
	}

	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}
