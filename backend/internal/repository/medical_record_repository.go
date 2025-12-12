package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"google.golang.org/api/iterator"
)

// MedicalRecordRepository handles medical record data operations
type MedicalRecordRepository struct {
	spannerRepo *SpannerRepository
}

// NewMedicalRecordRepository creates a new medical record repository
func NewMedicalRecordRepository(spannerRepo *SpannerRepository) *MedicalRecordRepository {
	return &MedicalRecordRepository{
		spannerRepo: spannerRepo,
	}
}

// Create creates a new medical record
func (r *MedicalRecordRepository) Create(ctx context.Context, patientID string, req *models.MedicalRecordCreateRequest, createdBy string) (*models.MedicalRecord, error) {
	recordID := uuid.New().String()
	now := time.Now()

	record := &models.MedicalRecord{
		RecordID:       recordID,
		PatientID:      patientID,
		VisitStartedAt: req.VisitStartedAt,
		VisitEndedAt:   req.VisitEndedAt,
		VisitType:      req.VisitType,
		PerformedBy:    req.PerformedBy,
		Status:         req.Status,
		ScheduleID:     req.ScheduleID,
		SOAPContent:    req.SOAPContent,
		TemplateID:     req.TemplateID,
		SourceRecordID: req.SourceRecordID,
		SourceType:     req.SourceType,
		AudioFileURL:   req.AudioFileURL,
		Version:        1,
		CreatedAt:      now,
		CreatedBy:      createdBy,
		UpdatedAt:      now,
		Deleted:        false,
	}

	// Convert JSONB field to sql.NullString
	var soapContentStr sql.NullString
	if len(req.SOAPContent) > 0 {
		soapContentStr = sql.NullString{String: string(req.SOAPContent), Valid: true}
	}

	// Convert optional fields to sql.Null types
	var visitEndedAt sql.NullTime
	if req.VisitEndedAt != nil {
		visitEndedAt = sql.NullTime{Time: *req.VisitEndedAt, Valid: true}
	}

	var scheduleID, templateID, sourceRecordID, audioFileURL sql.NullString
	if req.ScheduleID != nil {
		scheduleID = sql.NullString{String: *req.ScheduleID, Valid: true}
	}
	if req.TemplateID != nil {
		templateID = sql.NullString{String: *req.TemplateID, Valid: true}
	}
	if req.SourceRecordID != nil {
		sourceRecordID = sql.NullString{String: *req.SourceRecordID, Valid: true}
	}
	if req.AudioFileURL != nil {
		audioFileURL = sql.NullString{String: *req.AudioFileURL, Valid: true}
	}

	mutation := spanner.Insert("medical_records",
		[]string{
			"record_id", "patient_id",
			"visit_started_at", "visit_ended_at", "visit_type", "performed_by", "status",
			"schedule_id", "soap_content",
			"template_id", "source_record_id", "source_type", "audio_file_url",
			"version",
			"created_at", "created_by", "updated_at", "deleted",
		},
		[]interface{}{
			recordID, patientID,
			req.VisitStartedAt, visitEndedAt, req.VisitType, req.PerformedBy, req.Status,
			scheduleID, soapContentStr,
			templateID, sourceRecordID, req.SourceType, audioFileURL,
			1,
			now, createdBy, now, false,
		},
	)

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to create medical record: %w", err)
	}

	return record, nil
}

