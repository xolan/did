# PROJECT KNOWLEDGE BASE

**Generated:** 2026-01-11
**Commit:** 40d8407
**Branch:** master

## OVERVIEW

Go CLI time tracker using Cobra. Log work with durations, filter by project/tag, timer mode for real-time tracking.

**Tech Stack:** Go 1.25+, Cobra CLI framework, Bubble Tea TUI, bubbletint theming

## STRUCTURE

```
did/
├── main.go           # Entry point, version injection via ldflags
├── cmd/              # 30 files: Cobra commands + DI (see cmd/AGENTS.md)
└── internal/         # 9 domain packages (below)
```

### internal/ packages

| Package | Files | Purpose |
|---------|-------|---------|
| `entry/` | 4 | `Entry` struct, `ParseDuration`, `ParseProjectAndTags` |
| `storage/` | 5 | JSONL persistence, atomic writes, soft delete, backups |
| `timeutil/` | 6 | Date ranges, week boundaries, timezone handling |
| `config/` | 2 | TOML config, `WeekStartDay`, `Timezone` validation |
| `filter/` | 2 | `Filter` struct with AND logic (keyword+project+tags) |
| `timer/` | 2 | Timer state persistence across sessions |
| `stats/` | 2 | Statistics calculations, project/tag breakdowns |
| `osutil/` | 2 | `PathProvider` interface for cross-platform paths |
| `app/` | 1 | `const Name = "did"` |
| `tui/` | 10+ | Bubble Tea TUI, views, theming via bubbletint |
| `service/` | 8 | Business logic services for TUI |
| `cli/` | 2 | CLI helpers, formatters |

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add new command | `cmd/` | Copy existing pattern, add to `rootCmd` in `init()` |
| Modify entry format | `internal/entry/entry.go` | Update struct + JSON tags |
| Change storage format | `internal/storage/jsonl.go` | Atomic writes via temp file |
| Add time filter | `internal/timeutil/datefilter.go` | Follow `ThisWeek()`/`LastWeek()` pattern |
| Modify config | `internal/config/config.go` | Add field, update `Validate()` |
| Test any command | `cmd/*_test.go` | Use `SetDeps()` pattern |
| Modify TUI views | `internal/tui/views/*.go` | Follow existing view pattern |
| Add TUI theme | `internal/tui/ui/theme.go` | Uses bubbletint registry |
| Modify TUI styles | `internal/tui/ui/styles.go` | Update `NewStylesFromRegistry()` |

## DEPENDENCY GRAPH

```
cmd/ ──┬── storage ──── entry
       ├── timer        │
       ├── config       ▼
       ├── filter ───── entry
       └── timeutil
           │
       osutil (cross-platform paths)
```

## CONVENTIONS

### Testing (100% coverage target)
- `Deps` struct in `cmd/deps.go` for DI
- `SetDeps(d)` / `ResetDeps()` in tests
- `osutil.Provider` swappable for path errors
- Timer tests use `setupTimerTest(t)` helper

### File writes
- **Atomic**: temp file + `os.Rename()` (storage, timer)
- **JSONL**: one JSON object per line, append-only

### Output
- Success → `deps.Stdout`
- Errors → `deps.Stderr` + `deps.Exit(1)`
- Helpful hints in error messages

## ANTI-PATTERNS (DO NOT)

| Pattern | Why |
|---------|-----|
| Duration > 24h | Max `1440` minutes per entry |
| Zero duration | Rejected by `ParseDuration` |
| Tags OR logic | Tags use AND: must match ALL specified |
| Parallel tests without `SetDeps` | Global `deps` will race |
| Direct `os.Exit()` in cmd | Use `deps.Exit()` for testability |
| Skip `defer ResetDeps()` | Leaks test state |

## GOTCHAS

- **Week start**: Configurable (`monday`/`sunday`), affects `--this-week`/`--prev-week`
- **Soft delete**: 7-day grace period, then auto-purged
- **Timer override**: Need `--force` flag if timer already running
- **Date format**: ISO `YYYY-MM-DD` preferred over `DD/MM/YYYY` for ambiguous dates
- **Entry index**: 1-based for users, 0-based internally
- **Multiple @project**: Last one wins

## COMMANDS

```bash
just setup         # Install mise tools + deps
just test          # go test ./... with coverage
just lint          # golangci-lint v2.7.2
just build         # Build to dist/did
just release       # GoReleaser snapshot
```

## BUILD

- **Version**: Injected via ldflags (`-X main.version=...`)
- **CGO**: Disabled (`CGO_ENABLED=0`)
- **Platforms**: linux/darwin (amd64/arm64), windows (amd64)
- **Release**: GoReleaser → GitHub, Homebrew tap, Docker

## DATA STORAGE

| Data | Location | Format |
|------|----------|--------|
| Entries | `~/.config/did/entries.jsonl` | JSONL (one JSON per line) |
| Timer | `~/.config/did/timer.json` | JSON (auto-removed on stop) |
| Config | `~/.config/did/config.toml` | TOML (optional) |

Cross-platform via `os.UserConfigDir()`: Linux `~/.config/`, macOS `~/Library/Application Support/`, Windows `%AppData%`.

## CONFIGURATION

Config file is **optional** — all settings have sensible defaults.

| Option | Values | Default | Affects |
|--------|--------|---------|---------|
| `week_start_day` | `"monday"`, `"sunday"` | `"monday"` | `--this-week`, `--prev-week`, stats |
| `timezone` | IANA name or `"Local"` | `"Local"` | All time operations |
| `theme` | Any bubbletint theme name | `"dracula"` | TUI color scheme |

```bash
did config --init  # Create sample config.toml
did config         # Show current settings
```

## TUI

Launch with `did tui`. 280+ themes available via bubbletint.

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab`, `1-5` | Switch views |
| `j/k`, `↑/↓` | Navigate entries |
| `t` | Today's entries (Entries view) / Open theme selector (Config view) |
| `y` | Yesterday's entries |
| `w` | This week's entries |
| `r` | Refresh data |
| `Enter` | Select / Open theme selector |
| `Esc` | Cancel / Close selector |
| `q` | Quit |

### Theme Architecture

```
ThemeProvider (ui/theme.go)
    ├── Uses bubbletint.Registry
    ├── Stores current theme name
    └── Generates Styles from theme colors

ThemeChangedMsg broadcasts to all views when theme changes
```

## DEVELOPMENT NOTES

- If running in sandboxed environment with strict permissions, execute commands through `bash -c`
- Always keep documentation up to date (all markdown files and code docstrings)

---
