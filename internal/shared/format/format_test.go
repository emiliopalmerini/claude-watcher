package format

import (
	"database/sql"
	"testing"
)

func TestNullStr(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected string
	}{
		{
			name:     "valid string",
			input:    sql.NullString{String: "hello", Valid: true},
			expected: "hello",
		},
		{
			name:     "empty valid string",
			input:    sql.NullString{String: "", Valid: true},
			expected: "",
		},
		{
			name:     "invalid string",
			input:    sql.NullString{String: "", Valid: false},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NullStr(tt.input)
			if result != tt.expected {
				t.Errorf("NullStr(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNullInt(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullInt64
		expected string
	}{
		{
			name:     "valid positive int",
			input:    sql.NullInt64{Int64: 42, Valid: true},
			expected: "42",
		},
		{
			name:     "valid zero",
			input:    sql.NullInt64{Int64: 0, Valid: true},
			expected: "0",
		},
		{
			name:     "valid negative int",
			input:    sql.NullInt64{Int64: -10, Valid: true},
			expected: "-10",
		},
		{
			name:     "invalid int",
			input:    sql.NullInt64{Int64: 0, Valid: false},
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NullInt(tt.input)
			if result != tt.expected {
				t.Errorf("NullInt(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "small number",
			input:    42,
			expected: "42",
		},
		{
			name:     "just under 1K",
			input:    999,
			expected: "999",
		},
		{
			name:     "exactly 1K",
			input:    1000,
			expected: "1.0K",
		},
		{
			name:     "thousands",
			input:    1500,
			expected: "1.5K",
		},
		{
			name:     "tens of thousands",
			input:    45000,
			expected: "45.0K",
		},
		{
			name:     "just under 1M",
			input:    999999,
			expected: "1000.0K",
		},
		{
			name:     "exactly 1M",
			input:    1000000,
			expected: "1.0M",
		},
		{
			name:     "millions",
			input:    2500000,
			expected: "2.5M",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNullIntFormatted(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullInt64
		expected string
	}{
		{
			name:     "valid small number",
			input:    sql.NullInt64{Int64: 500, Valid: true},
			expected: "500",
		},
		{
			name:     "valid thousands",
			input:    sql.NullInt64{Int64: 5000, Valid: true},
			expected: "5.0K",
		},
		{
			name:     "invalid",
			input:    sql.NullInt64{Int64: 0, Valid: false},
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NullIntFormatted(tt.input)
			if result != tt.expected {
				t.Errorf("NullIntFormatted(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNullCost(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullFloat64
		expected string
	}{
		{
			name:     "valid cost",
			input:    sql.NullFloat64{Float64: 1.50, Valid: true},
			expected: "$1.50",
		},
		{
			name:     "valid zero",
			input:    sql.NullFloat64{Float64: 0, Valid: true},
			expected: "$0.00",
		},
		{
			name:     "valid small cost",
			input:    sql.NullFloat64{Float64: 0.05, Valid: true},
			expected: "$0.05",
		},
		{
			name:     "invalid cost",
			input:    sql.NullFloat64{Float64: 0, Valid: false},
			expected: "$0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NullCost(tt.input)
			if result != tt.expected {
				t.Errorf("NullCost(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNullCostPrecise(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullFloat64
		expected string
	}{
		{
			name:     "valid cost",
			input:    sql.NullFloat64{Float64: 1.5678, Valid: true},
			expected: "$1.5678",
		},
		{
			name:     "valid small cost",
			input:    sql.NullFloat64{Float64: 0.0001, Valid: true},
			expected: "$0.0001",
		},
		{
			name:     "invalid cost",
			input:    sql.NullFloat64{Float64: 0, Valid: false},
			expected: "$0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NullCostPrecise(tt.input)
			if result != tt.expected {
				t.Errorf("NullCostPrecise(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullInt64
		expected string
	}{
		{
			name:     "seconds only",
			input:    sql.NullInt64{Int64: 45, Valid: true},
			expected: "45s",
		},
		{
			name:     "minutes and seconds",
			input:    sql.NullInt64{Int64: 125, Valid: true},
			expected: "2m 5s",
		},
		{
			name:     "exactly one minute",
			input:    sql.NullInt64{Int64: 60, Valid: true},
			expected: "1m 0s",
		},
		{
			name:     "zero seconds",
			input:    sql.NullInt64{Int64: 0, Valid: true},
			expected: "0s",
		},
		{
			name:     "invalid duration",
			input:    sql.NullInt64{Int64: 0, Valid: false},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Duration(tt.input)
			if result != tt.expected {
				t.Errorf("Duration(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestShortID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "long id",
			input:    "abcdefghijklmnop",
			expected: "abcdefgh",
		},
		{
			name:     "exactly 8 chars",
			input:    "abcdefgh",
			expected: "abcdefgh",
		},
		{
			name:     "short id",
			input:    "abc",
			expected: "abc",
		},
		{
			name:     "empty id",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShortID(tt.input)
			if result != tt.expected {
				t.Errorf("ShortID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseToolsJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected map[string]int
	}{
		{
			name:     "valid json",
			input:    sql.NullString{String: `{"Read":5,"Write":3}`, Valid: true},
			expected: map[string]int{"Read": 5, "Write": 3},
		},
		{
			name:     "empty json object",
			input:    sql.NullString{String: `{}`, Valid: true},
			expected: map[string]int{},
		},
		{
			name:     "invalid null string",
			input:    sql.NullString{String: "", Valid: false},
			expected: map[string]int{},
		},
		{
			name:     "empty string",
			input:    sql.NullString{String: "", Valid: true},
			expected: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseToolsJSON(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseToolsJSON(%v) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("ParseToolsJSON(%v)[%q] = %d, want %d", tt.input, k, result[k], v)
				}
			}
		})
	}
}
