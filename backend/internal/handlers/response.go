package handlers

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/visitas/backend/internal/errors"
	"github.com/visitas/backend/pkg/logger"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error("Failed to encode JSON response", err)
		}
	}
}

// WriteError writes an error response based on the error type
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := apperrors.GetHTTPStatusCode(err)
	errorType := getErrorType(statusCode)

	response := ErrorResponse{
		Error:   errorType,
		Message: err.Error(),
	}

	// Add details if available
	var appErr *apperrors.AppError
	if ok := apperrors.Wrap(err, "") != nil; ok {
		if ae, isAppErr := err.(*apperrors.AppError); isAppErr && ae.Details != nil {
			response.Details = ae.Details
		}
	}
	_ = appErr // suppress unused variable warning

	// Log server errors
	if statusCode >= 500 {
		logger.ErrorContext(r.Context(), "Server error", err, map[string]interface{}{
			"path":   r.URL.Path,
			"method": r.Method,
		})
	}

	WriteJSON(w, statusCode, response)
}

// WriteNotFound writes a 404 not found response
func WriteNotFound(w http.ResponseWriter, resource string) {
	WriteJSON(w, http.StatusNotFound, ErrorResponse{
		Error:   "not_found",
		Message: resource + " not found",
	})
}

// WriteBadRequest writes a 400 bad request response
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusBadRequest, ErrorResponse{
		Error:   "bad_request",
		Message: message,
	})
}

// WriteUnauthorized writes a 401 unauthorized response
func WriteUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "authentication required"
	}
	WriteJSON(w, http.StatusUnauthorized, ErrorResponse{
		Error:   "unauthorized",
		Message: message,
	})
}

// WriteForbidden writes a 403 forbidden response
func WriteForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "access denied"
	}
	WriteJSON(w, http.StatusForbidden, ErrorResponse{
		Error:   "forbidden",
		Message: message,
	})
}

// WriteConflict writes a 409 conflict response
func WriteConflict(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusConflict, ErrorResponse{
		Error:   "conflict",
		Message: message,
	})
}

// WriteInternalError writes a 500 internal server error response
func WriteInternalError(w http.ResponseWriter, r *http.Request, err error) {
	logger.ErrorContext(r.Context(), "Internal server error", err, map[string]interface{}{
		"path":   r.URL.Path,
		"method": r.Method,
	})
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error:   "internal_server_error",
		Message: "an internal error occurred",
	})
}

// WriteNoContent writes a 204 no content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteCreated writes a 201 created response with data
func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, data)
}

// WriteOK writes a 200 OK response with data
func WriteOK(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, data)
}

// getErrorType returns a string error type based on status code
func getErrorType(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusTooManyRequests:
		return "rate_limit_exceeded"
	case http.StatusInternalServerError:
		return "internal_server_error"
	default:
		return "error"
	}
}
