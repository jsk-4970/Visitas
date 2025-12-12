package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
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

// CreateCarePlan creates a new care plan with validation
func (s *CarePlanService) CreateCarePlan(ctx context.Context, patientID string, req *models.CarePlanCreateRequest) (*models.CarePlan, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
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
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate intent
	validIntents := map[string]bool{
		"proposal": true,
		"plan":     true,
		"order":    true,
	}
	if !validIntents[req.Intent] {
		return nil, fmt.Errorf("invalid intent: %s", req.Intent)
	}

	// Validate period_end is after period_start if provided
	if req.PeriodEnd != nil && req.PeriodEnd.Before(req.PeriodStart) {
		return nil, fmt.Errorf("period_end cannot be before period_start")
	}

	// Validate title is not empty
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	return s.carePlanRepo.Create(ctx, patientID, req)
}

// GetCarePlan retrieves a care plan by ID
func (s *CarePlanService) GetCarePlan(ctx context.Context, patientID, planID string) (*models.CarePlan, error) {
	return s.carePlanRepo.GetByID(ctx, patientID, planID)
}

// ListCarePlans lists care plans with filters
func (s *CarePlanService) ListCarePlans(ctx context.Context, filter *models.CarePlanFilter) ([]*models.CarePlan, error) {
	return s.carePlanRepo.List(ctx, filter)
}

// UpdateCarePlan updates a care plan with validation
func (s *CarePlanService) UpdateCarePlan(ctx context.Context, patientID, planID string, req *models.CarePlanUpdateRequest) (*models.CarePlan, error) {
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
			return nil, fmt.Errorf("invalid intent: %s", *req.Intent)
		}
	}

	// Validate period_end is after period_start if both provided
	if req.PeriodStart != nil && req.PeriodEnd != nil && req.PeriodEnd.Before(*req.PeriodStart) {
		return nil, fmt.Errorf("period_end cannot be before period_start")
	}

	// Validate title is not empty if provided
	if req.Title != nil && *req.Title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	return s.carePlanRepo.Update(ctx, patientID, planID, req)
}

// DeleteCarePlan deletes a care plan
func (s *CarePlanService) DeleteCarePlan(ctx context.Context, patientID, planID string) error {
	// Verify the care plan exists before deletion
	_, err := s.carePlanRepo.GetByID(ctx, patientID, planID)
	if err != nil {
		return fmt.Errorf("care plan not found: %w", err)
	}

	return s.carePlanRepo.Delete(ctx, patientID, planID)
}

// GetActiveCarePlans retrieves active care plans for a patient
func (s *CarePlanService) GetActiveCarePlans(ctx context.Context, patientID string) ([]*models.CarePlan, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	return s.carePlanRepo.GetActiveCarePlans(ctx, patientID)
}
