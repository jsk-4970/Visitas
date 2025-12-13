package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// MedicalRecordService handles business logic for medical records
type MedicalRecordService struct {
	medicalRecordRepo *repository.MedicalRecordRepository
	patientRepo       *repository.PatientRepository
	templateRepo      *repository.MedicalRecordTemplateRepository
}

// NewMedicalRecordService creates a new medical record service
func NewMedicalRecordService(
	medicalRecordRepo *repository.MedicalRecordRepository,
	patientRepo *repository.PatientRepository,
	templateRepo *repository.MedicalRecordTemplateRepository,
) *MedicalRecordService {
	return &MedicalRecordService{
		medicalRecordRepo: medicalRecordRepo,
		patientRepo:       patientRepo,
		templateRepo:      templateRepo,
	}
}

// CreateRecord creates a new medical record with access control
func (s *MedicalRecordService) CreateRecord(ctx context.Context, patientID string, req *models.MedicalRecordCreateRequest, createdBy string) (*models.MedicalRecord, error) {
	// Check staff access
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, createdBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized medical record creation attempt", map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to create medical records for this patient")
	}

	// Validate visit type
	validVisitTypes := map[string]bool{
		"regular":       true,
		"emergency":     true,
		"initial":       true,
		"follow_up":     true,
		"terminal_care": true,
	}
	if !validVisitTypes[req.VisitType] {
		logger.WarnContext(ctx, "Invalid visit type", map[string]interface{}{
			"visit_type": req.VisitType,
		})
		return nil, fmt.Errorf("invalid visit type: %s", req.VisitType)
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft":       true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}
	if !validStatuses[req.Status] {
		logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
			"status": req.Status,
		})
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate source type
	validSourceTypes := map[string]bool{
		"manual":       true,
		"voice_to_text": true,
		"ai_generated": true,
		"template":     true,
	}
	if !validSourceTypes[req.SourceType] {
		logger.WarnContext(ctx, "Invalid source type", map[string]interface{}{
			"source_type": req.SourceType,
		})
		return nil, fmt.Errorf("invalid source type: %s", req.SourceType)
	}

	// Validate visit_started_at is not zero
	if req.VisitStartedAt.IsZero() {
		logger.WarnContext(ctx, "Missing visit_started_at", nil)
		return nil, fmt.Errorf("visit_started_at is required")
	}

	// Validate performed_by is not empty
	if req.PerformedBy == "" {
		logger.WarnContext(ctx, "Missing performed_by", nil)
		return nil, fmt.Errorf("performed_by is required")
	}

	// Create record
	record, err := s.medicalRecordRepo.Create(ctx, patientID, req, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create medical record", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, err
	}

	// If using a template, increment usage count
	if req.TemplateID != nil {
		templateID := *req.TemplateID
		go func() {
			if err := s.templateRepo.IncrementUsageCount(context.Background(), templateID); err != nil {
				logger.Error("Failed to increment template usage count", err, map[string]interface{}{
					"template_id": templateID,
					"record_id":   record.RecordID,
				})
			}
		}()
	}

	logger.InfoContext(ctx, "Medical record created successfully", map[string]interface{}{
		"record_id":  record.RecordID,
		"patient_id": patientID,
		"created_by": createdBy,
	})

	return record, nil
}

// GetRecord retrieves a medical record with access control
func (s *MedicalRecordService) GetRecord(ctx context.Context, patientID, recordID, requestorID string) (*models.MedicalRecord, error) {
	// Check staff access
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"record_id":    recordID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized medical record access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"record_id":    recordID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this medical record")
	}

	// Get record
	record, err := s.medicalRecordRepo.GetByID(ctx, patientID, recordID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch medical record", err, map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Medical record retrieved successfully", map[string]interface{}{
		"record_id":    recordID,
		"patient_id":   patientID,
		"requestor_id": requestorID,
	})

	return record, nil
}

// ListRecords retrieves medical records with access control
func (s *MedicalRecordService) ListRecords(ctx context.Context, patientID string, filter *models.MedicalRecordFilter, requestorID string) ([]*models.MedicalRecord, error) {
	// Check staff access
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized medical records list attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view medical records for this patient")
	}

	// Set patient ID in filter
	filter.PatientID = &patientID

	// Get records
	records, err := s.medicalRecordRepo.List(ctx, filter)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list medical records", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Medical records listed successfully", map[string]interface{}{
		"patient_id":   patientID,
		"count":        len(records),
		"requestor_id": requestorID,
	})

	return records, nil
}

