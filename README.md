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
- View entries for today, yesterday, this week, or last week
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
