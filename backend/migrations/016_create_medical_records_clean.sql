-- Migration: Create medical_records and medical_record_templates tables
-- Cloud Spanner PostgreSQL Interface (Emulator Compatible)
-- Phase 1 Sprint 6: 基本カルテ機能

-- Table: medical_record_templates
CREATE TABLE medical_record_templates (
    template_id VARCHAR(36) NOT NULL,
    template_name VARCHAR(200) NOT NULL,
    template_description TEXT,
    specialty VARCHAR(50),
    soap_template JSONB NOT NULL,
    is_system_template BOOLEAN NOT NULL DEFAULT FALSE,
    usage_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(36) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(36),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (template_id)
);

-- Indexes for templates
CREATE INDEX idx_templates_specialty ON medical_record_templates(specialty);
CREATE INDEX idx_templates_system ON medical_record_templates(is_system_template);
CREATE INDEX idx_templates_creator ON medical_record_templates(created_by);
CREATE INDEX idx_templates_usage ON medical_record_templates(usage_count DESC);

-- Table: medical_records
CREATE TABLE medical_records (
    record_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,
    visit_started_at TIMESTAMPTZ NOT NULL,
    visit_ended_at TIMESTAMPTZ,
    visit_type VARCHAR(50) NOT NULL,
    performed_by VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    schedule_id VARCHAR(36),
    soap_content JSONB,
    template_id VARCHAR(36),
    source_record_id VARCHAR(36),
    source_type VARCHAR(20) NOT NULL DEFAULT 'manual',
    audio_file_url TEXT,
    soap_completed BOOLEAN NOT NULL DEFAULT FALSE,
    has_ai_assistance BOOLEAN NOT NULL DEFAULT FALSE,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(36) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(36),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    deleted_by VARCHAR(36),
    PRIMARY KEY (record_id)
);

-- Indexes for medical_records
CREATE INDEX idx_medical_records_patient ON medical_records(patient_id);
CREATE INDEX idx_medical_records_visit_date ON medical_records(patient_id, visit_started_at DESC);
CREATE INDEX idx_medical_records_performer ON medical_records(performed_by, visit_started_at DESC);
CREATE INDEX idx_medical_records_status ON medical_records(status);
CREATE INDEX idx_medical_records_draft ON medical_records(patient_id, status);
CREATE INDEX idx_medical_records_template ON medical_records(template_id);
CREATE INDEX idx_medical_records_schedule ON medical_records(schedule_id);
CREATE INDEX idx_medical_records_source ON medical_records(source_record_id);
