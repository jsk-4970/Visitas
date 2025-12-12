package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// MedicalConditionService handles business logic for medical condition operations
type MedicalConditionService struct {
	conditionRepo *repository.MedicalConditionRepository
	patientRepo   *repository.PatientRepository
}

// NewMedicalConditionService creates a new medical condition service
func NewMedicalConditionService(
	conditionRepo *repository.MedicalConditionRepository,
	patientRepo *repository.PatientRepository,
) *MedicalConditionService {
	return &MedicalConditionService{
		conditionRepo: conditionRepo,
		patientRepo:   patientRepo,
	}
}

// CreateCondition creates a new medical condition with access control
func (s *MedicalConditionService) CreateCondition(ctx context.Context, req *models.MedicalConditionCreateRequest, createdBy string) (*models.MedicalCondition, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		logger.WarnContext(ctx, "Invalid medical condition create request", map[string]interface{}{
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
		logger.WarnContext(ctx, "Unauthorized condition creation attempt", map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to add conditions for this patient")
	}

	// Create condition
	condition, err := s.conditionRepo.CreateCondition(ctx, req, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create condition", err, map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create condition: %w", err)
	}

	logger.InfoContext(ctx, "Medical condition created successfully", map[string]interface{}{
		"condition_id": condition.ConditionID,
		"patient_id":   condition.PatientID,
		"created_by":   createdBy,
	})

	return condition, nil
}

// GetCondition retrieves a condition by ID with access control
func (s *MedicalConditionService) GetCondition(ctx context.Context, conditionID, requestorID string) (*models.MedicalCondition, error) {
	// Get condition first to check patient ID
	condition, err := s.conditionRepo.GetConditionByID(ctx, conditionID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get condition", err, map[string]interface{}{
			"condition_id": conditionID,
		})
		return nil, fmt.Errorf("failed to get condition: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, condition.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"condition_id": conditionID,
			"patient_id":   condition.PatientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized condition access attempt", map[string]interface{}{
			"condition_id": conditionID,
			"patient_id":   condition.PatientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this condition")
	}

	return condition, nil
}

// GetActiveConditions retrieves all active conditions for a patient with access control
func (s *MedicalConditionService) GetActiveConditions(ctx context.Context, patientID, requestorID string) ([]*models.MedicalCondition, error) {
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
		logger.WarnContext(ctx, "Unauthorized active conditions access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view conditions for this patient")
	}

	conditions, err := s.conditionRepo.GetActiveConditions(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get active conditions", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get active conditions: %w", err)
	}

	return conditions, nil
}

// GetConditionsByPatient retrieves all conditions for a patient with access control
func (s *MedicalConditionService) GetConditionsByPatient(ctx context.Context, patientID, requestorID string) ([]*models.MedicalCondition, error) {
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
		logger.WarnContext(ctx, "Unauthorized conditions access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view conditions for this patient")
	}

	conditions, err := s.conditionRepo.GetConditionsByPatient(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get conditions", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get conditions: %w", err)
	}

	return conditions, nil
}

// UpdateCondition updates a condition with access control
func (s *MedicalConditionService) UpdateCondition(ctx context.Context, conditionID string, req *models.MedicalConditionUpdateRequest, updatedBy string) (*models.MedicalCondition, error) {
	// Get condition first to check patient ID
	condition, err := s.conditionRepo.GetConditionByID(ctx, conditionID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get condition for update", err, map[string]interface{}{
			"condition_id": conditionID,
		})
		return nil, fmt.Errorf("condition not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, condition.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"condition_id": conditionID,
			"patient_id":   condition.PatientID,
			"updated_by":   updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized condition update attempt", map[string]interface{}{
			"condition_id": conditionID,
			"patient_id":   condition.PatientID,
			"updated_by":   updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this condition")
	}

	// Update condition
	updatedCondition, err := s.conditionRepo.UpdateCondition(ctx, conditionID, req, updatedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update condition", err, map[string]interface{}{
			"condition_id": conditionID,
			"updated_by":   updatedBy,
		})
		return nil, fmt.Errorf("failed to update condition: %w", err)
	}

	logger.InfoContext(ctx, "Medical condition updated successfully", map[string]interface{}{
		"condition_id": updatedCondition.ConditionID,
		"patient_id":   updatedCondition.PatientID,
		"updated_by":   updatedBy,
	})

	return updatedCondition, nil
}

// DeleteCondition soft deletes a condition with access control
func (s *MedicalConditionService) DeleteCondition(ctx context.Context, conditionID, deletedBy string) error {
	// Get condition first to check patient ID
	condition, err := s.conditionRepo.GetConditionByID(ctx, conditionID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get condition for deletion", err, map[string]interface{}{
			"condition_id": conditionID,
		})
		return fmt.Errorf("condition not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, condition.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"condition_id": conditionID,
			"patient_id":   condition.PatientID,
			"deleted_by":   deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized condition delete attempt", map[string]interface{}{
			"condition_id": conditionID,
			"patient_id":   condition.PatientID,
			"deleted_by":   deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this condition")
	}

	// Delete condition
	err = s.conditionRepo.DeleteCondition(ctx, conditionID, deletedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete condition", err, map[string]interface{}{
			"condition_id": conditionID,
			"deleted_by":   deletedBy,
		})
		return fmt.Errorf("failed to delete condition: %w", err)
	}

	logger.InfoContext(ctx, "Medical condition deleted successfully", map[string]interface{}{
		"condition_id": conditionID,
		"patient_id":   condition.PatientID,
		"deleted_by":   deletedBy,
	})

	return nil
}

// validateCreateRequest validates condition create request
func (s *MedicalConditionService) validateCreateRequest(req *models.MedicalConditionCreateRequest) error {
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
		string(models.ClinicalStatusActive),
		string(models.ClinicalStatusRecurrence),
		string(models.ClinicalStatusRelapse),
		string(models.ClinicalStatusInactive),
		string(models.ClinicalStatusRemission),
		string(models.ClinicalStatusResolved),
	}
	if !isValidValue(req.ClinicalStatus, validClinicalStatuses) {
		return fmt.Errorf("invalid clinical_status: must be one of [active, recurrence, relapse, inactive, remission, resolved]")
	}

	if req.VerificationStatus == "" {
		return fmt.Errorf("verification_status is required")
	}

	// Validate verification status
	validVerificationStatuses := []string{
		string(models.VerificationStatusUnconfirmed),
		string(models.VerificationStatusProvisional),
		string(models.VerificationStatusDifferential),
		string(models.VerificationStatusConfirmed),
		string(models.VerificationStatusRefuted),
		string(models.VerificationStatusEnteredInError),
	}
	if !isValidValue(req.VerificationStatus, validVerificationStatuses) {
		return fmt.Errorf("invalid verification_status: must be one of [unconfirmed, provisional, differential, confirmed, refuted, entered-in-error]")
	}

	// Validate severity if provided
	if req.Severity != "" {
		validSeverities := []string{
			string(models.SeverityMild),
			string(models.SeverityModerate),
			string(models.SeveritySevere),
			string(models.SeverityLifeThreatening),
		}
		if !isValidValue(req.Severity, validSeverities) {
			return fmt.Errorf("invalid severity: must be one of [mild, moderate, severe, life-threatening]")
		}
	}

	return nil
}

// isValidValue checks if a value exists in a list of valid values
func isValidValue(value string, validValues []string) bool {
	for _, v := range validValues {
		if value == v {
			return true
		}
	}
	return false
}
