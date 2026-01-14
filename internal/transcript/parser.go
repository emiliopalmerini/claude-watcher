package transcript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

// JSONLParser implements the Parser interface for JSONL transcripts
type JSONLParser struct{}

// NewParser creates a new JSONLParser
func NewParser() *JSONLParser {
	return &JSONLParser{}
}

// Parse parses a transcript file and extracts statistics and limit events
func (p *JSONLParser) Parse(transcriptPath string) (ParsedTranscript, error) {
	result := ParsedTranscript{
		Statistics:  NewSessionStatistics(),
		LimitEvents: []LimitEvent{},
	}

	if transcriptPath == "" {
		return result, nil
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, fmt.Errorf("open transcript: %w", err)
	}
	defer file.Close()

	return p.ParseReader(file)
}

// ParseReader parses from an io.Reader
func (p *JSONLParser) ParseReader(r io.Reader) (ParsedTranscript, error) {
	result := ParsedTranscript{
		Statistics:  NewSessionStatistics(),
		LimitEvents: []LimitEvent{},
	}

	scanner := bufio.NewScanner(r)
	const maxCapacity = 1024 * 1024 // 1MB buffer for long lines
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	filesAccessedSet := make(map[string]bool)
	filesModifiedSet := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry RawEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip malformed entries
		}

		p.processEntry(&entry, &result, filesAccessedSet, filesModifiedSet)
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("scan transcript: %w", err)
	}

	// Convert sets to slices
	for file := range filesAccessedSet {
		result.Statistics.FilesAccessed = append(result.Statistics.FilesAccessed, file)
	}
	for file := range filesModifiedSet {
		result.Statistics.FilesModified = append(result.Statistics.FilesModified, file)
	}

	return result, nil
}

func (p *JSONLParser) processEntry(
	entry *RawEntry,
	result *ParsedTranscript,
	filesAccessed, filesModified map[string]bool,
) {
	p.updateTimestamps(entry, &result.Statistics)
	p.extractMetadata(entry, &result.Statistics)

	switch entry.Type {
	case "human", "user":
		p.processUserMessage(entry, &result.Statistics)
	case "assistant":
		p.processAssistantMessage(entry, &result.Statistics, filesAccessed, filesModified)
	case "tool_use":
		p.processToolUse(entry, &result.Statistics)
	case "tool_result":
		p.processToolResult(entry, &result.Statistics)
	case "system":
		p.processSystemMessage(entry, result)
	}
}

func (p *JSONLParser) updateTimestamps(entry *RawEntry, stats *SessionStatistics) {
	if entry.Timestamp == "" {
		return
	}

	t, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		return
	}

	if stats.StartTime == nil {
		stats.StartTime = &t
	}
	stats.EndTime = &t
}

func (p *JSONLParser) extractMetadata(entry *RawEntry, stats *SessionStatistics) {
	if stats.GitBranch == "" && entry.GitBranch != "" {
		stats.GitBranch = entry.GitBranch
	}
	if stats.ClaudeVersion == "" && entry.Version != "" {
		stats.ClaudeVersion = entry.Version
	}
}

func (p *JSONLParser) processUserMessage(entry *RawEntry, stats *SessionStatistics) {
	stats.UserPrompts++

	if stats.Summary != "" {
		return
	}

	var msg RawMessage
	if err := json.Unmarshal(entry.Message, &msg); err != nil {
		return
	}

	content := extractTextContent(msg.Content)
	if content != "" {
		if len(content) > 200 {
			stats.Summary = content[:200]
		} else {
			stats.Summary = content
		}
	}
}

// extractTextContent handles both string content and array content formats
// Real transcripts use: {"content": "text"} or {"content": [{"type":"text","text":"..."}]}
func extractTextContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try parsing as simple string first
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Try parsing as array of content items
	var items []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &items); err == nil {
		for _, item := range items {
			if item.Type == "text" && item.Text != "" {
				return item.Text
			}
		}
	}

	return ""
}

func (p *JSONLParser) processAssistantMessage(
	entry *RawEntry,
	stats *SessionStatistics,
	filesAccessed, filesModified map[string]bool,
) {
	stats.AssistantResponses++

	var msg RawMessage
	if err := json.Unmarshal(entry.Message, &msg); err != nil {
		return
	}

	// Extract token usage
	stats.InputTokens += msg.Usage.InputTokens
	stats.OutputTokens += msg.Usage.OutputTokens
	stats.CacheReadTokens += msg.Usage.CacheReadInputTokens
	stats.CacheWriteTokens += msg.Usage.CacheCreationInputTokens
	stats.ThinkingTokens += msg.Usage.ThinkingTokens

	// Get model
	if stats.Model == "" {
		if msg.Model != "" {
			stats.Model = msg.Model
		} else if entry.Model != "" {
			stats.Model = entry.Model
		}
	}

	// Process tool uses in content
	var contentItems []ContentItem
	if err := json.Unmarshal(msg.Content, &contentItems); err != nil {
		return
	}

	for _, item := range contentItems {
		if item.Type == "tool_use" {
			p.recordToolUse(item.Name, item.Input, stats, filesAccessed, filesModified)
		}
	}
}

