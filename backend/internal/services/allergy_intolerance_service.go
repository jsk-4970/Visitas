package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// AllergyIntoleranceService handles business logic for allergy/intolerance operations
type AllergyIntoleranceService struct {
	allergyRepo *repository.AllergyIntoleranceRepository
	patientRepo *repository.PatientRepository
}

// NewAllergyIntoleranceService creates a new allergy intolerance service
func NewAllergyIntoleranceService(
	allergyRepo *repository.AllergyIntoleranceRepository,
	patientRepo *repository.PatientRepository,
) *AllergyIntoleranceService {
	return &AllergyIntoleranceService{
		allergyRepo: allergyRepo,
		patientRepo: patientRepo,
	}
}

// CreateAllergy creates a new allergy intolerance with access control
func (s *AllergyIntoleranceService) CreateAllergy(ctx context.Context, req *models.AllergyIntoleranceCreateRequest, createdBy string) (*models.AllergyIntolerance, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		logger.WarnContext(ctx, "Invalid allergy create request", map[string]interface{}{
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
		logger.WarnContext(ctx, "Unauthorized allergy creation attempt", map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to add allergies for this patient")
	}

	// Create allergy
	allergy, err := s.allergyRepo.CreateAllergy(ctx, req, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create allergy", err, map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create allergy: %w", err)
	}

	logger.InfoContext(ctx, "Allergy intolerance created successfully", map[string]interface{}{
		"allergy_id": allergy.AllergyID,
		"patient_id": allergy.PatientID,
		"created_by": createdBy,
		"category":   allergy.Category,
		"criticality": allergy.Criticality,
	})

	return allergy, nil
}

// GetAllergy retrieves an allergy by ID with access control
func (s *AllergyIntoleranceService) GetAllergy(ctx context.Context, allergyID, requestorID string) (*models.AllergyIntolerance, error) {
	// Get allergy first to check patient ID
	allergy, err := s.allergyRepo.GetAllergyByID(ctx, allergyID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get allergy", err, map[string]interface{}{
			"allergy_id": allergyID,
		})
		return nil, fmt.Errorf("failed to get allergy: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, allergy.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"allergy_id":   allergyID,
			"patient_id":   allergy.PatientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized allergy access attempt", map[string]interface{}{
			"allergy_id":   allergyID,
			"patient_id":   allergy.PatientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this allergy")
	}

	return allergy, nil
}

// GetActiveAllergies retrieves all active allergies for a patient with access control
func (s *AllergyIntoleranceService) GetActiveAllergies(ctx context.Context, patientID, requestorID string) ([]*models.AllergyIntolerance, error) {
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
		logger.WarnContext(ctx, "Unauthorized active allergies access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view allergies for this patient")
	}

	allergies, err := s.allergyRepo.GetActiveAllergies(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get active allergies", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get active allergies: %w", err)
	}

	return allergies, nil
}

// GetMedicationAllergies retrieves all medication allergies for a patient with access control
func (s *AllergyIntoleranceService) GetMedicationAllergies(ctx context.Context, patientID, requestorID string) ([]*models.AllergyIntolerance, error) {
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
		logger.WarnContext(ctx, "Unauthorized medication allergies access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view medication allergies for this patient")
	}

	allergies, err := s.allergyRepo.GetMedicationAllergies(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get medication allergies", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get medication allergies: %w", err)
	}

	return allergies, nil
}

// GetAllergiesByPatient retrieves all allergies for a patient with access control
func (s *AllergyIntoleranceService) GetAllergiesByPatient(ctx context.Context, patientID, requestorID string) ([]*models.AllergyIntolerance, error) {
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
		logger.WarnContext(ctx, "Unauthorized allergies access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view allergies for this patient")
	}

	allergies, err := s.allergyRepo.GetAllergiesByPatient(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get allergies", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get allergies: %w", err)
	}

	return allergies, nil
}