// GetByID retrieves a medical record by ID
func (r *MedicalRecordRepository) GetByID(ctx context.Context, patientID, recordID string) (*models.MedicalRecord, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			record_id, patient_id,
			visit_started_at, visit_ended_at, visit_type, performed_by, status,
			schedule_id, soap_content,
			template_id, source_record_id, source_type, audio_file_url,
			soap_completed, has_ai_assistance,
			version,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at, deleted_by
		FROM medical_records
		WHERE patient_id = @patient_id AND record_id = @record_id AND deleted = false`,
		Params: map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
		},
	}

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("medical record not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query medical record: %w", err)
	}

	return scanMedicalRecord(row)
}

// List retrieves medical records with filters
func (r *MedicalRecordRepository) List(ctx context.Context, filter *models.MedicalRecordFilter) ([]*models.MedicalRecord, error) {
	conditions := []string{"deleted = false"}
	params := make(map[string]interface{})

	if filter.PatientID != nil {
		conditions = append(conditions, "patient_id = @patient_id")
		params["patient_id"] = *filter.PatientID
	}

	if filter.PerformedBy != nil {
		conditions = append(conditions, "performed_by = @performed_by")
		params["performed_by"] = *filter.PerformedBy
	}

	if filter.Status != nil {
		conditions = append(conditions, "status = @status")
		params["status"] = *filter.Status
	}

	if filter.VisitType != nil {
		conditions = append(conditions, "visit_type = @visit_type")
		params["visit_type"] = *filter.VisitType
	}

	if filter.ScheduleID != nil {
		conditions = append(conditions, "schedule_id = @schedule_id")
		params["schedule_id"] = *filter.ScheduleID
	}

	if filter.VisitDateFrom != nil {
		conditions = append(conditions, "visit_started_at >= @visit_date_from")
		params["visit_date_from"] = *filter.VisitDateFrom
	}

	if filter.VisitDateTo != nil {
		conditions = append(conditions, "visit_started_at <= @visit_date_to")
		params["visit_date_to"] = *filter.VisitDateTo
	}

	if filter.SOAPCompleted != nil {
		conditions = append(conditions, "soap_completed = @soap_completed")
		params["soap_completed"] = *filter.SOAPCompleted
	}

	if filter.TemplateID != nil {
		conditions = append(conditions, "template_id = @template_id")
		params["template_id"] = *filter.TemplateID
	}

	if filter.SourceType != nil {
		conditions = append(conditions, "source_type = @source_type")
		params["source_type"] = *filter.SourceType
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	limit := 100
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	stmt := spanner.Statement{
		SQL: fmt.Sprintf(`SELECT
			record_id, patient_id,
			visit_started_at, visit_ended_at, visit_type, performed_by, status,
			schedule_id, soap_content,
			template_id, source_record_id, source_type, audio_file_url,
			soap_completed, has_ai_assistance,
			version,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at, deleted_by
		FROM medical_records
		%s
		ORDER BY visit_started_at DESC, created_at DESC
		LIMIT @limit OFFSET @offset`, whereClause),
		Params: params,
	}
	stmt.Params["limit"] = int64(limit)
	stmt.Params["offset"] = int64(offset)

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var records []*models.MedicalRecord
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate medical records: %w", err)
		}

		record, err := scanMedicalRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// Update updates a medical record
func (r *MedicalRecordRepository) Update(ctx context.Context, patientID, recordID string, req *models.MedicalRecordUpdateRequest, updatedBy string) (*models.MedicalRecord, error) {
	// First, get the existing record
	existing, err := r.GetByID(ctx, patientID, recordID)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.VisitEndedAt != nil {
		updates["visit_ended_at"] = sql.NullTime{Time: *req.VisitEndedAt, Valid: true}
		existing.VisitEndedAt = req.VisitEndedAt
	}

	if req.VisitType != nil {
		updates["visit_type"] = *req.VisitType
		existing.VisitType = *req.VisitType
	}

	if req.Status != nil {
		updates["status"] = *req.Status
		existing.Status = *req.Status
	}

	if len(req.SOAPContent) > 0 {
		updates["soap_content"] = sql.NullString{String: string(req.SOAPContent), Valid: true}
		existing.SOAPContent = req.SOAPContent
	}

	if req.ScheduleID != nil {
		updates["schedule_id"] = sql.NullString{String: *req.ScheduleID, Valid: true}
		existing.ScheduleID = req.ScheduleID
	}

	if req.TemplateID != nil {
		updates["template_id"] = sql.NullString{String: *req.TemplateID, Valid: true}
		existing.TemplateID = req.TemplateID
	}

	if req.AudioFileURL != nil {
		updates["audio_file_url"] = sql.NullString{String: *req.AudioFileURL, Valid: true}
		existing.AudioFileURL = req.AudioFileURL
	}

	if len(updates) == 0 {
		return existing, nil
	}

	// Update audit fields
	now := time.Now()
	updates["updated_at"] = now
	updates["updated_by"] = sql.NullString{String: updatedBy, Valid: true}
	existing.UpdatedAt = now
	existing.UpdatedBy = &updatedBy

	// Increment version
	updates["version"] = existing.Version + 1
	existing.Version++

	// Build column list and values
	columns := []string{"record_id"}
	values := []interface{}{recordID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("medical_records", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update medical record: %w", err)
	}

	return existing, nil
}

// UpdateWithVersion updates a medical record with optimistic locking
func (r *MedicalRecordRepository) UpdateWithVersion(ctx context.Context, patientID, recordID string, expectedVersion int, req *models.MedicalRecordUpdateRequest, updatedBy string) (*models.MedicalRecord, error) {
	// First, get the existing record
	existing, err := r.GetByID(ctx, patientID, recordID)
	if err != nil {
		return nil, err
	}

	// Check version for optimistic locking
	if existing.Version != expectedVersion {
		return nil, fmt.Errorf("CONFLICT: Record was modified by another user. Expected version %d but found %d", expectedVersion, existing.Version)
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.VisitEndedAt != nil {
		updates["visit_ended_at"] = sql.NullTime{Time: *req.VisitEndedAt, Valid: true}
		existing.VisitEndedAt = req.VisitEndedAt
	}

	if req.VisitType != nil {
		updates["visit_type"] = *req.VisitType
		existing.VisitType = *req.VisitType
	}

	if req.Status != nil {
		updates["status"] = *req.Status
		existing.Status = *req.Status
	}

	if len(req.SOAPContent) > 0 {
		updates["soap_content"] = sql.NullString{String: string(req.SOAPContent), Valid: true}
		existing.SOAPContent = req.SOAPContent
	}

	if req.ScheduleID != nil {
		updates["schedule_id"] = sql.NullString{String: *req.ScheduleID, Valid: true}
		existing.ScheduleID = req.ScheduleID
	}

	if req.TemplateID != nil {
		updates["template_id"] = sql.NullString{String: *req.TemplateID, Valid: true}
		existing.TemplateID = req.TemplateID
	}

	if req.AudioFileURL != nil {
		updates["audio_file_url"] = sql.NullString{String: *req.AudioFileURL, Valid: true}
		existing.AudioFileURL = req.AudioFileURL
	}

	if len(updates) == 0 {
		return existing, nil
	}

	// Update audit fields
	now := time.Now()
	updates["updated_at"] = now
	updates["updated_by"] = sql.NullString{String: updatedBy, Valid: true}
	existing.UpdatedAt = now
	existing.UpdatedBy = &updatedBy

	// Increment version
	updates["version"] = existing.Version + 1
	existing.Version++

	// Build column list and values
	columns := []string{"record_id"}
	values := []interface{}{recordID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("medical_records", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update medical record: %w", err)
	}

	return existing, nil
}

// Delete soft-deletes a medical record
func (r *MedicalRecordRepository) Delete(ctx context.Context, patientID, recordID string, deletedBy string) error {
	// Check if record exists
	_, err := r.GetByID(ctx, patientID, recordID)
	if err != nil {
		return err
	}

	now := time.Now()
	mutation := spanner.Update("medical_records",
		[]string{"record_id", "deleted", "deleted_at", "deleted_by", "updated_at", "updated_by"},
		[]interface{}{
			recordID,
			true,
			sql.NullTime{Time: now, Valid: true},
			sql.NullString{String: deletedBy, Valid: true},
			now,
			sql.NullString{String: deletedBy, Valid: true},
		},
	)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete medical record: %w", err)
	}

	return nil
}

// GetLatestByPatient retrieves the latest medical records for a patient
func (r *MedicalRecordRepository) GetLatestByPatient(ctx context.Context, patientID string, limit int) ([]*models.MedicalRecord, error) {
	if limit <= 0 {
		limit = 10
	}

	stmt := spanner.Statement{
		SQL: `SELECT
			record_id, patient_id,
			visit_started_at, visit_ended_at, visit_type, performed_by, status,
			schedule_id, soap_content,
			template_id, source_record_id, source_type, audio_file_url,
			soap_completed, has_ai_assistance,
			version,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at, deleted_by
		FROM medical_records
		WHERE patient_id = @patient_id AND deleted = false
		ORDER BY visit_started_at DESC, created_at DESC
		LIMIT @limit`,
		Params: map[string]interface{}{
			"patient_id": patientID,
			"limit":      int64(limit),
		},
	}

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var records []*models.MedicalRecord
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate medical records: %w", err)
		}

		record, err := scanMedicalRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// GetByScheduleID retrieves medical records by schedule ID
func (r *MedicalRecordRepository) GetByScheduleID(ctx context.Context, scheduleID string) ([]*models.MedicalRecord, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			record_id, patient_id,
			visit_started_at, visit_ended_at, visit_type, performed_by, status,
			schedule_id, soap_content,
			template_id, source_record_id, source_type, audio_file_url,
			soap_completed, has_ai_assistance,
			version,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at, deleted_by
		FROM medical_records
		WHERE schedule_id = @schedule_id AND deleted = false
		ORDER BY visit_started_at DESC`,
		Params: map[string]interface{}{
			"schedule_id": scheduleID,
		},
	}

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var records []*models.MedicalRecord
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate medical records: %w", err)
		}

		record, err := scanMedicalRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// GetDraftRecords retrieves draft or in-progress medical records
