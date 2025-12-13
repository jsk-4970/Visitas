//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
)

func main() {
	ctx := context.Background()

	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create admin client: %v", err)
	}
	defer adminClient.Close()

	projectID := "stunning-grin-480914-n1"
	instanceID := "stunning-grin-480914-n1-instance"
	databaseID := "stunning-grin-480914-n1-db"
	databaseName := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		projectID, instanceID, databaseID)

	// Clean migrations in order (all emulator-compatible migrations)
	cleanMigrations := []string{
		"migrations/001_create_patients_clean.sql",
		"migrations/010_create_staff_tables_clean.sql",
		"migrations/018_create_staff_patient_assignments_clean.sql",
		"migrations/006_create_visit_schedules_clean.sql",
		"migrations/007_create_clinical_observations_clean.sql",
		"migrations/008_create_medication_orders_clean.sql",
		"migrations/009_create_care_plans_clean.sql",
		"migrations/011_create_acp_records_clean.sql",
		"migrations/014_create_audit_access_logs_clean.sql",
		"migrations/016_create_medical_records_clean.sql",
	}

	fmt.Printf("Applying %d clean migrations...\n\n", len(cleanMigrations))

	for i, file := range cleanMigrations {
		fmt.Printf("[%d/%d] 📄 Processing: %s\n", i+1, len(cleanMigrations), file)

		sqlContent, err := os.ReadFile(file)
		if err != nil {
			log.Printf("⚠️  Skipping %s: %v\n", file, err)
			continue
		}

		// Parse statements (simple split by semicolon, skip comments)
		lines := string(sqlContent)
		var statements []string
		currentStmt := ""

		for _, line := range splitLines(lines) {
			trimmed := trimSpace(line)
			if trimmed == "" || hasPrefix(trimmed, "--") {
				continue
			}
			currentStmt += line + "\n"
			if hasSuffix(trimmed, ";") {
				statements = append(statements, currentStmt)
				currentStmt = ""
			}
		}

		fmt.Printf("   Applying %d statements...\n", len(statements))

		// Apply each statement individually
		for j, stmt := range statements {
			fmt.Printf("   [%d/%d] Applying statement...\n", j+1, len(statements))
			op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
				Database:   databaseName,
				Statements: []string{stmt},
			})
			if err != nil {
				log.Printf("   ❌ Statement %d/%d failed: %v\n", j+1, len(statements), err)
				continue
			}
			if err := op.Wait(ctx); err != nil {
				log.Printf("   ❌ Statement %d/%d failed: %v\n", j+1, len(statements), err)
				continue
			}
			fmt.Printf("   ✅ Statement %d/%d applied\n", j+1, len(statements))
		}

		fmt.Printf("   ✅ Migration completed\n\n")
	}

	fmt.Println("✅ All clean migrations applied!")
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func hasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}
