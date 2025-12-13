package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/pkg/encryption"
	"google.golang.org/api/iterator"
)

// IdentifierRepository handles patient identifier data operations
type IdentifierRepository struct {
	client    *spanner.Client
	encryptor *encryption.KMSEncryptor
}

// NewIdentifierRepository creates a new identifier repository
func NewIdentifierRepository(spannerRepo *SpannerRepository, encryptor *encryption.KMSEncryptor) *IdentifierRepository {
	return &IdentifierRepository{
		client:    spannerRepo.client,
		encryptor: encryptor,
	}
}

// CreateIdentifier creates a new patient identifier
func (r *IdentifierRepository) CreateIdentifier(ctx context.Context, req *models.PatientIdentifierCreateRequest, createdBy string) (*models.PatientIdentifier, error) {
	identifierID := uuid.New().String()
	now := time.Now()

	// Determine if encryption is needed (for My Number)
	identifierValue := req.IdentifierValue
	var err error

	if req.IdentifierType == string(models.IdentifierTypeMyNumber) {
		// Encrypt the My Number with patient-specific AAD
		if r.encryptor == nil {
			return nil, fmt.Errorf("KMS encryptor not configured for My Number encryption")
		}
		identifierValue, err = r.encryptor.EncryptMyNumberWithPatient(ctx, req.IdentifierValue, req.PatientID)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt My Number: %w", err)
		}
	}

	// Create insert mutation
	mutation := spanner.InsertMap("patient_identifiers", map[string]interface{}{
		"identifier_id":       identifierID,
		"patient_id":          req.PatientID,
		"identifier_type":     req.IdentifierType,
		"identifier_value":    identifierValue,
		"is_primary":          req.IsPrimary,
		"valid_from":          req.ValidFrom,
		"valid_to":            req.ValidTo,
		"issuer_name":         req.IssuerName,
		"issuer_code":         req.IssuerCode,
		"verification_status": string(models.VerificationStatusUnverified),
		"verified_at":         nil,
		"verified_by":         "",
		"deleted":             false,
		"deleted_at":          nil,
		"created_at":          now,
		"created_by":          createdBy,
		"updated_at":          now,
		"updated_by":          createdBy,
	})

	// Apply the mutation
	_, err = r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to insert identifier: %w", err)
	}

	// Retrieve the created identifier (without decrypting)
	identifier, err := r.GetIdentifierByID(ctx, identifierID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created identifier: %w", err)
	}

	return identifier, nil
}

// GetIdentifierByID retrieves an identifier by ID
// If decrypt is true and the identifier is a My Number, it will be decrypted
func (r *IdentifierRepository) GetIdentifierByID(ctx context.Context, identifierID string, decrypt bool) (*models.PatientIdentifier, error) {
	stmt := NewStatement(`SELECT
			identifier_id, patient_id, identifier_type, identifier_value,
			is_primary, valid_from, valid_to, issuer_name, issuer_code,
			verification_status, verified_at, verified_by,
			deleted, deleted_at, created_at, created_by, updated_at, updated_by
		FROM patient_identifiers
		WHERE identifier_id = @identifierID AND deleted = false`,
		map[string]interface{}{
			"identifierID": identifierID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("identifier not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query identifier: %w", err)
	}

	identifier, err := scanIdentifier(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan identifier: %w", err)
	}

	// Decrypt My Number if requested
	if decrypt && identifier.IsMyNumber() {
		if r.encryptor == nil {
			return nil, fmt.Errorf("KMS encryptor not configured for My Number decryption")
		}
		decryptedValue, err := r.encryptor.DecryptMyNumberWithPatient(ctx, identifier.IdentifierValue, identifier.PatientID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt My Number: %w", err)
		}
		identifier.IdentifierValue = decryptedValue
	}

	return identifier, nil
}

// GetIdentifiersByPatientID retrieves all identifiers for a patient
// If decrypt is true, My Numbers will be decrypted
func (r *IdentifierRepository) GetIdentifiersByPatientID(ctx context.Context, patientID string, decrypt bool) ([]*models.PatientIdentifier, error) {
	stmt := NewStatement(`SELECT
			identifier_id, patient_id, identifier_type, identifier_value,
			is_primary, valid_from, valid_to, issuer_name, issuer_code,
			verification_status, verified_at, verified_by,
			deleted, deleted_at, created_at, created_by, updated_at, updated_by
		FROM patient_identifiers
		WHERE patient_id = @patientID AND deleted = false
		ORDER BY is_primary DESC, created_at ASC`,
		map[string]interface{}{
			"patientID": patientID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var identifiers []*models.PatientIdentifier
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate identifiers: %w", err)
		}

		identifier, err := scanIdentifier(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan identifier: %w", err)
		}

		// Decrypt My Number if requested
		if decrypt && identifier.IsMyNumber() {
			if r.encryptor == nil {
				return nil, fmt.Errorf("KMS encryptor not configured for My Number decryption")
			}
			decryptedValue, err := r.encryptor.DecryptMyNumberWithPatient(ctx, identifier.IdentifierValue, patientID)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt My Number: %w", err)
			}
			identifier.IdentifierValue = decryptedValue
		}

		identifiers = append(identifiers, identifier)
	}

	return identifiers, nil
}

