package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"google.golang.org/api/iterator"
)

// MedicationOrderRepository handles medication order data operations
type MedicationOrderRepository struct {
	spannerRepo *SpannerRepository
}

// NewMedicationOrderRepository creates a new medication order repository
func NewMedicationOrderRepository(spannerRepo *SpannerRepository) *MedicationOrderRepository {
	return &MedicationOrderRepository{
		spannerRepo: spannerRepo,
	}
}

// Create creates a new medication order
func (r *MedicationOrderRepository) Create(ctx context.Context, patientID string, req *models.MedicationOrderCreateRequest) (*models.MedicationOrder, error) {
	orderID := uuid.New().String()

	order := &models.MedicationOrder{
		OrderID:           orderID,
		PatientID:         patientID,
		Status:            req.Status,
		Intent:            req.Intent,
		Medication:        req.Medication,
		DosageInstruction: req.DosageInstruction,
		PrescribedDate:    req.PrescribedDate,
		PrescribedBy:      req.PrescribedBy,
		DispensePharmacy:  req.DispensePharmacy,
		Version:           1,
	}

	if req.ReasonReference != nil {
		order.ReasonReference = spanner.NullString{StringVal: *req.ReasonReference, Valid: true}
	}

	// Convert JSONB fields to strings for Spanner
	medicationStr := string(req.Medication)
	dosageInstructionStr := string(req.DosageInstruction)

	var dispensePharmacyStr spanner.NullString
	if len(req.DispensePharmacy) > 0 {
		dispensePharmacyStr = spanner.NullString{StringVal: string(req.DispensePharmacy), Valid: true}
	}

	mutation := spanner.Insert("medication_orders",
		[]string{
			"order_id", "patient_id", "status", "intent",
			"medication", "dosage_instruction",
			"prescribed_date", "prescribed_by",
			"dispense_pharmacy", "reason_reference", "version",
		},
		[]interface{}{
			orderID, patientID, req.Status, req.Intent,
			medicationStr, dosageInstructionStr,
			req.PrescribedDate, req.PrescribedBy,
			dispensePharmacyStr, order.ReasonReference, 1,
		},
	)

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to create medication order: %w", err)
	}

	return order, nil
}

