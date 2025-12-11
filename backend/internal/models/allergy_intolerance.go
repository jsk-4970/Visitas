package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// AllergyClinicalStatus represents the clinical status of an allergy (FHIR)
type AllergyClinicalStatus string

const (
	AllergyClinicalStatusActive   AllergyClinicalStatus = "active"
	AllergyClinicalStatusInactive AllergyClinicalStatus = "inactive"
	AllergyClinicalStatusResolved AllergyClinicalStatus = "resolved"
)

// AllergyVerificationStatus represents the verification status (FHIR)
type AllergyVerificationStatus string

const (
	AllergyVerificationStatusUnconfirmed AllergyVerificationStatus = "unconfirmed"
	AllergyVerificationStatusConfirmed   AllergyVerificationStatus = "confirmed"
	AllergyVerificationStatusRefuted     AllergyVerificationStatus = "refuted"
)

// AllergyType represents the type of allergy or intolerance (FHIR)
type AllergyType string

const (
	AllergyTypeAllergy     AllergyType = "allergy"     // Immune response
	AllergyTypeIntolerance AllergyType = "intolerance" // Non-immune response
)

// AllergyCategory represents the category of allergen (FHIR)
type AllergyCategory string

const (
	AllergyCategoryFood        AllergyCategory = "food"
	AllergyCategoryMedication  AllergyCategory = "medication"
	AllergyCategoryEnvironment AllergyCategory = "environment"
	AllergyCategoryBiologic    AllergyCategory = "biologic"
)

// AllergyCriticality represents the potential severity (FHIR)
type AllergyCriticality string

const (
	AllergyCriticalityLow            AllergyCriticality = "low"
	AllergyCriticalityHigh           AllergyCriticality = "high"
	AllergyCriticalityUnableToAssess AllergyCriticality = "unable-to-assess"
)

// AllergyIntolerance represents a patient's allergy or adverse reaction
// Based on FHIR R4 AllergyIntolerance Resource
type AllergyIntolerance struct {
	AllergyID          string `json:"allergy_id" spanner:"allergy_id"`
	PatientID          string `json:"patient_id" spanner:"patient_id"`

	// Status (FHIR)
	ClinicalStatus     string `json:"clinical_status" spanner:"clinical_status"`
	VerificationStatus string `json:"verification_status" spanner:"verification_status"`

	// Type & Category
	Type        string `json:"type" spanner:"type"`               // allergy, intolerance
	Category    string `json:"category" spanner:"category"`       // food, medication, environment, biologic
	Criticality string `json:"criticality" spanner:"criticality"` // low, high, unable-to-assess

	// Allergen Code (SNOMED CT, RxNorm, etc.)
	CodeSystem  string `json:"code_system,omitempty" spanner:"code_system"`
	Code        string `json:"code,omitempty" spanner:"code"`
	DisplayName string `json:"display_name" spanner:"display_name"` // アレルゲン名

	// Reactions (JSONB array)
	Reactions json.RawMessage `json:"reactions,omitempty" spanner:"reactions"`

	// Generated Column: Maximum Severity
	MaxSeverity string `json:"max_severity,omitempty" spanner:"max_severity"`

	// Onset Information
	OnsetDate sql.NullTime `json:"onset_date,omitempty" spanner:"onset_date"`
	OnsetAge  int          `json:"onset_age,omitempty" spanner:"onset_age"`
	OnsetNote string       `json:"onset_note,omitempty" spanner:"onset_note"`

	// Last Occurrence
	LastOccurrenceDate sql.NullTime `json:"last_occurrence_date,omitempty" spanner:"last_occurrence_date"`

	// Recorded Information
	RecordedDate sql.NullTime `json:"recorded_date" spanner:"recorded_date"`
	RecordedBy   string       `json:"recorded_by,omitempty" spanner:"recorded_by"`

	// Notes
	ClinicalNotes   string `json:"clinical_notes,omitempty" spanner:"clinical_notes"`
	PatientComments string `json:"patient_comments,omitempty" spanner:"patient_comments"`

	// Audit Timestamps
	CreatedAt time.Time `json:"created_at" spanner:"created_at"`
	CreatedBy string    `json:"created_by,omitempty" spanner:"created_by"`
	UpdatedAt time.Time `json:"updated_at" spanner:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty" spanner:"updated_by"`

	// Soft Delete
	Deleted   bool         `json:"deleted" spanner:"deleted"`
	DeletedAt sql.NullTime `json:"deleted_at,omitempty" spanner:"deleted_at"`
}

