package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
)

// Mock repositories
type MockIdentifierRepository struct {
	mock.Mock
}

func (m *MockIdentifierRepository) CreateIdentifier(ctx context.Context, req *models.PatientIdentifierCreateRequest, createdBy string) (*models.PatientIdentifier, error) {
	args := m.Called(ctx, req, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PatientIdentifier), args.Error(1)
}

func (m *MockIdentifierRepository) GetIdentifierByID(ctx context.Context, identifierID string, decrypt bool) (*models.PatientIdentifier, error) {
	args := m.Called(ctx, identifierID, decrypt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PatientIdentifier), args.Error(1)
}

func (m *MockIdentifierRepository) GetIdentifiersByPatientID(ctx context.Context, patientID string, decrypt bool) ([]*models.PatientIdentifier, error) {
	args := m.Called(ctx, patientID, decrypt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PatientIdentifier), args.Error(1)
}

func (m *MockIdentifierRepository) UpdateIdentifier(ctx context.Context, identifierID string, req *models.PatientIdentifierUpdateRequest, updatedBy string) (*models.PatientIdentifier, error) {
	args := m.Called(ctx, identifierID, req, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PatientIdentifier), args.Error(1)
}

func (m *MockIdentifierRepository) DeleteIdentifier(ctx context.Context, identifierID string) error {
	args := m.Called(ctx, identifierID)
	return args.Error(0)
}

func (m *MockIdentifierRepository) GetPrimaryIdentifier(ctx context.Context, patientID string, identifierType models.IdentifierType, decrypt bool) (*models.PatientIdentifier, error) {
	args := m.Called(ctx, patientID, identifierType, decrypt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PatientIdentifier), args.Error(1)
}

type MockPatientRepository struct {
	mock.Mock
}

func (m *MockPatientRepository) CheckStaffAccess(ctx context.Context, staffID, patientID string) (bool, error) {
	args := m.Called(ctx, staffID, patientID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPatientRepository) GetPatientByID(ctx context.Context, patientID string) (*models.Patient, error) {
	args := m.Called(ctx, patientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Patient), args.Error(1)
}

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) LogMyNumberAccess(ctx context.Context, patientID, identifierID, action, userID string) error {
	args := m.Called(ctx, patientID, identifierID, action, userID)
	return args.Error(0)
}

func (m *MockAuditRepository) LogAccess(ctx context.Context, log *repository.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// Helper function to create test service
func setupIdentifierServiceTest() (*IdentifierService, *MockIdentifierRepository, *MockPatientRepository, *MockAuditRepository) {
	mockIdentifierRepo := new(MockIdentifierRepository)
	mockPatientRepo := new(MockPatientRepository)
	mockAuditRepo := new(MockAuditRepository)

	service := NewIdentifierService(mockIdentifierRepo, mockPatientRepo, mockAuditRepo)

	return service, mockIdentifierRepo, mockPatientRepo, mockAuditRepo
}

// Tests for CreateIdentifier
func TestIdentifierService_CreateIdentifier_Success(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, mockAuditRepo := setupIdentifierServiceTest()
	ctx := context.Background()

	req := &models.PatientIdentifierCreateRequest{
		PatientID:       "patient-123",
		IdentifierType:  string(models.IdentifierTypeMyNumber),
		IdentifierValue: "123456789012",
		IsPrimary:       true,
	}
	createdBy := "staff-456"

	expectedIdentifier := &models.PatientIdentifier{
		IdentifierID:    "id-789",
		PatientID:       req.PatientID,
		IdentifierType:  req.IdentifierType,
		IdentifierValue: req.IdentifierValue,
		IsPrimary:       req.IsPrimary,
		CreatedAt:       time.Now(),
	}

	// Setup mocks
	mockPatientRepo.On("CheckStaffAccess", ctx, createdBy, req.PatientID).Return(true, nil)
	mockIdentifierRepo.On("CreateIdentifier", ctx, req, createdBy).Return(expectedIdentifier, nil)
	mockAuditRepo.On("LogMyNumberAccess", ctx, req.PatientID, expectedIdentifier.IdentifierID, "create", createdBy).Return(nil)

	// Execute
	result, err := service.CreateIdentifier(ctx, req, createdBy)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedIdentifier.IdentifierID, result.IdentifierID)
	assert.Equal(t, expectedIdentifier.PatientID, result.PatientID)
	mockPatientRepo.AssertExpectations(t)
	mockIdentifierRepo.AssertExpectations(t)
	mockAuditRepo.AssertExpectations(t)
}

