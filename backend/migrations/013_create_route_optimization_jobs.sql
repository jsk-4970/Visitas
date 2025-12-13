-- Migration: Create route_optimization_jobs table
-- Cloud Spanner PostgreSQL Interface
-- Google Maps Route Optimization API連携記録
-- 訪問ルート最適化ジョブの実行履歴と結果保存

CREATE TABLE route_optimization_jobs (
    job_id VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36),

    -- Job Metadata
    job_type VARCHAR(50) NOT NULL, -- daily_route, weekly_route, emergency_insertion, manual_optimization
    job_status VARCHAR(30) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed, cancelled

    -- Target Date/Period
    target_date DATE NOT NULL, -- The date(s) for which routes are being optimized
    target_start_time TIMESTAMPTZ,
    target_end_time TIMESTAMPTZ,

    -- Staff/Vehicle Assignment
    staff_id VARCHAR(100), -- Staff member for whom route is optimized
    vehicle_id VARCHAR(36), -- Vehicle to be used

    -- Start & End Locations
    start_location_id VARCHAR(36), -- FK to logistics_locations
    end_location_id VARCHAR(36), -- FK to logistics_locations (can be same as start)
    start_latitude NUMERIC(10, 7),
    start_longitude NUMERIC(10, 7),
    end_latitude NUMERIC(10, 7),
    end_longitude NUMERIC(10, 7),

    -- Input Parameters (JSONB)
    optimization_params JSONB NOT NULL,
    -- Example structure:
    -- {
    --   "objective": "minimize_travel_time", // or "minimize_distance", "balance_workload"
    --   "constraints": {
    --     "max_visits_per_route": 12,
    --     "max_route_duration_minutes": 480,
    --     "lunch_break": {"start": "12:00", "duration_minutes": 60},
    --     "priority_visits": ["schedule_id_1", "schedule_id_2"]
    --   },
    --   "preferences": {
    --     "avoid_tolls": false,
    --     "avoid_highways": false,
    --     "traffic_model": "best_guess"
    --   }
    -- }

    -- Visit Schedules Included (JSONB array)
    included_schedule_ids JSONB, -- ["schedule_id_1", "schedule_id_2", ...]
    total_visits_count INT GENERATED ALWAYS AS (
        jsonb_array_length(COALESCE(included_schedule_ids, '[]'::jsonb))
    ) STORED,

    -- Google Maps API Request
    google_api_request_payload JSONB, -- Full request sent to Route Optimization API
    google_api_request_time TIMESTAMPTZ,

    -- Google Maps API Response
    google_api_response_payload JSONB, -- Full response from Route Optimization API
    google_api_response_time TIMESTAMPTZ,
    google_api_computation_time_ms INT,

    -- Optimized Route Results (JSONB)
    optimized_route JSONB,
    -- Example structure:
    -- {
    --   "routes": [{
    --     "vehicle_index": 0,
    --     "visits": [
    --       {"schedule_id": "...", "sequence": 1, "arrival_time": "09:30", "departure_time": "10:00"},
    --       {"schedule_id": "...", "sequence": 2, "arrival_time": "10:25", "departure_time": "10:55"}
    --     ],
    --     "total_distance_meters": 45000,
    --     "total_duration_seconds": 18000,
    --     "polyline": "encoded_polyline_string"
    --   }],
    --   "metrics": {
    --     "total_distance_km": 45.0,
    --     "total_duration_hours": 5.0,
    --     "cost_saved_vs_baseline": 1200
    --   }
    -- }

    -- Performance Metrics (extracted from response)
    total_distance_meters INT,
    total_duration_seconds INT,
    total_distance_km NUMERIC(10, 2) GENERATED ALWAYS AS (
        CAST(total_distance_meters AS NUMERIC) / 1000.0
    ) STORED,
    total_duration_hours NUMERIC(5, 2) GENERATED ALWAYS AS (
        CAST(total_duration_seconds AS NUMERIC) / 3600.0
    ) STORED,

    -- Cost Estimation (based on distance, fuel, time)
    estimated_fuel_cost_jpy INT,
    estimated_toll_cost_jpy INT,
    estimated_total_cost_jpy INT,

    -- Comparison with Previous Route (if re-optimization)
    baseline_job_id VARCHAR(36), -- Previous optimization job for comparison
    improvement_distance_percent NUMERIC(5, 2), -- % reduction in distance
    improvement_duration_percent NUMERIC(5, 2), -- % reduction in time

    -- Application Status
    applied_to_schedules BOOLEAN NOT NULL DEFAULT FALSE, -- Whether optimized route was applied to visit_schedules
    applied_at TIMESTAMPTZ,
    applied_by VARCHAR(100),

    -- Override/Manual Adjustments
    has_manual_overrides BOOLEAN NOT NULL DEFAULT FALSE,
    manual_override_notes TEXT,

    -- Error Handling
    error_code VARCHAR(100),
    error_message TEXT,
    retry_count INT DEFAULT 0,

    -- Execution Timing
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    execution_duration_seconds INT GENERATED ALWAYS AS (
        CASE
            WHEN started_at IS NOT NULL AND completed_at IS NOT NULL
            THEN EXTRACT(EPOCH FROM (completed_at - started_at))::INT
            ELSE NULL
        END
    ) STORED,

    -- User Feedback
    user_rating INT, -- 1-5 stars
    user_feedback TEXT,

    -- Audit & Soft Delete
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (job_id),
    FOREIGN KEY (staff_id) REFERENCES staff_members(staff_id) ON DELETE SET NULL,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(vehicle_id) ON DELETE SET NULL,
    FOREIGN KEY (start_location_id) REFERENCES logistics_locations(location_id) ON DELETE SET NULL,
    FOREIGN KEY (end_location_id) REFERENCES logistics_locations(location_id) ON DELETE SET NULL,
    FOREIGN KEY (baseline_job_id) REFERENCES route_optimization_jobs(job_id) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_route_jobs_organization ON route_optimization_jobs(organization_id);
