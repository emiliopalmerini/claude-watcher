package pricing

// ModelPricing represents pricing per million tokens for a Claude model
type ModelPricing struct {
	Input      float64
	Output     float64
	CacheRead  float64
	CacheWrite float64
}

// TokenUsage represents token counts for cost calculation
type TokenUsage struct {
	Model            string
	InputTokens      int
	OutputTokens     int
	CacheReadTokens  int
	CacheWriteTokens int
}

// modelPricing maps model names to their pricing
var modelPricing = map[string]ModelPricing{
	// Current models (January 2026)
	"claude-opus-4-5":   {Input: 5.00, Output: 25.00, CacheRead: 0.50, CacheWrite: 6.25},
	"claude-sonnet-4-5": {Input: 3.00, Output: 15.00, CacheRead: 0.30, CacheWrite: 3.75},
	"claude-haiku-4-5":  {Input: 1.00, Output: 5.00, CacheRead: 0.10, CacheWrite: 1.25},
	// Legacy models
	"claude-opus-4-1":   {Input: 15.00, Output: 75.00, CacheRead: 1.50, CacheWrite: 18.75},
	"claude-sonnet-4":   {Input: 3.00, Output: 15.00, CacheRead: 0.30, CacheWrite: 3.75},
	"claude-opus-4":     {Input: 15.00, Output: 75.00, CacheRead: 1.50, CacheWrite: 18.75},
	"claude-3-5-sonnet": {Input: 3.00, Output: 15.00, CacheRead: 0.30, CacheWrite: 3.75},
	"claude-3-5-haiku":  {Input: 0.80, Output: 4.00, CacheRead: 0.08, CacheWrite: 1.00},
	"claude-3-haiku":    {Input: 0.25, Output: 1.25, CacheRead: 0.03, CacheWrite: 0.30},
}

var defaultPricing = modelPricing["claude-opus-4-5"]
