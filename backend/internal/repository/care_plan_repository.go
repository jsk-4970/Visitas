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

// CarePlanRepository handles care plan data operations
type CarePlanRepository struct {
	spannerRepo *SpannerRepository
}

// NewCarePlanRepository creates a new care plan repository
func NewCarePlanRepository(spannerRepo *SpannerRepository) *CarePlanRepository {
	return &CarePlanRepository{
		spannerRepo: spannerRepo,
	}
}

// Create creates a new care plan
func (r *CarePlanRepository) Create(ctx context.Context, patientID string, req *models.CarePlanCreateRequest) (*models.CarePlan, error) {
	planID := uuid.New().String()
	now := time.Now()

	carePlan := &models.CarePlan{
		PlanID:      planID,
		PatientID:   patientID,
		Status:      req.Status,
		Intent:      req.Intent,
		Title:       req.Title,
		PeriodStart: req.PeriodStart,
		Goals:       req.Goals,
		Activities:  req.Activities,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if req.Description != nil {
		carePlan.Description = sql.NullString{String: *req.Description, Valid: true}
	}
	if req.PeriodEnd != nil {
		carePlan.PeriodEnd = sql.NullTime{Time: *req.PeriodEnd, Valid: true}
	}

	// Convert JSONB fields to strings for Spanner
	var goalsStr, activitiesStr sql.NullString
	if len(req.Goals) > 0 {
		goalsStr = sql.NullString{String: string(req.Goals), Valid: true}
	}
	if len(req.Activities) > 0 {
		activitiesStr = sql.NullString{String: string(req.Activities), Valid: true}
	}

	mutation := spanner.Insert("care_plans",
		[]string{
			"plan_id", "patient_id", "status", "intent",
			"title", "description", "period_start", "period_end",
			"goals", "activities",
			"created_by", "created_at", "updated_at",
		},
		[]interface{}{
			planID, patientID, req.Status, req.Intent,
			req.Title, carePlan.Description, req.PeriodStart, carePlan.PeriodEnd,
			goalsStr, activitiesStr,
			req.CreatedBy, now, now,
		},
	)

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to create care plan: %w", err)
	}

	return carePlan, nil
}