CREATE INDEX idx_route_jobs_status ON route_optimization_jobs(job_status);
CREATE INDEX idx_route_jobs_staff ON route_optimization_jobs(staff_id);
CREATE INDEX idx_route_jobs_vehicle ON route_optimization_jobs(vehicle_id);

-- Critical: Recent jobs by date and staff
CREATE INDEX idx_route_jobs_recent ON route_optimization_jobs(staff_id, target_date DESC, created_at DESC)
    WHERE deleted = FALSE;

-- Jobs by date
CREATE INDEX idx_route_jobs_target_date ON route_optimization_jobs(target_date, job_status);

-- Applied routes
CREATE INDEX idx_route_jobs_applied ON route_optimization_jobs(applied_to_schedules, applied_at)
    WHERE deleted = FALSE AND applied_to_schedules = TRUE;

-- Failed jobs (for debugging)
CREATE INDEX idx_route_jobs_failed ON route_optimization_jobs(job_status, error_code)
    WHERE job_status = 'failed';

-- Performance tracking
CREATE INDEX idx_route_jobs_performance ON route_optimization_jobs(total_distance_km, total_duration_hours)
    WHERE deleted = FALSE AND job_status = 'completed';

-- Baseline comparison
CREATE INDEX idx_route_jobs_baseline ON route_optimization_jobs(baseline_job_id);

-- User feedback
CREATE INDEX idx_route_jobs_rating ON route_optimization_jobs(user_rating)
    WHERE user_rating IS NOT NULL;

-- Comments
COMMENT ON TABLE route_optimization_jobs IS 'Route optimization job history (Google Maps Route Optimization API)';
COMMENT ON COLUMN route_optimization_jobs.optimization_params IS 'JSONB: Optimization objectives, constraints, and preferences';
COMMENT ON COLUMN route_optimization_jobs.google_api_request_payload IS 'JSONB: Full request sent to Google Route Optimization API';
COMMENT ON COLUMN route_optimization_jobs.google_api_response_payload IS 'JSONB: Full response from Google API';
COMMENT ON COLUMN route_optimization_jobs.optimized_route IS 'JSONB: Structured optimized route with visit sequence and metrics';
COMMENT ON COLUMN route_optimization_jobs.included_schedule_ids IS 'JSONB array of visit_schedules.schedule_id included in optimization';
COMMENT ON COLUMN route_optimization_jobs.total_visits_count IS 'Auto-computed count of visits in the job';
