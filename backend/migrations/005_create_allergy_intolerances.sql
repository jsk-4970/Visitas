-- Migration: Create allergy_intolerances table
-- Cloud Spanner PostgreSQL Interface
-- Based on FHIR R4 AllergyIntolerance Resource

CREATE TABLE allergy_intolerances (
    allergy_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,

    -- Status (FHIR)
    clinical_status VARCHAR(20) NOT NULL, -- active, inactive, resolved
    verification_status VARCHAR(20) NOT NULL, -- unconfirmed, confirmed, refuted

    -- Type & Category
    type VARCHAR(20) NOT NULL, -- allergy, intolerance
    category VARCHAR(20) NOT NULL, -- food, medication, environment, biologic
    criticality VARCHAR(30) NOT NULL, -- low, high, unable-to-assess

    -- Allergen Code (SNOMED CT, RxNorm, etc.)
    code_system VARCHAR(50),
    code VARCHAR(50),
    display_name VARCHAR(200) NOT NULL, -- アレルゲン名

    -- Reactions (JSONB array)
    reactions JSONB,

    -- Generated Column: Maximum Severity from reactions
    max_severity VARCHAR(20) GENERATED ALWAYS AS (
        (SELECT MAX(CAST(value->>'severity' AS TEXT))
         FROM jsonb_array_elements(COALESCE(reactions, '[]'::jsonb)))
    ) STORED,

    -- Onset Information
    onset_date TIMESTAMPTZ,
    onset_age INT,
    onset_note TEXT,

    -- Last Occurrence
    last_occurrence_date TIMESTAMPTZ,

    -- Recorded Information
    recorded_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recorded_by VARCHAR(100),

    -- Notes
    clinical_notes TEXT,
    patient_comments TEXT,

    -- Audit Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),

    -- Soft Delete
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (allergy_id),
    FOREIGN KEY (patient_id) REFERENCES patients(patient_id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_allergies_patient ON allergy_intolerances(patient_id);
CREATE INDEX idx_allergies_clinical_status ON allergy_intolerances(clinical_status);
CREATE INDEX idx_allergies_active ON allergy_intolerances(patient_id, clinical_status)
    WHERE deleted = FALSE AND clinical_status = 'active';
CREATE INDEX idx_allergies_category ON allergy_intolerances(category);
CREATE INDEX idx_allergies_medication ON allergy_intolerances(patient_id, category)
    WHERE category = 'medication' AND clinical_status = 'active';
CREATE INDEX idx_allergies_high_risk ON allergy_intolerances(patient_id, criticality)
    WHERE criticality = 'high' AND clinical_status = 'active';
CREATE INDEX idx_allergies_display_name ON allergy_intolerances(display_name);

-- Comments
COMMENT ON TABLE allergy_intolerances IS 'Patient allergies and adverse reactions (FHIR AllergyIntolerance)';
COMMENT ON COLUMN allergy_intolerances.reactions IS 'JSONB array of specific reaction events';
COMMENT ON COLUMN allergy_intolerances.max_severity IS 'Auto-computed max severity from reactions array';
