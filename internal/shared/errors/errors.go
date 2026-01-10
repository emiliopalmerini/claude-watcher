package errors

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
)

// AppError represents an application error with HTTP status code.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Common error constructors

// NotFound creates a 404 error.
func NotFound(message string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: message}
}

// InternalError creates a 500 error.
func InternalError(err error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Err: err}
}

// BadRequest creates a 400 error.
func BadRequest(message string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: message}
}

// HandleError writes an appropriate HTTP error response.
// It logs internal errors and returns a generic message to clients.
func HandleError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Code >= 500 {
			log.Printf("internal error: %v", appErr.Err)
			http.Error(w, "internal server error", appErr.Code)
			return
		}
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	// Handle common database errors
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Default to internal server error
	log.Printf("unhandled error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

// HandleDBError wraps a database error with appropriate HTTP status.
func HandleDBError(err error, notFoundMsg string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return NotFound(notFoundMsg)
	}
	return InternalError(err)
}
