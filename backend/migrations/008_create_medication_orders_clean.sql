-- Migration: Create medication_orders table (Emulator Compatible)
-- Removed: GENERATED columns, FOREIGN KEY, WHERE clauses in indexes, COMMENT

CREATE TABLE medication_orders (
    order_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,
    visit_schedule_id VARCHAR(36),

    status VARCHAR(30) NOT NULL DEFAULT 'active',
    intent VARCHAR(30) NOT NULL DEFAULT 'order',
    priority VARCHAR(20) NOT NULL DEFAULT 'routine',

    medication_code_system VARCHAR(50),
    medication_code VARCHAR(50),
    medication_name VARCHAR(300) NOT NULL,
    medication_generic_name VARCHAR(300),

    dosage_instruction JSONB NOT NULL,

    -- Removed GENERATED columns: dose_quantity, dose_unit, frequency

    quantity_value FLOAT8,
    quantity_unit VARCHAR(50),
    expected_supply_duration INT,
    number_of_repeats_allowed INT DEFAULT 0,

    route VARCHAR(50),
    method VARCHAR(100),

    authored_on TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    validity_period_start TIMESTAMPTZ,
    validity_period_end TIMESTAMPTZ,

    prescriber_id VARCHAR(100) NOT NULL,
    prescriber_name VARCHAR(200),
    prescriber_license_number VARCHAR(50),

    reason_code VARCHAR(50),
    reason_display VARCHAR(200),
    reason_reference VARCHAR(36),

    substitution_allowed BOOLEAN NOT NULL DEFAULT TRUE,
    substitution_reason TEXT,

    dispense_request JSONB,
    last_dispensed_date TIMESTAMPTZ,
    total_dispensed_quantity FLOAT8,

    patient_instructions TEXT,
    clinical_notes TEXT,

    insurance_type VARCHAR(30),
    billing_code VARCHAR(50),

    allergy_checked BOOLEAN NOT NULL DEFAULT FALSE,
    interaction_checked BOOLEAN NOT NULL DEFAULT FALSE,
    contraindication_checked BOOLEAN NOT NULL DEFAULT FALSE,
    check_warnings JSONB,

    stopped_at TIMESTAMPTZ,
    stopped_by VARCHAR(100),
    stop_reason TEXT,
    modified_from VARCHAR(36),

    version INT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (order_id)
);

CREATE INDEX idx_medication_orders_patient ON medication_orders(patient_id);
CREATE INDEX idx_medication_orders_visit ON medication_orders(visit_schedule_id);
CREATE INDEX idx_medication_orders_status ON medication_orders(status);
CREATE INDEX idx_medication_orders_prescriber ON medication_orders(prescriber_id);
CREATE INDEX idx_medication_orders_drug_name ON medication_orders(medication_name);
CREATE INDEX idx_medication_orders_authored ON medication_orders(patient_id, authored_on);
CREATE INDEX idx_medication_orders_modified ON medication_orders(modified_from);
