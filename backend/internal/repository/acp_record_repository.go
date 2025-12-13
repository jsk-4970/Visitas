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

// ACPRecordRepository handles ACP record data operations
type ACPRecordRepository struct {
	spannerRepo *SpannerRepository
}

// NewACPRecordRepository creates a new ACP record repository
func NewACPRecordRepository(spannerRepo *SpannerRepository) *ACPRecordRepository {
	return &ACPRecordRepository{
		spannerRepo: spannerRepo,
	}
}

// Create creates a new ACP record
func (r *ACPRecordRepository) Create(ctx context.Context, patientID string, req *models.ACPRecordCreateRequest) (*models.ACPRecord, error) {
	acpID := uuid.New().String()
	now := time.Now()

	// Default version is 1 for new records
	version := 1

	record := &models.ACPRecord{
		ACPID:        acpID,
		PatientID:    patientID,
		RecordedDate: req.RecordedDate,
		Version:      version,
		Status:       req.Status,
		DecisionMaker: req.DecisionMaker,
		Directives:   req.Directives,
		CreatedBy:    req.CreatedBy,
		CreatedAt:    now,
	}

	// Set default data sensitivity if not provided
	if req.DataSensitivity != nil {
		record.DataSensitivity = *req.DataSensitivity
	} else {
		record.DataSensitivity = "highly_confidential"
	}

	// Handle optional fields
	if req.ProxyPersonID != nil {
		record.ProxyPersonID = spanner.NullString{StringVal: *req.ProxyPersonID, Valid: true}
	}
	if req.ValuesNarrative != nil {
		record.ValuesNarrative = spanner.NullString{StringVal: *req.ValuesNarrative, Valid: true}
	}

	// Convert JSONB fields to strings for Spanner
	var directivesStr spanner.NullString
	if len(req.Directives) > 0 {
		directivesStr = spanner.NullString{StringVal: string(req.Directives), Valid: true}
	}

	var legalDocumentsStr spanner.NullString
	if len(req.LegalDocuments) > 0 {
		record.LegalDocuments = req.LegalDocuments
		legalDocumentsStr = spanner.NullString{StringVal: string(req.LegalDocuments), Valid: true}
	}

	var discussionLogStr spanner.NullString
	if len(req.DiscussionLog) > 0 {
		record.DiscussionLog = req.DiscussionLog
		discussionLogStr = spanner.NullString{StringVal: string(req.DiscussionLog), Valid: true}
	}

	var accessRestrictedToStr spanner.NullString
	if len(req.AccessRestrictedTo) > 0 {
		record.AccessRestrictedTo = req.AccessRestrictedTo
		accessRestrictedToStr = spanner.NullString{StringVal: string(req.AccessRestrictedTo), Valid: true}
	}

	mutation := spanner.Insert("acp_records",
		[]string{
			"acp_id", "patient_id", "recorded_date", "version", "status",
			"decision_maker", "proxy_person_id",
			"directives", "values_narrative",
			"legal_documents", "discussion_log",
			"data_sensitivity", "access_restricted_to",
			"created_by", "created_at",
		},
		[]interface{}{
			acpID, patientID, req.RecordedDate, version, req.Status,
			req.DecisionMaker, record.ProxyPersonID,
			directivesStr, record.ValuesNarrative,
			legalDocumentsStr, discussionLogStr,
			record.DataSensitivity, accessRestrictedToStr,
			req.CreatedBy, now,
		},
	)

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to create ACP record: %w", err)
	}

	return record, nil
}

