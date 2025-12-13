//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/api/iterator"
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

	// Get database DDL
	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create admin client: %v", err)
	}
	defer adminClient.Close()

	ddl, err := adminClient.GetDatabaseDdl(ctx, &databasepb.GetDatabaseDdlRequest{
		Database: databaseName,
	})
	if err != nil {
		log.Fatalf("Failed to get DDL: %v", err)
	}

	fmt.Println("📋 Database DDL Statements:")
	fmt.Printf("Total statements: %d\n\n", len(ddl.Statements))
	
	tableCount := 0
	indexCount := 0
	
	for i, stmt := range ddl.Statements {
		fmt.Printf("[%d] %s\n", i+1, stmt)
		if len(stmt) >= 12 && stmt[:12] == "CREATE TABLE" {
			tableCount++
		}
		if len(stmt) >= 12 && stmt[:12] == "CREATE INDEX" {
			indexCount++
		}
	}
	
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("Tables: %d\n", tableCount)
	fmt.Printf("Indexes: %d\n", indexCount)

	// Query information schema
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

	fmt.Println("\n📁 Tables in database:")
	count := 0
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
		count++
		fmt.Printf("  %d. %s\n", count, tableName)
	}
	
	if count == 0 {
		fmt.Println("  (No tables found)")
	}
}