func (p *JSONLParser) processToolUse(entry *RawEntry, stats *SessionStatistics) {
	toolName := entry.Name
	if toolName == "" {
		toolName = "unknown"
	}
	stats.ToolCalls++
	stats.ToolsBreakdown[toolName]++
}

func (p *JSONLParser) processToolResult(entry *RawEntry, stats *SessionStatistics) {
	if entry.IsError {
		stats.ErrorsCount++
		return
	}

	var content string
	if err := json.Unmarshal(entry.Content, &content); err != nil {
		return
	}

	// Check first 100 chars for error indicators
	if len(content) > 100 {
		content = content[:100]
	}
	if strings.Contains(strings.ToLower(content), "error") {
		stats.ErrorsCount++
	}
}

func (p *JSONLParser) processSystemMessage(entry *RawEntry, result *ParsedTranscript) {
	var content string
	if err := json.Unmarshal(entry.Content, &content); err != nil {
		return
	}

	contentLower := strings.ToLower(content)

	// Check for limit-related messages
	if !isLimitMessage(contentLower) {
		return
	}

	event := p.extractLimitEvent(content, entry.Timestamp)
	if event != nil {
		result.LimitEvents = append(result.LimitEvents, *event)
	}
}

func (p *JSONLParser) extractLimitEvent(content, timestamp string) *LimitEvent {
	contentLower := strings.ToLower(content)

	event := &LimitEvent{
		Message: content,
	}

	// Parse timestamp
	if timestamp != "" {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			event.Timestamp = t
		} else {
			event.Timestamp = time.Now().UTC()
		}
	} else {
		event.Timestamp = time.Now().UTC()
	}

	// Determine event type (hit or reset)
	// Check for reset first with more specific patterns to avoid false positives
	// (e.g., "It resets in 6 hours" should be a hit, not a reset)
	isReset := (strings.Contains(contentLower, "has been reset") ||
		strings.Contains(contentLower, "limit reset") ||
		strings.Contains(contentLower, "restored") ||
		strings.Contains(contentLower, "renewed")) &&
		!strings.Contains(contentLower, "hit") &&
		!strings.Contains(contentLower, "reached") &&
		!strings.Contains(contentLower, "exceeded")

	if isReset {
		event.EventType = LimitEventReset
	} else {
		event.EventType = LimitEventHit
	}

	// Determine limit type (daily or weekly)
	if strings.Contains(contentLower, "daily") ||
		strings.Contains(contentLower, "24 hour") ||
		strings.Contains(contentLower, "today") {
		event.LimitType = LimitTypeDaily
	} else if strings.Contains(contentLower, "weekly") ||
		strings.Contains(contentLower, "7 day") ||
		strings.Contains(contentLower, "week") {
		event.LimitType = LimitTypeWeekly
	} else {
		event.LimitType = LimitTypeDaily // Default to daily
	}

	// Try to extract token count from message
	event.TokensUsed = extractTokenCount(content)

	return event
}

func (p *JSONLParser) recordToolUse(
	toolName string,
	inputRaw json.RawMessage,
	stats *SessionStatistics,
	filesAccessed, filesModified map[string]bool,
) {
	if toolName == "" {
		toolName = "unknown"
	}

	stats.ToolCalls++
	stats.ToolsBreakdown[toolName]++

	if !isFileTool(toolName) {
		return
	}

	var input ToolInput
	if err := json.Unmarshal(inputRaw, &input); err != nil {
		return
	}

	filePath := input.FilePath
	if filePath == "" {
		filePath = input.Path
	}
	if filePath == "" {
		filePath = input.NotebookPath
	}

	if filePath == "" {
		return
	}

	filesAccessed[filePath] = true
	if isModifyingTool(toolName) {
		filesModified[filePath] = true
	}
}

func isLimitMessage(content string) bool {
	limitIndicators := []string{
		"limit",
		"quota",
		"rate",
		"exceeded",
		"throttle",
		"resets",
	}

	for _, indicator := range limitIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	return false
}

var tokenCountPattern = regexp.MustCompile(`(\d{1,3}(?:,\d{3})*|\d+)\s*(?:tokens?|k)`)

func extractTokenCount(content string) int {
	matches := tokenCountPattern.FindStringSubmatch(strings.ToLower(content))
	if len(matches) < 2 {
		return 0
	}

	numStr := strings.ReplaceAll(matches[1], ",", "")
	var count int
	_, err := fmt.Sscanf(numStr, "%d", &count)
	if err != nil {
		return 0
	}

	// Handle 'k' suffix (e.g., "500k tokens")
	if strings.Contains(matches[0], "k") && !strings.Contains(matches[0], "token") {
		count *= 1000
	}

	return count
}

func isFileTool(toolName string) bool {
	fileTools := map[string]bool{
		"Read":         true,
		"Edit":         true,
		"Write":        true,
		"Glob":         true,
		"Grep":         true,
		"LSP":          true,
		"NotebookEdit": true,
	}
	return fileTools[toolName]
}

func isModifyingTool(toolName string) bool {
	modifyingTools := map[string]bool{
		"Edit":         true,
		"Write":        true,
		"NotebookEdit": true,
	}
	return modifyingTools[toolName]
}
