package api

import (
	"context"
	"database/sql"

	"claude-watcher/internal/database/sqlc"
)

// MockRepository is a mock implementation of Repository for testing.
type MockRepository struct {
	GetDailyMetricsFunc          func(ctx context.Context, days sql.NullString) ([]sqlc.GetDailyMetricsRow, error)
	GetModelDistributionFunc     func(ctx context.Context, hours sql.NullString) ([]sqlc.GetModelDistributionRow, error)
	GetHourOfDayDistributionFunc func(ctx context.Context, hours sql.NullString) ([]sqlc.GetHourOfDayDistributionRow, error)
}

func (m *MockRepository) GetDailyMetrics(ctx context.Context, days sql.NullString) ([]sqlc.GetDailyMetricsRow, error) {
	if m.GetDailyMetricsFunc != nil {
		return m.GetDailyMetricsFunc(ctx, days)
	}
	return []sqlc.GetDailyMetricsRow{}, nil
}

func (m *MockRepository) GetModelDistribution(ctx context.Context, hours sql.NullString) ([]sqlc.GetModelDistributionRow, error) {
	if m.GetModelDistributionFunc != nil {
		return m.GetModelDistributionFunc(ctx, hours)
	}
	return []sqlc.GetModelDistributionRow{}, nil
}

func (m *MockRepository) GetHourOfDayDistribution(ctx context.Context, hours sql.NullString) ([]sqlc.GetHourOfDayDistributionRow, error) {
	if m.GetHourOfDayDistributionFunc != nil {
		return m.GetHourOfDayDistributionFunc(ctx, hours)
	}
	return []sqlc.GetHourOfDayDistributionRow{}, nil
}
