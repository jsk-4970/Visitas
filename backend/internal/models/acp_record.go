package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ACPRecord represents an Advance Care Planning record for end-of-life care decision support
type ACPRecord struct {
	ACPID       string    `json:"acp_id"`
	PatientID   string    `json:"patient_id"`
	RecordedDate time.Time `json:"recorded_date"`
	Version     int       `json:"version"`
	Status      string    `json:"status"` // "draft" | "active" | "superseded"

	// Decision maker
	DecisionMaker  string         `json:"decision_maker"` // "patient" | "proxy" | "guardian"
	ProxyPersonID  sql.NullString `json:"proxy_person_id,omitempty"`

	// ACP content
	Directives       json.RawMessage `json:"directives"`          // JSONB - Specific directives (DNAR, ventilator, etc.)
	ValuesNarrative  sql.NullString  `json:"values_narrative,omitempty"` // Value description

	// Legal documents
	LegalDocuments   json.RawMessage `json:"legal_documents,omitempty"`  // JSONB - Links to consent forms, living wills, etc.

	// ACP process
	DiscussionLog    json.RawMessage `json:"discussion_log,omitempty"`   // JSONB - Discussion history

	// Security
	DataSensitivity      string          `json:"data_sensitivity"`       // Default "highly_confidential"
	AccessRestrictedTo   json.RawMessage `json:"access_restricted_to,omitempty"` // JSONB - Access permission list

	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// ACPRecordCreateRequest represents the request body for creating an ACP record
type ACPRecordCreateRequest struct {
	RecordedDate      time.Time       `json:"recorded_date" validate:"required"`
	Status            string          `json:"status" validate:"required,oneof=draft active superseded"`
	DecisionMaker     string          `json:"decision_maker" validate:"required,oneof=patient proxy guardian"`
	ProxyPersonID     *string         `json:"proxy_person_id,omitempty"`
	Directives        json.RawMessage `json:"directives" validate:"required"`
	ValuesNarrative   *string         `json:"values_narrative,omitempty"`
	LegalDocuments    json.RawMessage `json:"legal_documents,omitempty"`
	DiscussionLog     json.RawMessage `json:"discussion_log,omitempty"`
	DataSensitivity   *string         `json:"data_sensitivity,omitempty"`
	AccessRestrictedTo json.RawMessage `json:"access_restricted_to,omitempty"`
	CreatedBy         string          `json:"created_by" validate:"required"`
}

// ACPRecordUpdateRequest represents the request body for updating an ACP record
type ACPRecordUpdateRequest struct {
	RecordedDate      *time.Time      `json:"recorded_date,omitempty"`
	Status            *string         `json:"status,omitempty" validate:"omitempty,oneof=draft active superseded"`
	DecisionMaker     *string         `json:"decision_maker,omitempty" validate:"omitempty,oneof=patient proxy guardian"`
	ProxyPersonID     *string         `json:"proxy_person_id,omitempty"`
	Directives        json.RawMessage `json:"directives,omitempty"`
	ValuesNarrative   *string         `json:"values_narrative,omitempty"`
	LegalDocuments    json.RawMessage `json:"legal_documents,omitempty"`
	DiscussionLog     json.RawMessage `json:"discussion_log,omitempty"`
	DataSensitivity   *string         `json:"data_sensitivity,omitempty"`
	AccessRestrictedTo json.RawMessage `json:"access_restricted_to,omitempty"`
}

// ACPRecordFilter represents filter options for listing ACP records
type ACPRecordFilter struct {
	PatientID    *string
	Status       *string
	RecordedFrom *time.Time
	RecordedTo   *time.Time
	DecisionMaker *string
	Limit        int
	Offset       int
}
