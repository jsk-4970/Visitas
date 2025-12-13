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

// VisitScheduleRepository handles visit schedule data operations
type VisitScheduleRepository struct {
	spannerRepo *SpannerRepository
}

// NewVisitScheduleRepository creates a new visit schedule repository
func NewVisitScheduleRepository(spannerRepo *SpannerRepository) *VisitScheduleRepository {
	return &VisitScheduleRepository{
		spannerRepo: spannerRepo,
	}
}

// Create creates a new visit schedule
func (r *VisitScheduleRepository) Create(ctx context.Context, patientID string, req *models.VisitScheduleCreateRequest) (*models.VisitSchedule, error) {
	scheduleID := uuid.New().String()
	now := time.Now()

	schedule := &models.VisitSchedule{
		ScheduleID:               scheduleID,
		PatientID:                patientID,
		VisitDate:                req.VisitDate,
		VisitType:                req.VisitType,
		EstimatedDurationMinutes: req.EstimatedDurationMinutes,
		Status:                   req.Status,
		PriorityScore:            req.PriorityScore,
		Constraints:              req.Constraints,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	if req.TimeWindowStart != nil {
		schedule.TimeWindowStart = spanner.NullTime{Time: *req.TimeWindowStart, Valid: true}
	}
	if req.TimeWindowEnd != nil {
		schedule.TimeWindowEnd = spanner.NullTime{Time: *req.TimeWindowEnd, Valid: true}
	}
	if req.AssignedStaffID != nil {
		schedule.AssignedStaffID = spanner.NullString{StringVal: *req.AssignedStaffID, Valid: true}
	}
	if req.AssignedVehicleID != nil {
		schedule.AssignedVehicleID = spanner.NullString{StringVal: *req.AssignedVehicleID, Valid: true}
	}
	if req.CarePlanRef != nil {
		schedule.CarePlanRef = spanner.NullString{StringVal: *req.CarePlanRef, Valid: true}
	}
	if req.ActivityRef != nil {
		schedule.ActivityRef = spanner.NullString{StringVal: *req.ActivityRef, Valid: true}
	}

	// Convert JSONB fields to strings for Spanner
	var constraintsStr spanner.NullString
	if len(req.Constraints) > 0 {
		constraintsStr = spanner.NullString{StringVal: string(req.Constraints), Valid: true}
	}

	mutation := spanner.Insert("visit_schedules",
		[]string{
			"schedule_id", "patient_id", "visit_date", "visit_type",
			"time_window_start", "time_window_end", "estimated_duration_minutes",
			"assigned_staff_id", "assigned_vehicle_id",
			"status", "priority_score", "constraints",
			"care_plan_ref", "activity_ref",
			"created_at", "updated_at",
		},
		[]interface{}{
			scheduleID, patientID, req.VisitDate, req.VisitType,
			schedule.TimeWindowStart, schedule.TimeWindowEnd, req.EstimatedDurationMinutes,
			schedule.AssignedStaffID, schedule.AssignedVehicleID,
			req.Status, req.PriorityScore, constraintsStr,
			schedule.CarePlanRef, schedule.ActivityRef,
			now, now,
		},
	)

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to create visit schedule: %w", err)
	}

	return schedule, nil
}

// GetByID retrieves a visit schedule by ID
func (r *VisitScheduleRepository) GetByID(ctx context.Context, patientID, scheduleID string) (*models.VisitSchedule, error) {
	stmt := NewStatement(`SELECT
			schedule_id, patient_id, visit_date, visit_type,
			time_window_start, time_window_end, estimated_duration_minutes,
			assigned_staff_id, assigned_vehicle_id,
			status, priority_score, constraints, optimization_result,
			care_plan_ref, activity_ref,
			created_at, updated_at
		FROM visit_schedules
		WHERE patient_id = @patient_id AND schedule_id = @schedule_id`,
		map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("visit schedule not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query visit schedule: %w", err)
	}

	return scanVisitSchedule(row)
}

