package services

import (
	"context"
	"fmt"
	"time"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
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

// CreateClinicalObservation creates a new clinical observation with validation
func (s *ClinicalObservationService) CreateClinicalObservation(ctx context.Context, patientID string, req *models.ClinicalObservationCreateRequest) (*models.ClinicalObservation, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	// Validate category
	validCategories := map[string]bool{
		"vital_signs":           true,
		"adl_assessment":        true,
		"cognitive_assessment":  true,
		"pain_scale":            true,
	}
	if !validCategories[req.Category] {
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
			return nil, fmt.Errorf("invalid interpretation: %s", *req.Interpretation)
		}
	}

	// Validate code and value are valid JSON
	if len(req.Code) == 0 {
		return nil, fmt.Errorf("code is required")
	}
	if len(req.Value) == 0 {
		return nil, fmt.Errorf("value is required")
	}

	return s.clinicalObservationRepo.Create(ctx, patientID, req)
}

// GetClinicalObservation retrieves a clinical observation by ID
func (s *ClinicalObservationService) GetClinicalObservation(ctx context.Context, patientID, observationID string) (*models.ClinicalObservation, error) {
	return s.clinicalObservationRepo.GetByID(ctx, patientID, observationID)
}

// ListClinicalObservations lists clinical observations with filters
func (s *ClinicalObservationService) ListClinicalObservations(ctx context.Context, filter *models.ClinicalObservationFilter) ([]*models.ClinicalObservation, error) {
	return s.clinicalObservationRepo.List(ctx, filter)
}

// UpdateClinicalObservation updates a clinical observation with validation
func (s *ClinicalObservationService) UpdateClinicalObservation(ctx context.Context, patientID, observationID string, req *models.ClinicalObservationUpdateRequest) (*models.ClinicalObservation, error) {
	// Validate category if provided
	if req.Category != nil {
		validCategories := map[string]bool{
			"vital_signs":           true,
			"adl_assessment":        true,
			"cognitive_assessment":  true,
			"pain_scale":            true,
		}
		if !validCategories[*req.Category] {
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
			return nil, fmt.Errorf("invalid interpretation: %s", *req.Interpretation)
		}
	}

	return s.clinicalObservationRepo.Update(ctx, patientID, observationID, req)
}

// DeleteClinicalObservation deletes a clinical observation
func (s *ClinicalObservationService) DeleteClinicalObservation(ctx context.Context, patientID, observationID string) error {
	// Verify the observation exists before deletion
	_, err := s.clinicalObservationRepo.GetByID(ctx, patientID, observationID)
	if err != nil {
		return fmt.Errorf("clinical observation not found: %w", err)
	}

	return s.clinicalObservationRepo.Delete(ctx, patientID, observationID)
}

// GetLatestObservationByCategory retrieves the latest observation for a given category
func (s *ClinicalObservationService) GetLatestObservationByCategory(ctx context.Context, patientID, category string) (*models.ClinicalObservation, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	validCategories := map[string]bool{
		"vital_signs":           true,
		"adl_assessment":        true,
		"cognitive_assessment":  true,
		"pain_scale":            true,
	}
	if !validCategories[category] {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	return s.clinicalObservationRepo.GetLatestByCategory(ctx, patientID, category)
}

// GetTimeSeriesData retrieves time series observation data for trend analysis
func (s *ClinicalObservationService) GetTimeSeriesData(ctx context.Context, patientID, category string, from, to time.Time) ([]*models.ClinicalObservation, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	if to.Before(from) {
		return nil, fmt.Errorf("to date cannot be before from date")
	}

	validCategories := map[string]bool{
		"vital_signs":           true,
		"adl_assessment":        true,
		"cognitive_assessment":  true,
		"pain_scale":            true,
	}
	if !validCategories[category] {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	return s.clinicalObservationRepo.GetTimeSeriesData(ctx, patientID, category, from, to)
}
