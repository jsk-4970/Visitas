package models

import (
	"encoding/json"
	"time"
)

// MedicalRecord represents a medical record with SOAP notes
// Phase 1 Sprint 6: 基本カルテ機能
type MedicalRecord struct {
	RecordID       string    `json:"record_id"`
	PatientID      string    `json:"patient_id"`
	VisitStartedAt time.Time `json:"visit_started_at"`
	VisitEndedAt   *time.Time `json:"visit_ended_at,omitempty"`
	VisitType      string    `json:"visit_type"` // regular, emergency, initial, follow_up, terminal_care
	PerformedBy    string    `json:"performed_by"`
	Status         string    `json:"status"` // draft, in_progress, completed, cancelled
	ScheduleID     *string   `json:"schedule_id,omitempty"`

	// SOAP Content (JSONB)
	SOAPContent json.RawMessage `json:"soap_content,omitempty"`

	// Template Support
	TemplateID     *string `json:"template_id,omitempty"`
	SourceRecordID *string `json:"source_record_id,omitempty"`

	// AI Integration
	SourceType   string  `json:"source_type"` // manual, voice_to_text, ai_generated, template
	AudioFileURL *string `json:"audio_file_url,omitempty"`

	// Generated Columns (read-only)
	SOAPCompleted   bool `json:"soap_completed"`
	HasAIAssistance bool `json:"has_ai_assistance"`

	// Optimistic Locking
	Version int64 `json:"version"`

	// Audit
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy string     `json:"created_by"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdatedBy *string    `json:"updated_by,omitempty"`
	Deleted   bool       `json:"deleted"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	DeletedBy *string    `json:"deleted_by,omitempty"`
}

// MedicalRecordCreateRequest represents the request body for creating a medical record
type MedicalRecordCreateRequest struct {
	VisitStartedAt time.Time       `json:"visit_started_at" validate:"required"`
	VisitEndedAt   *time.Time      `json:"visit_ended_at,omitempty"`
	VisitType      string          `json:"visit_type" validate:"required,oneof=regular emergency initial follow_up terminal_care"`
	PerformedBy    string          `json:"performed_by" validate:"required"`
	Status         string          `json:"status" validate:"required,oneof=draft in_progress completed cancelled"`
	ScheduleID     *string         `json:"schedule_id,omitempty"`
	SOAPContent    json.RawMessage `json:"soap_content,omitempty"`
	TemplateID     *string         `json:"template_id,omitempty"`
	SourceRecordID *string         `json:"source_record_id,omitempty"`
	SourceType     string          `json:"source_type" validate:"required,oneof=manual voice_to_text ai_generated template"`
	AudioFileURL   *string         `json:"audio_file_url,omitempty"`
}

// MedicalRecordUpdateRequest represents the request body for updating a medical record
type MedicalRecordUpdateRequest struct {
	VisitEndedAt    *time.Time      `json:"visit_ended_at,omitempty"`
	VisitType       *string         `json:"visit_type,omitempty" validate:"omitempty,oneof=regular emergency initial follow_up terminal_care"`
	Status          *string         `json:"status,omitempty" validate:"omitempty,oneof=draft in_progress completed cancelled"`
	SOAPContent     json.RawMessage `json:"soap_content,omitempty"`
	ScheduleID      *string         `json:"schedule_id,omitempty"`
	TemplateID      *string         `json:"template_id,omitempty"`
	AudioFileURL    *string         `json:"audio_file_url,omitempty"`
	ExpectedVersion *int64          `json:"expected_version,omitempty"` // Optimistic locking
}

// MedicalRecordFilter represents filter options for listing medical records
type MedicalRecordFilter struct {
	PatientID     *string
	PerformedBy   *string
	Status        *string
	VisitType     *string
	ScheduleID    *string
	VisitDateFrom *time.Time
	VisitDateTo   *time.Time
	SOAPCompleted *bool
	TemplateID    *string
	SourceType    *string
	Limit         int
	Offset        int
}

// MedicalRecordTemplate represents a reusable SOAP note template
type MedicalRecordTemplate struct {
	TemplateID          string          `json:"template_id"`
	TemplateName        string          `json:"template_name"`
	TemplateDescription *string         `json:"template_description,omitempty"`
	Specialty           *string         `json:"specialty,omitempty"` // general, internal_medicine, neurology, palliative_care
	SOAPTemplate        json.RawMessage `json:"soap_template"`
	IsSystemTemplate    bool            `json:"is_system_template"`
	UsageCount          int64           `json:"usage_count"`
	CreatedAt           time.Time       `json:"created_at"`
	CreatedBy           string          `json:"created_by"`
	UpdatedAt           time.Time       `json:"updated_at"`
	UpdatedBy           *string         `json:"updated_by,omitempty"`
	Deleted             bool            `json:"deleted"`
	DeletedAt           *time.Time      `json:"deleted_at,omitempty"`
}

