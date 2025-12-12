package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
)

// MedicationOrderService handles business logic for medication orders
type MedicationOrderService struct {
	medicationOrderRepo *repository.MedicationOrderRepository
	patientRepo         *repository.PatientRepository
}

// NewMedicationOrderService creates a new medication order service
func NewMedicationOrderService(
	medicationOrderRepo *repository.MedicationOrderRepository,
	patientRepo *repository.PatientRepository,
) *MedicationOrderService {
	return &MedicationOrderService{
		medicationOrderRepo: medicationOrderRepo,
		patientRepo:         patientRepo,
	}
}

// CreateMedicationOrder creates a new medication order with validation
func (s *MedicationOrderService) CreateMedicationOrder(ctx context.Context, patientID string, req *models.MedicationOrderCreateRequest) (*models.MedicationOrder, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":            true,
		"on-hold":           true,
		"cancelled":         true,
		"completed":         true,
		"entered-in-error":  true,
	}
	if !validStatuses[req.Status] {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate intent
	validIntents := map[string]bool{
		"order": true,
		"plan":  true,
	}
	if !validIntents[req.Intent] {
		return nil, fmt.Errorf("invalid intent: %s", req.Intent)
	}

	// Validate medication is valid JSON
	if len(req.Medication) == 0 {
		return nil, fmt.Errorf("medication is required")
	}

	// Validate dosage_instruction is valid JSON
	if len(req.DosageInstruction) == 0 {
		return nil, fmt.Errorf("dosage_instruction is required")
	}

	// Validate prescribed_by
	if req.PrescribedBy == "" {
		return nil, fmt.Errorf("prescribed_by is required")
	}

	return s.medicationOrderRepo.Create(ctx, patientID, req)
}

// GetMedicationOrder retrieves a medication order by ID
func (s *MedicationOrderService) GetMedicationOrder(ctx context.Context, patientID, orderID string) (*models.MedicationOrder, error) {
	return s.medicationOrderRepo.GetByID(ctx, patientID, orderID)
}

// ListMedicationOrders lists medication orders with filters
func (s *MedicationOrderService) ListMedicationOrders(ctx context.Context, filter *models.MedicationOrderFilter) ([]*models.MedicationOrder, error) {
	// Validate status if provided
	if filter.Status != nil {
		validStatuses := map[string]bool{
			"active":            true,
			"on-hold":           true,
			"cancelled":         true,
			"completed":         true,
			"entered-in-error":  true,
		}
		if !validStatuses[*filter.Status] {
			return nil, fmt.Errorf("invalid status: %s", *filter.Status)
		}
	}

	// Validate intent if provided
	if filter.Intent != nil {
		validIntents := map[string]bool{
			"order": true,
			"plan":  true,
		}
		if !validIntents[*filter.Intent] {
			return nil, fmt.Errorf("invalid intent: %s", *filter.Intent)
		}
	}

	return s.medicationOrderRepo.List(ctx, filter)
}

// UpdateMedicationOrder updates a medication order with validation
func (s *MedicationOrderService) UpdateMedicationOrder(ctx context.Context, patientID, orderID string, req *models.MedicationOrderUpdateRequest) (*models.MedicationOrder, error) {
	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"active":            true,
			"on-hold":           true,
			"cancelled":         true,
			"completed":         true,
			"entered-in-error":  true,
		}
		if !validStatuses[*req.Status] {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Validate intent if provided
	if req.Intent != nil {
		validIntents := map[string]bool{
			"order": true,
			"plan":  true,
		}
		if !validIntents[*req.Intent] {
			return nil, fmt.Errorf("invalid intent: %s", *req.Intent)
		}
	}

	return s.medicationOrderRepo.Update(ctx, patientID, orderID, req)
}

// DeleteMedicationOrder deletes a medication order
func (s *MedicationOrderService) DeleteMedicationOrder(ctx context.Context, patientID, orderID string) error {
	// Verify the order exists before deletion
	_, err := s.medicationOrderRepo.GetByID(ctx, patientID, orderID)
	if err != nil {
		return fmt.Errorf("medication order not found: %w", err)
	}

	return s.medicationOrderRepo.Delete(ctx, patientID, orderID)
}

// GetActiveOrders retrieves all active medication orders for a patient
func (s *MedicationOrderService) GetActiveOrders(ctx context.Context, patientID string) ([]*models.MedicationOrder, error) {
	// Verify patient exists
	_, err := s.patientRepo.GetPatientByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	return s.medicationOrderRepo.GetActiveOrders(ctx, patientID)
}
