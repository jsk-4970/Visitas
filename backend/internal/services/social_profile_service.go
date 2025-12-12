package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// SocialProfileService handles business logic for social profile operations
type SocialProfileService struct {
	socialProfileRepo *repository.SocialProfileRepository
	patientRepo       *repository.PatientRepository
}

// NewSocialProfileService creates a new social profile service
func NewSocialProfileService(
	socialProfileRepo *repository.SocialProfileRepository,
	patientRepo *repository.PatientRepository,
) *SocialProfileService {
	return &SocialProfileService{
		socialProfileRepo: socialProfileRepo,
		patientRepo:       patientRepo,
	}
}

// CreateSocialProfile creates a new social profile with access control
func (s *SocialProfileService) CreateSocialProfile(ctx context.Context, req *models.PatientSocialProfileCreateRequest, createdBy string) (*models.PatientSocialProfile, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		logger.WarnContext(ctx, "Invalid social profile create request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if user has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, createdBy, req.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized social profile creation attempt", map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to create social profile for this patient")
	}

	// Create social profile
	profile, err := s.socialProfileRepo.CreateSocialProfile(ctx, req, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create social profile", err, map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create social profile: %w", err)
	}

	logger.InfoContext(ctx, "Social profile created successfully", map[string]interface{}{
		"profile_id": profile.ProfileID,
		"patient_id": profile.PatientID,
		"created_by": createdBy,
	})

	return profile, nil
}

// GetSocialProfile retrieves a social profile by ID with access control
func (s *SocialProfileService) GetSocialProfile(ctx context.Context, profileID, requestorID string) (*models.PatientSocialProfile, error) {
	// Get social profile first to check patient ID
	profile, err := s.socialProfileRepo.GetSocialProfileByID(ctx, profileID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get social profile", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return nil, fmt.Errorf("failed to get social profile: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, profile.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"profile_id":   profileID,
			"patient_id":   profile.PatientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized social profile access attempt", map[string]interface{}{
			"profile_id":   profileID,
			"patient_id":   profile.PatientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this social profile")
	}

	return profile, nil
}

// GetCurrentSocialProfile retrieves the current valid social profile for a patient with access control
func (s *SocialProfileService) GetCurrentSocialProfile(ctx context.Context, patientID, requestorID string) (*models.PatientSocialProfile, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized current social profile access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view social profile for this patient")
	}

	profile, err := s.socialProfileRepo.GetCurrentSocialProfile(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get current social profile", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get current social profile: %w", err)
	}

	return profile, nil
}

// GetSocialProfileHistory retrieves all social profiles for a patient with access control
func (s *SocialProfileService) GetSocialProfileHistory(ctx context.Context, patientID, requestorID string) ([]*models.PatientSocialProfile, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized social profile history access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view social profile history for this patient")
	}

	profiles, err := s.socialProfileRepo.GetSocialProfileHistory(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get social profile history", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get social profile history: %w", err)
	}

	return profiles, nil
}

// UpdateSocialProfile updates a social profile with access control
func (s *SocialProfileService) UpdateSocialProfile(ctx context.Context, profileID string, req *models.PatientSocialProfileUpdateRequest, updatedBy string) (*models.PatientSocialProfile, error) {
	// Get social profile first to check patient ID
	profile, err := s.socialProfileRepo.GetSocialProfileByID(ctx, profileID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get social profile for update", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return nil, fmt.Errorf("social profile not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, profile.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"profile_id": profileID,
			"patient_id": profile.PatientID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized social profile update attempt", map[string]interface{}{
			"profile_id": profileID,
			"patient_id": profile.PatientID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this social profile")
	}

	// Update social profile
	updatedProfile, err := s.socialProfileRepo.UpdateSocialProfile(ctx, profileID, req, updatedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update social profile", err, map[string]interface{}{
			"profile_id": profileID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to update social profile: %w", err)
	}

	logger.InfoContext(ctx, "Social profile updated successfully", map[string]interface{}{
		"profile_id": updatedProfile.ProfileID,
		"patient_id": updatedProfile.PatientID,
		"updated_by": updatedBy,
	})

	return updatedProfile, nil
}

// DeleteSocialProfile soft deletes a social profile with access control
func (s *SocialProfileService) DeleteSocialProfile(ctx context.Context, profileID, deletedBy string) error {
	// Get social profile first to check patient ID
	profile, err := s.socialProfileRepo.GetSocialProfileByID(ctx, profileID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get social profile for deletion", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return fmt.Errorf("social profile not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, profile.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"profile_id": profileID,
			"patient_id": profile.PatientID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized social profile delete attempt", map[string]interface{}{
			"profile_id": profileID,
			"patient_id": profile.PatientID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this social profile")
	}

	// Delete social profile
	err = s.socialProfileRepo.DeleteSocialProfile(ctx, profileID, deletedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete social profile", err, map[string]interface{}{
			"profile_id": profileID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to delete social profile: %w", err)
	}

	logger.InfoContext(ctx, "Social profile deleted successfully", map[string]interface{}{
		"profile_id": profileID,
		"patient_id": profile.PatientID,
		"deleted_by": deletedBy,
	})

	return nil
}

// validateCreateRequest validates social profile create request
func (s *SocialProfileService) validateCreateRequest(req *models.PatientSocialProfileCreateRequest) error {
	if req.PatientID == "" {
		return fmt.Errorf("patient_id is required")
	}

	if req.ValidFrom.IsZero() {
		return fmt.Errorf("valid_from is required")
	}

	// Validate content is not empty
	if req.Content.LivingSituation == nil && len(req.Content.KeyPersons) == 0 &&
	   req.Content.FinancialBackground == nil && req.Content.SocialSupport == nil {
		return fmt.Errorf("content must contain at least one section (livingSituation, keyPersons, financialBackground, or socialSupport)")
	}

	// Validate housing type if provided
	if req.Content.LivingSituation != nil {
		validHousingTypes := []string{"detached", "apartment", "facility", "other"}
		if req.Content.LivingSituation.HousingType != "" && !isValidValue(req.Content.LivingSituation.HousingType, validHousingTypes) {
			return fmt.Errorf("invalid housing_type: must be one of [detached, apartment, facility, other]")
		}
	}

	return nil
}
