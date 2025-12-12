package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/spanner"
)

type TemplateData struct {
	TemplateID          string                 `json:"template_id"`
	TemplateName        string                 `json:"template_name"`
	TemplateDescription string                 `json:"template_description"`
	Specialty           string                 `json:"specialty"`
	IsSystemTemplate    bool                   `json:"is_system_template"`
	SoapTemplate        map[string]interface{} `json:"soap_template"`
}

func main() {
	ctx := context.Background()

	// Set emulator host if not already set
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	projectID := "stunning-grin-480914-n1"
	instanceID := "stunning-grin-480914-n1-instance"
	databaseID := "stunning-grin-480914-n1-db"

	databaseName := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		projectID, instanceID, databaseID)

	// Create Spanner client
	client, err := spanner.NewClient(ctx, databaseName)
	if err != nil {
		log.Fatalf("Failed to create Spanner client: %v", err)
	}
	defer client.Close()

	// Template files to load
	templateFiles := []string{
		"medical_record_templates/soap_template.json",
		"medical_record_templates/specialty_templates/internal_medicine.json",
		"medical_record_templates/specialty_templates/palliative_care.json",
	}

	systemUserID := "system_admin_000001"
	now := time.Now()

	fmt.Println("Seeding system templates...")

	for _, filePath := range templateFiles {
		// Read template file
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", filePath, err)
			continue
		}

		var template TemplateData
		if err := json.Unmarshal(data, &template); err != nil {
			log.Printf("Warning: Failed to parse %s: %v", filePath, err)
			continue
		}

		// Convert soap_template to JSON string for JSONB
		soapTemplateJSON, err := json.Marshal(template.SoapTemplate)
		if err != nil {
			log.Printf("Warning: Failed to marshal soap_template for %s: %v", filePath, err)
			continue
		}

		// Insert template
		_, err = client.Apply(ctx, []*spanner.Mutation{
			spanner.InsertOrUpdate("medical_record_templates",
				[]string{
					"template_id", "template_name", "template_description",
					"specialty", "soap_template", "is_system_template",
					"usage_count", "created_at", "created_by",
					"updated_at", "deleted",
				},
				[]interface{}{
					template.TemplateID,
					template.TemplateName,
					template.TemplateDescription,
					template.Specialty,
					spanner.NullJSON{Value: string(soapTemplateJSON), Valid: true},
					template.IsSystemTemplate,
					0,
					now,
					systemUserID,
					now,
					false,
				},
			),
		})

		if err != nil {
			log.Printf("Warning: Failed to insert template %s: %v", template.TemplateID, err)
			continue
		}

		fmt.Printf("✅ Seeded: %s (%s)\n", template.TemplateName, template.Specialty)
	}

	fmt.Println("\n✅ All system templates seeded successfully!")
}
