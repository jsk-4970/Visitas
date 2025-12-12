package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// CarePlanService handles business logic for care plans
type CarePlanService struct {
	carePlanRepo *repository.CarePlanRepository
	patientRepo  *repository.PatientRepository
}

// NewCarePlanService creates a new care plan service
func NewCarePlanService(
	carePlanRepo *repository.CarePlanRepository,
	patientRepo *repository.PatientRepository,
) *CarePlanService {
	return &CarePlanService{
		carePlanRepo: carePlanRepo,
		patientRepo:  patientRepo,
	}
}

// CreateCarePlan creates a new care plan with validation and access control
func (s *CarePlanService) CreateCarePlan(ctx context.Context, patientID string, req *models.CarePlanCreateRequest, createdBy string) (*models.CarePlan, error) {
	// Check if user has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, createdBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized care plan creation attempt", map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to create care plans for this patient")
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft":     true,
		"active":    true,
		"on-hold":   true,
		"revoked":   true,
		"completed": true,
	}
	if !validStatuses[req.Status] {
		logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
			"status": req.Status,
		})
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate intent
	validIntents := map[string]bool{
		"proposal": true,
		"plan":     true,
		"order":    true,
	}
	if !validIntents[req.Intent] {
		logger.WarnContext(ctx, "Invalid intent", map[string]interface{}{
			"intent": req.Intent,
		})
		return nil, fmt.Errorf("invalid intent: %s", req.Intent)
	}

	// Validate period_end is after period_start if provided
	if req.PeriodEnd != nil && req.PeriodEnd.Before(req.PeriodStart) {
		logger.WarnContext(ctx, "Invalid period", map[string]interface{}{
			"period_start": req.PeriodStart,
			"period_end":   req.PeriodEnd,
		})
		return nil, fmt.Errorf("period_end cannot be before period_start")
	}

	// Validate title is not empty
	if req.Title == "" {
		logger.WarnContext(ctx, "Missing title", nil)
		return nil, fmt.Errorf("title is required")
	}

	plan, err := s.carePlanRepo.Create(ctx, patientID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create care plan", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create care plan: %w", err)
	}

	logger.InfoContext(ctx, "Care plan created successfully", map[string]interface{}{
		"plan_id":    plan.PlanID,
		"patient_id": plan.PatientID,
		"title":      plan.Title,
		"created_by": createdBy,
	})

	return plan, nil
}

// GetCarePlan retrieves a care plan by ID with access control
func (s *CarePlanService) GetCarePlan(ctx context.Context, patientID, planID, requestorID string) (*models.CarePlan, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"plan_id":      planID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized care plan access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"plan_id":      planID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this care plan")
	}

	return s.carePlanRepo.GetByID(ctx, patientID, planID)
}

// ListCarePlans lists care plans with filters and access control
func (s *CarePlanService) ListCarePlans(ctx context.Context, filter *models.CarePlanFilter, requestorID string) ([]*models.CarePlan, error) {
	// If filtering by patient ID, check access
	if filter.PatientID != nil {
		hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, *filter.PatientID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
				"patient_id":   *filter.PatientID,
				"requestor_id": requestorID,
			})
			return nil, fmt.Errorf("failed to check access: %w", err)
		}

		if !hasAccess {
			logger.WarnContext(ctx, "Unauthorized care plans list attempt", map[string]interface{}{
				"patient_id":   *filter.PatientID,
				"requestor_id": requestorID,
			})
			return nil, fmt.Errorf("access denied: you do not have permission to view care plans for this patient")
		}
	}

	return s.carePlanRepo.List(ctx, filter)
}

// UpdateCarePlan updates a care plan with validation and access control
func (s *CarePlanService) UpdateCarePlan(ctx context.Context, patientID, planID string, req *models.CarePlanUpdateRequest, updatedBy string) (*models.CarePlan, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"plan_id":    planID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized care plan update attempt", map[string]interface{}{
			"patient_id": patientID,
			"plan_id":    planID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this care plan")
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"draft":     true,
			"active":    true,
			"on-hold":   true,
			"revoked":   true,
			"completed": true,
		}
		if !validStatuses[*req.Status] {
			logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
				"status": *req.Status,
			})
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Validate intent if provided
	if req.Intent != nil {
		validIntents := map[string]bool{
			"proposal": true,
			"plan":     true,
			"order":    true,
		}
		if !validIntents[*req.Intent] {
			logger.WarnContext(ctx, "Invalid intent", map[string]interface{}{
				"intent": *req.Intent,
			})
			return nil, fmt.Errorf("invalid intent: %s", *req.Intent)
		}
	}

	// Validate period_end is after period_start if both provided
	if req.PeriodStart != nil && req.PeriodEnd != nil && req.PeriodEnd.Before(*req.PeriodStart) {
		logger.WarnContext(ctx, "Invalid period", map[string]interface{}{
			"period_start": *req.PeriodStart,
			"period_end":   *req.PeriodEnd,
		})
		return nil, fmt.Errorf("period_end cannot be before period_start")
	}

	// Validate title is not empty if provided
	if req.Title != nil && *req.Title == "" {
		logger.WarnContext(ctx, "Empty title", nil)
		return nil, fmt.Errorf("title cannot be empty")
	}

	plan, err := s.carePlanRepo.Update(ctx, patientID, planID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update care plan", err, map[string]interface{}{
			"patient_id": patientID,
			"plan_id":    planID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to update care plan: %w", err)
	}

	logger.InfoContext(ctx, "Care plan updated successfully", map[string]interface{}{
		"plan_id":    plan.PlanID,
		"patient_id": plan.PatientID,
		"updated_by": updatedBy,
	})

	return plan, nil
}

// DeleteCarePlan deletes a care plan with access control
func (s *CarePlanService) DeleteCarePlan(ctx context.Context, patientID, planID, deletedBy string) error {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"plan_id":    planID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized care plan deletion attempt", map[string]interface{}{
			"patient_id": patientID,
			"plan_id":    planID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this care plan")
	}

	// Verify the care plan exists before deletion
	_, err = s.carePlanRepo.GetByID(ctx, patientID, planID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get care plan for deletion", err, map[string]interface{}{
			"patient_id": patientID,
			"plan_id":    planID,
		})
		return fmt.Errorf("care plan not found: %w", err)
	}

	err = s.carePlanRepo.Delete(ctx, patientID, planID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete care plan", err, map[string]interface{}{
			"patient_id": patientID,
			"plan_id":    planID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to delete care plan: %w", err)
	}

	logger.InfoContext(ctx, "Care plan deleted successfully", map[string]interface{}{
		"patient_id": patientID,
		"plan_id":    planID,
		"deleted_by": deletedBy,
	})

	return nil
}

// GetActiveCarePlans retrieves active care plans for a patient with access control
func (s *CarePlanService) GetActiveCarePlans(ctx context.Context, patientID, requestorID string) ([]*models.CarePlan, error) {
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
		logger.WarnContext(ctx, "Unauthorized active care plans access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view care plans for this patient")
	}

	return s.carePlanRepo.GetActiveCarePlans(ctx, patientID)
}
