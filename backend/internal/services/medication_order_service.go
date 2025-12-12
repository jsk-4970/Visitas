package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
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

// CreateMedicationOrder creates a new medication order with validation and access control
func (s *MedicationOrderService) CreateMedicationOrder(ctx context.Context, patientID string, req *models.MedicationOrderCreateRequest, createdBy string) (*models.MedicationOrder, error) {
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
		logger.WarnContext(ctx, "Unauthorized medication order creation attempt", map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to create medication orders for this patient")
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
		logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
			"status": req.Status,
		})
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate intent
	validIntents := map[string]bool{
		"order": true,
		"plan":  true,
	}
	if !validIntents[req.Intent] {
		logger.WarnContext(ctx, "Invalid intent", map[string]interface{}{
			"intent": req.Intent,
		})
		return nil, fmt.Errorf("invalid intent: %s", req.Intent)
	}

	// Validate medication is valid JSON
	if len(req.Medication) == 0 {
		logger.WarnContext(ctx, "Missing medication", nil)
		return nil, fmt.Errorf("medication is required")
	}

	// Validate dosage_instruction is valid JSON
	if len(req.DosageInstruction) == 0 {
		logger.WarnContext(ctx, "Missing dosage instruction", nil)
		return nil, fmt.Errorf("dosage_instruction is required")
	}

	// Validate prescribed_by
	if req.PrescribedBy == "" {
		logger.WarnContext(ctx, "Missing prescribed_by", nil)
		return nil, fmt.Errorf("prescribed_by is required")
	}

	order, err := s.medicationOrderRepo.Create(ctx, patientID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create medication order", err, map[string]interface{}{
			"patient_id": patientID,
			"created_by": createdBy,
		})
		return nil, fmt.Errorf("failed to create medication order: %w", err)
	}

	logger.InfoContext(ctx, "Medication order created successfully", map[string]interface{}{
		"order_id":     order.OrderID,
		"patient_id":   order.PatientID,
		"prescribed_by": req.PrescribedBy,
		"created_by":   createdBy,
	})

	return order, nil
}

// GetMedicationOrder retrieves a medication order by ID with access control
func (s *MedicationOrderService) GetMedicationOrder(ctx context.Context, patientID, orderID, requestorID string) (*models.MedicationOrder, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, requestorID, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id":   patientID,
			"order_id":     orderID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized medication order access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"order_id":     orderID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view this medication order")
	}

	return s.medicationOrderRepo.GetByID(ctx, patientID, orderID)
}

// ListMedicationOrders lists medication orders with filters and access control
func (s *MedicationOrderService) ListMedicationOrders(ctx context.Context, filter *models.MedicationOrderFilter, requestorID string) ([]*models.MedicationOrder, error) {
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
			logger.WarnContext(ctx, "Unauthorized medication orders list attempt", map[string]interface{}{
				"patient_id":   *filter.PatientID,
				"requestor_id": requestorID,
			})
			return nil, fmt.Errorf("access denied: you do not have permission to view medication orders for this patient")
		}
	}

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
			logger.WarnContext(ctx, "Invalid status filter", map[string]interface{}{
				"status": *filter.Status,
			})
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
			logger.WarnContext(ctx, "Invalid intent filter", map[string]interface{}{
				"intent": *filter.Intent,
			})
			return nil, fmt.Errorf("invalid intent: %s", *filter.Intent)
		}
	}

	return s.medicationOrderRepo.List(ctx, filter)
}

// UpdateMedicationOrder updates a medication order with validation and access control
func (s *MedicationOrderService) UpdateMedicationOrder(ctx context.Context, patientID, orderID string, req *models.MedicationOrderUpdateRequest, updatedBy string) (*models.MedicationOrder, error) {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, updatedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"order_id":   orderID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized medication order update attempt", map[string]interface{}{
			"patient_id": patientID,
			"order_id":   orderID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to update this medication order")
	}

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
			logger.WarnContext(ctx, "Invalid status", map[string]interface{}{
				"status": *req.Status,
			})
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
			logger.WarnContext(ctx, "Invalid intent", map[string]interface{}{
				"intent": *req.Intent,
			})
			return nil, fmt.Errorf("invalid intent: %s", *req.Intent)
		}
	}

	order, err := s.medicationOrderRepo.Update(ctx, patientID, orderID, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update medication order", err, map[string]interface{}{
			"patient_id": patientID,
			"order_id":   orderID,
			"updated_by": updatedBy,
		})
		return nil, fmt.Errorf("failed to update medication order: %w", err)
	}

	logger.InfoContext(ctx, "Medication order updated successfully", map[string]interface{}{
		"order_id":   order.OrderID,
		"patient_id": order.PatientID,
		"updated_by": updatedBy,
	})

	return order, nil
}

// DeleteMedicationOrder deletes a medication order with access control
func (s *MedicationOrderService) DeleteMedicationOrder(ctx context.Context, patientID, orderID, deletedBy string) error {
	// Check if requestor has access to this patient
	hasAccess, err := s.patientRepo.CheckStaffAccess(ctx, deletedBy, patientID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check staff access", err, map[string]interface{}{
			"patient_id": patientID,
			"order_id":   orderID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to check access: %w", err)
	}

	if !hasAccess {
		logger.WarnContext(ctx, "Unauthorized medication order deletion attempt", map[string]interface{}{
			"patient_id": patientID,
			"order_id":   orderID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("access denied: you do not have permission to delete this medication order")
	}

	// Verify the order exists before deletion
	_, err = s.medicationOrderRepo.GetByID(ctx, patientID, orderID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get medication order for deletion", err, map[string]interface{}{
			"patient_id": patientID,
			"order_id":   orderID,
		})
		return fmt.Errorf("medication order not found: %w", err)
	}

	err = s.medicationOrderRepo.Delete(ctx, patientID, orderID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete medication order", err, map[string]interface{}{
			"patient_id": patientID,
			"order_id":   orderID,
			"deleted_by": deletedBy,
		})
		return fmt.Errorf("failed to delete medication order: %w", err)
	}

	logger.InfoContext(ctx, "Medication order deleted successfully", map[string]interface{}{
		"patient_id": patientID,
		"order_id":   orderID,
		"deleted_by": deletedBy,
	})

	return nil
}

// GetActiveOrders retrieves all active medication orders for a patient with access control
func (s *MedicationOrderService) GetActiveOrders(ctx context.Context, patientID, requestorID string) ([]*models.MedicationOrder, error) {
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
		logger.WarnContext(ctx, "Unauthorized active medication orders access attempt", map[string]interface{}{
			"patient_id":   patientID,
			"requestor_id": requestorID,
		})
		return nil, fmt.Errorf("access denied: you do not have permission to view medication orders for this patient")
	}

	return s.medicationOrderRepo.GetActiveOrders(ctx, patientID)
}
