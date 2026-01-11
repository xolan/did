package handlers

import (
	"fmt"
	"strings"

	"github.com/xolan/did/internal/cli"
)

// ShowConfig displays the current configuration
func ShowConfig(deps *cli.Deps) {
	cfg := deps.Services.Config.Get()
	path := deps.Services.Config.GetPath()

	_, _ = fmt.Fprintln(deps.Stdout, "Configuration:")
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("=", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "Config file: %s\n", path)
	if deps.Services.Config.Exists() {
		_, _ = fmt.Fprintln(deps.Stdout, "Status: File exists")
	} else {
		_, _ = fmt.Fprintln(deps.Stdout, "Status: Using defaults (no config file)")
	}
	_, _ = fmt.Fprintln(deps.Stdout, strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(deps.Stdout, "week_start_day: %s\n", cfg.WeekStartDay)
	_, _ = fmt.Fprintf(deps.Stdout, "timezone:       %s\n", cfg.Timezone)
}

// InitConfig creates a sample config file
func InitConfig(deps *cli.Deps) {
	err := deps.Services.Config.Init()
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		deps.Exit(1)
		return
	}

	path := deps.Services.Config.GetPath()
	_, _ = fmt.Fprintf(deps.Stdout, "Created config file: %s\n", path)
	_, _ = fmt.Fprintln(deps.Stdout, "Edit this file to customize your settings.")
}
