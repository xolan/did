# did

[![CI](https://github.com/xolan/did/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/xolan/did/actions/workflows/ci.yml)
[![Lint](https://github.com/xolan/did/actions/workflows/lint.yml/badge.svg?branch=master)](https://github.com/xolan/did/actions/workflows/lint.yml)
[![codecov](https://codecov.io/gh/xolan/did/branch/master/graph/badge.svg)](https://codecov.io/gh/xolan/did)

A simple CLI tool for tracking time spent on tasks.

## Features

- Log work activities with duration
- View entries for today, yesterday, this week, or last week
- Simple duration format (hours and minutes)
- Data stored locally in JSONL format

## Prerequisites

- **Go 1.25+** - Required to build the project
- **mise** (optional) - For managing Go version
- **just** (optional) - For running build commands

## Installation

### Using just (recommended)

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

### View entries

| Command   | Description              |
|-----------|--------------------------|
| `did`     | List today's entries     |
| `did y`   | List yesterday's entries |
| `did w`   | List this week's entries |
| `did lw`  | List last week's entries |

**Example output:**

```
Entries for today:
--------------------------------------------------
  09:30  feature X (2h)
  14:00  fixing login bug (30m)
--------------------------------------------------
Total: 2h 30m
```

## Duration Format

Durations are specified using a simple format:

| Format | Description | Example |
|--------|-------------|---------|
| `Yh`   | Hours       | `2h` = 2 hours |
| `Ym`   | Minutes     | `30m` = 30 minutes |

**Valid examples:** `1h`, `2h`, `30m`, `45m`, `8h`

**Note:** Maximum duration per entry is 24 hours.

## Data Storage

Entries are stored in JSONL (JSON Lines) format at:

| Platform | Location |
|----------|----------|
| Linux    | `~/.config/did/entries.jsonl` |
| macOS    | `~/Library/Application Support/did/entries.jsonl` |
| Windows  | `%AppData%/did/entries.jsonl` |

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
