# CLAUDE.md

## Project Overview

**did** - A Go CLI tool for time tracking. Log work activities with durations and view entries by time period.

## Tech Stack

- Go 1.25+
- Cobra CLI framework

## Project Structure

```
main.go                           # Entry point
cmd/root.go                       # Cobra command definitions (did, y, w, lw)
internal/
  entry/
    entry.go                      # Entry struct (Timestamp, Description, DurationMinutes, RawInput)
    parser.go                     # Duration parsing (ParseDuration)
    parser_test.go
  storage/
    jsonl.go                      # JSONL storage (AppendEntry, ReadEntries, GetStoragePath)
    jsonl_test.go
  timeutil/
    datefilter.go                 # Date range utilities (Today, Yesterday, ThisWeek, LastWeek)
    datefilter_test.go
```

## Commands

```bash
just setup    # Install mise tools and download Go dependencies
just test     # Run test suite: go test ./...
just build    # Build binary: go build -o did .
just install  # Build and install to ~/.local/bin/
```

## CLI Usage

```bash
did work on <description> for <duration>  # Log entry (e.g., "did work on feature X for 2h")
did                                        # List today's entries
did y                                      # List yesterday's entries
did w                                      # List this week's entries
did lw                                     # List last week's entries
```

Duration format: `Yh` (hours) or `Ym` (minutes). Max 24 hours per entry.

## Data Storage

Entries stored in JSONL format at `~/.config/did/entries.jsonl` (Linux), uses `os.UserConfigDir()` for cross-platform support.

## Conventions

- Tests alongside source files (`*_test.go`)
- Internal packages under `internal/`
- Errors written to stderr, success to stdout
- ISO week standard (Monday-Sunday)
