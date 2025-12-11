package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// AuditLoggerMiddleware logs patient access for audit purposes
type AuditLoggerMiddleware struct {
	auditRepo *repository.AuditRepository
}

// NewAuditLoggerMiddleware creates a new audit logger middleware
func NewAuditLoggerMiddleware(auditRepo *repository.AuditRepository) *AuditLoggerMiddleware {
	return &AuditLoggerMiddleware{
		auditRepo: auditRepo,
	}
}

// responseWriter is a wrapper for http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.written = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the write
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// LogPatientAccess logs access to patient resources
func (alm *AuditLoggerMiddleware) LogPatientAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only log patient-related endpoints
		if !isPatientEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Get user ID from context
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			userID = "anonymous"
		}

		// Extract patient ID from URL path
		patientID := chi.URLParam(r, "id")
		if patientID == "" {
			patientID = chi.URLParam(r, "patient_id")
		}

		// Determine action based on HTTP method
		action := determineAction(r.Method)

		// Start time for performance tracking
		startTime := time.Now()

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(startTime)

		// Determine if the request was successful
		success := wrapped.statusCode >= 200 && wrapped.statusCode < 300

		// Create audit log entry
		auditLog := &repository.AuditLog{
			EventTime:  time.Now(),
			ActorID:    userID,
			Action:     action,
			ResourceID: patientID,
			PatientID:  patientID,
			Success:    success,
			IPAddress:  getClientIP(r),
			UserAgent:  r.UserAgent(),
		}

		// Add accessed fields if available (from response body)
		if success {
			accessedFields := extractAccessedFields(r)
			if len(accessedFields) > 0 {
				fieldsJSON, _ := json.Marshal(accessedFields)
				auditLog.AccessedFields = fieldsJSON
			}
		} else {
			// Log error message for failed requests
			auditLog.ErrorMessage = http.StatusText(wrapped.statusCode)
		}

		// Log the audit entry
		if err := alm.auditRepo.LogAccess(r.Context(), auditLog); err != nil {
			logger.ErrorContext(r.Context(), "Failed to write audit log", err, map[string]interface{}{
				"patient_id": patientID,
				"action":     string(action),
				"user_id":    userID,
			})
		}

		// Log request details
		logger.InfoContext(r.Context(), "Patient access logged", map[string]interface{}{
			"patient_id":  patientID,
			"action":      string(action),
			"user_id":     userID,
			"status_code": wrapped.statusCode,
			"duration_ms": duration.Milliseconds(),
			"success":     success,
		})
	})
}

// isPatientEndpoint checks if the path is a patient-related endpoint
func isPatientEndpoint(path string) bool {
	return strings.Contains(path, "/patients") ||
		strings.Contains(path, "/patient_identifiers") ||
		strings.Contains(path, "/social_profiles") ||
		strings.Contains(path, "/coverages")
}

// determineAction determines the audit action based on HTTP method
func determineAction(method string) repository.AuditAction {
	switch method {
	case http.MethodGet:
		return repository.AuditActionView
	case http.MethodPost:
		return repository.AuditActionCreate
	case http.MethodPut, http.MethodPatch:
		return repository.AuditActionUpdate
	case http.MethodDelete:
		return repository.AuditActionDelete
	default:
		return repository.AuditActionView
	}
}

// extractAccessedFields extracts fields that were accessed from the request
func extractAccessedFields(r *http.Request) []string {
	// For GET requests, we can extract query parameters as accessed fields
	if r.Method == http.MethodGet {
		fields := []string{}
		for key := range r.URL.Query() {
			fields = append(fields, key)
		}
		return fields
	}

	// For POST/PUT requests, we could parse the body to determine accessed fields
	// This is a simplified version - you might want to enhance this based on your needs
	return []string{}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for requests behind a proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// LogDecryptAccess logs when sensitive data (like My Number) is decrypted
func (alm *AuditLoggerMiddleware) LogDecryptAccess(ctx context.Context, patientID, identifierID, userID string) error {
	auditLog := &repository.AuditLog{
		EventTime:  time.Now(),
		ActorID:    userID,
		Action:     repository.AuditActionDecrypt,
		ResourceID: identifierID,
		PatientID:  patientID,
		Success:    true,
	}

	accessedFields := []string{"my_number"}
	fieldsJSON, _ := json.Marshal(accessedFields)
	auditLog.AccessedFields = fieldsJSON

	if err := alm.auditRepo.LogAccess(ctx, auditLog); err != nil {
		logger.ErrorContext(ctx, "Failed to write decrypt audit log", err, map[string]interface{}{
			"patient_id":    patientID,
			"identifier_id": identifierID,
			"user_id":       userID,
		})
		return err
	}

	logger.InfoContext(ctx, "My Number decryption logged", map[string]interface{}{
		"patient_id":    patientID,
		"identifier_id": identifierID,
		"user_id":       userID,
	})

	return nil
}
