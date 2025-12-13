-- Migration: Create care_plans table (Emulator Compatible)
-- Removed: GENERATED columns, FOREIGN KEY, WHERE clauses in indexes, COMMENT

CREATE TABLE care_plans (
    plan_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'active',
    intent VARCHAR(30) NOT NULL DEFAULT 'plan',

    title VARCHAR(300) NOT NULL,
    description TEXT,

    period_start DATE NOT NULL,
    period_end DATE,

    goals JSONB,
    activities JSONB,

    version INT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),

    PRIMARY KEY (plan_id)
);

CREATE INDEX idx_care_plans_patient ON care_plans(patient_id);
CREATE INDEX idx_care_plans_status ON care_plans(status);
CREATE INDEX idx_care_plans_period ON care_plans(patient_id, period_start, period_end);
