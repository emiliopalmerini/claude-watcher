package dashboard

import (
	"context"

	"claude-watcher/internal/database/sqlc"
)

// Repository defines the data access interface for dashboard metrics.
type Repository interface {
	GetDashboardMetrics(ctx context.Context) (sqlc.GetDashboardMetricsRow, error)
	GetTodayMetrics(ctx context.Context) (sqlc.GetTodayMetricsRow, error)
	GetWeekMetrics(ctx context.Context) (sqlc.GetWeekMetricsRow, error)
}

// SQLCRepository implements Repository using sqlc.Queries.
type SQLCRepository struct {
	queries *sqlc.Queries
}

// NewSQLCRepository creates a new SQLCRepository.
func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) GetDashboardMetrics(ctx context.Context) (sqlc.GetDashboardMetricsRow, error) {
	return r.queries.GetDashboardMetrics(ctx)
}

func (r *SQLCRepository) GetTodayMetrics(ctx context.Context) (sqlc.GetTodayMetricsRow, error) {
	return r.queries.GetTodayMetrics(ctx)
}

func (r *SQLCRepository) GetWeekMetrics(ctx context.Context) (sqlc.GetWeekMetricsRow, error) {
	return r.queries.GetWeekMetrics(ctx)
}