// GetByID retrieves a care plan by ID
func (r *CarePlanRepository) GetByID(ctx context.Context, patientID, planID string) (*models.CarePlan, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			plan_id, patient_id, status, intent,
			title, description, period_start, period_end,
			goals, activities,
			created_by, created_at, updated_at
		FROM care_plans
		WHERE patient_id = @patient_id AND plan_id = @plan_id`,
		Params: map[string]interface{}{
			"patient_id": patientID,
			"plan_id":    planID,
		},
	}

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("care plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query care plan: %w", err)
	}

	return scanCarePlan(row)
}

// List retrieves care plans with filters
func (r *CarePlanRepository) List(ctx context.Context, filter *models.CarePlanFilter) ([]*models.CarePlan, error) {
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

	if filter.Intent != nil {
		conditions = append(conditions, "intent = @intent")
		params["intent"] = *filter.Intent
	}

	if filter.PeriodStartFrom != nil {
		conditions = append(conditions, "period_start >= @period_start_from")
		params["period_start_from"] = *filter.PeriodStartFrom
	}

	if filter.PeriodStartTo != nil {
		conditions = append(conditions, "period_start <= @period_start_to")
		params["period_start_to"] = *filter.PeriodStartTo
	}

	if filter.CreatedBy != nil {
		conditions = append(conditions, "created_by = @created_by")
		params["created_by"] = *filter.CreatedBy
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

	stmt := spanner.Statement{
		SQL: fmt.Sprintf(`SELECT
			plan_id, patient_id, status, intent,
			title, description, period_start, period_end,
			goals, activities,
			created_by, created_at, updated_at
		FROM care_plans
		%s
		ORDER BY period_start DESC, created_at DESC
		LIMIT @limit OFFSET @offset`, whereClause),
		Params: params,
	}
	stmt.Params["limit"] = limit
	stmt.Params["offset"] = offset

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var carePlans []*models.CarePlan
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate care plans: %w", err)
		}

		carePlan, err := scanCarePlan(row)
		if err != nil {
			return nil, err
		}
		carePlans = append(carePlans, carePlan)
	}

	return carePlans, nil
}

// Update updates a care plan
func (r *CarePlanRepository) Update(ctx context.Context, patientID, planID string, req *models.CarePlanUpdateRequest) (*models.CarePlan, error) {
	// First, get the existing care plan
	existing, err := r.GetByID(ctx, patientID, planID)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
		existing.Status = *req.Status
	}

	if req.Intent != nil {
		updates["intent"] = *req.Intent
		existing.Intent = *req.Intent
	}

	if req.Title != nil {
		updates["title"] = *req.Title
		existing.Title = *req.Title
	}

	if req.Description != nil {
		updates["description"] = sql.NullString{String: *req.Description, Valid: true}
		existing.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	if req.PeriodStart != nil {
		updates["period_start"] = *req.PeriodStart
		existing.PeriodStart = *req.PeriodStart
	}

	if req.PeriodEnd != nil {
		updates["period_end"] = sql.NullTime{Time: *req.PeriodEnd, Valid: true}
		existing.PeriodEnd = sql.NullTime{Time: *req.PeriodEnd, Valid: true}
	}

	if len(req.Goals) > 0 {
		updates["goals"] = sql.NullString{String: string(req.Goals), Valid: true}
		existing.Goals = req.Goals
	}

	if len(req.Activities) > 0 {
		updates["activities"] = sql.NullString{String: string(req.Activities), Valid: true}
		existing.Activities = req.Activities
	}

	if len(updates) == 0 {
		return existing, nil
	}

	updates["updated_at"] = time.Now()
	existing.UpdatedAt = updates["updated_at"].(time.Time)

	// Build column list and values
	columns := []string{"patient_id", "plan_id"}
	values := []interface{}{patientID, planID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("care_plans", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update care plan: %w", err)
	}

	return existing, nil
}

// Delete deletes a care plan
func (r *CarePlanRepository) Delete(ctx context.Context, patientID, planID string) error {
	mutation := spanner.Delete("care_plans", spanner.Key{patientID, planID})

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete care plan: %w", err)
	}

	return nil
}

// GetActiveCarePlans retrieves active care plans for a patient
func (r *CarePlanRepository) GetActiveCarePlans(ctx context.Context, patientID string) ([]*models.CarePlan, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			plan_id, patient_id, status, intent,
			title, description, period_start, period_end,
			goals, activities,
			created_by, created_at, updated_at
		FROM care_plans
		WHERE patient_id = @patient_id AND status = 'active'
		ORDER BY period_start DESC`,
		Params: map[string]interface{}{
			"patient_id": patientID,
		},
	}

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var carePlans []*models.CarePlan
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate active care plans: %w", err)
		}

		carePlan, err := scanCarePlan(row)
		if err != nil {
			return nil, err
		}
		carePlans = append(carePlans, carePlan)
	}

	return carePlans, nil
}

// scanCarePlan scans a Spanner row into a CarePlan model
func scanCarePlan(row *spanner.Row) (*models.CarePlan, error) {
	var carePlan models.CarePlan
	var goalsStr, activitiesStr sql.NullString

	err := row.Columns(
		&carePlan.PlanID,
		&carePlan.PatientID,
		&carePlan.Status,
		&carePlan.Intent,
		&carePlan.Title,
		&carePlan.Description,
		&carePlan.PeriodStart,
		&carePlan.PeriodEnd,
		&goalsStr,
		&activitiesStr,
		&carePlan.CreatedBy,
		&carePlan.CreatedAt,
		&carePlan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan care plan: %w", err)
	}

	// Convert JSONB strings back to json.RawMessage
	if goalsStr.Valid {
		carePlan.Goals = json.RawMessage(goalsStr.String)
	}
	if activitiesStr.Valid {
		carePlan.Activities = json.RawMessage(activitiesStr.String)
	}

	return &carePlan, nil
}
