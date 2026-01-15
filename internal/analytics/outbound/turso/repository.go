package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"claude-watcher/internal/analytics"
	"claude-watcher/internal/database/sqlc"
)

// Repository implements analytics.Repository using sqlc queries
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new Turso repository for analytics
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

// GetOverviewMetrics retrieves aggregate metrics for the dashboard overview
func (r *Repository) GetOverviewMetrics(ctx context.Context) (analytics.OverviewMetrics, error) {
	// Get week metrics (last 7 days) for the overview
	weekRow, err := r.queries.GetWeekMetrics(ctx)
	if err != nil {
		return analytics.OverviewMetrics{}, fmt.Errorf("get week metrics: %w", err)
	}

	// Get limit hits count (non-fatal if table doesn't exist)
	var limitHits int64
	limitHits, _ = r.queries.CountLimitEvents(ctx)

	// Get last limit hit timestamp (non-fatal if table doesn't exist)
	var lastLimitHit *time.Time
	lastLimitStr, err := r.queries.GetLastLimitEvent(ctx)
	if err == nil && lastLimitStr != "" {
		if t, parseErr := time.Parse(time.RFC3339, lastLimitStr); parseErr == nil {
			lastLimitHit = &t
		}
	}

	return analytics.OverviewMetrics{
		TotalSessions: int(weekRow.SessionsWeek),
		TotalCost:     toFloat64(weekRow.CostWeek),
		Tokens: analytics.TokenSummary{
			Input:    toInt64(weekRow.InputTokensWeek),
			Output:   toInt64(weekRow.OutputTokensWeek),
			Thinking: toInt64(weekRow.ThinkingTokensWeek),
		},
		LimitHits:    int(limitHits),
		LastLimitHit: lastLimitHit,
	}, nil
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case nil:
		return 0
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case nil:
		return 0
	default:
		return 0
	}
}