// MedicalRecordTemplateCreateRequest represents the request body for creating a template
type MedicalRecordTemplateCreateRequest struct {
	TemplateName        string          `json:"template_name" validate:"required,min=1,max=200"`
	TemplateDescription *string         `json:"template_description,omitempty"`
	Specialty           *string         `json:"specialty,omitempty" validate:"omitempty,oneof=general internal_medicine neurology palliative_care"`
	SOAPTemplate        json.RawMessage `json:"soap_template" validate:"required"`
	IsSystemTemplate    bool            `json:"is_system_template"`
}

// MedicalRecordTemplateUpdateRequest represents the request body for updating a template
type MedicalRecordTemplateUpdateRequest struct {
	TemplateName        *string         `json:"template_name,omitempty" validate:"omitempty,min=1,max=200"`
	TemplateDescription *string         `json:"template_description,omitempty"`
	Specialty           *string         `json:"specialty,omitempty" validate:"omitempty,oneof=general internal_medicine neurology palliative_care"`
	SOAPTemplate        json.RawMessage `json:"soap_template,omitempty"`
}

// MedicalRecordTemplateFilter represents filter options for listing templates
type MedicalRecordTemplateFilter struct {
	Specialty        *string
	IsSystemTemplate *bool
	CreatedBy        *string
	Limit            int
	Offset           int
}

// SOAPContent represents the structure of SOAP content in medical records
// This is for documentation/validation purposes - actual data is stored as json.RawMessage
type SOAPContent struct {
	Subjective *SubjectiveSection `json:"subjective,omitempty"`
	Objective  *ObjectiveSection  `json:"objective,omitempty"`
	Assessment *AssessmentSection `json:"assessment,omitempty"`
	Plan       *PlanSection       `json:"plan,omitempty"`
	Metadata   *SOAPMetadata      `json:"_metadata,omitempty"`
}

// SubjectiveSection represents the S (Subjective) section of SOAP
type SubjectiveSection struct {
	ChiefComplaint     string     `json:"chiefComplaint,omitempty"`
	Symptoms           []Symptom  `json:"symptoms,omitempty"`
	PatientNarrative   string     `json:"patientNarrative,omitempty"`
	FamilyObservations string     `json:"familyObservations,omitempty"`
	PainScale          *PainScale `json:"painScale,omitempty"`
}

// Symptom represents a symptom entry
type Symptom struct {
	Code     string `json:"code,omitempty"`     // SNOMED CT code
	Display  string `json:"display,omitempty"`  // Human-readable name
	Severity string `json:"severity,omitempty"` // mild, moderate, severe
	Onset    string `json:"onset,omitempty"`    // When it started
}

// PainScale represents pain assessment
type PainScale struct {
	Score    int    `json:"score,omitempty"`    // 0-10
	Location string `json:"location,omitempty"` // Body location
}

// ObjectiveSection represents the O (Objective) section of SOAP
type ObjectiveSection struct {
	VitalSigns         *VitalSigns       `json:"vitalSigns,omitempty"`
	PhysicalExam       map[string]string `json:"physicalExam,omitempty"` // Area -> Findings
	LinkedObservations []string          `json:"linkedObservations,omitempty"`
	LaboratoryResults  map[string]interface{} `json:"laboratoryResults,omitempty"`
}

// VitalSigns represents vital sign measurements
type VitalSigns struct {
	BloodPressure   *BloodPressure   `json:"bloodPressure,omitempty"`
	HeartRate       *Measurement     `json:"heartRate,omitempty"`
	SPO2            *Measurement     `json:"spo2,omitempty"`
	Temperature     *Measurement     `json:"temperature,omitempty"`
	RespiratoryRate *Measurement     `json:"respiratoryRate,omitempty"`
}

// BloodPressure represents blood pressure measurement
type BloodPressure struct {
	Systolic  int    `json:"systolic,omitempty"`
	Diastolic int    `json:"diastolic,omitempty"`
	Unit      string `json:"unit,omitempty"` // mmHg
}

// Measurement represents a general measurement value
type Measurement struct {
	Value float64 `json:"value,omitempty"`
	Unit  string  `json:"unit,omitempty"`
}

// AssessmentSection represents the A (Assessment) section of SOAP
type AssessmentSection struct {
	Diagnoses          []Diagnosis `json:"diagnoses,omitempty"`
	ClinicalImpression string      `json:"clinicalImpression,omitempty"`
	RiskFactors        []string    `json:"riskFactors,omitempty"`
	PrognosisNote      string      `json:"prognosisNote,omitempty"`
}

