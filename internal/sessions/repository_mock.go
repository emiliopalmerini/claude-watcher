package sessions

import (
	"context"

	"claude-watcher/internal/database/sqlc"
)

// MockRepository is a mock implementation of Repository for testing.
type MockRepository struct {
	ListSessionsFunc  func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error)
	CountSessionsFunc func(ctx context.Context) (int64, error)
}

func (m *MockRepository) ListSessions(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
	if m.ListSessionsFunc != nil {
		return m.ListSessionsFunc(ctx, params)
	}
	return []sqlc.ListSessionsRow{}, nil
}

func (m *MockRepository) CountSessions(ctx context.Context) (int64, error) {
	if m.CountSessionsFunc != nil {
		return m.CountSessionsFunc(ctx)
	}
	return 0, nil
}
