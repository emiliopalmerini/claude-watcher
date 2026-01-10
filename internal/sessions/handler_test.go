package sessions

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"claude-watcher/internal/database/sqlc"
)

func TestHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockRepo       *MockRepository
		pageSize       int64
		expectedStatus int
	}{
		{
			name:        "success first page",
			queryParams: "",
			mockRepo: &MockRepository{
				ListSessionsFunc: func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
					if params.Offset != 0 {
						t.Errorf("expected offset 0, got %d", params.Offset)
					}
					if params.Limit != 20 {
						t.Errorf("expected limit 20, got %d", params.Limit)
					}
					return []sqlc.ListSessionsRow{
						{
							ID:        1,
							SessionID: "abc123",
							Hostname:  "localhost",
							Timestamp: "2024-01-01",
						},
					}, nil
				},
				CountSessionsFunc: func(ctx context.Context) (int64, error) {
					return 100, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "success second page",
			queryParams: "?page=2",
			mockRepo: &MockRepository{
				ListSessionsFunc: func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
					if params.Offset != 20 {
						t.Errorf("expected offset 20, got %d", params.Offset)
					}
					return []sqlc.ListSessionsRow{}, nil
				},
				CountSessionsFunc: func(ctx context.Context) (int64, error) {
					return 100, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid page defaults to 1",
			queryParams: "?page=0",
			mockRepo: &MockRepository{
				ListSessionsFunc: func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
					if params.Offset != 0 {
						t.Errorf("expected offset 0, got %d", params.Offset)
					}
					return []sqlc.ListSessionsRow{}, nil
				},
				CountSessionsFunc: func(ctx context.Context) (int64, error) {
					return 0, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "negative page defaults to 1",
			queryParams: "?page=-5",
			mockRepo: &MockRepository{
				ListSessionsFunc: func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
					if params.Offset != 0 {
						t.Errorf("expected offset 0, got %d", params.Offset)
					}
					return []sqlc.ListSessionsRow{}, nil
				},
				CountSessionsFunc: func(ctx context.Context) (int64, error) {
					return 0, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "list sessions error",
			queryParams: "",
			mockRepo: &MockRepository{
				ListSessionsFunc: func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
					return nil, errors.New("database error")
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "count sessions error",
			queryParams: "",
			mockRepo: &MockRepository{
				ListSessionsFunc: func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
					return []sqlc.ListSessionsRow{}, nil
				},
				CountSessionsFunc: func(ctx context.Context) (int64, error) {
					return 0, errors.New("database error")
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "custom page size",
			queryParams: "",
			mockRepo: &MockRepository{
				ListSessionsFunc: func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
					if params.Limit != 10 {
						t.Errorf("expected limit 10, got %d", params.Limit)
					}
					return []sqlc.ListSessionsRow{}, nil
				},
				CountSessionsFunc: func(ctx context.Context) (int64, error) {
					return 50, nil
				},
			},
			pageSize:       10,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mockRepo, tt.pageSize)

			req := httptest.NewRequest(http.MethodGet, "/sessions"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("List() status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestHandler_List_Pagination(t *testing.T) {
	mockRepo := &MockRepository{
		ListSessionsFunc: func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
			return []sqlc.ListSessionsRow{
				{ID: 1, SessionID: "sess1", Hostname: "host1", Timestamp: "2024-01-01"},
			}, nil
		},
		CountSessionsFunc: func(ctx context.Context) (int64, error) {
			return 45, nil
		},
	}

	handler := NewHandler(mockRepo, 20)

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestNewHandler_Sessions(t *testing.T) {
	mockRepo := &MockRepository{}
	pageSize := int64(25)
	handler := NewHandler(mockRepo, pageSize)

	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if handler.repo != mockRepo {
		t.Error("NewHandler() did not set repository correctly")
	}
	if handler.pageSize != pageSize {
		t.Errorf("NewHandler() pageSize = %d, want %d", handler.pageSize, pageSize)
	}
}

func TestSessionsData(t *testing.T) {
	data := SessionsData{
		Sessions: []sqlc.ListSessionsRow{
			{
				ID:               1,
				SessionID:        "abc123",
				Hostname:         "localhost",
				Timestamp:        "2024-01-01",
				GitBranch:        sql.NullString{String: "main", Valid: true},
				DurationSeconds:  sql.NullInt64{Int64: 120, Valid: true},
				UserPrompts:      sql.NullInt64{Int64: 5, Valid: true},
				ToolCalls:        sql.NullInt64{Int64: 10, Valid: true},
				EstimatedCostUsd: sql.NullFloat64{Float64: 0.50, Valid: true},
			},
		},
		Page:       1,
		TotalPages: 5,
	}

	if len(data.Sessions) != 1 {
		t.Errorf("SessionsData.Sessions length = %d, want 1", len(data.Sessions))
	}
	if data.Page != 1 {
		t.Errorf("SessionsData.Page = %d, want 1", data.Page)
	}
	if data.TotalPages != 5 {
		t.Errorf("SessionsData.TotalPages = %d, want 5", data.TotalPages)
	}
}
