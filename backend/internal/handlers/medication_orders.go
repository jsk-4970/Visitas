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

// MedicationOrderHandler handles HTTP requests for medication orders
type MedicationOrderHandler struct {
	medicationOrderService *services.MedicationOrderService
}

// NewMedicationOrderHandler creates a new medication order handler
func NewMedicationOrderHandler(medicationOrderService *services.MedicationOrderService) *MedicationOrderHandler {
	return &MedicationOrderHandler{
		medicationOrderService: medicationOrderService,
	}
}

// CreateMedicationOrder handles POST /patients/{patient_id}/medication-orders
func (h *MedicationOrderHandler) CreateMedicationOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.MedicationOrderCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.medicationOrderService.CreateMedicationOrder(ctx, patientID, &req, userID)
	if err != nil {
		logger.Error("Failed to create medication order", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to create medication orders for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// GetMedicationOrder handles GET /patients/{patient_id}/medication-orders/{id}
func (h *MedicationOrderHandler) GetMedicationOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	orderID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	order, err := h.medicationOrderService.GetMedicationOrder(ctx, patientID, orderID, userID)
	if err != nil {
		logger.Error("Failed to get medication order", err)
		if err.Error() == "access denied: you do not have permission to view this patient's medication orders" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Medication order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// GetMedicationOrders handles GET /patients/{patient_id}/medication-orders
func (h *MedicationOrderHandler) GetMedicationOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := &models.MedicationOrderFilter{
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

	// Parse prescribed_by
	if prescribedBy := r.URL.Query().Get("prescribed_by"); prescribedBy != "" {
		filter.PrescribedBy = &prescribedBy
	}

	// Parse prescribed_date_from
	if dateFrom := r.URL.Query().Get("prescribed_date_from"); dateFrom != "" {
		t, err := time.Parse("2006-01-02", dateFrom)
		if err != nil {
			http.Error(w, "Invalid prescribed_date_from format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		filter.PrescribedDateFrom = &t
	}

	// Parse prescribed_date_to
	if dateTo := r.URL.Query().Get("prescribed_date_to"); dateTo != "" {
		t, err := time.Parse("2006-01-02", dateTo)
		if err != nil {
			http.Error(w, "Invalid prescribed_date_to format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		filter.PrescribedDateTo = &t
	}

	// Parse reason_reference
	if reasonRef := r.URL.Query().Get("reason_reference"); reasonRef != "" {
		filter.ReasonReference = &reasonRef
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

	orders, err := h.medicationOrderService.ListMedicationOrders(ctx, filter, userID)
	if err != nil {
		logger.Error("Failed to list medication orders", err)
		if err.Error() == "access denied: you do not have permission to view this patient's medication orders" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to retrieve medication orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// UpdateMedicationOrder handles PUT /patients/{patient_id}/medication-orders/{id}
func (h *MedicationOrderHandler) UpdateMedicationOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	orderID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.MedicationOrderUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.medicationOrderService.UpdateMedicationOrder(ctx, patientID, orderID, &req, userID)
	if err != nil {
		logger.Error("Failed to update medication order", err)
		if err.Error() == "medication order not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to update medication orders for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// DeleteMedicationOrder handles DELETE /patients/{patient_id}/medication-orders/{id}
func (h *MedicationOrderHandler) DeleteMedicationOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")
	orderID := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.medicationOrderService.DeleteMedicationOrder(ctx, patientID, orderID, userID)
	if err != nil {
		logger.Error("Failed to delete medication order", err)
		if err.Error() == "access denied: you do not have permission to delete medication orders for this patient" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetActiveOrders handles GET /patients/{patient_id}/medication-orders/active
func (h *MedicationOrderHandler) GetActiveOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	patientID := chi.URLParam(r, "patient_id")

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.medicationOrderService.GetActiveOrders(ctx, patientID, userID)
	if err != nil {
		logger.Error("Failed to get active orders", err)
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "access denied: you do not have permission to view this patient's medication orders" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to retrieve active medication orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
