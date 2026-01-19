package util

// ToInt64 safely converts an any to int64.
// Handles int64, int, and float64 types. Returns 0 for nil or unsupported types.
func ToInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}

// ToFloat64 safely converts an any to float64.
// Handles float64, int64, and int types. Returns 0 for nil or unsupported types.
func ToFloat64(v any) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}
