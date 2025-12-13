-- Migration: Create staff and vehicle tables (Emulator Compatible)
-- Removed: GENERATED columns, FOREIGN KEY, WHERE clauses in indexes, COMMENT

-- ========================================
-- 1. Staff Members Table
-- ========================================

CREATE TABLE staff_members (
    staff_id VARCHAR(100) NOT NULL,
    organization_id VARCHAR(36),

    family_name VARCHAR(100) NOT NULL,
    given_name VARCHAR(100) NOT NULL,
    family_name_kana VARCHAR(100),
    given_name_kana VARCHAR(100),

    -- Removed GENERATED column: full_name

    email VARCHAR(200) NOT NULL,
    phone VARCHAR(50),
    emergency_contact_phone VARCHAR(50),

    role VARCHAR(50) NOT NULL,
    qualification_type VARCHAR(50),
    license_number VARCHAR(100),
    license_issued_date DATE,
    license_expiry_date DATE,

    specialties JSONB,
    certifications JSONB,

    work_schedule JSONB,
    availability_status VARCHAR(30) NOT NULL DEFAULT 'available',

    base_location_id VARCHAR(36),
    current_latitude FLOAT8,
    current_longitude FLOAT8,
    last_location_update TIMESTAMPTZ,

    assigned_vehicle_id VARCHAR(36),

    total_patients_assigned INT DEFAULT 0,
    total_visits_completed INT DEFAULT 0,
    average_visit_duration_minutes INT,
    last_visit_date DATE,

    can_prescribe BOOLEAN NOT NULL DEFAULT FALSE,
    can_view_all_patients BOOLEAN NOT NULL DEFAULT FALSE,
    access_level VARCHAR(30) NOT NULL DEFAULT 'standard',

    account_status VARCHAR(30) NOT NULL DEFAULT 'active',
    onboarding_completed BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,

    bio TEXT,
    internal_notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (staff_id)
);

CREATE INDEX idx_staff_organization ON staff_members(organization_id);
CREATE INDEX idx_staff_role ON staff_members(role);
CREATE INDEX idx_staff_email ON staff_members(email);
CREATE INDEX idx_staff_vehicle ON staff_members(assigned_vehicle_id);

-- ========================================
-- 2. Vehicles Table
-- ========================================

CREATE TABLE vehicles (
    vehicle_id VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36),

    vehicle_type VARCHAR(50) NOT NULL,
    make VARCHAR(100),
    model VARCHAR(100),
    year INT,
    color VARCHAR(50),

    license_plate VARCHAR(50) NOT NULL,
    registration_number VARCHAR(100),
    registration_expiry_date DATE,

    insurance_company VARCHAR(200),
    insurance_policy_number VARCHAR(100),
    insurance_expiry_date DATE,

    last_maintenance_date DATE,
    next_maintenance_date DATE,
    odometer_reading INT,
    fuel_type VARCHAR(30),

    medical_equipment JSONB,

    passenger_capacity INT,
    cargo_capacity_kg FLOAT8,

    gps_device_id VARCHAR(100),
    current_latitude FLOAT8,
    current_longitude FLOAT8,
    last_location_update TIMESTAMPTZ,

    vehicle_status VARCHAR(30) NOT NULL DEFAULT 'available',

    currently_assigned_to VARCHAR(100),
    assignment_start_date DATE,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (vehicle_id)
);

CREATE INDEX idx_vehicles_organization ON vehicles(organization_id);
CREATE INDEX idx_vehicles_license_plate ON vehicles(license_plate);
CREATE INDEX idx_vehicles_status ON vehicles(vehicle_status);
CREATE INDEX idx_vehicles_assigned_to ON vehicles(currently_assigned_to);

-- ========================================
-- 3. Staff Assignments Table (for patient-staff relationships)
-- ========================================

CREATE TABLE staff_assignments (
    assignment_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL,
    staff_id VARCHAR(100) NOT NULL,

    assignment_role VARCHAR(50) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,

    start_date DATE NOT NULL,
    end_date DATE,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (assignment_id)
);

CREATE INDEX idx_staff_assignments_patient ON staff_assignments(patient_id);
CREATE INDEX idx_staff_assignments_staff ON staff_assignments(staff_id);
CREATE INDEX idx_staff_assignments_role ON staff_assignments(assignment_role);
