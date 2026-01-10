package format

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// NullStr formats a sql.NullString, returning "-" for invalid values.
func NullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "-"
}

// NullInt formats a sql.NullInt64 as a string, returning "0" for invalid values.
func NullInt(i sql.NullInt64) string {
	if i.Valid {
		return fmt.Sprintf("%d", i.Int64)
	}
	return "0"
}

// NullIntFormatted formats a sql.NullInt64 with K/M suffixes for large numbers.
func NullIntFormatted(i sql.NullInt64) string {
	if i.Valid {
		return FormatNumber(i.Int64)
	}
	return "0"
}

// FormatNumber formats an int64 with K/M suffixes for large numbers.
func FormatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

// NullCost formats a sql.NullFloat64 as currency with 2 decimal places.
func NullCost(f sql.NullFloat64) string {
	if f.Valid {
		return fmt.Sprintf("$%.2f", f.Float64)
	}
	return "$0.00"
}

// NullCostPrecise formats a sql.NullFloat64 as currency with 4 decimal places.
func NullCostPrecise(f sql.NullFloat64) string {
	if f.Valid {
		return fmt.Sprintf("$%.4f", f.Float64)
	}
	return "$0.00"
}

// Duration formats a sql.NullInt64 representing seconds as "Xm Ys" format.
func Duration(d sql.NullInt64) string {
	if !d.Valid {
		return "-"
	}
	mins := d.Int64 / 60
	secs := d.Int64 % 60
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// ShortID truncates an ID string to 8 characters.
func ShortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// ParseToolsJSON parses a JSON string of tool counts into a map.
func ParseToolsJSON(s sql.NullString) map[string]int {
	result := make(map[string]int)
	if !s.Valid || s.String == "" {
		return result
	}
	json.Unmarshal([]byte(s.String), &result)
	return result
}
