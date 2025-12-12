package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// CoverageService handles business logic for insurance coverage operations
type CoverageService struct {
	coverageRepo *repository.CoverageRepository
	patientRepo  *repository.PatientRepository
}

// NewCoverageService creates a new coverage service
func NewCoverageService(
	coverageRepo *repository.CoverageRepository,
	patientRepo *repository.PatientRepository,
) *CoverageService {
	return &CoverageService{
		coverageRepo: coverageRepo,
		patientRepo:  patientRepo,
	}
}

// CreateCoverage creates a new coverage with access control
func (s *CoverageService) CreateCoverage(ctx context.Context, req *models.PatientCoverageCreateRequest, createdBy string) (*models.PatientCoverage, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		logger.WarnContext(ctx, "Invalid coverage create request", map[string]interface{}{
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
		logger.WarnContext(ctx, "Unauthorized coverage creation attempt", map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to create coverage for this patient")
	}

	// Create coverage
	coverage, err := s.coverageRepo.CreateCoverage(ctx, req, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create coverage", err, map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create coverage: %w", err)
	}

	logger.InfoContext(ctx, "Coverage created successfully", map[string]interface{}{
		"coverage_id":    coverage.CoverageID,
		"patient_id":     coverage.PatientID,
		"insurance_type": coverage.InsuranceType,
		"created_by":     createdBy,
	})

	return coverage, nil
}

// GetCoverage retrieves a coverage by ID with access control
func (s *CoverageService) GetCoverage(ctx context.Context, coverageID, requestorID string) (*models.PatientCoverage, error) {
	// Get coverage first to check patient ID
	coverage, err := s.coverageRepo.GetCoverageByID(ctx, coverageID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get coverage", err, map[string]interface{}{
			"coverage_id": coverageID,
		})
		return nil, fmt.Errorf("failed to get coverage: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, coverage.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"coverage_id":  coverageID,
			"patient_id":   coverage.PatientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized coverage access attempt", map[string]interface{}{
			"coverage_id":  coverageID,
			"patient_id":   coverage.PatientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this coverage")
	}

	return coverage, nil
}

// GetActiveCoverages retrieves all active coverages for a patient with access control
func (s *CoverageService) GetActiveCoverages(ctx context.Context, patientID, requestorID string) ([]*models.PatientCoverage, error) {
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
		logger.WarnContext(ctx, "Unauthorized active coverages access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view coverages for this patient")
	}

	coverages, err := s.coverageRepo.GetActiveCoverages(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get active coverages", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get active coverages: %w", err)
	}

	return coverages, nil
}

// GetCoveragesByPatient retrieves all coverages for a patient with access control
func (s *CoverageService) GetCoveragesByPatient(ctx context.Context, patientID, requestorID string) ([]*models.PatientCoverage, error) {
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
		logger.WarnContext(ctx, "Unauthorized coverages access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view coverages for this patient")
	}

	coverages, err := s.coverageRepo.GetCoveragesByPatient(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get coverages", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get coverages: %w", err)
	}

	return coverages, nil
}

// GetCoveragesByPatientAndType retrieves coverages filtered by insurance type with access control
func (s *CoverageService) GetCoveragesByPatientAndType(ctx context.Context, patientID, insuranceType, requestorID string) ([]*models.PatientCoverage, error) {
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
		logger.WarnContext(ctx, "Unauthorized coverages by type access attempt", map[string]interface{}{
			"patient_id":     patientID,
			"insurance_type": insuranceType,
			"requestor_id":   requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view coverages for this patient")
	}

	// Validate insurance type
	validInsuranceTypes := []string{
		string(models.InsuranceTypeMedical),
		string(models.InsuranceTypeLongTermCare),
		string(models.InsuranceTypePublicExpense),
	}
	if !isValidValue(insuranceType, validInsuranceTypes) {
		return nil, fmt.Errorf("invalid insurance_type: must be one of [medical, long_term_care, public_expense]")
	}

	coverages, err := s.coverageRepo.GetCoveragesByPatientAndType(ctx, patientID, insuranceType)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get coverages by type", err, map[string]interface{}{
			"patient_id":     patientID,
			"insurance_type": insuranceType,
		})
		return nil, fmt.Errorf("failed to get coverages by type: %w", err)
	}

	return coverages, nil
}

