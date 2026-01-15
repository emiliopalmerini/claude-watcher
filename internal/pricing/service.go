package pricing

import "strings"

// GetModelPricing returns pricing for a model, with fallback to default
func GetModelPricing(model string) ModelPricing {
	if model == "" {
		return defaultPricing
	}

	// Try exact match first
	if pricing, ok := modelPricing[model]; ok {
		return pricing
	}

	// Try prefix match (e.g., "claude-opus-4-20241022" -> "claude-opus-4")
	for key, pricing := range modelPricing {
		if strings.HasPrefix(model, key) || strings.Contains(model, key) {
			return pricing
		}
	}

	return defaultPricing
}

// CalculateCost calculates estimated API cost based on token usage
func CalculateCost(usage TokenUsage) float64 {
	pricing := GetModelPricing(usage.Model)

	inputCost := (float64(usage.InputTokens) / 1_000_000) * pricing.Input
	outputCost := (float64(usage.OutputTokens) / 1_000_000) * pricing.Output
	cacheReadCost := (float64(usage.CacheReadTokens) / 1_000_000) * pricing.CacheRead
	cacheWriteCost := (float64(usage.CacheWriteTokens) / 1_000_000) * pricing.CacheWrite

	total := inputCost + outputCost + cacheReadCost + cacheWriteCost

	// Round to 6 decimal places
	return float64(int(total*1_000_000)) / 1_000_000
}
