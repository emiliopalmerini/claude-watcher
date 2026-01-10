package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"claude-watcher/internal/database/sqlc"
)

func TestHandler_GetChartData(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockRepo       *MockRepository
		expectedStatus int
		expectedRange  string
	}{
		{
			name:        "success with default range",
			queryParams: "",
			mockRepo: &MockRepository{
				GetDailyMetricsFunc: func(ctx context.Context, days sql.NullString) ([]sqlc.GetDailyMetricsRow, error) {
					return []sqlc.GetDailyMetricsRow{
						{Period: "2024-01-01", Sessions: 10, Cost: float64(5.0)},
					}, nil
				},
				GetModelDistributionFunc: func(ctx context.Context, hours sql.NullString) ([]sqlc.GetModelDistributionRow, error) {
					return []sqlc.GetModelDistributionRow{
						{Model: "claude-3-opus", Sessions: 5, Cost: float64(2.5)},
					}, nil
				},
				GetHourOfDayDistributionFunc: func(ctx context.Context, hours sql.NullString) ([]sqlc.GetHourOfDayDistributionRow, error) {
					return []sqlc.GetHourOfDayDistributionRow{
						{Hour: 14, Sessions: 3, Cost: float64(1.0)},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedRange:  "7d",
		},
		{
			name:        "success with 24h range",
			queryParams: "?range=24h",
			mockRepo: &MockRepository{
				GetDailyMetricsFunc: func(ctx context.Context, days sql.NullString) ([]sqlc.GetDailyMetricsRow, error) {
					return []sqlc.GetDailyMetricsRow{}, nil
				},
				GetModelDistributionFunc: func(ctx context.Context, hours sql.NullString) ([]sqlc.GetModelDistributionRow, error) {
					return []sqlc.GetModelDistributionRow{}, nil
				},
				GetHourOfDayDistributionFunc: func(ctx context.Context, hours sql.NullString) ([]sqlc.GetHourOfDayDistributionRow, error) {
					return []sqlc.GetHourOfDayDistributionRow{}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedRange:  "24h",
		},
		{
			name:        "success with 30d range",
			queryParams: "?range=30d",
			mockRepo: &MockRepository{
				GetDailyMetricsFunc: func(ctx context.Context, days sql.NullString) ([]sqlc.GetDailyMetricsRow, error) {
					return []sqlc.GetDailyMetricsRow{}, nil
				},
				GetModelDistributionFunc: func(ctx context.Context, hours sql.NullString) ([]sqlc.GetModelDistributionRow, error) {
					return []sqlc.GetModelDistributionRow{}, nil
				},
				GetHourOfDayDistributionFunc: func(ctx context.Context, hours sql.NullString) ([]sqlc.GetHourOfDayDistributionRow, error) {
					return []sqlc.GetHourOfDayDistributionRow{}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedRange:  "30d",
		},
		{
			name:        "partial failure - daily metrics error",
			queryParams: "",
			mockRepo: &MockRepository{
				GetDailyMetricsFunc: func(ctx context.Context, days sql.NullString) ([]sqlc.GetDailyMetricsRow, error) {
					return nil, errors.New("database error")
				},
				GetModelDistributionFunc: func(ctx context.Context, hours sql.NullString) ([]sqlc.GetModelDistributionRow, error) {
					return []sqlc.GetModelDistributionRow{}, nil
				},
				GetHourOfDayDistributionFunc: func(ctx context.Context, hours sql.NullString) ([]sqlc.GetHourOfDayDistributionRow, error) {
					return []sqlc.GetHourOfDayDistributionRow{}, nil
				},
			},
			expectedStatus: http.StatusOK, // API returns partial data
			expectedRange:  "7d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mockRepo, 168) // 7 days default

			req := httptest.NewRequest(http.MethodGet, "/api/charts"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.GetChartData(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GetChartData() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("GetChartData() Content-Type = %q, want %q", contentType, "application/json")
			}

			// Parse response
			var data ChartData
			if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
				t.Fatalf("GetChartData() invalid JSON: %v", err)
			}

			if data.Range != tt.expectedRange {
				t.Errorf("GetChartData() Range = %q, want %q", data.Range, tt.expectedRange)
			}
		})
	}
}

func TestHandler_rangeToHours(t *testing.T) {
	handler := NewHandler(&MockRepository{}, 168)

	tests := []struct {
		input    string
		expected int
	}{
		{"24h", 24},
		{"7d", 168},
		{"30d", 720},
		{"90d", 2160},
		{"invalid", 168}, // default
		{"", 168},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := handler.rangeToHours(tt.input)
			if result != tt.expected {
				t.Errorf("rangeToHours(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewHandler_API(t *testing.T) {
	mockRepo := &MockRepository{}
	defaultHours := 168
	handler := NewHandler(mockRepo, defaultHours)

	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if handler.repo != mockRepo {
		t.Error("NewHandler() did not set repository correctly")
	}
	if handler.defaultRangeHours != defaultHours {
		t.Errorf("NewHandler() defaultRangeHours = %d, want %d", handler.defaultRangeHours, defaultHours)
	}
}

func TestChartDataStructure(t *testing.T) {
	data := ChartData{
		TimeSeries: []TimePoint{
			{Period: "2024-01-01", Sessions: 10, Cost: 5.0, Tokens: Tokens{Input: 1000, Output: 500, Thinking: 200}},
		},
		Models: []ModelPoint{
			{Model: "claude-3-opus", Sessions: 5, Cost: 2.5},
		},
		HourOfDay: []HourPoint{
			{Hour: 14, Sessions: 3, Cost: 1.0},
		},
		Range: "7d",
	}

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal ChartData: %v", err)
	}

	var decoded ChartData
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ChartData: %v", err)
	}

	if len(decoded.TimeSeries) != 1 {
		t.Errorf("TimeSeries length = %d, want 1", len(decoded.TimeSeries))
	}
	if len(decoded.Models) != 1 {
		t.Errorf("Models length = %d, want 1", len(decoded.Models))
	}
	if len(decoded.HourOfDay) != 1 {
		t.Errorf("HourOfDay length = %d, want 1", len(decoded.HourOfDay))
	}
	if decoded.Range != "7d" {
		t.Errorf("Range = %q, want %q", decoded.Range, "7d")
	}
}

func TestMapSlice(t *testing.T) {
	input := []int{1, 2, 3}
	result := mapSlice(input, func(i int) int { return i * 2 })

	expected := []int{2, 4, 6}
	if len(result) != len(expected) {
		t.Fatalf("mapSlice() length = %d, want %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("mapSlice()[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestAsFloat(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
	}{
		{float64(1.5), 1.5},
		{int64(10), 10.0},
		{"invalid", 0},
		{nil, 0},
	}

	for _, tt := range tests {
		result := asFloat(tt.input)
		if result != tt.expected {
			t.Errorf("asFloat(%v) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestAsInt(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int64
	}{
		{int64(10), 10},
		{float64(10.5), 10},
		{"invalid", 0},
		{nil, 0},
	}

	for _, tt := range tests {
		result := asInt(tt.input)
		if result != tt.expected {
			t.Errorf("asInt(%v) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestAsString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{[]byte("bytes"), "bytes"},
		{123, ""},
		{nil, ""},
	}

	for _, tt := range tests {
		result := asString(tt.input)
		if result != tt.expected {
			t.Errorf("asString(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
