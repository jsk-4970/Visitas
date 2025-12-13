-- Migration: Create patients table
-- Cloud Spanner PostgreSQL Interface (Emulator Compatible)

CREATE TABLE patients (
    patient_id VARCHAR(36) NOT NULL,
    birth_date TIMESTAMPTZ,
    gender VARCHAR(20) NOT NULL,
    blood_type VARCHAR(10),
    name_history JSONB NOT NULL,
    contact_points JSONB,
    addresses JSONB,
    consent_details JSONB,
    current_family_name VARCHAR(100),
    current_given_name VARCHAR(100),
    primary_phone VARCHAR(50),
    current_prefecture VARCHAR(50),
    current_city VARCHAR(100),
    consent_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    consent_obtained_at TIMESTAMPTZ,
    consent_withdrawn_at TIMESTAMPTZ,
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    deleted_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100) NOT NULL,
    PRIMARY KEY (patient_id)
);

CREATE INDEX idx_patients_birth_date ON patients(birth_date);
CREATE INDEX idx_patients_deleted ON patients(deleted);
CREATE INDEX idx_patients_consent_status ON patients(consent_status);
