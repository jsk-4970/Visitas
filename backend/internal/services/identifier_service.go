package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// IdentifierService handles business logic for patient identifier operations
type IdentifierService struct {
	identifierRepo repository.IdentifierRepositoryInterface
	patientRepo    repository.PatientRepositoryInterface
	auditRepo      repository.AuditRepositoryInterface
}

// NewIdentifierService creates a new identifier service
func NewIdentifierService(
	identifierRepo repository.IdentifierRepositoryInterface,
	patientRepo repository.PatientRepositoryInterface,
	auditRepo repository.AuditRepositoryInterface,
) *IdentifierService {
	return &IdentifierService{
		identifierRepo: identifierRepo,
		patientRepo:    patientRepo,
		auditRepo:      auditRepo,
	}
}

// CreateIdentifier creates a new patient identifier with access control
func (s *IdentifierService) CreateIdentifier(ctx context.Context, req *models.PatientIdentifierCreateRequest, createdBy string) (*models.PatientIdentifier, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		logger.WarnContext(ctx, "Invalid identifier create request", map[string]interface{}{
			"error":      err.Error(),
			"patient_id": req.PatientID,
		})
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if user has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, createdBy, req.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized identifier creation attempt", map[string]interface{}{
			"patient_id": req.PatientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to add identifiers for this patient")
	}

	// Create identifier (repository handles encryption for My Number)
	identifier, err := s.identifierRepo.CreateIdentifier(ctx, req, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create identifier", err, map[string]interface{}{
			"patient_id":      req.PatientID,
			"identifier_type": req.IdentifierType,
			"created_by":      createdBy,
		})
		return nil, fmt.Errorf("failed to create identifier: %w", err)
	}

	logger.InfoContext(ctx, "Patient identifier created successfully", map[string]interface{}{
		"identifier_id":   identifier.IdentifierID,
		"patient_id":      identifier.PatientID,
		"identifier_type": identifier.IdentifierType,
		"is_primary":      identifier.IsPrimary,
		"created_by":      createdBy,
	})

	// Log audit trail for My Number creation
	if identifier.IsMyNumber() {
		if err := s.auditRepo.LogMyNumberAccess(ctx, identifier.PatientID, identifier.IdentifierID, "create", createdBy); err != nil {
			logger.ErrorContext(ctx, "Failed to log My Number access audit", err, map[string]interface{}{
				"identifier_id": identifier.IdentifierID,
				"patient_id":    identifier.PatientID,
				"action":        "create",
			})
		}
	}

	return identifier, nil
}

// GetIdentifier retrieves an identifier by ID with access control
func (s *IdentifierService) GetIdentifier(ctx context.Context, identifierID, requestorID string, decrypt bool) (*models.PatientIdentifier, error) {
	// Get identifier first (without decryption)
	identifier, err := s.identifierRepo.GetIdentifierByID(ctx, identifierID, false)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get identifier", err, map[string]interface{}{
			"identifier_id": identifierID,
		})
		return nil, fmt.Errorf("failed to get identifier: %w", err)
	}

	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, identifier.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"requestor_id":  requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized identifier access attempt", map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"requestor_id":  requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this patient's identifiers")
	}

	// If decryption is requested for My Number
	if decrypt && identifier.IsMyNumber() {
		decryptedIdentifier, err := s.identifierRepo.GetIdentifierByID(ctx, identifierID, true)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to decrypt My Number", err, map[string]interface{}{
				"identifier_id": identifierID,
				"patient_id":    identifier.PatientID,
			})
			return nil, fmt.Errorf("failed to decrypt My Number: %w", err)
		}

		// Log audit trail for My Number decryption
		if err := s.auditRepo.LogMyNumberAccess(ctx, identifier.PatientID, identifierID, "decrypt", requestorID); err != nil {
			logger.ErrorContext(ctx, "Failed to log My Number access audit", err, map[string]interface{}{
				"identifier_id": identifierID,
				"patient_id":    identifier.PatientID,
				"action":        "decrypt",
			})
		}

		logger.InfoContext(ctx, "My Number decrypted successfully", map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"requestor_id":  requestorID,
		})

		return decryptedIdentifier, nil
	}

	return identifier, nil
}

