# did

[![CI](https://github.com/xolan/did/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/xolan/did/actions/workflows/ci.yml)
[![Lint](https://github.com/xolan/did/actions/workflows/lint.yml/badge.svg?branch=master)](https://github.com/xolan/did/actions/workflows/lint.yml)
[![codecov](https://codecov.io/gh/xolan/did/branch/master/graph/badge.svg)](https://codecov.io/gh/xolan/did)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple CLI tool for tracking time spent on tasks.

## Disclaimer

This is an experimental project created almost exclusively using [Claude Code](https://claude.ai/claude-code) and [AutoClaude](https://github.com/almaraz97/AutoClaude). While functional, it should be treated as a proof-of-concept for AI-assisted development rather than production-ready software.

## Features

- Log work activities with duration
- Timer mode for tracking work in real-time
- View entries for today, yesterday, this week, or last week
- Organize entries with projects (`@project`) and tags (`#tag`)
- Search entries by keyword
- Export to JSON or CSV
- Generate reports grouped by project or tag
- View statistics for week or month
- Simple duration format (hours and minutes)
- Data stored locally in JSONL format

## Test coverage

![TestCoverage](https://codecov.io/github/xolan/did/graphs/tree.svg?token=54AUNJMJVP)

## Prerequisites

- **Go 1.25+** - Required to build the project
- **mise** (optional) - For managing Go version
- **just** (optional) - For running build commands

## Installation

### Using Homebrew (recommended for macOS/Linux)

```bash
# Add the tap
brew tap xolan/tap

# Install did
brew install did
```

Shell completions are automatically installed. To enable them:

**Bash:**
```bash
# Add to ~/.bashrc or ~/.bash_profile
eval "$(did completion bash)"
```

**Zsh:**
```bash
# Add to ~/.zshrc
eval "$(did completion zsh)"
```

**Fish:**
```bash
# Add to ~/.config/fish/config.fish
did completion fish | source
```

### Using just

```bash
# Install dependencies and build tools
just setup

# Build and install to ~/.local/bin/
just install
```

### Manual installation

```bash
# Download dependencies
go mod download

# Build the binary
go build -o dist/did .

# Install to your PATH (example)
cp dist/did ~/.local/bin/
```

## Usage

### Log a work entry

```bash
did <description> for <duration>
```

**Examples:**

```bash
did feature X for 2h
did fixing login bug for 30m
did code review for 1h
did meeting with team for 45m
```

### Projects and Tags

Organize entries with `@project` and `#tag` in descriptions:

```bash
did fix login bug @acme for 1h              # Assign to project 'acme'
did code review #review for 30m             # Add tag 'review'
did API work @client #backend #api for 2h   # Project with multiple tags
```

### Timer Mode

As an alternative to specifying duration upfront, you can start a timer and stop it when done:

```bash
did start <description>           # Start a timer
did status                        # Check current timer status
did stop                          # Stop timer and create entry
```

**Examples:**

```bash
did start fixing auth bug                   # Start simple timer
did start code review @acme                 # Start timer with project
did start API work @client #backend #api    # Start timer with project and tags
did status                                  # Shows elapsed time and description
did stop                                    # Creates entry with calculated duration
```

**Timer flags:**

| Flag | Description |
|------|-------------|
| `--force`, `-f` | Override existing timer when starting a new one |

**Notes:**
- Timer state persists across terminal sessions (closing the terminal doesn't lose your tracking)
- Duration is automatically calculated and rounded to the nearest minute (minimum 1 minute)
- If a timer is already running when you try to start a new one, you'll be warned (use `--force` to override)
- The original `did X for Y` syntax continues to work unchanged

### View entries

| Command | Description |
|---------|-------------|
| `did` | List today's entries |
| `did y` | List yesterday's entries |
| `did w` | List this week's entries |
| `did lw` | List last week's entries |
| `did 2024-01-15` | List entries for a specific date |
| `did from 2024-01-01 to 2024-01-31` | List entries for a date range |
| `did last 7 days` | List entries from the past 7 days |

**Example output:**

```
Entries for today:
--------------------------------------------------
  09:30  feature X (2h)
  14:00  fixing login bug (30m)
--------------------------------------------------
Total: 2h 30m
```

### Filter by project or tag

```bash
did --project acme                # Today's entries for project 'acme'
did @acme                         # Same as above (shorthand)
did w --tag bugfix                # This week's entries tagged 'bugfix'
did #bugfix                       # Today's entries tagged 'bugfix'
did y @client #urgent             # Yesterday's entries filtered
did --project acme --tag review   # Multiple filters
```

### Edit entries

```bash
did edit <index> --description 'new text'    # Update description
did edit <index> --duration 2h               # Update duration
did edit <index> --description 'text' --duration 2h    # Update both
```

### Delete and restore entries

```bash
did delete <index>      # Delete entry (with confirmation)
did delete <index> -y   # Delete without confirmation
did undo                # Restore most recently deleted entry
did purge               # Permanently remove all deleted entries
did purge -y            # Purge without confirmation
```

### Search entries

```bash
did search meeting                           # Search for 'meeting'
did search bug --from 2024-01-01             # Search from a date
did search review --last 7                   # Search last 7 days
did search api --from 2024-01-01 --to 2024-01-31    # Search date range
```

### Export entries

```bash
# JSON export
did export json                    # Export all entries
did export json > backup.json      # Export to file
did export json --from 2024-01-01  # From a specific date
did export json --last 7           # Last 7 days
did export json @acme #review      # With filters

# CSV export
did export csv                     # Export all entries
did export csv > backup.csv        # Export to file
did export csv --last 30           # Last 30 days
```

**Export flags:**

| Flag | Description |
|------|-------------|
| `--from <date>` | Start date (YYYY-MM-DD or DD/MM/YYYY) |
| `--to <date>` | End date (YYYY-MM-DD or DD/MM/YYYY) |
| `--last <n>` | Last N days |

### Reports

```bash
# Single project/tag reports
did report @acme                   # All entries for project 'acme'
did report #review                 # All entries tagged 'review'
did report @acme --last 7          # Project report for last 7 days

# Grouped reports
did report --by project            # Hours grouped by all projects
did report --by tag                # Hours grouped by all tags
did report --by project --last 30  # Project breakdown for last 30 days
```

**Report flags:**

| Flag | Description |
|------|-------------|
| `--by <type>` | Group by 'project' or 'tag' |
| `--from <date>` | Start date |
| `--to <date>` | End date |
| `--last <n>` | Last N days |

### Statistics

```bash
did stats           # Statistics for current week
did stats --month   # Statistics for current month
```

### Maintenance commands

```bash
did validate              # Check storage file health
did restore               # Restore from most recent backup
did restore 2             # Restore from backup #2 (1-3 available)
did config                # Display current configuration
did config --init         # Create sample config file
```

### Global flags

| Flag | Description |
|------|-------------|
| `--project <name>` | Filter entries by project |
| `--tag <name>` | Filter entries by tag (can be repeated) |
| `-h, --help` | Help for any command |
| `-v, --version` | Show version |

## Duration Format

| Format | Description | Example |
|--------|-------------|---------|
| `Yh` | Hours | `2h` = 2 hours |
| `Ym` | Minutes | `30m` = 30 minutes |
| `YhYm` | Combined | `1h30m` = 1 hour 30 minutes |

**Note:** Maximum duration per entry is 24 hours.

## Date Format

Dates can be specified in two formats:
- `YYYY-MM-DD` (e.g., `2024-01-15`)
- `DD/MM/YYYY` (e.g., `15/01/2024`)

## Data Storage

Entries are stored in JSONL (JSON Lines) format at:

| Platform | Location |
|----------|----------|
| Linux    | `~/.config/did/entries.jsonl` |
| macOS    | `~/Library/Application Support/did/entries.jsonl` |
| Windows  | `%AppData%/did/entries.jsonl` |

**Timer State:**

Active timer state is stored in `timer.json` in the same config directory:

| Platform | Location |
|----------|----------|
| Linux    | `~/.config/did/timer.json` |
| macOS    | `~/Library/Application Support/did/timer.json` |
| Windows  | `%AppData%/did/timer.json` |

The timer file is automatically created when you start a timer and removed when you stop it.

## Development

```bash
# Run tests
just test

# Run linter
just lint

# Build the binary
just build
```

## License

See [LICENSE](LICENSE) for details.
