package models

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/spanner"
)

// CarePlan represents a care plan for home healthcare
// FHIR R4 CarePlan resource mapping
type CarePlan struct {
	PlanID    string `json:"plan_id"`
	PatientID string `json:"patient_id"`

	Status string `json:"status"` // "draft" | "active" | "on-hold" | "revoked" | "completed"
	Intent string `json:"intent"` // "proposal" | "plan" | "order"

	Title       string         `json:"title"`
	Description spanner.NullString `json:"description,omitempty"`

	PeriodStart time.Time    `json:"period_start"`
	PeriodEnd   spanner.NullTime `json:"period_end,omitempty"`

	// JSONB fields - Goals (SMART format) and Activity plans
	Goals      json.RawMessage `json:"goals,omitempty"`
	Activities json.RawMessage `json:"activities,omitempty"`

	// Optimistic Locking
	Version int64 `json:"version"`

	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CarePlanCreateRequest represents the request body for creating a care plan
type CarePlanCreateRequest struct {
	Status string `json:"status" validate:"required,oneof=draft active on-hold revoked completed"`
	Intent string `json:"intent" validate:"required,oneof=proposal plan order"`

	Title       string  `json:"title" validate:"required,min=1,max=200"`
	Description *string `json:"description,omitempty"`

	PeriodStart time.Time  `json:"period_start" validate:"required"`
	PeriodEnd   *time.Time `json:"period_end,omitempty"`

	Goals      json.RawMessage `json:"goals,omitempty"`
	Activities json.RawMessage `json:"activities,omitempty"`

	CreatedBy string `json:"created_by" validate:"required"`
}

// CarePlanUpdateRequest represents the request body for updating a care plan
type CarePlanUpdateRequest struct {
	Status *string `json:"status,omitempty" validate:"omitempty,oneof=draft active on-hold revoked completed"`
	Intent *string `json:"intent,omitempty" validate:"omitempty,oneof=proposal plan order"`

	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty"`

	PeriodStart *time.Time `json:"period_start,omitempty"`
	PeriodEnd   *time.Time `json:"period_end,omitempty"`

	Goals      json.RawMessage `json:"goals,omitempty"`
	Activities json.RawMessage `json:"activities,omitempty"`

	ExpectedVersion *int64 `json:"expected_version,omitempty"` // Optimistic locking
}

// CarePlanFilter represents filter options for listing care plans
type CarePlanFilter struct {
	PatientID       *string
	Status          *string
	Intent          *string
	PeriodStartFrom *time.Time
	PeriodStartTo   *time.Time
	CreatedBy       *string
	Limit           int
	Offset          int
}
