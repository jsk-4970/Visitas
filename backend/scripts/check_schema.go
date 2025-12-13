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

	dbAdminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer dbAdminClient.Close()

	dbPath := "projects/stunning-grin-480914-n1/instances/stunning-grin-480914-n1-instance/databases/stunning-grin-480914-n1-db"
	resp, err := dbAdminClient.GetDatabaseDdl(ctx, &databasepb.GetDatabaseDdlRequest{Database: dbPath})
	if err != nil {
		log.Fatalf("Failed to get DDL: %v", err)
	}

	fmt.Printf("Found %d DDL statements:\n\n", len(resp.Statements))
	for i, stmt := range resp.Statements {
		fmt.Printf("--- Statement %d ---\n%s\n\n", i+1, stmt)
	}
}
