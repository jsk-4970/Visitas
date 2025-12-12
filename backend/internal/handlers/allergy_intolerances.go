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

// AllergyIntoleranceHandler handles allergy intolerance-related HTTP requests
type AllergyIntoleranceHandler struct {
	allergyRepo *repository.AllergyIntoleranceRepository
	patientRepo *repository.PatientRepository
}

// NewAllergyIntoleranceHandler creates a new allergy intolerance handler
func NewAllergyIntoleranceHandler(allergyRepo *repository.AllergyIntoleranceRepository, patientRepo *repository.PatientRepository) *AllergyIntoleranceHandler {
	return &AllergyIntoleranceHandler{
		allergyRepo: allergyRepo,
		patientRepo: patientRepo,
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

	// Verify patient exists
	_, err := h.patientRepo.GetPatientByID(r.Context(), patientID)
	if err != nil {
		if err.Error() == "patient not found" {
			respondError(w, http.StatusNotFound, "Patient not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to verify patient", err)
			respondError(w, http.StatusInternalServerError, "Failed to verify patient")
		}
		return
	}

	allergy, err := h.allergyRepo.CreateAllergy(r.Context(), &req, userID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to create allergy intolerance", err)
		respondError(w, http.StatusInternalServerError, "Failed to create allergy intolerance")
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

	// Parse query parameters
	activeOnly := r.URL.Query().Get("active_only") == "true"
	medicationOnly := r.URL.Query().Get("medication_only") == "true"

	var allergies []*models.AllergyIntolerance
	var err error

	if medicationOnly {
		// Get only medication allergies
		allergies, err = h.allergyRepo.GetMedicationAllergies(r.Context(), patientID)
	} else if activeOnly {
		// Get only active allergies
		allergies, err = h.allergyRepo.GetActiveAllergies(r.Context(), patientID)
	} else {
		// Get all allergies
		allergies, err = h.allergyRepo.GetAllergiesByPatient(r.Context(), patientID)
	}

	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to get allergy intolerances", err)
		respondError(w, http.StatusInternalServerError, "Failed to get allergy intolerances")
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

	allergy, err := h.allergyRepo.GetAllergyByID(r.Context(), allergyID)
	if err != nil {
		if err.Error() == "allergy intolerance not found" {
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

	allergy, err := h.allergyRepo.UpdateAllergy(r.Context(), allergyID, &req, userID)
	if err != nil {
		if err.Error() == "allergy intolerance not found" {
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

	err := h.allergyRepo.DeleteAllergy(r.Context(), allergyID, userID)
	if err != nil {
		if err.Error() == "allergy intolerance not found" {
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