// List retrieves visit schedules with filters
func (r *VisitScheduleRepository) List(ctx context.Context, filter *models.VisitScheduleFilter) ([]*models.VisitSchedule, error) {
	var conditions []string
	params := make(map[string]interface{})

	if filter.PatientID != nil {
		conditions = append(conditions, "patient_id = @patient_id")
		params["patient_id"] = *filter.PatientID
	}

	if filter.VisitDateFrom != nil {
		conditions = append(conditions, "visit_date >= @visit_date_from")
		params["visit_date_from"] = *filter.VisitDateFrom
	}

	if filter.VisitDateTo != nil {
		conditions = append(conditions, "visit_date <= @visit_date_to")
		params["visit_date_to"] = *filter.VisitDateTo
	}

	if filter.VisitType != nil {
		conditions = append(conditions, "visit_type = @visit_type")
		params["visit_type"] = *filter.VisitType
	}

	if filter.AssignedStaffID != nil {
		conditions = append(conditions, "assigned_staff_id = @assigned_staff_id")
		params["assigned_staff_id"] = *filter.AssignedStaffID
	}

	if filter.AssignedVehicleID != nil {
		conditions = append(conditions, "assigned_vehicle_id = @assigned_vehicle_id")
		params["assigned_vehicle_id"] = *filter.AssignedVehicleID
	}

	if filter.Status != nil {
		conditions = append(conditions, "status = @status")
		params["status"] = *filter.Status
	}

	if filter.PriorityScoreMin != nil {
		conditions = append(conditions, "priority_score >= @priority_score_min")
		params["priority_score_min"] = *filter.PriorityScoreMin
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
			schedule_id, patient_id, visit_date, visit_type,
			time_window_start, time_window_end, estimated_duration_minutes,
			assigned_staff_id, assigned_vehicle_id,
			status, priority_score, constraints, optimization_result,
			care_plan_ref, activity_ref,
			created_at, updated_at
		FROM visit_schedules
		%s
		ORDER BY visit_date DESC, created_at DESC
		LIMIT @limit OFFSET @offset`, whereClause),
		params)

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var schedules []*models.VisitSchedule
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate visit schedules: %w", err)
		}

		schedule, err := scanVisitSchedule(row)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// Update updates a visit schedule
func (r *VisitScheduleRepository) Update(ctx context.Context, patientID, scheduleID string, req *models.VisitScheduleUpdateRequest) (*models.VisitSchedule, error) {
	// First, get the existing schedule
	existing, err := r.GetByID(ctx, patientID, scheduleID)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.VisitDate != nil {
		updates["visit_date"] = *req.VisitDate
		existing.VisitDate = *req.VisitDate
	}

	if req.VisitType != nil {
		updates["visit_type"] = *req.VisitType
		existing.VisitType = *req.VisitType
	}

	if req.TimeWindowStart != nil {
		updates["time_window_start"] = spanner.NullTime{Time: *req.TimeWindowStart, Valid: true}
		existing.TimeWindowStart = spanner.NullTime{Time: *req.TimeWindowStart, Valid: true}
	}

	if req.TimeWindowEnd != nil {
		updates["time_window_end"] = spanner.NullTime{Time: *req.TimeWindowEnd, Valid: true}
		existing.TimeWindowEnd = spanner.NullTime{Time: *req.TimeWindowEnd, Valid: true}
	}

	if req.EstimatedDurationMinutes != nil {
		updates["estimated_duration_minutes"] = *req.EstimatedDurationMinutes
		existing.EstimatedDurationMinutes = *req.EstimatedDurationMinutes
	}

	if req.AssignedStaffID != nil {
		updates["assigned_staff_id"] = spanner.NullString{StringVal: *req.AssignedStaffID, Valid: true}
		existing.AssignedStaffID = spanner.NullString{StringVal: *req.AssignedStaffID, Valid: true}
	}

	if req.AssignedVehicleID != nil {
		updates["assigned_vehicle_id"] = spanner.NullString{StringVal: *req.AssignedVehicleID, Valid: true}
		existing.AssignedVehicleID = spanner.NullString{StringVal: *req.AssignedVehicleID, Valid: true}
	}

	if req.Status != nil {
		updates["status"] = *req.Status
		existing.Status = *req.Status
	}

	if req.PriorityScore != nil {
		updates["priority_score"] = *req.PriorityScore
		existing.PriorityScore = *req.PriorityScore
	}

	if len(req.Constraints) > 0 {
		updates["constraints"] = spanner.NullString{StringVal: string(req.Constraints), Valid: true}
		existing.Constraints = req.Constraints
	}

	if len(req.OptimizationResult) > 0 {
		updates["optimization_result"] = spanner.NullString{StringVal: string(req.OptimizationResult), Valid: true}
		existing.OptimizationResult = req.OptimizationResult
	}

	if req.CarePlanRef != nil {
		updates["care_plan_ref"] = spanner.NullString{StringVal: *req.CarePlanRef, Valid: true}
		existing.CarePlanRef = spanner.NullString{StringVal: *req.CarePlanRef, Valid: true}
	}

	if req.ActivityRef != nil {
		updates["activity_ref"] = spanner.NullString{StringVal: *req.ActivityRef, Valid: true}
		existing.ActivityRef = spanner.NullString{StringVal: *req.ActivityRef, Valid: true}
	}

	if len(updates) == 0 {
		return existing, nil
	}

	updates["updated_at"] = time.Now()
	existing.UpdatedAt = updates["updated_at"].(time.Time)

	// Build column list and values
	columns := []string{"patient_id", "schedule_id"}
	values := []interface{}{patientID, scheduleID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("visit_schedules", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update visit schedule: %w", err)
	}

	return existing, nil
}

// Delete soft deletes a visit schedule (if soft delete is implemented, otherwise hard delete)
func (r *VisitScheduleRepository) Delete(ctx context.Context, patientID, scheduleID string) error {
	mutation := spanner.Delete("visit_schedules", spanner.Key{patientID, scheduleID})

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete visit schedule: %w", err)
	}

	return nil
}

// GetUpcomingSchedules retrieves upcoming schedules for a patient
func (r *VisitScheduleRepository) GetUpcomingSchedules(ctx context.Context, patientID string, days int) ([]*models.VisitSchedule, error) {
	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	stmt := NewStatement(`SELECT
			schedule_id, patient_id, visit_date, visit_type,
			time_window_start, time_window_end, estimated_duration_minutes,
			assigned_staff_id, assigned_vehicle_id,
			status, priority_score, constraints, optimization_result,
			care_plan_ref, activity_ref,
			created_at, updated_at
		FROM visit_schedules
		WHERE patient_id = @patient_id
		  AND visit_date >= @now
		  AND visit_date <= @end_date
		  AND status NOT IN ('cancelled', 'completed')
		ORDER BY visit_date ASC, time_window_start ASC`,
		map[string]interface{}{
			"patient_id": patientID,
			"now":        now,
			"end_date":   endDate,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var schedules []*models.VisitSchedule
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate upcoming schedules: %w", err)
		}

		schedule, err := scanVisitSchedule(row)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// scanVisitSchedule scans a Spanner row into a VisitSchedule model
func scanVisitSchedule(row *spanner.Row) (*models.VisitSchedule, error) {
	var schedule models.VisitSchedule
	var constraintsStr, optimizationResultStr spanner.NullString

	err := row.Columns(
		&schedule.ScheduleID,
		&schedule.PatientID,
		&schedule.VisitDate,
		&schedule.VisitType,
		&schedule.TimeWindowStart,
		&schedule.TimeWindowEnd,
		&schedule.EstimatedDurationMinutes,
		&schedule.AssignedStaffID,
		&schedule.AssignedVehicleID,
		&schedule.Status,
		&schedule.PriorityScore,
		&constraintsStr,
		&optimizationResultStr,
		&schedule.CarePlanRef,
		&schedule.ActivityRef,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan visit schedule: %w", err)
	}

	// Convert JSONB strings back to json.RawMessage
	if constraintsStr.Valid {
		schedule.Constraints = json.RawMessage(constraintsStr.StringVal)
	}
	if optimizationResultStr.Valid {
		schedule.OptimizationResult = json.RawMessage(optimizationResultStr.StringVal)
	}

	return &schedule, nil
}
