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

// AllergyIntoleranceRepository handles allergy and intolerance data operations
type AllergyIntoleranceRepository struct {
	client *spanner.Client
}

// NewAllergyIntoleranceRepository creates a new allergy intolerance repository
func NewAllergyIntoleranceRepository(spannerRepo *SpannerRepository) *AllergyIntoleranceRepository {
	return &AllergyIntoleranceRepository{
		client: spannerRepo.client,
	}
}

// CreateAllergy creates a new allergy intolerance record
func (r *AllergyIntoleranceRepository) CreateAllergy(ctx context.Context, req *models.AllergyIntoleranceCreateRequest, createdBy string) (*models.AllergyIntolerance, error) {
	allergyID := uuid.New().String()
	now := time.Now()

	// Marshal reactions to JSONB
	var reactionsJSON []byte
	var err error
	if len(req.Reactions) > 0 {
		reactionsJSON, err = json.Marshal(req.Reactions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal reactions: %w", err)
		}
	}

	mutation := spanner.InsertMap("allergy_intolerances", map[string]interface{}{
		"allergy_id":           allergyID,
		"patient_id":           req.PatientID,
		"clinical_status":      req.ClinicalStatus,
		"verification_status":  req.VerificationStatus,
		"type":                 req.Type,
		"category":             req.Category,
		"criticality":          req.Criticality,
		"code_system":          req.CodeSystem,
		"code":                 req.Code,
		"display_name":         req.DisplayName,
		"reactions":            string(reactionsJSON),
		"onset_date":           req.OnsetDate,
		"onset_age":            req.OnsetAge,
		"onset_note":           req.OnsetNote,
		"last_occurrence_date": nil,
		"recorded_date":        now,
		"recorded_by":          createdBy,
		"clinical_notes":       req.ClinicalNotes,
		"patient_comments":     "",
		"created_at":           now,
		"created_by":           createdBy,
		"updated_at":           now,
		"updated_by":           createdBy,
		"deleted":              false,
		"deleted_at":           nil,
	})

	_, err = r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to insert allergy: %w", err)
	}

	allergy, err := r.GetAllergyByID(ctx, allergyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created allergy: %w", err)
	}

	return allergy, nil
}

