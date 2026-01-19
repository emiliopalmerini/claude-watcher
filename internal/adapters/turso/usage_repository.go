package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type UsageLimitsRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewUsageLimitsRepository(db *sql.DB) *UsageLimitsRepository {
	return &UsageLimitsRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *UsageLimitsRepository) Upsert(ctx context.Context, limit *domain.UsageLimit) error {
	enabled := int64(0)
	if limit.Enabled {
		enabled = 1
	}
	return r.queries.CreateUsageLimit(ctx, sqlc.CreateUsageLimitParams{
		ID:            limit.ID,
		LimitValue:    limit.LimitValue,
		WarnThreshold: sql.NullFloat64{Float64: limit.WarnThreshold, Valid: true},
		Enabled:       enabled,
	})
}

func (r *UsageLimitsRepository) Get(ctx context.Context, id string) (*domain.UsageLimit, error) {
	row, err := r.queries.GetUsageLimit(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get usage limit: %w", err)
	}
	return limitFromRow(row), nil
}

func (r *UsageLimitsRepository) List(ctx context.Context) ([]*domain.UsageLimit, error) {
	rows, err := r.queries.ListUsageLimits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list usage limits: %w", err)
	}
	limits := make([]*domain.UsageLimit, len(rows))
	for i, row := range rows {
		limits[i] = limitFromRow(row)
	}
	return limits, nil
}

func (r *UsageLimitsRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteUsageLimit(ctx, id)
}

func limitFromRow(row sqlc.UsageLimit) *domain.UsageLimit {
	limit := &domain.UsageLimit{
		ID:            row.ID,
		LimitValue:    row.LimitValue,
		WarnThreshold: 0.8,
		Enabled:       row.Enabled == 1,
		CreatedAt:     util.ParseTimeRFC3339(row.CreatedAt),
		UpdatedAt:     util.ParseTimeRFC3339(row.UpdatedAt),
	}
	if row.WarnThreshold.Valid {
		limit.WarnThreshold = row.WarnThreshold.Float64
	}
	return limit
}

type PlanConfigRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewPlanConfigRepository(db *sql.DB) *PlanConfigRepository {
	return &PlanConfigRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *PlanConfigRepository) Upsert(ctx context.Context, config *domain.PlanConfig) error {
	var learnedLimit sql.NullFloat64
	var learnedAt sql.NullString

	if config.LearnedTokenLimit != nil {
		learnedLimit = sql.NullFloat64{Float64: *config.LearnedTokenLimit, Valid: true}
	}
	if config.LearnedAt != nil {
		learnedAt = sql.NullString{String: config.LearnedAt.Format(time.RFC3339), Valid: true}
	}

	return r.queries.UpsertPlanConfig(ctx, sqlc.UpsertPlanConfigParams{
		PlanType:          config.PlanType,
		WindowHours:       int64(config.WindowHours),
		LearnedTokenLimit: learnedLimit,
		LearnedAt:         learnedAt,
	})
}

func (r *PlanConfigRepository) Get(ctx context.Context) (*domain.PlanConfig, error) {
	row, err := r.queries.GetPlanConfig(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get plan config: %w", err)
	}
	return planConfigFromRow(row), nil
}

func (r *PlanConfigRepository) UpdateLearnedLimit(ctx context.Context, limit float64) error {
	return r.queries.UpdateLearnedLimit(ctx, sql.NullFloat64{Float64: limit, Valid: true})
}

