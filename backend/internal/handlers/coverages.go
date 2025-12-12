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

// CoverageHandler handles insurance coverage-related HTTP requests
type CoverageHandler struct {
	coverageRepo *repository.CoverageRepository
	patientRepo  *repository.PatientRepository
}

// NewCoverageHandler creates a new coverage handler
func NewCoverageHandler(coverageRepo *repository.CoverageRepository, patientRepo *repository.PatientRepository) *CoverageHandler {
	return &CoverageHandler{
		coverageRepo: coverageRepo,
		patientRepo:  patientRepo,
	}
}

// CreateCoverage handles POST /api/v1/patients/:patient_id/coverages
func (h *CoverageHandler) CreateCoverage(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	var req models.PatientCoverageCreateRequest
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

	coverage, err := h.coverageRepo.CreateCoverage(r.Context(), &req, userID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to create coverage", err)
		respondError(w, http.StatusInternalServerError, "Failed to create coverage")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"coverage_id": coverage.CoverageID,
		"created_at":  coverage.CreatedAt,
		"message":     "Coverage created successfully",
	})
}

// GetCoverages handles GET /api/v1/patients/:patient_id/coverages
func (h *CoverageHandler) GetCoverages(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	// Parse query parameters
	insuranceType := r.URL.Query().Get("insurance_type")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	var coverages []*models.PatientCoverage
	var err error

	if insuranceType != "" {
		// Filter by insurance type
		coverages, err = h.coverageRepo.GetCoveragesByPatientAndType(r.Context(), patientID, insuranceType)
	} else if activeOnly {
		// Get only active coverages
		coverages, err = h.coverageRepo.GetActiveCoverages(r.Context(), patientID)
	} else {
		// Get all coverages
		coverages, err = h.coverageRepo.GetCoveragesByPatient(r.Context(), patientID)
	}

	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to get coverages", err)
		respondError(w, http.StatusInternalServerError, "Failed to get coverages")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"coverages": coverages,
		"total":     len(coverages),
	})
}

// GetCoverage handles GET /api/v1/patients/:patient_id/coverages/:id
func (h *CoverageHandler) GetCoverage(w http.ResponseWriter, r *http.Request) {
	coverageID := chi.URLParam(r, "id")
	if coverageID == "" {
		respondError(w, http.StatusBadRequest, "Coverage ID is required")
		return
	}

	coverage, err := h.coverageRepo.GetCoverageByID(r.Context(), coverageID)
	if err != nil {
		if err.Error() == "coverage not found" {
			respondError(w, http.StatusNotFound, "Coverage not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to get coverage", err)
			respondError(w, http.StatusInternalServerError, "Failed to get coverage")
		}
		return
	}

	respondJSON(w, http.StatusOK, coverage)
}

// UpdateCoverage handles PUT /api/v1/patients/:patient_id/coverages/:id
func (h *CoverageHandler) UpdateCoverage(w http.ResponseWriter, r *http.Request) {
	coverageID := chi.URLParam(r, "id")
	if coverageID == "" {
		respondError(w, http.StatusBadRequest, "Coverage ID is required")
		return
	}

	var req models.PatientCoverageUpdateRequest
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

	coverage, err := h.coverageRepo.UpdateCoverage(r.Context(), coverageID, &req, userID)
	if err != nil {
		if err.Error() == "coverage not found" {
			respondError(w, http.StatusNotFound, "Coverage not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to update coverage", err)
			respondError(w, http.StatusInternalServerError, "Failed to update coverage")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"coverage_id": coverage.CoverageID,
		"updated_at":  coverage.UpdatedAt,
		"message":     "Coverage updated successfully",
	})
}

// DeleteCoverage handles DELETE /api/v1/patients/:patient_id/coverages/:id
func (h *CoverageHandler) DeleteCoverage(w http.ResponseWriter, r *http.Request) {
	coverageID := chi.URLParam(r, "id")
	if coverageID == "" {
		respondError(w, http.StatusBadRequest, "Coverage ID is required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.coverageRepo.DeleteCoverage(r.Context(), coverageID, userID)
	if err != nil {
		if err.Error() == "coverage not found" {
			respondError(w, http.StatusNotFound, "Coverage not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to delete coverage", err)
			respondError(w, http.StatusInternalServerError, "Failed to delete coverage")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Coverage deleted successfully",
	})
}

// VerifyCoverage handles POST /api/v1/patients/:patient_id/coverages/:id/verify
func (h *CoverageHandler) VerifyCoverage(w http.ResponseWriter, r *http.Request) {
	coverageID := chi.URLParam(r, "id")
	if coverageID == "" {
		respondError(w, http.StatusBadRequest, "Coverage ID is required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.coverageRepo.VerifyCoverage(r.Context(), coverageID, userID)
	if err != nil {
		if err.Error() == "coverage not found" {
			respondError(w, http.StatusNotFound, "Coverage not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to verify coverage", err)
			respondError(w, http.StatusInternalServerError, "Failed to verify coverage")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Coverage verified successfully",
	})
}
