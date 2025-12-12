package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// ACPRecordService handles business logic for ACP records
type ACPRecordService struct {
	acpRecordRepo *repository.ACPRecordRepository
	patientRepo   *repository.PatientRepository
}

// NewACPRecordService creates a new ACP record service
func NewACPRecordService(
	acpRecordRepo *repository.ACPRecordRepository,
	patientRepo *repository.PatientRepository,
) *ACPRecordService {
	return &ACPRecordService{
		acpRecordRepo: acpRecordRepo,
		patientRepo:   patientRepo,
	}
}

// CreateACPRecord creates a new ACP record with validation and access control
func (s *ACPRecordService) CreateACPRecord(ctx context.Context, patientID string, req *models.ACPRecordCreateRequest, createdBy string) (*models.ACPRecord, error) {
	// Check if user has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, createdBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized ACP record creation attempt", map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to create ACP records for this patient")
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft":      true,
		"active":     true,
		"superseded": true,
	}
	if !validStatuses[req.Status] {
		logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
			"status": req.Status,
		})
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate decision_maker
	validDecisionMakers := map[string]bool{
		"patient":  true,
		"proxy":    true,
		"guardian": true,
	}
	if !validDecisionMakers[req.DecisionMaker] {
		logger.WarnContext(ctx, "Invalid decision_maker", map[string]interface{}{
			"decision_maker": req.DecisionMaker,
		})
		return nil, fmt.Errorf("invalid decision_maker: %s", req.DecisionMaker)
	}

	// If decision_maker is proxy or guardian, proxy_person_id should be provided
	if (req.DecisionMaker == "proxy" || req.DecisionMaker == "guardian") && req.ProxyPersonID == nil {
		logger.WarnContext(ctx, "Missing proxy_person_id", map[string]interface{}{
			"decision_maker": req.DecisionMaker,
		})
		return nil, fmt.Errorf("proxy_person_id is required when decision_maker is %s", req.DecisionMaker)
	}

	// Validate data_sensitivity if provided
	if req.DataSensitivity != nil {
		validSensitivities := map[string]bool{
			"highly_confidential": true,
			"confidential":        true,
			"restricted":          true,
		}
		if !validSensitivities[*req.DataSensitivity] {
			logger.WarnContext(ctx, "Invalid data_sensitivity", map[string]interface{}{
				"data_sensitivity": *req.DataSensitivity,
			})
			return nil, fmt.Errorf("invalid data_sensitivity: %s", *req.DataSensitivity)
		}
	}

	// Validate directives is valid JSON
	if len(req.Directives) == 0 {
		logger.WarnContext(ctx, "Missing directives", nil)
		return nil, fmt.Errorf("directives is required")
	}

	record, err := s.acpRecordRepo.Create(ctx, patientID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create ACP record", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create ACP record: %w", err)
	}

	logger.InfoContext(ctx, "ACP record created successfully", map[string]interface{}{
		"acp_id":         record.ACPID,
		"patient_id":     record.PatientID,
		"decision_maker": record.DecisionMaker,
		"created_by":     createdBy,
	})

	return record, nil
}

// GetACPRecord retrieves an ACP record by ID with access control
func (s *ACPRecordService) GetACPRecord(ctx context.Context, patientID, acpID, requestorID string) (*models.ACPRecord, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"acp_id":       acpID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized ACP record access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"acp_id":       acpID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this ACP record")
	}

	return s.acpRecordRepo.GetByID(ctx, patientID, acpID)
}

// ListACPRecords lists ACP records with filters and access control
func (s *ACPRecordService) ListACPRecords(ctx context.Context, filter *models.ACPRecordFilter, requestorID string) ([]*models.ACPRecord, error) {
	// If filtering by patient ID, check access
	if filter.PatientID != nil {
		hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, *filter.PatientID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
				"patient_id":   *filter.PatientID,
				"requestor_id": requestorID,
			})
			return nil, fmt.Errorf("failed to check access: %w", err)
		}

		if !hasAccess {
			logger.WarnContext(ctx, "Unauthorized ACP records list attempt", map[string]interface{}{
				"patient_id":   *filter.PatientID,
				"requestor_id": requestorID,
			})
			return nil, fmt.Errorf("access denied: you do not have permission to view ACP records for this patient")
		}
	}

	// Validate status filter if provided
	if filter.Status != nil {
		validStatuses := map[string]bool{
			"draft":      true,
			"active":     true,
			"superseded": true,
		}
		if !validStatuses[*filter.Status] {
			logger.WarnContext(ctx, "Invalid status filter", map[string]interface{}{
				"status": *filter.Status,
			})
			return nil, fmt.Errorf("invalid status filter: %s", *filter.Status)
		}
	}

	// Validate decision_maker filter if provided
	if filter.DecisionMaker != nil {
		validDecisionMakers := map[string]bool{
			"patient":  true,
			"proxy":    true,
			"guardian": true,
		}
		if !validDecisionMakers[*filter.DecisionMaker] {
			logger.WarnContext(ctx, "Invalid decision_maker filter", map[string]interface{}{
				"decision_maker": *filter.DecisionMaker,
			})
			return nil, fmt.Errorf("invalid decision_maker filter: %s", *filter.DecisionMaker)
		}
	}

	return s.acpRecordRepo.List(ctx, filter)
}