func TestIdentifierService_CreateIdentifier_AccessDenied(t *testing.T) {
	service, _, mockPatientRepo, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	req := &models.PatientIdentifierCreateRequest{
		PatientID:       "patient-123",
		IdentifierType:  string(models.IdentifierTypeMyNumber),
		IdentifierValue: "123456789012",
		IsPrimary:       true,
	}
	createdBy := "staff-456"

	// Setup mock - no access
	mockPatientRepo.On("CheckStaffAccess", ctx, createdBy, req.PatientID).Return(false, nil)

	// Execute
	result, err := service.CreateIdentifier(ctx, req, createdBy)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "access denied")
	mockPatientRepo.AssertExpectations(t)
}

func TestIdentifierService_CreateIdentifier_ValidationError_EmptyPatientID(t *testing.T) {
	service, _, _, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	req := &models.PatientIdentifierCreateRequest{
		PatientID:       "", // Empty patient ID
		IdentifierType:  string(models.IdentifierTypeMyNumber),
		IdentifierValue: "123456789012",
	}
	createdBy := "staff-456"

	// Execute
	result, err := service.CreateIdentifier(ctx, req, createdBy)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "patient_id is required")
}

func TestIdentifierService_CreateIdentifier_ValidationError_InvalidMyNumber(t *testing.T) {
	service, _, _, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	req := &models.PatientIdentifierCreateRequest{
		PatientID:       "patient-123",
		IdentifierType:  string(models.IdentifierTypeMyNumber),
		IdentifierValue: "12345", // Invalid - not 12 digits
		IsPrimary:       true,
	}
	createdBy := "staff-456"

	// Execute
	result, err := service.CreateIdentifier(ctx, req, createdBy)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "my_number must be 12 digits")
}

func TestIdentifierService_CreateIdentifier_ValidationError_InsuranceNoIssuer(t *testing.T) {
	service, _, _, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	req := &models.PatientIdentifierCreateRequest{
		PatientID:       "patient-123",
		IdentifierType:  string(models.IdentifierTypeInsuranceID),
		IdentifierValue: "12345678",
		IssuerName:      "", // Missing issuer name
	}
	createdBy := "staff-456"

	// Execute
	result, err := service.CreateIdentifier(ctx, req, createdBy)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "issuer_name is required")
}

// Tests for GetIdentifier
func TestIdentifierService_GetIdentifier_Success_NoDecrypt(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	identifierID := "id-789"
	requestorID := "staff-456"
	decrypt := false

	expectedIdentifier := &models.PatientIdentifier{
		IdentifierID:    identifierID,
		PatientID:       "patient-123",
		IdentifierType:  string(models.IdentifierTypeInsuranceID),
		IdentifierValue: "12345678",
		CreatedAt:       time.Now(),
	}

	// Setup mocks
	mockIdentifierRepo.On("GetIdentifierByID", ctx, identifierID, false).Return(expectedIdentifier, nil)
	mockPatientRepo.On("CheckStaffAccess", ctx, requestorID, expectedIdentifier.PatientID).Return(true, nil)

	// Execute
	result, err := service.GetIdentifier(ctx, identifierID, requestorID, decrypt)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedIdentifier.IdentifierID, result.IdentifierID)
	mockIdentifierRepo.AssertExpectations(t)
	mockPatientRepo.AssertExpectations(t)
}

func TestIdentifierService_GetIdentifier_Success_WithDecrypt(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, mockAuditRepo := setupIdentifierServiceTest()
	ctx := context.Background()

	identifierID := "id-789"
	requestorID := "staff-456"
	decrypt := true

	encryptedIdentifier := &models.PatientIdentifier{
		IdentifierID:    identifierID,
		PatientID:       "patient-123",
		IdentifierType:  string(models.IdentifierTypeMyNumber),
		IdentifierValue: "***********", // Encrypted
		CreatedAt:       time.Now(),
	}

	decryptedIdentifier := &models.PatientIdentifier{
		IdentifierID:    identifierID,
		PatientID:       "patient-123",
		IdentifierType:  string(models.IdentifierTypeMyNumber),
		IdentifierValue: "123456789012", // Decrypted
		CreatedAt:       time.Now(),
	}

	// Setup mocks
	mockIdentifierRepo.On("GetIdentifierByID", ctx, identifierID, false).Return(encryptedIdentifier, nil)
	mockPatientRepo.On("CheckStaffAccess", ctx, requestorID, encryptedIdentifier.PatientID).Return(true, nil)
	mockIdentifierRepo.On("GetIdentifierByID", ctx, identifierID, true).Return(decryptedIdentifier, nil)
	mockAuditRepo.On("LogMyNumberAccess", ctx, encryptedIdentifier.PatientID, identifierID, "decrypt", requestorID).Return(nil)

	// Execute
	result, err := service.GetIdentifier(ctx, identifierID, requestorID, decrypt)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "123456789012", result.IdentifierValue) // Should be decrypted
	mockIdentifierRepo.AssertExpectations(t)
	mockPatientRepo.AssertExpectations(t)
	mockAuditRepo.AssertExpectations(t)
}

