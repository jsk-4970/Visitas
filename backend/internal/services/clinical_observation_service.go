package services

import (
	"context"
	"fmt"
	"time"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// ClinicalObservationService handles business logic for clinical observations
type ClinicalObservationService struct {
	clinicalObservationRepo *repository.ClinicalObservationRepository
	patientRepo             *repository.PatientRepository
}

// NewClinicalObservationService creates a new clinical observation service
func NewClinicalObservationService(
	clinicalObservationRepo *repository.ClinicalObservationRepository,
	patientRepo *repository.PatientRepository,
) *ClinicalObservationService {
	return &ClinicalObservationService{
		clinicalObservationRepo: clinicalObservationRepo,
		patientRepo:             patientRepo,
	}
}

// CreateClinicalObservation creates a new clinical observation with validation and access control
func (s *ClinicalObservationService) CreateClinicalObservation(ctx context.Context, patientID string, req *models.ClinicalObservationCreateRequest, createdBy string) (*models.ClinicalObservation, error) {
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
		logger.WarnContext(ctx, "Unauthorized clinical observation creation attempt", map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to create clinical observations for this patient")
	}

	// Validate category
	validCategories := map[string]bool{
		"vital_signs":           true,
		"adl_assessment":        true,
		"cognitive_assessment":  true,
		"pain_scale":            true,
	}
	if !validCategories[req.Category] {
		logger.WarnContext(ctx, "Invalid category", map[string]interface{}{
			"category": req.Category,
		})
		return nil, fmt.Errorf("invalid category: %s", req.Category)
	}

	// Validate interpretation if provided
	if req.Interpretation != nil {
		validInterpretations := map[string]bool{
			"normal":   true,
			"high":     true,
			"low":      true,
			"critical": true,
		}
		if !validInterpretations[*req.Interpretation] {
			logger.WarnContext(ctx, "Invalid interpretation", map[string]interface{}{
				"interpretation": *req.Interpretation,
			})
			return nil, fmt.Errorf("invalid interpretation: %s", *req.Interpretation)
		}
	}

	// Validate code and value are valid JSON
	if len(req.Code) == 0 {
		logger.WarnContext(ctx, "Missing code", nil)
		return nil, fmt.Errorf("code is required")
	}
	if len(req.Value) == 0 {
		logger.WarnContext(ctx, "Missing value", nil)
		return nil, fmt.Errorf("value is required")
	}

	observation, err := s.clinicalObservationRepo.Create(ctx, patientID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create clinical observation", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create clinical observation: %w", err)
	}

	logger.InfoContext(ctx, "Clinical observation created successfully", map[string]interface{}{
		"observation_id": observation.ObservationID,
		"patient_id":     observation.PatientID,
		"category":       observation.Category,
		"created_by":     createdBy,
	})

	return observation, nil
}

// GetClinicalObservation retrieves a clinical observation by ID with access control
func (s *ClinicalObservationService) GetClinicalObservation(ctx context.Context, patientID, observationID, requestorID string) (*models.ClinicalObservation, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
			"requestor_id":   requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized clinical observation access attempt", map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
			"requestor_id":   requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this clinical observation")
	}

	return s.clinicalObservationRepo.GetByID(ctx, patientID, observationID)
}

// ListClinicalObservations lists clinical observations with filters and access control
func (s *ClinicalObservationService) ListClinicalObservations(ctx context.Context, filter *models.ClinicalObservationFilter, requestorID string) ([]*models.ClinicalObservation, error) {
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
			logger.WarnContext(ctx, "Unauthorized clinical observations list attempt", map[string]interface{}{
				"patient_id":   *filter.PatientID,
				"requestor_id": requestorID,
			})
			return nil, fmt.Errorf("access denied: you do not have permission to view clinical observations for this patient")
		}
	}

	return s.clinicalObservationRepo.List(ctx, filter)
}

