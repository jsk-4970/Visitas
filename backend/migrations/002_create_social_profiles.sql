-- Migration: Create patient_social_profiles table
-- Cloud Spanner PostgreSQL Interface

CREATE TABLE patient_social_profiles (
    profile_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,
    profile_version INT NOT NULL DEFAULT 1,

    -- JSONB Content
    content JSONB NOT NULL,

    -- Generated Columns (for fast queries)
    lives_alone BOOLEAN GENERATED ALWAYS AS (
        CAST(content->'livingSituation'->>'livingAlone' AS BOOLEAN)
    ) STORED,
    requires_caregiver_support BOOLEAN GENERATED ALWAYS AS (
        CAST(content->'livingSituation'->>'requiresCaregiverSupport' AS BOOLEAN)
    ) STORED,

    -- Validity Period
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ,

    -- Assessment Information
    assessed_by VARCHAR(100),
    assessed_at TIMESTAMPTZ,
    assessment_notes TEXT,

    -- Audit Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),

    -- Soft Delete
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (profile_id),
    FOREIGN KEY (patient_id) REFERENCES patients(patient_id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_social_profiles_patient ON patient_social_profiles(patient_id);
CREATE INDEX idx_social_profiles_validity ON patient_social_profiles(valid_from, valid_to);
CREATE INDEX idx_social_profiles_lives_alone ON patient_social_profiles(lives_alone) WHERE lives_alone = TRUE;
CREATE INDEX idx_social_profiles_caregiver_support ON patient_social_profiles(requires_caregiver_support) WHERE requires_caregiver_support = TRUE;
CREATE INDEX idx_social_profiles_current ON patient_social_profiles(patient_id, valid_from, valid_to)
    WHERE deleted = FALSE AND valid_to IS NULL;

-- Comments
COMMENT ON TABLE patient_social_profiles IS 'Patient social context and living situation (FHIR SDOH)';
COMMENT ON COLUMN patient_social_profiles.content IS 'JSONB with living situation, key persons, financial background, social support';
