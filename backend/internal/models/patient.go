package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Patient represents the enhanced patient model aligned with DATABASE_REQUIREMENTS.md
type Patient struct {
	PatientID string `json:"patient_id" spanner:"patient_id"`

	// Basic Demographics
	BirthDate sql.NullTime `json:"birth_date" spanner:"birth_date"`
	Gender    string       `json:"gender" spanner:"gender"`
	BloodType string       `json:"blood_type,omitempty" spanner:"blood_type"`

	// JSONB Fields (stored as JSON in database)
	// Name includes previousNames array internally
	Name           json.RawMessage `json:"name" spanner:"name"`
	ContactPoints  json.RawMessage `json:"contact_points,omitempty" spanner:"contact_points"`
	Addresses      json.RawMessage `json:"addresses,omitempty" spanner:"addresses"`
	ConsentRecords json.RawMessage `json:"consent_records,omitempty" spanner:"consent_records"`

	// Photo
	PhotoURL string `json:"photo_url,omitempty" spanner:"photo_url"`

	// Generated Columns (read-only from database)
	FullName     string `json:"full_name,omitempty" spanner:"full_name"`
	FullNameKana string `json:"full_name_kana,omitempty" spanner:"full_name_kana"`
	PrimaryPhone string `json:"primary_phone,omitempty" spanner:"primary_phone"`

	// System Management
	Active bool `json:"active" spanner:"active"`

	// Soft Delete
	IsDeleted bool         `json:"is_deleted" spanner:"is_deleted"`
	DeletedAt sql.NullTime `json:"deleted_at,omitempty" spanner:"deleted_at"`

	// Audit Timestamps
	CreatedAt time.Time `json:"created_at" spanner:"created_at"`
	UpdatedAt time.Time `json:"updated_at" spanner:"updated_at"`

	// Security
	DataClassification   string `json:"data_classification" spanner:"data_classification"`
	EncryptionKeyVersion int    `json:"encryption_key_version" spanner:"encryption_key_version"`
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

	// Name (will be stored as JSONB object)
	Name Name `json:"name" validate:"required"`

	// Contact Points (JSONB array)
	ContactPoints []ContactPoint `json:"contact_points,omitempty"`

	// Addresses (JSONB array)
	Addresses []Address `json:"addresses,omitempty"`

	// Photo
	PhotoURL string `json:"photo_url,omitempty"`

	// Consent Records (JSONB)
	ConsentRecords json.RawMessage `json:"consent_records,omitempty"`
}

// PatientUpdateRequest represents the request payload for updating a patient
type PatientUpdateRequest struct {
	// Basic Demographics
	BirthDate *string `json:"birth_date,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	BloodType *string `json:"blood_type,omitempty"`

	// Name (will update the entire name object)
	Name *Name `json:"name,omitempty"`

	// Update contact points (replaces entire array)
	ContactPoints *[]ContactPoint `json:"contact_points,omitempty"`

	// Update addresses (replaces entire array)
	Addresses *[]Address `json:"addresses,omitempty"`

	// Photo
	PhotoURL *string `json:"photo_url,omitempty"`

	// Consent Records
	ConsentRecords *json.RawMessage `json:"consent_records,omitempty"`

	// Active status
	Active *bool `json:"active,omitempty"`
}

// PatientListResponse represents paginated patient list
type PatientListResponse struct {
	Patients   []Patient `json:"patients"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}

// Helper method to convert Name to structured format
func (p *Patient) GetName() (*Name, error) {
	var name Name
	if len(p.Name) == 0 {
		return nil, nil
	}
	err := json.Unmarshal(p.Name, &name)
	return &name, err
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
