package live

import (
	"context"

	"claude-watcher/internal/database/sqlc"
)

// Repository defines the data access interface for live sessions.
type Repository interface {
	GetLiveSessions(ctx context.Context) ([]sqlc.GetLiveSessionsRow, error)
	GetRecentSessions(ctx context.Context) ([]sqlc.GetRecentSessionsRow, error)
}

// SQLCRepository implements Repository using sqlc.Queries.
type SQLCRepository struct {
	queries *sqlc.Queries
}

// NewSQLCRepository creates a new SQLCRepository.
func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) GetLiveSessions(ctx context.Context) ([]sqlc.GetLiveSessionsRow, error) {
	return r.queries.GetLiveSessions(ctx)
}

func (r *SQLCRepository) GetRecentSessions(ctx context.Context) ([]sqlc.GetRecentSessionsRow, error) {
	return r.queries.GetRecentSessions(ctx)
}
