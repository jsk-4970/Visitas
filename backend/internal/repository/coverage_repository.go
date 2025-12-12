package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"google.golang.org/api/iterator"
)

// CoverageRepository handles insurance coverage data operations
type CoverageRepository struct {
	client *spanner.Client
}

// NewCoverageRepository creates a new coverage repository
func NewCoverageRepository(spannerRepo *SpannerRepository) *CoverageRepository {
	return &CoverageRepository{
		client: spannerRepo.client,
	}
}

// CreateCoverage creates a new coverage record
func (r *CoverageRepository) CreateCoverage(ctx context.Context, req *models.PatientCoverageCreateRequest, createdBy string) (*models.PatientCoverage, error) {
	coverageID := uuid.New().String()
	now := time.Now()

	// Marshal details to JSONB
	detailsJSON, err := json.Marshal(req.Details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal details: %w", err)
	}

	mutation := spanner.InsertMap("patient_coverages", map[string]interface{}{
		"coverage_id":         coverageID,
		"patient_id":          req.PatientID,
		"insurance_type":      req.InsuranceType,
		"details":             string(detailsJSON),
		"valid_from":          req.ValidFrom,
		"valid_to":            req.ValidTo,
		"status":              req.Status,
		"priority":            req.Priority,
		"verification_status": "unverified",
		"verified_at":         nil,
		"verified_by":         "",
		"created_at":          now,
		"created_by":          createdBy,
		"updated_at":          now,
		"updated_by":          createdBy,
		"deleted":             false,
		"deleted_at":          nil,
	})

	_, err = r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to insert coverage: %w", err)
	}

	coverage, err := r.GetCoverageByID(ctx, coverageID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created coverage: %w", err)
	}

	return coverage, nil
}

// GetCoverageByID retrieves a coverage by ID
func (r *CoverageRepository) GetCoverageByID(ctx context.Context, coverageID string) (*models.PatientCoverage, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			coverage_id, patient_id, insurance_type, details,
			care_level_code, copay_rate,
			valid_from, valid_to,
			status, priority,
			verification_status, verified_at, verified_by,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM patient_coverages
		WHERE coverage_id = @coverageID AND deleted = false`,
		Params: map[string]interface{}{
			"coverageID": coverageID,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("coverage not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query coverage: %w", err)
	}

	coverage, err := scanCoverage(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan coverage: %w", err)
	}

	return coverage, nil
}

// GetActiveCoverages retrieves all active coverages for a patient
func (r *CoverageRepository) GetActiveCoverages(ctx context.Context, patientID string) ([]*models.PatientCoverage, error) {
	now := time.Now()

	stmt := spanner.Statement{
		SQL: `SELECT
			coverage_id, patient_id, insurance_type, details,
			care_level_code, copay_rate,
			valid_from, valid_to,
			status, priority,
			verification_status, verified_at, verified_by,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM patient_coverages
		WHERE patient_id = @patientID
			AND deleted = false
			AND status = 'active'
			AND valid_from <= @now
			AND (valid_to IS NULL OR valid_to > @now)
		ORDER BY priority ASC`,
		Params: map[string]interface{}{
			"patientID": patientID,
			"now":       now,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var coverages []*models.PatientCoverage
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate coverages: %w", err)
		}

		coverage, err := scanCoverage(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coverage: %w", err)
		}

		coverages = append(coverages, coverage)
	}

	return coverages, nil
}

// GetCoveragesByPatient retrieves all coverages for a patient (including expired)
func (r *CoverageRepository) GetCoveragesByPatient(ctx context.Context, patientID string) ([]*models.PatientCoverage, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			coverage_id, patient_id, insurance_type, details,
			care_level_code, copay_rate,
			valid_from, valid_to,
			status, priority,
			verification_status, verified_at, verified_by,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM patient_coverages
		WHERE patient_id = @patientID AND deleted = false
		ORDER BY priority ASC, created_at DESC`,
		Params: map[string]interface{}{
			"patientID": patientID,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var coverages []*models.PatientCoverage
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate coverages: %w", err)
		}

		coverage, err := scanCoverage(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coverage: %w", err)
		}

		coverages = append(coverages, coverage)
	}

	return coverages, nil
}

// UpdateCoverage updates a coverage record
func (r *CoverageRepository) UpdateCoverage(ctx context.Context, coverageID string, req *models.PatientCoverageUpdateRequest, updatedBy string) (*models.PatientCoverage, error) {
	now := time.Now()
	updates := make(map[string]interface{})
	updates["coverage_id"] = coverageID
	updates["updated_at"] = now
	updates["updated_by"] = updatedBy

	if req.Details != nil {
		detailsJSON, err := json.Marshal(*req.Details)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal details: %w", err)
		}
		updates["details"] = string(detailsJSON)
	}

	if req.ValidTo != nil {
		updates["valid_to"] = *req.ValidTo
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	if req.VerificationStatus != nil {
		updates["verification_status"] = *req.VerificationStatus
		if *req.VerificationStatus == "verified" {
			updates["verified_at"] = now
			updates["verified_by"] = updatedBy
		}
	}

	mutation := spanner.UpdateMap("patient_coverages", updates)

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update coverage: %w", err)
	}

	coverage, err := r.GetCoverageByID(ctx, coverageID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated coverage: %w", err)
	}

	return coverage, nil
}

// DeleteCoverage performs a soft delete on a coverage record
func (r *CoverageRepository) DeleteCoverage(ctx context.Context, coverageID, deletedBy string) error {
	now := time.Now()

	mutation := spanner.UpdateMap("patient_coverages", map[string]interface{}{
		"coverage_id": coverageID,
		"deleted":     true,
		"deleted_at":  now,
		"updated_at":  now,
		"updated_by":  deletedBy,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete coverage: %w", err)
	}

	return nil
}

// scanCoverage scans a Spanner row into a PatientCoverage model
func scanCoverage(row *spanner.Row) (*models.PatientCoverage, error) {
	var coverage models.PatientCoverage
	var detailsStr string

	err := row.Columns(
		&coverage.CoverageID,
		&coverage.PatientID,
		&coverage.InsuranceType,
		&detailsStr,
		&coverage.CareLevelCode,
		&coverage.CopayRate,
		&coverage.ValidFrom,
		&coverage.ValidTo,
		&coverage.Status,
		&coverage.Priority,
		&coverage.VerificationStatus,
		&coverage.VerifiedAt,
		&coverage.VerifiedBy,
		&coverage.CreatedAt,
		&coverage.CreatedBy,
		&coverage.UpdatedAt,
		&coverage.UpdatedBy,
		&coverage.Deleted,
		&coverage.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	coverage.Details = json.RawMessage(detailsStr)

	return &coverage, nil
}
