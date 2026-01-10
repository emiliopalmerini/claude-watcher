package session_detail

import (
	"context"

	"claude-watcher/internal/database/sqlc"
)

// MockRepository is a mock implementation of Repository for testing.
type MockRepository struct {
	GetSessionByIDFunc func(ctx context.Context, sessionID string) (sqlc.Session, error)
}

func (m *MockRepository) GetSessionByID(ctx context.Context, sessionID string) (sqlc.Session, error) {
	if m.GetSessionByIDFunc != nil {
		return m.GetSessionByIDFunc(ctx, sessionID)
	}
	return sqlc.Session{}, nil
}
