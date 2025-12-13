-- Migration: Create staff_patient_assignments table
-- Cloud Spanner PostgreSQL Interface (Emulator Compatible)

CREATE TABLE staff_patient_assignments (
    assignment_id VARCHAR(36) NOT NULL,
    staff_id VARCHAR(100) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,
    role VARCHAR(50) NOT NULL,
    assignment_type VARCHAR(50) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    assigned_at TIMESTAMPTZ NOT NULL,
    assigned_by VARCHAR(100) NOT NULL,
    inactivated_at TIMESTAMPTZ,
    inactivated_by VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (assignment_id)
);

CREATE INDEX idx_spa_staff_id ON staff_patient_assignments(staff_id);
CREATE INDEX idx_spa_patient_id ON staff_patient_assignments(patient_id);
CREATE INDEX idx_spa_status ON staff_patient_assignments(status);
CREATE INDEX idx_spa_role ON staff_patient_assignments(role);