// UpdateRecord updates a medical record with access control and optimistic locking
func (s *MedicalRecordService) UpdateRecord(ctx context.Context, patientID, recordID string, req *models.MedicalRecordUpdateRequest, updatedBy string) (*models.MedicalRecord, error) {
	// Check staff access
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized medical record update attempt", map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this medical record")
	}

	// Get existing record for optimistic locking
	existing, err := s.medicalRecordRepo.GetByID(ctx, patientID, recordID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch medical record for update", err, map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
		})
		return nil, err
	}

	// Check version for optimistic locking if provided
	if req.ExpectedVersion != nil && existing.Version != *req.ExpectedVersion {
		logger.WarnContext(ctx, "Concurrent edit conflict detected", map[string]interface{}{
			"record_id":        recordID,
			"expected_version": *req.ExpectedVersion,
			"actual_version":   existing.Version,
		})
		return nil, fmt.Errorf("CONFLICT: Record was modified by another user. Please refresh and try again. Expected version %d but found %d", *req.ExpectedVersion, existing.Version)
	}

	// Validate visit type if provided
	if req.VisitType != nil {
		validVisitTypes := map[string]bool{
			"regular":       true,
			"emergency":     true,
			"initial":       true,
			"follow_up":     true,
			"terminal_care": true,
		}
		if !validVisitTypes[*req.VisitType] {
			return nil, fmt.Errorf("invalid visit type: %s", *req.VisitType)
		}
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"draft":       true,
			"in_progress": true,
			"completed":   true,
			"cancelled":   true,
		}
		if !validStatuses[*req.Status] {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Merge SOAP content if provided (intelligent merge)
	if len(req.SOAPContent) > 0 && len(existing.SOAPContent) > 0 {
		mergedSOAP, err := s.mergeSOAPContent(existing.SOAPContent, req.SOAPContent)
		if err != nil {
			logger.WarnContext(ctx, "Failed to merge SOAP content, using new content as-is", map[string]interface{}{
				"record_id": recordID,
				"error":     err.Error(),
			})
		} else {
			req.SOAPContent = mergedSOAP
		}
	}

	// Update record with version check
	var record *models.MedicalRecord
	if req.ExpectedVersion != nil {
		record, err = s.medicalRecordRepo.UpdateWithVersion(ctx, patientID, recordID, *req.ExpectedVersion, req, updatedBy)
	} else {
		record, err = s.medicalRecordRepo.Update(ctx, patientID, recordID, req, updatedBy)
	}

	if err != nil {
		logger.ErrorContext(ctx, "Failed to update medical record", err, map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
			"updated_by": updatedBy,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Medical record updated successfully", map[string]interface{}{
		"record_id":  recordID,
		"patient_id": patientID,
		"updated_by": updatedBy,
		"version":    record.Version,
	})

	return record, nil
}

// DeleteRecord soft-deletes a medical record with access control
func (s *MedicalRecordService) DeleteRecord(ctx context.Context, patientID, recordID, deletedBy string) error {
	// Check staff access
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized medical record deletion attempt", map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this medical record")
	}

	// Delete record
	err = s.medicalRecordRepo.Delete(ctx, patientID, recordID, deletedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete medical record", err, map[string]interface{}{
			"patient_id": patientID,
			"record_id":  recordID,
			"deleted_by": deletedBy,
		})
		return err
	}

	logger.InfoContext(ctx, "Medical record deleted successfully", map[string]interface{}{
		"record_id":  recordID,
		"patient_id": patientID,
		"deleted_by": deletedBy,
	})

	return nil
}

// CopyRecord copies an existing record to create a new one
func (s *MedicalRecordService) CopyRecord(ctx context.Context, sourcePatientID, sourceRecordID, targetPatientID string, req *models.CopyAsMedicalRecordRequest, createdBy string) (*models.MedicalRecord, error) {
	// Check access to source patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, createdBy, sourcePatientID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("access denied: you do not have permission to copy this medical record")
	}

	// Get source record
	sourceRecord, err := s.medicalRecordRepo.GetByID(ctx, sourcePatientID, sourceRecordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source record: %w", err)
	}

	// If target patient is different, check access
	if targetPatientID != sourcePatientID {
		hasAccess, err = s.patientRepo.CheckStaffAccess(ctx, createdBy, targetPatientID)
		if err != nil {
			return nil, fmt.Errorf("failed to check access: %w", err)
		}
		if !hasAccess {
			return nil, fmt.Errorf("access denied: you do not have permission to create records for target patient")
		}
	}

	// Prepare SOAP content (merge with modifications if provided)
	soapContent := sourceRecord.SOAPContent
	if len(req.ModifySOAP) > 0 {
		soapContent, err = s.mergeSOAPContent(sourceRecord.SOAPContent, req.ModifySOAP)
		if err != nil {
			soapContent = sourceRecord.SOAPContent // Fallback to original
		}
	}

	// Create new record request
	createReq := &models.MedicalRecordCreateRequest{
		VisitStartedAt: req.VisitStartedAt,
		VisitType:      req.VisitType,
		PerformedBy:    req.PerformedBy,
		Status:         "draft",
		SOAPContent:    soapContent,
		SourceRecordID: &sourceRecordID,
		SourceType:     "template",
	}

	// Create the copied record
	record, err := s.medicalRecordRepo.Create(ctx, targetPatientID, createReq, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create copied medical record", err, map[string]interface{}{
			"source_record_id": sourceRecordID,
			"target_patient_id": targetPatientID,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Medical record copied successfully", map[string]interface{}{
		"source_record_id": sourceRecordID,
		"new_record_id":    record.RecordID,
		"created_by":       createdBy,
	})

	return record, nil
}

