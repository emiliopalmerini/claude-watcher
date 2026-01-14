package limits

import "time"

// Repository defines the interface for persisting and querying limit events
type Repository interface {
	// Save persists a limit event to the database
	Save(event LimitEvent) error

	// GetUsageSinceLastLimit returns aggregate usage metrics since the last limit event
	GetUsageSinceLastLimit() (UsageSummary, error)

	// GetLastLimitTimestamp returns the timestamp of the most recent limit event
	GetLastLimitTimestamp() (*time.Time, error)

	// ListRecent returns the most recent limit events within the given number of days
	ListRecent(days int) ([]LimitEvent, error)

	// ListByType returns limit events of a specific type
	ListByType(limitType LimitType, limit int) ([]LimitEvent, error)
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string)
	Error(msg string)
}
