package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"google.golang.org/api/iterator"
)

// MedicalConditionRepository handles medical condition data operations
type MedicalConditionRepository struct {
	client *spanner.Client
}

// NewMedicalConditionRepository creates a new medical condition repository
func NewMedicalConditionRepository(spannerRepo *SpannerRepository) *MedicalConditionRepository {
	return &MedicalConditionRepository{
		client: spannerRepo.client,
	}
}

// CreateCondition creates a new medical condition record
func (r *MedicalConditionRepository) CreateCondition(ctx context.Context, req *models.MedicalConditionCreateRequest, createdBy string) (*models.MedicalCondition, error) {
	conditionID := uuid.New().String()
	now := time.Now()

	mutation := spanner.InsertMap("medical_conditions", map[string]interface{}{
		"condition_id":        conditionID,
		"patient_id":          req.PatientID,
		"clinical_status":     req.ClinicalStatus,
		"verification_status": req.VerificationStatus,
		"category":            req.Category,
		"severity":            req.Severity,
		"code_system":         req.CodeSystem,
		"code":                req.Code,
		"display_name":        req.DisplayName,
		"body_site":           req.BodySite,
		"onset_date":          req.OnsetDate,
		"onset_age":           req.OnsetAge,
		"onset_note":          req.OnsetNote,
		"abatement_date":      nil,
		"abatement_note":      "",
		"recorded_date":       now,
		"recorded_by":         createdBy,
		"clinical_notes":      req.ClinicalNotes,
		"patient_comments":    "",
		"created_at":          now,
		"created_by":          createdBy,
		"updated_at":          now,
		"updated_by":          createdBy,
		"deleted":             false,
		"deleted_at":          nil,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to insert condition: %w", err)
	}

	condition, err := r.GetConditionByID(ctx, conditionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created condition: %w", err)
	}

	return condition, nil
}

// GetConditionByID retrieves a medical condition by ID
func (r *MedicalConditionRepository) GetConditionByID(ctx context.Context, conditionID string) (*models.MedicalCondition, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			condition_id, patient_id,
			clinical_status, verification_status,
			category, severity,
			code_system, code, display_name,
			body_site,
			onset_date, onset_age, onset_note,
			abatement_date, abatement_note,
			recorded_date, recorded_by,
			clinical_notes, patient_comments,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM medical_conditions
		WHERE condition_id = @conditionID AND deleted = false`,
		Params: map[string]interface{}{
			"conditionID": conditionID,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("condition not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query condition: %w", err)
	}

	condition, err := scanCondition(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan condition: %w", err)
	}

	return condition, nil
}

// GetActiveConditions retrieves all active conditions for a patient
func (r *MedicalConditionRepository) GetActiveConditions(ctx context.Context, patientID string) ([]*models.MedicalCondition, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			condition_id, patient_id,
			clinical_status, verification_status,
			category, severity,
			code_system, code, display_name,
			body_site,
			onset_date, onset_age, onset_note,
			abatement_date, abatement_note,
			recorded_date, recorded_by,
			clinical_notes, patient_comments,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM medical_conditions
		WHERE patient_id = @patientID
			AND deleted = false
			AND clinical_status IN ('active', 'recurrence', 'relapse')
		ORDER BY recorded_date DESC`,
		Params: map[string]interface{}{
			"patientID": patientID,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var conditions []*models.MedicalCondition
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate conditions: %w", err)
		}

		condition, err := scanCondition(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan condition: %w", err)
		}

		conditions = append(conditions, condition)
	}

	return conditions, nil
}

// GetConditionsByPatient retrieves all conditions for a patient
func (r *MedicalConditionRepository) GetConditionsByPatient(ctx context.Context, patientID string) ([]*models.MedicalCondition, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			condition_id, patient_id,
			clinical_status, verification_status,
			category, severity,
			code_system, code, display_name,
			body_site,
			onset_date, onset_age, onset_note,
			abatement_date, abatement_note,
			recorded_date, recorded_by,
			clinical_notes, patient_comments,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM medical_conditions
		WHERE patient_id = @patientID AND deleted = false
		ORDER BY recorded_date DESC`,
		Params: map[string]interface{}{
			"patientID": patientID,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var conditions []*models.MedicalCondition
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate conditions: %w", err)
		}

		condition, err := scanCondition(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan condition: %w", err)
		}

		conditions = append(conditions, condition)
	}

	return conditions, nil
}

// UpdateCondition updates a medical condition record
func (r *MedicalConditionRepository) UpdateCondition(ctx context.Context, conditionID string, req *models.MedicalConditionUpdateRequest, updatedBy string) (*models.MedicalCondition, error) {
	now := time.Now()
	updates := make(map[string]interface{})
	updates["condition_id"] = conditionID
	updates["updated_at"] = now
	updates["updated_by"] = updatedBy

	if req.ClinicalStatus != nil {
		updates["clinical_status"] = *req.ClinicalStatus
	}

	if req.VerificationStatus != nil {
		updates["verification_status"] = *req.VerificationStatus
	}

	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}

	if req.AbatementDate != nil {
		updates["abatement_date"] = *req.AbatementDate
	}

	if req.AbatementNote != nil {
		updates["abatement_note"] = *req.AbatementNote
	}

	if req.ClinicalNotes != nil {
		updates["clinical_notes"] = *req.ClinicalNotes
	}

	if req.PatientComments != nil {
		updates["patient_comments"] = *req.PatientComments
	}

	mutation := spanner.UpdateMap("medical_conditions", updates)

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update condition: %w", err)
	}

	condition, err := r.GetConditionByID(ctx, conditionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated condition: %w", err)
	}

	return condition, nil
}

// DeleteCondition performs a soft delete on a medical condition record
func (r *MedicalConditionRepository) DeleteCondition(ctx context.Context, conditionID, deletedBy string) error {
	now := time.Now()

	mutation := spanner.UpdateMap("medical_conditions", map[string]interface{}{
		"condition_id": conditionID,
		"deleted":      true,
		"deleted_at":   now,
		"updated_at":   now,
		"updated_by":   deletedBy,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete condition: %w", err)
	}

	return nil
}

// scanCondition scans a Spanner row into a MedicalCondition model
func scanCondition(row *spanner.Row) (*models.MedicalCondition, error) {
	var condition models.MedicalCondition

	err := row.Columns(
		&condition.ConditionID,
		&condition.PatientID,
		&condition.ClinicalStatus,
		&condition.VerificationStatus,
		&condition.Category,
		&condition.Severity,
		&condition.CodeSystem,
		&condition.Code,
		&condition.DisplayName,
		&condition.BodySite,
		&condition.OnsetDate,
		&condition.OnsetAge,
		&condition.OnsetNote,
		&condition.AbatementDate,
		&condition.AbatementNote,
		&condition.RecordedDate,
		&condition.RecordedBy,
		&condition.ClinicalNotes,
		&condition.PatientComments,
		&condition.CreatedAt,
		&condition.CreatedBy,
		&condition.UpdatedAt,
		&condition.UpdatedBy,
		&condition.Deleted,
		&condition.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &condition, nil
}
