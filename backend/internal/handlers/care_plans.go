package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/services"
	"github.com/visitas/backend/pkg/logger"
)

// CarePlanHandler handles HTTP requests for care plans
type CarePlanHandler struct {
	carePlanService *services.CarePlanService
}

// NewCarePlanHandler creates a new care plan handler
func NewCarePlanHandler(carePlanService *services.CarePlanService) *CarePlanHandler {
	return &CarePlanHandler{
		carePlanService: carePlanService,
	}
}

// CreateCarePlan handles POST /patients/{patient_id}/care-plans
func (h *CarePlanHandler) CreateCarePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	var req models.CarePlanCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	carePlan, err := h.carePlanService.CreateCarePlan(ctx, patientID, &req)
	if err != nil {
		logger.Error("Failed to create care plan", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(carePlan)
}

// GetCarePlan handles GET /patients/{patient_id}/care-plans/{id}
func (h *CarePlanHandler) GetCarePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	planID := chi.URLParam(r, "id")

	carePlan, err := h.carePlanService.GetCarePlan(ctx, patientID, planID)
	if err != nil {
		logger.Error("Failed to get care plan", err)
		http.Error(w, "Care plan not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(carePlan)
}

// GetCarePlans handles GET /patients/{patient_id}/care-plans
func (h *CarePlanHandler) GetCarePlans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Parse query parameters
	filter := &models.CarePlanFilter{
		PatientID: &patientID,
	}

	// Parse status
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	// Parse intent
	if intent := r.URL.Query().Get("intent"); intent != "" {
		filter.Intent = &intent
	}

	// Parse period_start_from
	if periodStartFrom := r.URL.Query().Get("period_start_from"); periodStartFrom != "" {
		t, err := time.Parse("2006-01-02", periodStartFrom)
		if err != nil {
			http.Error(w, "Invalid period_start_from format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		filter.PeriodStartFrom = &t
	}

	// Parse period_start_to
	if periodStartTo := r.URL.Query().Get("period_start_to"); periodStartTo != "" {
		t, err := time.Parse("2006-01-02", periodStartTo)
		if err != nil {
			http.Error(w, "Invalid period_start_to format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		filter.PeriodStartTo = &t
	}

	// Parse created_by
	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		filter.CreatedBy = &createdBy
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

	carePlans, err := h.carePlanService.ListCarePlans(ctx, filter)
	if err != nil {
		logger.Error("Failed to list care plans", err)
		http.Error(w, "Failed to retrieve care plans", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(carePlans)
}

// UpdateCarePlan handles PUT /patients/{patient_id}/care-plans/{id}
func (h *CarePlanHandler) UpdateCarePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	planID := chi.URLParam(r, "id")

	var req models.CarePlanUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	carePlan, err := h.carePlanService.UpdateCarePlan(ctx, patientID, planID, &req)
	if err != nil {
		logger.Error("Failed to update care plan", err)
		if err.Error() == "care plan not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(carePlan)
}

// DeleteCarePlan handles DELETE /patients/{patient_id}/care-plans/{id}
func (h *CarePlanHandler) DeleteCarePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	planID := chi.URLParam(r, "id")

	err := h.carePlanService.DeleteCarePlan(ctx, patientID, planID)
	if err != nil {
		logger.Error("Failed to delete care plan", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetActiveCarePlans handles GET /patients/{patient_id}/care-plans/active
func (h *CarePlanHandler) GetActiveCarePlans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	carePlans, err := h.carePlanService.GetActiveCarePlans(ctx, patientID)
	if err != nil {
		logger.Error("Failed to get active care plans", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve active care plans", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(carePlans)
}
