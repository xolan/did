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
internal/
  entry/
    entry.go                      # Entry struct (Timestamp, Description, DurationMinutes, RawInput, DeletedAt)
    parser.go                     # Duration parsing (ParseDuration)
    parser_test.go
  storage/
    jsonl.go                      # JSONL storage (AppendEntry, ReadEntries, GetStoragePath, soft delete functions)
    jsonl_test.go
    backup.go                     # Backup management (CreateBackup, ListBackups, RestoreBackup)
    backup_test.go
  timeutil/
    datefilter.go                 # Date range utilities (Today, Yesterday, ThisWeek, LastWeek)
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
did <description> for <duration>  # Log entry (e.g., "did feature X for 2h")
did                               # List today's entries
did y                             # List yesterday's entries
did w                             # List this week's entries
did lw                            # List last week's entries
did edit <index> --description X  # Edit entry description
did edit <index> --duration 2h    # Edit entry duration
did delete <index>                # Soft delete an entry (can be undone, auto-purged after 7 days)
did undo                          # Restore the most recently deleted entry
did purge                         # Permanently remove all soft-deleted entries (with confirmation)
did purge --yes                   # Permanently remove all soft-deleted entries (skip confirmation)
did validate                      # Check storage file health
did restore                       # Restore from most recent backup
did restore <n>                   # Restore from backup #n (1-3)
did report @project               # Show all entries for a specific project with totals
did report #tag                   # Show all entries with a specific tag
did report --by project           # Show hours grouped by all projects
did report --by tag               # Show hours grouped by all tags
did report @project --last 7      # Project report for last 7 days
did report --by project --from 2024-01-01 --to 2024-01-31  # Project breakdown for date range
```

Duration format: `Yh` (hours), `Ym` (minutes), or `YhYm` (combined). Max 24 hours per entry.

## Data Storage

Entries stored in JSONL format at `~/.config/did/entries.jsonl` (Linux), uses `os.UserConfigDir()` for cross-platform support.

**Soft Delete Behavior:**
- Deleted entries are marked with a `deleted_at` timestamp rather than removed
- Deleted entries can be recovered with `did undo`
- Entries deleted more than 7 days ago are automatically purged during delete operations
- Use `did purge` to manually remove all soft-deleted entries immediately

## Conventions

- Tests alongside source files (`*_test.go`)
- Internal packages under `internal/`
- Errors written to stderr, success to stdout
- ISO week standard (Monday-Sunday)


## Development help

- If running in a sandboxed environment with strict permissions on what you can run, execute commands through `bash -c`.
- Always keep documentation up to date. This includes _all_ markdown files, and code docstrings.
