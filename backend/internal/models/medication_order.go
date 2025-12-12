package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// MedicationOrder represents a medication order for a patient
// FHIR R4 MedicationRequest resource mapping
type MedicationOrder struct {
	OrderID      string    `json:"order_id"`
	PatientID    string    `json:"patient_id"`
	Status       string    `json:"status"` // "active" | "on-hold" | "cancelled" | "completed" | "entered-in-error"
	Intent       string    `json:"intent"` // "order" | "plan"

	// Medication information (YJ code, generic name, brand name)
	Medication         json.RawMessage `json:"medication"`
	// Dosage and administration (FHIR DosageInstruction compliant)
	DosageInstruction  json.RawMessage `json:"dosage_instruction"`

	PrescribedDate     time.Time       `json:"prescribed_date"`
	PrescribedBy       string          `json:"prescribed_by"` // Physician ID

	// Dispensing pharmacy information
	DispensePharmacy   json.RawMessage `json:"dispense_pharmacy,omitempty"`

	// Prescription reason (reference to condition ID)
	ReasonReference    sql.NullString  `json:"reason_reference,omitempty"`
}

// MedicationOrderCreateRequest represents the request body for creating a medication order
type MedicationOrderCreateRequest struct {
	Status             string          `json:"status" validate:"required,oneof=active on-hold cancelled completed entered-in-error"`
	Intent             string          `json:"intent" validate:"required,oneof=order plan"`
	Medication         json.RawMessage `json:"medication" validate:"required"`
	DosageInstruction  json.RawMessage `json:"dosage_instruction" validate:"required"`
	PrescribedDate     time.Time       `json:"prescribed_date" validate:"required"`
	PrescribedBy       string          `json:"prescribed_by" validate:"required"`
	DispensePharmacy   json.RawMessage `json:"dispense_pharmacy,omitempty"`
	ReasonReference    *string         `json:"reason_reference,omitempty"`
}

// MedicationOrderUpdateRequest represents the request body for updating a medication order
type MedicationOrderUpdateRequest struct {
	Status             *string         `json:"status,omitempty" validate:"omitempty,oneof=active on-hold cancelled completed entered-in-error"`
	Intent             *string         `json:"intent,omitempty" validate:"omitempty,oneof=order plan"`
	Medication         json.RawMessage `json:"medication,omitempty"`
	DosageInstruction  json.RawMessage `json:"dosage_instruction,omitempty"`
	PrescribedDate     *time.Time      `json:"prescribed_date,omitempty"`
	PrescribedBy       *string         `json:"prescribed_by,omitempty"`
	DispensePharmacy   json.RawMessage `json:"dispense_pharmacy,omitempty"`
	ReasonReference    *string         `json:"reason_reference,omitempty"`
}

// MedicationOrderFilter represents filter options for listing medication orders
type MedicationOrderFilter struct {
	PatientID           *string
	Status              *string
	Intent              *string
	PrescribedBy        *string
	PrescribedDateFrom  *time.Time
	PrescribedDateTo    *time.Time
	ReasonReference     *string
	Limit               int
	Offset              int
}
