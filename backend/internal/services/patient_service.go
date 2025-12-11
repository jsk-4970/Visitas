package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// PatientService handles business logic for patient operations
type PatientService struct {
	patientRepo    *repository.PatientRepository
	assignmentRepo *repository.AssignmentRepository
	auditRepo      *repository.AuditRepository
}

// NewPatientService creates a new patient service
func NewPatientService(
	patientRepo *repository.PatientRepository,
	assignmentRepo *repository.AssignmentRepository,
	auditRepo *repository.AuditRepository,
) *PatientService {
	return &PatientService{
		patientRepo:    patientRepo,
		assignmentRepo: assignmentRepo,
		auditRepo:      auditRepo,
	}
}

// CreatePatient creates a new patient
func (s *PatientService) CreatePatient(ctx context.Context, req *models.PatientCreateRequest, createdBy string) (*models.Patient, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		logger.WarnContext(ctx, "Invalid patient create request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Create patient
	patient, err := s.patientRepo.CreatePatient(ctx, req, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create patient", err, map[string]interface{}{
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create patient: %w", err)
	}

	logger.InfoContext(ctx, "Patient created successfully", map[string]interface{}{
		"patient_id": patient.PatientID,
		"created_by": createdBy,
	})

	return patient, nil
}

// GetPatient retrieves a patient by ID with access control
func (s *PatientService) GetPatient(ctx context.Context, patientID, requestorID string) (*models.Patient, error) {
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
		logger.WarnContext(ctx, "Unauthorized access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this patient")
	}

	// Get patient
	patient, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get patient", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, fmt.Errorf("failed to get patient: %w", err)
	}

	return patient, nil
}

// GetMyPatients retrieves all patients assigned to a staff member
func (s *PatientService) GetMyPatients(ctx context.Context, staffID string, page, perPage int) (*models.PatientListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	patients, total, err := s.patientRepo.GetPatientsByStaffID(ctx, staffID, perPage, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get patients for staff", err, map[string]interface{}{
			"staff_id": staffID,
		})
		return nil, fmt.Errorf("failed to get patients: %w", err)
	}

	// Convert []*Patient to []Patient
	patientList := make([]models.Patient, len(patients))
	for i, p := range patients {
		patientList[i] = *p
	}

	totalPages := (total + perPage - 1) / perPage

	response := &models.PatientListResponse{
		Patients:   patientList,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}

	return response, nil
}

// UpdatePatient updates a patient with access control
func (s *PatientService) UpdatePatient(ctx context.Context, patientID string, req *models.PatientUpdateRequest, updatedBy string) (*models.Patient, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized update attempt", map[string]interface{}{
			"patient_id": patientID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this patient")
	}

	// Update patient
	patient, err := s.patientRepo.UpdatePatient(ctx, patientID, req, updatedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update patient", err, map[string]interface{}{
			"patient_id": patientID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to update patient: %w", err)
	}

	logger.InfoContext(ctx, "Patient updated successfully", map[string]interface{}{
		"patient_id": patient.PatientID,
		"updated_by": updatedBy,
	})

	return patient, nil
}

// DeletePatient soft deletes a patient with access control
func (s *PatientService) DeletePatient(ctx context.Context, patientID, deletedBy, reason string) error {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized delete attempt", map[string]interface{}{
			"patient_id": patientID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this patient")
	}

	// Delete patient
	err = s.patientRepo.DeletePatient(ctx, patientID, deletedBy, reason)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete patient", err, map[string]interface{}{
			"patient_id": patientID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to delete patient: %w", err)
	}

	logger.InfoContext(ctx, "Patient deleted successfully", map[string]interface{}{
		"patient_id": patientID,
		"deleted_by": deletedBy,
		"reason":     reason,
	})

	return nil
}

// AssignPatientToStaff assigns a patient to a staff member
func (s *PatientService) AssignPatientToStaff(ctx context.Context, patientID, staffID string, role repository.StaffRole, assignmentType repository.AssignmentType, assignedBy string) error {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Patient not found for assignment", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return fmt.Errorf("patient not found: %w", err)
	}

	// Create assignment
	_, err = s.assignmentRepo.CreateAssignment(ctx, staffID, patientID, role, assignmentType, assignedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create assignment", err, map[string]interface{}{
			"patient_id": patientID,
			"staff_id":   staffID,
			"role":       role,
		})
		return fmt.Errorf("failed to assign patient to staff: %w", err)
	}

	logger.InfoContext(ctx, "Patient assigned to staff successfully", map[string]interface{}{
		"patient_id":      patientID,
		"staff_id":        staffID,
		"role":            role,
		"assignment_type": assignmentType,
		"assigned_by":     assignedBy,
	})

	return nil
}

// validateCreateRequest validates patient create request
func (s *PatientService) validateCreateRequest(req *models.PatientCreateRequest) error {
	if req.BirthDate == "" {
		return fmt.Errorf("birth_date is required")
	}

	if req.Name.Family == "" || req.Name.Given == "" {
		return fmt.Errorf("family name and given name are required")
	}

	if len(req.ContactPoints) == 0 {
		return fmt.Errorf("at least one contact point is required")
	}

	if len(req.Addresses) == 0 {
		return fmt.Errorf("at least one address is required")
	}

	if req.ConsentStatus == "" {
		return fmt.Errorf("consent_status is required")
	}

	// Validate consent status
	validConsentStatuses := []string{"obtained", "not_obtained", "conditional"}
	isValid := false
	for _, status := range validConsentStatuses {
		if req.ConsentStatus == status {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid consent_status: must be one of [obtained, not_obtained, conditional]")
	}

	return nil
}