// UpdateCoverage updates a coverage with access control
func (s *CoverageService) UpdateCoverage(ctx context.Context, coverageID string, req *models.PatientCoverageUpdateRequest, updatedBy string) (*models.PatientCoverage, error) {
	// Get coverage first to check patient ID
	coverage, err := s.coverageRepo.GetCoverageByID(ctx, coverageID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get coverage for update", err, map[string]interface{}{
			"coverage_id": coverageID,
		})
		return nil, fmt.Errorf("coverage not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, coverage.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"coverage_id": coverageID,
			"patient_id":  coverage.PatientID,
			"updated_by":  updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized coverage update attempt", map[string]interface{}{
			"coverage_id": coverageID,
			"patient_id":  coverage.PatientID,
			"updated_by":  updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this coverage")
	}

	// Validate update request
	if req.Status != nil {
		validStatuses := []string{
			string(models.CoverageStatusActive),
			string(models.CoverageStatusExpired),
			string(models.CoverageStatusSuspended),
			string(models.CoverageStatusTerminated),
		}
		if !isValidValue(*req.Status, validStatuses) {
			return nil, fmt.Errorf("invalid status: must be one of [active, expired, suspended, terminated]")
		}
	}

	// Update coverage
	updatedCoverage, err := s.coverageRepo.UpdateCoverage(ctx, coverageID, req, updatedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update coverage", err, map[string]interface{}{
			"coverage_id": coverageID,
			"updated_by":  updatedBy,
		})
		return nil, fmt.Errorf("failed to update coverage: %w", err)
	}

	logger.InfoContext(ctx, "Coverage updated successfully", map[string]interface{}{
		"coverage_id": updatedCoverage.CoverageID,
		"patient_id":  updatedCoverage.PatientID,
		"updated_by":  updatedBy,
	})

	return updatedCoverage, nil
}

// DeleteCoverage soft deletes a coverage with access control
func (s *CoverageService) DeleteCoverage(ctx context.Context, coverageID, deletedBy string) error {
	// Get coverage first to check patient ID
	coverage, err := s.coverageRepo.GetCoverageByID(ctx, coverageID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get coverage for deletion", err, map[string]interface{}{
			"coverage_id": coverageID,
		})
		return fmt.Errorf("coverage not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, coverage.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"coverage_id": coverageID,
			"patient_id":  coverage.PatientID,
			"deleted_by":  deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized coverage delete attempt", map[string]interface{}{
			"coverage_id": coverageID,
			"patient_id":  coverage.PatientID,
			"deleted_by":  deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this coverage")
	}

	// Delete coverage
	err = s.coverageRepo.DeleteCoverage(ctx, coverageID, deletedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete coverage", err, map[string]interface{}{
			"coverage_id": coverageID,
			"deleted_by":  deletedBy,
		})
		return fmt.Errorf("failed to delete coverage: %w", err)
	}

	logger.InfoContext(ctx, "Coverage deleted successfully", map[string]interface{}{
		"coverage_id": coverageID,
		"patient_id":  coverage.PatientID,
		"deleted_by":  deletedBy,
	})

	return nil
}

// VerifyCoverage marks a coverage as verified with access control
func (s *CoverageService) VerifyCoverage(ctx context.Context, coverageID, verifiedBy string) error {
	// Get coverage first to check patient ID
	coverage, err := s.coverageRepo.GetCoverageByID(ctx, coverageID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get coverage for verification", err, map[string]interface{}{
			"coverage_id": coverageID,
		})
		return fmt.Errorf("coverage not found: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, verifiedBy, coverage.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"coverage_id": coverageID,
			"patient_id":  coverage.PatientID,
			"verified_by": verifiedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized coverage verification attempt", map[string]interface{}{
			"coverage_id": coverageID,
			"patient_id":  coverage.PatientID,
			"verified_by": verifiedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to verify this coverage")
	}

	// Verify coverage
	err = s.coverageRepo.VerifyCoverage(ctx, coverageID, verifiedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to verify coverage", err, map[string]interface{}{
			"coverage_id": coverageID,
			"verified_by": verifiedBy,
		})
		return fmt.Errorf("failed to verify coverage: %w", err)
	}

	logger.InfoContext(ctx, "Coverage verified successfully", map[string]interface{}{
		"coverage_id": coverageID,
		"patient_id":  coverage.PatientID,
		"verified_by": verifiedBy,
	})

	return nil
}

// validateCreateRequest validates coverage create request
func (s *CoverageService) validateCreateRequest(req *models.PatientCoverageCreateRequest) error {
	if req.PatientID == "" {
		return fmt.Errorf("patient_id is required")
	}

	if req.InsuranceType == "" {
		return fmt.Errorf("insurance_type is required")
	}

	// Validate insurance type
	validInsuranceTypes := []string{
		string(models.InsuranceTypeMedical),
		string(models.InsuranceTypeLongTermCare),
		string(models.InsuranceTypePublicExpense),
	}
	if !isValidValue(req.InsuranceType, validInsuranceTypes) {
		return fmt.Errorf("invalid insurance_type: must be one of [medical, long_term_care, public_expense]")
	}

	if len(req.Details) == 0 {
		return fmt.Errorf("details is required")
	}

	if req.ValidFrom.IsZero() {
		return fmt.Errorf("valid_from is required")
	}

	if req.Status == "" {
		return fmt.Errorf("status is required")
	}

	// Validate status
	validStatuses := []string{
		string(models.CoverageStatusActive),
		string(models.CoverageStatusExpired),
		string(models.CoverageStatusSuspended),
		string(models.CoverageStatusTerminated),
	}
	if !isValidValue(req.Status, validStatuses) {
		return fmt.Errorf("invalid status: must be one of [active, expired, suspended, terminated]")
	}

	if req.Priority < 1 {
		return fmt.Errorf("priority must be at least 1")
	}

	return nil
}
