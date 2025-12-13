-- Migration: Create visit_schedules table (Emulator Compatible)
-- Aligned with models/visit_schedule.go

CREATE TABLE visit_schedules (
    schedule_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,
    visit_date DATE NOT NULL,
    visit_type VARCHAR(50) NOT NULL,

    -- Visit time window
    time_window_start TIMESTAMPTZ,
    time_window_end TIMESTAMPTZ,
    estimated_duration_minutes INT NOT NULL DEFAULT 30,

    -- Staff assignment (nullable)
    assigned_staff_id VARCHAR(100),
    assigned_vehicle_id VARCHAR(100),

    -- Status
    status VARCHAR(30) NOT NULL DEFAULT 'draft',

    -- Route Optimization integration
    priority_score INT NOT NULL DEFAULT 5,
    constraints JSONB,
    optimization_result JSONB,

    -- Link to care plan
    care_plan_ref VARCHAR(36),
    activity_ref VARCHAR(36),

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (schedule_id)
);

CREATE INDEX idx_schedules_patient ON visit_schedules(patient_id);
CREATE INDEX idx_schedules_staff ON visit_schedules(assigned_staff_id);
CREATE INDEX idx_schedules_status ON visit_schedules(status);
CREATE INDEX idx_schedules_visit_date ON visit_schedules(visit_date);
CREATE INDEX idx_schedules_care_plan ON visit_schedules(care_plan_ref);
