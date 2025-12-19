# hevycli

[![Go Version](https://img.shields.io/github/go-mod/go-version/obay/hevycli)](https://go.dev/)
[![License](https://img.shields.io/github/license/obay/hevycli)](LICENSE)
[![Release](https://img.shields.io/github/v/release/obay/hevycli)](https://github.com/obay/hevycli/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/obay/hevycli)](https://goreportcard.com/report/github.com/obay/hevycli)
[![CI](https://github.com/obay/hevycli/actions/workflows/ci.yml/badge.svg)](https://github.com/obay/hevycli/actions/workflows/ci.yml)

A powerful command-line interface for the [Hevy](https://www.hevyapp.com/) fitness tracking platform.

## Features

- **Full CRUD Operations** - Manage workouts, routines, exercises, and folders
- **Analytics & Stats** - Track progress, view personal records, workout summaries
- **Interactive TUI** - Terminal UI for workout sessions, exercise search, routine builder
- **Multiple Output Formats** - JSON, table, and plain text for scripting
- **Shell Completion** - Bash, Zsh, Fish, and PowerShell support
- **AI-Agent Ready** - Structured output suitable for LLM consumption

## Installation

### Homebrew (macOS/Linux)

```bash
brew install --cask obay/tap/hevycli
```

### Scoop (Windows)

```bash
scoop bucket add obay https://github.com/obay/scoop-bucket
scoop install hevycli
```

### Go Install

```bash
go install github.com/obay/hevycli@latest
```

### Binary Download

Download pre-built binaries from the [Releases](https://github.com/obay/hevycli/releases) page.

## Quick Start

1. Get your API key from [Hevy Developer Settings](https://hevy.com/settings?developer) (requires Hevy Pro)

2. Initialize configuration:
   ```bash
   hevycli config init
   ```

3. Start using hevycli:
   ```bash
   hevycli workout list          # List recent workouts
   hevycli stats summary         # View workout summary
   hevycli routine builder       # Create routine interactively
   ```

## Commands

### Workouts

```bash
hevycli workout list              # List recent workouts
hevycli workout list --all        # List all workouts
hevycli workout get <id>          # Get workout details
hevycli workout count             # Get total workout count
hevycli workout create --file w.json   # Create from JSON
hevycli workout update <id> --file w.json  # Update workout
hevycli workout delete <id>       # Delete workout
hevycli workout start             # Start interactive session
```

### Routines

```bash
hevycli routine list              # List all routines
hevycli routine get <id>          # Get routine details
hevycli routine create --file r.json   # Create from JSON
hevycli routine update <id> --file r.json  # Update routine
hevycli routine builder           # Interactive routine builder
```

### Exercises

```bash
hevycli exercise list             # List exercise templates
hevycli exercise get <id>         # Get exercise details
hevycli exercise search "bench"   # Search exercises
hevycli exercise create --title "My Exercise" --type weight_reps --muscle chest
hevycli exercise interactive      # Interactive exercise browser
```

### Folders

```bash
hevycli folder list               # List routine folders
hevycli folder get <id>           # Get folder details
hevycli folder create "Name"      # Create new folder
```

### Analytics

```bash
hevycli stats summary             # Monthly workout summary
hevycli stats summary --period week    # Weekly summary
hevycli stats summary --period year    # Yearly summary
hevycli stats progress "Bench Press"   # Track exercise progress
hevycli stats progress "Squat" --metric 1rm  # Estimated 1RM over time
hevycli stats records             # View personal records
hevycli stats records --exercise "Bench"  # Filter by exercise
```

### Configuration

```bash
hevycli config init               # Interactive setup
hevycli config show               # Display current config
hevycli config set api-key <key>  # Set API key
```

### Shell Completion

```bash
# Bash
hevycli completion bash > /etc/bash_completion.d/hevycli

# Zsh
hevycli completion zsh > "${fpath[1]}/_hevycli"

# Fish
hevycli completion fish > ~/.config/fish/completions/hevycli.fish

# PowerShell
hevycli completion powershell > hevycli.ps1
```

## Output Formats

### Table (default)

```bash
hevycli workout list
```

### JSON (for scripting/AI agents)

```bash
hevycli workout list --output json
```

### Plain (pipe-delimited)

```bash
hevycli workout list --output plain
```

## Configuration

Configuration file location: `~/.hevycli/config.yaml`

```yaml
api:
  key: "your-api-key"

display:
  output_format: table  # table, json, plain
  color: true
  units: metric  # metric, imperial
```

### Environment Variables

```bash
HEVYCLI_API_KEY=your-key
HEVYCLI_OUTPUT_FORMAT=json
HEVYCLI_UNITS=imperial
HEVYCLI_NO_COLOR=true
```

## AI Agent Integration

hevycli is designed for AI agent consumption:

```bash
# Fetch workout history as structured JSON
hevycli workout list --since 2024-12-01 --output json

# Analyze exercise progress
hevycli stats progress "Bench Press" --output json

# Create workouts programmatically
hevycli workout create --file /tmp/workout.json --output json
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | API authentication error |
| 4 | API rate limit exceeded |
| 5 | Network error |
| 6 | Resource not found |
| 7 | Validation error |

## Development

### Build from Source

```bash
git clone https://github.com/obay/hevycli.git
cd hevycli
go build -o hevycli .
```

### Run Tests

```bash
go test ./...
```

### Release

Releases are automated via GoReleaser on version tags:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Requirements

- Hevy Pro subscription (for API access)
- API key from [Hevy Developer Settings](https://hevy.com/settings?developer)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related

- [Hevy App](https://www.hevyapp.com/) - The official Hevy mobile app
- [Hevy API Docs](https://api.hevyapp.com/docs/) - Official API documentation
