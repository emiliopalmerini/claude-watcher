package session_detail

import (
	"context"

	"claude-watcher/internal/database/sqlc"
)

// Repository defines the data access interface for session details.
type Repository interface {
	GetSessionByID(ctx context.Context, sessionID string) (sqlc.Session, error)
}

// SQLCRepository implements Repository using sqlc.Queries.
type SQLCRepository struct {
	queries *sqlc.Queries
}

// NewSQLCRepository creates a new SQLCRepository.
func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) GetSessionByID(ctx context.Context, sessionID string) (sqlc.Session, error) {
	return r.queries.GetSessionByID(ctx, sessionID)
}