// GetByID retrieves an ACP record by ID
func (r *ACPRecordRepository) GetByID(ctx context.Context, patientID, acpID string) (*models.ACPRecord, error) {
	stmt := NewStatement(`SELECT
			acp_id, patient_id, recorded_date, version, status,
			decision_maker, proxy_person_id,
			directives, values_narrative,
			legal_documents, discussion_log,
			data_sensitivity, access_restricted_to,
			created_by, created_at
		FROM acp_records
		WHERE patient_id = @patient_id AND acp_id = @acp_id`,
		map[string]interface{}{
			"patient_id": patientID,
			"acp_id":     acpID,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("ACP record not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query ACP record: %w", err)
	}

	return scanACPRecord(row)
}

// List retrieves ACP records with filters
func (r *ACPRecordRepository) List(ctx context.Context, filter *models.ACPRecordFilter) ([]*models.ACPRecord, error) {
	var conditions []string
	params := make(map[string]interface{})

	if filter.PatientID != nil {
		conditions = append(conditions, "patient_id = @patient_id")
		params["patient_id"] = *filter.PatientID
	}

	if filter.Status != nil {
		conditions = append(conditions, "status = @status")
		params["status"] = *filter.Status
	}

	if filter.RecordedFrom != nil {
		conditions = append(conditions, "recorded_date >= @recorded_from")
		params["recorded_from"] = *filter.RecordedFrom
	}

	if filter.RecordedTo != nil {
		conditions = append(conditions, "recorded_date <= @recorded_to")
		params["recorded_to"] = *filter.RecordedTo
	}

	if filter.DecisionMaker != nil {
		conditions = append(conditions, "decision_maker = @decision_maker")
		params["decision_maker"] = *filter.DecisionMaker
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
			acp_id, patient_id, recorded_date, version, status,
			decision_maker, proxy_person_id,
			directives, values_narrative,
			legal_documents, discussion_log,
			data_sensitivity, access_restricted_to,
			created_by, created_at
		FROM acp_records
		%s
		ORDER BY recorded_date DESC, version DESC, created_at DESC
		LIMIT @limit OFFSET @offset`, whereClause),
		params)

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var records []*models.ACPRecord
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate ACP records: %w", err)
		}

		record, err := scanACPRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// Update updates an ACP record
func (r *ACPRecordRepository) Update(ctx context.Context, patientID, acpID string, req *models.ACPRecordUpdateRequest) (*models.ACPRecord, error) {
	// First, get the existing record
	existing, err := r.GetByID(ctx, patientID, acpID)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.RecordedDate != nil {
		updates["recorded_date"] = *req.RecordedDate
		existing.RecordedDate = *req.RecordedDate
	}

	if req.Status != nil {
		updates["status"] = *req.Status
		existing.Status = *req.Status
	}

	if req.DecisionMaker != nil {
		updates["decision_maker"] = *req.DecisionMaker
		existing.DecisionMaker = *req.DecisionMaker
	}

	if req.ProxyPersonID != nil {
		updates["proxy_person_id"] = spanner.NullString{StringVal: *req.ProxyPersonID, Valid: true}
		existing.ProxyPersonID = spanner.NullString{StringVal: *req.ProxyPersonID, Valid: true}
	}

	if len(req.Directives) > 0 {
		updates["directives"] = spanner.NullString{StringVal: string(req.Directives), Valid: true}
		existing.Directives = req.Directives
	}

	if req.ValuesNarrative != nil {
		updates["values_narrative"] = spanner.NullString{StringVal: *req.ValuesNarrative, Valid: true}
		existing.ValuesNarrative = spanner.NullString{StringVal: *req.ValuesNarrative, Valid: true}
	}

	if len(req.LegalDocuments) > 0 {
		updates["legal_documents"] = spanner.NullString{StringVal: string(req.LegalDocuments), Valid: true}
		existing.LegalDocuments = req.LegalDocuments
	}

	if len(req.DiscussionLog) > 0 {
		updates["discussion_log"] = spanner.NullString{StringVal: string(req.DiscussionLog), Valid: true}
		existing.DiscussionLog = req.DiscussionLog
	}

	if req.DataSensitivity != nil {
		updates["data_sensitivity"] = *req.DataSensitivity
		existing.DataSensitivity = *req.DataSensitivity
	}

	if len(req.AccessRestrictedTo) > 0 {
		updates["access_restricted_to"] = spanner.NullString{StringVal: string(req.AccessRestrictedTo), Valid: true}
		existing.AccessRestrictedTo = req.AccessRestrictedTo
	}

	if len(updates) == 0 {
		return existing, nil
	}

	// Build column list and values
	columns := []string{"patient_id", "acp_id"}
	values := []interface{}{patientID, acpID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("acp_records", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update ACP record: %w", err)
	}

	return existing, nil
}

// Delete deletes an ACP record
func (r *ACPRecordRepository) Delete(ctx context.Context, patientID, acpID string) error {
	mutation := spanner.Delete("acp_records", spanner.Key{patientID, acpID})

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete ACP record: %w", err)
	}

	return nil
}

// GetLatestACP retrieves the latest active ACP record for a patient
func (r *ACPRecordRepository) GetLatestACP(ctx context.Context, patientID string) (*models.ACPRecord, error) {
	stmt := NewStatement(`SELECT
			acp_id, patient_id, recorded_date, version, status,
			decision_maker, proxy_person_id,
			directives, values_narrative,
			legal_documents, discussion_log,
			data_sensitivity, access_restricted_to,
			created_by, created_at
		FROM acp_records
		WHERE patient_id = @patient_id AND status = 'active'
		ORDER BY version DESC, recorded_date DESC
		LIMIT 1`,
		map[string]interface{}{
			"patient_id": patientID,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no active ACP record found for patient")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest ACP record: %w", err)
	}

	return scanACPRecord(row)
}

// GetACPHistory retrieves the complete history of ACP records for a patient
func (r *ACPRecordRepository) GetACPHistory(ctx context.Context, patientID string) ([]*models.ACPRecord, error) {
	stmt := NewStatement(`SELECT
			acp_id, patient_id, recorded_date, version, status,
			decision_maker, proxy_person_id,
			directives, values_narrative,
			legal_documents, discussion_log,
			data_sensitivity, access_restricted_to,
			created_by, created_at
		FROM acp_records
		WHERE patient_id = @patient_id
		ORDER BY version DESC, recorded_date DESC, created_at DESC`,
		map[string]interface{}{
			"patient_id": patientID,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var records []*models.ACPRecord
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate ACP history: %w", err)
		}

		record, err := scanACPRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// scanACPRecord scans a Spanner row into an ACPRecord model
func scanACPRecord(row *spanner.Row) (*models.ACPRecord, error) {
	var record models.ACPRecord
	var directivesStr, legalDocumentsStr, discussionLogStr, accessRestrictedToStr spanner.NullString

	err := row.Columns(
		&record.ACPID,
		&record.PatientID,
		&record.RecordedDate,
		&record.Version,
		&record.Status,
		&record.DecisionMaker,
		&record.ProxyPersonID,
		&directivesStr,
		&record.ValuesNarrative,
		&legalDocumentsStr,
		&discussionLogStr,
		&record.DataSensitivity,
		&accessRestrictedToStr,
		&record.CreatedBy,
		&record.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan ACP record: %w", err)
	}

	// Convert JSONB strings back to json.RawMessage
	if directivesStr.Valid {
		record.Directives = json.RawMessage(directivesStr.StringVal)
	}
	if legalDocumentsStr.Valid {
		record.LegalDocuments = json.RawMessage(legalDocumentsStr.StringVal)
	}
	if discussionLogStr.Valid {
		record.DiscussionLog = json.RawMessage(discussionLogStr.StringVal)
	}
	if accessRestrictedToStr.Valid {
		record.AccessRestrictedTo = json.RawMessage(accessRestrictedToStr.StringVal)
	}

	return &record, nil
}
