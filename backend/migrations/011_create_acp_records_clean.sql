-- Migration: Create acp_records table (Emulator Compatible)
-- Aligned with models/acp_record.go

CREATE TABLE acp_records (
    acp_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,
    recorded_date TIMESTAMPTZ NOT NULL,
    version INT NOT NULL DEFAULT 1,
    status VARCHAR(30) NOT NULL DEFAULT 'draft',

    -- Decision maker
    decision_maker VARCHAR(50) NOT NULL,
    proxy_person_id VARCHAR(100),

    -- ACP content
    directives JSONB NOT NULL,
    values_narrative TEXT,

    -- Legal documents
    legal_documents JSONB,

    -- ACP process
    discussion_log JSONB,

    -- Security
    data_sensitivity VARCHAR(50) NOT NULL DEFAULT 'highly_confidential',
    access_restricted_to JSONB,

    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (acp_id)
);

CREATE INDEX idx_acp_patient ON acp_records(patient_id);
CREATE INDEX idx_acp_status ON acp_records(status);
CREATE INDEX idx_acp_recorded_date ON acp_records(patient_id, recorded_date);
