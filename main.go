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
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
