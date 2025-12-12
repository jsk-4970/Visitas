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

// MedicalConditionHandler handles medical condition-related HTTP requests
type MedicalConditionHandler struct {
	conditionService *services.MedicalConditionService
}

// NewMedicalConditionHandler creates a new medical condition handler
func NewMedicalConditionHandler(conditionService *services.MedicalConditionService) *MedicalConditionHandler {
	return &MedicalConditionHandler{
		conditionService: conditionService,
	}
}

// CreateMedicalCondition handles POST /api/v1/patients/:patient_id/conditions
func (h *MedicalConditionHandler) CreateMedicalCondition(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	var req models.MedicalConditionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Override patientID from URL
	req.PatientID = patientID

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	condition, err := h.conditionService.CreateCondition(r.Context(), &req, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to add conditions for this patient" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			logger.ErrorContext(r.Context(), "Failed to create medical condition", err)
			respondError(w, http.StatusInternalServerError, "Failed to create medical condition")
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"condition_id": condition.ConditionID,
		"created_at":   condition.CreatedAt,
		"message":      "Medical condition created successfully",
	})
}

// GetMedicalConditions handles GET /api/v1/patients/:patient_id/conditions
func (h *MedicalConditionHandler) GetMedicalConditions(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	activeOnly := r.URL.Query().Get("active_only") == "true"

	var conditions []*models.MedicalCondition
	var err error

	if activeOnly {
		// Get only active conditions
		conditions, err = h.conditionService.GetActiveConditions(r.Context(), patientID, userID)
	} else {
		// Get all conditions
		conditions, err = h.conditionService.GetConditionsByPatient(r.Context(), patientID, userID)
	}

	if err != nil {
		if err.Error() == "access denied: you do not have permission to view conditions for this patient" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			logger.ErrorContext(r.Context(), "Failed to get medical conditions", err)
			respondError(w, http.StatusInternalServerError, "Failed to get medical conditions")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"conditions": conditions,
		"total":      len(conditions),
	})
}

// GetMedicalCondition handles GET /api/v1/patients/:patient_id/conditions/:id
func (h *MedicalConditionHandler) GetMedicalCondition(w http.ResponseWriter, r *http.Request) {
	conditionID := chi.URLParam(r, "id")
	if conditionID == "" {
		respondError(w, http.StatusBadRequest, "Condition ID is required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	condition, err := h.conditionService.GetCondition(r.Context(), conditionID, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to view this condition" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "condition not found" {
			respondError(w, http.StatusNotFound, "Medical condition not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to get medical condition", err)
			respondError(w, http.StatusInternalServerError, "Failed to get medical condition")
		}
		return
	}

	respondJSON(w, http.StatusOK, condition)
}

// UpdateMedicalCondition handles PUT /api/v1/patients/:patient_id/conditions/:id
func (h *MedicalConditionHandler) UpdateMedicalCondition(w http.ResponseWriter, r *http.Request) {
	conditionID := chi.URLParam(r, "id")
	if conditionID == "" {
		respondError(w, http.StatusBadRequest, "Condition ID is required")
		return
	}

	var req models.MedicalConditionUpdateRequest
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

	condition, err := h.conditionService.UpdateCondition(r.Context(), conditionID, &req, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to update this condition" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "condition not found" {
			respondError(w, http.StatusNotFound, "Medical condition not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to update medical condition", err)
			respondError(w, http.StatusInternalServerError, "Failed to update medical condition")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"condition_id": condition.ConditionID,
		"updated_at":   condition.UpdatedAt,
		"message":      "Medical condition updated successfully",
	})
}

// DeleteMedicalCondition handles DELETE /api/v1/patients/:patient_id/conditions/:id
func (h *MedicalConditionHandler) DeleteMedicalCondition(w http.ResponseWriter, r *http.Request) {
	conditionID := chi.URLParam(r, "id")
	if conditionID == "" {
		respondError(w, http.StatusBadRequest, "Condition ID is required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.conditionService.DeleteCondition(r.Context(), conditionID, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to delete this condition" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "condition not found" {
			respondError(w, http.StatusNotFound, "Medical condition not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to delete medical condition", err)
			respondError(w, http.StatusInternalServerError, "Failed to delete medical condition")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Medical condition deleted successfully",
	})
}
