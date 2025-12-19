package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a machine-readable error code for frontend handling
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrCodeConflict      ErrorCode = "CONFLICT"
	ErrCodeBadRequest    ErrorCode = "BAD_REQUEST"
	ErrCodeRateLimited   ErrorCode = "RATE_LIMITED"

	// Server errors (5xx)
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase      ErrorCode = "DATABASE_ERROR"
	ErrCodeCache         ErrorCode = "CACHE_ERROR"
	ErrCodeExternal      ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
)

// AppError represents an application-level error with HTTP status mapping
type AppError struct {
	Code       ErrorCode              // Machine-readable error code
	Message    string                 // Human-readable message
	StatusCode int                    // HTTP status code
	Internal   error                  // Internal error (not exposed to client)
	Fields     map[string]string      // Field-level validation errors
	Meta       map[string]interface{} // Additional metadata
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the internal error for error chain support
func (e *AppError) Unwrap() error {
	return e.Internal
}

// WithField adds a field-level error
func (e *AppError) WithField(field, message string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	e.Fields[field] = message
	return e
}

// WithMeta adds metadata to the error
func (e *AppError) WithMeta(key string, value interface{}) *AppError {
	if e.Meta == nil {
		e.Meta = make(map[string]interface{})
	}
	e.Meta[key] = value
	return e
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: codeToStatus(code),
	}
}

// Wrap wraps an internal error with an AppError
func Wrap(code ErrorCode, message string, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: codeToStatus(code),
		Internal:   internal,
	}
}

// codeToStatus maps error codes to HTTP status codes
func codeToStatus(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeRateLimited:
		return http.StatusTooManyRequests
	case ErrCodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Predefined errors for common scenarios
var (
	ErrNotFound          = New(ErrCodeNotFound, "Resource not found")
	ErrUnauthorized      = New(ErrCodeUnauthorized, "Authentication required")
	ErrForbidden         = New(ErrCodeForbidden, "Access denied")
	ErrValidation        = New(ErrCodeValidation, "Validation failed")
	ErrInternal          = New(ErrCodeInternal, "Internal server error")
	ErrDatabaseQuery     = New(ErrCodeDatabase, "Database query failed")
	ErrDatabaseConnect   = New(ErrCodeDatabase, "Database connection failed")
	ErrCacheOperation    = New(ErrCodeCache, "Cache operation failed")
	ErrExternalService   = New(ErrCodeExternal, "External service error")
	ErrConflict          = New(ErrCodeConflict, "Resource conflict")
	ErrBadRequest        = New(ErrCodeBadRequest, "Invalid request")
	ErrRateLimited       = New(ErrCodeRateLimited, "Rate limit exceeded")
	ErrServiceUnavailable = New(ErrCodeUnavailable, "Service temporarily unavailable")
)

// NotFoundf creates a not found error with formatted message
func NotFoundf(format string, args ...interface{}) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf(format, args...))
}

// Unauthorizedf creates an unauthorized error with formatted message
func Unauthorizedf(format string, args ...interface{}) *AppError {
	return New(ErrCodeUnauthorized, fmt.Sprintf(format, args...))
}

// Forbiddenf creates a forbidden error with formatted message
func Forbiddenf(format string, args ...interface{}) *AppError {
	return New(ErrCodeForbidden, fmt.Sprintf(format, args...))
}

// Validationf creates a validation error with formatted message
func Validationf(format string, args ...interface{}) *AppError {
	return New(ErrCodeValidation, fmt.Sprintf(format, args...))
}

// Internalf creates an internal error with formatted message and wraps internal error
func Internalf(internal error, format string, args ...interface{}) *AppError {
	return Wrap(ErrCodeInternal, fmt.Sprintf(format, args...), internal)
}

// Conflictf creates a conflict error with formatted message
func Conflictf(format string, args ...interface{}) *AppError {
	return New(ErrCodeConflict, fmt.Sprintf(format, args...))
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// FromError converts a generic error to AppError
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return Wrap(ErrCodeInternal, "An unexpected error occurred", err)
}
