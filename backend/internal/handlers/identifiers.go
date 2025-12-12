package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/services"
	"github.com/visitas/backend/pkg/logger"
)

// IdentifierHandler handles patient identifier-related HTTP requests
type IdentifierHandler struct {
	identifierService *services.IdentifierService
}

// NewIdentifierHandler creates a new identifier handler
func NewIdentifierHandler(
	identifierService *services.IdentifierService,
) *IdentifierHandler {
	return &IdentifierHandler{
		identifierService: identifierService,
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

	// Create identifier via service layer (handles access control and encryption)
	identifier, err := h.identifierService.CreateIdentifier(r.Context(), &req, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to add identifiers for this patient" {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
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

	// Check if decrypt parameter is set (only for authorized users)
	decrypt := r.URL.Query().Get("decrypt") == "true"

	// Get identifiers via service layer (handles access control, decryption, and audit logging)
	identifiers, err := h.identifierService.GetIdentifiersByPatientID(r.Context(), patientID, userID, decrypt)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to view this patient's identifiers" {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		logger.ErrorContext(r.Context(), "Failed to get identifiers", err)
		respondError(w, http.StatusInternalServerError, "Failed to get identifiers")
		return
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

	// Check if decrypt parameter is set
	decrypt := r.URL.Query().Get("decrypt") == "true"

	// Get identifier via service layer (handles access control, decryption, and audit logging)
	identifier, err := h.identifierService.GetIdentifier(r.Context(), identifierID, userID, decrypt)
	if err != nil {
		if err.Error() == "failed to get identifier: identifier not found" {
			respondError(w, http.StatusNotFound, "Identifier not found")
			return
		}
		if err.Error() == "access denied: you do not have permission to view this patient's identifiers" {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		logger.ErrorContext(r.Context(), "Failed to get identifier", err)
		respondError(w, http.StatusInternalServerError, "Failed to get identifier")
		return
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

	// Update identifier via service layer (handles access control and encryption)
	identifier, err := h.identifierService.UpdateIdentifier(r.Context(), identifierID, &req, userID)
	if err != nil {
		if err.Error() == "failed to get identifier: identifier not found" {
			respondError(w, http.StatusNotFound, "Identifier not found")
			return
		}
		if err.Error() == "access denied: you do not have permission to update this patient's identifiers" {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		logger.ErrorContext(r.Context(), "Failed to update identifier", err)
		respondError(w, http.StatusInternalServerError, "Failed to update identifier")
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

	// Delete identifier via service layer (handles access control and audit logging)
	err := h.identifierService.DeleteIdentifier(r.Context(), identifierID, userID)
	if err != nil {
		if err.Error() == "failed to get identifier: identifier not found" {
			respondError(w, http.StatusNotFound, "Identifier not found")
			return
		}
		if err.Error() == "access denied: you do not have permission to delete this patient's identifiers" {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		logger.ErrorContext(r.Context(), "Failed to delete identifier", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete identifier")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Identifier deleted successfully",
	})
}
