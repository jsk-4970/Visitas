package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ClinicalObservation represents vital signs, ADL assessments, and other clinical observations
type ClinicalObservation struct {
	ObservationID string    `json:"observation_id"`
	PatientID     string    `json:"patient_id"`
	Category      string    `json:"category"` // "vital_signs" | "adl_assessment" | "cognitive_assessment" | "pain_scale"
	Code          json.RawMessage `json:"code"` // LOINC/SNOMED CT compliant JSONB

	EffectiveDatetime time.Time `json:"effective_datetime"` // Measurement datetime
	Issued            time.Time `json:"issued"`             // Recording datetime

	Value          json.RawMessage `json:"value"`                    // Measured value (numeric, coded value, score, etc.) JSONB
	Interpretation sql.NullString  `json:"interpretation,omitempty"` // "normal" | "high" | "low" | "critical"

	PerformerID sql.NullString `json:"performer_id,omitempty"` // Measurer
	DeviceID    sql.NullString `json:"device_id,omitempty"`    // Device ID (for IoT integration)

	// Reference to visit record
	VisitRecordID sql.NullString `json:"visit_record_id,omitempty"`
}

// ClinicalObservationCreateRequest represents the request body for creating a clinical observation
type ClinicalObservationCreateRequest struct {
	Category          string          `json:"category" validate:"required,oneof=vital_signs adl_assessment cognitive_assessment pain_scale"`
	Code              json.RawMessage `json:"code" validate:"required"`
	EffectiveDatetime time.Time       `json:"effective_datetime" validate:"required"`
	Value             json.RawMessage `json:"value" validate:"required"`
	Interpretation    *string         `json:"interpretation,omitempty" validate:"omitempty,oneof=normal high low critical"`
	PerformerID       *string         `json:"performer_id,omitempty"`
	DeviceID          *string         `json:"device_id,omitempty"`
	VisitRecordID     *string         `json:"visit_record_id,omitempty"`
}

// ClinicalObservationUpdateRequest represents the request body for updating a clinical observation
type ClinicalObservationUpdateRequest struct {
	Category          *string         `json:"category,omitempty" validate:"omitempty,oneof=vital_signs adl_assessment cognitive_assessment pain_scale"`
	Code              json.RawMessage `json:"code,omitempty"`
	EffectiveDatetime *time.Time      `json:"effective_datetime,omitempty"`
	Value             json.RawMessage `json:"value,omitempty"`
	Interpretation    *string         `json:"interpretation,omitempty" validate:"omitempty,oneof=normal high low critical"`
	PerformerID       *string         `json:"performer_id,omitempty"`
	DeviceID          *string         `json:"device_id,omitempty"`
	VisitRecordID     *string         `json:"visit_record_id,omitempty"`
}

// ClinicalObservationFilter represents filter options for listing clinical observations
type ClinicalObservationFilter struct {
	PatientID             *string
	Category              *string
	EffectiveDatetimeFrom *time.Time
	EffectiveDatetimeTo   *time.Time
	PerformerID           *string
	VisitRecordID         *string
	Interpretation        *string
	Limit                 int
	Offset                int
}

// Common code structures for LOINC/SNOMED CT compliance

// ObservationCode represents a standardized code for clinical observations
type ObservationCode struct {
	System  string `json:"system"`  // "LOINC" | "SNOMED-CT"
	Code    string `json:"code"`    // e.g., "8867-4" for heart rate
	Display string `json:"display"` // Human-readable name
}

// QuantityValue represents a numeric observation value
type QuantityValue struct {
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Comparator *string `json:"comparator,omitempty"` // "<" | "<=" | ">=" | ">"
}

// CodedValue represents a coded observation value (for discrete choices)
type CodedValue struct {
	System  string `json:"system"`
	Code    string `json:"code"`
	Display string `json:"display"`
}

// BloodPressureValue represents blood pressure observation value
type BloodPressureValue struct {
	Systolic  QuantityValue `json:"systolic"`
	Diastolic QuantityValue `json:"diastolic"`
}

// ADLScore represents ADL assessment score
type ADLScore struct {
	TotalScore int               `json:"total_score"`
	Items      map[string]int    `json:"items"` // e.g., "bathing": 1, "dressing": 2
	Method     string            `json:"method"` // "Barthel", "Katz", etc.
	Notes      *string           `json:"notes,omitempty"`
}
