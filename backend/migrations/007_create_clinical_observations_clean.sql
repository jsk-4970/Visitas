-- Migration: Create clinical_observations table (Emulator Compatible)
-- Aligned with models/clinical_observation.go (JSONB-based design)

CREATE TABLE clinical_observations (
    observation_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,

    category VARCHAR(50) NOT NULL,
    code JSONB NOT NULL,

    effective_datetime TIMESTAMPTZ NOT NULL,
    issued TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    value JSONB NOT NULL,
    interpretation VARCHAR(30),

    performer_id VARCHAR(100),
    device_id VARCHAR(100),

    visit_record_id VARCHAR(36),

    PRIMARY KEY (observation_id)
);

CREATE INDEX idx_observations_patient ON clinical_observations(patient_id);
CREATE INDEX idx_observations_category ON clinical_observations(category);
CREATE INDEX idx_observations_effective ON clinical_observations(patient_id, effective_datetime);
CREATE INDEX idx_observations_performer ON clinical_observations(performer_id);
CREATE INDEX idx_observations_visit ON clinical_observations(visit_record_id);
