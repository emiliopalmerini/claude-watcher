package web

import (
	"database/sql"
	"math"
	"testing"
	"time"

	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func TestCalculateSuccessRate_BothValid(t *testing.T) {
	rate := calculateSuccessRate(
		sql.NullFloat64{Float64: 7, Valid: true},
		sql.NullFloat64{Float64: 3, Valid: true},
	)
	if rate == nil {
		t.Fatal("expected non-nil rate")
	}
	if math.Abs(*rate-0.7) > 0.001 {
		t.Errorf("expected 0.7, got %f", *rate)
	}
}

func TestCalculateSuccessRate_ZeroTotal(t *testing.T) {
	rate := calculateSuccessRate(
		sql.NullFloat64{Float64: 0, Valid: true},
		sql.NullFloat64{Float64: 0, Valid: true},
	)
	if rate != nil {
		t.Errorf("expected nil rate for zero total, got %f", *rate)
	}
}

func TestCalculateSuccessRate_NullValues(t *testing.T) {
	rate := calculateSuccessRate(
		sql.NullFloat64{Valid: false},
		sql.NullFloat64{Valid: false},
	)
	if rate != nil {
		t.Error("expected nil rate for null values")
	}
}

func TestCalculateSuccessRate_AllSuccess(t *testing.T) {
	rate := calculateSuccessRate(
		sql.NullFloat64{Float64: 5, Valid: true},
		sql.NullFloat64{Float64: 0, Valid: true},
	)
	if rate == nil {
		t.Fatal("expected non-nil rate")
	}
	if math.Abs(*rate-1.0) > 0.001 {
		t.Errorf("expected 1.0, got %f", *rate)
	}
}

func TestFormatChartDate_TimeType(t *testing.T) {
	tm := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	result := formatChartDate(tm)
	if result != "2024-06-15" {
		t.Errorf("expected 2024-06-15, got %s", result)
	}
}

func TestFormatChartDate_StringType(t *testing.T) {
	result := formatChartDate("2024-06-15")
	if result != "2024-06-15" {
		t.Errorf("expected 2024-06-15, got %s", result)
	}
}

func TestFormatChartDate_OtherType(t *testing.T) {
	result := formatChartDate(42)
	if result != "42" {
		t.Errorf("expected '42', got %s", result)
	}
}

func TestBuildSessionDetail_BasicFields(t *testing.T) {
	session := sqlc.Session{
		ID:             "sess-123",
		ProjectID:      "proj-456",
		Cwd:            "/home/test",
		PermissionMode: "default",
		ExitReason:     "exit",
		CreatedAt:      "2024-06-15T10:00:00Z",
	}

	detail := buildSessionDetail(session, nil)

	if detail.ID != "sess-123" {
		t.Errorf("expected sess-123, got %s", detail.ID)
	}
	if detail.ProjectID != "proj-456" {
		t.Errorf("expected proj-456, got %s", detail.ProjectID)
	}
	if detail.TurnCount != 0 {
		t.Errorf("expected 0 turns with nil metrics, got %d", detail.TurnCount)
	}
}

func TestBuildSessionDetail_WithMetrics(t *testing.T) {
	session := sqlc.Session{
		ID:             "sess-123",
		ProjectID:      "proj-456",
		Cwd:            "/home/test",
		PermissionMode: "default",
		ExitReason:     "exit",
		CreatedAt:      "2024-06-15T10:00:00Z",
	}
	metrics := &sqlc.SessionMetric{
		TurnCount:       5,
		TokenInput:      1000,
		TokenOutput:     500,
		ErrorCount:      1,
		CostEstimateUsd: sql.NullFloat64{Float64: 0.05, Valid: true},
	}

	detail := buildSessionDetail(session, metrics)

	if detail.TurnCount != 5 {
		t.Errorf("expected 5 turns, got %d", detail.TurnCount)
	}
	if detail.TokenInput != 1000 {
		t.Errorf("expected 1000 input tokens, got %d", detail.TokenInput)
	}
	if detail.CostEstimateUsd != 0.05 {
		t.Errorf("expected cost 0.05, got %f", detail.CostEstimateUsd)
	}
}

func TestBuildSessionDetail_NullableSessionFields(t *testing.T) {
	session := sqlc.Session{
		ID:              "sess-123",
		ProjectID:       "proj-456",
		Cwd:             "/home/test",
		PermissionMode:  "default",
		ExitReason:      "exit",
		CreatedAt:       "2024-06-15T10:00:00Z",
		ExperimentID:    sql.NullString{String: "exp-789", Valid: true},
		StartedAt:       sql.NullString{String: "2024-06-15T10:00:00Z", Valid: true},
		EndedAt:         sql.NullString{String: "2024-06-15T10:30:00Z", Valid: true},
		DurationSeconds: sql.NullInt64{Int64: 1800, Valid: true},
	}

	detail := buildSessionDetail(session, nil)

	if detail.ExperimentID != "exp-789" {
		t.Errorf("expected exp-789, got %s", detail.ExperimentID)
	}
	if detail.StartedAt != "2024-06-15T10:00:00Z" {
		t.Errorf("expected started_at, got %s", detail.StartedAt)
	}
	if detail.DurationSeconds != 1800 {
		t.Errorf("expected 1800s duration, got %d", detail.DurationSeconds)
	}
}