// Diagnosis represents a diagnosis entry
type Diagnosis struct {
	Code            *DiagnosisCode `json:"code,omitempty"`
	Status          string         `json:"status,omitempty"` // provisional, confirmed, differential
	LinkedCondition string         `json:"linkedCondition,omitempty"`
}

// DiagnosisCode represents a diagnosis code (ICD-10, etc.)
type DiagnosisCode struct {
	System  string `json:"system,omitempty"`  // ICD-10, SNOMED CT
	Code    string `json:"code,omitempty"`    // Code value
	Display string `json:"display,omitempty"` // Human-readable
}

// PlanSection represents the P (Plan) section of SOAP
type PlanSection struct {
	Medications      []MedicationPlan      `json:"medications,omitempty"`
	Procedures       []ProcedurePlan       `json:"procedures,omitempty"`
	CareInstructions map[string]interface{} `json:"careInstructions,omitempty"`
	NextVisit        *NextVisitPlan        `json:"nextVisit,omitempty"`
	Referrals        []ReferralPlan        `json:"referrals,omitempty"`
}

// MedicationPlan represents a medication in the plan
type MedicationPlan struct {
	Action      string `json:"action,omitempty"` // prescribe, continue, discontinue, adjust
	DrugName    string `json:"drugName,omitempty"`
	Dosage      string `json:"dosage,omitempty"`
	LinkedOrder string `json:"linkedOrder,omitempty"` // Link to medication_orders
	Note        string `json:"note,omitempty"`
}

// ProcedurePlan represents a procedure in the plan
type ProcedurePlan struct {
	Code      string `json:"code,omitempty"`
	Scheduled string `json:"scheduled,omitempty"` // Date
	Location  string `json:"location,omitempty"`
}

// NextVisitPlan represents next visit planning
type NextVisitPlan struct {
	ScheduledDate  string `json:"scheduledDate,omitempty"`
	Purpose        string `json:"purpose,omitempty"`
	LinkedSchedule string `json:"linkedSchedule,omitempty"` // Link to visit_schedules
}

// ReferralPlan represents a referral to another provider
type ReferralPlan struct {
	Specialty string `json:"specialty,omitempty"`
	Facility  string `json:"facility,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Urgency   string `json:"urgency,omitempty"` // routine, urgent, emergent
}

// SOAPMetadata represents metadata about the SOAP record
type SOAPMetadata struct {
	RecordedBy     string          `json:"recordedBy,omitempty"`
	RecordedAt     string          `json:"recordedAt,omitempty"`
	SourceType     string          `json:"sourceType,omitempty"`
	TemplateUsed   string          `json:"templateUsed,omitempty"`
	DraftMode      bool            `json:"draftMode,omitempty"`
	AIAssistance   *AIAssistance   `json:"aiAssistance,omitempty"`
	LinkedEntities *LinkedEntities `json:"linkedEntities,omitempty"`
}

// AIAssistance represents AI-assisted creation metadata
type AIAssistance struct {
	Enabled       bool    `json:"enabled,omitempty"`
	AIGenerated   bool    `json:"aiGenerated,omitempty"`
	Model         string  `json:"model,omitempty"`
	Confidence    float64 `json:"confidence,omitempty"`
	TranscriptURL string  `json:"transcriptUrl,omitempty"`
}

// LinkedEntities represents links to other domain entities
type LinkedEntities struct {
	Schedule     string   `json:"schedule,omitempty"`
	Observations []string `json:"observations,omitempty"`
	Conditions   []string `json:"conditions,omitempty"`
	Medications  []string `json:"medications,omitempty"`
	CarePlans    []string `json:"carePlans,omitempty"`
}

// CopyAsMedicalRecordRequest represents request to copy a record
type CopyAsMedicalRecordRequest struct {
	TargetPatientID *string         `json:"target_patient_id,omitempty"` // If different from source
	VisitStartedAt  time.Time       `json:"visit_started_at" validate:"required"`
	VisitType       string          `json:"visit_type" validate:"required,oneof=regular emergency initial follow_up terminal_care"`
	PerformedBy     string          `json:"performed_by" validate:"required"`
	ModifySOAP      json.RawMessage `json:"modify_soap,omitempty"` // Partial modifications to apply
}

// CreateFromTemplateRequest represents request to create from template
type CreateFromTemplateRequest struct {
	TemplateID     string          `json:"template_id" validate:"required"`
	VisitStartedAt time.Time       `json:"visit_started_at" validate:"required"`
	VisitType      string          `json:"visit_type" validate:"required,oneof=regular emergency initial follow_up terminal_care"`
	PerformedBy    string          `json:"performed_by" validate:"required"`
	InitialSOAP    json.RawMessage `json:"initial_soap,omitempty"` // Initial values to merge with template
}
