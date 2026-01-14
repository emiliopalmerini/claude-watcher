package transcript

import (
	"encoding/json"
	"time"
)

// ParsedTranscript contains all data extracted from a Claude Code transcript
type ParsedTranscript struct {
	Statistics  SessionStatistics
	LimitEvents []LimitEvent
}

// SessionStatistics holds metrics extracted from a transcript
type SessionStatistics struct {
	// Interaction metrics
	UserPrompts        int
	AssistantResponses int
	ToolCalls          int
	ToolsBreakdown     map[string]int
	ErrorsCount        int

	// File tracking
	FilesAccessed []string
	FilesModified []string

	// Token usage
	InputTokens      int
	OutputTokens     int
	ThinkingTokens   int
	CacheReadTokens  int
	CacheWriteTokens int

	// Session metadata
	Model         string
	GitBranch     string
	ClaudeVersion string
	Summary       string
	StartTime     *time.Time
	EndTime       *time.Time
}

// Duration returns session duration in seconds
func (s SessionStatistics) Duration() int {
	if s.StartTime == nil || s.EndTime == nil {
		return 0
	}
	return int(s.EndTime.Sub(*s.StartTime).Seconds())
}

// TotalTokens returns the sum of all tokens (excluding cache)
func (s SessionStatistics) TotalTokens() int {
	return s.InputTokens + s.OutputTokens + s.ThinkingTokens
}

// NewSessionStatistics creates an initialized SessionStatistics
func NewSessionStatistics() SessionStatistics {
	return SessionStatistics{
		ToolsBreakdown: make(map[string]int),
		FilesAccessed:  []string{},
		FilesModified:  []string{},
	}
}

// LimitEvent represents a usage limit event extracted from a transcript
type LimitEvent struct {
	EventType  LimitEventType
	LimitType  LimitType
	Timestamp  time.Time
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

// RawEntry represents a single line in the JSONL transcript
type RawEntry struct {
	Type      string          `json:"type"`
	Subtype   string          `json:"subtype"`
	Timestamp string          `json:"timestamp"`
	GitBranch string          `json:"gitBranch"`
	Version   string          `json:"version"`
	Model     string          `json:"model"`
	Message   json.RawMessage `json:"message"`
	Name      string          `json:"name"`
	IsError   bool            `json:"is_error"`
	Content   json.RawMessage `json:"content"`
	Level     string          `json:"level"`
	Error     json.RawMessage `json:"error"`
}

// RawMessage represents the message field in transcript entries
type RawMessage struct {
	Usage   TokenUsage      `json:"usage"`
	Model   string          `json:"model"`
	Content json.RawMessage `json:"content"`
}

// TokenUsage represents token counts in a message
type TokenUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	ThinkingTokens           int `json:"thinking_tokens"`
}

// ContentItem represents a content item in assistant messages
type ContentItem struct {
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolInput represents generic tool input with file paths
type ToolInput struct {
	FilePath     string `json:"file_path"`
	Path         string `json:"path"`
	NotebookPath string `json:"notebook_path"`
}
