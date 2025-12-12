package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
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

// CreateACPRecord creates a new ACP record with validation
func (s *ACPRecordService) CreateACPRecord(ctx context.Context, patientID string, req *models.ACPRecordCreateRequest) (*models.ACPRecord, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft":      true,
		"active":     true,
		"superseded": true,
	}
	if !validStatuses[req.Status] {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate decision_maker
	validDecisionMakers := map[string]bool{
		"patient":  true,
		"proxy":    true,
		"guardian": true,
	}
	if !validDecisionMakers[req.DecisionMaker] {
		return nil, fmt.Errorf("invalid decision_maker: %s", req.DecisionMaker)
	}

	// If decision_maker is proxy or guardian, proxy_person_id should be provided
	if (req.DecisionMaker == "proxy" || req.DecisionMaker == "guardian") && req.ProxyPersonID == nil {
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
			return nil, fmt.Errorf("invalid data_sensitivity: %s", *req.DataSensitivity)
		}
	}

	// Validate directives is valid JSON
	if len(req.Directives) == 0 {
		return nil, fmt.Errorf("directives is required")
	}

	return s.acpRecordRepo.Create(ctx, patientID, req)
}

// GetACPRecord retrieves an ACP record by ID
func (s *ACPRecordService) GetACPRecord(ctx context.Context, patientID, acpID string) (*models.ACPRecord, error) {
	return s.acpRecordRepo.GetByID(ctx, patientID, acpID)
}

// ListACPRecords lists ACP records with filters
func (s *ACPRecordService) ListACPRecords(ctx context.Context, filter *models.ACPRecordFilter) ([]*models.ACPRecord, error) {
	// Validate status filter if provided
	if filter.Status != nil {
		validStatuses := map[string]bool{
			"draft":      true,
			"active":     true,
			"superseded": true,
		}
		if !validStatuses[*filter.Status] {
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
			return nil, fmt.Errorf("invalid decision_maker filter: %s", *filter.DecisionMaker)
		}
	}

	return s.acpRecordRepo.List(ctx, filter)
}

// UpdateACPRecord updates an ACP record with validation
func (s *ACPRecordService) UpdateACPRecord(ctx context.Context, patientID, acpID string, req *models.ACPRecordUpdateRequest) (*models.ACPRecord, error) {
	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"draft":      true,
			"active":     true,
			"superseded": true,
		}
		if !validStatuses[*req.Status] {
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
			return nil, fmt.Errorf("invalid data_sensitivity: %s", *req.DataSensitivity)
		}
	}

	return s.acpRecordRepo.Update(ctx, patientID, acpID, req)
}

// DeleteACPRecord deletes an ACP record
func (s *ACPRecordService) DeleteACPRecord(ctx context.Context, patientID, acpID string) error {
	// Verify the record exists before deletion
	_, err := s.acpRecordRepo.GetByID(ctx, patientID, acpID)
	if err != nil {
		return fmt.Errorf("ACP record not found: %w", err)
	}

	return s.acpRecordRepo.Delete(ctx, patientID, acpID)
}

// GetLatestACP retrieves the latest active ACP record for a patient
func (s *ACPRecordService) GetLatestACP(ctx context.Context, patientID string) (*models.ACPRecord, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	return s.acpRecordRepo.GetLatestACP(ctx, patientID)
}

// GetACPHistory retrieves the complete history of ACP records for a patient
func (s *ACPRecordService) GetACPHistory(ctx context.Context, patientID string) ([]*models.ACPRecord, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	return s.acpRecordRepo.GetACPHistory(ctx, patientID)
}
