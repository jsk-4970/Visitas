package models

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/spanner"
)

// Patient represents the enhanced patient model aligned with DATABASE_REQUIREMENTS.md
type Patient struct {
	PatientID string `json:"patient_id" spanner:"patient_id"`

	// Basic Demographics
	BirthDate spanner.NullTime `json:"birth_date" spanner:"birth_date"`
	Gender    string           `json:"gender" spanner:"gender"`
	BloodType string       `json:"blood_type,omitempty" spanner:"blood_type"`

	// JSONB Fields (stored as JSON in database)
	NameHistory    json.RawMessage `json:"name_history" spanner:"name_history"`
	ContactPoints  json.RawMessage `json:"contact_points,omitempty" spanner:"contact_points"`
	Addresses      json.RawMessage `json:"addresses,omitempty" spanner:"addresses"`
	ConsentDetails json.RawMessage `json:"consent_details,omitempty" spanner:"consent_details"`

	// Generated Columns (read-only from database)
	CurrentFamilyName string `json:"current_family_name,omitempty" spanner:"current_family_name"`
	CurrentGivenName  string `json:"current_given_name,omitempty" spanner:"current_given_name"`
	PrimaryPhone      string `json:"primary_phone,omitempty" spanner:"primary_phone"`
	CurrentPrefecture string `json:"current_prefecture,omitempty" spanner:"current_prefecture"`
	CurrentCity       string `json:"current_city,omitempty" spanner:"current_city"`

	// Consent Management
	ConsentStatus      string        `json:"consent_status" spanner:"consent_status"`
	ConsentObtainedAt  *time.Time    `json:"consent_obtained_at,omitempty" spanner:"consent_obtained_at"`
	ConsentWithdrawnAt *time.Time    `json:"consent_withdrawn_at,omitempty" spanner:"consent_withdrawn_at"`

	// Soft Delete
	Deleted       bool       `json:"deleted" spanner:"deleted"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" spanner:"deleted_at"`
	DeletedReason string     `json:"deleted_reason,omitempty" spanner:"deleted_reason"`

	// Audit Timestamps
	CreatedAt time.Time `json:"created_at" spanner:"created_at"`
	CreatedBy string    `json:"created_by" spanner:"created_by"`
	UpdatedAt time.Time `json:"updated_at" spanner:"updated_at"`
	UpdatedBy string    `json:"updated_by" spanner:"updated_by"`
}

// NameRecord represents a single name entry (used in name_history array)
type NameRecord struct {
	Family       string     `json:"family"`
	Given        string     `json:"given"`
	Kana         string     `json:"kana,omitempty"`
	ValidFrom    time.Time  `json:"valid_from"`
	ValidTo      *time.Time `json:"valid_to,omitempty"`
	ChangeReason string     `json:"change_reason,omitempty"` // 婚姻, 離婚, etc.
}

// Name represents patient name with history
// Aligned with DATABASE_REQUIREMENTS.md specification
type Name struct {
	Use           string         `json:"use"` // official, maiden, nickname
	Family        string         `json:"family"`
	Given         string         `json:"given"`
	Kana          string         `json:"kana,omitempty"`
	PreviousNames []PreviousName `json:"previousNames,omitempty"`
}

// PreviousName represents a historical name entry
type PreviousName struct {
	Family       string  `json:"family"`
	Given        string  `json:"given"`
	Period       Period  `json:"period"`
	ChangeReason string  `json:"changeReason,omitempty"` // 婚姻, 離婚, etc.
}

// Period represents a time period
type Period struct {
	Start string  `json:"start,omitempty"` // ISO 8601 date
	End   string  `json:"end,omitempty"`   // ISO 8601 date
}

// ContactPoint represents a contact method
type ContactPoint struct {
	System     string     `json:"system"` // phone, email, fax
	Value      string     `json:"value"`
	Use        string     `json:"use,omitempty"` // home, work, mobile
	Rank       int        `json:"rank"`           // 1 = primary
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

// Address represents a patient address with geolocation
type Address struct {
	Use                string       `json:"use"` // home, billing
	PostalCode         string       `json:"postal_code,omitempty"`
	Prefecture         string       `json:"prefecture"`
	City               string       `json:"city"`
	Line               string       `json:"line"` // Street address
	Building           string       `json:"building,omitempty"`
	Geolocation        *Geolocation `json:"geolocation,omitempty"`
	AccessInstructions string       `json:"access_instructions,omitempty"`
	ValidFrom          time.Time    `json:"valid_from"`
	ValidTo            *time.Time   `json:"valid_to,omitempty"`
}

// Geolocation represents latitude/longitude coordinates
type Geolocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// PatientCreateRequest represents the request payload for creating a patient
type PatientCreateRequest struct {
	// Basic Demographics
	BirthDate string `json:"birth_date" validate:"required"` // YYYY-MM-DD
	Gender    string `json:"gender" validate:"required"`
	BloodType string `json:"blood_type,omitempty"`

	// Name (will be stored as name_history JSONB array)
	Name NameRecord `json:"name" validate:"required"`

	// Contact Points (JSONB array)
	ContactPoints []ContactPoint `json:"contact_points,omitempty"`

	// Addresses (JSONB array)
	Addresses []Address `json:"addresses,omitempty"`

	// Consent
	ConsentStatus     string     `json:"consent_status"`
	ConsentObtainedAt *time.Time `json:"consent_obtained_at,omitempty"`
}

// PatientUpdateRequest represents the request payload for updating a patient
type PatientUpdateRequest struct {
	// Basic Demographics
	BirthDate *string `json:"birth_date,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	BloodType *string `json:"blood_type,omitempty"`

	// Add a new name to name_history
	AddName *NameRecord `json:"add_name,omitempty"`

	// Update contact points (replaces entire array)
	ContactPoints *[]ContactPoint `json:"contact_points,omitempty"`
	// Add a single contact point
	AddContactPoint *ContactPoint `json:"add_contact_point,omitempty"`

	// Update addresses (replaces entire array)
	Addresses *[]Address `json:"addresses,omitempty"`
	// Add a single address
	AddAddress *Address `json:"add_address,omitempty"`

	// Consent
	ConsentStatus      *string    `json:"consent_status,omitempty"`
	ConsentObtainedAt  *time.Time `json:"consent_obtained_at,omitempty"`
	ConsentWithdrawnAt *time.Time `json:"consent_withdrawn_at,omitempty"`
}

// PatientListResponse represents paginated patient list
type PatientListResponse struct {
	Patients   []Patient `json:"patients"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}

// Helper method to convert NameHistory to structured format
func (p *Patient) GetNameHistory() ([]NameRecord, error) {
	var nameHistory []NameRecord
	if len(p.NameHistory) == 0 {
		return nameHistory, nil
	}
	err := json.Unmarshal(p.NameHistory, &nameHistory)
	return nameHistory, err
}

// Helper method to convert ContactPoints to structured format
func (p *Patient) GetContactPoints() ([]ContactPoint, error) {
	var contacts []ContactPoint
	if len(p.ContactPoints) == 0 {
		return contacts, nil
	}
	err := json.Unmarshal(p.ContactPoints, &contacts)
	return contacts, err
}

// Helper method to convert Addresses to structured format
func (p *Patient) GetAddresses() ([]Address, error) {
	var addresses []Address
	if len(p.Addresses) == 0 {
		return addresses, nil
	}
	err := json.Unmarshal(p.Addresses, &addresses)
	return addresses, err
}
