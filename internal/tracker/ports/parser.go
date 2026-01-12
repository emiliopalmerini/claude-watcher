package ports

import "claude-watcher/internal/tracker/domain"

// TranscriptParser defines the interface for parsing session transcripts
type TranscriptParser interface {
	Parse(transcriptPath string) (domain.Statistics, error)
}
