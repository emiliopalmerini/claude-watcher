package sessions

import (
	"context"

	"claude-watcher/internal/database/sqlc"
)

// Repository defines the data access interface for sessions list.
type Repository interface {
	ListSessions(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error)
	CountSessions(ctx context.Context) (int64, error)
}

// SQLCRepository implements Repository using sqlc.Queries.
type SQLCRepository struct {
	queries *sqlc.Queries
}

// NewSQLCRepository creates a new SQLCRepository.
func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) ListSessions(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
	return r.queries.ListSessions(ctx, params)
}

func (r *SQLCRepository) CountSessions(ctx context.Context) (int64, error) {
	return r.queries.CountSessions(ctx)
}
