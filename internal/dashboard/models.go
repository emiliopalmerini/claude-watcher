package dashboard

import "claude-watcher/internal/database/sqlc"

// PlanLimits defines token limits for different Claude subscription plans
var PlanLimits = map[string]int64{
	"Pro":    19000,
	"Max5":   88000,
	"Max20":  220000,
	"Custom": 500000, // Default high limit for custom plans
}

type PlanStatus struct {
	PlanName     string
	CurrentUsage int64
	MaxLimit     int64
	BurnRate     float64 // tokens per minute
	TimeToLimit  string  // formatted time remaining
}

type DashboardData struct {
	Metrics           sqlc.GetDashboardMetricsRow
	Today             sqlc.GetTodayMetricsRow
	Week              sqlc.GetWeekMetricsRow
	CacheMetrics      sqlc.GetCacheMetricsRow
	TopProject        sqlc.GetTopProjectRow
	EfficiencyMetrics sqlc.GetEfficiencyMetricsRow
	TopTool           string
	CacheHitRate      float64
	PlanStatus        PlanStatus
}
