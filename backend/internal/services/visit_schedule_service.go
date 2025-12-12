package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// VisitScheduleService handles business logic for visit schedules
type VisitScheduleService struct {
	visitScheduleRepo *repository.VisitScheduleRepository
	patientRepo       *repository.PatientRepository
}

// NewVisitScheduleService creates a new visit schedule service
func NewVisitScheduleService(
	visitScheduleRepo *repository.VisitScheduleRepository,
	patientRepo *repository.PatientRepository,
) *VisitScheduleService {
	return &VisitScheduleService{
		visitScheduleRepo: visitScheduleRepo,
		patientRepo:       patientRepo,
	}
}

// CreateVisitSchedule creates a new visit schedule with validation and access control
func (s *VisitScheduleService) CreateVisitSchedule(ctx context.Context, patientID string, req *models.VisitScheduleCreateRequest, createdBy string) (*models.VisitSchedule, error) {
	// Check if user has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, createdBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized visit schedule creation attempt", map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to create visit schedules for this patient")
	}

	// Validate visit type
	validVisitTypes := map[string]bool{
		"regular":            true,
		"emergency":          true,
		"initial_assessment": true,
		"terminal_care":      true,
	}
	if !validVisitTypes[req.VisitType] {
		logger.WarnContext(ctx, "Invalid visit type", map[string]interface{}{
			"visit_type": req.VisitType,
		})
		return nil, fmt.Errorf("invalid visit type: %s", req.VisitType)
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft":       true,
		"optimized":   true,
		"assigned":    true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}
	if !validStatuses[req.Status] {
		logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
			"status": req.Status,
		})
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate priority score
	if req.PriorityScore < 1 || req.PriorityScore > 10 {
		req.PriorityScore = 5 // Default
	}

	// Validate estimated duration
	if req.EstimatedDurationMinutes < 5 || req.EstimatedDurationMinutes > 480 {
		logger.WarnContext(ctx, "Invalid estimated duration", map[string]interface{}{
			"estimated_duration": req.EstimatedDurationMinutes,
		})
		return nil, fmt.Errorf("invalid estimated duration: must be between 5 and 480 minutes")
	}

	// Validate time window consistency
	if req.TimeWindowStart != nil && req.TimeWindowEnd != nil {
		if req.TimeWindowEnd.Before(*req.TimeWindowStart) {
			logger.WarnContext(ctx, "Invalid time window", map[string]interface{}{
				"time_window_start": req.TimeWindowStart,
				"time_window_end":   req.TimeWindowEnd,
			})
			return nil, fmt.Errorf("time window end cannot be before time window start")
		}
	}

	schedule, err := s.visitScheduleRepo.Create(ctx, patientID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create visit schedule", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create visit schedule: %w", err)
	}

	logger.InfoContext(ctx, "Visit schedule created successfully", map[string]interface{}{
		"schedule_id": schedule.ScheduleID,
		"patient_id":  schedule.PatientID,
		"created_by":  createdBy,
	})

	return schedule, nil
}

// GetVisitSchedule retrieves a visit schedule by ID with access control
func (s *VisitScheduleService) GetVisitSchedule(ctx context.Context, patientID, scheduleID, requestorID string) (*models.VisitSchedule, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"schedule_id":  scheduleID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized visit schedule access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"schedule_id":  scheduleID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this visit schedule")
	}

	return s.visitScheduleRepo.GetByID(ctx, patientID, scheduleID)
}

// ListVisitSchedules lists visit schedules with filters and access control
func (s *VisitScheduleService) ListVisitSchedules(ctx context.Context, filter *models.VisitScheduleFilter, requestorID string) ([]*models.VisitSchedule, error) {
	// If filtering by patient ID, check access
	if filter.PatientID != nil {
		hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, *filter.PatientID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
				"patient_id":   *filter.PatientID,
				"requestor_id": requestorID,
			})
			return nil, fmt.Errorf("failed to check access: %w", err)
		}

		if !hasAccess {
			logger.WarnContext(ctx, "Unauthorized visit schedules list attempt", map[string]interface{}{
				"patient_id":   *filter.PatientID,
				"requestor_id": requestorID,
			})
			return nil, fmt.Errorf("access denied: you do not have permission to view visit schedules for this patient")
		}
	}

	return s.visitScheduleRepo.List(ctx, filter)
}

