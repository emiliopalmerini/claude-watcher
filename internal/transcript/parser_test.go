package transcript

import (
	"strings"
	"testing"
	"time"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
}

func TestParseReader_EmptyInput(t *testing.T) {
	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.UserPrompts != 0 {
		t.Errorf("expected 0 user prompts, got %d", result.Statistics.UserPrompts)
	}
	if len(result.LimitEvents) != 0 {
		t.Errorf("expected 0 limit events, got %d", len(result.LimitEvents))
	}
}

func TestParseReader_UserMessage(t *testing.T) {
	input := `{"type":"user","timestamp":"2024-01-15T10:00:00Z","message":{"content":"Hello, Claude!"}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.UserPrompts != 1 {
		t.Errorf("expected 1 user prompt, got %d", result.Statistics.UserPrompts)
	}
	if result.Statistics.Summary != "Hello, Claude!" {
		t.Errorf("expected summary 'Hello, Claude!', got '%s'", result.Statistics.Summary)
	}
}

func TestParseReader_HumanMessage(t *testing.T) {
	input := `{"type":"human","timestamp":"2024-01-15T10:00:00Z","message":{"content":"\"Test prompt\""}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.UserPrompts != 1 {
		t.Errorf("expected 1 user prompt, got %d", result.Statistics.UserPrompts)
	}
}

func TestParseReader_AssistantMessage_TokenUsage(t *testing.T) {
	input := `{"type":"assistant","timestamp":"2024-01-15T10:00:00Z","message":{"model":"claude-3-sonnet","usage":{"input_tokens":100,"output_tokens":200,"cache_read_input_tokens":50,"cache_creation_input_tokens":25,"thinking_tokens":75},"content":"[]"}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := result.Statistics
	if stats.AssistantResponses != 1 {
		t.Errorf("expected 1 assistant response, got %d", stats.AssistantResponses)
	}
	if stats.InputTokens != 100 {
		t.Errorf("expected 100 input tokens, got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 200 {
		t.Errorf("expected 200 output tokens, got %d", stats.OutputTokens)
	}
	if stats.CacheReadTokens != 50 {
		t.Errorf("expected 50 cache read tokens, got %d", stats.CacheReadTokens)
	}
	if stats.CacheWriteTokens != 25 {
		t.Errorf("expected 25 cache write tokens, got %d", stats.CacheWriteTokens)
	}
	if stats.ThinkingTokens != 75 {
		t.Errorf("expected 75 thinking tokens, got %d", stats.ThinkingTokens)
	}
	if stats.Model != "claude-3-sonnet" {
		t.Errorf("expected model 'claude-3-sonnet', got '%s'", stats.Model)
	}
}

func TestParseReader_ToolUse(t *testing.T) {
	input := `{"type":"tool_use","name":"Read","timestamp":"2024-01-15T10:00:00Z"}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.ToolCalls != 1 {
		t.Errorf("expected 1 tool call, got %d", result.Statistics.ToolCalls)
	}
	if result.Statistics.ToolsBreakdown["Read"] != 1 {
		t.Errorf("expected 1 Read call in breakdown, got %d", result.Statistics.ToolsBreakdown["Read"])
	}
}

func TestParseReader_ToolResult_Error(t *testing.T) {
	input := `{"type":"tool_result","is_error":true,"timestamp":"2024-01-15T10:00:00Z"}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.ErrorsCount != 1 {
		t.Errorf("expected 1 error, got %d", result.Statistics.ErrorsCount)
	}
}

func TestParseReader_ToolResult_ErrorInContent(t *testing.T) {
	input := `{"type":"tool_result","is_error":false,"content":"\"File not found error\"","timestamp":"2024-01-15T10:00:00Z"}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.ErrorsCount != 1 {
		t.Errorf("expected 1 error detected in content, got %d", result.Statistics.ErrorsCount)
	}
}

func TestParseReader_Timestamps(t *testing.T) {
	input := `{"type":"user","timestamp":"2024-01-15T10:00:00Z","message":{"content":"\"First\""}}
{"type":"user","timestamp":"2024-01-15T10:30:00Z","message":{"content":"\"Second\""}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.StartTime == nil {
		t.Fatal("expected start time to be set")
	}
	if result.Statistics.EndTime == nil {
		t.Fatal("expected end time to be set")
	}

	expectedStart := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	if !result.Statistics.StartTime.Equal(expectedStart) {
		t.Errorf("expected start time %v, got %v", expectedStart, *result.Statistics.StartTime)
	}
	if !result.Statistics.EndTime.Equal(expectedEnd) {
		t.Errorf("expected end time %v, got %v", expectedEnd, *result.Statistics.EndTime)
	}

	if result.Statistics.Duration() != 1800 {
		t.Errorf("expected duration 1800 seconds, got %d", result.Statistics.Duration())
	}
}

func TestParseReader_Metadata(t *testing.T) {
	input := `{"type":"user","gitBranch":"main","version":"1.0.0","timestamp":"2024-01-15T10:00:00Z","message":{"content":"\"Test\""}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.GitBranch != "main" {
		t.Errorf("expected git branch 'main', got '%s'", result.Statistics.GitBranch)
	}
	if result.Statistics.ClaudeVersion != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", result.Statistics.ClaudeVersion)
	}
}

func TestParseReader_LimitEvent_Hit(t *testing.T) {
	input := `{"type":"system","timestamp":"2024-01-15T10:00:00Z","content":"\"You've hit your daily limit. It resets in 6 hours.\""}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.LimitEvents) != 1 {
		t.Fatalf("expected 1 limit event, got %d", len(result.LimitEvents))
	}

	event := result.LimitEvents[0]
	if event.EventType != LimitEventHit {
		t.Errorf("expected event type 'hit', got '%s'", event.EventType)
	}
	if event.LimitType != LimitTypeDaily {
		t.Errorf("expected limit type 'daily', got '%s'", event.LimitType)
	}
}

func TestParseReader_LimitEvent_WeeklyHit(t *testing.T) {
	input := `{"type":"system","timestamp":"2024-01-15T10:00:00Z","content":"\"You've reached your weekly limit.\""}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.LimitEvents) != 1 {
		t.Fatalf("expected 1 limit event, got %d", len(result.LimitEvents))
	}

	event := result.LimitEvents[0]
	if event.EventType != LimitEventHit {
		t.Errorf("expected event type 'hit', got '%s'", event.EventType)
	}
	if event.LimitType != LimitTypeWeekly {
		t.Errorf("expected limit type 'weekly', got '%s'", event.LimitType)
	}
}

