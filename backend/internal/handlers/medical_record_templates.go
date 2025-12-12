package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/services"
	"github.com/visitas/backend/pkg/logger"
)

// MedicalRecordTemplateHandler handles HTTP requests for medical record templates
type MedicalRecordTemplateHandler struct {
	templateService *services.MedicalRecordTemplateService
}

// NewMedicalRecordTemplateHandler creates a new template handler
func NewMedicalRecordTemplateHandler(templateService *services.MedicalRecordTemplateService) *MedicalRecordTemplateHandler {
	return &MedicalRecordTemplateHandler{
		templateService: templateService,
	}
}

// CreateTemplate handles POST /medical-record-templates
func (h *MedicalRecordTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var req models.MedicalRecordTemplateCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create template
	template, err := h.templateService.CreateTemplate(ctx, &req, userID)
	if err != nil {
		logger.Error("Failed to create template", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

// GetTemplate handles GET /medical-record-templates/{id}
func (h *MedicalRecordTemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	templateID := chi.URLParam(r, "id")

	// Extract user ID from context (for logging purposes)
	_, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get template
	template, err := h.templateService.GetTemplate(ctx, templateID)
	if err != nil {
		logger.Error("Failed to get template", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// ListTemplates handles GET /medical-record-templates
func (h *MedicalRecordTemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from context (for logging purposes)
	_, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := &models.MedicalRecordTemplateFilter{}

	if specialty := r.URL.Query().Get("specialty"); specialty != "" {
		filter.Specialty = &specialty
	}
	if isSystemStr := r.URL.Query().Get("is_system_template"); isSystemStr != "" {
		isSystem := isSystemStr == "true"
		filter.IsSystemTemplate = &isSystem
	}
	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		filter.CreatedBy = &createdBy
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

	// List templates
	templates, err := h.templateService.ListTemplates(ctx, filter)
	if err != nil {
		logger.Error("Failed to list templates", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// UpdateTemplate handles PUT /medical-record-templates/{id}
func (h *MedicalRecordTemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	templateID := chi.URLParam(r, "id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var req models.MedicalRecordTemplateUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update template
	template, err := h.templateService.UpdateTemplate(ctx, templateID, &req, userID)
	if err != nil {
		logger.Error("Failed to update template", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// DeleteTemplate handles DELETE /medical-record-templates/{id}
func (h *MedicalRecordTemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	templateID := chi.URLParam(r, "id")

	// Extract user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete template
	err := h.templateService.DeleteTemplate(ctx, templateID, userID)
	if err != nil {
		logger.Error("Failed to delete template", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "cannot delete system template") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetSystemTemplates handles GET /medical-record-templates/system
func (h *MedicalRecordTemplateHandler) GetSystemTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from context (for logging purposes)
	_, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get system templates
	templates, err := h.templateService.GetSystemTemplates(ctx)
	if err != nil {
		logger.Error("Failed to get system templates", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// GetTemplatesBySpecialty handles GET /medical-record-templates/specialty/{specialty}
func (h *MedicalRecordTemplateHandler) GetTemplatesBySpecialty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	specialty := chi.URLParam(r, "specialty")

	// Extract user ID from context (for logging purposes)
	_, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get templates by specialty
	templates, err := h.templateService.GetTemplatesBySpecialty(ctx, specialty)
	if err != nil {
		logger.Error("Failed to get templates by specialty", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}