func TestIdentifierService_GetIdentifier_NotFound(t *testing.T) {
	service, mockIdentifierRepo, _, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	identifierID := "non-existent-id"
	requestorID := "staff-456"

	// Setup mock
	mockIdentifierRepo.On("GetIdentifierByID", ctx, identifierID, false).Return(nil, errors.New("identifier not found"))

	// Execute
	result, err := service.GetIdentifier(ctx, identifierID, requestorID, false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "identifier not found")
	mockIdentifierRepo.AssertExpectations(t)
}

func TestIdentifierService_GetIdentifier_AccessDenied(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	identifierID := "id-789"
	requestorID := "staff-456"

	expectedIdentifier := &models.PatientIdentifier{
		IdentifierID:   identifierID,
		PatientID:      "patient-123",
		IdentifierType: string(models.IdentifierTypeMyNumber),
	}

	// Setup mocks
	mockIdentifierRepo.On("GetIdentifierByID", ctx, identifierID, false).Return(expectedIdentifier, nil)
	mockPatientRepo.On("CheckStaffAccess", ctx, requestorID, expectedIdentifier.PatientID).Return(false, nil)

	// Execute
	result, err := service.GetIdentifier(ctx, identifierID, requestorID, false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "access denied")
	mockIdentifierRepo.AssertExpectations(t)
	mockPatientRepo.AssertExpectations(t)
}

// Tests for GetIdentifiersByPatientID
func TestIdentifierService_GetIdentifiersByPatientID_Success(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	patientID := "patient-123"
	requestorID := "staff-456"
	decrypt := false

	expectedIdentifiers := []*models.PatientIdentifier{
		{
			IdentifierID:   "id-1",
			PatientID:      patientID,
			IdentifierType: string(models.IdentifierTypeMyNumber),
		},
		{
			IdentifierID:   "id-2",
			PatientID:      patientID,
			IdentifierType: string(models.IdentifierTypeInsuranceID),
		},
	}

	// Setup mocks
	mockPatientRepo.On("CheckStaffAccess", ctx, requestorID, patientID).Return(true, nil)
	mockIdentifierRepo.On("GetIdentifiersByPatientID", ctx, patientID, decrypt).Return(expectedIdentifiers, nil)

	// Execute
	result, err := service.GetIdentifiersByPatientID(ctx, patientID, requestorID, decrypt)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockPatientRepo.AssertExpectations(t)
	mockIdentifierRepo.AssertExpectations(t)
}

func TestIdentifierService_GetIdentifiersByPatientID_WithDecrypt(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, mockAuditRepo := setupIdentifierServiceTest()
	ctx := context.Background()

	patientID := "patient-123"
	requestorID := "staff-456"
	decrypt := true

	myNumberIdentifier := &models.PatientIdentifier{
		IdentifierID:   "id-1",
		PatientID:      patientID,
		IdentifierType: string(models.IdentifierTypeMyNumber),
	}

	expectedIdentifiers := []*models.PatientIdentifier{
		myNumberIdentifier,
		{
			IdentifierID:   "id-2",
			PatientID:      patientID,
			IdentifierType: string(models.IdentifierTypeInsuranceID),
		},
	}

	// Setup mocks
	mockPatientRepo.On("CheckStaffAccess", ctx, requestorID, patientID).Return(true, nil)
	mockIdentifierRepo.On("GetIdentifiersByPatientID", ctx, patientID, decrypt).Return(expectedIdentifiers, nil)
	mockAuditRepo.On("LogMyNumberAccess", ctx, patientID, myNumberIdentifier.IdentifierID, "decrypt", requestorID).Return(nil)

	// Execute
	result, err := service.GetIdentifiersByPatientID(ctx, patientID, requestorID, decrypt)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockPatientRepo.AssertExpectations(t)
	mockIdentifierRepo.AssertExpectations(t)
	mockAuditRepo.AssertExpectations(t)
}

