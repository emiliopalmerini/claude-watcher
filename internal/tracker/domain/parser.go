package domain

// TranscriptParser defines the interface for parsing session transcripts
type TranscriptParser interface {
	Parse(transcriptPath string) (ParsedTranscript, error)
}

// ParsedTranscript contains all data extracted from a Claude Code transcript
type ParsedTranscript struct {
	Statistics  Statistics
	LimitEvents []LimitEvent
}

// LimitEvent represents a usage limit event extracted from a transcript
type LimitEvent struct {
	EventType  LimitEventType
	LimitType  LimitType
	Timestamp  string
	Message    string
	TokensUsed int
	CostUsed   float64
}

// LimitEventType indicates whether a limit was hit or reset
type LimitEventType string

const (
	LimitEventHit   LimitEventType = "hit"
	LimitEventReset LimitEventType = "reset"
)

// LimitType indicates the period of the usage limit
type LimitType string

const (
	LimitTypeDaily   LimitType = "daily"
	LimitTypeWeekly  LimitType = "weekly"
	LimitTypeMonthly LimitType = "monthly"
)