// GetIdentifiersByPatientID retrieves all identifiers for a patient with access control
func (s *IdentifierService) GetIdentifiersByPatientID(ctx context.Context, patientID, requestorID string, decrypt bool) ([]*models.PatientIdentifier, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized patient identifiers access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this patient's identifiers")
	}

	// Get identifiers with optional decryption
	identifiers, err := s.identifierRepo.GetIdentifiersByPatientID(ctx, patientID, decrypt)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get patient identifiers", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get identifiers: %w", err)
	}

	// Log audit trail if My Number was decrypted
	if decrypt {
		for _, identifier := range identifiers {
			if identifier.IsMyNumber() {
				if err := s.auditRepo.LogMyNumberAccess(ctx, patientID, identifier.IdentifierID, "decrypt", requestorID); err != nil {
					logger.ErrorContext(ctx, "Failed to log My Number access audit", err, map[string]interface{}{
						"identifier_id": identifier.IdentifierID,
						"patient_id":    patientID,
						"action":        "decrypt",
					})
				}
			}
		}

		logger.InfoContext(ctx, "Patient identifiers retrieved with decryption", map[string]interface{}{
			"patient_id":   patientID,
			"count":        len(identifiers),
			"requestor_id": requestorID,
		})
	}

	return identifiers, nil
}

// UpdateIdentifier updates an identifier with access control
func (s *IdentifierService) UpdateIdentifier(ctx context.Context, identifierID string, req *models.PatientIdentifierUpdateRequest, updatedBy string) (*models.PatientIdentifier, error) {
	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		logger.WarnContext(ctx, "Invalid identifier update request", map[string]interface{}{
			"error":         err.Error(),
			"identifier_id": identifierID,
		})
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get identifier first to check patient ID
	identifier, err := s.identifierRepo.GetIdentifierByID(ctx, identifierID, false)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get identifier", err, map[string]interface{}{
			"identifier_id": identifierID,
		})
		return nil, fmt.Errorf("failed to get identifier: %w", err)
	}

	// Check if user has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, identifier.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"updated_by":    updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized identifier update attempt", map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"updated_by":    updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this patient's identifiers")
	}

	// Update identifier (repository handles encryption if needed)
	updatedIdentifier, err := s.identifierRepo.UpdateIdentifier(ctx, identifierID, req, updatedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update identifier", err, map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"updated_by":    updatedBy,
		})
		return nil, fmt.Errorf("failed to update identifier: %w", err)
	}

	logger.InfoContext(ctx, "Patient identifier updated successfully", map[string]interface{}{
		"identifier_id": updatedIdentifier.IdentifierID,
		"patient_id":    updatedIdentifier.PatientID,
		"updated_by":    updatedBy,
	})

	// Log audit trail for My Number updates
	if identifier.IsMyNumber() {
		if err := s.auditRepo.LogMyNumberAccess(ctx, identifier.PatientID, identifierID, "update", updatedBy); err != nil {
			logger.ErrorContext(ctx, "Failed to log My Number access audit", err, map[string]interface{}{
				"identifier_id": identifierID,
				"patient_id":    identifier.PatientID,
				"action":        "update",
			})
		}
	}

	return updatedIdentifier, nil
}

// DeleteIdentifier soft deletes an identifier with access control
func (s *IdentifierService) DeleteIdentifier(ctx context.Context, identifierID, deletedBy string) error {
	// Get identifier first to check patient ID
	identifier, err := s.identifierRepo.GetIdentifierByID(ctx, identifierID, false)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get identifier", err, map[string]interface{}{
			"identifier_id": identifierID,
		})
		return fmt.Errorf("failed to get identifier: %w", err)
	}

	// Check if user has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, identifier.PatientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"deleted_by":    deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized identifier delete attempt", map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"deleted_by":    deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this patient's identifiers")
	}

	// Delete identifier
	err = s.identifierRepo.DeleteIdentifier(ctx, identifierID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete identifier", err, map[string]interface{}{
			"identifier_id": identifierID,
			"patient_id":    identifier.PatientID,
			"deleted_by":    deletedBy,
		})
		return fmt.Errorf("failed to delete identifier: %w", err)
	}

	logger.InfoContext(ctx, "Patient identifier deleted successfully", map[string]interface{}{
		"identifier_id": identifierID,
		"patient_id":    identifier.PatientID,
		"deleted_by":    deletedBy,
	})

	// Log audit trail for My Number deletion
	if identifier.IsMyNumber() {
		if err := s.auditRepo.LogMyNumberAccess(ctx, identifier.PatientID, identifierID, "delete", deletedBy); err != nil {
			logger.ErrorContext(ctx, "Failed to log My Number access audit", err, map[string]interface{}{
				"identifier_id": identifierID,
				"patient_id":    identifier.PatientID,
				"action":        "delete",
			})
		}
	}

	return nil
}

