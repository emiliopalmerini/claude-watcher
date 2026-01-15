# claude-watcher

A TUI dashboard for tracking Claude Code session usage.

## Features

- **TUI Dashboard**: Interactive terminal dashboard showing session metrics, costs, and usage history
- **Session Tracking**: Automatically captures session data via Claude Code hooks
- **Quality Feedback**: Rate sessions with prompt specificity, task completion, and code confidence
- **Limit Tracking**: Monitors daily/weekly/monthly usage limits from transcripts
- **Cost Analysis**: Tracks token usage and estimated costs across models

## Prerequisites

- Go 1.25+
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI with libsql support
- A Turso database

## Installation

### Install golang-migrate with libsql support

```bash
go install -tags 'libsql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Build from source

```bash
make build
```

This builds two binaries:
- `claude-watcher` - TUI dashboard
- `session-tracker` - Hook handler for session tracking

### Using Nix

```bash
nix build
```

## Environment Variables

### For Dashboard

```bash
export TURSO_DATABASE_URL_CLAUDE_WATCHER="libsql://your-database.turso.io"
export TURSO_AUTH_TOKEN_CLAUDE_WATCHER="your-auth-token"
```

### For Session Tracker

```bash
export CLAUDE_WATCHER_DATABASE_URL="libsql://your-database.turso.io"
export CLAUDE_WATCHER_AUTH_TOKEN="your-auth-token"
```

## Database Setup

### Run migrations

```bash
# Apply all migrations
migrate -database "libsql://your-database.turso.io?authToken=your-token" \
        -path internal/database/migrations up
```

### Rollback migrations

```bash
# Rollback last migration
migrate -database "libsql://your-database.turso.io?authToken=your-token" \
        -path internal/database/migrations down 1
```

## Usage

### Running the Dashboard

```bash
./claude-watcher
# or
make run
```

Navigate with:
- `Tab` / `Shift+Tab` - Switch between screens
- `j/k` or arrows - Navigate lists
- `Enter` - View session details
- `q` - Quit

### Setting up Session Tracking

Add the session tracker to your Claude Code hooks in `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PostSessionHook": [
      {
        "type": "command",
        "command": "/path/to/session-tracker"
      }
    ]
  }
}
```

The tracker will:
1. Parse the session transcript
2. Prompt for quality feedback (TUI)
3. Save session data to the database

## Development

```bash
# Generate sqlc code
make sqlc

# Run tests
make test

# Format code
make fmt

# Build everything
make build

# Clean build artifacts
make clean
```

## Architecture

```
cmd/
├── dashboard/          # TUI dashboard entry point
└── session-tracker/    # Hook handler entry point

internal/
├── analytics/          # Analytics domain (metrics, queries)
│   ├── inbound/tui/    # TUI screens (overview, sessions, costs)
│   └── outbound/turso/ # Database repository
├── app/tui/            # Main TUI app shell
├── database/           # Database infrastructure
│   ├── migrations/     # SQL migrations
│   ├── queries/        # sqlc query definitions
│   └── sqlc/           # Generated sqlc code
├── limits/             # Limits domain (usage tracking)
├── pkg/tui/            # Shared TUI components
│   ├── components/     # Reusable widgets
│   └── theme/          # Styling
├── pricing/            # Pricing domain (cost calculation)
├── tracker/            # Session tracker domain
│   ├── adapters/       # Repository, prompter, parser
│   └── domain/         # Core session tracking logic
└── transcript/         # Transcript parser
```
