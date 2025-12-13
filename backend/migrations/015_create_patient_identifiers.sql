-- Migration: Create patient_identifiers table
-- Cloud Spanner PostgreSQL Interface
-- Purpose: Store patient identification numbers (insurance, My Number, MRN, etc.) with encryption support

CREATE TABLE patient_identifiers (
    identifier_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,

    -- Identifier Type (my_number, insurance_id, care_insurance_id, mrn, other)
    identifier_type VARCHAR(50) NOT NULL,

    -- Identifier Value
    -- For my_number: stores encrypted ciphertext (base64-encoded)
    -- For other types: stores plaintext value
    identifier_value TEXT NOT NULL,

    -- Primary Identifier Flag
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,

    -- Validity Period (for insurance cards with expiration dates)
    valid_from TIMESTAMPTZ,
    valid_to TIMESTAMPTZ,

    -- Issuer Information (for insurance cards)
    issuer_name VARCHAR(200),
    issuer_code VARCHAR(50),

    -- Verification Status
    verification_status VARCHAR(20) NOT NULL DEFAULT 'unverified',
    verified_at TIMESTAMPTZ,
    verified_by VARCHAR(100),

    -- Soft Delete
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    -- Audit Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100) NOT NULL,

    PRIMARY KEY (identifier_id),

    -- Foreign Key to patients table
    CONSTRAINT fk_patient_identifiers_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(patient_id)
        ON DELETE CASCADE
);

-- Indexes for common queries
CREATE INDEX idx_patient_identifiers_patient_id ON patient_identifiers(patient_id);
CREATE INDEX idx_patient_identifiers_type ON patient_identifiers(identifier_type);
CREATE INDEX idx_patient_identifiers_primary ON patient_identifiers(patient_id, is_primary) WHERE is_primary = TRUE;
CREATE INDEX idx_patient_identifiers_deleted ON patient_identifiers(deleted) WHERE deleted = FALSE;
CREATE INDEX idx_patient_identifiers_verification ON patient_identifiers(verification_status);

-- Unique constraint: only one primary identifier per patient per type
CREATE UNIQUE INDEX idx_patient_identifiers_unique_primary
    ON patient_identifiers(patient_id, identifier_type)
    WHERE is_primary = TRUE AND deleted = FALSE;

-- Comments
COMMENT ON TABLE patient_identifiers IS 'Patient identification numbers with encryption support for My Number';
COMMENT ON COLUMN patient_identifiers.identifier_value IS 'Encrypted for my_number, plaintext for others';
COMMENT ON COLUMN patient_identifiers.verification_status IS 'Values: verified, unverified, expired, invalid';
COMMENT ON COLUMN patient_identifiers.is_primary IS 'Only one primary identifier per patient per type';