// UpdateACPRecord updates an ACP record with validation and access control
func (s *ACPRecordService) UpdateACPRecord(ctx context.Context, patientID, acpID string, req *models.ACPRecordUpdateRequest, updatedBy string) (*models.ACPRecord, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"acp_id":     acpID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized ACP record update attempt", map[string]interface{}{
			"patient_id": patientID,
			"acp_id":     acpID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this ACP record")
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"draft":      true,
			"active":     true,
			"superseded": true,
		}
		if !validStatuses[*req.Status] {
			logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
				"status": *req.Status,
			})
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Validate decision_maker if provided
	if req.DecisionMaker != nil {
		validDecisionMakers := map[string]bool{
			"patient":  true,
			"proxy":    true,
			"guardian": true,
		}
		if !validDecisionMakers[*req.DecisionMaker] {
			logger.WarnContext(ctx, "Invalid decision_maker", map[string]interface{}{
				"decision_maker": *req.DecisionMaker,
			})
			return nil, fmt.Errorf("invalid decision_maker: %s", *req.DecisionMaker)
		}
	}

	// Validate data_sensitivity if provided
	if req.DataSensitivity != nil {
		validSensitivities := map[string]bool{
			"highly_confidential": true,
			"confidential":        true,
			"restricted":          true,
		}
		if !validSensitivities[*req.DataSensitivity] {
			logger.WarnContext(ctx, "Invalid data_sensitivity", map[string]interface{}{
				"data_sensitivity": *req.DataSensitivity,
			})
			return nil, fmt.Errorf("invalid data_sensitivity: %s", *req.DataSensitivity)
		}
	}

	record, err := s.acpRecordRepo.Update(ctx, patientID, acpID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update ACP record", err, map[string]interface{}{
			"patient_id": patientID,
			"acp_id":     acpID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to update ACP record: %w", err)
	}

	logger.InfoContext(ctx, "ACP record updated successfully", map[string]interface{}{
		"acp_id":     record.ACPID,
		"patient_id": record.PatientID,
		"updated_by": updatedBy,
	})

	return record, nil
}

// DeleteACPRecord deletes an ACP record with access control
func (s *ACPRecordService) DeleteACPRecord(ctx context.Context, patientID, acpID, deletedBy string) error {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"acp_id":     acpID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized ACP record deletion attempt", map[string]interface{}{
			"patient_id": patientID,
			"acp_id":     acpID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this ACP record")
	}

	// Verify the record exists before deletion
	_, err = s.acpRecordRepo.GetByID(ctx, patientID, acpID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get ACP record for deletion", err, map[string]interface{}{
			"patient_id": patientID,
			"acp_id":     acpID,
		})
		return fmt.Errorf("ACP record not found: %w", err)
	}

	err = s.acpRecordRepo.Delete(ctx, patientID, acpID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete ACP record", err, map[string]interface{}{
			"patient_id": patientID,
			"acp_id":     acpID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to delete ACP record: %w", err)
	}

	logger.InfoContext(ctx, "ACP record deleted successfully", map[string]interface{}{
		"patient_id": patientID,
		"acp_id":     acpID,
		"deleted_by": deletedBy,
	})

	return nil
}

// GetLatestACP retrieves the latest active ACP record for a patient with access control
func (s *ACPRecordService) GetLatestACP(ctx context.Context, patientID, requestorID string) (*models.ACPRecord, error) {
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
		logger.WarnContext(ctx, "Unauthorized latest ACP access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view ACP records for this patient")
	}

	return s.acpRecordRepo.GetLatestACP(ctx, patientID)
}

// GetACPHistory retrieves the complete history of ACP records for a patient with access control
func (s *ACPRecordService) GetACPHistory(ctx context.Context, patientID, requestorID string) ([]*models.ACPRecord, error) {
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
		logger.WarnContext(ctx, "Unauthorized ACP history access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view ACP history for this patient")
	}

	return s.acpRecordRepo.GetACPHistory(ctx, patientID)
}
