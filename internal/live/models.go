package live

import "time"

// LiveSession represents an active or recent session with velocity metrics
type LiveSession struct {
	SessionID       string
	Hostname        string
	WorkingDir      string
	GitBranch       string
	Duration        int64 // seconds
	Tokens          int64
	TokensPerMinute float64
	Timestamp       time.Time
	LastActivity    string // relative time string
	Status          string // "active", "idle", "stale"
}

// LiveData holds all data for the live sessions page
type LiveData struct {
	ActiveSessions []LiveSession
	RecentSessions []LiveSession
}
