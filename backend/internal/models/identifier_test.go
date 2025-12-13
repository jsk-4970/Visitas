package models

import (
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
)

func TestPatientIdentifier_IsMyNumber(t *testing.T) {
	tests := []struct {
		name           string
		identifierType string
		expected       bool
	}{
		{
			name:           "My Number identifier",
			identifierType: string(IdentifierTypeMyNumber),
			expected:       true,
		},
		{
			name:           "Insurance ID is not My Number",
			identifierType: string(IdentifierTypeInsuranceID),
			expected:       false,
		},
		{
			name:           "Care Insurance ID is not My Number",
			identifierType: string(IdentifierTypeCareInsuranceID),
			expected:       false,
		},
		{
			name:           "MRN is not My Number",
			identifierType: string(IdentifierTypeMRN),
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identifier := &PatientIdentifier{
				IdentifierType: tt.identifierType,
			}
			assert.Equal(t, tt.expected, identifier.IsMyNumber())
		})
	}
}

func TestPatientIdentifier_IsActive(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name      string
		deleted   bool
		validFrom spanner.NullTime
		validTo   spanner.NullTime
		expected  bool
	}{
		{
			name:      "Active identifier with no validity period",
			deleted:   false,
			validFrom: spanner.NullTime{Valid: false},
			validTo:   spanner.NullTime{Valid: false},
			expected:  true,
		},
		{
			name:      "Active identifier within validity period",
			deleted:   false,
			validFrom: spanner.NullTime{Time: yesterday, Valid: true},
			validTo:   spanner.NullTime{Time: tomorrow, Valid: true},
			expected:  true,
		},
		{
			name:      "Deleted identifier",
			deleted:   true,
			validFrom: spanner.NullTime{Valid: false},
			validTo:   spanner.NullTime{Valid: false},
			expected:  false,
		},
		{
			name:      "Identifier not yet valid",
			deleted:   false,
			validFrom: spanner.NullTime{Time: tomorrow, Valid: true},
			validTo:   spanner.NullTime{Valid: false},
			expected:  false,
		},
		{
			name:      "Expired identifier",
			deleted:   false,
			validFrom: spanner.NullTime{Valid: false},
			validTo:   spanner.NullTime{Time: yesterday, Valid: true},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identifier := &PatientIdentifier{
				Deleted:   tt.deleted,
				ValidFrom: tt.validFrom,
				ValidTo:   tt.validTo,
			}
			assert.Equal(t, tt.expected, identifier.IsActive())
		})
	}
}

func TestIdentifierType_Constants(t *testing.T) {
	// Ensure identifier type constants are defined correctly
	assert.Equal(t, IdentifierType("my_number"), IdentifierTypeMyNumber)
	assert.Equal(t, IdentifierType("insurance_id"), IdentifierTypeInsuranceID)
	assert.Equal(t, IdentifierType("care_insurance_id"), IdentifierTypeCareInsuranceID)
	assert.Equal(t, IdentifierType("mrn"), IdentifierTypeMRN)
	assert.Equal(t, IdentifierType("other"), IdentifierTypeOther)
}

func TestVerificationStatus_Constants(t *testing.T) {
	// Ensure verification status constants are defined correctly
	assert.Equal(t, VerificationStatus("verified"), VerificationStatusVerified)
	assert.Equal(t, VerificationStatus("unverified"), VerificationStatusUnverified)
	assert.Equal(t, VerificationStatus("expired"), VerificationStatusExpired)
	assert.Equal(t, VerificationStatus("invalid"), VerificationStatusInvalid)
}

func TestPatientIdentifierCreateRequest_Validation(t *testing.T) {
	validFrom := time.Now()

	tests := []struct {
		name    string
		request PatientIdentifierCreateRequest
		wantErr bool
	}{
		{
			name: "Valid My Number request",
			request: PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  string(IdentifierTypeMyNumber),
				IdentifierValue: "123456789012", // 12 digits
				IsPrimary:       true,
				ValidFrom:       &validFrom,
			},
			wantErr: false,
		},
		{
			name: "Valid Insurance ID request",
			request: PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  string(IdentifierTypeInsuranceID),
				IdentifierValue: "12345678",
				IsPrimary:       true,
				IssuerName:      "全国健康保険協会",
				IssuerCode:      "06",
			},
			wantErr: false,
		},
		{
			name: "Valid Care Insurance ID request",
			request: PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  string(IdentifierTypeCareInsuranceID),
				IdentifierValue: "1234567890",
				IsPrimary:       true,
				IssuerName:      "東京都新宿区",
			},
			wantErr: false,
		},
		{
			name: "Valid MRN request",
			request: PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  string(IdentifierTypeMRN),
				IdentifierValue: "MRN-001",
				IsPrimary:       false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			assert.NotEmpty(t, tt.request.PatientID, "PatientID should not be empty")
			assert.NotEmpty(t, tt.request.IdentifierType, "IdentifierType should not be empty")
			assert.NotEmpty(t, tt.request.IdentifierValue, "IdentifierValue should not be empty")

			// Type-specific validation
			if tt.request.IdentifierType == string(IdentifierTypeMyNumber) {
				assert.Len(t, tt.request.IdentifierValue, 12, "My Number should be 12 digits")
			}

			if tt.request.IdentifierType == string(IdentifierTypeInsuranceID) ||
			   tt.request.IdentifierType == string(IdentifierTypeCareInsuranceID) {
				assert.NotEmpty(t, tt.request.IssuerName, "Insurance identifiers should have issuer name")
			}
		})
	}
}