func (r *MedicalRecordRepository) GetDraftRecords(ctx context.Context, performedBy string) ([]*models.MedicalRecord, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			record_id, patient_id,
			visit_started_at, visit_ended_at, visit_type, performed_by, status,
			schedule_id, soap_content,
			template_id, source_record_id, source_type, audio_file_url,
			soap_completed, has_ai_assistance,
			version,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at, deleted_by
		FROM medical_records
		WHERE performed_by = @performed_by AND status IN ('draft', 'in_progress') AND deleted = false
		ORDER BY visit_started_at DESC`,
		Params: map[string]interface{}{
			"performed_by": performedBy,
		},
	}

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var records []*models.MedicalRecord
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate draft records: %w", err)
		}

		record, err := scanMedicalRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// scanMedicalRecord scans a Spanner row into a MedicalRecord model
func scanMedicalRecord(row *spanner.Row) (*models.MedicalRecord, error) {
	var record models.MedicalRecord

	// Nullable fields
	var visitEndedAt sql.NullTime
	var scheduleID, soapContentStr sql.NullString
	var templateID, sourceRecordID, audioFileURL sql.NullString
	var soapCompleted, hasAIAssistance sql.NullBool
	var updatedBy, deletedBy sql.NullString
	var deletedAt sql.NullTime

	err := row.Columns(
		&record.RecordID,
		&record.PatientID,
		&record.VisitStartedAt,
		&visitEndedAt,
		&record.VisitType,
		&record.PerformedBy,
		&record.Status,
		&scheduleID,
		&soapContentStr,
		&templateID,
		&sourceRecordID,
		&record.SourceType,
		&audioFileURL,
		&soapCompleted,
		&hasAIAssistance,
		&record.Version,
		&record.CreatedAt,
		&record.CreatedBy,
		&record.UpdatedAt,
		&updatedBy,
		&record.Deleted,
		&deletedAt,
		&deletedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan medical record: %w", err)
	}

	// Convert nullable fields
	if visitEndedAt.Valid {
		record.VisitEndedAt = &visitEndedAt.Time
	}
	if scheduleID.Valid {
		record.ScheduleID = &scheduleID.String
	}
	if soapContentStr.Valid {
		record.SOAPContent = json.RawMessage(soapContentStr.String)
	}
	if templateID.Valid {
		record.TemplateID = &templateID.String
	}
	if sourceRecordID.Valid {
		record.SourceRecordID = &sourceRecordID.String
	}
	if audioFileURL.Valid {
		record.AudioFileURL = &audioFileURL.String
	}
	if soapCompleted.Valid {
		record.SOAPCompleted = soapCompleted.Bool
	}
	if hasAIAssistance.Valid {
		record.HasAIAssistance = hasAIAssistance.Bool
	}
	if updatedBy.Valid {
		record.UpdatedBy = &updatedBy.String
	}
	if deletedAt.Valid {
		record.DeletedAt = &deletedAt.Time
	}
	if deletedBy.Valid {
		record.DeletedBy = &deletedBy.String
	}

	return &record, nil
}
