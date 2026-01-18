package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
)

// ViewerMessage represents a single message for display in the transcript viewer
type ViewerMessage struct {
	Role      string
	Content   string
	Timestamp string
	Tools     []ViewerToolUse
}

// ViewerToolUse represents a tool invocation for display
type ViewerToolUse struct {
	Name  string
	Input string
}

// viewerEntry is a flexible entry struct for the viewer that handles
// user messages with string content (vs array content for assistant)
type viewerEntry struct {
	Type      string         `json:"type"`
	Timestamp string         `json:"timestamp,omitempty"`
	Message   *viewerMessage `json:"message,omitempty"`
}

type viewerMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"` // Can be string or []Content
}

// ParseTranscriptForViewer parses JSONL bytes into a viewer-friendly format.
// This is separate from ParseTranscript to keep concerns separate:
// - ParseTranscript: extracts metrics during recording
// - ParseTranscriptForViewer: renders transcript for display
// Consecutive messages from the same role are merged into a single message.
func ParseTranscriptForViewer(data []byte) ([]ViewerMessage, error) {
	var messages []ViewerMessage
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Increase buffer for large lines
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry viewerEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		switch entry.Type {
		case "user", "human":
			content := ""
			if entry.Message != nil {
				content = extractContentFlexible(entry.Message.Content)
			}
			if content != "" {
				// Merge with previous message if same role
				if len(messages) > 0 && messages[len(messages)-1].Role == "user" {
					messages[len(messages)-1].Content += "\n\n" + content
				} else {
					messages = append(messages, ViewerMessage{
						Role:      "user",
						Timestamp: entry.Timestamp,
						Content:   content,
					})
				}
			}

		case "assistant":
			content := ""
			var tools []ViewerToolUse
			if entry.Message != nil {
				content = extractContentFlexible(entry.Message.Content)
				tools = extractToolUsesFromRaw(entry.Message.Content)
			}
			// Include message if it has content or tool uses
			if content != "" || len(tools) > 0 {
				// Merge with previous message if same role
				if len(messages) > 0 && messages[len(messages)-1].Role == "assistant" {
					prev := &messages[len(messages)-1]
					if content != "" {
						if prev.Content != "" {
							prev.Content += "\n\n" + content
						} else {
							prev.Content = content
						}
					}
					prev.Tools = append(prev.Tools, tools...)
				} else {
					messages = append(messages, ViewerMessage{
						Role:      "assistant",
						Timestamp: entry.Timestamp,
						Content:   content,
						Tools:     tools,
					})
				}
			}
		}
	}

	return messages, scanner.Err()
}

// extractContentFlexible handles content that can be either a string or []Content
func extractContentFlexible(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try as string first (common for user messages)
	var strContent string
	if err := json.Unmarshal(raw, &strContent); err == nil {
		return strContent
	}

	// Try as array of Content objects (common for assistant messages)
	var contentArray []Content
	if err := json.Unmarshal(raw, &contentArray); err == nil {
		return extractTextContent(contentArray)
	}

	return ""
}

// extractToolUsesFromRaw extracts tool uses from raw content JSON
func extractToolUsesFromRaw(raw json.RawMessage) []ViewerToolUse {
	if len(raw) == 0 {
		return nil
	}

	var contentArray []Content
	if err := json.Unmarshal(raw, &contentArray); err != nil {
		return nil
	}

	return extractToolUses(contentArray)
}

func extractTextContent(content []Content) string {
	var parts []string
	for _, c := range content {
		if c.Type == "text" && c.Text != "" {
			parts = append(parts, c.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func extractToolUses(content []Content) []ViewerToolUse {
	var tools []ViewerToolUse
	for _, c := range content {
		if c.Type == "tool_use" && c.Name != "" {
			input := ""
			if len(c.Input) > 0 {
				// Pretty-print JSON, truncate if too long
				var v any
				if json.Unmarshal(c.Input, &v) == nil {
					if b, err := json.MarshalIndent(v, "", "  "); err == nil {
						input = string(b)
						if len(input) > 2000 {
							input = input[:2000] + "\n..."
						}
					}
				}
			}
			tools = append(tools, ViewerToolUse{
				Name:  c.Name,
				Input: input,
			})
		}
	}
	return tools
}
