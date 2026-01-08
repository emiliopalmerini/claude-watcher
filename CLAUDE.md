# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Claude Watcher is a dashboard for tracking Claude Code session usage. It displays metrics like session counts, costs, token usage, and tool calls from data stored in a Turso (libsql) database.

## Tech Stack

- **Go 1.25+** with Chi router for HTTP handling
- **templ** for type-safe HTML templates
- **sqlc** for type-safe SQL queries
- **HTMX** for dynamic UI updates
- **Turso/libsql** as the database

## Build Commands

```bash
make build          # Build the application (runs vet, sqlc, templ first)
make run            # Build and run
make test           # Run tests
make generate       # Generate sqlc and templ code
make sqlc           # Generate sqlc code only
make templ          # Generate templ code only
make fmt            # Format code
```

## Database Migrations

Requires golang-migrate with libsql support:
```bash
go install -tags 'libsql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Environment variables required:
- `TURSO_DATABASE_URL` - libsql connection URL
- `TURSO_AUTH_TOKEN` - Turso auth token

## Architecture

The codebase follows a vertical slice architecture where each feature is self-contained:

```
internal/
├── app/              # Application bootstrap and config
├── database/
│   ├── migrations/   # SQL migration files
│   ├── queries/      # sqlc query definitions (.sql)
│   └── sqlc/         # Generated sqlc code (DO NOT EDIT)
├── dashboard/        # Dashboard feature (handler, routes, templates, models)
├── sessions/         # Sessions list feature
├── session_detail/   # Session detail feature
├── server/           # HTTP server setup and routing
└── shared/
    ├── middleware/   # HTMX middleware
    └── templates/    # Shared templ components (layout, nav, cards, pagination)
```

Each feature module contains:
- `handler.go` - HTTP handlers
- `routes.go` - Route registration
- `models.go` - Data structures
- `*.templ` - Template files (generate `*_templ.go`)

## Adding New SQL Queries

1. Add query to `internal/database/queries/sessions.sql` with sqlc annotation
2. Run `make sqlc` to regenerate `internal/database/sqlc/sessions.sql.go`

## Adding New Templates

1. Create/edit `.templ` files in the feature directory
2. Run `make templ` to generate corresponding `*_templ.go` files
