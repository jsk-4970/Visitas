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

	// Set emulator host if not already set
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	// Database admin client
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

	// Read migration file
	sqlContent, err := os.ReadFile("migrations/016_create_medical_records_clean.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Split into individual statements (simple split by semicolon)
	lines := strings.Split(string(sqlContent), "\n")
	var statements []string
	var currentStmt strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		currentStmt.WriteString(line)
		currentStmt.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, currentStmt.String())
			currentStmt.Reset()
		}
	}

	fmt.Printf("Applying %d DDL statements...\n", len(statements))

	// Apply each statement individually
	for i, stmt := range statements {
		fmt.Printf("[%d/%d] Applying statement...\n", i+1, len(statements))
		op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
			Database:   databaseName,
			Statements: []string{stmt},
		})
		if err != nil {
			log.Fatalf("Failed to apply statement %d: %v\nStatement: %s", i+1, err, stmt)
		}
		if err := op.Wait(ctx); err != nil {
			log.Fatalf("Statement %d failed: %v\nStatement: %s", i+1, err, stmt)
		}
	}

	fmt.Println("✅ Migration 016_create_medical_records.sql applied successfully!")
}
