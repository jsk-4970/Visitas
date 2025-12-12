package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/services"
	"github.com/visitas/backend/pkg/logger"
)

// ClinicalObservationHandler handles HTTP requests for clinical observations
type ClinicalObservationHandler struct {
	clinicalObservationService *services.ClinicalObservationService
}

// NewClinicalObservationHandler creates a new clinical observation handler
func NewClinicalObservationHandler(clinicalObservationService *services.ClinicalObservationService) *ClinicalObservationHandler {
	return &ClinicalObservationHandler{
		clinicalObservationService: clinicalObservationService,
	}
}

// CreateClinicalObservation handles POST /patients/{patient_id}/observations
func (h *ClinicalObservationHandler) CreateClinicalObservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ClinicalObservationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	observation, err := h.clinicalObservationService.CreateClinicalObservation(ctx, patientID, &req, userID)
	if err != nil {
		logger.Error("Failed to create clinical observation", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to create observations for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(observation)
}

// GetClinicalObservation handles GET /patients/{patient_id}/observations/{id}
func (h *ClinicalObservationHandler) GetClinicalObservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	observationID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	observation, err := h.clinicalObservationService.GetClinicalObservation(ctx, patientID, observationID, userID)
	if err != nil {
		logger.Error("Failed to get clinical observation", err)
		if err.Error() == "access denied: you do not have permission to view this patient's observations" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Clinical observation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(observation)
}

// GetClinicalObservations handles GET /patients/{patient_id}/observations
func (h *ClinicalObservationHandler) GetClinicalObservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := &models.ClinicalObservationFilter{
		PatientID: &patientID,
	}

	// Parse category
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Category = &category
	}

	// Parse effective_datetime_from
	if dateFrom := r.URL.Query().Get("effective_datetime_from"); dateFrom != "" {
		t, err := time.Parse(time.RFC3339, dateFrom)
		if err != nil {
			http.Error(w, "Invalid effective_datetime_from format (expected RFC3339)", http.StatusBadRequest)
			return
		}
		filter.EffectiveDatetimeFrom = &t
	}

	// Parse effective_datetime_to
	if dateTo := r.URL.Query().Get("effective_datetime_to"); dateTo != "" {
		t, err := time.Parse(time.RFC3339, dateTo)
		if err != nil {
			http.Error(w, "Invalid effective_datetime_to format (expected RFC3339)", http.StatusBadRequest)
			return
		}
		filter.EffectiveDatetimeTo = &t
	}

	// Parse performer_id
	if performerID := r.URL.Query().Get("performer_id"); performerID != "" {
		filter.PerformerID = &performerID
	}

	// Parse visit_record_id
	if visitRecordID := r.URL.Query().Get("visit_record_id"); visitRecordID != "" {
		filter.VisitRecordID = &visitRecordID
	}

	// Parse interpretation
	if interpretation := r.URL.Query().Get("interpretation"); interpretation != "" {
		filter.Interpretation = &interpretation
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		filter.Limit = limit
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
		filter.Offset = offset
	}

	observations, err := h.clinicalObservationService.ListClinicalObservations(ctx, filter, userID)
	if err != nil {
		logger.Error("Failed to list clinical observations", err)
		if err.Error() == "access denied: you do not have permission to view this patient's observations" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to retrieve clinical observations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(observations)
}

// UpdateClinicalObservation handles PUT /patients/{patient_id}/observations/{id}
func (h *ClinicalObservationHandler) UpdateClinicalObservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	observationID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ClinicalObservationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	observation, err := h.clinicalObservationService.UpdateClinicalObservation(ctx, patientID, observationID, &req, userID)
	if err != nil {
		logger.Error("Failed to update clinical observation", err)
		if err.Error() == "clinical observation not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to update observations for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(observation)
}

// DeleteClinicalObservation handles DELETE /patients/{patient_id}/observations/{id}
func (h *ClinicalObservationHandler) DeleteClinicalObservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	observationID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.clinicalObservationService.DeleteClinicalObservation(ctx, patientID, observationID, userID)
	if err != nil {
		logger.Error("Failed to delete clinical observation", err)
		if err.Error() == "access denied: you do not have permission to delete observations for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetLatestObservation handles GET /patients/{patient_id}/observations/latest/{category}
func (h *ClinicalObservationHandler) GetLatestObservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	category := chi.URLParam(r, "category")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	observation, err := h.clinicalObservationService.GetLatestObservationByCategory(ctx, patientID, category, userID)
	if err != nil {
		logger.Error("Failed to get latest observation", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to view this patient's observations" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(observation)
}

// GetTimeSeriesData handles GET /patients/{patient_id}/observations/timeseries/{category}
func (h *ClinicalObservationHandler) GetTimeSeriesData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	category := chi.URLParam(r, "category")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse from and to dates
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		http.Error(w, "from and to query parameters are required (RFC3339 format)", http.StatusBadRequest)
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		http.Error(w, "Invalid from format (expected RFC3339)", http.StatusBadRequest)
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		http.Error(w, "Invalid to format (expected RFC3339)", http.StatusBadRequest)
		return
	}

	observations, err := h.clinicalObservationService.GetTimeSeriesData(ctx, patientID, category, from, to, userID)
	if err != nil {
		logger.Error("Failed to get time series data", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to view this patient's observations" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(observations)
}
