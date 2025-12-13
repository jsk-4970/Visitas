-- Migration: Create logistics_locations table
-- Cloud Spanner PostgreSQL Interface
-- ロジスティクス拠点管理（クリニック、事務所、薬局等）
-- Google Maps Route Optimization API連携

CREATE TABLE logistics_locations (
    location_id VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36), -- 所属医療機関・事業所

    -- Location Type
    location_type VARCHAR(50) NOT NULL, -- clinic, office, pharmacy, hospital, care_facility, warehouse

    -- Basic Information
    location_name VARCHAR(300) NOT NULL, -- e.g., "山田在宅クリニック 本院"
    location_code VARCHAR(50), -- Internal code for route optimization

    -- Address
    postal_code VARCHAR(10),
    prefecture VARCHAR(50) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address_line VARCHAR(500) NOT NULL,
    building_name VARCHAR(200),

    -- Full address (generated)
    full_address VARCHAR(800) GENERATED ALWAYS AS (
        COALESCE(postal_code || ' ', '') ||
        prefecture || city || address_line ||
        COALESCE(' ' || building_name, '')
    ) STORED,

    -- Geolocation (critical for route optimization)
    latitude NUMERIC(10, 7) NOT NULL,
    longitude NUMERIC(10, 7) NOT NULL,
    geolocation_verified BOOLEAN NOT NULL DEFAULT FALSE,
    geolocation_verified_at TIMESTAMPTZ,

    -- Google Maps Place ID (for API integration)
    google_place_id VARCHAR(200),
    google_maps_url TEXT,

    -- Contact Information
    phone VARCHAR(50),
    fax VARCHAR(50),
    email VARCHAR(200),
    website_url TEXT,

    -- Operating Hours (JSONB)
    operating_hours JSONB,
    -- Example: {"monday": {"open": "09:00", "close": "18:00", "closed": false}, "tuesday": {...}}

    -- Capacity & Resources
    staff_capacity INT, -- Number of staff members based here
    vehicle_capacity INT, -- Number of vehicles based here
    parking_spots INT,
    has_medical_equipment_storage BOOLEAN NOT NULL DEFAULT FALSE,

    -- Route Optimization Settings
    is_route_start_point BOOLEAN NOT NULL DEFAULT FALSE, -- Can be used as route start
    is_route_end_point BOOLEAN NOT NULL DEFAULT FALSE, -- Can be used as route end
    default_departure_time TIME, -- Default time staff depart from this location
    default_return_time TIME, -- Expected return time

    -- Service Area (JSONB polygon or radius)
    service_area JSONB,
    -- Example: {"type": "radius", "radius_km": 10} or {"type": "polygon", "coordinates": [...]}
    max_service_radius_km NUMERIC(5, 2),

    -- Facilities (JSONB array)
    facilities JSONB, -- ["wheelchair_accessible", "parking", "emergency_equipment", "medication_storage"]

    -- Relationships
    parent_location_id VARCHAR(36), -- For branch offices/clinics

    -- Status
    location_status VARCHAR(30) NOT NULL DEFAULT 'active', -- active, inactive, temporary_closed, relocated

    -- Special Notes
    access_instructions TEXT, -- Parking, entrance instructions
    special_notes TEXT,

    -- Billing/Administrative
    is_billing_location BOOLEAN NOT NULL DEFAULT FALSE,
    billing_code VARCHAR(50),

    -- Audit & Soft Delete
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (location_id),
    FOREIGN KEY (parent_location_id) REFERENCES logistics_locations(location_id) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_locations_organization ON logistics_locations(organization_id);
CREATE INDEX idx_locations_type ON logistics_locations(location_type);
CREATE INDEX idx_locations_status ON logistics_locations(location_status);

-- Critical: Active locations by type
CREATE INDEX idx_locations_active ON logistics_locations(organization_id, location_type, location_status)
    WHERE deleted = FALSE AND location_status = 'active';

-- Geolocation queries (for nearest location search)
CREATE INDEX idx_locations_geolocation ON logistics_locations(latitude, longitude)
    WHERE deleted = FALSE AND location_status = 'active';

-- Route start/end points
CREATE INDEX idx_locations_route_points ON logistics_locations(organization_id, is_route_start_point, is_route_end_point)
    WHERE deleted = FALSE AND (is_route_start_point = TRUE OR is_route_end_point = TRUE);

-- Location name search
CREATE INDEX idx_locations_name ON logistics_locations(location_name);

-- Google Place ID lookup
CREATE INDEX idx_locations_google_place ON logistics_locations(google_place_id)
    WHERE google_place_id IS NOT NULL;

-- Parent-child relationships
CREATE INDEX idx_locations_parent ON logistics_locations(parent_location_id);

-- Prefecture/city lookup
CREATE INDEX idx_locations_prefecture_city ON logistics_locations(prefecture, city);

-- Comments
COMMENT ON TABLE logistics_locations IS 'Healthcare facility locations for route optimization and logistics';
COMMENT ON COLUMN logistics_locations.operating_hours IS 'JSONB: Weekly operating hours schedule';
COMMENT ON COLUMN logistics_locations.service_area IS 'JSONB: Service area definition (radius or polygon)';
COMMENT ON COLUMN logistics_locations.facilities IS 'JSONB array of available facilities and amenities';
COMMENT ON COLUMN logistics_locations.google_place_id IS 'Google Maps Place ID for API integration';
COMMENT ON COLUMN logistics_locations.is_route_start_point IS 'Can be used as starting point for route optimization';

-- Add FK to staff_members for base_location_id
-- (Assuming staff_members table exists from previous migration)
-- This establishes the relationship between staff and their base location
