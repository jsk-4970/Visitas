package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// SocialProfileHandler handles social profile-related HTTP requests
type SocialProfileHandler struct {
	socialProfileRepo *repository.SocialProfileRepository
	patientRepo       *repository.PatientRepository
}

// NewSocialProfileHandler creates a new social profile handler
func NewSocialProfileHandler(socialProfileRepo *repository.SocialProfileRepository, patientRepo *repository.PatientRepository) *SocialProfileHandler {
	return &SocialProfileHandler{
		socialProfileRepo: socialProfileRepo,
		patientRepo:       patientRepo,
	}
}

// CreateSocialProfile handles POST /api/v1/patients/:patient_id/social-profiles
func (h *SocialProfileHandler) CreateSocialProfile(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
		return
	}

	var req models.PatientSocialProfileCreateRequest
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
	_, err := h.patientRepo.GetByID(r.Context(), patientID)
	if err != nil {
		if err.Error() == "patient not found" {
			respondError(w, http.StatusNotFound, "Patient not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to verify patient", err)
			respondError(w, http.StatusInternalServerError, "Failed to verify patient")
		}
		return
	}

	// Marshal content to JSON
	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to marshal content", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusBadRequest, "Invalid content format")
		return
	}

	profile, err := h.socialProfileRepo.Create(r.Context(), &req, json.RawMessage(contentJSON), userID)
	if err != nil {
		logger.ErrorContext(r.Context(), "Failed to create social profile", err)
		respondError(w, http.StatusInternalServerError, "Failed to create social profile")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"profile_id": profile.ProfileID,
		"created_at": profile.CreatedAt,
		"message":    "Social profile created successfully",
	})
}

// GetSocialProfiles handles GET /api/v1/patients/:patient_id/social-profiles
func (h *SocialProfileHandler) GetSocialProfiles(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	if patientID == "" {
		respondError(w, http.StatusBadRequest, "Patient ID is required")
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

	// Check for current_only parameter
	currentOnly := r.URL.Query().Get("current_only") == "true"

	var profiles []*models.PatientSocialProfile
	var err error

	if currentOnly {
		// Get only the current profile
		currentProfile, err := h.socialProfileRepo.GetCurrentByPatientID(r.Context(), patientID)
		if err != nil {
			if err.Error() == "no current social profile found" {
				respondJSON(w, http.StatusOK, map[string]interface{}{
					"profiles": []interface{}{},
					"total":    0,
					"page":     page,
					"per_page": perPage,
				})
				return
			}
			logger.ErrorContext(r.Context(), "Failed to get current social profile", err)
			respondError(w, http.StatusInternalServerError, "Failed to get current social profile")
			return
		}
		profiles = []*models.PatientSocialProfile{currentProfile}
	} else {
		// Get all profiles (versioned history)
		profiles, err = h.socialProfileRepo.GetByPatientID(r.Context(), patientID)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get social profiles", err)
			respondError(w, http.StatusInternalServerError, "Failed to get social profiles")
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"profiles": profiles,
		"total":    len(profiles),
		"page":     page,
		"per_page": perPage,
	})
}

// GetSocialProfile handles GET /api/v1/patients/:patient_id/social-profiles/:id
func (h *SocialProfileHandler) GetSocialProfile(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	profile, err := h.socialProfileRepo.GetByID(r.Context(), profileID)
	if err != nil {
		if err.Error() == "social profile not found" {
			respondError(w, http.StatusNotFound, "Social profile not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to get social profile", err)
			respondError(w, http.StatusInternalServerError, "Failed to get social profile")
		}
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

// UpdateSocialProfile handles PUT /api/v1/patients/:patient_id/social-profiles/:id
func (h *SocialProfileHandler) UpdateSocialProfile(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	var req models.PatientSocialProfileUpdateRequest
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

	profile, err := h.socialProfileRepo.Update(r.Context(), profileID, &req, userID)
	if err != nil {
		if err.Error() == "social profile not found" {
			respondError(w, http.StatusNotFound, "Social profile not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to update social profile", err)
			respondError(w, http.StatusInternalServerError, "Failed to update social profile")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"profile_id": profile.ProfileID,
		"updated_at": profile.UpdatedAt,
		"message":    "Social profile updated successfully",
	})
}

// DeleteSocialProfile handles DELETE /api/v1/patients/:patient_id/social-profiles/:id
func (h *SocialProfileHandler) DeleteSocialProfile(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.socialProfileRepo.Delete(r.Context(), profileID, userID)
	if err != nil {
		if err.Error() == "social profile not found" {
			respondError(w, http.StatusNotFound, "Social profile not found")
		} else {
			logger.ErrorContext(r.Context(), "Failed to delete social profile", err)
			respondError(w, http.StatusInternalServerError, "Failed to delete social profile")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Social profile deleted successfully",
	})
}
