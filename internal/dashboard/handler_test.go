package dashboard

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"claude-watcher/internal/database/sqlc"
)

func TestHandler_Show(t *testing.T) {
	tests := []struct {
		name           string
		mockRepo       *MockRepository
		expectedStatus int
	}{
		{
			name: "success",
			mockRepo: &MockRepository{
				GetDashboardMetricsFunc: func(ctx context.Context) (sqlc.GetDashboardMetricsRow, error) {
					return sqlc.GetDashboardMetricsRow{
						TotalSessions: 100,
						TotalCostUsd:  float64(50.0),
					}, nil
				},
				GetTodayMetricsFunc: func(ctx context.Context) (sqlc.GetTodayMetricsRow, error) {
					return sqlc.GetTodayMetricsRow{
						SessionsToday: 10,
						CostToday:     float64(5.0),
					}, nil
				},
				GetWeekMetricsFunc: func(ctx context.Context) (sqlc.GetWeekMetricsRow, error) {
					return sqlc.GetWeekMetricsRow{
						SessionsWeek: 50,
						CostWeek:     float64(25.0),
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "dashboard metrics error",
			mockRepo: &MockRepository{
				GetDashboardMetricsFunc: func(ctx context.Context) (sqlc.GetDashboardMetricsRow, error) {
					return sqlc.GetDashboardMetricsRow{}, errors.New("database error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "today metrics error",
			mockRepo: &MockRepository{
				GetDashboardMetricsFunc: func(ctx context.Context) (sqlc.GetDashboardMetricsRow, error) {
					return sqlc.GetDashboardMetricsRow{TotalSessions: 100}, nil
				},
				GetTodayMetricsFunc: func(ctx context.Context) (sqlc.GetTodayMetricsRow, error) {
					return sqlc.GetTodayMetricsRow{}, errors.New("database error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "week metrics error",
			mockRepo: &MockRepository{
				GetDashboardMetricsFunc: func(ctx context.Context) (sqlc.GetDashboardMetricsRow, error) {
					return sqlc.GetDashboardMetricsRow{TotalSessions: 100}, nil
				},
				GetTodayMetricsFunc: func(ctx context.Context) (sqlc.GetTodayMetricsRow, error) {
					return sqlc.GetTodayMetricsRow{SessionsToday: 10}, nil
				},
				GetWeekMetricsFunc: func(ctx context.Context) (sqlc.GetWeekMetricsRow, error) {
					return sqlc.GetWeekMetricsRow{}, errors.New("database error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
			w := httptest.NewRecorder()

			handler.Show(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Show() status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	mockRepo := &MockRepository{}
	handler := NewHandler(mockRepo)

	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if handler.repo != mockRepo {
		t.Error("NewHandler() did not set repository correctly")
	}
}
