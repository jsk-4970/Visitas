-- Migration: Create medical_conditions table
-- Cloud Spanner PostgreSQL Interface
-- Based on FHIR R4 Condition Resource

CREATE TABLE medical_conditions (
    condition_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,

    -- Clinical Status (FHIR)
    clinical_status VARCHAR(20) NOT NULL, -- active, recurrence, relapse, inactive, remission, resolved
    verification_status VARCHAR(30) NOT NULL, -- unconfirmed, provisional, differential, confirmed, refuted, entered-in-error

    -- Category & Severity
    category VARCHAR(30), -- problem-list-item, encounter-diagnosis
    severity VARCHAR(30), -- mild, moderate, severe, life-threatening

    -- Condition Code (ICD-10 or SNOMED CT)
    code_system VARCHAR(50), -- ICD-10, SNOMED-CT
    code VARCHAR(50), -- e.g., E11.9 for Type 2 diabetes
    display_name VARCHAR(200) NOT NULL, -- 疾患名

    -- Body Site
    body_site VARCHAR(100),

    -- Onset Information
    onset_date TIMESTAMPTZ,
    onset_age INT,
    onset_note TEXT,

    -- Abatement Information (if resolved)
    abatement_date TIMESTAMPTZ,
    abatement_note TEXT,

    -- Recorded Information
    recorded_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recorded_by VARCHAR(100),

    -- Notes
    clinical_notes TEXT,
    patient_comments TEXT,

    -- Audit Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),

    -- Soft Delete
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (condition_id),
    FOREIGN KEY (patient_id) REFERENCES patients(patient_id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_conditions_patient ON medical_conditions(patient_id);
CREATE INDEX idx_conditions_clinical_status ON medical_conditions(clinical_status);
CREATE INDEX idx_conditions_active ON medical_conditions(patient_id, clinical_status)
    WHERE deleted = FALSE AND clinical_status IN ('active', 'recurrence', 'relapse');
CREATE INDEX idx_conditions_code ON medical_conditions(code_system, code);
CREATE INDEX idx_conditions_display_name ON medical_conditions(display_name);
CREATE INDEX idx_conditions_severity ON medical_conditions(severity) WHERE severity IN ('severe', 'life-threatening');

-- Comments
COMMENT ON TABLE medical_conditions IS 'Patient medical conditions and diagnoses (FHIR Condition)';
COMMENT ON COLUMN medical_conditions.clinical_status IS 'Current clinical status per FHIR';
COMMENT ON COLUMN medical_conditions.verification_status IS 'Verification status per FHIR';
COMMENT ON COLUMN medical_conditions.code IS 'ICD-10 or SNOMED-CT code';
