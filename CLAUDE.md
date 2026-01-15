# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Claude Watcher is a TUI dashboard for tracking Claude Code session usage. It displays metrics like session counts, costs, token usage, and tool calls from data stored in a Turso (libsql) database.

## Tech Stack

- **Go 1.25+** for backend
- **BubbleTea** for TUI framework
- **Lipgloss** for TUI styling
- **sqlc** for type-safe SQL queries
- **Turso/libsql** as the database

## Build Commands

```bash
make build          # Build both binaries (dashboard + session-tracker)
make dashboard      # Build only the dashboard
make session-tracker # Build only the session tracker
make run            # Build and run the dashboard
make test           # Run tests
make sqlc           # Generate sqlc code
make fmt            # Format code
make clean          # Remove build artifacts
```

## Environment Variables

Dashboard:
- `TURSO_DATABASE_URL_CLAUDE_WATCHER` - libsql connection URL
- `TURSO_AUTH_TOKEN_CLAUDE_WATCHER` - Turso auth token

Session Tracker:
- `CLAUDE_WATCHER_DATABASE_URL` - libsql connection URL
- `CLAUDE_WATCHER_AUTH_TOKEN` - Turso auth token

## Architecture

The codebase follows hexagonal architecture (ports & adapters) with vertical slices:

```
cmd/
├── dashboard/          # TUI dashboard entry point
└── session-tracker/    # Hook handler entry point

internal/
├── analytics/          # Analytics domain
│   ├── model.go        # Domain models (Metrics, Session summaries)
│   ├── ports.go        # Repository interface
│   ├── service.go      # Query use cases
│   ├── inbound/tui/    # TUI screens (overview, sessions, costs, detail)
│   └── outbound/turso/ # Database repository implementation
├── app/tui/            # Main TUI app shell and navigation
├── database/           # Database infrastructure
│   ├── migrations/     # SQL migration files
│   ├── queries/        # sqlc query definitions (.sql)
│   └── sqlc/           # Generated sqlc code (DO NOT EDIT)
├── limits/             # Usage limits domain
│   ├── model.go        # LimitEvent, LimitType
│   ├── ports.go        # Repository interface
│   ├── service.go      # Limit tracking logic
│   └── outbound/       # Turso repository
├── pkg/tui/            # Shared TUI components
│   ├── components/     # Reusable widgets (scale, selector, help)
│   └── theme/          # Lipgloss styles and colors
├── pricing/            # Pricing domain (pure calculation)
│   ├── model.go        # ModelPricing, TokenUsage
│   └── service.go      # Cost calculator
├── tracker/            # Session tracker domain
│   ├── domain/         # Core models and interfaces
│   └── adapters/       # Repository, prompter, parser implementations
└── transcript/         # Transcript parser
    ├── model.go        # ParsedTranscript, SessionStatistics
    ├── ports.go        # Parser interface
    └── parser.go       # JSONL transcript parser
```

## Key Patterns

### Hexagonal Architecture
- **Domain** defines ports (interfaces) in `ports.go`
- **Inbound adapters** (TUI screens) call domain services
- **Outbound adapters** (Turso repositories) implement domain interfaces

### Adding New Features
1. Create domain models in `model.go`
2. Define repository interface in `ports.go`
3. Implement business logic in `service.go`
4. Create TUI screens in `inbound/tui/`
5. Implement repository in `outbound/turso/`

### Adding SQL Queries
1. Add query to `internal/database/queries/*.sql` with sqlc annotation
2. Run `make sqlc` to regenerate code in `internal/database/sqlc/`

### TUI Development
- Use shared styles from `internal/pkg/tui/theme/`
- Reusable components in `internal/pkg/tui/components/`
- Screens implement `tea.Model` interface from BubbleTea
