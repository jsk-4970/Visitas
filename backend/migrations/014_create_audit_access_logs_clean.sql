-- Migration: Create audit_patient_access_logs table (Emulator Compatible)
-- Simplified audit log table matching repository schema

CREATE TABLE audit_patient_access_logs (
    log_id VARCHAR(36) NOT NULL,
    event_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actor_id VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource_id VARCHAR(36),
    patient_id VARCHAR(36),
    accessed_fields TEXT,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    ip_address VARCHAR(50),
    user_agent TEXT,
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (log_id)
);

CREATE INDEX idx_audit_actor ON audit_patient_access_logs(actor_id);
CREATE INDEX idx_audit_patient ON audit_patient_access_logs(patient_id);
CREATE INDEX idx_audit_event_time ON audit_patient_access_logs(event_time DESC);
CREATE INDEX idx_audit_action ON audit_patient_access_logs(action);
CREATE INDEX idx_audit_success ON audit_patient_access_logs(success);
