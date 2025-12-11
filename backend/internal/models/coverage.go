package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// InsuranceType represents the type of insurance coverage
type InsuranceType string

const (
	InsuranceTypeMedical       InsuranceType = "medical"        // 医療保険
	InsuranceTypeLongTermCare  InsuranceType = "long_term_care" // 介護保険
	InsuranceTypePublicExpense InsuranceType = "public_expense" // 公費負担医療
)

// CoverageStatus represents the status of an insurance coverage
type CoverageStatus string

const (
	CoverageStatusActive     CoverageStatus = "active"
	CoverageStatusExpired    CoverageStatus = "expired"
	CoverageStatusSuspended  CoverageStatus = "suspended"
	CoverageStatusTerminated CoverageStatus = "terminated"
)

// PatientCoverage represents insurance and coverage information
type PatientCoverage struct {
	CoverageID     string          `json:"coverage_id" spanner:"coverage_id"`
	PatientID      string          `json:"patient_id" spanner:"patient_id"`
	InsuranceType  string          `json:"insurance_type" spanner:"insurance_type"`
	Details        json.RawMessage `json:"details" spanner:"details"` // JSONB

	// Generated Columns
	CareLevelCode string `json:"care_level_code,omitempty" spanner:"care_level_code"` // For long_term_care
	CopayRate     int    `json:"copay_rate" spanner:"copay_rate"`                      // Percentage

	// Validity Period
	ValidFrom sql.NullTime `json:"valid_from" spanner:"valid_from"`
	ValidTo   sql.NullTime `json:"valid_to,omitempty" spanner:"valid_to"`

	// Status
	Status   string `json:"status" spanner:"status"`
	Priority int    `json:"priority" spanner:"priority"` // Lower number = higher priority

	// Verification
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

// MedicalInsuranceDetails represents medical insurance details
type MedicalInsuranceDetails struct {
	InsurerName              string `json:"insurerName"`
	InsurerNumber            string `json:"insurerNumber"`
	CertificateNumber        string `json:"certificateNumber"`
	CertificateSymbol        string `json:"certificateSymbol,omitempty"`
	CopayRate                int    `json:"copayRate"` // Percentage (typically 30, 20, or 10)
	InsuredPersonCategory    string `json:"insuredPersonCategory"` // 本人 or 家族
	EmployerName             string `json:"employerName,omitempty"`
}

// LongTermCareInsuranceDetails represents long-term care insurance details
type LongTermCareInsuranceDetails struct {
	InsurerName             string    `json:"insurerName"`
	InsurerNumber           string    `json:"insurerNumber"`
	CertificateNumber       string    `json:"certificateNumber"`
	CareLevelCode           string    `json:"careLevelCode"` // 要支援1-2, 要介護1-5
	CareLevelAssessedAt     time.Time `json:"careLevelAssessedAt"`
	CopayRate               int       `json:"copayRate"` // Typically 10%
	MonthlyServiceLimit     int       `json:"monthlyServiceLimit"` // Yen
	CertificationValidFrom  string    `json:"certificationValidFrom"` // YYYY-MM-DD
	CertificationValidTo    string    `json:"certificationValidTo"`   // YYYY-MM-DD
}

// PublicExpenseDetails represents public assistance details
type PublicExpenseDetails struct {
	ProgramName        string `json:"programName"` // e.g., 特定疾患医療費助成制度, 生活保護
	RecipientNumber    string `json:"recipientNumber"`
	IssuingAuthority   string `json:"issuingAuthority"`
	CopayRate          int    `json:"copayRate"` // Often 0%
	ExemptionReason    string `json:"exemptionReason,omitempty"`
}

// PatientCoverageCreateRequest represents the request to add insurance coverage
type PatientCoverageCreateRequest struct {
	PatientID     string          `json:"patient_id" validate:"required"`
	InsuranceType string          `json:"insurance_type" validate:"required,oneof=medical long_term_care public_expense"`
	Details       json.RawMessage `json:"details" validate:"required"`
	ValidFrom     time.Time       `json:"valid_from" validate:"required"`
	ValidTo       *time.Time      `json:"valid_to,omitempty"`
	Status        string          `json:"status" validate:"required,oneof=active expired suspended terminated"`
	Priority      int             `json:"priority" validate:"required,min=1"`
}

// PatientCoverageUpdateRequest represents the request to update insurance coverage
type PatientCoverageUpdateRequest struct {
	Details                *json.RawMessage `json:"details,omitempty"`
	ValidTo                *time.Time       `json:"valid_to,omitempty"`
	Status                 *string          `json:"status,omitempty"`
	Priority               *int             `json:"priority,omitempty"`
	VerificationStatus     *string          `json:"verification_status,omitempty"`
}

// Helper methods

// GetMedicalInsuranceDetails parses medical insurance details
func (c *PatientCoverage) GetMedicalInsuranceDetails() (*MedicalInsuranceDetails, error) {
	if c.InsuranceType != string(InsuranceTypeMedical) {
		return nil, nil
	}
	var details MedicalInsuranceDetails
	err := json.Unmarshal(c.Details, &details)
	return &details, err
}

// GetLongTermCareDetails parses long-term care insurance details
func (c *PatientCoverage) GetLongTermCareDetails() (*LongTermCareInsuranceDetails, error) {
	if c.InsuranceType != string(InsuranceTypeLongTermCare) {
		return nil, nil
	}
	var details LongTermCareInsuranceDetails
	err := json.Unmarshal(c.Details, &details)
	return &details, err
}

// GetPublicExpenseDetails parses public expense details
func (c *PatientCoverage) GetPublicExpenseDetails() (*PublicExpenseDetails, error) {
	if c.InsuranceType != string(InsuranceTypePublicExpense) {
		return nil, nil
	}
	var details PublicExpenseDetails
	err := json.Unmarshal(c.Details, &details)
	return &details, err
}

// IsActive checks if the coverage is currently active
func (c *PatientCoverage) IsActive() bool {
	now := time.Now()

	// Check if deleted
	if c.Deleted {
		return false
	}

	// Check status
	if c.Status != string(CoverageStatusActive) {
		return false
	}

	// Check validity period
	if c.ValidFrom.Valid && c.ValidFrom.Time.After(now) {
		return false
	}

	if c.ValidTo.Valid && c.ValidTo.Time.Before(now) {
		return false
	}

	return true
}

// IsExpiringSoon checks if the coverage will expire within the given days
func (c *PatientCoverage) IsExpiringSoon(days int) bool {
	if !c.ValidTo.Valid {
		return false
	}

	now := time.Now()
	expirationThreshold := now.AddDate(0, 0, days)

	return c.ValidTo.Time.After(now) && c.ValidTo.Time.Before(expirationThreshold)
}
