package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"google.golang.org/api/iterator"
)

// PatientRepository handles patient data operations
type PatientRepository struct {
	client *spanner.Client
}

// NewPatientRepository creates a new patient repository
func NewPatientRepository(spannerRepo *SpannerRepository) *PatientRepository {
	return &PatientRepository{
		client: spannerRepo.client,
	}
}

// CreatePatient creates a new patient record
func (r *PatientRepository) CreatePatient(ctx context.Context, req *models.PatientCreateRequest, createdBy string) (*models.Patient, error) {
	patientID := uuid.New().String()
	now := time.Now()

	// Convert name to name_history JSONB array
	nameHistory := []models.NameRecord{req.Name}
	nameHistoryJSON, err := json.Marshal(nameHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal name history: %w", err)
	}

	// Convert contact points to JSONB
	contactPointsJSON, err := json.Marshal(req.ContactPoints)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contact points: %w", err)
	}

	// Convert addresses to JSONB
	addressesJSON, err := json.Marshal(req.Addresses)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal addresses: %w", err)
	}

	// Parse birth_date
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("invalid birth_date format: %w", err)
	}

	// Create insert mutation
	mutation := spanner.InsertMap("patients", map[string]interface{}{
		"patient_id":           patientID,
		"birth_date":           birthDate,
		"gender":               req.Gender,
		"blood_type":           req.BloodType,
		"name_history":         string(nameHistoryJSON),
		"contact_points":       string(contactPointsJSON),
		"addresses":            string(addressesJSON),
		"consent_details":      nil, // Optional JSONB field for consent details
		"consent_status":       req.ConsentStatus,
		"consent_obtained_at":  req.ConsentObtainedAt,
		"consent_withdrawn_at": nil,
		"deleted":              false,
		"deleted_at":           nil,
		"deleted_reason":       "",
		"created_at":           now,
		"created_by":           createdBy,
		"updated_at":           now,
		"updated_by":           createdBy,
	})

	// Apply the mutation
	_, err = r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to insert patient: %w", err)
	}

	// Retrieve the created patient
	patient, err := r.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created patient: %w", err)
	}

	return patient, nil
}

// GetPatientByID retrieves a patient by ID
func (r *PatientRepository) GetPatientByID(ctx context.Context, patientID string) (*models.Patient, error) {
	stmt := NewStatement(`SELECT
			patient_id, birth_date, gender, blood_type,
			name_history::text, COALESCE(contact_points::text, '[]'), COALESCE(addresses::text, '[]'), consent_details::text,
			COALESCE(current_family_name, ''), COALESCE(current_given_name, ''), COALESCE(primary_phone, ''),
			COALESCE(current_prefecture, ''), COALESCE(current_city, ''),
			consent_status, consent_obtained_at, consent_withdrawn_at,
			deleted, deleted_at, deleted_reason,
			created_at, created_by, updated_at, updated_by
		FROM patients
		WHERE patient_id = @patientID AND deleted = false`,
		map[string]interface{}{
			"patientID": patientID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("patient not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query patient: %w", err)
	}

	patient, err := scanPatient(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan patient: %w", err)
	}

	return patient, nil
}

// GetPatientsByStaffID retrieves all patients assigned to a staff member (RLS implementation)
func (r *PatientRepository) GetPatientsByStaffID(ctx context.Context, staffID string, limit, offset int) ([]*models.Patient, int, error) {
	// Query using RLS view
	stmt := NewStatement(`SELECT
			p.patient_id, p.birth_date, p.gender, p.blood_type,
			p.name_history, p.contact_points, p.addresses, p.consent_details,
			p.current_family_name, p.current_given_name, p.primary_phone,
			p.current_prefecture, p.current_city,
			p.consent_status, p.consent_obtained_at, p.consent_withdrawn_at,
			p.deleted, p.deleted_at, p.deleted_reason,
			p.created_at, p.created_by, p.updated_at, p.updated_by
		FROM patients p
		INNER JOIN staff_patient_assignments spa
			ON p.patient_id = spa.patient_id
		WHERE spa.staff_id = @staffID
			AND spa.status = 'active'
			AND p.deleted = false
		ORDER BY p.updated_at DESC
		LIMIT @limit OFFSET @offset`,
		map[string]interface{}{
			"staffID": staffID,
			"limit":   limit,
			"offset":  offset,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var patients []*models.Patient
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("failed to iterate patients: %w", err)
		}

		patient, err := scanPatient(row)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan patient: %w", err)
		}

		patients = append(patients, patient)
	}

	// Get total count
	countStmt := NewStatement(`SELECT COUNT(*) as total
		FROM patients p
		INNER JOIN staff_patient_assignments spa
			ON p.patient_id = spa.patient_id
		WHERE spa.staff_id = @staffID
			AND spa.status = 'active'
			AND p.deleted = false`,
		map[string]interface{}{
			"staffID": staffID,
		})

	countIter := r.client.Single().Query(ctx, countStmt)
	defer countIter.Stop()

	countRow, err := countIter.Next()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	var total int64
	if err := countRow.Columns(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return patients, int(total), nil
}

// UpdatePatient updates a patient record
func (r *PatientRepository) UpdatePatient(ctx context.Context, patientID string, req *models.PatientUpdateRequest, updatedBy string) (*models.Patient, error) {
	// First, get the current patient to merge updates
	currentPatient, err := r.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	now := time.Now()
	updates := make(map[string]interface{})
	updates["patient_id"] = patientID
	updates["updated_at"] = now
	updates["updated_by"] = updatedBy

	// Update basic demographics if provided
	if req.BirthDate != nil {
		birthDate, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("invalid birth_date format: %w", err)
		}
		updates["birth_date"] = birthDate
	}

	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}

	if req.BloodType != nil {
		updates["blood_type"] = *req.BloodType
	}

	// Handle name history updates
	if req.AddName != nil {
		nameHistory, err := currentPatient.GetNameHistory()
		if err != nil {
			return nil, fmt.Errorf("failed to get name history: %w", err)
		}
		// Append new name
		nameHistory = append(nameHistory, *req.AddName)
		nameHistoryJSON, err := json.Marshal(nameHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal name history: %w", err)
		}
		updates["name_history"] = string(nameHistoryJSON)
	}

	// Handle contact points updates
	if req.ContactPoints != nil {
		contactPointsJSON, err := json.Marshal(*req.ContactPoints)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal contact points: %w", err)
		}
		updates["contact_points"] = string(contactPointsJSON)
	} else if req.AddContactPoint != nil {
		contactPoints, err := currentPatient.GetContactPoints()
		if err != nil {
			return nil, fmt.Errorf("failed to get contact points: %w", err)
		}
		contactPoints = append(contactPoints, *req.AddContactPoint)
		contactPointsJSON, err := json.Marshal(contactPoints)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal contact points: %w", err)
		}
		updates["contact_points"] = string(contactPointsJSON)
	}

	// Handle address updates
	if req.Addresses != nil {
		addressesJSON, err := json.Marshal(*req.Addresses)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal addresses: %w", err)
		}
		updates["addresses"] = string(addressesJSON)
	} else if req.AddAddress != nil {
		addresses, err := currentPatient.GetAddresses()
		if err != nil {
			return nil, fmt.Errorf("failed to get addresses: %w", err)
		}
		addresses = append(addresses, *req.AddAddress)
		addressesJSON, err := json.Marshal(addresses)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal addresses: %w", err)
		}
		updates["addresses"] = string(addressesJSON)
	}

	// Handle consent status updates
	if req.ConsentStatus != nil {
		updates["consent_status"] = *req.ConsentStatus
	}

	if req.ConsentObtainedAt != nil {
		updates["consent_obtained_at"] = *req.ConsentObtainedAt
	}

	if req.ConsentWithdrawnAt != nil {
		updates["consent_withdrawn_at"] = *req.ConsentWithdrawnAt
	}

	// Create update mutation
	mutation := spanner.UpdateMap("patients", updates)

	// Apply the mutation
	_, err = r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update patient: %w", err)
	}

	// Retrieve the updated patient
	patient, err := r.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated patient: %w", err)
	}

	return patient, nil
}

