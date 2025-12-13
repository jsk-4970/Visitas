-- Migration: Create patient_coverages table
-- Cloud Spanner PostgreSQL Interface

CREATE TABLE patient_coverages (
    coverage_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,
    insurance_type VARCHAR(30) NOT NULL, -- medical, long_term_care, public_expense

    -- JSONB Details
    details JSONB NOT NULL,

    -- Generated Columns
    care_level_code VARCHAR(20) GENERATED ALWAYS AS (
        CASE
            WHEN insurance_type = 'long_term_care' THEN CAST(details->>'careLevelCode' AS VARCHAR(20))
            ELSE NULL
        END
    ) STORED,
    copay_rate INT GENERATED ALWAYS AS (
        CAST(details->>'copayRate' AS INT)
    ) STORED,

    -- Validity Period
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, expired, suspended, terminated
    priority INT NOT NULL DEFAULT 1, -- Lower number = higher priority

    -- Verification
    verification_status VARCHAR(20) NOT NULL DEFAULT 'unverified',
    verified_at TIMESTAMPTZ,
    verified_by VARCHAR(100),

    -- Audit Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),

    -- Soft Delete
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (coverage_id),
    FOREIGN KEY (patient_id) REFERENCES patients(patient_id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_coverages_patient ON patient_coverages(patient_id);
CREATE INDEX idx_coverages_type ON patient_coverages(insurance_type);
CREATE INDEX idx_coverages_status ON patient_coverages(status) WHERE status = 'active';
CREATE INDEX idx_coverages_validity ON patient_coverages(valid_from, valid_to);
CREATE INDEX idx_coverages_care_level ON patient_coverages(care_level_code) WHERE care_level_code IS NOT NULL;
CREATE INDEX idx_coverages_active ON patient_coverages(patient_id, priority)
    WHERE deleted = FALSE AND status = 'active';

-- Comments
COMMENT ON TABLE patient_coverages IS 'Patient insurance coverage (medical, long-term care, public expense)';
COMMENT ON COLUMN patient_coverages.details IS 'JSONB with insurance-type-specific details';
COMMENT ON COLUMN patient_coverages.priority IS 'Coverage priority (1 = primary, 2 = secondary, etc.)';
