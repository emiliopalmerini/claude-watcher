package outbound

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"claude-watcher/internal/database/sqlc"
	"claude-watcher/internal/limits"
)

// TursoRepository implements limits.Repository using sqlc queries
type TursoRepository struct {
	queries *sqlc.Queries
}

// NewTursoRepository creates a new Turso repository for limits
func NewTursoRepository(db *sql.DB) *TursoRepository {
	return &TursoRepository{
		queries: sqlc.New(db),
	}
}

// Save persists a limit event to the database
func (r *TursoRepository) Save(event limits.LimitEvent) error {
	ctx := context.Background()

	var resetTime sql.NullString
	if event.ResetTime != nil {
		resetTime = sql.NullString{
			String: event.ResetTime.Format(time.RFC3339),
			Valid:  true,
		}
	}

	params := sqlc.InsertLimitEventParams{
		Timestamp:      event.Timestamp.Format(time.RFC3339),
		LimitType:      string(event.LimitType),
		ResetTime:      resetTime,
		SessionsCount:  sql.NullInt64{Int64: int64(event.SessionsCount), Valid: true},
		InputTokens:    sql.NullInt64{Int64: int64(event.InputTokens), Valid: true},
		OutputTokens:   sql.NullInt64{Int64: int64(event.OutputTokens), Valid: true},
		ThinkingTokens: sql.NullInt64{Int64: int64(event.ThinkingTokens), Valid: true},
		TotalCostUsd:   sql.NullFloat64{Float64: event.TotalCostUSD, Valid: true},
	}

	if err := r.queries.InsertLimitEvent(ctx, params); err != nil {
		return fmt.Errorf("insert limit event: %w", err)
	}

	return nil
}

// GetUsageSinceLastLimit returns aggregate usage metrics since the last limit event
func (r *TursoRepository) GetUsageSinceLastLimit() (limits.UsageSummary, error) {
	ctx := context.Background()

	row, err := r.queries.GetUsageSinceLastLimit(ctx)
	if err != nil {
		return limits.UsageSummary{}, fmt.Errorf("get usage since last limit: %w", err)
	}

	return limits.UsageSummary{
		SessionsCount:  int(row.SessionsCount),
		InputTokens:    toInt(row.InputTokens),
		OutputTokens:   toInt(row.OutputTokens),
		ThinkingTokens: toInt(row.ThinkingTokens),
		TotalCostUSD:   toFloat64(row.TotalCostUsd),
	}, nil
}

// GetLastLimitTimestamp returns the timestamp of the most recent limit event
func (r *TursoRepository) GetLastLimitTimestamp() (*time.Time, error) {
	ctx := context.Background()

	ts, err := r.queries.GetLastLimitEvent(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get last limit event: %w", err)
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil, fmt.Errorf("parse timestamp: %w", err)
	}

	return &t, nil
}

// ListRecent returns the most recent limit events within the given number of days
func (r *TursoRepository) ListRecent(days int) ([]limits.LimitEvent, error) {
	ctx := context.Background()

	daysStr := sql.NullString{String: fmt.Sprintf("-%d", days), Valid: true}
	rows, err := r.queries.GetRecentLimitEvents(ctx, daysStr)
	if err != nil {
		return nil, fmt.Errorf("get recent limit events: %w", err)
	}

	return convertLimitEvents(rows), nil
}

// ListByType returns limit events of a specific type
func (r *TursoRepository) ListByType(limitType limits.LimitType, limit int) ([]limits.LimitEvent, error) {
	ctx := context.Background()

	params := sqlc.GetLimitEventsByTypeParams{
		LimitType: string(limitType),
		Limit:     int64(limit),
	}

	rows, err := r.queries.GetLimitEventsByType(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get limit events by type: %w", err)
	}

	return convertLimitEvents(rows), nil
}

func convertLimitEvents(rows []sqlc.LimitEvent) []limits.LimitEvent {
	events := make([]limits.LimitEvent, 0, len(rows))
	for _, row := range rows {
		event := limits.LimitEvent{
			ID:             row.ID,
			LimitType:      limits.LimitType(row.LimitType),
			SessionsCount:  int(row.SessionsCount.Int64),
			InputTokens:    int(row.InputTokens.Int64),
			OutputTokens:   int(row.OutputTokens.Int64),
			ThinkingTokens: int(row.ThinkingTokens.Int64),
			TotalCostUSD:   row.TotalCostUsd.Float64,
		}

		if t, err := time.Parse(time.RFC3339, row.Timestamp); err == nil {
			event.Timestamp = t
		}

		if row.ResetTime.Valid {
			if t, err := time.Parse(time.RFC3339, row.ResetTime.String); err == nil {
				event.ResetTime = &t
			}
		}

		events = append(events, event)
	}
	return events
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int64:
		return int(val)
	case float64:
		return int(val)
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
