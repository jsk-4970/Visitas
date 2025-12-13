package models

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/spanner"
)

// PatientSocialProfile represents a patient's social context and living situation
// Based on FHIR Observation (Social History) and SDOH extensions
type PatientSocialProfile struct {
	ProfileID      string          `json:"profile_id" spanner:"profile_id"`
	PatientID      string          `json:"patient_id" spanner:"patient_id"`
	ProfileVersion int64           `json:"profile_version" spanner:"profile_version"`
	Content        json.RawMessage `json:"content" spanner:"content"` // JSONB

	// Generated Columns (for fast queries)
	LivesAlone                 bool `json:"lives_alone" spanner:"lives_alone"`
	RequiresCaregiverSupport   bool `json:"requires_caregiver_support" spanner:"requires_caregiver_support"`

	// Validity Period
	ValidFrom spanner.NullTime `json:"valid_from" spanner:"valid_from"`
	ValidTo   spanner.NullTime `json:"valid_to,omitempty" spanner:"valid_to"`

	// Assessment Information
	AssessedBy      string           `json:"assessed_by,omitempty" spanner:"assessed_by"`
	AssessedAt      spanner.NullTime `json:"assessed_at,omitempty" spanner:"assessed_at"`
	AssessmentNotes string           `json:"assessment_notes,omitempty" spanner:"assessment_notes"`

	// Audit Timestamps
	CreatedAt time.Time `json:"created_at" spanner:"created_at"`
	CreatedBy string    `json:"created_by,omitempty" spanner:"created_by"`
	UpdatedAt time.Time `json:"updated_at" spanner:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty" spanner:"updated_by"`

	// Soft Delete
	Deleted   bool             `json:"deleted" spanner:"deleted"`
	DeletedAt spanner.NullTime `json:"deleted_at,omitempty" spanner:"deleted_at"`
}

// SocialProfileContent represents the structured content of a social profile
type SocialProfileContent struct {
	LivingSituation      *LivingSituation      `json:"livingSituation,omitempty"`
	KeyPersons           []KeyPerson           `json:"keyPersons,omitempty"`
	FinancialBackground  *FinancialBackground  `json:"financialBackground,omitempty"`
	SocialSupport        *SocialSupport        `json:"socialSupport,omitempty"`
}

// LivingSituation represents the patient's living situation
type LivingSituation struct {
	LivingAlone             bool         `json:"livingAlone"`
	RequiresCaregiverSupport bool        `json:"requiresCaregiverSupport"`
	HousingType             string       `json:"housingType"` // detached, apartment, facility, other
	Accessibility           *Accessibility `json:"accessibility,omitempty"`
}

// Accessibility represents housing accessibility features
type Accessibility struct {
	WheelchairAccessible bool `json:"wheelchairAccessible"`
	ElevatorAvailable    bool `json:"elevatorAvailable"`
	StairsCount          int  `json:"stairsCount"`
}

// KeyPerson represents a significant person in the patient's life
type KeyPerson struct {
	Relationship      string           `json:"relationship"` // spouse, child, parent, sibling, other
	Name              string           `json:"name"`
	Age               int              `json:"age,omitempty"`
	ContactInfo       *ContactInfo     `json:"contactInfo,omitempty"`
	LivesWith         bool             `json:"livesWith"`
	IsPrimaryCaregiver bool            `json:"isPrimaryCaregiver"`
	CaregiverBurden   *CaregiverBurden `json:"caregiverBurden,omitempty"`
}

// ContactInfo represents contact information for a key person
type ContactInfo struct {
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

// CaregiverBurden represents caregiver burden assessment (Zarit scale)
type CaregiverBurden struct {
	ZaritScore  int       `json:"zaritScore"`     // 0-88 (higher = more burden)
	BurnoutRisk string    `json:"burnoutRisk"`    // low, moderate, high
	AssessedAt  time.Time `json:"assessedAt"`
	Notes       string    `json:"notes,omitempty"`
}

// FinancialBackground represents the patient's financial situation
type FinancialBackground struct {
	IncomeLevel          string                `json:"income_level"` // low, middle, high
	PublicAssistance     bool                  `json:"publicAssistance"`
	PublicAssistanceTypes []string             `json:"publicAssistanceTypes,omitempty"` // livelihood, medical, housing
	InsuranceCoverage    *InsuranceCoverage   `json:"insuranceCoverage,omitempty"`
}

// InsuranceCoverage represents insurance coverage status
type InsuranceCoverage struct {
	MedicalInsurance    bool `json:"medicalInsurance"`
	LongTermCareInsurance bool `json:"longTermCareInsurance"`
	PrivateInsurance    bool `json:"privateInsurance"`
}

// SocialSupport represents social support systems
type SocialSupport struct {
	CommunityServices    []string `json:"communityServices,omitempty"` // meal_delivery, daycare, respite_care
	NeighborSupport      string   `json:"neighborSupport"`             // none, minimal, moderate, strong
	ReligiousAffiliation string   `json:"religiousAffiliation,omitempty"`
}

// PatientSocialProfileCreateRequest represents the request to create a social profile
type PatientSocialProfileCreateRequest struct {
	PatientID   string                `json:"patient_id" validate:"required"`
	Content     SocialProfileContent  `json:"content" validate:"required"`
	ValidFrom   time.Time             `json:"valid_from" validate:"required"`
	AssessedBy  string                `json:"assessed_by,omitempty"`
	AssessedAt  *time.Time            `json:"assessed_at,omitempty"`
	AssessmentNotes string             `json:"assessment_notes,omitempty"`
}

// PatientSocialProfileUpdateRequest represents the request to update a social profile
type PatientSocialProfileUpdateRequest struct {
	Content         *SocialProfileContent `json:"content,omitempty"`
	ValidTo         *time.Time            `json:"valid_to,omitempty"`
	AssessmentNotes *string               `json:"assessment_notes,omitempty"`
}

// Helper method to parse content JSONB
func (p *PatientSocialProfile) GetContent() (*SocialProfileContent, error) {
	var content SocialProfileContent
	if len(p.Content) == 0 {
		return &content, nil
	}
	err := json.Unmarshal(p.Content, &content)
	return &content, err
}

// Helper method to check if profile is current
func (p *PatientSocialProfile) IsCurrent() bool {
	now := time.Now()

	// Check if deleted
	if p.Deleted {
		return false
	}

	// Check validity period
	if p.ValidFrom.Valid && p.ValidFrom.Time.After(now) {
		return false
	}

	if p.ValidTo.Valid && p.ValidTo.Time.Before(now) {
		return false
	}

	return true
}