// GetByID retrieves a medication order by ID
func (r *MedicationOrderRepository) GetByID(ctx context.Context, patientID, orderID string) (*models.MedicationOrder, error) {
	stmt := NewStatement(`SELECT
			order_id, patient_id, status, intent,
			medication, dosage_instruction,
			prescribed_date, prescribed_by,
			dispense_pharmacy, reason_reference, version
		FROM medication_orders
		WHERE patient_id = @patient_id AND order_id = @order_id`,
		map[string]interface{}{
			"patient_id": patientID,
			"order_id":   orderID,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("medication order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query medication order: %w", err)
	}

	return scanMedicationOrder(row)
}

// List retrieves medication orders with filters
func (r *MedicationOrderRepository) List(ctx context.Context, filter *models.MedicationOrderFilter) ([]*models.MedicationOrder, error) {
	var conditions []string
	params := make(map[string]interface{})

	if filter.PatientID != nil {
		conditions = append(conditions, "patient_id = @patient_id")
		params["patient_id"] = *filter.PatientID
	}

	if filter.Status != nil {
		conditions = append(conditions, "status = @status")
		params["status"] = *filter.Status
	}

	if filter.Intent != nil {
		conditions = append(conditions, "intent = @intent")
		params["intent"] = *filter.Intent
	}

	if filter.PrescribedBy != nil {
		conditions = append(conditions, "prescribed_by = @prescribed_by")
		params["prescribed_by"] = *filter.PrescribedBy
	}

	if filter.PrescribedDateFrom != nil {
		conditions = append(conditions, "prescribed_date >= @prescribed_date_from")
		params["prescribed_date_from"] = *filter.PrescribedDateFrom
	}

	if filter.PrescribedDateTo != nil {
		conditions = append(conditions, "prescribed_date <= @prescribed_date_to")
		params["prescribed_date_to"] = *filter.PrescribedDateTo
	}

	if filter.ReasonReference != nil {
		conditions = append(conditions, "reason_reference = @reason_reference")
		params["reason_reference"] = *filter.ReasonReference
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := 100
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	params["limit"] = limit
	params["offset"] = offset

	stmt := NewStatement(fmt.Sprintf(`SELECT
			order_id, patient_id, status, intent,
			medication, dosage_instruction,
			prescribed_date, prescribed_by,
			dispense_pharmacy, reason_reference, version
		FROM medication_orders
		%s
		ORDER BY prescribed_date DESC
		LIMIT @limit OFFSET @offset`, whereClause),
		params)

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var orders []*models.MedicationOrder
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate medication orders: %w", err)
		}

		order, err := scanMedicationOrder(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// Update updates a medication order
func (r *MedicationOrderRepository) Update(ctx context.Context, patientID, orderID string, req *models.MedicationOrderUpdateRequest) (*models.MedicationOrder, error) {
	// First, get the existing order
	existing, err := r.GetByID(ctx, patientID, orderID)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
		existing.Status = *req.Status
	}

	if req.Intent != nil {
		updates["intent"] = *req.Intent
		existing.Intent = *req.Intent
	}

	if len(req.Medication) > 0 {
		updates["medication"] = string(req.Medication)
		existing.Medication = req.Medication
	}

	if len(req.DosageInstruction) > 0 {
		updates["dosage_instruction"] = string(req.DosageInstruction)
		existing.DosageInstruction = req.DosageInstruction
	}

	if req.PrescribedDate != nil {
		updates["prescribed_date"] = *req.PrescribedDate
		existing.PrescribedDate = *req.PrescribedDate
	}

	if req.PrescribedBy != nil {
		updates["prescribed_by"] = *req.PrescribedBy
		existing.PrescribedBy = *req.PrescribedBy
	}

	if len(req.DispensePharmacy) > 0 {
		updates["dispense_pharmacy"] = spanner.NullString{StringVal: string(req.DispensePharmacy), Valid: true}
		existing.DispensePharmacy = req.DispensePharmacy
	}

	if req.ReasonReference != nil {
		updates["reason_reference"] = spanner.NullString{StringVal: *req.ReasonReference, Valid: true}
		existing.ReasonReference = spanner.NullString{StringVal: *req.ReasonReference, Valid: true}
	}

	if len(updates) == 0 {
		return existing, nil
	}

	// Increment version for optimistic locking
	updates["version"] = existing.Version + 1
	existing.Version++

	// Build column list and values
	columns := []string{"patient_id", "order_id"}
	values := []interface{}{patientID, orderID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("medication_orders", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update medication order: %w", err)
	}

	return existing, nil
}

// UpdateWithVersion updates a medication order with optimistic locking
func (r *MedicationOrderRepository) UpdateWithVersion(ctx context.Context, patientID, orderID string, expectedVersion int, req *models.MedicationOrderUpdateRequest) (*models.MedicationOrder, error) {
	// First, get the existing order
	existing, err := r.GetByID(ctx, patientID, orderID)
	if err != nil {
		return nil, err
	}

	// Check version for optimistic locking
	if existing.Version != expectedVersion {
		return nil, fmt.Errorf("CONFLICT: Medication order was modified by another user. Expected version %d but found %d", expectedVersion, existing.Version)
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
		existing.Status = *req.Status
	}

	if req.Intent != nil {
		updates["intent"] = *req.Intent
		existing.Intent = *req.Intent
	}

	if len(req.Medication) > 0 {
		updates["medication"] = string(req.Medication)
		existing.Medication = req.Medication
	}

	if len(req.DosageInstruction) > 0 {
		updates["dosage_instruction"] = string(req.DosageInstruction)
		existing.DosageInstruction = req.DosageInstruction
	}

	if req.PrescribedDate != nil {
		updates["prescribed_date"] = *req.PrescribedDate
		existing.PrescribedDate = *req.PrescribedDate
	}

	if req.PrescribedBy != nil {
		updates["prescribed_by"] = *req.PrescribedBy
		existing.PrescribedBy = *req.PrescribedBy
	}

	if len(req.DispensePharmacy) > 0 {
		updates["dispense_pharmacy"] = spanner.NullString{StringVal: string(req.DispensePharmacy), Valid: true}
		existing.DispensePharmacy = req.DispensePharmacy
	}

	if req.ReasonReference != nil {
		updates["reason_reference"] = spanner.NullString{StringVal: *req.ReasonReference, Valid: true}
		existing.ReasonReference = spanner.NullString{StringVal: *req.ReasonReference, Valid: true}
	}

	if len(updates) == 0 {
		return existing, nil
	}

	// Increment version
	updates["version"] = existing.Version + 1
	existing.Version++

	// Build column list and values
	columns := []string{"patient_id", "order_id"}
	values := []interface{}{patientID, orderID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("medication_orders", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update medication order: %w", err)
	}

	return existing, nil
}

// Delete deletes a medication order
func (r *MedicationOrderRepository) Delete(ctx context.Context, patientID, orderID string) error {
	mutation := spanner.Delete("medication_orders", spanner.Key{patientID, orderID})

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete medication order: %w", err)
	}

	return nil
}

// GetActiveOrders retrieves all active medication orders for a patient
func (r *MedicationOrderRepository) GetActiveOrders(ctx context.Context, patientID string) ([]*models.MedicationOrder, error) {
	stmt := NewStatement(`SELECT
			order_id, patient_id, status, intent,
			medication, dosage_instruction,
			prescribed_date, prescribed_by,
			dispense_pharmacy, reason_reference, version
		FROM medication_orders
		WHERE patient_id = @patient_id AND status = 'active'
		ORDER BY prescribed_date DESC`,
		map[string]interface{}{
			"patient_id": patientID,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var orders []*models.MedicationOrder
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate active orders: %w", err)
		}

		order, err := scanMedicationOrder(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrdersByPrescription retrieves medication orders by prescription details
func (r *MedicationOrderRepository) GetOrdersByPrescription(ctx context.Context, patientID, prescribedBy string, prescribedDate time.Time) ([]*models.MedicationOrder, error) {
	stmt := NewStatement(`SELECT
			order_id, patient_id, status, intent,
			medication, dosage_instruction,
			prescribed_date, prescribed_by,
			dispense_pharmacy, reason_reference, version
		FROM medication_orders
		WHERE patient_id = @patient_id
		  AND prescribed_by = @prescribed_by
		  AND prescribed_date = @prescribed_date
		ORDER BY order_id`,
		map[string]interface{}{
			"patient_id":      patientID,
			"prescribed_by":   prescribedBy,
			"prescribed_date": prescribedDate,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var orders []*models.MedicationOrder
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate orders by prescription: %w", err)
		}

		order, err := scanMedicationOrder(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// scanMedicationOrder scans a Spanner row into a MedicationOrder model
func scanMedicationOrder(row *spanner.Row) (*models.MedicationOrder, error) {
	var order models.MedicationOrder
	var medicationStr, dosageInstructionStr string
	var dispensePharmacyStr spanner.NullString

	err := row.Columns(
		&order.OrderID,
		&order.PatientID,
		&order.Status,
		&order.Intent,
		&medicationStr,
		&dosageInstructionStr,
		&order.PrescribedDate,
		&order.PrescribedBy,
		&dispensePharmacyStr,
		&order.ReasonReference,
		&order.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan medication order: %w", err)
	}

	// Convert JSONB strings back to json.RawMessage
	order.Medication = json.RawMessage(medicationStr)
	order.DosageInstruction = json.RawMessage(dosageInstructionStr)
	if dispensePharmacyStr.Valid {
		order.DispensePharmacy = json.RawMessage(dispensePharmacyStr.StringVal)
	}

	return &order, nil
}
