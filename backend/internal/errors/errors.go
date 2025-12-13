package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard error types for the application
var (
	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrAccessDenied indicates the user doesn't have permission
	ErrAccessDenied = errors.New("access denied")

	// ErrUnauthorized indicates the user is not authenticated
	ErrUnauthorized = errors.New("unauthorized")

	// ErrBadRequest indicates invalid input
	ErrBadRequest = errors.New("bad request")

	// ErrConflict indicates a conflict (e.g., optimistic locking failure)
	ErrConflict = errors.New("conflict")

	// ErrInternal indicates an internal server error
	ErrInternal = errors.New("internal server error")

	// ErrValidation indicates validation failure
	ErrValidation = errors.New("validation error")
)

// AppError represents an application-level error with context
type AppError struct {
	// Err is the underlying error
	Err error
	// Message is a human-readable message
	Message string
	// Code is the HTTP status code
	Code int
	// Details contains additional error details
	Details map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatusCode returns the HTTP status code for this error
func (e *AppError) HTTPStatusCode() int {
	if e.Code != 0 {
		return e.Code
	}
	return http.StatusInternalServerError
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Err:     ErrNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Code:    http.StatusNotFound,
	}
}

// NewAccessDeniedError creates an access denied error
func NewAccessDeniedError(message string) *AppError {
	if message == "" {
		message = "you do not have permission to access this resource"
	}
	return &AppError{
		Err:     ErrAccessDenied,
		Message: message,
		Code:    http.StatusForbidden,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "authentication required"
	}
	return &AppError{
		Err:     ErrUnauthorized,
		Message: message,
		Code:    http.StatusUnauthorized,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Err:     ErrBadRequest,
		Message: message,
		Code:    http.StatusBadRequest,
	}
}

// NewValidationError creates a validation error with field details
func NewValidationError(message string, fields map[string]string) *AppError {
	details := make(map[string]interface{})
	for k, v := range fields {
		details[k] = v
	}
	return &AppError{
		Err:     ErrValidation,
		Message: message,
		Code:    http.StatusBadRequest,
		Details: details,
	}
}

// NewConflictError creates a conflict error (e.g., optimistic locking)
func NewConflictError(message string) *AppError {
	return &AppError{
		Err:     ErrConflict,
		Message: message,
		Code:    http.StatusConflict,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string, cause error) *AppError {
	if message == "" {
		message = "an internal error occurred"
	}
	return &AppError{
		Err:     cause,
		Message: message,
		Code:    http.StatusInternalServerError,
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, preserve the code
	var appErr *AppError
	if errors.As(err, &appErr) {
		return &AppError{
			Err:     err,
			Message: message + ": " + appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details,
		}
	}

	// Check for standard errors
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, ErrAccessDenied):
		code = http.StatusForbidden
	case errors.Is(err, ErrUnauthorized):
		code = http.StatusUnauthorized
	case errors.Is(err, ErrBadRequest), errors.Is(err, ErrValidation):
		code = http.StatusBadRequest
	case errors.Is(err, ErrConflict):
		code = http.StatusConflict
	}

	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAccessDenied checks if the error is an access denied error
func IsAccessDenied(err error) bool {
	return errors.Is(err, ErrAccessDenied)
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsBadRequest checks if the error is a bad request error
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest) || errors.Is(err, ErrValidation)
}

// IsConflict checks if the error is a conflict error
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// GetHTTPStatusCode returns the HTTP status code for an error
func GetHTTPStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatusCode()
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrAccessDenied):
		return http.StatusForbidden
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrBadRequest), errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