// GetAllergyByID retrieves an allergy intolerance by ID
func (r *AllergyIntoleranceRepository) GetAllergyByID(ctx context.Context, allergyID string) (*models.AllergyIntolerance, error) {
	stmt := NewStatement(`SELECT
			allergy_id, patient_id,
			clinical_status, verification_status,
			type, category, criticality,
			code_system, code, display_name,
			reactions, max_severity,
			onset_date, onset_age, onset_note,
			last_occurrence_date,
			recorded_date, recorded_by,
			clinical_notes, patient_comments,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM allergy_intolerances
		WHERE allergy_id = @allergyID AND deleted = false`,
		map[string]interface{}{
			"allergyID": allergyID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("allergy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query allergy: %w", err)
	}

	allergy, err := scanAllergy(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan allergy: %w", err)
	}

	return allergy, nil
}

// GetActiveAllergies retrieves all active allergies for a patient
func (r *AllergyIntoleranceRepository) GetActiveAllergies(ctx context.Context, patientID string) ([]*models.AllergyIntolerance, error) {
	stmt := NewStatement(`SELECT
			allergy_id, patient_id,
			clinical_status, verification_status,
			type, category, criticality,
			code_system, code, display_name,
			reactions, max_severity,
			onset_date, onset_age, onset_note,
			last_occurrence_date,
			recorded_date, recorded_by,
			clinical_notes, patient_comments,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM allergy_intolerances
		WHERE patient_id = @patientID
			AND deleted = false
			AND clinical_status = 'active'
		ORDER BY criticality DESC, recorded_date DESC`,
		map[string]interface{}{
			"patientID": patientID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var allergies []*models.AllergyIntolerance
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate allergies: %w", err)
		}

		allergy, err := scanAllergy(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allergy: %w", err)
		}

		allergies = append(allergies, allergy)
	}

	return allergies, nil
}

// GetMedicationAllergies retrieves all medication allergies for a patient
func (r *AllergyIntoleranceRepository) GetMedicationAllergies(ctx context.Context, patientID string) ([]*models.AllergyIntolerance, error) {
	stmt := NewStatement(`SELECT
			allergy_id, patient_id,
			clinical_status, verification_status,
			type, category, criticality,
			code_system, code, display_name,
			reactions, max_severity,
			onset_date, onset_age, onset_note,
			last_occurrence_date,
			recorded_date, recorded_by,
			clinical_notes, patient_comments,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM allergy_intolerances
		WHERE patient_id = @patientID
			AND deleted = false
			AND category = 'medication'
			AND clinical_status = 'active'
		ORDER BY criticality DESC, recorded_date DESC`,
		map[string]interface{}{
			"patientID": patientID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var allergies []*models.AllergyIntolerance
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate allergies: %w", err)
		}

		allergy, err := scanAllergy(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allergy: %w", err)
		}

		allergies = append(allergies, allergy)
	}

	return allergies, nil
}

// GetAllergiesByPatient retrieves all allergies for a patient
func (r *AllergyIntoleranceRepository) GetAllergiesByPatient(ctx context.Context, patientID string) ([]*models.AllergyIntolerance, error) {
	stmt := NewStatement(`SELECT
			allergy_id, patient_id,
			clinical_status, verification_status,
			type, category, criticality,
			code_system, code, display_name,
			reactions, max_severity,
			onset_date, onset_age, onset_note,
			last_occurrence_date,
			recorded_date, recorded_by,
			clinical_notes, patient_comments,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM allergy_intolerances
		WHERE patient_id = @patientID AND deleted = false
		ORDER BY recorded_date DESC`,
		map[string]interface{}{
			"patientID": patientID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var allergies []*models.AllergyIntolerance
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate allergies: %w", err)
		}

		allergy, err := scanAllergy(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allergy: %w", err)
		}

		allergies = append(allergies, allergy)
	}

	return allergies, nil
}

// UpdateAllergy updates an allergy intolerance record
func (r *AllergyIntoleranceRepository) UpdateAllergy(ctx context.Context, allergyID string, req *models.AllergyIntoleranceUpdateRequest, updatedBy string) (*models.AllergyIntolerance, error) {
	now := time.Now()
	updates := make(map[string]interface{})
	updates["allergy_id"] = allergyID
	updates["updated_at"] = now
	updates["updated_by"] = updatedBy

	if req.ClinicalStatus != nil {
		updates["clinical_status"] = *req.ClinicalStatus
	}

	if req.VerificationStatus != nil {
		updates["verification_status"] = *req.VerificationStatus
	}

	if req.Criticality != nil {
		updates["criticality"] = *req.Criticality
	}

	if req.Reactions != nil {
		reactionsJSON, err := json.Marshal(*req.Reactions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal reactions: %w", err)
		}
		updates["reactions"] = string(reactionsJSON)
	}

	if req.LastOccurrenceDate != nil {
		updates["last_occurrence_date"] = *req.LastOccurrenceDate
	}

	if req.ClinicalNotes != nil {
		updates["clinical_notes"] = *req.ClinicalNotes
	}

	if req.PatientComments != nil {
		updates["patient_comments"] = *req.PatientComments
	}

	mutation := spanner.UpdateMap("allergy_intolerances", updates)

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update allergy: %w", err)
	}

	allergy, err := r.GetAllergyByID(ctx, allergyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated allergy: %w", err)
	}

	return allergy, nil
}

// DeleteAllergy performs a soft delete on an allergy intolerance record
func (r *AllergyIntoleranceRepository) DeleteAllergy(ctx context.Context, allergyID, deletedBy string) error {
	now := time.Now()

	mutation := spanner.UpdateMap("allergy_intolerances", map[string]interface{}{
		"allergy_id": allergyID,
		"deleted":    true,
		"deleted_at": now,
		"updated_at": now,
		"updated_by": deletedBy,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete allergy: %w", err)
	}

	return nil
}

// scanAllergy scans a Spanner row into an AllergyIntolerance model
func scanAllergy(row *spanner.Row) (*models.AllergyIntolerance, error) {
	var allergy models.AllergyIntolerance
	var reactionsStr *string

	err := row.Columns(
		&allergy.AllergyID,
		&allergy.PatientID,
		&allergy.ClinicalStatus,
		&allergy.VerificationStatus,
		&allergy.Type,
		&allergy.Category,
		&allergy.Criticality,
		&allergy.CodeSystem,
		&allergy.Code,
		&allergy.DisplayName,
		&reactionsStr,
		&allergy.MaxSeverity,
		&allergy.OnsetDate,
		&allergy.OnsetAge,
		&allergy.OnsetNote,
		&allergy.LastOccurrenceDate,
		&allergy.RecordedDate,
		&allergy.RecordedBy,
		&allergy.ClinicalNotes,
		&allergy.PatientComments,
		&allergy.CreatedAt,
		&allergy.CreatedBy,
		&allergy.UpdatedAt,
		&allergy.UpdatedBy,
		&allergy.Deleted,
		&allergy.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	if reactionsStr != nil && *reactionsStr != "" {
		allergy.Reactions = json.RawMessage(*reactionsStr)
	}

	return &allergy, nil
}
