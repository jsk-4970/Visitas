package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"google.golang.org/api/iterator"
)

// ClinicalObservationRepository handles clinical observation data operations
type ClinicalObservationRepository struct {
	spannerRepo *SpannerRepository
}

// NewClinicalObservationRepository creates a new clinical observation repository
func NewClinicalObservationRepository(spannerRepo *SpannerRepository) *ClinicalObservationRepository {
	return &ClinicalObservationRepository{
		spannerRepo: spannerRepo,
	}
}

// Create creates a new clinical observation
func (r *ClinicalObservationRepository) Create(ctx context.Context, patientID string, req *models.ClinicalObservationCreateRequest, createdBy string) (*models.ClinicalObservation, error) {
	observationID := uuid.New().String()
	now := time.Now()

	observation := &models.ClinicalObservation{
		ObservationID:     observationID,
		PatientID:         patientID,
		Category:          req.Category,
		Code:              req.Code,
		EffectiveDatetime: req.EffectiveDatetime,
		Issued:            now,
		Value:             req.Value,
		CreatedAt:         now,
		CreatedBy:         createdBy,
		UpdatedAt:         now,
	}

	if req.Interpretation != nil {
		observation.Interpretation = spanner.NullString{StringVal: *req.Interpretation, Valid: true}
	}
	if req.PerformerID != nil {
		observation.PerformerID = spanner.NullString{StringVal: *req.PerformerID, Valid: true}
	}
	if req.DeviceID != nil {
		observation.DeviceID = spanner.NullString{StringVal: *req.DeviceID, Valid: true}
	}
	if req.VisitRecordID != nil {
		observation.VisitRecordID = spanner.NullString{StringVal: *req.VisitRecordID, Valid: true}
	}

	// Convert JSONB fields to strings for Spanner
	codeStr := spanner.NullString{StringVal: string(req.Code), Valid: true}
	valueStr := spanner.NullString{StringVal: string(req.Value), Valid: true}

	mutation := spanner.Insert("clinical_observations",
		[]string{
			"observation_id", "patient_id", "category", "code",
			"effective_datetime", "issued", "value", "interpretation",
			"performer_id", "device_id", "visit_record_id",
			"created_at", "created_by", "updated_at",
		},
		[]interface{}{
			observationID, patientID, req.Category, codeStr,
			req.EffectiveDatetime, now, valueStr, observation.Interpretation,
			observation.PerformerID, observation.DeviceID, observation.VisitRecordID,
			now, createdBy, now,
		},
	)

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to create clinical observation: %w", err)
	}

	return observation, nil
}

// GetByID retrieves a clinical observation by ID
func (r *ClinicalObservationRepository) GetByID(ctx context.Context, patientID, observationID string) (*models.ClinicalObservation, error) {
	stmt := NewStatement(`SELECT
			observation_id, patient_id, category, code::text,
			effective_datetime, issued, value::text, interpretation,
			performer_id, device_id, visit_record_id,
			created_at, created_by, updated_at, updated_by
		FROM clinical_observations
		WHERE patient_id = @patient_id AND observation_id = @observation_id`,
		map[string]interface{}{
			"patient_id":     patientID,
			"observation_id": observationID,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("clinical observation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query clinical observation: %w", err)
	}

	return scanClinicalObservation(row)
}

// List retrieves clinical observations with filters
func (r *ClinicalObservationRepository) List(ctx context.Context, filter *models.ClinicalObservationFilter) ([]*models.ClinicalObservation, error) {
	var conditions []string
	params := make(map[string]interface{})

	if filter.PatientID != nil {
		conditions = append(conditions, "patient_id = @patient_id")
		params["patient_id"] = *filter.PatientID
	}

	if filter.Category != nil {
		conditions = append(conditions, "category = @category")
		params["category"] = *filter.Category
	}

	if filter.EffectiveDatetimeFrom != nil {
		conditions = append(conditions, "effective_datetime >= @effective_datetime_from")
		params["effective_datetime_from"] = *filter.EffectiveDatetimeFrom
	}

	if filter.EffectiveDatetimeTo != nil {
		conditions = append(conditions, "effective_datetime <= @effective_datetime_to")
		params["effective_datetime_to"] = *filter.EffectiveDatetimeTo
	}

	if filter.PerformerID != nil {
		conditions = append(conditions, "performer_id = @performer_id")
		params["performer_id"] = *filter.PerformerID
	}

	if filter.VisitRecordID != nil {
		conditions = append(conditions, "visit_record_id = @visit_record_id")
		params["visit_record_id"] = *filter.VisitRecordID
	}

	if filter.Interpretation != nil {
		conditions = append(conditions, "interpretation = @interpretation")
		params["interpretation"] = *filter.Interpretation
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := 100
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	params["limit"] = limit
	params["offset"] = offset

	stmt := NewStatement(fmt.Sprintf(`SELECT
			observation_id, patient_id, category, code::text,
			effective_datetime, issued, value::text, interpretation,
			performer_id, device_id, visit_record_id,
			created_at, created_by, updated_at, updated_by
		FROM clinical_observations
		%s
		ORDER BY effective_datetime DESC, issued DESC
		LIMIT @limit OFFSET @offset`, whereClause),
		params)

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var observations []*models.ClinicalObservation
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate clinical observations: %w", err)
		}

		observation, err := scanClinicalObservation(row)
		if err != nil {
			return nil, err
		}
		observations = append(observations, observation)
	}

	return observations, nil
}