func TestParseReader_LimitEvent_Reset(t *testing.T) {
	input := `{"type":"system","timestamp":"2024-01-15T10:00:00Z","content":"\"Your daily limit has been reset.\""}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.LimitEvents) != 1 {
		t.Fatalf("expected 1 limit event, got %d", len(result.LimitEvents))
	}

	event := result.LimitEvents[0]
	if event.EventType != LimitEventReset {
		t.Errorf("expected event type 'reset', got '%s'", event.EventType)
	}
}

func TestParseReader_LimitEvent_TokenExtraction(t *testing.T) {
	input := `{"type":"system","timestamp":"2024-01-15T10:00:00Z","content":"\"You've used 1,000,000 tokens and hit your limit.\""}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.LimitEvents) != 1 {
		t.Fatalf("expected 1 limit event, got %d", len(result.LimitEvents))
	}

	event := result.LimitEvents[0]
	if event.TokensUsed != 1000000 {
		t.Errorf("expected 1000000 tokens, got %d", event.TokensUsed)
	}
}

func TestParseReader_AssistantMessage_ToolUseInContent(t *testing.T) {
	input := `{"type":"assistant","timestamp":"2024-01-15T10:00:00Z","message":{"model":"claude-3","usage":{"input_tokens":50,"output_tokens":100},"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/test/file.go"}}]}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.ToolCalls != 1 {
		t.Errorf("expected 1 tool call, got %d", result.Statistics.ToolCalls)
	}
	if result.Statistics.ToolsBreakdown["Read"] != 1 {
		t.Errorf("expected 1 Read call, got %d", result.Statistics.ToolsBreakdown["Read"])
	}
	if len(result.Statistics.FilesAccessed) != 1 {
		t.Fatalf("expected 1 file accessed, got %d", len(result.Statistics.FilesAccessed))
	}
	if result.Statistics.FilesAccessed[0] != "/test/file.go" {
		t.Errorf("expected file '/test/file.go', got '%s'", result.Statistics.FilesAccessed[0])
	}
}

func TestParseReader_FileModification(t *testing.T) {
	input := `{"type":"assistant","timestamp":"2024-01-15T10:00:00Z","message":{"model":"claude-3","usage":{},"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"/test/file.go"}}]}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Statistics.FilesModified) != 1 {
		t.Errorf("expected 1 file modified, got %d", len(result.Statistics.FilesModified))
	}
}

func TestParseReader_MalformedLine(t *testing.T) {
	input := `not valid json
{"type":"user","timestamp":"2024-01-15T10:00:00Z","message":{"content":"\"Valid line\""}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still process valid lines
	if result.Statistics.UserPrompts != 1 {
		t.Errorf("expected 1 user prompt, got %d", result.Statistics.UserPrompts)
	}
}

func TestParseReader_SummaryTruncation(t *testing.T) {
	longPrompt := strings.Repeat("a", 300)
	input := `{"type":"user","timestamp":"2024-01-15T10:00:00Z","message":{"content":"\"` + longPrompt + `\""}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Statistics.Summary) != 200 {
		t.Errorf("expected summary truncated to 200 chars, got %d", len(result.Statistics.Summary))
	}
}

func TestParseReader_CompleteSession(t *testing.T) {
	input := `{"type":"user","timestamp":"2024-01-15T10:00:00Z","gitBranch":"feature","version":"1.0.0","message":{"content":"Fix the bug"}}
{"type":"assistant","timestamp":"2024-01-15T10:01:00Z","message":{"model":"claude-3-opus","usage":{"input_tokens":500,"output_tokens":1000,"thinking_tokens":200},"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/src/main.go"}}]}}
{"type":"tool_result","is_error":false,"content":"package main","timestamp":"2024-01-15T10:01:05Z"}
{"type":"assistant","timestamp":"2024-01-15T10:02:00Z","message":{"model":"claude-3-opus","usage":{"input_tokens":600,"output_tokens":800},"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"/src/main.go"}}]}}
{"type":"user","timestamp":"2024-01-15T10:05:00Z","message":{"content":"Thanks!"}}
{"type":"system","timestamp":"2024-01-15T10:06:00Z","content":"You've hit your daily limit."}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := result.Statistics

	// Check interactions
	if stats.UserPrompts != 2 {
		t.Errorf("expected 2 user prompts, got %d", stats.UserPrompts)
	}
	if stats.AssistantResponses != 2 {
		t.Errorf("expected 2 assistant responses, got %d", stats.AssistantResponses)
	}

	// Check tokens
	if stats.InputTokens != 1100 {
		t.Errorf("expected 1100 input tokens, got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 1800 {
		t.Errorf("expected 1800 output tokens, got %d", stats.OutputTokens)
	}

	// Check metadata
	if stats.Model != "claude-3-opus" {
		t.Errorf("expected model 'claude-3-opus', got '%s'", stats.Model)
	}
	if stats.GitBranch != "feature" {
		t.Errorf("expected branch 'feature', got '%s'", stats.GitBranch)
	}
	if stats.Summary != "Fix the bug" {
		t.Errorf("expected summary 'Fix the bug', got '%s'", stats.Summary)
	}

	// Check tool usage
	if stats.ToolCalls != 2 {
		t.Errorf("expected 2 tool calls, got %d", stats.ToolCalls)
	}

	// Check files
	if len(stats.FilesAccessed) != 1 {
		t.Errorf("expected 1 file accessed, got %d", len(stats.FilesAccessed))
	}
	if len(stats.FilesModified) != 1 {
		t.Errorf("expected 1 file modified, got %d", len(stats.FilesModified))
	}

	// Check limit events
	if len(result.LimitEvents) != 1 {
		t.Errorf("expected 1 limit event, got %d", len(result.LimitEvents))
	}
}

func TestParse_NonExistentFile(t *testing.T) {
	p := NewParser()
	result, err := p.Parse("/nonexistent/path/transcript.jsonl")
	if err != nil {
		t.Fatalf("unexpected error for non-existent file: %v", err)
	}

	// Should return empty result without error
	if result.Statistics.UserPrompts != 0 {
		t.Errorf("expected 0 user prompts, got %d", result.Statistics.UserPrompts)
	}
}

func TestParse_EmptyPath(t *testing.T) {
	p := NewParser()
	result, err := p.Parse("")
	if err != nil {
		t.Fatalf("unexpected error for empty path: %v", err)
	}

	if result.Statistics.UserPrompts != 0 {
		t.Errorf("expected 0 user prompts, got %d", result.Statistics.UserPrompts)
	}
}

func TestSessionStatistics_TotalTokens(t *testing.T) {
	stats := SessionStatistics{
		InputTokens:    100,
		OutputTokens:   200,
		ThinkingTokens: 50,
	}

	if stats.TotalTokens() != 350 {
		t.Errorf("expected total tokens 350, got %d", stats.TotalTokens())
	}
}

func TestSessionStatistics_Duration_NilTimes(t *testing.T) {
	stats := SessionStatistics{}
	if stats.Duration() != 0 {
		t.Errorf("expected duration 0 for nil times, got %d", stats.Duration())
	}
}

func TestExtractTokenCount(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"You've used 1000 tokens", 1000},
		{"You've used 1,000,000 tokens", 1000000},
		{"500 tokens remaining", 500},
		{"No tokens mentioned", 0},
		{"", 0},
	}

	for _, tc := range tests {
		result := extractTokenCount(tc.input)
		if result != tc.expected {
			t.Errorf("extractTokenCount(%q) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestIsFileTool(t *testing.T) {
	fileTools := []string{"Read", "Edit", "Write", "Glob", "Grep", "LSP", "NotebookEdit"}
	nonFileTools := []string{"Bash", "Task", "WebSearch", "unknown"}

	for _, tool := range fileTools {
		if !isFileTool(tool) {
			t.Errorf("expected %s to be a file tool", tool)
		}
	}

	for _, tool := range nonFileTools {
		if isFileTool(tool) {
			t.Errorf("expected %s to not be a file tool", tool)
		}
	}
}

func TestIsModifyingTool(t *testing.T) {
	modifying := []string{"Edit", "Write", "NotebookEdit"}
	nonModifying := []string{"Read", "Glob", "Grep", "Bash"}

	for _, tool := range modifying {
		if !isModifyingTool(tool) {
			t.Errorf("expected %s to be a modifying tool", tool)
		}
	}

	for _, tool := range nonModifying {
		if isModifyingTool(tool) {
			t.Errorf("expected %s to not be a modifying tool", tool)
		}
	}
}

// Realistic tests based on actual Claude Code transcript format

func TestParseReader_RealisticUserMessage(t *testing.T) {
	// Actual format from Claude Code transcripts
	input := `{"parentUuid":null,"isSidechain":false,"userType":"external","cwd":"/Users/test/project","sessionId":"abc-123","version":"2.0.76","gitBranch":"main","type":"user","message":{"role":"user","content":"Fix the authentication bug"},"uuid":"msg-001","timestamp":"2025-01-15T10:00:00Z"}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.UserPrompts != 1 {
		t.Errorf("expected 1 user prompt, got %d", result.Statistics.UserPrompts)
	}
	if result.Statistics.Summary != "Fix the authentication bug" {
		t.Errorf("expected summary 'Fix the authentication bug', got '%s'", result.Statistics.Summary)
	}
	if result.Statistics.GitBranch != "main" {
		t.Errorf("expected git branch 'main', got '%s'", result.Statistics.GitBranch)
	}
	if result.Statistics.ClaudeVersion != "2.0.76" {
		t.Errorf("expected version '2.0.76', got '%s'", result.Statistics.ClaudeVersion)
	}
}

func TestParseReader_RealisticUserMessageWithArrayContent(t *testing.T) {
	// User message with array content format (used for tool results)
	input := `{"type":"user","timestamp":"2025-01-15T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Please review the changes"}]}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.Summary != "Please review the changes" {
		t.Errorf("expected summary 'Please review the changes', got '%s'", result.Statistics.Summary)
	}
}

func TestParseReader_RealisticAssistantMessage(t *testing.T) {
	// Actual format from Claude Code transcripts with full usage structure
	input := `{"parentUuid":"msg-001","isSidechain":false,"userType":"external","cwd":"/Users/test/project","sessionId":"abc-123","version":"2.0.76","gitBranch":"main","message":{"model":"claude-opus-4-1-20250805","id":"msg_abc123","type":"message","role":"assistant","content":[{"type":"text","text":"I'll help you fix that bug."}],"stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":1500,"cache_creation_input_tokens":500,"cache_read_input_tokens":12000,"output_tokens":250,"service_tier":"standard"}},"requestId":"req_xyz","type":"assistant","uuid":"msg-002","timestamp":"2025-01-15T10:00:05Z"}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := result.Statistics
	if stats.AssistantResponses != 1 {
		t.Errorf("expected 1 assistant response, got %d", stats.AssistantResponses)
	}
	if stats.InputTokens != 1500 {
		t.Errorf("expected 1500 input tokens, got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 250 {
		t.Errorf("expected 250 output tokens, got %d", stats.OutputTokens)
	}
	if stats.CacheReadTokens != 12000 {
		t.Errorf("expected 12000 cache read tokens, got %d", stats.CacheReadTokens)
	}
	if stats.CacheWriteTokens != 500 {
		t.Errorf("expected 500 cache write tokens, got %d", stats.CacheWriteTokens)
	}
	if stats.Model != "claude-opus-4-1-20250805" {
		t.Errorf("expected model 'claude-opus-4-1-20250805', got '%s'", stats.Model)
	}
}

func TestParseReader_RealisticAssistantWithThinking(t *testing.T) {
	// Assistant message with thinking content (extended thinking mode)
	input := `{"type":"assistant","timestamp":"2025-01-15T10:00:00Z","message":{"model":"claude-opus-4-1-20250805","usage":{"input_tokens":100,"output_tokens":50,"thinking_tokens":500},"content":[{"type":"thinking","thinking":"Let me analyze this problem...","signature":"abc123"},{"type":"text","text":"Here's my analysis."}]}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.ThinkingTokens != 500 {
		t.Errorf("expected 500 thinking tokens, got %d", result.Statistics.ThinkingTokens)
	}
}

func TestParseReader_RealisticAssistantWithToolUse(t *testing.T) {
	// Assistant message with tool_use in content array
	input := `{"type":"assistant","timestamp":"2025-01-15T10:00:00Z","message":{"model":"claude-3-sonnet","usage":{"input_tokens":200,"output_tokens":100},"content":[{"type":"text","text":"Let me read that file."},{"type":"tool_use","id":"toolu_abc123","name":"Read","input":{"file_path":"/src/main.go"}}]}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Statistics.ToolCalls != 1 {
		t.Errorf("expected 1 tool call, got %d", result.Statistics.ToolCalls)
	}
	if result.Statistics.ToolsBreakdown["Read"] != 1 {
		t.Errorf("expected 1 Read call, got %d", result.Statistics.ToolsBreakdown["Read"])
	}
	if len(result.Statistics.FilesAccessed) != 1 {
		t.Fatalf("expected 1 file accessed, got %d", len(result.Statistics.FilesAccessed))
	}
	if result.Statistics.FilesAccessed[0] != "/src/main.go" {
		t.Errorf("expected file '/src/main.go', got '%s'", result.Statistics.FilesAccessed[0])
	}
}

func TestParseReader_RealisticCompleteSession(t *testing.T) {
	// Complete realistic session with multiple interactions
	input := `{"parentUuid":null,"isSidechain":false,"cwd":"/Users/dev/myproject","sessionId":"session-xyz","version":"2.0.76","gitBranch":"feature/auth","type":"user","message":{"role":"user","content":"Add login validation"},"uuid":"msg-001","timestamp":"2025-01-15T10:00:00Z"}
{"parentUuid":"msg-001","cwd":"/Users/dev/myproject","sessionId":"session-xyz","version":"2.0.76","gitBranch":"feature/auth","message":{"model":"claude-opus-4-1-20250805","id":"msg_001","type":"message","role":"assistant","content":[{"type":"text","text":"I'll help add login validation."},{"type":"tool_use","id":"toolu_001","name":"Read","input":{"file_path":"/src/auth/login.go"}}],"usage":{"input_tokens":3000,"cache_creation_input_tokens":1000,"cache_read_input_tokens":15000,"output_tokens":150}},"type":"assistant","uuid":"msg-002","timestamp":"2025-01-15T10:00:05Z"}
{"parentUuid":"msg-002","cwd":"/Users/dev/myproject","sessionId":"session-xyz","version":"2.0.76","gitBranch":"feature/auth","message":{"model":"claude-opus-4-1-20250805","id":"msg_002","type":"message","role":"assistant","content":[{"type":"tool_use","id":"toolu_002","name":"Edit","input":{"file_path":"/src/auth/login.go"}}],"usage":{"input_tokens":4000,"output_tokens":500}},"type":"assistant","uuid":"msg-003","timestamp":"2025-01-15T10:00:30Z"}
{"parentUuid":"msg-003","cwd":"/Users/dev/myproject","sessionId":"session-xyz","type":"user","message":{"role":"user","content":"Looks good, thanks!"},"uuid":"msg-004","timestamp":"2025-01-15T10:01:00Z"}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := result.Statistics

	// Check interaction counts
	if stats.UserPrompts != 2 {
		t.Errorf("expected 2 user prompts, got %d", stats.UserPrompts)
	}
	if stats.AssistantResponses != 2 {
		t.Errorf("expected 2 assistant responses, got %d", stats.AssistantResponses)
	}

	// Check tokens (cumulative across responses)
	if stats.InputTokens != 7000 {
		t.Errorf("expected 7000 input tokens (3000+4000), got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 650 {
		t.Errorf("expected 650 output tokens (150+500), got %d", stats.OutputTokens)
	}
	if stats.CacheReadTokens != 15000 {
		t.Errorf("expected 15000 cache read tokens, got %d", stats.CacheReadTokens)
	}

	// Check metadata
	if stats.Model != "claude-opus-4-1-20250805" {
		t.Errorf("expected model 'claude-opus-4-1-20250805', got '%s'", stats.Model)
	}
	if stats.GitBranch != "feature/auth" {
		t.Errorf("expected branch 'feature/auth', got '%s'", stats.GitBranch)
	}
	if stats.ClaudeVersion != "2.0.76" {
		t.Errorf("expected version '2.0.76', got '%s'", stats.ClaudeVersion)
	}
	if stats.Summary != "Add login validation" {
		t.Errorf("expected summary 'Add login validation', got '%s'", stats.Summary)
	}

	// Check tool usage
	if stats.ToolCalls != 2 {
		t.Errorf("expected 2 tool calls, got %d", stats.ToolCalls)
	}
	if stats.ToolsBreakdown["Read"] != 1 {
		t.Errorf("expected 1 Read call, got %d", stats.ToolsBreakdown["Read"])
	}
	if stats.ToolsBreakdown["Edit"] != 1 {
		t.Errorf("expected 1 Edit call, got %d", stats.ToolsBreakdown["Edit"])
	}

	// Check files
	if len(stats.FilesAccessed) != 1 {
		t.Errorf("expected 1 file accessed, got %d", len(stats.FilesAccessed))
	}
	if len(stats.FilesModified) != 1 {
		t.Errorf("expected 1 file modified, got %d", len(stats.FilesModified))
	}

	// Check timestamps
	if stats.StartTime == nil || stats.EndTime == nil {
		t.Fatal("expected start and end times to be set")
	}
	if stats.Duration() != 60 {
		t.Errorf("expected duration 60 seconds, got %d", stats.Duration())
	}
}

func TestParseReader_SkipsFileHistorySnapshot(t *testing.T) {
	// Real transcripts include file-history-snapshot entries that should be ignored
	input := `{"type":"user","timestamp":"2025-01-15T10:00:00Z","message":{"role":"user","content":"Hello"}}
{"type":"file-history-snapshot","messageId":"msg-001","snapshot":{"trackedFileBackups":{}},"isSnapshotUpdate":false}
{"type":"assistant","timestamp":"2025-01-15T10:00:05Z","message":{"model":"claude-3","usage":{"input_tokens":100,"output_tokens":50},"content":[{"type":"text","text":"Hi!"}]}}`

	p := NewParser()
	result, err := p.ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only count the user and assistant messages
	if result.Statistics.UserPrompts != 1 {
		t.Errorf("expected 1 user prompt, got %d", result.Statistics.UserPrompts)
	}
	if result.Statistics.AssistantResponses != 1 {
		t.Errorf("expected 1 assistant response, got %d", result.Statistics.AssistantResponses)
	}
}

func TestExtractTextContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    `"Hello world"`,
			expected: "Hello world",
		},
		{
			name:     "array with text item",
			input:    `[{"type":"text","text":"Hello from array"}]`,
			expected: "Hello from array",
		},
		{
			name:     "array with multiple items",
			input:    `[{"type":"thinking","thinking":"hmm"},{"type":"text","text":"Response text"}]`,
			expected: "Response text",
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: "",
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
		},
		{
			name:     "null",
			input:    `null`,
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractTextContent([]byte(tc.input))
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
