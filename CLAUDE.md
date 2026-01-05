# CLAUDE.md

## Project Overview

**did** - A Go CLI tool for time tracking. Log work activities with durations and view entries by time period.

## Tech Stack

- Go 1.25+
- Cobra CLI framework

## Project Structure

```
main.go                           # Entry point
cmd/
  root.go                         # Cobra command definitions (did, y, w, lw, edit, validate)
  delete.go                       # Delete command (soft delete)
  undo.go                         # Undo last delete command
  purge.go                        # Purge all soft-deleted entries command
  restore.go                      # Restore from backup command
  stats.go                        # Stats command (weekly/monthly statistics)
  config.go                       # Config command (display and init configuration)
  search.go                       # Search command (keyword search with date filters)
  export.go                       # Export command (JSON and CSV export)
  report.go                       # Report command (project/tag reports and grouping)
  completion.go                   # Shell completion generation
internal/
  config/
    config.go                     # Configuration (Config struct, TOML loading, validation)
    config_test.go
  entry/
    entry.go                      # Entry struct (Timestamp, Description, DurationMinutes, RawInput, DeletedAt)
    parser.go                     # Duration parsing (ParseDuration)
    parser_test.go
  storage/
    jsonl.go                      # JSONL storage (AppendEntry, ReadEntries, GetStoragePath, soft delete functions)
    jsonl_test.go
    backup.go                     # Backup management (CreateBackup, ListBackups, RestoreBackup)
    backup_test.go
  stats/
    stats.go                      # Statistics calculations (CalculateStatistics, project/tag breakdowns, comparisons)
    stats_test.go
  timeutil/
    datefilter.go                 # Date range utilities (Today, Yesterday, ThisWeek, LastWeek, ThisMonth, LastMonth)
    datefilter_test.go
```

## Commands

```bash
just setup         # Install mise tools and download Go dependencies
just test          # Run test suite: go test ./...
just format        # Format code: go fmt ./...
just lint          # Run linter: golangci-lint
just build         # Build binary to dist/did
just install       # Build and install to ~/.local/bin/
just release       # Build release artifacts with GoReleaser (snapshot)
just release-check # Validate GoReleaser configuration
```

## CLI Usage

```bash
# Log entries
did <description> for <duration>  # Log entry (e.g., "did feature X for 2h")
did fix bug @acme for 1h          # Log with project
did review #code #urgent for 30m  # Log with tags

# View entries
did                               # List today's entries
did y                             # List yesterday's entries
did w                             # List this week's entries
did lw                            # List last week's entries
did 2024-01-15                    # List entries for specific date
did from 2024-01-01 to 2024-01-31 # List entries for date range
did last 7 days                   # List entries from last 7 days

# Filter by project/tag
did @acme                         # Today's entries for project 'acme'
did w #bugfix                     # This week's entries tagged 'bugfix'
did --project acme --tag review   # Multiple filters

# Edit entries
did edit <index> --description X  # Edit entry description
did edit <index> --duration 2h    # Edit entry duration

# Delete and restore
did delete <index>                # Soft delete (can be undone, auto-purged after 7 days)
did delete <index> -y             # Delete without confirmation
did undo                          # Restore the most recently deleted entry
did purge                         # Permanently remove all soft-deleted entries
did purge -y                      # Purge without confirmation

# Search
did search <keyword>              # Search entries by keyword
did search bug --from 2024-01-01  # Search from specific date
did search api --last 7           # Search last 7 days

# Export
did export json                   # Export all entries as JSON
did export json --last 7          # Export last 7 days
did export json @acme #review     # Export with filters
did export csv                    # Export all entries as CSV
did export csv > backup.csv       # Export to file

# Reports
did report @project               # Show all entries for a specific project
did report #tag                   # Show all entries with a specific tag
did report --by project           # Show hours grouped by all projects
did report --by tag               # Show hours grouped by all tags
did report @acme --last 7         # Project report for last 7 days

# Statistics
did stats                         # Statistics for current week
did stats --month                 # Statistics for current month

# Maintenance
did validate                      # Check storage file health
did restore                       # Restore from most recent backup
did restore <n>                   # Restore from backup #n (1-3)
did config                        # Display current configuration
did config --init                 # Create sample config file
did completion [bash|zsh|fish|powershell]  # Generate shell completions
```

Duration format: `Yh` (hours), `Ym` (minutes), or `YhYm` (combined). Max 24 hours per entry.

## Data Storage

Entries stored in JSONL format at `~/.config/did/entries.jsonl` (Linux), uses `os.UserConfigDir()` for cross-platform support.

**Soft Delete Behavior:**
- Deleted entries are marked with a `deleted_at` timestamp rather than removed
- Deleted entries can be recovered with `did undo`
- Entries deleted more than 7 days ago are automatically purged during delete operations
- Use `did purge` to manually remove all soft-deleted entries immediately

## Configuration

Configuration file is **optional** - did works perfectly without any configuration. All settings have sensible defaults.

**Config file location:**
- Linux/macOS: `~/.config/did/config.toml`
- Windows: `%APPDATA%\did\config.toml`

**Create a config file:**
```bash
did config --init  # Creates a sample config.toml with all options documented
```

**View current configuration:**
```bash
did config  # Shows current settings and config file location
```

### Configuration Options

**`week_start_day`** - Which day starts the week (affects `did w`, `did lw`, and `did stats`)
- Valid values: `"monday"` or `"sunday"`
- Default: `"monday"` (ISO 8601 standard)
- Example: `week_start_day = "sunday"` for US convention

**`timezone`** - Timezone for time operations and display
- Valid values: Any IANA timezone name (e.g., `"America/New_York"`, `"Europe/London"`, `"Asia/Tokyo"`) or `"Local"` for system timezone
- Default: `"Local"` (uses system timezone)
- Example: `timezone = "America/New_York"`
- See available timezones: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

**`default_output_format`** - Default output format for entry listings (reserved for future use)
- Valid values: Currently unused (reserved for future custom formats)
- Default: `""` (uses built-in default format)

### Sample Configuration

```toml
# Uncomment and modify only the settings you want to customize

# Week starts on Sunday (US convention)
week_start_day = "sunday"

# Use Eastern Time
timezone = "America/New_York"

# Default output format (reserved for future use)
# default_output_format = ""
```

## Conventions

- Tests alongside source files (`*_test.go`)
- Internal packages under `internal/`
- Errors written to stderr, success to stdout
- Week start day is configurable (defaults to Monday per ISO 8601)


## Development help

- If running in a sandboxed environment with strict permissions on what you can run, execute commands through `bash -c`.
- Always keep documentation up to date. This includes _all_ markdown files, and code docstrings.