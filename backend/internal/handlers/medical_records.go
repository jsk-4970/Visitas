package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/services"
	"github.com/visitas/backend/pkg/logger"
)

// MedicalRecordHandler handles HTTP requests for medical records
type MedicalRecordHandler struct {
	medicalRecordService *services.MedicalRecordService
}

// NewMedicalRecordHandler creates a new medical record handler
func NewMedicalRecordHandler(medicalRecordService *services.MedicalRecordService) *MedicalRecordHandler {
	return &MedicalRecordHandler{
		medicalRecordService: medicalRecordService,
	}
}

// CreateMedicalRecord handles POST /patients/{patient_id}/medical-records
func (h *MedicalRecordHandler) CreateMedicalRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var req models.MedicalRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create record
	record, err := h.medicalRecordService.CreateRecord(ctx, patientID, &req, userID)
	if err != nil {
		logger.Error("Failed to create medical record", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// GetMedicalRecord handles GET /patients/{patient_id}/medical-records/{id}
func (h *MedicalRecordHandler) GetMedicalRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	recordID := chi.URLParam(r, "id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get record
	record, err := h.medicalRecordService.GetRecord(ctx, patientID, recordID, userID)
	if err != nil {
		logger.Error("Failed to get medical record", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// ListMedicalRecords handles GET /patients/{patient_id}/medical-records
func (h *MedicalRecordHandler) ListMedicalRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := &models.MedicalRecordFilter{}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	if visitType := r.URL.Query().Get("visit_type"); visitType != "" {
		filter.VisitType = &visitType
	}
	if performedBy := r.URL.Query().Get("performed_by"); performedBy != "" {
		filter.PerformedBy = &performedBy
	}
	if sourceType := r.URL.Query().Get("source_type"); sourceType != "" {
		filter.SourceType = &sourceType
	}
	if templateID := r.URL.Query().Get("template_id"); templateID != "" {
		filter.TemplateID = &templateID
	}
	if scheduleID := r.URL.Query().Get("schedule_id"); scheduleID != "" {
		filter.ScheduleID = &scheduleID
	}
	if soapCompletedStr := r.URL.Query().Get("soap_completed"); soapCompletedStr != "" {
		soapCompleted := soapCompletedStr == "true"
		filter.SOAPCompleted = &soapCompleted
	}
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.VisitDateFrom = &t
		}
	}
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.VisitDateTo = &t
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	// List records
	records, err := h.medicalRecordService.ListRecords(ctx, patientID, filter, userID)
	if err != nil {
		logger.Error("Failed to list medical records", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// UpdateMedicalRecord handles PUT /patients/{patient_id}/medical-records/{id}
func (h *MedicalRecordHandler) UpdateMedicalRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	recordID := chi.URLParam(r, "id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var req models.MedicalRecordUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update record
	record, err := h.medicalRecordService.UpdateRecord(ctx, patientID, recordID, &req, userID)
	if err != nil {
		logger.Error("Failed to update medical record", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "CONFLICT") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "CONFLICT",
				"message": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// DeleteMedicalRecord handles DELETE /patients/{patient_id}/medical-records/{id}
func (h *MedicalRecordHandler) DeleteMedicalRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	recordID := chi.URLParam(r, "id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete record
	err := h.medicalRecordService.DeleteRecord(ctx, patientID, recordID, userID)
	if err != nil {
		logger.Error("Failed to delete medical record", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CopyMedicalRecord handles POST /medical-records/{record_id}/copy
func (h *MedicalRecordHandler) CopyMedicalRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recordID := chi.URLParam(r, "record_id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var req models.CopyAsMedicalRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get source patient ID from query param (required for copy operation)
	sourcePatientID := r.URL.Query().Get("source_patient_id")
	if sourcePatientID == "" {
		http.Error(w, "source_patient_id query parameter is required", http.StatusBadRequest)
		return
	}

	// Determine target patient ID
	targetPatientID := sourcePatientID
	if req.TargetPatientID != nil {
		targetPatientID = *req.TargetPatientID
	}

	// Copy record
	record, err := h.medicalRecordService.CopyRecord(ctx, sourcePatientID, recordID, targetPatientID, &req, userID)
	if err != nil {
		logger.Error("Failed to copy medical record", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// CreateFromTemplate handles POST /patients/{patient_id}/medical-records/from-template
func (h *MedicalRecordHandler) CreateFromTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var req models.CreateFromTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create from template
	record, err := h.medicalRecordService.CreateFromTemplate(ctx, patientID, &req, userID)
	if err != nil {
		logger.Error("Failed to create medical record from template", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// GetLatestRecords handles GET /patients/{patient_id}/medical-records/latest
func (h *MedicalRecordHandler) GetLatestRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse limit
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get latest records
	records, err := h.medicalRecordService.GetLatestRecords(ctx, patientID, limit, userID)
	if err != nil {
		logger.Error("Failed to get latest medical records", err)
		if strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// GetDraftRecords handles GET /medical-records/drafts
func (h *MedicalRecordHandler) GetDraftRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get draft records
	records, err := h.medicalRecordService.GetDraftRecords(ctx, userID)
	if err != nil {
		logger.Error("Failed to get draft records", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}
