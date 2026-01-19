package domain

import "time"

type UsageLimit struct {
	ID            string
	LimitValue    float64
	WarnThreshold float64
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UsageSummary struct {
	TotalTokens float64
	TotalCost   float64
}

type PlanConfig struct {
	PlanType          string
	WindowHours       int
	WindowStartTime   *time.Time
	LearnedTokenLimit *float64
	LearnedAt         *time.Time
	// Weekly window fields
	WeeklyWindowStartTime   *time.Time
	WeeklyLearnedTokenLimit *float64
	WeeklyLearnedAt         *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// Plan type constants
const (
	PlanPro    = "pro"
	PlanMax5x  = "max_5x"
	PlanMax20x = "max_20x"
	PlanCustom = "custom"
)

// Plan presets (messages per 5 hours - tokens not documented)
// These are rough estimates based on documented message limits
var PlanPresets = map[string]struct {
	Name              string
	MessagesPerWindow int
	TokenEstimate     float64 // Rough estimate: ~3K tokens per message avg
}{
	PlanPro:    {Name: "Pro", MessagesPerWindow: 45, TokenEstimate: 135000},
	PlanMax5x:  {Name: "Max 5x", MessagesPerWindow: 225, TokenEstimate: 675000},
	PlanMax20x: {Name: "Max 20x", MessagesPerWindow: 900, TokenEstimate: 2700000},
}

// Weekly plan presets (7-day rolling window)
// Based on documented "active hours" per week, converted to rough token estimates
// Active hours = time Claude is processing tokens (not idle time)
var WeeklyPlanPresets = map[string]struct {
	Name          string
	HoursEstimate float64 // Active hours per week (midpoint of documented range)
	TokenEstimate float64 // Rough token equivalent
}{
	PlanPro:    {Name: "Pro", HoursEstimate: 60, TokenEstimate: 4_000_000},      // 40-80 hours
	PlanMax5x:  {Name: "Max 5x", HoursEstimate: 210, TokenEstimate: 14_000_000}, // 140-280 hours
	PlanMax20x: {Name: "Max 20x", HoursEstimate: 360, TokenEstimate: 24_000_000}, // 240-480 hours
}

// WeeklyWindowHours is the fixed window size for weekly limits (7 days)
const WeeklyWindowHours = 168

// Limit type constants (legacy - for manual limits)
const (
	LimitDailyTokens  = "daily_tokens"
	LimitWeeklyTokens = "weekly_tokens"
	LimitDailyCost    = "daily_cost"
	LimitWeeklyCost   = "weekly_cost"
)
