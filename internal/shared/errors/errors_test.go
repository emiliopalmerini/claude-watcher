package errors

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name:     "with wrapped error",
			appErr:   &AppError{Code: 500, Err: errors.New("database error")},
			expected: "database error",
		},
		{
			name:     "with message only",
			appErr:   &AppError{Code: 404, Message: "not found"},
			expected: "not found",
		},
		{
			name:     "with both message and error",
			appErr:   &AppError{Code: 500, Message: "fallback", Err: errors.New("wrapped")},
			expected: "wrapped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appErr.Error(); got != tt.expected {
				t.Errorf("AppError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := &AppError{Code: 500, Err: originalErr}

	if !errors.Is(appErr, originalErr) {
		t.Error("AppError should unwrap to original error")
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("resource not found")

	if err.Code != http.StatusNotFound {
		t.Errorf("NotFound().Code = %d, want %d", err.Code, http.StatusNotFound)
	}
	if err.Message != "resource not found" {
		t.Errorf("NotFound().Message = %q, want %q", err.Message, "resource not found")
	}
}

func TestInternalError(t *testing.T) {
	originalErr := errors.New("db connection failed")
	err := InternalError(originalErr)

	if err.Code != http.StatusInternalServerError {
		t.Errorf("InternalError().Code = %d, want %d", err.Code, http.StatusInternalServerError)
	}
	if err.Err != originalErr {
		t.Error("InternalError should wrap the original error")
	}
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("invalid input")

	if err.Code != http.StatusBadRequest {
		t.Errorf("BadRequest().Code = %d, want %d", err.Code, http.StatusBadRequest)
	}
	if err.Message != "invalid input" {
		t.Errorf("BadRequest().Message = %q, want %q", err.Message, "invalid input")
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedBody string
		bodyContains bool
	}{
		{
			name:         "app error 404",
			err:          NotFound("session not found"),
			expectedCode: http.StatusNotFound,
			expectedBody: "session not found",
			bodyContains: false,
		},
		{
			name:         "app error 400",
			err:          BadRequest("invalid page"),
			expectedCode: http.StatusBadRequest,
			expectedBody: "invalid page",
			bodyContains: false,
		},
		{
			name:         "app error 500 hides details",
			err:          InternalError(errors.New("secret db error")),
			expectedCode: http.StatusInternalServerError,
			expectedBody: "internal server error",
			bodyContains: true,
		},
		{
			name:         "sql.ErrNoRows",
			err:          sql.ErrNoRows,
			expectedCode: http.StatusNotFound,
			expectedBody: "not found",
			bodyContains: true,
		},
		{
			name:         "generic error",
			err:          errors.New("unknown error"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: "internal server error",
			bodyContains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(w, tt.err)

			if w.Code != tt.expectedCode {
				t.Errorf("HandleError() status = %d, want %d", w.Code, tt.expectedCode)
			}

			body := strings.TrimSpace(w.Body.String())
			if tt.bodyContains {
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("HandleError() body = %q, want to contain %q", body, tt.expectedBody)
				}
			} else {
				if body != tt.expectedBody {
					t.Errorf("HandleError() body = %q, want %q", body, tt.expectedBody)
				}
			}
		})
	}
}

func TestHandleDBError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		notFoundMsg  string
		expectedNil  bool
		expectedCode int
	}{
		{
			name:        "nil error",
			err:         nil,
			notFoundMsg: "not found",
			expectedNil: true,
		},
		{
			name:         "sql.ErrNoRows",
			err:          sql.ErrNoRows,
			notFoundMsg:  "session not found",
			expectedNil:  false,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "other error",
			err:          errors.New("connection failed"),
			notFoundMsg:  "not found",
			expectedNil:  false,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleDBError(tt.err, tt.notFoundMsg)

			if tt.expectedNil {
				if result != nil {
					t.Errorf("HandleDBError() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("HandleDBError() = nil, want error")
			}

			var appErr *AppError
			if !errors.As(result, &appErr) {
				t.Fatal("HandleDBError() should return *AppError")
			}

			if appErr.Code != tt.expectedCode {
				t.Errorf("HandleDBError().Code = %d, want %d", appErr.Code, tt.expectedCode)
			}
		})
	}
}