// Tests for UpdateIdentifier
func TestIdentifierService_UpdateIdentifier_Success(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	identifierID := "id-789"
	updatedBy := "staff-456"
	verificationStatus := string(models.VerificationStatusVerified)

	req := &models.PatientIdentifierUpdateRequest{
		VerificationStatus: &verificationStatus,
	}

	existingIdentifier := &models.PatientIdentifier{
		IdentifierID:   identifierID,
		PatientID:      "patient-123",
		IdentifierType: string(models.IdentifierTypeInsuranceID),
	}

	updatedIdentifier := &models.PatientIdentifier{
		IdentifierID:       identifierID,
		PatientID:          "patient-123",
		IdentifierType:     string(models.IdentifierTypeInsuranceID),
		VerificationStatus: verificationStatus,
		UpdatedAt:          time.Now(),
	}

	// Setup mocks
	mockIdentifierRepo.On("GetIdentifierByID", ctx, identifierID, false).Return(existingIdentifier, nil)
	mockPatientRepo.On("CheckStaffAccess", ctx, updatedBy, existingIdentifier.PatientID).Return(true, nil)
	mockIdentifierRepo.On("UpdateIdentifier", ctx, identifierID, req, updatedBy).Return(updatedIdentifier, nil)

	// Execute
	result, err := service.UpdateIdentifier(ctx, identifierID, req, updatedBy)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, verificationStatus, result.VerificationStatus)
	mockIdentifierRepo.AssertExpectations(t)
	mockPatientRepo.AssertExpectations(t)
}

func TestIdentifierService_UpdateIdentifier_ValidationError(t *testing.T) {
	service, _, _, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	identifierID := "id-789"
	updatedBy := "staff-456"
	invalidStatus := "invalid-status"

	req := &models.PatientIdentifierUpdateRequest{
		VerificationStatus: &invalidStatus,
	}

	// Execute
	result, err := service.UpdateIdentifier(ctx, identifierID, req, updatedBy)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid verification_status")
}

// Tests for DeleteIdentifier
func TestIdentifierService_DeleteIdentifier_Success(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, mockAuditRepo := setupIdentifierServiceTest()
	ctx := context.Background()

	identifierID := "id-789"
	deletedBy := "staff-456"

	existingIdentifier := &models.PatientIdentifier{
		IdentifierID:   identifierID,
		PatientID:      "patient-123",
		IdentifierType: string(models.IdentifierTypeMyNumber),
	}

	// Setup mocks
	mockIdentifierRepo.On("GetIdentifierByID", ctx, identifierID, false).Return(existingIdentifier, nil)
	mockPatientRepo.On("CheckStaffAccess", ctx, deletedBy, existingIdentifier.PatientID).Return(true, nil)
	mockIdentifierRepo.On("DeleteIdentifier", ctx, identifierID).Return(nil)
	mockAuditRepo.On("LogMyNumberAccess", ctx, existingIdentifier.PatientID, identifierID, "delete", deletedBy).Return(nil)

	// Execute
	err := service.DeleteIdentifier(ctx, identifierID, deletedBy)

	// Assert
	assert.NoError(t, err)
	mockIdentifierRepo.AssertExpectations(t)
	mockPatientRepo.AssertExpectations(t)
	mockAuditRepo.AssertExpectations(t)
}

func TestIdentifierService_DeleteIdentifier_AccessDenied(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	identifierID := "id-789"
	deletedBy := "staff-456"

	existingIdentifier := &models.PatientIdentifier{
		IdentifierID:   identifierID,
		PatientID:      "patient-123",
		IdentifierType: string(models.IdentifierTypeInsuranceID),
	}

	// Setup mocks
	mockIdentifierRepo.On("GetIdentifierByID", ctx, identifierID, false).Return(existingIdentifier, nil)
	mockPatientRepo.On("CheckStaffAccess", ctx, deletedBy, existingIdentifier.PatientID).Return(false, nil)

	// Execute
	err := service.DeleteIdentifier(ctx, identifierID, deletedBy)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
	mockIdentifierRepo.AssertExpectations(t)
	mockPatientRepo.AssertExpectations(t)
}

