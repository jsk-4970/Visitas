package models

import (
	"time"

	"cloud.google.com/go/spanner"
)

// ClinicalStatus represents the clinical status of a condition (FHIR)
type ClinicalStatus string

const (
	ClinicalStatusActive     ClinicalStatus = "active"
	ClinicalStatusRecurrence ClinicalStatus = "recurrence"
	ClinicalStatusRelapse    ClinicalStatus = "relapse"
	ClinicalStatusInactive   ClinicalStatus = "inactive"
	ClinicalStatusRemission  ClinicalStatus = "remission"
	ClinicalStatusResolved   ClinicalStatus = "resolved"
)

// ConditionVerificationStatus represents the verification status (FHIR)
type ConditionVerificationStatus string

const (
	VerificationStatusUnconfirmed    ConditionVerificationStatus = "unconfirmed"
	VerificationStatusProvisional    ConditionVerificationStatus = "provisional"
	VerificationStatusDifferential   ConditionVerificationStatus = "differential"
	VerificationStatusConfirmed      ConditionVerificationStatus = "confirmed"
	VerificationStatusRefuted        ConditionVerificationStatus = "refuted"
	VerificationStatusEnteredInError ConditionVerificationStatus = "entered-in-error"
)

// ConditionSeverity represents the severity of a condition (FHIR)
type ConditionSeverity string

const (
	SeverityMild           ConditionSeverity = "mild"
	SeverityModerate       ConditionSeverity = "moderate"
	SeveritySevere         ConditionSeverity = "severe"
	SeverityLifeThreatening ConditionSeverity = "life-threatening"
)

// MedicalCondition represents a patient's medical condition/diagnosis
// Based on FHIR R4 Condition Resource
type MedicalCondition struct {
	ConditionID        string `json:"condition_id" spanner:"condition_id"`
	PatientID          string `json:"patient_id" spanner:"patient_id"`

	// Clinical Status (FHIR)
	ClinicalStatus     string `json:"clinical_status" spanner:"clinical_status"`
	VerificationStatus string `json:"verification_status" spanner:"verification_status"`

	// Category & Severity
	Category string `json:"category,omitempty" spanner:"category"` // problem-list-item, encounter-diagnosis
	Severity string `json:"severity,omitempty" spanner:"severity"`

	// Condition Code (ICD-10 or SNOMED CT)
	CodeSystem  string `json:"code_system,omitempty" spanner:"code_system"` // ICD-10, SNOMED-CT
	Code        string `json:"code,omitempty" spanner:"code"`                // e.g., E11.9 for Type 2 diabetes
	DisplayName string `json:"display_name" spanner:"display_name"`          // 疾患名

	// Body Site
	BodySite string `json:"body_site,omitempty" spanner:"body_site"`

	// Onset Information
	OnsetDate spanner.NullTime `json:"onset_date,omitempty" spanner:"onset_date"`
	OnsetAge  int              `json:"onset_age,omitempty" spanner:"onset_age"`
	OnsetNote string           `json:"onset_note,omitempty" spanner:"onset_note"`

	// Abatement Information (if resolved)
	AbatementDate spanner.NullTime `json:"abatement_date,omitempty" spanner:"abatement_date"`
	AbatementNote string           `json:"abatement_note,omitempty" spanner:"abatement_note"`

	// Recorded Information
	RecordedDate spanner.NullTime `json:"recorded_date" spanner:"recorded_date"`
	RecordedBy   string           `json:"recorded_by,omitempty" spanner:"recorded_by"`

	// Notes
	ClinicalNotes   string `json:"clinical_notes,omitempty" spanner:"clinical_notes"`
	PatientComments string `json:"patient_comments,omitempty" spanner:"patient_comments"`

	// Audit Timestamps
	CreatedAt time.Time `json:"created_at" spanner:"created_at"`
	CreatedBy string    `json:"created_by,omitempty" spanner:"created_by"`
	UpdatedAt time.Time `json:"updated_at" spanner:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty" spanner:"updated_by"`

	// Soft Delete
	Deleted   bool             `json:"deleted" spanner:"deleted"`
	DeletedAt spanner.NullTime `json:"deleted_at,omitempty" spanner:"deleted_at"`
}

// MedicalConditionCreateRequest represents the request to add a condition
type MedicalConditionCreateRequest struct {
	PatientID          string     `json:"patient_id" validate:"required"`
	ClinicalStatus     string     `json:"clinical_status" validate:"required"`
	VerificationStatus string     `json:"verification_status" validate:"required"`
	Category           string     `json:"category,omitempty"`
	Severity           string     `json:"severity,omitempty"`
	CodeSystem         string     `json:"code_system,omitempty"`
	Code               string     `json:"code,omitempty"`
	DisplayName        string     `json:"display_name" validate:"required"`
	BodySite           string     `json:"body_site,omitempty"`
	OnsetDate          *time.Time `json:"onset_date,omitempty"`
	OnsetAge           int        `json:"onset_age,omitempty"`
	OnsetNote          string     `json:"onset_note,omitempty"`
	ClinicalNotes      string     `json:"clinical_notes,omitempty"`
}

// MedicalConditionUpdateRequest represents the request to update a condition
type MedicalConditionUpdateRequest struct {
	ClinicalStatus     *string    `json:"clinical_status,omitempty"`
	VerificationStatus *string    `json:"verification_status,omitempty"`
	Severity           *string    `json:"severity,omitempty"`
	AbatementDate      *time.Time `json:"abatement_date,omitempty"`
	AbatementNote      *string    `json:"abatement_note,omitempty"`
	ClinicalNotes      *string    `json:"clinical_notes,omitempty"`
	PatientComments    *string    `json:"patient_comments,omitempty"`
}

// IsActive checks if the condition is currently active
func (c *MedicalCondition) IsActive() bool {
	if c.Deleted {
		return false
	}

	activeStatuses := []string{
		string(ClinicalStatusActive),
		string(ClinicalStatusRecurrence),
		string(ClinicalStatusRelapse),
	}

	for _, status := range activeStatuses {
		if c.ClinicalStatus == status {
			return true
		}
	}

	return false
}

// IsResolved checks if the condition is resolved
func (c *MedicalCondition) IsResolved() bool {
	return c.ClinicalStatus == string(ClinicalStatusResolved)
}
