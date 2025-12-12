package main

import (
	"context"
	"fmt"
	"log"
	"os"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
)

func main() {
	ctx := context.Background()

	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	projectID := "stunning-grin-480914-n1"
	instanceID := "stunning-grin-480914-n1-instance"
	databaseID := "stunning-grin-480914-n1-db"

	// Create instance
	fmt.Println("Creating Spanner instance...")
	instanceAdminClient, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create instance admin client: %v", err)
	}
	defer instanceAdminClient.Close()

	instanceOp, err := instanceAdminClient.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", projectID),
		InstanceId: instanceID,
		Instance: &instancepb.Instance{
			Config:      fmt.Sprintf("projects/%s/instanceConfigs/emulator-config", projectID),
			DisplayName: "Emulator Instance",
			NodeCount:   1,
		},
	})
	if err != nil {
		log.Printf("Warning: Instance creation may have failed (may already exist): %v", err)
	} else {
		_, err = instanceOp.Wait(ctx)
		if err != nil {
			log.Printf("Warning: Instance creation may have failed: %v", err)
		} else {
			fmt.Println("✅ Instance created successfully")
		}
	}

	// Create database
	fmt.Println("Creating Spanner database...")
	dbAdminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create database admin client: %v", err)
	}
	defer dbAdminClient.Close()

	dbOp, err := dbAdminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID),
		CreateStatement: fmt.Sprintf("CREATE DATABASE \"%s\"", databaseID),
		DatabaseDialect: databasepb.DatabaseDialect_POSTGRESQL,
	})
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	_, err = dbOp.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to wait for database creation: %v", err)
	}

	fmt.Println("✅ Database created successfully")
	fmt.Printf("\nInstance: %s\n", instanceID)
	fmt.Printf("Database: %s\n", databaseID)
	fmt.Printf("Dialect: PostgreSQL\n")
}
