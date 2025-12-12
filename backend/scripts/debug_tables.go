package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/spanner"
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

	client, err := spanner.NewClient(ctx, databaseName)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Try different queries
	queries := []string{
		"SELECT table_name FROM information_schema.tables WHERE table_catalog = '' AND table_schema = ''",
		"SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'",
		"SELECT table_name FROM information_schema.tables",
	}

	for i, q := range queries {
		fmt.Printf("\n=== Query %d ===\n%s\n", i+1, q)
		stmt := spanner.Statement{SQL: q}
		iter := client.Single().Query(ctx, stmt)
		count := 0
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				break
			}
			var tableName string
			if err := row.Columns(&tableName); err != nil {
				fmt.Printf("Column error: %v\n", err)
				continue
			}
			count++
			fmt.Printf("  %d. %s\n", count, tableName)
		}
		iter.Stop()
		fmt.Printf("Total: %d tables\n", count)
	}
}
