# cmd/ - CLI Commands

## OVERVIEW

30 Go files implementing Cobra commands + dependency injection infrastructure.

## STRUCTURE

| File | Command | Key Function |
|------|---------|--------------|
| `root.go` | `did` (default) | Entry creation, listing, edit, validate |
| `deps.go` | — | `Deps` struct, `SetDeps`, `ResetDeps` |
| `io_errors.go` | — | Error helpers (excluded from coverage) |
| **Timer** |||
| `start.go` | `did start` | `startTimer()` |
| `stop.go` | `did stop` | `stopTimer()`, `calculateDurationMinutes()` |
| `status.go` | `did status` | `showStatus()` |
| **CRUD** |||
| `delete.go` | `did delete` | Soft delete with confirmation |
| `undo.go` | `did undo` | Restore most recent delete |
| `purge.go` | `did purge` | Permanently remove deleted |
| **Query** |||
| `search.go` | `did search` | Keyword search with date filters |
| `report.go` | `did report` | Project/tag reports, `--by` grouping |
| `stats.go` | `did stats` | Weekly/monthly statistics |
| **Data** |||
| `export.go` | `did export` | JSON/CSV export with filters |
| `restore.go` | `did restore` | Restore from backup (1-3) |
| `config.go` | `did config` | Display/init config file |
| `completion.go` | `did completion` | Shell completions |

## DEPENDENCY INJECTION

```go
// deps.go - The pattern ALL commands use
var deps = DefaultDeps()  // Global singleton

type Deps struct {
    Stdout      io.Writer              // Capture output in tests
    Stderr      io.Writer              // Capture errors in tests
    Stdin       io.Reader              // Mock user input (y/n)
    Exit        func(code int)         // Prevent real exit in tests
    StoragePath func() (string, error) // Mock storage location
    TimerPath   func() (string, error) // Mock timer location
    Config      config.Config          // Test config values
}
```

### Test setup pattern

```go
func TestXxx(t *testing.T) {
    stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
    SetDeps(&Deps{
        Stdout: stdout,
        Stderr: stderr,
        Stdin:  strings.NewReader("y\n"),
        Exit:   func(int) {},
        StoragePath: func() (string, error) { return tmpFile, nil },
        TimerPath:   timer.GetTimerPath,
        Config:      DefaultDeps().Config,
    })
    defer ResetDeps()
}
```

## WHERE TO LOOK

| Task | File | Pattern |
|------|------|---------|
| Add new command | New `xxx.go` | Copy `search.go`, add `rootCmd.AddCommand()` |
| Add flag | Target command | `cmd.Flags().StringVarP()` in `init()` |
| Confirmation prompt | `delete.go:79` | `promptConfirmation()` with `deps.Stdin` |
| Date filtering | `root.go:169-214` | Mutually exclusive flag validation |
| Project/tag parsing | `root.go:96` | `parseShorthandFilters()` for `@proj #tag` |

## FLAG PATTERNS

### Time period flags (mutually exclusive)
--yesterday, -y | --this-week, -w | --prev-week | --this-month, -m | --prev-month | --last n, -l | --date date, -d | --from date --to date

### Filter flags (inherited by subcommands)
--project name (or @name shorthand) | --tag name (or #name shorthand, repeatable)

## CONVENTIONS

- Handler functions named after command: `startTimer()`, `stopTimer()`, `showStatus()`
- All output via `deps.Stdout`/`deps.Stderr`
- Fatal errors: `deps.Exit(1)` after printing to stderr
- Tests have matching `*_test.go` files
- Table-driven tests with `t.Run()` subtests

## DO NOT

- Call `os.Exit()` directly — use `deps.Exit()`
- Read `os.Stdin` directly — use `deps.Stdin`
- Print to stdout/stderr directly — use `deps`
- Forget `defer ResetDeps()` in tests
- Mix time period flags (validation rejects)

## CLI USAGE REFERENCE

### Entry Creation

```bash
did <description> for <duration>      # Log entry (e.g., "did feature X for 2h")
did fix bug @acme for 1h              # Log with project
did review #code #urgent for 30m      # Log with tags
```

### Timer Mode

```bash
did start <description>           # Start a timer
did start code review @acme       # Start timer with project
did status                        # Show current timer status
did stop                          # Stop timer and create entry
did start <desc> --force          # Override existing timer
```

**Notes:** Duration rounded to nearest minute (min 1m). Timer persists across sessions.

### View Entries

```bash
did                               # Today's entries
did -y                            # Yesterday
did -w                            # This week
did --prev-week                   # Previous week
did -m                            # This month
did --prev-month                  # Previous month
did -d 2024-01-15                 # Specific date
did --from 2024-01-01 --to 2024-01-31  # Date range
did -l 7                          # Last 7 days
```

### Filter, Edit, Delete

```bash
did @acme                         # Filter by project
did -w #bugfix                    # Filter by tag
did edit <index> --description X  # Edit description
did edit <index> --duration 2h    # Edit duration
did delete <index>                # Soft delete (7-day recovery)
did undo                          # Restore last delete
did purge                         # Permanent removal
```

### Search, Export, Reports

```bash
did search <keyword>              # Search entries
did export json                   # Export as JSON
did export csv                    # Export as CSV
did report @project               # Project report
did report --by project           # Hours by all projects
did stats                         # Weekly statistics
did stats --month                 # Monthly statistics
```

### Duration Format

| Format | Example |
|--------|---------|
| `Yh` | `2h` = 2 hours |
| `Ym` | `30m` = 30 minutes |
| `YhYm` | `1h30m` = 1.5 hours |

**Max:** 24 hours (1440 minutes) per entry.

---
