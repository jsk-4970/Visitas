-- Migration: Add optimistic locking to care_plans and medication_orders
-- Cloud Spanner PostgreSQL Interface
-- Adds version column for concurrent update protection

-- Add version column to care_plans
ALTER TABLE care_plans ADD COLUMN version INT NOT NULL DEFAULT 1;

-- Add version column to medication_orders
ALTER TABLE medication_orders ADD COLUMN version INT NOT NULL DEFAULT 1;

-- Comments
COMMENT ON COLUMN care_plans.version IS 'Optimistic locking version counter (incremented on each update)';
COMMENT ON COLUMN medication_orders.version IS 'Optimistic locking version counter (incremented on each update)';
