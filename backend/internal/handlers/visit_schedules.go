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

// VisitScheduleHandler handles HTTP requests for visit schedules
type VisitScheduleHandler struct {
	visitScheduleService *services.VisitScheduleService
}

// NewVisitScheduleHandler creates a new visit schedule handler
func NewVisitScheduleHandler(visitScheduleService *services.VisitScheduleService) *VisitScheduleHandler {
	return &VisitScheduleHandler{
		visitScheduleService: visitScheduleService,
	}
}

// CreateVisitSchedule handles POST /patients/{patient_id}/schedules
func (h *VisitScheduleHandler) CreateVisitSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.VisitScheduleCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	schedule, err := h.visitScheduleService.CreateVisitSchedule(ctx, patientID, &req, userID)
	if err != nil {
		logger.Error("Failed to create visit schedule", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to create schedules for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

// GetVisitSchedule handles GET /patients/{patient_id}/schedules/{id}
func (h *VisitScheduleHandler) GetVisitSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	scheduleID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	schedule, err := h.visitScheduleService.GetVisitSchedule(ctx, patientID, scheduleID, userID)
	if err != nil {
		logger.Error("Failed to get visit schedule", err)
		if err.Error() == "access denied: you do not have permission to view this patient's schedules" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Visit schedule not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// GetVisitSchedules handles GET /patients/{patient_id}/schedules
func (h *VisitScheduleHandler) GetVisitSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := &models.VisitScheduleFilter{
		PatientID: &patientID,
	}

	// Parse visit_date_from
	if dateFrom := r.URL.Query().Get("visit_date_from"); dateFrom != "" {
		t, err := time.Parse("2006-01-02", dateFrom)
		if err != nil {
			http.Error(w, "Invalid visit_date_from format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		filter.VisitDateFrom = &t
	}

	// Parse visit_date_to
	if dateTo := r.URL.Query().Get("visit_date_to"); dateTo != "" {
		t, err := time.Parse("2006-01-02", dateTo)
		if err != nil {
			http.Error(w, "Invalid visit_date_to format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		filter.VisitDateTo = &t
	}

	// Parse visit_type
	if visitType := r.URL.Query().Get("visit_type"); visitType != "" {
		filter.VisitType = &visitType
	}

	// Parse assigned_staff_id
	if staffID := r.URL.Query().Get("assigned_staff_id"); staffID != "" {
		filter.AssignedStaffID = &staffID
	}

	// Parse assigned_vehicle_id
	if vehicleID := r.URL.Query().Get("assigned_vehicle_id"); vehicleID != "" {
		filter.AssignedVehicleID = &vehicleID
	}

	// Parse status
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	// Parse priority_score_min
	if priorityMin := r.URL.Query().Get("priority_score_min"); priorityMin != "" {
		val, err := strconv.Atoi(priorityMin)
		if err != nil {
			http.Error(w, "Invalid priority_score_min", http.StatusBadRequest)
			return
		}
		filter.PriorityScoreMin = &val
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

	schedules, err := h.visitScheduleService.ListVisitSchedules(ctx, filter, userID)
	if err != nil {
		logger.Error("Failed to list visit schedules", err)
		if err.Error() == "access denied: you do not have permission to view this patient's schedules" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to retrieve visit schedules", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

// UpdateVisitSchedule handles PUT /patients/{patient_id}/schedules/{id}
func (h *VisitScheduleHandler) UpdateVisitSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	scheduleID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.VisitScheduleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	schedule, err := h.visitScheduleService.UpdateVisitSchedule(ctx, patientID, scheduleID, &req, userID)
	if err != nil {
		logger.Error("Failed to update visit schedule", err)
		if err.Error() == "visit schedule not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to update schedules for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// DeleteVisitSchedule handles DELETE /patients/{patient_id}/schedules/{id}
func (h *VisitScheduleHandler) DeleteVisitSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	scheduleID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.visitScheduleService.DeleteVisitSchedule(ctx, patientID, scheduleID, userID)
	if err != nil {
		logger.Error("Failed to delete visit schedule", err)
		if err.Error() == "access denied: you do not have permission to delete schedules for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUpcomingSchedules handles GET /patients/{patient_id}/schedules/upcoming
func (h *VisitScheduleHandler) GetUpcomingSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse days parameter (default 7)
	days := 7
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		val, err := strconv.Atoi(daysStr)
		if err != nil || val <= 0 {
			http.Error(w, "Invalid days parameter", http.StatusBadRequest)
			return
		}
		days = val
	}

	schedules, err := h.visitScheduleService.GetUpcomingSchedules(ctx, patientID, days, userID)
	if err != nil {
		logger.Error("Failed to get upcoming schedules", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to view this patient's schedules" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to retrieve upcoming schedules", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

// AssignStaff handles POST /patients/{patient_id}/schedules/{id}/assign-staff
func (h *VisitScheduleHandler) AssignStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	scheduleID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		StaffID string `json:"staff_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.StaffID == "" {
		http.Error(w, "staff_id is required", http.StatusBadRequest)
		return
	}

	schedule, err := h.visitScheduleService.AssignStaff(ctx, patientID, scheduleID, req.StaffID, userID)
	if err != nil {
		logger.Error("Failed to assign staff", err)
		if err.Error() == "access denied: you do not have permission to assign staff for this patient's schedules" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// UpdateStatus handles POST /patients/{patient_id}/schedules/{id}/status
func (h *VisitScheduleHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	scheduleID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	schedule, err := h.visitScheduleService.UpdateStatus(ctx, patientID, scheduleID, req.Status, userID)
	if err != nil {
		logger.Error("Failed to update status", err)
		if err.Error() == "access denied: you do not have permission to update status for this patient's schedules" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}
