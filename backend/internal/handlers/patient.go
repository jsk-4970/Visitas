package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
)

type PatientHandler struct {
	repo *repository.SpannerRepository
}

func NewPatientHandler(repo *repository.SpannerRepository) *PatientHandler {
	return &PatientHandler{
		repo: repo,
	}
}

func (h *PatientHandler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement list patients with pagination and filters
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "List patients - TODO",
		"data":    []models.Patient{},
	})
}

func (h *PatientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "patient ID is required")
		return
	}

	// TODO: Implement get patient by ID
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Get patient - TODO",
		"id":      id,
	})
}

func (h *PatientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var patient models.Patient
	if err := json.NewDecoder(r.Body).Decode(&patient); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// TODO: Implement create patient
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Create patient - TODO",
		"data":    patient,
	})
}

func (h *PatientHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "patient ID is required")
		return
	}

	var patient models.Patient
	if err := json.NewDecoder(r.Body).Decode(&patient); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// TODO: Implement update patient
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Update patient - TODO",
		"id":      id,
		"data":    patient,
	})
}

func (h *PatientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "patient ID is required")
		return
	}

	// TODO: Implement soft delete patient
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Delete patient - TODO",
		"id":      id,
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