// UpdateAllergy updates an allergy with access control
func (s *AllergyIntoleranceService) UpdateAllergy(ctx context.Context, allergyID string, req *models.AllergyIntoleranceUpdateRequest, updatedBy string) (*models.AllergyIntolerance, error) {
	// Get allergy first to check patient ID
	allergy, err := s.allergyRepo.GetAllergyByID(ctx, allergyID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get allergy for update", err, map[string]interface{}{
			"allergy_id": allergyID,
		})
		return nil, fmt.Errorf("allergy not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, allergy.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"allergy_id": allergyID,
			"patient_id": allergy.PatientID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized allergy update attempt", map[string]interface{}{
			"allergy_id": allergyID,
			"patient_id": allergy.PatientID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this allergy")
	}

	// Update allergy
	updatedAllergy, err := s.allergyRepo.UpdateAllergy(ctx, allergyID, req, updatedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update allergy", err, map[string]interface{}{
			"allergy_id": allergyID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to update allergy: %w", err)
	}

	logger.InfoContext(ctx, "Allergy intolerance updated successfully", map[string]interface{}{
		"allergy_id": updatedAllergy.AllergyID,
		"patient_id": updatedAllergy.PatientID,
		"updated_by": updatedBy,
	})

	return updatedAllergy, nil
}

// DeleteAllergy soft deletes an allergy with access control
func (s *AllergyIntoleranceService) DeleteAllergy(ctx context.Context, allergyID, deletedBy string) error {
	// Get allergy first to check patient ID
	allergy, err := s.allergyRepo.GetAllergyByID(ctx, allergyID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get allergy for deletion", err, map[string]interface{}{
			"allergy_id": allergyID,
		})
		return fmt.Errorf("allergy not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, allergy.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"allergy_id": allergyID,
			"patient_id": allergy.PatientID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized allergy delete attempt", map[string]interface{}{
			"allergy_id": allergyID,
			"patient_id": allergy.PatientID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this allergy")
	}

	// Delete allergy
	err = s.allergyRepo.DeleteAllergy(ctx, allergyID, deletedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete allergy", err, map[string]interface{}{
			"allergy_id": allergyID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to delete allergy: %w", err)
	}

	logger.InfoContext(ctx, "Allergy intolerance deleted successfully", map[string]interface{}{
		"allergy_id": allergyID,
		"patient_id": allergy.PatientID,
		"deleted_by": deletedBy,
	})

	return nil
}

// validateCreateRequest validates allergy create request
func (s *AllergyIntoleranceService) validateCreateRequest(req *models.AllergyIntoleranceCreateRequest) error {
	if req.PatientID == "" {
		return fmt.Errorf("patient_id is required")
	}

	if req.DisplayName == "" {
		return fmt.Errorf("display_name is required")
	}

	if req.ClinicalStatus == "" {
		return fmt.Errorf("clinical_status is required")
	}

	// Validate clinical status
	validClinicalStatuses := []string{
		string(models.AllergyClinicalStatusActive),
		string(models.AllergyClinicalStatusInactive),
		string(models.AllergyClinicalStatusResolved),
	}
	if !isValidValue(req.ClinicalStatus, validClinicalStatuses) {
		return fmt.Errorf("invalid clinical_status: must be one of [active, inactive, resolved]")
	}

	if req.VerificationStatus == "" {
		return fmt.Errorf("verification_status is required")
	}

	// Validate verification status
	validVerificationStatuses := []string{
		string(models.AllergyVerificationStatusUnconfirmed),
		string(models.AllergyVerificationStatusConfirmed),
		string(models.AllergyVerificationStatusRefuted),
	}
	if !isValidValue(req.VerificationStatus, validVerificationStatuses) {
		return fmt.Errorf("invalid verification_status: must be one of [unconfirmed, confirmed, refuted]")
	}

	if req.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Validate type
	validTypes := []string{
		string(models.AllergyTypeAllergy),
		string(models.AllergyTypeIntolerance),
	}
	if !isValidValue(req.Type, validTypes) {
		return fmt.Errorf("invalid type: must be one of [allergy, intolerance]")
	}

	if req.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate category
	validCategories := []string{
		string(models.AllergyCategoryFood),
		string(models.AllergyCategoryMedication),
		string(models.AllergyCategoryEnvironment),
		string(models.AllergyCategoryBiologic),
	}
	if !isValidValue(req.Category, validCategories) {
		return fmt.Errorf("invalid category: must be one of [food, medication, environment, biologic]")
	}

	if req.Criticality == "" {
		return fmt.Errorf("criticality is required")
	}

	// Validate criticality
	validCriticalities := []string{
		string(models.AllergyCriticalityLow),
		string(models.AllergyCriticalityHigh),
		string(models.AllergyCriticalityUnableToAssess),
	}
	if !isValidValue(req.Criticality, validCriticalities) {
		return fmt.Errorf("invalid criticality: must be one of [low, high, unable-to-assess]")
	}

	return nil
}