// Tests for GetPrimaryIdentifier
func TestIdentifierService_GetPrimaryIdentifier_Success(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, _ := setupIdentifierServiceTest()
	ctx := context.Background()

	patientID := "patient-123"
	requestorID := "staff-456"
	identifierType := models.IdentifierTypeInsuranceID
	decrypt := false

	expectedIdentifier := &models.PatientIdentifier{
		IdentifierID:   "id-primary",
		PatientID:      patientID,
		IdentifierType: string(identifierType),
		IsPrimary:      true,
	}

	// Setup mocks
	mockPatientRepo.On("CheckStaffAccess", ctx, requestorID, patientID).Return(true, nil)
	mockIdentifierRepo.On("GetPrimaryIdentifier", ctx, patientID, identifierType, decrypt).Return(expectedIdentifier, nil)

	// Execute
	result, err := service.GetPrimaryIdentifier(ctx, patientID, identifierType, requestorID, decrypt)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsPrimary)
	mockPatientRepo.AssertExpectations(t)
	mockIdentifierRepo.AssertExpectations(t)
}

func TestIdentifierService_GetPrimaryIdentifier_WithDecrypt(t *testing.T) {
	service, mockIdentifierRepo, mockPatientRepo, mockAuditRepo := setupIdentifierServiceTest()
	ctx := context.Background()

	patientID := "patient-123"
	requestorID := "staff-456"
	identifierType := models.IdentifierTypeMyNumber
	decrypt := true

	expectedIdentifier := &models.PatientIdentifier{
		IdentifierID:    "id-primary",
		PatientID:       patientID,
		IdentifierType:  string(identifierType),
		IdentifierValue: "123456789012",
		IsPrimary:       true,
	}

	// Setup mocks
	mockPatientRepo.On("CheckStaffAccess", ctx, requestorID, patientID).Return(true, nil)
	mockIdentifierRepo.On("GetPrimaryIdentifier", ctx, patientID, identifierType, decrypt).Return(expectedIdentifier, nil)
	mockAuditRepo.On("LogMyNumberAccess", ctx, patientID, expectedIdentifier.IdentifierID, "decrypt", requestorID).Return(nil)

	// Execute
	result, err := service.GetPrimaryIdentifier(ctx, patientID, identifierType, requestorID, decrypt)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsPrimary)
	mockPatientRepo.AssertExpectations(t)
	mockIdentifierRepo.AssertExpectations(t)
	mockAuditRepo.AssertExpectations(t)
}

// Tests for validation methods
func TestIdentifierService_validateCreateRequest(t *testing.T) {
	service, _, _, _ := setupIdentifierServiceTest()

	tests := []struct {
		name    string
		req     *models.PatientIdentifierCreateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid My Number",
			req: &models.PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  string(models.IdentifierTypeMyNumber),
				IdentifierValue: "123456789012",
			},
			wantErr: false,
		},
		{
			name: "Invalid - empty patient ID",
			req: &models.PatientIdentifierCreateRequest{
				PatientID:       "",
				IdentifierType:  string(models.IdentifierTypeMyNumber),
				IdentifierValue: "123456789012",
			},
			wantErr: true,
			errMsg:  "patient_id is required",
		},
		{
			name: "Invalid - empty identifier type",
			req: &models.PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  "",
				IdentifierValue: "123456789012",
			},
			wantErr: true,
			errMsg:  "identifier_type is required",
		},
		{
			name: "Invalid - invalid identifier type",
			req: &models.PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  "invalid_type",
				IdentifierValue: "123456789012",
			},
			wantErr: true,
			errMsg:  "invalid identifier_type",
		},
		{
			name: "Invalid - My Number wrong length",
			req: &models.PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  string(models.IdentifierTypeMyNumber),
				IdentifierValue: "12345",
			},
			wantErr: true,
			errMsg:  "my_number must be 12 digits",
		},
		{
			name: "Invalid - Insurance ID missing issuer",
			req: &models.PatientIdentifierCreateRequest{
				PatientID:       "patient-123",
				IdentifierType:  string(models.IdentifierTypeInsuranceID),
				IdentifierValue: "12345678",
			},
			wantErr: true,
			errMsg:  "issuer_name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateCreateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIdentifierService_validateUpdateRequest(t *testing.T) {
	service, _, _, _ := setupIdentifierServiceTest()

	validStatus := string(models.VerificationStatusVerified)
	invalidStatus := "invalid_status"

	tests := []struct {
		name    string
		req     *models.PatientIdentifierUpdateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid - verified status",
			req: &models.PatientIdentifierUpdateRequest{
				VerificationStatus: &validStatus,
			},
			wantErr: false,
		},
		{
			name: "Invalid - invalid verification status",
			req: &models.PatientIdentifierUpdateRequest{
				VerificationStatus: &invalidStatus,
			},
			wantErr: true,
			errMsg:  "invalid verification_status",
		},
		{
			name:    "Valid - no updates",
			req:     &models.PatientIdentifierUpdateRequest{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateUpdateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
