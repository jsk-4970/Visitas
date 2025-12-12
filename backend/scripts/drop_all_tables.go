package main

import (
	"context"
	"fmt"
	"log"
	"os"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/api/iterator"
	"cloud.google.com/go/spanner"
)

func main() {
	ctx := context.Background()

	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	projectID := "stunning-grin-480914-n1"
	instanceID := "stunning-grin-480914-n1-instance"
	databaseID := "stunning-grin-480914-n1-db"
	databaseName := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		projectID, instanceID, databaseID)

	// Get all tables
	client, err := spanner.NewClient(ctx, databaseName)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	stmt := spanner.Statement{
		SQL: `SELECT table_name
		      FROM information_schema.tables
		      WHERE table_schema = 'public'
		      ORDER BY table_name`,
	}

	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var tables []string
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading row: %v", err)
			break
		}

		var tableName string
		if err := row.Columns(&tableName); err != nil {
			log.Printf("Error reading column: %v", err)
			continue
		}
		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		fmt.Println("No tables to drop")
		return
	}

	fmt.Printf("Found %d tables to drop:\n", len(tables))
	for _, table := range tables {
		fmt.Printf("  - %s\n", table)
	}

	// Drop all tables
	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create admin client: %v", err)
	}
	defer adminClient.Close()

	// First, get all indexes
	indexStmt := spanner.Statement{
		SQL: `SELECT index_name, table_name
		      FROM information_schema.indexes
		      WHERE table_schema = 'public' AND index_type != 'PRIMARY_KEY'
		      ORDER BY table_name, index_name`,
	}
	indexIter := client.Single().Query(ctx, indexStmt)

	var dropStatements []string
	for {
		row, err := indexIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading index: %v", err)
			break
		}
		var indexName, tableName string
		if err := row.Columns(&indexName, &tableName); err != nil {
			log.Printf("Error reading index columns: %v", err)
			continue
		}
		dropStatements = append(dropStatements, fmt.Sprintf("DROP INDEX %s", indexName))
	}
	indexIter.Stop()

	fmt.Printf("Found %d indexes to drop\n", len(dropStatements))

	// Then add table drops
	for _, table := range tables {
		dropStatements = append(dropStatements, fmt.Sprintf("DROP TABLE %s", table))
	}

	fmt.Printf("\n🗑️  Dropping %d indexes and %d tables...\n", len(dropStatements)-len(tables), len(tables))
	op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database:   databaseName,
		Statements: dropStatements,
	})
	if err != nil {
		log.Fatalf("Failed to drop tables: %v", err)
	}

	if err := op.Wait(ctx); err != nil {
		log.Fatalf("Failed to wait for drop operation: %v", err)
	}

	fmt.Println("✅ All tables dropped successfully!")
}
