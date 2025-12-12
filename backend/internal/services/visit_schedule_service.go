package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
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

// CreateVisitSchedule creates a new visit schedule with validation
func (s *VisitScheduleService) CreateVisitSchedule(ctx context.Context, patientID string, req *models.VisitScheduleCreateRequest) (*models.VisitSchedule, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	// Validate visit type
	validVisitTypes := map[string]bool{
		"regular":            true,
		"emergency":          true,
		"initial_assessment": true,
		"terminal_care":      true,
	}
	if !validVisitTypes[req.VisitType] {
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
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate priority score
	if req.PriorityScore < 1 || req.PriorityScore > 10 {
		req.PriorityScore = 5 // Default
	}

	// Validate estimated duration
	if req.EstimatedDurationMinutes < 5 || req.EstimatedDurationMinutes > 480 {
		return nil, fmt.Errorf("invalid estimated duration: must be between 5 and 480 minutes")
	}

	// Validate time window consistency
	if req.TimeWindowStart != nil && req.TimeWindowEnd != nil {
		if req.TimeWindowEnd.Before(*req.TimeWindowStart) {
			return nil, fmt.Errorf("time window end cannot be before time window start")
		}
	}

	return s.visitScheduleRepo.Create(ctx, patientID, req)
}

// GetVisitSchedule retrieves a visit schedule by ID
func (s *VisitScheduleService) GetVisitSchedule(ctx context.Context, patientID, scheduleID string) (*models.VisitSchedule, error) {
	return s.visitScheduleRepo.GetByID(ctx, patientID, scheduleID)
}

// ListVisitSchedules lists visit schedules with filters
func (s *VisitScheduleService) ListVisitSchedules(ctx context.Context, filter *models.VisitScheduleFilter) ([]*models.VisitSchedule, error) {
	return s.visitScheduleRepo.List(ctx, filter)
}

// UpdateVisitSchedule updates a visit schedule with validation
func (s *VisitScheduleService) UpdateVisitSchedule(ctx context.Context, patientID, scheduleID string, req *models.VisitScheduleUpdateRequest) (*models.VisitSchedule, error) {
	// Validate visit type if provided
	if req.VisitType != nil {
		validVisitTypes := map[string]bool{
			"regular":            true,
			"emergency":          true,
			"initial_assessment": true,
			"terminal_care":      true,
		}
		if !validVisitTypes[*req.VisitType] {
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
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Validate priority score if provided
	if req.PriorityScore != nil {
		if *req.PriorityScore < 1 || *req.PriorityScore > 10 {
			return nil, fmt.Errorf("invalid priority score: must be between 1 and 10")
		}
	}

	// Validate estimated duration if provided
	if req.EstimatedDurationMinutes != nil {
		if *req.EstimatedDurationMinutes < 5 || *req.EstimatedDurationMinutes > 480 {
			return nil, fmt.Errorf("invalid estimated duration: must be between 5 and 480 minutes")
		}
	}

	// Validate time window consistency if both provided
	if req.TimeWindowStart != nil && req.TimeWindowEnd != nil {
		if req.TimeWindowEnd.Before(*req.TimeWindowStart) {
			return nil, fmt.Errorf("time window end cannot be before time window start")
		}
	}

	return s.visitScheduleRepo.Update(ctx, patientID, scheduleID, req)
}

// DeleteVisitSchedule deletes a visit schedule
func (s *VisitScheduleService) DeleteVisitSchedule(ctx context.Context, patientID, scheduleID string) error {
	// Verify the schedule exists before deletion
	_, err := s.visitScheduleRepo.GetByID(ctx, patientID, scheduleID)
	if err != nil {
		return fmt.Errorf("visit schedule not found: %w", err)
	}

	return s.visitScheduleRepo.Delete(ctx, patientID, scheduleID)
}

// GetUpcomingSchedules retrieves upcoming schedules for a patient
func (s *VisitScheduleService) GetUpcomingSchedules(ctx context.Context, patientID string, days int) ([]*models.VisitSchedule, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	if days <= 0 {
		days = 7 // Default to 7 days
	}

	return s.visitScheduleRepo.GetUpcomingSchedules(ctx, patientID, days)
}

// AssignStaff assigns a staff member to a visit schedule
func (s *VisitScheduleService) AssignStaff(ctx context.Context, patientID, scheduleID, staffID string) (*models.VisitSchedule, error) {
	req := &models.VisitScheduleUpdateRequest{
		AssignedStaffID: &staffID,
		Status:          stringPtr("assigned"),
	}

	return s.visitScheduleRepo.Update(ctx, patientID, scheduleID, req)
}

// AssignVehicle assigns a vehicle to a visit schedule
func (s *VisitScheduleService) AssignVehicle(ctx context.Context, patientID, scheduleID, vehicleID string) (*models.VisitSchedule, error) {
	req := &models.VisitScheduleUpdateRequest{
		AssignedVehicleID: &vehicleID,
	}

	return s.visitScheduleRepo.Update(ctx, patientID, scheduleID, req)
}

// UpdateStatus updates the status of a visit schedule
func (s *VisitScheduleService) UpdateStatus(ctx context.Context, patientID, scheduleID, status string) (*models.VisitSchedule, error) {
	validStatuses := map[string]bool{
		"draft":       true,
		"optimized":   true,
		"assigned":    true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}
	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	req := &models.VisitScheduleUpdateRequest{
		Status: &status,
	}

	return s.visitScheduleRepo.Update(ctx, patientID, scheduleID, req)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
