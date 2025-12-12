package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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

	// Get all .sql files from migrations directory
	files, err := filepath.Glob("migrations/[0-9][0-9][0-9]_*.sql")
	if err != nil {
		log.Fatalf("Failed to read migrations: %v", err)
	}
	sort.Strings(files)

	fmt.Printf("Found %d migration files\n", len(files))

	for _, file := range files {
		// Skip emulator-specific variants
		if strings.Contains(file, "_emulator.sql") || strings.Contains(file, "_clean.sql") {
			continue
		}

		fmt.Printf("\n📄 Processing: %s\n", filepath.Base(file))

		sqlContent, err := os.ReadFile(file)
		if err != nil {
			log.Printf("⚠️  Skipping %s: %v\n", file, err)
			continue
		}

		// Split into individual statements
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
				stmt := currentStmt.String()
				// Skip COMMENT ON and unsupported statements
				if !strings.Contains(stmt, "COMMENT ON") &&
					!strings.Contains(stmt, "ON DELETE CASCADE") &&
					!strings.Contains(stmt, "ON DELETE SET NULL") &&
					!strings.Contains(stmt, "GENERATED ALWAYS AS") {
					// Remove WHERE clauses from CREATE INDEX
					stmt = removeIndexWhereClauses(stmt)
					statements = append(statements, stmt)
				}
				currentStmt.Reset()
			}
		}

		if len(statements) == 0 {
			fmt.Printf("⚠️  No statements to apply in %s\n", filepath.Base(file))
			continue
		}

		fmt.Printf("   Applying %d statements...\n", len(statements))

		// Apply each statement
		for i, stmt := range statements {
			op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
				Database:   databaseName,
				Statements: []string{stmt},
			})
			if err != nil {
				log.Printf("   ❌ Statement %d/%d failed: %v\n", i+1, len(statements), err)
				continue
			}
			if err := op.Wait(ctx); err != nil {
				log.Printf("   ❌ Statement %d/%d failed: %v\n", i+1, len(statements), err)
				continue
			}
		}
		fmt.Printf("   ✅ Applied successfully\n")
	}

	fmt.Println("\n✅ All migrations applied!")
}

func removeIndexWhereClauses(stmt string) string {
	// Simple regex-like replacement for WHERE clauses in CREATE INDEX
	lines := strings.Split(stmt, "\n")
	var result strings.Builder
	skipNext := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "WHERE ") {
			skipNext = true
			continue
		}
		if !skipNext {
			result.WriteString(line)
			result.WriteString("\n")
		}
		skipNext = false
	}
	return result.String()
}
