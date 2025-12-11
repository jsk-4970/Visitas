package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/internal/services"
	"github.com/visitas/backend/pkg/logger"
)

// PatientHandler handles patient-related HTTP requests
type PatientHandler struct {
	patientService *services.PatientService
}

// NewPatientHandler creates a new patient handler
func NewPatientHandler(patientService *services.PatientService) *PatientHandler {
	return &PatientHandler{
		patientService: patientService,
	}
}

// CreatePatient handles POST /api/v1/patients
func (h *PatientHandler) CreatePatient(w http.ResponseWriter, r *http.Request) {
	var req models.PatientCreateRequest
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

	patient, err := h.patientService.CreatePatient(r.Context(), &req, userID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to create patient", err)
		respondError(w, http.StatusInternalServerError, "Failed to create patient")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"patient_id": patient.PatientID,
		"created_at": patient.CreatedAt,
		"message":    "Patient created successfully",
	})
}

// GetPatient handles GET /api/v1/patients/:id
func (h *PatientHandler) GetPatient(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
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

	patient, err := h.patientService.GetPatient(r.Context(), patientID, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to view this patient" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "patient not found" {
			respondError(w, http.StatusNotFound, "Patient not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to get patient", err)
			respondError(w, http.StatusInternalServerError, "Failed to get patient")
		}
		return
	}

	respondJSON(w, http.StatusOK, patient)
}

// GetMyPatients handles GET /api/v1/patients
func (h *PatientHandler) GetMyPatients(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	patients, err := h.patientService.GetMyPatients(r.Context(), userID, page, perPage)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to get patients", err)
		respondError(w, http.StatusInternalServerError, "Failed to get patients")
		return
	}

	respondJSON(w, http.StatusOK, patients)
}

// UpdatePatient handles PUT /api/v1/patients/:id
func (h *PatientHandler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	var req models.PatientUpdateRequest
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

	patient, err := h.patientService.UpdatePatient(r.Context(), patientID, &req, userID)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to update this patient" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "patient not found" {
			respondError(w, http.StatusNotFound, "Patient not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to update patient", err)
			respondError(w, http.StatusInternalServerError, "Failed to update patient")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"patient_id": patient.PatientID,
		"updated_at": patient.UpdatedAt,
		"message":    "Patient updated successfully",
	})
}

// DeletePatient handles DELETE /api/v1/patients/:id
func (h *PatientHandler) DeletePatient(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
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

	// Get reason from query or body
	reason := r.URL.Query().Get("reason")
	if reason == "" {
		reason = "Deleted by user request"
	}

	err := h.patientService.DeletePatient(r.Context(), patientID, userID, reason)
	if err != nil {
		if err.Error() == "access denied: you do not have permission to delete this patient" {
			respondError(w, http.StatusForbidden, err.Error())
		} else if err.Error() == "patient not found" {
			respondError(w, http.StatusNotFound, "Patient not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to delete patient", err)
			respondError(w, http.StatusInternalServerError, "Failed to delete patient")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Patient deleted successfully",
	})
}

// AssignPatientToStaff handles POST /api/v1/patients/:id/assign
func (h *PatientHandler) AssignPatientToStaff(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	var req struct {
		StaffID        string `json:"staff_id"`
		Role           string `json:"role"`
		AssignmentType string `json:"assignment_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.StaffID == "" || req.Role == "" {
		respondError(w, http.StatusBadRequest, "staff_id and role are required")
		return
	}

	// Default assignment type
	if req.AssignmentType == "" {
		req.AssignmentType = "primary"
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.patientService.AssignPatientToStaff(
		r.Context(),
		patientID,
		req.StaffID,
		repository.StaffRole(req.Role),
		repository.AssignmentType(req.AssignmentType),
		userID,
	)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to assign patient to staff", err)
		respondError(w, http.StatusInternalServerError, "Failed to assign patient to staff")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Patient assigned to staff successfully",
	})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
