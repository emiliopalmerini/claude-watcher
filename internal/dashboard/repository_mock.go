package dashboard

import (
	"context"

	"claude-watcher/internal/database/sqlc"
)

// MockRepository is a mock implementation of Repository for testing.
type MockRepository struct {
	GetDashboardMetricsFunc func(ctx context.Context) (sqlc.GetDashboardMetricsRow, error)
	GetTodayMetricsFunc     func(ctx context.Context) (sqlc.GetTodayMetricsRow, error)
	GetWeekMetricsFunc      func(ctx context.Context) (sqlc.GetWeekMetricsRow, error)
}

func (m *MockRepository) GetDashboardMetrics(ctx context.Context) (sqlc.GetDashboardMetricsRow, error) {
	if m.GetDashboardMetricsFunc != nil {
		return m.GetDashboardMetricsFunc(ctx)
	}
	return sqlc.GetDashboardMetricsRow{}, nil
}

func (m *MockRepository) GetTodayMetrics(ctx context.Context) (sqlc.GetTodayMetricsRow, error) {
	if m.GetTodayMetricsFunc != nil {
		return m.GetTodayMetricsFunc(ctx)
	}
	return sqlc.GetTodayMetricsRow{}, nil
}

func (m *MockRepository) GetWeekMetrics(ctx context.Context) (sqlc.GetWeekMetricsRow, error) {
	if m.GetWeekMetricsFunc != nil {
		return m.GetWeekMetricsFunc(ctx)
	}
	return sqlc.GetWeekMetricsRow{}, nil
}