// GetPrimaryIdentifier retrieves the primary identifier for a patient by type with access control
func (s *IdentifierService) GetPrimaryIdentifier(ctx context.Context, patientID string, identifierType models.IdentifierType, requestorID string, decrypt bool) (*models.PatientIdentifier, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":      patientID,
			"identifier_type": identifierType,
			"requestor_id":    requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized primary identifier access attempt", map[string]interface{}{
			"patient_id":      patientID,
			"identifier_type": identifierType,
			"requestor_id":    requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this patient's identifiers")
	}

	// Get primary identifier
	identifier, err := s.identifierRepo.GetPrimaryIdentifier(ctx, patientID, identifierType, decrypt)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get primary identifier", err, map[string]interface{}{
			"patient_id":      patientID,
			"identifier_type": identifierType,
		})
		return nil, fmt.Errorf("failed to get primary identifier: %w", err)
	}

	// Log audit trail if My Number was decrypted
	if decrypt && identifier.IsMyNumber() {
		if err := s.auditRepo.LogMyNumberAccess(ctx, patientID, identifier.IdentifierID, "decrypt", requestorID); err != nil {
			logger.ErrorContext(ctx, "Failed to log My Number access audit", err, map[string]interface{}{
				"identifier_id": identifier.IdentifierID,
				"patient_id":    patientID,
				"action":        "decrypt",
			})
		}

		logger.InfoContext(ctx, "Primary My Number decrypted successfully", map[string]interface{}{
			"identifier_id":   identifier.IdentifierID,
			"patient_id":      patientID,
			"identifier_type": identifierType,
			"requestor_id":    requestorID,
		})
	}

	return identifier, nil
}

// validateCreateRequest validates identifier create request
func (s *IdentifierService) validateCreateRequest(req *models.PatientIdentifierCreateRequest) error {
	if req.PatientID == "" {
		return fmt.Errorf("patient_id is required")
	}

	if req.IdentifierType == "" {
		return fmt.Errorf("identifier_type is required")
	}

	// Validate identifier type
	validTypes := []string{
		string(models.IdentifierTypeMyNumber),
		string(models.IdentifierTypeInsuranceID),
		string(models.IdentifierTypeCareInsuranceID),
		string(models.IdentifierTypeMRN),
		string(models.IdentifierTypeOther),
	}
	isValid := false
	for _, t := range validTypes {
		if req.IdentifierType == t {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid identifier_type: must be one of [my_number, insurance_id, care_insurance_id, mrn, other]")
	}

	if req.IdentifierValue == "" {
		return fmt.Errorf("identifier_value is required")
	}

	// Validate My Number format (12 digits)
	if req.IdentifierType == string(models.IdentifierTypeMyNumber) {
		if len(req.IdentifierValue) != 12 {
			return fmt.Errorf("my_number must be 12 digits")
		}
	}

	// Validate issuer information for insurance cards
	if req.IdentifierType == string(models.IdentifierTypeInsuranceID) || req.IdentifierType == string(models.IdentifierTypeCareInsuranceID) {
		if req.IssuerName == "" {
			return fmt.Errorf("issuer_name is required for insurance identifiers")
		}
	}

	return nil
}

// validateUpdateRequest validates identifier update request
func (s *IdentifierService) validateUpdateRequest(req *models.PatientIdentifierUpdateRequest) error {
	// Validate verification status if provided
	if req.VerificationStatus != nil {
		validStatuses := []string{
			string(models.VerificationStatusVerified),
			string(models.VerificationStatusUnverified),
			string(models.VerificationStatusExpired),
			string(models.VerificationStatusInvalid),
		}
		isValid := false
		for _, status := range validStatuses {
			if *req.VerificationStatus == status {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid verification_status: must be one of [verified, unverified, expired, invalid]")
		}
	}

	return nil
}
