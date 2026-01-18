package domain

import "time"

type SessionQuality struct {
	SessionID         string
	OverallRating     *int
	IsSuccess         *bool
	AccuracyRating    *int
	HelpfulnessRating *int
	EfficiencyRating  *int
	Notes             *string
	ReviewedAt        *time.Time
	CreatedAt         time.Time
}