// AllergyReaction represents a specific reaction to an allergen
type AllergyReaction struct {
	Substance      string    `json:"substance"`               // アレルゲン物質
	Manifestation  []string  `json:"manifestation"`           // 症状（蕁麻疹、呼吸困難等）
	Severity       string    `json:"severity"`                // mild, moderate, severe
	Onset          time.Time `json:"onset,omitempty"`         // 発症時刻
	Duration       string    `json:"duration,omitempty"`      // 持続時間
	ExposureRoute  string    `json:"exposureRoute,omitempty"` // 経口、注射、吸入、接触
	Note           string    `json:"note,omitempty"`          // 詳細な経過記録
}

// AllergyIntoleranceCreateRequest represents the request to add an allergy
type AllergyIntoleranceCreateRequest struct {
	PatientID          string             `json:"patient_id" validate:"required"`
	ClinicalStatus     string             `json:"clinical_status" validate:"required"`
	VerificationStatus string             `json:"verification_status" validate:"required"`
	Type               string             `json:"type" validate:"required,oneof=allergy intolerance"`
	Category           string             `json:"category" validate:"required,oneof=food medication environment biologic"`
	Criticality        string             `json:"criticality" validate:"required"`
	CodeSystem         string             `json:"code_system,omitempty"`
	Code               string             `json:"code,omitempty"`
	DisplayName        string             `json:"display_name" validate:"required"`
	Reactions          []AllergyReaction  `json:"reactions,omitempty"`
	OnsetDate          *time.Time         `json:"onset_date,omitempty"`
	OnsetAge           int                `json:"onset_age,omitempty"`
	OnsetNote          string             `json:"onset_note,omitempty"`
	ClinicalNotes      string             `json:"clinical_notes,omitempty"`
}

// AllergyIntoleranceUpdateRequest represents the request to update an allergy
type AllergyIntoleranceUpdateRequest struct {
	ClinicalStatus      *string            `json:"clinical_status,omitempty"`
	VerificationStatus  *string            `json:"verification_status,omitempty"`
	Criticality         *string            `json:"criticality,omitempty"`
	Reactions           *[]AllergyReaction `json:"reactions,omitempty"`
	LastOccurrenceDate  *time.Time         `json:"last_occurrence_date,omitempty"`
	ClinicalNotes       *string            `json:"clinical_notes,omitempty"`
	PatientComments     *string            `json:"patient_comments,omitempty"`
}

// Helper methods

// GetReactions parses reactions JSONB
func (a *AllergyIntolerance) GetReactions() ([]AllergyReaction, error) {
	var reactions []AllergyReaction
	if len(a.Reactions) == 0 {
		return reactions, nil
	}
	err := json.Unmarshal(a.Reactions, &reactions)
	return reactions, err
}

// IsActive checks if the allergy is currently active
func (a *AllergyIntolerance) IsActive() bool {
	if a.Deleted {
		return false
	}

	return a.ClinicalStatus == string(AllergyClinicalStatusActive)
}

// IsHighRisk checks if the allergy is high risk
func (a *AllergyIntolerance) IsHighRisk() bool {
	if !a.IsActive() {
		return false
	}

	return a.Criticality == string(AllergyCriticalityHigh)
}

// IsMedicationAllergy checks if this is a medication allergy
func (a *AllergyIntolerance) IsMedicationAllergy() bool {
	return a.Category == string(AllergyCategoryMedication)
}
