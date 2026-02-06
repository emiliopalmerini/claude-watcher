package domain

// AggregateStats holds summary statistics across sessions.
type AggregateStats struct {
	SessionCount           int64
	TotalUserMessages      int64
	TotalAssistantMessages int64
	TotalTurns             int64
	TotalTokenInput        int64
	TotalTokenOutput       int64
	TotalTokenCacheRead    int64
	TotalTokenCacheWrite   int64
	TotalCostUsd           float64
	TotalErrors            int64
}

// ToolUsageStats holds usage data for a single tool.
type ToolUsageStats struct {
	ToolName         string
	TotalInvocations int64
	TotalErrors      int64
}

// ExperimentStats holds aggregate stats for a specific experiment.
type ExperimentStats struct {
	ExperimentID   string
	ExperimentName string
	AggregateStats
}

// TranscriptPathInfo holds session ID and transcript path for cleanup operations.
type TranscriptPathInfo struct {
	ID             string
	TranscriptPath string
}
