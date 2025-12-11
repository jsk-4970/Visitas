package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// IdentifierHandler handles patient identifier-related HTTP requests
type IdentifierHandler struct {
	identifierRepo *repository.IdentifierRepository
	patientRepo    *repository.PatientRepository
	auditMiddleware *middleware.AuditLoggerMiddleware
}

// NewIdentifierHandler creates a new identifier handler
func NewIdentifierHandler(
	identifierRepo *repository.IdentifierRepository,
	patientRepo *repository.PatientRepository,
	auditMiddleware *middleware.AuditLoggerMiddleware,
) *IdentifierHandler {
	return &IdentifierHandler{
		identifierRepo:  identifierRepo,
		patientRepo:     patientRepo,
		auditMiddleware: auditMiddleware,
	}
}

// CreateIdentifier handles POST /api/v1/patients/:patient_id/identifiers
func (h *IdentifierHandler) CreateIdentifier(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	var req models.PatientIdentifierCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set patient ID from URL
	req.PatientID = patientID

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user has access to this patient
	hasAccess, err := h.patientRepo.CheckStaffAccess(r.Context(), userID, patientID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to check staff access", err)
		respondError(w, http.StatusInternalServerError, "Failed to check access")
		return
	}

	if !hasAccess {
		logger.WarnContext(r.Context(), "Unauthorized access attempt", map[string]interface{}{
			"patient_id": patientID,
			"user_id":    userID,
		})
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	identifier, err := h.identifierRepo.CreateIdentifier(r.Context(), &req, userID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to create identifier", err)
		respondError(w, http.StatusInternalServerError, "Failed to create identifier")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"identifier_id": identifier.IdentifierID,
		"created_at":    identifier.CreatedAt,
		"message":       "Identifier created successfully",
	})
}

// GetIdentifiers handles GET /api/v1/patients/:patient_id/identifiers
func (h *IdentifierHandler) GetIdentifiers(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user has access to this patient
	hasAccess, err := h.patientRepo.CheckStaffAccess(r.Context(), userID, patientID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to check staff access", err)
		respondError(w, http.StatusInternalServerError, "Failed to check access")
		return
	}

	if !hasAccess {
		logger.WarnContext(r.Context(), "Unauthorized access attempt", map[string]interface{}{
			"patient_id": patientID,
			"user_id":    userID,
		})
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Check if decrypt parameter is set (only for authorized users)
	decrypt := r.URL.Query().Get("decrypt") == "true"

	identifiers, err := h.identifierRepo.GetIdentifiersByPatientID(r.Context(), patientID, decrypt)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to get identifiers", err)
		respondError(w, http.StatusInternalServerError, "Failed to get identifiers")
		return
	}

	// If decryption was requested, log the audit trail for My Numbers
	if decrypt {
		for _, identifier := range identifiers {
			if identifier.IsMyNumber() {
				h.auditMiddleware.LogDecryptAccess(r.Context(), patientID, identifier.IdentifierID, userID)
			}
		}
	}

	response := models.PatientIdentifierListResponse{
		Identifiers: []models.PatientIdentifier{},
		Total:       len(identifiers),
	}

	for _, id := range identifiers {
		response.Identifiers = append(response.Identifiers, *id)
	}

	respondJSON(w, http.StatusOK, response)
}

// GetIdentifier handles GET /api/v1/patients/:patient_id/identifiers/:id
func (h *IdentifierHandler) GetIdentifier(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	identifierID := chi.URLParam(r, "id")

	if patientID == "" || identifierID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID and Identifier ID are required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user has access to this patient
	hasAccess, err := h.patientRepo.CheckStaffAccess(r.Context(), userID, patientID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to check staff access", err)
		respondError(w, http.StatusInternalServerError, "Failed to check access")
		return
	}

	if !hasAccess {
		logger.WarnContext(r.Context(), "Unauthorized access attempt", map[string]interface{}{
			"patient_id": patientID,
			"user_id":    userID,
		})
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Check if decrypt parameter is set
	decrypt := r.URL.Query().Get("decrypt") == "true"

	identifier, err := h.identifierRepo.GetIdentifierByID(r.Context(), identifierID, decrypt)
	if err != nil {
		if err.Error() == "identifier not found" {
			respondError(w, http.StatusNotFound, "Identifier not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to get identifier", err)
			respondError(w, http.StatusInternalServerError, "Failed to get identifier")
		}
		return
	}

	// Log decrypt access if My Number was decrypted
	if decrypt && identifier.IsMyNumber() {
		h.auditMiddleware.LogDecryptAccess(r.Context(), patientID, identifierID, userID)
	}

	respondJSON(w, http.StatusOK, identifier)
}

// UpdateIdentifier handles PUT /api/v1/patients/:patient_id/identifiers/:id
func (h *IdentifierHandler) UpdateIdentifier(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	identifierID := chi.URLParam(r, "id")

	if patientID == "" || identifierID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID and Identifier ID are required")
		return
	}

	var req models.PatientIdentifierUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user has access to this patient
	hasAccess, err := h.patientRepo.CheckStaffAccess(r.Context(), userID, patientID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to check staff access", err)
		respondError(w, http.StatusInternalServerError, "Failed to check access")
		return
	}

	if !hasAccess {
		logger.WarnContext(r.Context(), "Unauthorized access attempt", map[string]interface{}{
			"patient_id": patientID,
			"user_id":    userID,
		})
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	identifier, err := h.identifierRepo.UpdateIdentifier(r.Context(), identifierID, &req, userID)
	if err != nil {
		if err.Error() == "identifier not found" {
			respondError(w, http.StatusNotFound, "Identifier not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to update identifier", err)
			respondError(w, http.StatusInternalServerError, "Failed to update identifier")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"identifier_id": identifier.IdentifierID,
		"updated_at":    identifier.UpdatedAt,
		"message":       "Identifier updated successfully",
	})
}

// DeleteIdentifier handles DELETE /api/v1/patients/:patient_id/identifiers/:id
func (h *IdentifierHandler) DeleteIdentifier(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	identifierID := chi.URLParam(r, "id")

	if patientID == "" || identifierID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID and Identifier ID are required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user has access to this patient
	hasAccess, err := h.patientRepo.CheckStaffAccess(r.Context(), userID, patientID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to check staff access", err)
		respondError(w, http.StatusInternalServerError, "Failed to check access")
		return
	}

	if !hasAccess {
		logger.WarnContext(r.Context(), "Unauthorized access attempt", map[string]interface{}{
			"patient_id": patientID,
			"user_id":    userID,
		})
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	err = h.identifierRepo.DeleteIdentifier(r.Context(), identifierID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to delete identifier", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete identifier")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Identifier deleted successfully",
	})
}
