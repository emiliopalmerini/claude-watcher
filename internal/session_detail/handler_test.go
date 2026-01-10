package session_detail

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"claude-watcher/internal/database/sqlc"
)

func TestHandler_Show(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		mockRepo       *MockRepository
		expectedStatus int
	}{
		{
			name:      "success",
			sessionID: "abc123",
			mockRepo: &MockRepository{
				GetSessionByIDFunc: func(ctx context.Context, sessionID string) (sqlc.Session, error) {
					if sessionID != "abc123" {
						t.Errorf("expected sessionID 'abc123', got %q", sessionID)
					}
					return sqlc.Session{
						ID:        1,
						SessionID: "abc123",
						Hostname:  "localhost",
						Timestamp: "2024-01-01T10:00:00Z",
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "not found",
			sessionID: "nonexistent",
			mockRepo: &MockRepository{
				GetSessionByIDFunc: func(ctx context.Context, sessionID string) (sqlc.Session, error) {
					return sqlc.Session{}, sql.ErrNoRows
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "database error",
			sessionID: "abc123",
			mockRepo: &MockRepository{
				GetSessionByIDFunc: func(ctx context.Context, sessionID string) (sqlc.Session, error) {
					return sqlc.Session{}, errors.New("connection failed")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mockRepo)

			// Create a chi router context with URL params
			r := chi.NewRouter()
			r.Get("/sessions/{sessionID}", handler.Show)

			req := httptest.NewRequest(http.MethodGet, "/sessions/"+tt.sessionID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Show() status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestNewHandler_SessionDetail(t *testing.T) {
	mockRepo := &MockRepository{}
	handler := NewHandler(mockRepo)

	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if handler.repo != mockRepo {
		t.Error("NewHandler() did not set repository correctly")
	}
}

func TestHandler_Show_SessionData(t *testing.T) {
	expectedSession := sqlc.Session{
		ID:               1,
		SessionID:        "test-session-123",
		InstanceID:       "instance-1",
		Hostname:         "test-host",
		Timestamp:        "2024-01-15T14:30:00Z",
		ExitReason:       sql.NullString{String: "user_exit", Valid: true},
		WorkingDirectory: sql.NullString{String: "/home/user/project", Valid: true},
		GitBranch:        sql.NullString{String: "feature/test", Valid: true},
		DurationSeconds:  sql.NullInt64{Int64: 3600, Valid: true},
		UserPrompts:      sql.NullInt64{Int64: 25, Valid: true},
		ToolCalls:        sql.NullInt64{Int64: 100, Valid: true},
		InputTokens:      sql.NullInt64{Int64: 50000, Valid: true},
		OutputTokens:     sql.NullInt64{Int64: 25000, Valid: true},
		EstimatedCostUsd: sql.NullFloat64{Float64: 1.5, Valid: true},
		Model:            sql.NullString{String: "claude-3-opus", Valid: true},
	}

	mockRepo := &MockRepository{
		GetSessionByIDFunc: func(ctx context.Context, sessionID string) (sqlc.Session, error) {
			return expectedSession, nil
		},
	}

	handler := NewHandler(mockRepo)

	r := chi.NewRouter()
	r.Get("/sessions/{sessionID}", handler.Show)

	req := httptest.NewRequest(http.MethodGet, "/sessions/test-session-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Show() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify the response contains expected data
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Show() returned empty body")
	}
}
