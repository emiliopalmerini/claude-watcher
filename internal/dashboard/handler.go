package dashboard

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	apperrors "claude-watcher/internal/shared/errors"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.repo.GetDashboardMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	today, err := h.repo.GetTodayMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	week, err := h.repo.GetWeekMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	cacheMetrics, err := h.repo.GetCacheMetrics(ctx, "-7")
	if err != nil {
		log.Printf("error fetching cache metrics: %v", err)
	}

	topProject, err := h.repo.GetTopProject(ctx)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error fetching top project: %v", err)
	}

	efficiencyMetrics, err := h.repo.GetEfficiencyMetrics(ctx, "-7")
	if err != nil {
		log.Printf("error fetching efficiency metrics: %v", err)
	}

	toolsBreakdown, err := h.repo.GetToolsBreakdownAll(ctx, "-7")
	if err != nil {
		log.Printf("error fetching tools breakdown: %v", err)
	}

	topTool := TopTool(toolsBreakdown)

	var cacheHitRate float64
	if totalTokens := toInt64(cacheMetrics.TotalTokens); totalTokens > 0 {
		cacheHitRate = float64(toInt64(cacheMetrics.CacheRead)) / float64(totalTokens) * 100
	}

	// Fetch burn rate and window usage for plan status
	burnRateMetrics, err := h.repo.GetBurnRateMetrics(ctx)
	if err != nil {
		log.Printf("error fetching burn rate: %v", err)
	}

	windowUsage, err := h.repo.GetCurrentWindowUsage(ctx)
	if err != nil {
		log.Printf("error fetching window usage: %v", err)
	}

	// Extract values from interface{} types
	windowTokens := toInt64(windowUsage.WindowTokens)
	tokensPerMin := float64(burnRateMetrics.TokensPerMinute)

	// Calculate plan status (default to Max5 plan)
	planStatus := calculatePlanStatus("Max5", windowTokens, tokensPerMin)

	data := DashboardData{
		Metrics:           metrics,
		Today:             today,
		Week:              week,
		CacheMetrics:      cacheMetrics,
		TopProject:        topProject,
		EfficiencyMetrics: efficiencyMetrics,
		TopTool:           topTool,
		CacheHitRate:      cacheHitRate,
		PlanStatus:        planStatus,
	}

	Dashboard(data).Render(ctx, w)
}

func calculatePlanStatus(planName string, windowTokens int64, tokensPerMin float64) PlanStatus {
	limit := PlanLimits[planName]
	if limit == 0 {
		limit = PlanLimits["Max5"]
	}

	timeToLimit := calculateTimeToLimit(windowTokens, limit, tokensPerMin)

	return PlanStatus{
		PlanName:     planName,
		CurrentUsage: windowTokens,
		MaxLimit:     limit,
		BurnRate:     tokensPerMin,
		TimeToLimit:  timeToLimit,
	}
}

func calculateTimeToLimit(current, limit int64, burnRate float64) string {
	remaining := limit - current
	if remaining <= 0 {
		return "Limit reached"
	}
	if burnRate <= 0 {
		return ">24h"
	}

	minutesToLimit := float64(remaining) / burnRate
	if minutesToLimit > 1440 { // More than 24 hours
		return ">24h"
	}
	if minutesToLimit > 60 {
		hours := minutesToLimit / 60
		return fmt.Sprintf("~%.1fh", hours)
	}
	return fmt.Sprintf("~%.0fm", minutesToLimit)
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}