// CreateFromTemplate creates a new record from a template
func (s *MedicalRecordService) CreateFromTemplate(ctx context.Context, patientID string, req *models.CreateFromTemplateRequest, createdBy string) (*models.MedicalRecord, error) {
	// Check staff access
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, createdBy, patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("access denied: you do not have permission to create medical records for this patient")
	}

	// Get template
	template, err := s.templateRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Initialize SOAP content from template
	soapContent := template.SOAPTemplate
	if len(req.InitialSOAP) > 0 {
		soapContent, err = s.mergeSOAPContent(template.SOAPTemplate, req.InitialSOAP)
		if err != nil {
			soapContent = template.SOAPTemplate // Fallback
		}
	}

	// Create record request
	createReq := &models.MedicalRecordCreateRequest{
		VisitStartedAt: req.VisitStartedAt,
		VisitType:      req.VisitType,
		PerformedBy:    req.PerformedBy,
		Status:         "draft",
		SOAPContent:    soapContent,
		TemplateID:     &req.TemplateID,
		SourceType:     "template",
	}

	// Create record
	record, err := s.medicalRecordRepo.Create(ctx, patientID, createReq, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create record from template", err, map[string]interface{}{
			"patient_id":  patientID,
			"template_id": req.TemplateID,
		})
		return nil, err
	}

	// Increment template usage count
	templateID := req.TemplateID
	go func() {
		if err := s.templateRepo.IncrementUsageCount(context.Background(), templateID); err != nil {
			logger.Error("Failed to increment template usage count", err, map[string]interface{}{
				"template_id": templateID,
				"record_id":   record.RecordID,
			})
		}
	}()

	logger.InfoContext(ctx, "Medical record created from template successfully", map[string]interface{}{
		"record_id":   record.RecordID,
		"patient_id":  patientID,
		"template_id": req.TemplateID,
		"created_by":  createdBy,
	})

	return record, nil
}

// GetLatestRecords retrieves the latest medical records for a patient
func (s *MedicalRecordService) GetLatestRecords(ctx context.Context, patientID string, limit int, requestorID string) ([]*models.MedicalRecord, error) {
	// Check staff access
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("access denied: you do not have permission to view medical records for this patient")
	}

	records, err := s.medicalRecordRepo.GetLatestByPatient(ctx, patientID, limit)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get latest records", err, map[string]interface{}{
			"patient_id": patientID,
		})
		return nil, err
	}

	return records, nil
}

// GetDraftRecords retrieves draft/in-progress records for a staff member
func (s *MedicalRecordService) GetDraftRecords(ctx context.Context, staffID string) ([]*models.MedicalRecord, error) {
	records, err := s.medicalRecordRepo.GetDraftRecords(ctx, staffID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get draft records", err, map[string]interface{}{
			"staff_id": staffID,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Draft records retrieved", map[string]interface{}{
		"staff_id": staffID,
		"count":    len(records),
	})

	return records, nil
}

// mergeSOAPContent performs a deep merge of SOAP content
func (s *MedicalRecordService) mergeSOAPContent(existing, updates json.RawMessage) (json.RawMessage, error) {
	var existingMap, updatesMap map[string]interface{}

	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &existingMap); err != nil {
			return updates, nil // If existing is invalid, just use updates
		}
	} else {
		existingMap = make(map[string]interface{})
	}

	if err := json.Unmarshal(updates, &updatesMap); err != nil {
		return existing, fmt.Errorf("invalid update SOAP content: %w", err)
	}

	// Deep merge - updates override existing, preserve unmodified fields
	merged := deepMerge(existingMap, updatesMap)

	result, err := json.Marshal(merged)
	if err != nil {
		return existing, fmt.Errorf("failed to marshal merged content: %w", err)
	}

	return result, nil
}

// deepMerge recursively merges two maps
func deepMerge(existing, updates map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing values
	for key, value := range existing {
		result[key] = value
	}

	// Override with updates (deep merge for nested maps)
	for key, updateValue := range updates {
		if existingValue, ok := result[key]; ok {
			// If both are maps, merge recursively
			if existingMap, existingIsMap := existingValue.(map[string]interface{}); existingIsMap {
				if updateMap, updateIsMap := updateValue.(map[string]interface{}); updateIsMap {
					result[key] = deepMerge(existingMap, updateMap)
					continue
				}
			}
		}
		// Otherwise, override with update value
		result[key] = updateValue
	}

	return result
}
