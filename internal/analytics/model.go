package analytics

import "time"

// OverviewMetrics contains aggregate metrics for the dashboard overview
type OverviewMetrics struct {
	TotalSessions int
	TotalCost     float64
	Tokens        TokenSummary
	LimitHits     int
	LastLimitHit  *time.Time
}

// TokenSummary aggregates token usage
type TokenSummary struct {
	Input      int64
	Output     int64
	Thinking   int64
	CacheRead  int64
	CacheWrite int64
}

// Total returns the sum of input and output tokens
func (t TokenSummary) Total() int64 {
	return t.Input + t.Output
}
