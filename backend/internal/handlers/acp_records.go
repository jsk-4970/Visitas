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

// ACPRecordHandler handles HTTP requests for ACP records
type ACPRecordHandler struct {
	acpRecordService *services.ACPRecordService
}

// NewACPRecordHandler creates a new ACP record handler
func NewACPRecordHandler(acpRecordService *services.ACPRecordService) *ACPRecordHandler {
	return &ACPRecordHandler{
		acpRecordService: acpRecordService,
	}
}

// CreateACPRecord handles POST /patients/{patient_id}/acp-records
func (h *ACPRecordHandler) CreateACPRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ACPRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	record, err := h.acpRecordService.CreateACPRecord(ctx, patientID, &req, userID)
	if err != nil {
		logger.Error("Failed to create ACP record", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to create ACP records for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(record); err != nil {
		logger.Error("Failed to encode response", err)
	}
}

// GetACPRecord handles GET /patients/{patient_id}/acp-records/{id}
func (h *ACPRecordHandler) GetACPRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	acpID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	record, err := h.acpRecordService.GetACPRecord(ctx, patientID, acpID, userID)
	if err != nil {
		logger.Error("Failed to get ACP record", err)
		if err.Error() == "access denied: you do not have permission to view this patient's ACP records" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "ACP record not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(record); err != nil {
		logger.Error("Failed to encode response", err)
	}
}

// GetACPRecords handles GET /patients/{patient_id}/acp-records
func (h *ACPRecordHandler) GetACPRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := &models.ACPRecordFilter{
		PatientID: &patientID,
	}

	// Parse status
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	// Parse decision_maker
	if decisionMaker := r.URL.Query().Get("decision_maker"); decisionMaker != "" {
		filter.DecisionMaker = &decisionMaker
	}

	// Parse recorded_from
	if recordedFrom := r.URL.Query().Get("recorded_from"); recordedFrom != "" {
		t, err := time.Parse("2006-01-02", recordedFrom)
		if err != nil {
			http.Error(w, "Invalid recorded_from format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		filter.RecordedFrom = &t
	}

	// Parse recorded_to
	if recordedTo := r.URL.Query().Get("recorded_to"); recordedTo != "" {
		t, err := time.Parse("2006-01-02", recordedTo)
		if err != nil {
			http.Error(w, "Invalid recorded_to format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		filter.RecordedTo = &t
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

	records, err := h.acpRecordService.ListACPRecords(ctx, filter, userID)
	if err != nil {
		logger.Error("Failed to list ACP records", err)
		if err.Error() == "access denied: you do not have permission to view this patient's ACP records" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to retrieve ACP records", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(records); err != nil {
		logger.Error("Failed to encode response", err)
	}
}

// UpdateACPRecord handles PUT /patients/{patient_id}/acp-records/{id}
func (h *ACPRecordHandler) UpdateACPRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	acpID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ACPRecordUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	record, err := h.acpRecordService.UpdateACPRecord(ctx, patientID, acpID, &req, userID)
	if err != nil {
		logger.Error("Failed to update ACP record", err)
		if err.Error() == "ACP record not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to update ACP records for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(record); err != nil {
		logger.Error("Failed to encode response", err)
	}
}

// DeleteACPRecord handles DELETE /patients/{patient_id}/acp-records/{id}
func (h *ACPRecordHandler) DeleteACPRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	acpID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.acpRecordService.DeleteACPRecord(ctx, patientID, acpID, userID)
	if err != nil {
		logger.Error("Failed to delete ACP record", err)
		if err.Error() == "access denied: you do not have permission to delete ACP records for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetLatestACP handles GET /patients/{patient_id}/acp-records/latest
func (h *ACPRecordHandler) GetLatestACP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	record, err := h.acpRecordService.GetLatestACP(ctx, patientID, userID)
	if err != nil {
		logger.Error("Failed to get latest ACP record", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "no active ACP record found for patient" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to view this patient's ACP records" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to retrieve latest ACP record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(record); err != nil {
		logger.Error("Failed to encode response", err)
	}
}

// GetACPHistory handles GET /patients/{patient_id}/acp-records/history
func (h *ACPRecordHandler) GetACPHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	records, err := h.acpRecordService.GetACPHistory(ctx, patientID, userID)
	if err != nil {
		logger.Error("Failed to get ACP history", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to view this patient's ACP records" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to retrieve ACP history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(records); err != nil {
		logger.Error("Failed to encode response", err)
	}
}
