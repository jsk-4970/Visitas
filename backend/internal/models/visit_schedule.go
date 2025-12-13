package models

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/spanner"
)

// VisitSchedule represents a visit schedule for home healthcare
type VisitSchedule struct {
	ScheduleID   string     `json:"schedule_id"`
	PatientID    string     `json:"patient_id"`
	VisitDate    civil.Date `json:"visit_date"`
	VisitType    string     `json:"visit_type"` // "regular" | "emergency" | "initial_assessment" | "terminal_care"

	// Visit time window
	TimeWindowStart          spanner.NullTime `json:"time_window_start,omitempty"`
	TimeWindowEnd            spanner.NullTime `json:"time_window_end,omitempty"`
	EstimatedDurationMinutes int64            `json:"estimated_duration_minutes"`

	// Staff assignment
	AssignedStaffID   spanner.NullString `json:"assigned_staff_id,omitempty"`
	AssignedVehicleID spanner.NullString `json:"assigned_vehicle_id,omitempty"`

	// Status
	Status string `json:"status"` // "draft" | "optimized" | "assigned" | "in_progress" | "completed" | "cancelled"

	// Route Optimization integration
	PriorityScore      int64           `json:"priority_score"`
	Constraints        json.RawMessage `json:"constraints,omitempty"`         // JSONB - Google Maps API Shipment.VisitRequest equivalent
	OptimizationResult json.RawMessage `json:"optimization_result,omitempty"` // JSONB - API Response

	// Link to care plan
	CarePlanRef  spanner.NullString `json:"care_plan_ref,omitempty"`
	ActivityRef  spanner.NullString `json:"activity_ref,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VisitScheduleCreateRequest represents the request body for creating a visit schedule
type VisitScheduleCreateRequest struct {
	VisitDate                time.Time       `json:"visit_date" validate:"required"`
	VisitType                string          `json:"visit_type" validate:"required,oneof=regular emergency initial_assessment terminal_care"`
	TimeWindowStart          *time.Time      `json:"time_window_start,omitempty"`
	TimeWindowEnd            *time.Time      `json:"time_window_end,omitempty"`
	EstimatedDurationMinutes int64           `json:"estimated_duration_minutes" validate:"required,min=5,max=480"`
	AssignedStaffID          *string         `json:"assigned_staff_id,omitempty"`
	AssignedVehicleID        *string         `json:"assigned_vehicle_id,omitempty"`
	Status                   string          `json:"status" validate:"required,oneof=draft optimized assigned in_progress completed cancelled"`
	PriorityScore            int64           `json:"priority_score" validate:"min=1,max=10"`
	Constraints              json.RawMessage `json:"constraints,omitempty"`
	CarePlanRef              *string         `json:"care_plan_ref,omitempty"`
	ActivityRef              *string         `json:"activity_ref,omitempty"`
}

// VisitScheduleUpdateRequest represents the request body for updating a visit schedule
type VisitScheduleUpdateRequest struct {
	VisitDate                *time.Time      `json:"visit_date,omitempty"`
	VisitType                *string         `json:"visit_type,omitempty" validate:"omitempty,oneof=regular emergency initial_assessment terminal_care"`
	TimeWindowStart          *time.Time      `json:"time_window_start,omitempty"`
	TimeWindowEnd            *time.Time      `json:"time_window_end,omitempty"`
	EstimatedDurationMinutes *int64          `json:"estimated_duration_minutes,omitempty" validate:"omitempty,min=5,max=480"`
	AssignedStaffID          *string         `json:"assigned_staff_id,omitempty"`
	AssignedVehicleID        *string         `json:"assigned_vehicle_id,omitempty"`
	Status                   *string         `json:"status,omitempty" validate:"omitempty,oneof=draft optimized assigned in_progress completed cancelled"`
	PriorityScore            *int64          `json:"priority_score,omitempty" validate:"omitempty,min=1,max=10"`
	Constraints              json.RawMessage `json:"constraints,omitempty"`
	OptimizationResult       json.RawMessage `json:"optimization_result,omitempty"`
	CarePlanRef              *string         `json:"care_plan_ref,omitempty"`
	ActivityRef              *string         `json:"activity_ref,omitempty"`
}

// VisitScheduleFilter represents filter options for listing visit schedules
type VisitScheduleFilter struct {
	PatientID         *string
	VisitDateFrom     *time.Time
	VisitDateTo       *time.Time
	VisitType         *string
	AssignedStaffID   *string
	AssignedVehicleID *string
	Status            *string
	PriorityScoreMin  *int
	Limit             int
	Offset            int
}