// Update updates a clinical observation
func (r *ClinicalObservationRepository) Update(ctx context.Context, patientID, observationID string, req *models.ClinicalObservationUpdateRequest, updatedBy string) (*models.ClinicalObservation, error) {
	// First, get the existing observation
	existing, err := r.GetByID(ctx, patientID, observationID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Build update map
	updates := make(map[string]interface{})
	updates["updated_at"] = now
	updates["updated_by"] = spanner.NullString{StringVal: updatedBy, Valid: true}
	existing.UpdatedAt = now
	existing.UpdatedBy = spanner.NullString{StringVal: updatedBy, Valid: true}

	if req.Category != nil {
		updates["category"] = *req.Category
		existing.Category = *req.Category
	}

	if len(req.Code) > 0 {
		updates["code"] = spanner.NullString{StringVal: string(req.Code), Valid: true}
		existing.Code = req.Code
	}

	if req.EffectiveDatetime != nil {
		updates["effective_datetime"] = *req.EffectiveDatetime
		existing.EffectiveDatetime = *req.EffectiveDatetime
	}

	if len(req.Value) > 0 {
		updates["value"] = spanner.NullString{StringVal: string(req.Value), Valid: true}
		existing.Value = req.Value
	}

	if req.Interpretation != nil {
		updates["interpretation"] = spanner.NullString{StringVal: *req.Interpretation, Valid: true}
		existing.Interpretation = spanner.NullString{StringVal: *req.Interpretation, Valid: true}
	}

	if req.PerformerID != nil {
		updates["performer_id"] = spanner.NullString{StringVal: *req.PerformerID, Valid: true}
		existing.PerformerID = spanner.NullString{StringVal: *req.PerformerID, Valid: true}
	}

	if req.DeviceID != nil {
		updates["device_id"] = spanner.NullString{StringVal: *req.DeviceID, Valid: true}
		existing.DeviceID = spanner.NullString{StringVal: *req.DeviceID, Valid: true}
	}

	if req.VisitRecordID != nil {
		updates["visit_record_id"] = spanner.NullString{StringVal: *req.VisitRecordID, Valid: true}
		existing.VisitRecordID = spanner.NullString{StringVal: *req.VisitRecordID, Valid: true}
	}

	if len(updates) == 0 {
		return existing, nil
	}

	// Build column list and values
	columns := []string{"patient_id", "observation_id"}
	values := []interface{}{patientID, observationID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("clinical_observations", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update clinical observation: %w", err)
	}

	return existing, nil
}

// Delete deletes a clinical observation
func (r *ClinicalObservationRepository) Delete(ctx context.Context, patientID, observationID string) error {
	mutation := spanner.Delete("clinical_observations", spanner.Key{observationID})

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete clinical observation: %w", err)
	}

	return nil
}

// GetLatestByCategory retrieves the latest observation for a given category
func (r *ClinicalObservationRepository) GetLatestByCategory(ctx context.Context, patientID, category string) (*models.ClinicalObservation, error) {
	stmt := NewStatement(`SELECT
			observation_id, patient_id, category, code::text,
			effective_datetime, issued, value::text, interpretation,
			performer_id, device_id, visit_record_id,
			created_at, created_by, updated_at, updated_by
		FROM clinical_observations
		WHERE patient_id = @patient_id AND category = @category
		ORDER BY effective_datetime DESC, issued DESC
		LIMIT 1`,
		map[string]interface{}{
			"patient_id": patientID,
			"category":   category,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no observations found for category: %s", category)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest observation: %w", err)
	}

	return scanClinicalObservation(row)
}

// GetTimeSeriesData retrieves time series observation data for trend analysis
func (r *ClinicalObservationRepository) GetTimeSeriesData(ctx context.Context, patientID, category string, from, to time.Time) ([]*models.ClinicalObservation, error) {
	stmt := NewStatement(`SELECT
			observation_id, patient_id, category, code::text,
			effective_datetime, issued, value::text, interpretation,
			performer_id, device_id, visit_record_id,
			created_at, created_by, updated_at, updated_by
		FROM clinical_observations
		WHERE patient_id = @patient_id
		  AND category = @category
		  AND effective_datetime >= @from
		  AND effective_datetime <= @to
		ORDER BY effective_datetime ASC`,
		map[string]interface{}{
			"patient_id": patientID,
			"category":   category,
			"from":       from,
			"to":         to,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var observations []*models.ClinicalObservation
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate time series data: %w", err)
		}

		observation, err := scanClinicalObservation(row)
		if err != nil {
			return nil, err
		}
		observations = append(observations, observation)
	}

	return observations, nil
}

// scanClinicalObservation scans a Spanner row into a ClinicalObservation model
func scanClinicalObservation(row *spanner.Row) (*models.ClinicalObservation, error) {
	var observation models.ClinicalObservation
	var codeStr, valueStr spanner.NullString

	err := row.Columns(
		&observation.ObservationID,
		&observation.PatientID,
		&observation.Category,
		&codeStr,
		&observation.EffectiveDatetime,
		&observation.Issued,
		&valueStr,
		&observation.Interpretation,
		&observation.PerformerID,
		&observation.DeviceID,
		&observation.VisitRecordID,
		&observation.CreatedAt,
		&observation.CreatedBy,
		&observation.UpdatedAt,
		&observation.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan clinical observation: %w", err)
	}

	// Convert JSONB strings back to json.RawMessage
	if codeStr.Valid {
		observation.Code = json.RawMessage(codeStr.StringVal)
	}
	if valueStr.Valid {
		observation.Value = json.RawMessage(valueStr.StringVal)
	}

	return &observation, nil
}
