package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	AuditActionView   AuditAction = "view"
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionDecrypt AuditAction = "decrypt"
)

// AuditLog represents a patient access audit log entry
type AuditLog struct {
	LogID          string          `json:"log_id"`
	EventTime      time.Time       `json:"event_time"`
	ActorID        string          `json:"actor_id"`
	Action         AuditAction     `json:"action"`
	ResourceID     string          `json:"resource_id"`
	PatientID      string          `json:"patient_id"`
	AccessedFields json.RawMessage `json:"accessed_fields,omitempty"`
	Success        bool            `json:"success"`
	ErrorMessage   string          `json:"error_message,omitempty"`
	IPAddress      string          `json:"ip_address,omitempty"`
	UserAgent      string          `json:"user_agent,omitempty"`
}

// AuditRepository handles audit log operations
type AuditRepository struct {
	client *spanner.Client
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(spannerRepo *SpannerRepository) *AuditRepository {
	return &AuditRepository{
		client: spannerRepo.client,
	}
}

// LogAccess creates a new audit log entry
func (r *AuditRepository) LogAccess(ctx context.Context, log *AuditLog) error {
	if log.LogID == "" {
		log.LogID = uuid.New().String()
	}
	if log.EventTime.IsZero() {
		log.EventTime = time.Now()
	}

	var accessedFieldsStr string
	if len(log.AccessedFields) > 0 {
		accessedFieldsStr = string(log.AccessedFields)
	}

	mutation := spanner.InsertMap("audit_patient_access_logs", map[string]interface{}{
		"log_id":          log.LogID,
		"event_time":      log.EventTime,
		"actor_id":        log.ActorID,
		"action":          string(log.Action),
		"resource_id":     log.ResourceID,
		"patient_id":      log.PatientID,
		"accessed_fields": accessedFieldsStr,
		"success":         log.Success,
		"error_message":   log.ErrorMessage,
		"ip_address":      log.IPAddress,
		"user_agent":      log.UserAgent,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	return nil
}

// GetLogsByPatientID retrieves audit logs for a specific patient
func (r *AuditRepository) GetLogsByPatientID(ctx context.Context, patientID string, limit, offset int) ([]*AuditLog, int, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			log_id, event_time, actor_id, action, resource_id, patient_id,
			accessed_fields, success, error_message, ip_address, user_agent
		FROM audit_patient_access_logs
		WHERE patient_id = @patientID
		ORDER BY event_time DESC
		LIMIT @limit OFFSET @offset`,
		Params: map[string]interface{}{
			"patientID": patientID,
			"limit":     limit,
			"offset":    offset,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var logs []*AuditLog
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("failed to iterate audit logs: %w", err)
		}

		log, err := scanAuditLog(row)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		logs = append(logs, log)
	}

	// Get total count
	countStmt := spanner.Statement{
		SQL: `SELECT COUNT(*) as total
		FROM audit_patient_access_logs
		WHERE patient_id = @patientID`,
		Params: map[string]interface{}{
			"patientID": patientID,
		},
	}

	countIter := r.client.Single().Query(ctx, countStmt)
	defer countIter.Stop()

	countRow, err := countIter.Next()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	var total int64
	if err := countRow.Columns(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return logs, int(total), nil
}

// GetLogsByActorID retrieves audit logs for a specific actor (staff member)
func (r *AuditRepository) GetLogsByActorID(ctx context.Context, actorID string, limit, offset int) ([]*AuditLog, int, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			log_id, event_time, actor_id, action, resource_id, patient_id,
			accessed_fields, success, error_message, ip_address, user_agent
		FROM audit_patient_access_logs
		WHERE actor_id = @actorID
		ORDER BY event_time DESC
		LIMIT @limit OFFSET @offset`,
		Params: map[string]interface{}{
			"actorID": actorID,
			"limit":   limit,
			"offset":  offset,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var logs []*AuditLog
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("failed to iterate audit logs: %w", err)
		}

		log, err := scanAuditLog(row)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		logs = append(logs, log)
	}

	// Get total count
	countStmt := spanner.Statement{
		SQL: `SELECT COUNT(*) as total
		FROM audit_patient_access_logs
		WHERE actor_id = @actorID`,
		Params: map[string]interface{}{
			"actorID": actorID,
		},
	}

	countIter := r.client.Single().Query(ctx, countStmt)
	defer countIter.Stop()

	countRow, err := countIter.Next()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	var total int64
	if err := countRow.Columns(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return logs, int(total), nil
}

// GetLogsByTimeRange retrieves audit logs within a time range
func (r *AuditRepository) GetLogsByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*AuditLog, int, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			log_id, event_time, actor_id, action, resource_id, patient_id,
			accessed_fields, success, error_message, ip_address, user_agent
		FROM audit_patient_access_logs
		WHERE event_time >= @startTime AND event_time <= @endTime
		ORDER BY event_time DESC
		LIMIT @limit OFFSET @offset`,
		Params: map[string]interface{}{
			"startTime": startTime,
			"endTime":   endTime,
			"limit":     limit,
			"offset":    offset,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var logs []*AuditLog
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("failed to iterate audit logs: %w", err)
		}

		log, err := scanAuditLog(row)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		logs = append(logs, log)
	}

	// Get total count
	countStmt := spanner.Statement{
		SQL: `SELECT COUNT(*) as total
		FROM audit_patient_access_logs
		WHERE event_time >= @startTime AND event_time <= @endTime`,
		Params: map[string]interface{}{
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	countIter := r.client.Single().Query(ctx, countStmt)
	defer countIter.Stop()

	countRow, err := countIter.Next()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	var total int64
	if err := countRow.Columns(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return logs, int(total), nil
}

// GetFailedAccessLogs retrieves failed access attempts
func (r *AuditRepository) GetFailedAccessLogs(ctx context.Context, limit, offset int) ([]*AuditLog, int, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			log_id, event_time, actor_id, action, resource_id, patient_id,
			accessed_fields, success, error_message, ip_address, user_agent
		FROM audit_patient_access_logs
		WHERE success = false
		ORDER BY event_time DESC
		LIMIT @limit OFFSET @offset`,
		Params: map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var logs []*AuditLog
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("failed to iterate audit logs: %w", err)
		}

		log, err := scanAuditLog(row)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		logs = append(logs, log)
	}

	// Get total count
	countStmt := spanner.Statement{
		SQL: `SELECT COUNT(*) as total
		FROM audit_patient_access_logs
		WHERE success = false`,
	}

	countIter := r.client.Single().Query(ctx, countStmt)
	defer countIter.Stop()

	countRow, err := countIter.Next()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	var total int64
	if err := countRow.Columns(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return logs, int(total), nil
}

// LogMyNumberAccess logs access to My Number identifiers (create, update, decrypt, delete)
// This is a critical security function for HIPAA and 3省2ガイドライン compliance
func (r *AuditRepository) LogMyNumberAccess(ctx context.Context, patientID, identifierID, action, actorID string) error {
	log := &AuditLog{
		LogID:      uuid.New().String(),
		EventTime:  time.Now(),
		ActorID:    actorID,
		Action:     AuditAction(action),
		ResourceID: identifierID,
		PatientID:  patientID,
		Success:    true,
	}

	// Add accessed fields metadata for My Number operations
	accessedFields := map[string]string{
		"resource_type": "patient_identifier",
		"identifier_type": "my_number",
		"operation": action,
	}
	accessedFieldsJSON, err := json.Marshal(accessedFields)
	if err != nil {
		return fmt.Errorf("failed to marshal accessed fields: %w", err)
	}
	log.AccessedFields = accessedFieldsJSON

	return r.LogAccess(ctx, log)
}

// scanAuditLog scans a Spanner row into an AuditLog model
func scanAuditLog(row *spanner.Row) (*AuditLog, error) {
	var log AuditLog
	var actionStr string
	var accessedFieldsStr *string

	err := row.Columns(
		&log.LogID,
		&log.EventTime,
		&log.ActorID,
		&actionStr,
		&log.ResourceID,
		&log.PatientID,
		&accessedFieldsStr,
		&log.Success,
		&log.ErrorMessage,
		&log.IPAddress,
		&log.UserAgent,
	)
	if err != nil {
		return nil, err
	}

	log.Action = AuditAction(actionStr)
	if accessedFieldsStr != nil && *accessedFieldsStr != "" {
		log.AccessedFields = json.RawMessage(*accessedFieldsStr)
	}

	return &log, nil
}