// DeletePatient performs a soft delete on a patient record
func (r *PatientRepository) DeletePatient(ctx context.Context, patientID, deletedBy, reason string) error {
	now := time.Now()

	mutation := spanner.UpdateMap("patients", map[string]interface{}{
		"patient_id":     patientID,
		"deleted":        true,
		"deleted_at":     now,
		"deleted_reason": reason,
		"updated_at":     now,
		"updated_by":     deletedBy,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete patient: %w", err)
	}

	return nil
}

// CheckStaffAccess verifies if a staff member has access to a patient (for RLS)
func (r *PatientRepository) CheckStaffAccess(ctx context.Context, staffID, patientID string) (bool, error) {
	stmt := NewStatement(`SELECT COUNT(*) as count
		FROM staff_patient_assignments
		WHERE staff_id = @staffID
			AND patient_id = @patientID
			AND status = 'active'`,
		map[string]interface{}{
			"staffID":   staffID,
			"patientID": patientID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		return false, fmt.Errorf("failed to check access: %w", err)
	}

	var count int64
	if err := row.Columns(&count); err != nil {
		return false, fmt.Errorf("failed to scan count: %w", err)
	}

	return count > 0, nil
}

// scanPatient scans a Spanner row into a Patient model
func scanPatient(row *spanner.Row) (*models.Patient, error) {
	var patient models.Patient

	var nameHistoryStr, contactPointsStr, addressesStr string
	var consentDetailsStr *string

	err := row.Columns(
		&patient.PatientID,
		&patient.BirthDate,
		&patient.Gender,
		&patient.BloodType,
		&nameHistoryStr,
		&contactPointsStr,
		&addressesStr,
		&consentDetailsStr,
		&patient.CurrentFamilyName,
		&patient.CurrentGivenName,
		&patient.PrimaryPhone,
		&patient.CurrentPrefecture,
		&patient.CurrentCity,
		&patient.ConsentStatus,
		&patient.ConsentObtainedAt,
		&patient.ConsentWithdrawnAt,
		&patient.Deleted,
		&patient.DeletedAt,
		&patient.DeletedReason,
		&patient.CreatedAt,
		&patient.CreatedBy,
		&patient.UpdatedAt,
		&patient.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Parse JSONB fields
	patient.NameHistory = json.RawMessage(nameHistoryStr)
	patient.ContactPoints = json.RawMessage(contactPointsStr)
	patient.Addresses = json.RawMessage(addressesStr)
	if consentDetailsStr != nil {
		patient.ConsentDetails = json.RawMessage(*consentDetailsStr)
	}

	return &patient, nil
}