func (r *PlanConfigRepository) GetRollingWindowSummary(ctx context.Context, windowHours int) (*domain.UsageSummary, error) {
	hoursParam := fmt.Sprintf("-%d", windowHours)
	row, err := r.queries.GetRollingWindowUsage(ctx, sql.NullString{String: hoursParam, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get rolling window usage: %w", err)
	}
	return &domain.UsageSummary{
		TotalTokens: row.TotalTokens,
		TotalCost:   row.TotalCost,
	}, nil
}

func planConfigFromRow(row sqlc.PlanConfig) *domain.PlanConfig {
	config := &domain.PlanConfig{
		PlanType:    row.PlanType,
		WindowHours: int(row.WindowHours),
		CreatedAt:   util.ParseTimeRFC3339(row.CreatedAt),
		UpdatedAt:   util.ParseTimeRFC3339(row.UpdatedAt),
	}
	if row.LearnedTokenLimit.Valid {
		config.LearnedTokenLimit = &row.LearnedTokenLimit.Float64
	}
	if row.LearnedAt.Valid {
		t := util.ParseTimeSQLite(row.LearnedAt.String)
		config.LearnedAt = &t
	}
	if row.WindowStartTime.Valid {
		t := util.ParseTimeSQLite(row.WindowStartTime.String)
		config.WindowStartTime = &t
	}
	// Weekly fields
	if row.WeeklyLearnedTokenLimit.Valid {
		config.WeeklyLearnedTokenLimit = &row.WeeklyLearnedTokenLimit.Float64
	}
	if row.WeeklyLearnedAt.Valid {
		t := util.ParseTimeSQLite(row.WeeklyLearnedAt.String)
		config.WeeklyLearnedAt = &t
	}
	if row.WeeklyWindowStartTime.Valid {
		t := util.ParseTimeSQLite(row.WeeklyWindowStartTime.String)
		config.WeeklyWindowStartTime = &t
	}
	return config
}

func (r *PlanConfigRepository) UpdateWindowStartTime(ctx context.Context, t time.Time) error {
	return r.queries.UpdateWindowStartTime(ctx, sql.NullString{String: t.Format(time.RFC3339), Valid: true})
}

func (r *PlanConfigRepository) ResetWindowIfExpired(ctx context.Context, sessionStartTime time.Time) (bool, error) {
	config, err := r.Get(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get plan config: %w", err)
	}
	if config == nil {
		return false, nil
	}

	windowHours := config.WindowHours
	return r.resetWindowIfExpired(ctx, sessionStartTime, config.WindowStartTime, windowHours, r.UpdateWindowStartTime)
}

// Weekly window methods

func (r *PlanConfigRepository) GetWeeklyWindowSummary(ctx context.Context) (*domain.UsageSummary, error) {
	row, err := r.queries.GetWeeklyWindowUsage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly window usage: %w", err)
	}
	return &domain.UsageSummary{
		TotalTokens: row.TotalTokens,
		TotalCost:   row.TotalCost,
	}, nil
}

func (r *PlanConfigRepository) UpdateWeeklyWindowStartTime(ctx context.Context, t time.Time) error {
	return r.queries.UpdateWeeklyWindowStartTime(ctx, sql.NullString{String: t.Format(time.RFC3339), Valid: true})
}

func (r *PlanConfigRepository) UpdateWeeklyLearnedLimit(ctx context.Context, limit float64) error {
	return r.queries.UpdateWeeklyLearnedLimit(ctx, sql.NullFloat64{Float64: limit, Valid: true})
}

func (r *PlanConfigRepository) ResetWeeklyWindowIfExpired(ctx context.Context, sessionStartTime time.Time) (bool, error) {
	config, err := r.Get(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get plan config: %w", err)
	}
	if config == nil {
		return false, nil
	}

	return r.resetWindowIfExpired(ctx, sessionStartTime, config.WeeklyWindowStartTime, domain.WeeklyWindowHours, r.UpdateWeeklyWindowStartTime)
}

// resetWindowIfExpired is a helper that handles the common logic for resetting time windows.
// It resets if windowStartTime is nil OR sessionStartTime is after the window expired.
func (r *PlanConfigRepository) resetWindowIfExpired(
	ctx context.Context,
	sessionStartTime time.Time,
	windowStartTime *time.Time,
	windowHours int,
	updateFn func(context.Context, time.Time) error,
) (bool, error) {
	windowDuration := time.Duration(windowHours) * time.Hour

	if windowStartTime == nil || sessionStartTime.After(windowStartTime.Add(windowDuration)) {
		if err := updateFn(ctx, sessionStartTime); err != nil {
			return false, fmt.Errorf("failed to update window start time: %w", err)
		}
		return true, nil
	}

	return false, nil
}
