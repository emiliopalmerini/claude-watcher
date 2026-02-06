package web

import (
	"net/http"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get stats
	startDate := time.Unix(0, 0).Format(time.RFC3339) // All time
	statsRow, _ := queries.GetAggregateStats(ctx, startDate)

	stats := templates.DashboardStats{
		SessionCount: statsRow.SessionCount,
		TotalTokens:  util.ToInt64(statsRow.TotalTokenInput) + util.ToInt64(statsRow.TotalTokenOutput),
		TotalCost:    util.ToFloat64(statsRow.TotalCostUsd),
		TotalTurns:   util.ToInt64(statsRow.TotalTurns),
		TokenInput:   util.ToInt64(statsRow.TotalTokenInput),
		TokenOutput:  util.ToInt64(statsRow.TotalTokenOutput),
		CacheRead:    util.ToInt64(statsRow.TotalTokenCacheRead),
		CacheWrite:   util.ToInt64(statsRow.TotalTokenCacheWrite),
		TotalErrors:  util.ToInt64(statsRow.TotalErrors),
	}

	// Get usage limit stats
	if planConfig, err := s.planConfigRepo.Get(ctx); err == nil && planConfig != nil {
		usageStats := &templates.UsageLimitStats{
			PlanType:    planConfig.PlanType,
			WindowHours: planConfig.WindowHours,
		}

		// Get the 5-hour token limit (learned or estimated)
		if planConfig.LearnedTokenLimit != nil {
			usageStats.TokenLimit = *planConfig.LearnedTokenLimit
			usageStats.IsLearned = true
		} else if preset, ok := domain.PlanPresets[planConfig.PlanType]; ok {
			usageStats.TokenLimit = preset.TokenEstimate
		}

		// Get 5-hour rolling window usage
		if summary, err := s.planConfigRepo.GetRollingWindowSummary(ctx, planConfig.WindowHours); err == nil {
			usageStats.TokensUsed = summary.TotalTokens

			// Calculate percentage
			if usageStats.TokenLimit > 0 {
				usageStats.UsagePercent = (summary.TotalTokens / usageStats.TokenLimit) * 100
			}

			// Determine status
			usageStats.Status = domain.GetStatusFromPercent(usageStats.UsagePercent)

			// Approximate minutes left (rolling window refreshes continuously)
			usageStats.MinutesLeft = planConfig.WindowHours * 60
		}

		// Get the weekly token limit (learned or estimated)
		if planConfig.WeeklyLearnedTokenLimit != nil {
			usageStats.WeeklyTokenLimit = *planConfig.WeeklyLearnedTokenLimit
			usageStats.WeeklyIsLearned = true
		} else if preset, ok := domain.WeeklyPlanPresets[planConfig.PlanType]; ok {
			usageStats.WeeklyTokenLimit = preset.TokenEstimate
		}

		// Get weekly rolling window usage
		if weeklySummary, err := s.planConfigRepo.GetWeeklyWindowSummary(ctx); err == nil {
			usageStats.WeeklyTokensUsed = weeklySummary.TotalTokens

			// Calculate percentage
			if usageStats.WeeklyTokenLimit > 0 {
				usageStats.WeeklyUsagePercent = (weeklySummary.TotalTokens / usageStats.WeeklyTokenLimit) * 100
			}

			// Determine status
			usageStats.WeeklyStatus = domain.GetStatusFromPercent(usageStats.WeeklyUsagePercent)
		}

		stats.UsageStats = usageStats
	}

	// Get active experiment
	activeExp, _ := queries.GetActiveExperiment(ctx)
	if activeExp.Name != "" {
		stats.ActiveExperiment = activeExp.Name
	}

	// Get default model
	defaultModel, _ := queries.GetDefaultModelPricing(ctx)
	if defaultModel.DisplayName != "" {
		stats.DefaultModel = defaultModel.DisplayName
	}

	// Get top tools
	tools, _ := queries.GetTopToolsUsage(ctx, sqlc.GetTopToolsUsageParams{
		CreatedAt: startDate,
		Limit:     5,
	})

	topTools := make([]templates.ToolUsage, 0, len(tools))
	for _, t := range tools {
		if t.TotalInvocations.Valid {
			topTools = append(topTools, templates.ToolUsage{
				Name:  t.ToolName,
				Count: int64(t.TotalInvocations.Float64),
			})
		}
	}
	stats.TopTools = topTools

	// Get recent sessions (with metrics in single query)
	sessions, _ := queries.ListSessionsWithMetrics(ctx, 5)
	recentSessions := make([]templates.SessionSummary, 0, len(sessions))
	for _, sess := range sessions {
		summary := templates.SessionSummary{
			ID:         sess.ID,
			CreatedAt:  sess.CreatedAt,
			ExitReason: sess.ExitReason,
			Turns:      sess.TurnCount,
			Tokens:     sess.TotalTokens,
		}
		if sess.CostEstimateUsd.Valid {
			summary.Cost = sess.CostEstimateUsd.Float64
		}
		recentSessions = append(recentSessions, summary)
	}
	stats.RecentSessions = recentSessions

	// Get overall quality stats
	qualityStats, err := queries.GetOverallQualityStats(ctx)
	if err == nil && qualityStats.ReviewedCount > 0 {
		stats.ReviewedCount = qualityStats.ReviewedCount
		if qualityStats.AvgOverallRating.Valid {
			avg := qualityStats.AvgOverallRating.Float64
			stats.AvgOverall = &avg
		}
		stats.SuccessRate = calculateSuccessRate(qualityStats.SuccessCount, qualityStats.FailureCount)
	}

	templates.Dashboard(stats).Render(ctx, w)
}
