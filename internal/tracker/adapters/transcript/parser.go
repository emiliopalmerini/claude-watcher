package transcript

import (
	"claude-watcher/internal/tracker/domain"
	"claude-watcher/internal/transcript"
)

// Parser implements the TranscriptParser port by wrapping the Phase 2 parser
type Parser struct {
	parser *transcript.JSONLParser
}

// NewParser creates a new transcript parser adapter
func NewParser() *Parser {
	return &Parser{
		parser: transcript.NewParser(),
	}
}

// Parse parses a Claude Code transcript file and extracts statistics and limit events
func (p *Parser) Parse(transcriptPath string) (domain.ParsedTranscript, error) {
	result := domain.ParsedTranscript{
		Statistics:  domain.NewStatistics(),
		LimitEvents: []domain.LimitEvent{},
	}

	parsed, err := p.parser.Parse(transcriptPath)
	if err != nil {
		return result, err
	}

	// Convert statistics
	result.Statistics = convertStatistics(parsed.Statistics)

	// Convert limit events
	for _, event := range parsed.LimitEvents {
		result.LimitEvents = append(result.LimitEvents, convertLimitEvent(event))
	}

	return result, nil
}

// convertStatistics converts Phase 2 SessionStatistics to domain Statistics
func convertStatistics(src transcript.SessionStatistics) domain.Statistics {
	stats := domain.Statistics{
		UserPrompts:        src.UserPrompts,
		AssistantResponses: src.AssistantResponses,
		ToolCalls:          src.ToolCalls,
		ToolsBreakdown:     src.ToolsBreakdown,
		ErrorsCount:        src.ErrorsCount,
		FilesAccessed:      src.FilesAccessed,
		FilesModified:      src.FilesModified,
		InputTokens:        src.InputTokens,
		OutputTokens:       src.OutputTokens,
		ThinkingTokens:     src.ThinkingTokens,
		CacheReadTokens:    src.CacheReadTokens,
		CacheWriteTokens:   src.CacheWriteTokens,
		Model:              src.Model,
		GitBranch:          src.GitBranch,
		ClaudeVersion:      src.ClaudeVersion,
		Summary:            src.Summary,
		StartTime:          src.StartTime,
		EndTime:            src.EndTime,
	}

	// Ensure maps and slices are initialized
	if stats.ToolsBreakdown == nil {
		stats.ToolsBreakdown = make(map[string]int)
	}
	if stats.FilesAccessed == nil {
		stats.FilesAccessed = []string{}
	}
	if stats.FilesModified == nil {
		stats.FilesModified = []string{}
	}

	return stats
}

// convertLimitEvent converts Phase 2 LimitEvent to domain LimitEvent
func convertLimitEvent(src transcript.LimitEvent) domain.LimitEvent {
	return domain.LimitEvent{
		EventType:  domain.LimitEventType(src.EventType),
		LimitType:  domain.LimitType(src.LimitType),
		Timestamp:  src.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Message:    src.Message,
		TokensUsed: src.TokensUsed,
		CostUsed:   src.CostUsed,
	}
}
