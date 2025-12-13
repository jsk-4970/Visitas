package main

import (
	"context"
	"fmt"
	"log"
	"os"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
)

func main() {
	ctx := context.Background()

	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	projectID := "stunning-grin-480914-n1"
	instanceID := "stunning-grin-480914-n1-instance"
	databaseID := "stunning-grin-480914-n1-db"

	dbAdminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create database admin client: %v", err)
	}
	defer dbAdminClient.Close()

	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)

	// Migration statements for Spanner Emulator (PostgreSQL dialect, simplified)
	statements := []string{
		// 001: patients table (simplified - no GENERATED columns)
		`CREATE TABLE patients (
			patient_id VARCHAR(36) NOT NULL,
			birth_date TIMESTAMPTZ,
			gender VARCHAR(20) NOT NULL,
			blood_type VARCHAR(10),
			name_history JSONB NOT NULL,
			contact_points JSONB,
			addresses JSONB,
			consent_details JSONB,
			current_family_name VARCHAR(100),
			current_given_name VARCHAR(100),
			primary_phone VARCHAR(50),
			current_prefecture VARCHAR(50),
			current_city VARCHAR(100),
			consent_status VARCHAR(20) NOT NULL DEFAULT 'pending',
			consent_obtained_at TIMESTAMPTZ,
			consent_withdrawn_at TIMESTAMPTZ,
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			deleted_reason TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100) NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100) NOT NULL,
			PRIMARY KEY (patient_id)
		)`,
		`CREATE INDEX idx_patients_birth_date ON patients(birth_date)`,
		`CREATE INDEX idx_patients_deleted ON patients(deleted)`,
		`CREATE INDEX idx_patients_consent_status ON patients(consent_status)`,

		// 002: social_profiles
		`CREATE TABLE patient_social_profiles (
			profile_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			profile_version INT NOT NULL DEFAULT 1,
			content JSONB NOT NULL,
			lives_alone BOOLEAN,
			requires_caregiver_support BOOLEAN,
			valid_from TIMESTAMPTZ NOT NULL,
			valid_to TIMESTAMPTZ,
			assessed_by VARCHAR(100),
			assessed_at TIMESTAMPTZ,
			assessment_notes TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100),
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			PRIMARY KEY (profile_id)
		)`,
		`CREATE INDEX idx_social_profiles_patient ON patient_social_profiles(patient_id)`,

		// 003: coverages
		`CREATE TABLE patient_coverages (
			coverage_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			insurance_type VARCHAR(30) NOT NULL,
			details JSONB NOT NULL,
			care_level_code VARCHAR(20),
			copay_rate INT,
			valid_from TIMESTAMPTZ NOT NULL,
			valid_to TIMESTAMPTZ,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			priority INT NOT NULL DEFAULT 1,
			verification_status VARCHAR(20) NOT NULL DEFAULT 'unverified',
			verified_at TIMESTAMPTZ,
			verified_by VARCHAR(100),
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100),
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			PRIMARY KEY (coverage_id)
		)`,
		`CREATE INDEX idx_coverages_patient ON patient_coverages(patient_id)`,
		`CREATE INDEX idx_coverages_type ON patient_coverages(insurance_type)`,

		// 004: medical_conditions
		`CREATE TABLE medical_conditions (
			condition_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			clinical_status VARCHAR(20) NOT NULL,
			verification_status VARCHAR(30) NOT NULL,
			category VARCHAR(30),
			severity VARCHAR(30),
			code_system VARCHAR(50),
			code VARCHAR(50),
			display_name VARCHAR(200) NOT NULL,
			body_site VARCHAR(100),
			onset_date DATE,
			onset_age INT,
			onset_note TEXT,
			abatement_date DATE,
			abatement_note TEXT,
			recorded_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			recorded_by VARCHAR(100),
			clinical_notes TEXT,
			patient_comments TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100),
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			PRIMARY KEY (condition_id)
		)`,
		`CREATE INDEX idx_conditions_patient ON medical_conditions(patient_id)`,
		`CREATE INDEX idx_conditions_clinical_status ON medical_conditions(clinical_status)`,

		// 005: allergy_intolerances
		`CREATE TABLE allergy_intolerances (
			allergy_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			clinical_status VARCHAR(20) NOT NULL,
			verification_status VARCHAR(20) NOT NULL,
			type VARCHAR(20) NOT NULL,
			category VARCHAR(20) NOT NULL,
			criticality VARCHAR(30) NOT NULL,
			code_system VARCHAR(50),
			code VARCHAR(50),
			display_name VARCHAR(200) NOT NULL,
			reactions JSONB,
			max_severity VARCHAR(20),
			onset_date DATE,
			onset_age INT,
			onset_note TEXT,
			last_occurrence_date TIMESTAMPTZ,
			recorded_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			recorded_by VARCHAR(100),
			clinical_notes TEXT,
			patient_comments TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100),
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			PRIMARY KEY (allergy_id)
		)`,
		`CREATE INDEX idx_allergies_patient ON allergy_intolerances(patient_id)`,
		`CREATE INDEX idx_allergies_clinical_status ON allergy_intolerances(clinical_status)`,

		// 006: visit_schedules
		`CREATE TABLE visit_schedules (
			schedule_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			visit_date DATE NOT NULL,
			visit_type VARCHAR(50) NOT NULL,
			time_window_start TIMESTAMPTZ,
			time_window_end TIMESTAMPTZ,
			estimated_duration_minutes INT NOT NULL DEFAULT 30,
			assigned_staff_id VARCHAR(36),
			assigned_vehicle_id VARCHAR(36),
			status VARCHAR(30) NOT NULL DEFAULT 'scheduled',
			priority_score INT NOT NULL DEFAULT 0,
			constraints JSONB,
			optimization_result JSONB,
			care_plan_ref VARCHAR(36),
			activity_ref VARCHAR(36),
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (schedule_id)
		)`,
		`CREATE INDEX idx_visit_schedules_patient ON visit_schedules(patient_id)`,
		`CREATE INDEX idx_visit_schedules_date ON visit_schedules(visit_date)`,
		`CREATE INDEX idx_visit_schedules_status ON visit_schedules(status)`,
		`CREATE INDEX idx_visit_schedules_staff ON visit_schedules(assigned_staff_id)`,

		// 007: clinical_observations (matching repository schema)
		`CREATE TABLE clinical_observations (
			observation_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			category VARCHAR(50) NOT NULL,
			code JSONB NOT NULL,
			effective_datetime TIMESTAMPTZ NOT NULL,
			issued TIMESTAMPTZ NOT NULL,
			value JSONB,
			interpretation VARCHAR(50),
			performer_id VARCHAR(36),
			device_id VARCHAR(36),
			visit_record_id VARCHAR(36),
			PRIMARY KEY (observation_id)
		)`,
		`CREATE INDEX idx_observations_patient ON clinical_observations(patient_id)`,
		`CREATE INDEX idx_observations_category ON clinical_observations(category)`,
		`CREATE INDEX idx_observations_date ON clinical_observations(effective_datetime)`,

		// 008: medication_orders (matching repository schema)
		`CREATE TABLE medication_orders (
			order_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			status VARCHAR(30) NOT NULL DEFAULT 'active',
			intent VARCHAR(30) NOT NULL DEFAULT 'order',
			medication JSONB NOT NULL,
			dosage_instruction JSONB,
			prescribed_date TIMESTAMPTZ NOT NULL,
			prescribed_by VARCHAR(100) NOT NULL,
			dispense_pharmacy JSONB,
			reason_reference VARCHAR(36),
			version INT NOT NULL DEFAULT 1,
			PRIMARY KEY (order_id)
		)`,
		`CREATE INDEX idx_medication_orders_patient ON medication_orders(patient_id)`,
		`CREATE INDEX idx_medication_orders_status ON medication_orders(status)`,

		// 009: care_plans
		`CREATE TABLE care_plans (
			plan_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			title VARCHAR(200) NOT NULL,
			description TEXT,
			status VARCHAR(30) NOT NULL DEFAULT 'active',
			intent VARCHAR(30) NOT NULL DEFAULT 'plan',
			category VARCHAR(50),
			period_start TIMESTAMPTZ NOT NULL,
			period_end TIMESTAMPTZ,
			goals JSONB,
			activities JSONB,
			care_team JSONB,
			version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100) NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100),
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			PRIMARY KEY (plan_id)
		)`,
		`CREATE INDEX idx_care_plans_patient ON care_plans(patient_id)`,
		`CREATE INDEX idx_care_plans_status ON care_plans(status)`,

		// 010: staff tables
		`CREATE TABLE staff_members (
			staff_id VARCHAR(36) NOT NULL,
			email VARCHAR(255) NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			role VARCHAR(50) NOT NULL,
			department VARCHAR(100),
			license_number VARCHAR(50),
			license_type VARCHAR(50),
			phone VARCHAR(20),
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			PRIMARY KEY (staff_id)
		)`,
		`CREATE UNIQUE INDEX idx_staff_email ON staff_members(email)`,
		`CREATE INDEX idx_staff_role ON staff_members(role)`,

		// 011: acp_records (matching repository schema)
		`CREATE TABLE acp_records (
			acp_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			recorded_date TIMESTAMPTZ NOT NULL,
			version INT NOT NULL DEFAULT 1,
			status VARCHAR(30) NOT NULL DEFAULT 'active',
			decision_maker VARCHAR(50) NOT NULL,
			proxy_person_id VARCHAR(100),
			directives JSONB,
			values_narrative TEXT,
			legal_documents JSONB,
			discussion_log JSONB,
			data_sensitivity VARCHAR(30) NOT NULL DEFAULT 'highly_confidential',
			access_restricted_to JSONB,
			created_by VARCHAR(100) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (patient_id, acp_id)
		)`,
		`CREATE INDEX idx_acp_records_patient ON acp_records(patient_id)`,
		`CREATE INDEX idx_acp_records_status ON acp_records(status)`,

		// 014: audit_patient_access_logs
		`CREATE TABLE audit_patient_access_logs (
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
		)`,
		`CREATE INDEX idx_audit_actor ON audit_patient_access_logs(actor_id)`,
		`CREATE INDEX idx_audit_patient ON audit_patient_access_logs(patient_id)`,
		`CREATE INDEX idx_audit_event_time ON audit_patient_access_logs(event_time DESC)`,

		// 016: medical_records (matching repository schema)
		`CREATE TABLE medical_records (
			record_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			visit_started_at TIMESTAMPTZ NOT NULL,
			visit_ended_at TIMESTAMPTZ,
			visit_type VARCHAR(50) NOT NULL,
			performed_by VARCHAR(100) NOT NULL,
			status VARCHAR(30) NOT NULL DEFAULT 'draft',
			schedule_id VARCHAR(36),
			soap_content JSONB,
			template_id VARCHAR(36),
			source_record_id VARCHAR(36),
			source_type VARCHAR(30),
			audio_file_url TEXT,
			soap_completed BOOLEAN NOT NULL DEFAULT FALSE,
			has_ai_assistance BOOLEAN NOT NULL DEFAULT FALSE,
			version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100) NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100),
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			deleted_by VARCHAR(100),
			PRIMARY KEY (record_id)
		)`,
		`CREATE INDEX idx_medical_records_patient ON medical_records(patient_id)`,
		`CREATE INDEX idx_medical_records_date ON medical_records(visit_started_at)`,
		`CREATE INDEX idx_medical_records_status ON medical_records(status)`,

		// medical_record_templates (matching repository schema)
		`CREATE TABLE medical_record_templates (
			template_id VARCHAR(36) NOT NULL,
			template_name VARCHAR(200) NOT NULL,
			template_description TEXT,
			specialty VARCHAR(100),
			soap_template JSONB,
			is_system_template BOOLEAN NOT NULL DEFAULT FALSE,
			usage_count INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100) NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100),
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			PRIMARY KEY (template_id)
		)`,
		`CREATE INDEX idx_templates_specialty ON medical_record_templates(specialty)`,
		`CREATE INDEX idx_templates_system ON medical_record_templates(is_system_template)`,

		// 018: staff_patient_assignments
		`CREATE TABLE staff_patient_assignments (
			assignment_id VARCHAR(36) NOT NULL,
			staff_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			role VARCHAR(50) NOT NULL,
			assignment_type VARCHAR(50) NOT NULL DEFAULT 'primary',
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			assigned_by VARCHAR(100) NOT NULL,
			inactivated_at TIMESTAMPTZ,
			inactivated_by VARCHAR(100),
			notes TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (assignment_id)
		)`,
		`CREATE INDEX idx_assignments_staff ON staff_patient_assignments(staff_id)`,
		`CREATE INDEX idx_assignments_patient ON staff_patient_assignments(patient_id)`,
		`CREATE INDEX idx_assignments_status ON staff_patient_assignments(status)`,
		`CREATE UNIQUE INDEX idx_assignments_unique ON staff_patient_assignments(staff_id, patient_id, role)`,

		// patient_identifiers
		`CREATE TABLE patient_identifiers (
			identifier_id VARCHAR(36) NOT NULL,
			patient_id VARCHAR(36) NOT NULL,
			identifier_system VARCHAR(100) NOT NULL,
			identifier_value VARCHAR(100) NOT NULL,
			identifier_type VARCHAR(50),
			period_start TIMESTAMPTZ,
			period_end TIMESTAMPTZ,
			assigner VARCHAR(200),
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by VARCHAR(100),
			deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMPTZ,
			PRIMARY KEY (identifier_id)
		)`,
		`CREATE INDEX idx_identifiers_patient ON patient_identifiers(patient_id)`,
		`CREATE UNIQUE INDEX idx_identifiers_system_value ON patient_identifiers(identifier_system, identifier_value)`,
	}

	fmt.Printf("Applying %d DDL statements to %s...\n", len(statements), dbPath)

	op, err := dbAdminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database:   dbPath,
		Statements: statements,
	})
	if err != nil {
		log.Fatalf("Failed to update database DDL: %v", err)
	}

	if err := op.Wait(ctx); err != nil {
		log.Fatalf("Failed to wait for DDL update: %v", err)
	}

	fmt.Println("✅ All migrations applied successfully!")
}
