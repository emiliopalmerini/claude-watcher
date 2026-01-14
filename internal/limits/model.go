package limits

import "time"

// LimitType indicates the period of the usage limit
type LimitType string

const (
	LimitTypeDaily   LimitType = "daily"
	LimitTypeWeekly  LimitType = "weekly"
	LimitTypeMonthly LimitType = "monthly"
)

// LimitEvent represents a usage limit event with aggregate usage data
type LimitEvent struct {
	ID             int64
	Timestamp      time.Time
	LimitType      LimitType
	ResetTime      *time.Time
	SessionsCount  int
	InputTokens    int
	OutputTokens   int
	ThinkingTokens int
	TotalCostUSD   float64
}

// UsageSummary holds aggregate usage metrics since the last limit event
type UsageSummary struct {
	SessionsCount  int
	InputTokens    int
	OutputTokens   int
	ThinkingTokens int
	TotalCostUSD   float64
}

// ParsedLimitInfo represents limit information extracted from a transcript
// This is used as input to the service, not stored directly
type ParsedLimitInfo struct {
	LimitType LimitType
	Timestamp time.Time
	Message   string
}