// UpdateIdentifier updates an identifier
func (r *IdentifierRepository) UpdateIdentifier(ctx context.Context, identifierID string, req *models.PatientIdentifierUpdateRequest, updatedBy string) (*models.PatientIdentifier, error) {
	// First, get the current identifier
	currentIdentifier, err := r.GetIdentifierByID(ctx, identifierID, false)
	if err != nil {
		return nil, fmt.Errorf("identifier not found: %w", err)
	}

	now := time.Now()
	updates := make(map[string]interface{})
	updates["identifier_id"] = identifierID
	updates["updated_at"] = now
	updates["updated_by"] = updatedBy

	// Update identifier value if provided
	if req.IdentifierValue != nil {
		identifierValue := *req.IdentifierValue

		// Encrypt if My Number
		if currentIdentifier.IsMyNumber() {
			if r.encryptor == nil {
				return nil, fmt.Errorf("KMS encryptor not configured for My Number encryption")
			}
			encryptedValue, err := r.encryptor.EncryptMyNumberWithPatient(ctx, identifierValue, currentIdentifier.PatientID)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt My Number: %w", err)
			}
			identifierValue = encryptedValue
		}

		updates["identifier_value"] = identifierValue
	}

	if req.IsPrimary != nil {
		updates["is_primary"] = *req.IsPrimary
	}

	if req.ValidFrom != nil {
		updates["valid_from"] = *req.ValidFrom
	}

	if req.ValidTo != nil {
		updates["valid_to"] = *req.ValidTo
	}

	if req.IssuerName != nil {
		updates["issuer_name"] = *req.IssuerName
	}

	if req.IssuerCode != nil {
		updates["issuer_code"] = *req.IssuerCode
	}

	if req.VerificationStatus != nil {
		updates["verification_status"] = *req.VerificationStatus
		if *req.VerificationStatus == string(models.VerificationStatusVerified) {
			updates["verified_at"] = now
			updates["verified_by"] = updatedBy
		}
	}

	// Create update mutation
	mutation := spanner.UpdateMap("patient_identifiers", updates)

	// Apply the mutation
	_, err = r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update identifier: %w", err)
	}

	// Retrieve the updated identifier
	identifier, err := r.GetIdentifierByID(ctx, identifierID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated identifier: %w", err)
	}

	return identifier, nil
}

// DeleteIdentifier performs a soft delete on an identifier
func (r *IdentifierRepository) DeleteIdentifier(ctx context.Context, identifierID string) error {
	now := time.Now()

	mutation := spanner.UpdateMap("patient_identifiers", map[string]interface{}{
		"identifier_id": identifierID,
		"deleted":       true,
		"deleted_at":    now,
		"updated_at":    now,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete identifier: %w", err)
	}

	return nil
}

// GetPrimaryIdentifier retrieves the primary identifier for a patient by type
func (r *IdentifierRepository) GetPrimaryIdentifier(ctx context.Context, patientID string, identifierType models.IdentifierType, decrypt bool) (*models.PatientIdentifier, error) {
	stmt := NewStatement(`SELECT
			identifier_id, patient_id, identifier_type, identifier_value,
			is_primary, valid_from, valid_to, issuer_name, issuer_code,
			verification_status, verified_at, verified_by,
			deleted, deleted_at, created_at, created_by, updated_at, updated_by
		FROM patient_identifiers
		WHERE patient_id = @patientID
			AND identifier_type = @identifierType
			AND is_primary = true
			AND deleted = false
		LIMIT 1`,
		map[string]interface{}{
			"patientID":      patientID,
			"identifierType": string(identifierType),
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("primary identifier not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query primary identifier: %w", err)
	}

	identifier, err := scanIdentifier(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan identifier: %w", err)
	}

	// Decrypt My Number if requested
	if decrypt && identifier.IsMyNumber() {
		if r.encryptor == nil {
			return nil, fmt.Errorf("KMS encryptor not configured for My Number decryption")
		}
		decryptedValue, err := r.encryptor.DecryptMyNumberWithPatient(ctx, identifier.IdentifierValue, patientID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt My Number: %w", err)
		}
		identifier.IdentifierValue = decryptedValue
	}

	return identifier, nil
}

// scanIdentifier scans a Spanner row into a PatientIdentifier model
func scanIdentifier(row *spanner.Row) (*models.PatientIdentifier, error) {
	var identifier models.PatientIdentifier

	err := row.Columns(
		&identifier.IdentifierID,
		&identifier.PatientID,
		&identifier.IdentifierType,
		&identifier.IdentifierValue,
		&identifier.IsPrimary,
		&identifier.ValidFrom,
		&identifier.ValidTo,
		&identifier.IssuerName,
		&identifier.IssuerCode,
		&identifier.VerificationStatus,
		&identifier.VerifiedAt,
		&identifier.VerifiedBy,
		&identifier.Deleted,
		&identifier.DeletedAt,
		&identifier.CreatedAt,
		&identifier.CreatedBy,
		&identifier.UpdatedAt,
		&identifier.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}

	return &identifier, nil
}
