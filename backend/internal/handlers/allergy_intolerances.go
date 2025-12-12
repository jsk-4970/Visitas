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

// AllergyIntoleranceHandler handles allergy intolerance-related HTTP requests
type AllergyIntoleranceHandler struct {
	allergyService *services.AllergyIntoleranceService
}

// NewAllergyIntoleranceHandler creates a new allergy intolerance handler
func NewAllergyIntoleranceHandler(allergyService *services.AllergyIntoleranceService) *AllergyIntoleranceHandler {
	return &AllergyIntoleranceHandler{
		allergyService: allergyService,
	}
}

// CreateAllergyIntolerance handles POST /api/v1/patients/:patient_id/allergies
func (h *AllergyIntoleranceHandler) CreateAllergyIntolerance(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	var req models.AllergyIntoleranceCreateRequest
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

	allergy, err := h.allergyService.CreateAllergy(r.Context(), &req, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to add allergies for this patient" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			logger.ErrorContext(r.Context(), "Failed to create allergy intolerance", err)
			respondError(w, http.StatusInternalServerError, "Failed to create allergy intolerance")
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"allergy_id": allergy.AllergyID,
		"created_at": allergy.CreatedAt,
		"message":    "Allergy intolerance created successfully",
	})
}

// GetAllergyIntolerances handles GET /api/v1/patients/:patient_id/allergies
func (h *AllergyIntoleranceHandler) GetAllergyIntolerances(w http.ResponseWriter, r *http.Request) {
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
	medicationOnly := r.URL.Query().Get("medication_only") == "true"

	var allergies []*models.AllergyIntolerance
	var err error

	if medicationOnly {
		// Get only medication allergies
		allergies, err = h.allergyService.GetMedicationAllergies(r.Context(), patientID, userID)
	} else if activeOnly {
		// Get only active allergies
		allergies, err = h.allergyService.GetActiveAllergies(r.Context(), patientID, userID)
	} else {
		// Get all allergies
		allergies, err = h.allergyService.GetAllergiesByPatient(r.Context(), patientID, userID)
	}

	if err != nil {
		if err.Error() == "access denied: you do not have permission to view allergies for this patient" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			logger.ErrorContext(r.Context(), "Failed to get allergy intolerances", err)
			respondError(w, http.StatusInternalServerError, "Failed to get allergy intolerances")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"allergies": allergies,
		"total":     len(allergies),
	})
}

// GetAllergyIntolerance handles GET /api/v1/patients/:patient_id/allergies/:id
func (h *AllergyIntoleranceHandler) GetAllergyIntolerance(w http.ResponseWriter, r *http.Request) {
	allergyID := chi.URLParam(r, "id")
	if allergyID == "" {
		respondError(w, http.StatusBadRequest, "Allergy ID is required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	allergy, err := h.allergyService.GetAllergy(r.Context(), allergyID, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to view this allergy" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "allergy not found" {
			respondError(w, http.StatusNotFound, "Allergy intolerance not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to get allergy intolerance", err)
			respondError(w, http.StatusInternalServerError, "Failed to get allergy intolerance")
		}
		return
	}

	respondJSON(w, http.StatusOK, allergy)
}

// UpdateAllergyIntolerance handles PUT /api/v1/patients/:patient_id/allergies/:id
func (h *AllergyIntoleranceHandler) UpdateAllergyIntolerance(w http.ResponseWriter, r *http.Request) {
	allergyID := chi.URLParam(r, "id")
	if allergyID == "" {
		respondError(w, http.StatusBadRequest, "Allergy ID is required")
		return
	}

	var req models.AllergyIntoleranceUpdateRequest
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

	allergy, err := h.allergyService.UpdateAllergy(r.Context(), allergyID, &req, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to update this allergy" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "allergy not found" {
			respondError(w, http.StatusNotFound, "Allergy intolerance not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to update allergy intolerance", err)
			respondError(w, http.StatusInternalServerError, "Failed to update allergy intolerance")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"allergy_id": allergy.AllergyID,
		"updated_at": allergy.UpdatedAt,
		"message":    "Allergy intolerance updated successfully",
	})
}

// DeleteAllergyIntolerance handles DELETE /api/v1/patients/:patient_id/allergies/:id
func (h *AllergyIntoleranceHandler) DeleteAllergyIntolerance(w http.ResponseWriter, r *http.Request) {
	allergyID := chi.URLParam(r, "id")
	if allergyID == "" {
		respondError(w, http.StatusBadRequest, "Allergy ID is required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.allergyService.DeleteAllergy(r.Context(), allergyID, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to delete this allergy" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "allergy not found" {
			respondError(w, http.StatusNotFound, "Allergy intolerance not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to delete allergy intolerance", err)
			respondError(w, http.StatusInternalServerError, "Failed to delete allergy intolerance")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Allergy intolerance deleted successfully",
	})
}