// UpdateClinicalObservation updates a clinical observation with validation and access control
func (s *ClinicalObservationService) UpdateClinicalObservation(ctx context.Context, patientID, observationID string, req *models.ClinicalObservationUpdateRequest, updatedBy string) (*models.ClinicalObservation, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
			"updated_by":     updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized clinical observation update attempt", map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
			"updated_by":     updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this clinical observation")
	}

	// Validate category if provided
	if req.Category != nil {
		validCategories := map[string]bool{
			"vital_signs":           true,
			"adl_assessment":        true,
			"cognitive_assessment":  true,
			"pain_scale":            true,
		}
		if !validCategories[*req.Category] {
			logger.WarnContext(ctx, "Invalid category", map[string]interface{}{
				"category": *req.Category,
			})
			return nil, fmt.Errorf("invalid category: %s", *req.Category)
		}
	}

	// Validate interpretation if provided
	if req.Interpretation != nil {
		validInterpretations := map[string]bool{
			"normal":   true,
			"high":     true,
			"low":      true,
			"critical": true,
		}
		if !validInterpretations[*req.Interpretation] {
			logger.WarnContext(ctx, "Invalid interpretation", map[string]interface{}{
				"interpretation": *req.Interpretation,
			})
			return nil, fmt.Errorf("invalid interpretation: %s", *req.Interpretation)
		}
	}

	observation, err := s.clinicalObservationRepo.Update(ctx, patientID, observationID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update clinical observation", err, map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
			"updated_by":     updatedBy,
		})
		return nil, fmt.Errorf("failed to update clinical observation: %w", err)
	}

	logger.InfoContext(ctx, "Clinical observation updated successfully", map[string]interface{}{
		"observation_id": observation.ObservationID,
		"patient_id":     observation.PatientID,
		"updated_by":     updatedBy,
	})

	return observation, nil
}

// DeleteClinicalObservation deletes a clinical observation with access control
func (s *ClinicalObservationService) DeleteClinicalObservation(ctx context.Context, patientID, observationID, deletedBy string) error {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
			"deleted_by":     deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized clinical observation deletion attempt", map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
			"deleted_by":     deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this clinical observation")
	}

	// Verify the observation exists before deletion
	_, err = s.clinicalObservationRepo.GetByID(ctx, patientID, observationID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get clinical observation for deletion", err, map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
		})
		return fmt.Errorf("clinical observation not found: %w", err)
	}

	err = s.clinicalObservationRepo.Delete(ctx, patientID, observationID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete clinical observation", err, map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
			"deleted_by":     deletedBy,
		})
		return fmt.Errorf("failed to delete clinical observation: %w", err)
	}

	logger.InfoContext(ctx, "Clinical observation deleted successfully", map[string]interface{}{
		"patient_id":     patientID,
		"observation_id": observationID,
		"deleted_by":     deletedBy,
	})

	return nil
}

// GetLatestObservationByCategory retrieves the latest observation for a given category with access control
func (s *ClinicalObservationService) GetLatestObservationByCategory(ctx context.Context, patientID, category, requestorID string) (*models.ClinicalObservation, error) {
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
		logger.WarnContext(ctx, "Unauthorized latest observation access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view observations for this patient")
	}

	validCategories := map[string]bool{
		"vital_signs":           true,
		"adl_assessment":        true,
		"cognitive_assessment":  true,
		"pain_scale":            true,
	}
	if !validCategories[category] {
		logger.WarnContext(ctx, "Invalid category", map[string]interface{}{
			"category": category,
		})
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	return s.clinicalObservationRepo.GetLatestByCategory(ctx, patientID, category)
}

// GetTimeSeriesData retrieves time series observation data for trend analysis with access control
func (s *ClinicalObservationService) GetTimeSeriesData(ctx context.Context, patientID, category string, from, to time.Time, requestorID string) ([]*models.ClinicalObservation, error) {
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
		logger.WarnContext(ctx, "Unauthorized time series data access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view observation data for this patient")
	}

	if to.Before(from) {
		logger.WarnContext(ctx, "Invalid time range", map[string]interface{}{
			"from": from,
			"to":   to,
		})
		return nil, fmt.Errorf("to date cannot be before from date")
	}

	validCategories := map[string]bool{
		"vital_signs":           true,
		"adl_assessment":        true,
		"cognitive_assessment":  true,
		"pain_scale":            true,
	}
	if !validCategories[category] {
		logger.WarnContext(ctx, "Invalid category", map[string]interface{}{
			"category": category,
		})
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	return s.clinicalObservationRepo.GetTimeSeriesData(ctx, patientID, category, from, to)
}
