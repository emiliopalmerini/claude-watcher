package api

import (
	"context"
	"database/sql"

	"claude-watcher/internal/database/sqlc"
)

// Repository defines the data access interface for chart API.
type Repository interface {
	GetDailyMetrics(ctx context.Context, days sql.NullString) ([]sqlc.GetDailyMetricsRow, error)
	GetModelDistribution(ctx context.Context, hours sql.NullString) ([]sqlc.GetModelDistributionRow, error)
	GetHourOfDayDistribution(ctx context.Context, hours sql.NullString) ([]sqlc.GetHourOfDayDistributionRow, error)
}

// SQLCRepository implements Repository using sqlc.Queries.
type SQLCRepository struct {
	queries *sqlc.Queries
}

// NewSQLCRepository creates a new SQLCRepository.
func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) GetDailyMetrics(ctx context.Context, days sql.NullString) ([]sqlc.GetDailyMetricsRow, error) {
	return r.queries.GetDailyMetrics(ctx, days)
}

func (r *SQLCRepository) GetModelDistribution(ctx context.Context, hours sql.NullString) ([]sqlc.GetModelDistributionRow, error) {
	return r.queries.GetModelDistribution(ctx, hours)
}

func (r *SQLCRepository) GetHourOfDayDistribution(ctx context.Context, hours sql.NullString) ([]sqlc.GetHourOfDayDistributionRow, error) {
	return r.queries.GetHourOfDayDistribution(ctx, hours)
}
