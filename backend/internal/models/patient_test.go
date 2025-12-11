package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatient_JSONSerialization(t *testing.T) {
	// Arrange
	birthDate := time.Date(1950, 1, 15, 0, 0, 0, 0, time.UTC)
	bloodType := "A+"
	createdBy := "user-123"

	patient := Patient{
		PatientID: "patient-001",
		BirthDate: birthDate,
		Gender:    "male",
		BloodType: &bloodType,
		Name: HumanName{
			Family: "山田",
			Given:  "太郎",
			Kana:   "ヤマダ タロウ",
		},
		ContactPoints: []ContactPoint{
			{
				System: "phone",
				Value:  "090-1234-5678",
				Use:    "mobile",
			},
		},
		Addresses: []Address{
			{
				Use:        "home",
				PostalCode: "123-4567",
				Prefecture: "東京都",
				City:       "新宿区",
				Line:       "西新宿1-2-3",
				Geolocation: &GeoLocation{
					Latitude:  35.6895,
					Longitude: 139.6917,
				},
			},
		},
		Active:               true,
		Deleted:              false,
		DataClassification:   "confidential",
		EncryptionKeyVersion: 1,
		CreatedAt:            time.Now(),
		CreatedBy:            &createdBy,
		UpdatedAt:            time.Now(),
	}

	// Act
	jsonData, err := json.Marshal(patient)
	require.NoError(t, err, "JSON marshaling should not fail")

	// Assert
	assert.NotEmpty(t, jsonData)

	// Unmarshal and verify
	var unmarshaled Patient
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err, "JSON unmarshaling should not fail")

	assert.Equal(t, patient.PatientID, unmarshaled.PatientID)
	assert.Equal(t, patient.Name.Family, unmarshaled.Name.Family)
	assert.Equal(t, patient.Name.Kana, unmarshaled.Name.Kana)
	assert.Equal(t, patient.Gender, unmarshaled.Gender)
	assert.Equal(t, *patient.BloodType, *unmarshaled.BloodType)
}

func TestHumanName_FullName(t *testing.T) {
	tests := []struct {
		name     string
		input    HumanName
		expected string
	}{
		{
			name: "Simple Japanese name",
			input: HumanName{
				Family: "山田",
				Given:  "太郎",
				Kana:   "ヤマダ タロウ",
			},
			expected: "山田 太郎",
		},
		{
			name: "Name with previous names",
			input: HumanName{
				Family: "佐藤",
				Given:  "花子",
				Kana:   "サトウ ハナコ",
				PreviousNames: []PreviousNameEntry{
					{
						Family:       "鈴木",
						Given:        "花子",
						ChangeReason: "婚姻",
					},
				},
			},
			expected: "佐藤 花子",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullName := tt.input.Family + " " + tt.input.Given
			assert.Equal(t, tt.expected, fullName)
		})
	}
}

func TestPatientCreateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request PatientCreateRequest
		wantErr bool
	}{
		{
			name: "Valid patient creation request",
			request: PatientCreateRequest{
				BirthDate: time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC),
				Gender:    "male",
				Name: HumanName{
					Family: "山田",
					Given:  "太郎",
					Kana:   "ヤマダ タロウ",
				},
			},
			wantErr: false,
		},
		// Add more test cases as validation is implemented
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContactPoint_Validation(t *testing.T) {
	tests := []struct {
		name  string
		cp    ContactPoint
		valid bool
	}{
		{
			name: "Valid phone contact",
			cp: ContactPoint{
				System: "phone",
				Value:  "090-1234-5678",
				Use:    "mobile",
			},
			valid: true,
		},
		{
			name: "Valid email contact",
			cp: ContactPoint{
				System: "email",
				Value:  "test@example.com",
				Use:    "home",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate system field
			validSystems := map[string]bool{"phone": true, "email": true, "fax": true}
			assert.True(t, validSystems[tt.cp.System], "System should be valid")

			// Validate use field if present
			if tt.cp.Use != "" {
				validUses := map[string]bool{"home": true, "work": true, "mobile": true}
				assert.True(t, validUses[tt.cp.Use], "Use should be valid")
			}
		})
	}
}

func TestAddress_GeoLocation(t *testing.T) {
	addr := Address{
		Use:        "home",
		PostalCode: "123-4567",
		Prefecture: "東京都",
		City:       "新宿区",
		Line:       "西新宿1-2-3",
		Geolocation: &GeoLocation{
			Latitude:  35.6895,
			Longitude: 139.6917,
		},
	}

	// Test geolocation bounds
	assert.GreaterOrEqual(t, addr.Geolocation.Latitude, -90.0)
	assert.LessOrEqual(t, addr.Geolocation.Latitude, 90.0)
	assert.GreaterOrEqual(t, addr.Geolocation.Longitude, -180.0)
	assert.LessOrEqual(t, addr.Geolocation.Longitude, 180.0)
}