// UpdateVisitSchedule updates a visit schedule with validation and access control
func (s *VisitScheduleService) UpdateVisitSchedule(ctx context.Context, patientID, scheduleID string, req *models.VisitScheduleUpdateRequest, updatedBy string) (*models.VisitSchedule, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"updated_by":  updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized visit schedule update attempt", map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"updated_by":  updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this visit schedule")
	}

	// Validate visit type if provided
	if req.VisitType != nil {
		validVisitTypes := map[string]bool{
			"regular":            true,
			"emergency":          true,
			"initial_assessment": true,
			"terminal_care":      true,
		}
		if !validVisitTypes[*req.VisitType] {
			logger.WarnContext(ctx, "Invalid visit type", map[string]interface{}{
				"visit_type": *req.VisitType,
			})
			return nil, fmt.Errorf("invalid visit type: %s", *req.VisitType)
		}
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"draft":       true,
			"optimized":   true,
			"assigned":    true,
			"in_progress": true,
			"completed":   true,
			"cancelled":   true,
		}
		if !validStatuses[*req.Status] {
			logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
				"status": *req.Status,
			})
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Validate priority score if provided
	if req.PriorityScore != nil {
		if *req.PriorityScore < 1 || *req.PriorityScore > 10 {
			logger.WarnContext(ctx, "Invalid priority score", map[string]interface{}{
				"priority_score": *req.PriorityScore,
			})
			return nil, fmt.Errorf("invalid priority score: must be between 1 and 10")
		}
	}

	// Validate estimated duration if provided
	if req.EstimatedDurationMinutes != nil {
		if *req.EstimatedDurationMinutes < 5 || *req.EstimatedDurationMinutes > 480 {
			logger.WarnContext(ctx, "Invalid estimated duration", map[string]interface{}{
				"estimated_duration": *req.EstimatedDurationMinutes,
			})
			return nil, fmt.Errorf("invalid estimated duration: must be between 5 and 480 minutes")
		}
	}

	// Validate time window consistency if both provided
	if req.TimeWindowStart != nil && req.TimeWindowEnd != nil {
		if req.TimeWindowEnd.Before(*req.TimeWindowStart) {
			logger.WarnContext(ctx, "Invalid time window", map[string]interface{}{
				"time_window_start": req.TimeWindowStart,
				"time_window_end":   req.TimeWindowEnd,
			})
			return nil, fmt.Errorf("time window end cannot be before time window start")
		}
	}

	schedule, err := s.visitScheduleRepo.Update(ctx, patientID, scheduleID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update visit schedule", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"updated_by":  updatedBy,
		})
		return nil, fmt.Errorf("failed to update visit schedule: %w", err)
	}

	logger.InfoContext(ctx, "Visit schedule updated successfully", map[string]interface{}{
		"schedule_id": schedule.ScheduleID,
		"patient_id":  schedule.PatientID,
		"updated_by":  updatedBy,
	})

	return schedule, nil
}

// DeleteVisitSchedule deletes a visit schedule with access control
func (s *VisitScheduleService) DeleteVisitSchedule(ctx context.Context, patientID, scheduleID, deletedBy string) error {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"deleted_by":  deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized visit schedule deletion attempt", map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"deleted_by":  deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this visit schedule")
	}

	// Verify the schedule exists before deletion
	_, err = s.visitScheduleRepo.GetByID(ctx, patientID, scheduleID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get visit schedule for deletion", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
		})
		return fmt.Errorf("visit schedule not found: %w", err)
	}

	err = s.visitScheduleRepo.Delete(ctx, patientID, scheduleID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete visit schedule", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"deleted_by":  deletedBy,
		})
		return fmt.Errorf("failed to delete visit schedule: %w", err)
	}

	logger.InfoContext(ctx, "Visit schedule deleted successfully", map[string]interface{}{
		"patient_id":  patientID,
		"schedule_id": scheduleID,
		"deleted_by":  deletedBy,
	})

	return nil
}

// GetUpcomingSchedules retrieves upcoming schedules for a patient with access control
func (s *VisitScheduleService) GetUpcomingSchedules(ctx context.Context, patientID string, days int, requestorID string) ([]*models.VisitSchedule, error) {
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
		logger.WarnContext(ctx, "Unauthorized upcoming schedules access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view schedules for this patient")
	}

	if days <= 0 {
		days = 7 // Default to 7 days
	}

	return s.visitScheduleRepo.GetUpcomingSchedules(ctx, patientID, days)
}

// AssignStaff assigns a staff member to a visit schedule with access control
func (s *VisitScheduleService) AssignStaff(ctx context.Context, patientID, scheduleID, staffID, assignedBy string) (*models.VisitSchedule, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, assignedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"assigned_by": assignedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized staff assignment attempt", map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"assigned_by": assignedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to assign staff to this visit schedule")
	}

	req := &models.VisitScheduleUpdateRequest{
		AssignedStaffID: &staffID,
		Status:          stringPtr("assigned"),
	}

	schedule, err := s.visitScheduleRepo.Update(ctx, patientID, scheduleID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to assign staff", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"staff_id":    staffID,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Staff assigned to visit schedule", map[string]interface{}{
		"schedule_id": scheduleID,
		"staff_id":    staffID,
		"assigned_by": assignedBy,
	})

	return schedule, nil
}

// AssignVehicle assigns a vehicle to a visit schedule with access control
func (s *VisitScheduleService) AssignVehicle(ctx context.Context, patientID, scheduleID, vehicleID, assignedBy string) (*models.VisitSchedule, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, assignedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"assigned_by": assignedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized vehicle assignment attempt", map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"assigned_by": assignedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to assign vehicles to this visit schedule")
	}

	req := &models.VisitScheduleUpdateRequest{
		AssignedVehicleID: &vehicleID,
	}

	schedule, err := s.visitScheduleRepo.Update(ctx, patientID, scheduleID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to assign vehicle", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"vehicle_id":  vehicleID,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Vehicle assigned to visit schedule", map[string]interface{}{
		"schedule_id": scheduleID,
		"vehicle_id":  vehicleID,
		"assigned_by": assignedBy,
	})

	return schedule, nil
}

// UpdateStatus updates the status of a visit schedule with access control
func (s *VisitScheduleService) UpdateStatus(ctx context.Context, patientID, scheduleID, status, updatedBy string) (*models.VisitSchedule, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"updated_by":  updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized status update attempt", map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"updated_by":  updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this visit schedule")
	}

	validStatuses := map[string]bool{
		"draft":       true,
		"optimized":   true,
		"assigned":    true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}
	if !validStatuses[status] {
		logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
			"status": status,
		})
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	req := &models.VisitScheduleUpdateRequest{
		Status: &status,
	}

	schedule, err := s.visitScheduleRepo.Update(ctx, patientID, scheduleID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update status", err, map[string]interface{}{
			"patient_id":  patientID,
			"schedule_id": scheduleID,
			"status":      status,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Visit schedule status updated", map[string]interface{}{
		"schedule_id": scheduleID,
		"status":      status,
		"updated_by":  updatedBy,
	})

	return schedule, nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
