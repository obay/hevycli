# Product Requirements Document (PRD)
## hevycli - Hevy Workout CLI

**Version:** 1.0
**Author:** Ahmad Obay
**Date:** December 2024
**Status:** Draft

---

## 1. Executive Summary

**hevycli** is an open-source command-line interface for the [Hevy](https://www.hevyapp.com/) fitness tracking platform. It provides both human-friendly and machine-readable interfaces for managing workouts, routines, and exercise data. The tool is designed with dual audiences in mind:

1. **End users** who want efficient terminal-based workout management
2. **AI agents** (Claude Code, Warp AI, etc.) that act as personal trainers and need structured access to fitness data

The CLI will serve as the foundation for an AI-powered personal training system, enabling agentic workflows to read workout history, create new workouts, update existing sessions, and provide data-driven training recommendations.

---

## 2. Problem Statement

### Current Pain Points

1. **No reliable CLI tool exists** - The existing CLI is very basic and limited in functionality
2. **Mobile-first limitations** - Hevy's mobile app is excellent but inefficient for power users and automation
3. **AI integration gap** - The existing `hevy-mcp` MCP server depends on an external service that proved unreliable, making the complicated dependency chain useless for production
4. **Data accessibility** - Workout data is locked in the app, making analysis and automation difficult

### Target Users

| User Type | Needs |
|-----------|-------|
| **Power Users** | Fast workout logging, bulk operations, keyboard-driven workflow |
| **Developers** | Scriptable interface, structured output, automation capabilities |
| **AI Agents** | Machine-readable data, predictable output formats, comprehensive CRUD operations |
| **Open Source Community** | Well-documented, extensible, easy to contribute |

---

## 3. Goals & Success Criteria

### Primary Goals

1. **Complete API Coverage** - Support all endpoints available in the official Hevy API
2. **Dual-Mode Output** - Human-friendly tables AND machine-readable JSON/plain text
3. **AI-Agent Ready** - Structured, predictable output suitable for LLM consumption
4. **Rich TUI Experience** - Interactive menus and forms using Bubble Tea for complex operations
5. **Production Quality** - Reliable, well-tested, properly documented open-source project

### Success Metrics

- [ ] 100% coverage of Hevy API v1 endpoints
- [ ] Sub-second response time for cached operations
- [ ] Zero breaking changes to output format after v1.0
- [ ] Comprehensive test coverage (>80%)
- [ ] Clear documentation with examples for every command

---

## 4. Technical Architecture

### 4.1 Technology Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Language** | Go 1.21+ | Fast, single binary, excellent CLI ecosystem |
| **CLI Framework** | [Cobra](https://github.com/spf13/cobra) | Industry standard, subcommand support, auto-generated help |
| **Configuration** | [Viper](https://github.com/spf13/viper) | Config files, env vars, flags integration |
| **TUI Framework** | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Modern, composable terminal UI |
| **TUI Components** | [Bubbles](https://github.com/charmbracelet/bubbles) | Pre-built inputs, tables, spinners |
| **Styling** | [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Consistent terminal styling |
| **Tables** | [Table](https://github.com/charmbracelet/lipgloss/tree/master/table) | Beautiful ASCII/Unicode tables |
| **HTTP Client** | Standard library + [resty](https://github.com/go-resty/resty) | Simple REST client with retries |
| **JSON** | Standard library | Native Go JSON support |
| **Testing** | `testing` + [testify](https://github.com/stretchr/testify) | Assertions and mocking |
| **Linting** | `golangci-lint` | Comprehensive code quality |

### 4.2 Existing Go Package

An existing Go client library exists: [github.com/swrm-io/go-hevy](https://pkg.go.dev/github.com/swrm-io/go-hevy)

**Current capabilities:**
- `AllWorkouts()` - Fetch all workouts with pagination
- `Workout(id)` - Get single workout by UUID
- `WorkoutCount()` - Get total workout count
- `AllWorkoutEvents(since)` - Get workout events (updates/deletes)
- `Routines()` - Get all routines

**Limitations:**
- No create/update workout support
- No exercise templates support
- No routine folders support
- Marked as "work in progress"

**Decision:** We will either:
1. Fork and extend `go-hevy` with missing functionality, OR
2. Build our own client layer (preferred for full control)

### 4.3 Project Structure

```
hevycli/
├── cmd/                    # Cobra commands
│   ├── root.go            # Root command, global flags
│   ├── workout/           # Workout subcommands
│   │   ├── list.go
│   │   ├── get.go
│   │   ├── create.go
│   │   ├── update.go
│   │   ├── delete.go
│   │   └── start.go       # Interactive workout session
│   ├── routine/           # Routine subcommands
│   │   ├── list.go
│   │   ├── get.go
│   │   ├── create.go
│   │   └── update.go
│   ├── exercise/          # Exercise template subcommands
│   │   ├── list.go
│   │   ├── get.go
│   │   └── search.go
│   ├── folder/            # Routine folder subcommands
│   │   ├── list.go
│   │   ├── create.go
│   │   └── get.go
│   ├── stats/             # Analytics commands
│   │   ├── summary.go
│   │   ├── progress.go
│   │   └── records.go
│   └── config/            # Configuration commands
│       ├── init.go
│       ├── show.go
│       └── set.go
├── internal/
│   ├── api/               # Hevy API client
│   │   ├── client.go
│   │   ├── workouts.go
│   │   ├── routines.go
│   │   ├── exercises.go
│   │   ├── folders.go
│   │   └── types.go
│   ├── config/            # Configuration management
│   │   └── config.go
│   ├── output/            # Output formatting
│   │   ├── formatter.go
│   │   ├── json.go
│   │   ├── table.go
│   │   └── plain.go
│   └── tui/               # Bubble Tea components
│       ├── workout/
│       ├── routine/
│       └── common/
├── pkg/                   # Public packages (if needed)
├── docs/                  # Documentation
├── scripts/               # Build and release scripts
├── .goreleaser.yml        # Release automation
├── go.mod
├── go.sum
├── LICENSE                # MIT or Apache-2.0
└── README.md
```

---

## 5. Hevy API Reference

### 5.1 Authentication

- **Method:** API Key in header
- **Header:** `api-key: YOUR_API_KEY`
- **Requirement:** Hevy Pro subscription
- **Key Location:** https://hevy.com/settings?developer

### 5.2 Base URL

```
https://api.hevyapp.com/v1
```

### 5.3 Available Endpoints

Based on research from the official Swagger docs and community implementations:

#### Workouts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workouts` | List workouts (paginated, max 10/page) |
| GET | `/workouts/{id}` | Get single workout by ID |
| GET | `/workouts/count` | Get total workout count |
| GET | `/workouts/events` | Get workout events (updates/deletes) since timestamp |
| POST | `/workouts` | Create new workout |
| PUT | `/workouts/{id}` | Update existing workout |
| DELETE | `/workouts/{id}` | Delete workout |

#### Routines

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/routines` | List all routines |
| GET | `/routines/{id}` | Get single routine by ID |
| POST | `/routines` | Create new routine |
| PUT | `/routines/{id}` | Update existing routine |

#### Exercise Templates

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/exercise_templates` | List exercise templates (paginated) |
| GET | `/exercise_templates/{id}` | Get single exercise template |

#### Routine Folders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/routine_folders` | List routine folders |
| GET | `/routine_folders/{id}` | Get single folder by ID |
| POST | `/routine_folders` | Create new folder |

#### Webhooks (Future consideration)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/webhooks` | Get webhook subscription |
| POST | `/webhooks` | Create webhook subscription |
| DELETE | `/webhooks` | Delete webhook subscription |

### 5.4 Data Models

#### Workout

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string",
  "start_time": "ISO8601",
  "end_time": "ISO8601",
  "created_at": "ISO8601",
  "updated_at": "ISO8601",
  "exercises": [Exercise]
}
```

#### Exercise

```json
{
  "index": "int",
  "title": "string",
  "notes": "string",
  "exercise_template_id": "string",
  "superset_id": "int|null",
  "sets": [Set]
}
```

#### Set

```json
{
  "index": "int",
  "set_type": "normal|warmup|dropset|failure",
  "weight_kg": "float",
  "reps": "int",
  "distance_meters": "float",
  "duration_seconds": "int",
  "rpe": "float"
}
```

#### Routine

```json
{
  "id": "uuid",
  "title": "string",
  "folder_id": "uuid|null",
  "created_at": "ISO8601",
  "updated_at": "ISO8601",
  "exercises": [Exercise]
}
```

---

## 6. Command Reference

### 6.1 Global Flags

```bash
hevycli [command] [flags]

Global Flags:
  --config string    Config file (default: ~/.hevycli/config.yaml)
  --output string    Output format: json, table, plain (default: table)
  --no-color         Disable colored output
  --quiet            Suppress non-essential output
  --verbose          Enable verbose/debug output
  --help             Show help for command
```

### 6.2 Configuration Commands

```bash
# Initialize configuration (interactive)
hevycli config init

# Show current configuration
hevycli config show

# Set configuration value
hevycli config set api-key <key>
hevycli config set default-output json
hevycli config set units metric|imperial
```

### 6.3 Workout Commands

```bash
# List workouts
hevycli workout list [flags]
  --limit int        Number of workouts to fetch (default: 10)
  --page int         Page number for pagination
  --all              Fetch all workouts (auto-pagination)
  --since string     Filter workouts since date (YYYY-MM-DD)
  --until string     Filter workouts until date (YYYY-MM-DD)

# Get single workout
hevycli workout get <workout-id>
  --exercises        Include detailed exercise data (default: true)
  --sets             Include detailed set data (default: true)

# Get workout count
hevycli workout count

# Create workout (from JSON file or interactive)
hevycli workout create [flags]
  --file string      JSON file with workout data
  --from-routine     Create from routine (interactive selection)
  --interactive      Use interactive TUI mode
  --title string     Workout title
  --date string      Workout date (default: today)

# Update workout
hevycli workout update <workout-id> [flags]
  --file string      JSON file with updated workout data
  --interactive      Use interactive TUI mode
  --title string     Update workout title
  --add-exercise     Add exercise interactively

# Delete workout
hevycli workout delete <workout-id> [flags]
  --force            Skip confirmation prompt

# Start interactive workout session (TUI)
hevycli workout start [flags]
  --from-routine     Start from a routine template
  --routine-id       Specific routine ID to use

# Get workout events/changes
hevycli workout events [flags]
  --since string     Events since date (YYYY-MM-DD)
  --type string      Filter by event type: updated, deleted
```

### 6.4 Routine Commands

```bash
# List routines
hevycli routine list [flags]
  --folder string    Filter by folder ID

# Get single routine
hevycli routine get <routine-id>

# Create routine
hevycli routine create [flags]
  --file string      JSON file with routine data
  --interactive      Use interactive TUI mode
  --title string     Routine title
  --folder string    Folder ID to place routine in

# Update routine
hevycli routine update <routine-id> [flags]
  --file string      JSON file with updated routine data
  --interactive      Use interactive TUI mode
  --title string     Update routine title
```

### 6.5 Exercise Commands

```bash
# List exercise templates
hevycli exercise list [flags]
  --limit int        Number of exercises to fetch
  --page int         Page number
  --all              Fetch all exercises

# Get single exercise template
hevycli exercise get <exercise-id>

# Search exercises
hevycli exercise search <query> [flags]
  --muscle string    Filter by muscle group
  --equipment string Filter by equipment type
```

### 6.6 Folder Commands

```bash
# List routine folders
hevycli folder list

# Get folder details
hevycli folder get <folder-id>

# Create folder
hevycli folder create <name>
```

### 6.7 Stats/Analytics Commands

```bash
# Overall summary
hevycli stats summary [flags]
  --period string    Time period: week, month, year, all (default: month)

# Exercise progress tracking
hevycli stats progress <exercise-name> [flags]
  --metric string    Metric: weight, volume, reps, 1rm (default: weight)
  --period string    Time period to analyze

# Personal records
hevycli stats records [flags]
  --exercise string  Filter by exercise
  --limit int        Number of records to show
```

---

## 7. Output Formats

### 7.1 Table Format (Default - Human Readable)

```
$ hevycli workout list --limit 3

┌──────────────────────────────────────┬─────────────────────┬────────────────┬──────────┬────────────┐
│ ID                                   │ Title               │ Date           │ Duration │ Exercises  │
├──────────────────────────────────────┼─────────────────────┼────────────────┼──────────┼────────────┤
│ a1b2c3d4-e5f6-7890-abcd-ef1234567890 │ Push Day            │ 2024-12-17     │ 1h 15m   │ 5          │
│ b2c3d4e5-f6a7-8901-bcde-f12345678901 │ Pull Day            │ 2024-12-15     │ 1h 05m   │ 6          │
│ c3d4e5f6-a7b8-9012-cdef-123456789012 │ Leg Day             │ 2024-12-13     │ 1h 30m   │ 7          │
└──────────────────────────────────────┴─────────────────────┴────────────────┴──────────┴────────────┘

Showing 3 of 127 workouts
```

### 7.2 JSON Format (Machine Readable)

```bash
$ hevycli workout list --limit 1 --output json
```

```json
{
  "workouts": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "title": "Push Day",
      "description": "Chest, shoulders, triceps",
      "start_time": "2024-12-17T10:00:00Z",
      "end_time": "2024-12-17T11:15:00Z",
      "duration_minutes": 75,
      "created_at": "2024-12-17T11:16:00Z",
      "updated_at": "2024-12-17T11:16:00Z",
      "exercise_count": 5,
      "total_sets": 20,
      "total_volume_kg": 5400.0,
      "exercises": [
        {
          "index": 0,
          "title": "Bench Press (Barbell)",
          "exercise_template_id": "abc123",
          "sets": [
            {
              "index": 0,
              "set_type": "warmup",
              "weight_kg": 60.0,
              "reps": 10,
              "rpe": null
            },
            {
              "index": 1,
              "set_type": "normal",
              "weight_kg": 100.0,
              "reps": 8,
              "rpe": 8.0
            }
          ]
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 1,
    "total_count": 127,
    "has_more": true
  }
}
```

### 7.3 Plain Text Format (Minimal, Scriptable)

```bash
$ hevycli workout list --limit 3 --output plain

a1b2c3d4-e5f6-7890-abcd-ef1234567890|Push Day|2024-12-17|75|5
b2c3d4e5-f6a7-8901-bcde-f12345678901|Pull Day|2024-12-15|65|6
c3d4e5f6-a7b8-9012-cdef-123456789012|Leg Day|2024-12-13|90|7
```

---

## 8. Interactive TUI Features

### 8.1 Workout Session Mode

An interactive full-screen terminal UI for logging workouts in real-time:

```
┌─ hevycli: Active Workout ─────────────────────────────────────────────────────┐
│                                                                                │
│  Push Day                                              Duration: 00:45:23      │
│  Started: 10:00 AM                                     3/5 exercises done      │
│                                                                                │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
│  ✓ Bench Press (Barbell)                                                       │
│    Set 1: 60kg × 10 (warmup) ✓                                                │
│    Set 2: 100kg × 8 ✓                                                          │
│    Set 3: 100kg × 7 ✓                                                          │
│                                                                                │
│  ✓ Incline Dumbbell Press                                                      │
│    Set 1: 32kg × 10 ✓                                                          │
│    Set 2: 32kg × 9 ✓                                                           │
│    Set 3: 30kg × 10 ✓                                                          │
│                                                                                │
│  → Overhead Press (Barbell)                            ← Current exercise      │
│    Set 1: [weight] × [reps]  _                                                 │
│    Set 2: pending                                                              │
│    Set 3: pending                                                              │
│                                                                                │
│  ○ Tricep Pushdown                                                             │
│  ○ Lateral Raise                                                               │
│                                                                                │
├────────────────────────────────────────────────────────────────────────────────┤
│  [Tab] Next field  [Enter] Save set  [n] New set  [e] Edit  [f] Finish  [?]   │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 8.2 Exercise Search & Selection

Interactive fuzzy-search for selecting exercises:

```
┌─ Select Exercise ─────────────────────────────────────────────────────────────┐
│                                                                                │
│  Search: bench p█                                                              │
│                                                                                │
│  > Bench Press (Barbell)              Chest, Triceps, Shoulders               │
│    Bench Press (Dumbbell)             Chest, Triceps, Shoulders               │
│    Bench Press (Smith Machine)        Chest, Triceps, Shoulders               │
│    Close Grip Bench Press             Triceps, Chest                          │
│    Incline Bench Press (Barbell)      Upper Chest, Shoulders                  │
│    Decline Bench Press (Barbell)      Lower Chest                             │
│                                                                                │
│  6 results                                                                     │
│                                                                                │
├────────────────────────────────────────────────────────────────────────────────┤
│  [↑↓] Navigate  [Enter] Select  [Esc] Cancel                                  │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 8.3 Routine Builder

Interactive routine creation:

```
┌─ Create Routine ──────────────────────────────────────────────────────────────┐
│                                                                                │
│  Title: Push Day A█                                                            │
│  Folder: [None] ▼                                                              │
│                                                                                │
│  Exercises:                                                                    │
│  ┌────┬──────────────────────────────┬───────┬──────────────────────────────┐ │
│  │ #  │ Exercise                     │ Sets  │ Notes                        │ │
│  ├────┼──────────────────────────────┼───────┼──────────────────────────────┤ │
│  │ 1  │ Bench Press (Barbell)        │ 4     │ 3-5 rep range                │ │
│  │ 2  │ Incline Dumbbell Press       │ 3     │ 8-10 rep range               │ │
│  │ 3  │ Overhead Press (Barbell)     │ 3     │                              │ │
│  └────┴──────────────────────────────┴───────┴──────────────────────────────┘ │
│                                                                                │
│  [a] Add exercise  [d] Delete  [↑↓] Move  [e] Edit  [Enter] Save              │
│                                                                                │
└────────────────────────────────────────────────────────────────────────────────┘
```

---

## 9. Configuration

### 9.1 Config File Location

```
~/.hevycli/config.yaml
```

### 9.2 Config File Structure

```yaml
# Hevy API Configuration
api:
  key: "your-api-key-here"
  base_url: "https://api.hevyapp.com/v1"  # Optional override

# Display Preferences
display:
  output_format: table  # table, json, plain
  color: true
  units: metric  # metric, imperial
  date_format: "2006-01-02"  # Go date format
  time_format: "15:04"  # Go time format

# TUI Preferences
tui:
  theme: default  # default, minimal, colorful
  confirm_destructive: true  # Confirm before delete operations

# Aliases (future feature)
aliases:
  bp: "exercise get bench-press-barbell"
  today: "workout list --since today"
```

### 9.3 Environment Variables

All config options can be overridden via environment variables:

```bash
HEVYCLI_API_KEY=your-key
HEVYCLI_OUTPUT_FORMAT=json
HEVYCLI_UNITS=imperial
HEVYCLI_NO_COLOR=true
```

---

## 10. AI Agent Integration

### 10.1 Design Principles for AI Consumption

1. **Consistent JSON schema** - Output structure never changes between versions
2. **Explicit error messages** - Clear, actionable error descriptions
3. **Predictable command structure** - Verb-noun pattern (workout list, workout create)
4. **Comprehensive data** - Always include IDs, timestamps, and computed fields
5. **Exit codes** - Standard exit codes for success (0) and various failure modes

### 10.2 Recommended AI Agent Workflow

```bash
# 1. Agent reads recent workout history
hevycli workout list --since 2024-12-01 --output json

# 2. Agent analyzes progress on specific exercise
hevycli stats progress "Bench Press" --output json

# 3. Agent creates new workout based on analysis
hevycli workout create --file /tmp/workout.json --output json

# 4. Agent updates workout if needed
hevycli workout update <id> --file /tmp/updated.json --output json
```

### 10.3 Error Response Format (JSON)

```json
{
  "error": {
    "code": "INVALID_API_KEY",
    "message": "The provided API key is invalid or expired",
    "details": "Please verify your API key at https://hevy.com/settings?developer",
    "timestamp": "2024-12-17T10:00:00Z"
  }
}
```

### 10.4 Exit Codes

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

---

## 11. Analytics & Progress Tracking

### 11.1 Summary Statistics

```bash
$ hevycli stats summary --period month --output json
```

```json
{
  "period": {
    "start": "2024-11-17",
    "end": "2024-12-17"
  },
  "workouts": {
    "total": 16,
    "average_duration_minutes": 72,
    "total_duration_hours": 19.2
  },
  "volume": {
    "total_kg": 89500,
    "average_per_workout_kg": 5593.75
  },
  "exercises": {
    "unique_count": 34,
    "total_sets": 384,
    "most_frequent": [
      {"name": "Bench Press (Barbell)", "count": 12},
      {"name": "Squat (Barbell)", "count": 10},
      {"name": "Deadlift (Barbell)", "count": 8}
    ]
  },
  "consistency": {
    "workouts_per_week": 4.0,
    "longest_streak_days": 14,
    "current_streak_days": 3
  }
}
```

### 11.2 Exercise Progress

```bash
$ hevycli stats progress "Bench Press" --metric 1rm --output json
```

```json
{
  "exercise": "Bench Press (Barbell)",
  "metric": "estimated_1rm",
  "unit": "kg",
  "data_points": [
    {"date": "2024-10-01", "value": 95.0},
    {"date": "2024-10-15", "value": 97.5},
    {"date": "2024-11-01", "value": 100.0},
    {"date": "2024-11-15", "value": 102.5},
    {"date": "2024-12-01", "value": 105.0}
  ],
  "analysis": {
    "starting_value": 95.0,
    "current_value": 105.0,
    "absolute_change": 10.0,
    "percent_change": 10.53,
    "trend": "increasing"
  }
}
```

### 11.3 Personal Records

```bash
$ hevycli stats records --limit 5 --output json
```

```json
{
  "personal_records": [
    {
      "exercise": "Bench Press (Barbell)",
      "record_type": "weight",
      "value": 110.0,
      "unit": "kg",
      "reps": 5,
      "date": "2024-12-15",
      "workout_id": "abc123"
    },
    {
      "exercise": "Squat (Barbell)",
      "record_type": "weight",
      "value": 140.0,
      "unit": "kg",
      "reps": 5,
      "date": "2024-12-10",
      "workout_id": "def456"
    }
  ]
}
```

---

## 12. Development Roadmap

### Phase 1: Foundation (MVP)

- [ ] Project scaffolding (Go modules, Cobra, Viper)
- [ ] Configuration management (init, show, set)
- [ ] API client implementation
- [ ] Authentication flow
- [ ] Basic output formatters (JSON, table, plain)

### Phase 2: Core CRUD

- [ ] Workout commands (list, get, create, update, delete)
- [ ] Routine commands (list, get, create, update)
- [ ] Exercise commands (list, get, search)
- [ ] Folder commands (list, get, create)

### Phase 3: Analytics

- [ ] Summary statistics
- [ ] Exercise progress tracking
- [ ] Personal records detection
- [ ] Computed metrics (1RM estimates, volume calculations)

### Phase 4: Interactive TUI

- [ ] Workout session mode
- [ ] Exercise search/selection
- [ ] Routine builder
- [ ] Interactive workout creation/editing

### Phase 5: Polish & Release

- [ ] Comprehensive documentation
- [ ] Shell completions (bash, zsh, fish, PowerShell)
- [ ] Homebrew formula
- [ ] Release automation (GoReleaser)
- [ ] CI/CD pipeline

### Future Considerations

- [ ] Offline mode with local SQLite cache
- [ ] Webhook integration for real-time sync
- [ ] Export to CSV/Excel
- [ ] Import from other fitness apps
- [ ] Plugin system for custom commands

---

## 13. Open Source Considerations

### 13.1 License

**MIT License** - Maximum permissiveness for community adoption

### 13.2 Repository Structure

- Clear README with quick start guide
- CONTRIBUTING.md with development setup
- CODE_OF_CONDUCT.md
- Issue and PR templates
- GitHub Actions for CI/CD

### 13.3 Documentation

- README.md - Quick start, installation, basic usage
- docs/commands.md - Complete command reference
- docs/configuration.md - All config options
- docs/ai-integration.md - Guide for AI agent developers
- docs/development.md - Contributing guide

### 13.4 Distribution

- **Homebrew** (macOS/Linux): `brew install hevycli`
- **Scoop** (Windows): `scoop install hevycli`
- **Go install**: `go install github.com/obay/hevycli@latest`
- **Direct download**: GitHub releases with pre-built binaries

---

## 14. Non-Functional Requirements

### 14.1 Performance

- Cold start: < 500ms
- API calls: < 2s (network dependent)
- TUI responsiveness: < 100ms for user input

### 14.2 Reliability

- Graceful handling of network errors
- Retry logic for transient failures
- Clear error messages with recovery suggestions

### 14.3 Security

- API key stored in config file with 600 permissions
- No credentials in command history (use config or env vars)
- Secure HTTPS-only API communication

### 14.4 Compatibility

- **OS**: macOS, Linux, Windows
- **Architecture**: amd64, arm64
- **Terminal**: Any terminal emulator with UTF-8 support
- **Go version**: 1.21+

---

## 15. Appendix

### A. Related Resources

- [Hevy Official App](https://www.hevyapp.com/)
- [Hevy API Docs](https://api.hevyapp.com/docs/)
- [Hevy Developer Settings](https://hevy.com/settings?developer)
- [go-hevy Package](https://pkg.go.dev/github.com/swrm-io/go-hevy)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Bubble Tea TUI](https://github.com/charmbracelet/bubbletea)

### B. Glossary

| Term | Definition |
|------|------------|
| **Workout** | A single training session containing exercises and sets |
| **Routine** | A template/plan for a workout that can be reused |
| **Exercise** | A specific movement (e.g., Bench Press) within a workout |
| **Set** | A single bout of an exercise (weight × reps) |
| **Exercise Template** | Hevy's database of exercise definitions |
| **Superset** | Two or more exercises performed back-to-back |
| **RPE** | Rate of Perceived Exertion (1-10 scale) |
| **1RM** | One-Rep Maximum - theoretical max weight for 1 rep |

---

*Document version: 1.0*
*Last updated: December 2024*
