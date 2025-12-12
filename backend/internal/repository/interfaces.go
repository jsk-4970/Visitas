package repository

import (
	"context"

	"github.com/visitas/backend/internal/models"
)

// PatientRepositoryInterface defines the interface for patient repository operations
type PatientRepositoryInterface interface {
	CheckStaffAccess(ctx context.Context, staffID, patientID string) (bool, error)
	GetPatientByID(ctx context.Context, patientID string) (*models.Patient, error)
}

// IdentifierRepositoryInterface defines the interface for identifier repository operations
type IdentifierRepositoryInterface interface {
	CreateIdentifier(ctx context.Context, req *models.PatientIdentifierCreateRequest, createdBy string) (*models.PatientIdentifier, error)
	GetIdentifierByID(ctx context.Context, identifierID string, decrypt bool) (*models.PatientIdentifier, error)
	GetIdentifiersByPatientID(ctx context.Context, patientID string, decrypt bool) ([]*models.PatientIdentifier, error)
	UpdateIdentifier(ctx context.Context, identifierID string, req *models.PatientIdentifierUpdateRequest, updatedBy string) (*models.PatientIdentifier, error)
	DeleteIdentifier(ctx context.Context, identifierID string) error
	GetPrimaryIdentifier(ctx context.Context, patientID string, identifierType models.IdentifierType, decrypt bool) (*models.PatientIdentifier, error)
}

// AuditRepositoryInterface defines the interface for audit repository operations
type AuditRepositoryInterface interface {
	LogMyNumberAccess(ctx context.Context, patientID, identifierID, action, userID string) error
	LogAccess(ctx context.Context, log *AuditLog) error
}
