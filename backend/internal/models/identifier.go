package models

import (
	"database/sql"
	"time"
)

// IdentifierType represents the type of patient identifier
type IdentifierType string

const (
	IdentifierTypeMyNumber        IdentifierType = "my_number"         // マイナンバー (requires encryption)
	IdentifierTypeInsuranceID     IdentifierType = "insurance_id"      // 健康保険証番号
	IdentifierTypeCareInsuranceID IdentifierType = "care_insurance_id" // 介護保険証番号
	IdentifierTypeMRN             IdentifierType = "mrn"               // Medical Record Number
	IdentifierTypeOther           IdentifierType = "other"             // その他
)

// VerificationStatus represents the verification status of an identifier
type VerificationStatus string

const (
	VerificationStatusVerified   VerificationStatus = "verified"
	VerificationStatusUnverified VerificationStatus = "unverified"
	VerificationStatusExpired    VerificationStatus = "expired"
	VerificationStatusInvalid    VerificationStatus = "invalid"
)

// PatientIdentifier represents a patient identification number
type PatientIdentifier struct {
	IdentifierID   string `json:"identifier_id" spanner:"identifier_id"`
	PatientID      string `json:"patient_id" spanner:"patient_id"`
	IdentifierType string `json:"identifier_type" spanner:"identifier_type"`

	// For my_number: stores encrypted ciphertext (base64-encoded)
	// For other types: stores plaintext value
	IdentifierValue string `json:"identifier_value" spanner:"identifier_value"`

	IsPrimary bool `json:"is_primary" spanner:"is_primary"`

	// Validity Period
	ValidFrom sql.NullTime `json:"valid_from,omitempty" spanner:"valid_from"`
	ValidTo   sql.NullTime `json:"valid_to,omitempty" spanner:"valid_to"`

	// Issuer Information (for insurance cards)
	IssuerName string `json:"issuer_name,omitempty" spanner:"issuer_name"`
	IssuerCode string `json:"issuer_code,omitempty" spanner:"issuer_code"`

	// Verification Status
	VerificationStatus string       `json:"verification_status" spanner:"verification_status"`
	VerifiedAt         sql.NullTime `json:"verified_at,omitempty" spanner:"verified_at"`
	VerifiedBy         string       `json:"verified_by,omitempty" spanner:"verified_by"`

	// Audit Timestamps
	CreatedAt time.Time `json:"created_at" spanner:"created_at"`
	CreatedBy string    `json:"created_by,omitempty" spanner:"created_by"`
	UpdatedAt time.Time `json:"updated_at" spanner:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty" spanner:"updated_by"`

	// Soft Delete
	Deleted   bool         `json:"deleted" spanner:"deleted"`
	DeletedAt sql.NullTime `json:"deleted_at,omitempty" spanner:"deleted_at"`
}

// PatientIdentifierCreateRequest represents the request to add a new identifier
type PatientIdentifierCreateRequest struct {
	PatientID      string    `json:"patient_id" validate:"required"`
	IdentifierType string    `json:"identifier_type" validate:"required,oneof=my_number insurance_id care_insurance_id mrn other"`
	IdentifierValue string    `json:"identifier_value" validate:"required"`
	IsPrimary      bool      `json:"is_primary"`
	ValidFrom      *time.Time `json:"valid_from,omitempty"`
	ValidTo        *time.Time `json:"valid_to,omitempty"`
	IssuerName     string    `json:"issuer_name,omitempty"`
	IssuerCode     string    `json:"issuer_code,omitempty"`
}

// PatientIdentifierUpdateRequest represents the request to update an identifier
type PatientIdentifierUpdateRequest struct {
	IdentifierValue    *string    `json:"identifier_value,omitempty"`
	IsPrimary          *bool      `json:"is_primary,omitempty"`
	ValidFrom          *time.Time `json:"valid_from,omitempty"`
	ValidTo            *time.Time `json:"valid_to,omitempty"`
	IssuerName         *string    `json:"issuer_name,omitempty"`
	IssuerCode         *string    `json:"issuer_code,omitempty"`
	VerificationStatus *string    `json:"verification_status,omitempty"`
}

// IsMyNumber checks if the identifier is a mynumber (requires encryption)
func (i *PatientIdentifier) IsMyNumber() bool {
	return i.IdentifierType == string(IdentifierTypeMyNumber)
}

// IsActive checks if the identifier is currently valid
func (i *PatientIdentifier) IsActive() bool {
	now := time.Now()

	// Check if deleted
	if i.Deleted {
		return false
	}

	// Check validity period
	if i.ValidFrom.Valid && i.ValidFrom.Time.After(now) {
		return false
	}

	if i.ValidTo.Valid && i.ValidTo.Time.Before(now) {
		return false
	}

	return true
}

// PatientIdentifierListResponse represents a list of identifiers for a patient
type PatientIdentifierListResponse struct {
	Identifiers []PatientIdentifier `json:"identifiers"`
	Total       int                 `json:"total"`
}
