package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"google.golang.org/api/iterator"
)

// MedicalRecordTemplateRepository handles medical record template data operations
type MedicalRecordTemplateRepository struct {
	spannerRepo *SpannerRepository
}

// NewMedicalRecordTemplateRepository creates a new medical record template repository
func NewMedicalRecordTemplateRepository(spannerRepo *SpannerRepository) *MedicalRecordTemplateRepository {
	return &MedicalRecordTemplateRepository{
		spannerRepo: spannerRepo,
	}
}

// Create creates a new medical record template
func (r *MedicalRecordTemplateRepository) Create(ctx context.Context, req *models.MedicalRecordTemplateCreateRequest, createdBy string) (*models.MedicalRecordTemplate, error) {
	templateID := uuid.New().String()
	now := time.Now()

	template := &models.MedicalRecordTemplate{
		TemplateID:       templateID,
		TemplateName:     req.TemplateName,
		IsSystemTemplate: req.IsSystemTemplate,
		UsageCount:       0,
		CreatedAt:        now,
		CreatedBy:        createdBy,
		UpdatedAt:        now,
		Deleted:          false,
	}

	if req.TemplateDescription != nil {
		template.TemplateDescription = req.TemplateDescription
	}
	if req.Specialty != nil {
		template.Specialty = req.Specialty
	}
	template.SOAPTemplate = req.SOAPTemplate

	// Convert optional fields to spanner.Null types
	var description, specialty spanner.NullString
	if req.TemplateDescription != nil {
		description = spanner.NullString{StringVal: *req.TemplateDescription, Valid: true}
	}
	if req.Specialty != nil {
		specialty = spanner.NullString{StringVal: *req.Specialty, Valid: true}
	}

	// Convert JSONB to string
	soapTemplateStr := spanner.NullString{StringVal: string(req.SOAPTemplate), Valid: true}

	mutation := spanner.Insert("medical_record_templates",
		[]string{
			"template_id", "template_name", "template_description", "specialty",
			"soap_template", "is_system_template", "usage_count",
			"created_at", "created_by", "updated_at", "deleted",
		},
		[]interface{}{
			templateID, req.TemplateName, description, specialty,
			soapTemplateStr, req.IsSystemTemplate, 0,
			now, createdBy, now, false,
		},
	)

	_, err := r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

// GetByID retrieves a template by ID
func (r *MedicalRecordTemplateRepository) GetByID(ctx context.Context, templateID string) (*models.MedicalRecordTemplate, error) {
	stmt := NewStatement(`SELECT
			template_id, template_name, template_description, specialty,
			soap_template, is_system_template, usage_count,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM medical_record_templates
		WHERE template_id = @template_id AND deleted = false`,
		map[string]interface{}{
			"template_id": templateID,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("template not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query template: %w", err)
	}

	return scanTemplate(row)
}

// List retrieves templates with filters
func (r *MedicalRecordTemplateRepository) List(ctx context.Context, filter *models.MedicalRecordTemplateFilter) ([]*models.MedicalRecordTemplate, error) {
	conditions := []string{"deleted = false"}
	params := make(map[string]interface{})

	if filter.Specialty != nil {
		conditions = append(conditions, "specialty = @specialty")
		params["specialty"] = *filter.Specialty
	}

	if filter.IsSystemTemplate != nil {
		conditions = append(conditions, "is_system_template = @is_system")
		params["is_system"] = *filter.IsSystemTemplate
	}

	if filter.CreatedBy != nil {
		conditions = append(conditions, "created_by = @created_by")
		params["created_by"] = *filter.CreatedBy
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	limit := 100
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	params["limit"] = int64(limit)
	params["offset"] = int64(offset)

	stmt := NewStatement(fmt.Sprintf(`SELECT
			template_id, template_name, template_description, specialty,
			soap_template, is_system_template, usage_count,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM medical_record_templates
		%s
		ORDER BY usage_count DESC, template_name ASC
		LIMIT @limit OFFSET @offset`, whereClause),
		params)

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var templates []*models.MedicalRecordTemplate
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate templates: %w", err)
		}

		template, err := scanTemplate(row)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// Update updates a template
func (r *MedicalRecordTemplateRepository) Update(ctx context.Context, templateID string, req *models.MedicalRecordTemplateUpdateRequest, updatedBy string) (*models.MedicalRecordTemplate, error) {
	// Get existing template
	existing, err := r.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.TemplateName != nil {
		updates["template_name"] = *req.TemplateName
		existing.TemplateName = *req.TemplateName
	}

	if req.TemplateDescription != nil {
		updates["template_description"] = spanner.NullString{StringVal: *req.TemplateDescription, Valid: true}
		existing.TemplateDescription = req.TemplateDescription
	}

	if req.Specialty != nil {
		updates["specialty"] = spanner.NullString{StringVal: *req.Specialty, Valid: true}
		existing.Specialty = req.Specialty
	}

	if len(req.SOAPTemplate) > 0 {
		updates["soap_template"] = spanner.NullString{StringVal: string(req.SOAPTemplate), Valid: true}
		existing.SOAPTemplate = req.SOAPTemplate
	}

	if len(updates) == 0 {
		return existing, nil
	}

	// Update audit fields
	now := time.Now()
	updates["updated_at"] = now
	updates["updated_by"] = spanner.NullString{StringVal: updatedBy, Valid: true}
	existing.UpdatedAt = now
	existing.UpdatedBy = &updatedBy

	// Build column list and values
	columns := []string{"template_id"}
	values := []interface{}{templateID}

	for col, val := range updates {
		columns = append(columns, col)
		values = append(values, val)
	}

	mutation := spanner.Update("medical_record_templates", columns, values)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return existing, nil
}

// Delete soft-deletes a template
func (r *MedicalRecordTemplateRepository) Delete(ctx context.Context, templateID string) error {
	// Check if template exists
	_, err := r.GetByID(ctx, templateID)
	if err != nil {
		return err
	}

	now := time.Now()
	mutation := spanner.Update("medical_record_templates",
		[]string{"template_id", "deleted", "deleted_at", "updated_at"},
		[]interface{}{
			templateID,
			true,
			spanner.NullTime{Time: now, Valid: true},
			now,
		},
	)

	_, err = r.spannerRepo.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// IncrementUsageCount increments the usage count of a template
func (r *MedicalRecordTemplateRepository) IncrementUsageCount(ctx context.Context, templateID string) error {
	_, err := r.spannerRepo.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		// Read current usage count
		stmt := NewStatement(`SELECT usage_count FROM medical_record_templates WHERE template_id = @template_id`,
			map[string]interface{}{
				"template_id": templateID,
			})

		iter := txn.Query(ctx, stmt)
		defer iter.Stop()

		row, err := iter.Next()
		if err == iterator.Done {
			return fmt.Errorf("template not found")
		}
		if err != nil {
			return err
		}

		var usageCount int64
		if err := row.Columns(&usageCount); err != nil {
			return err
		}

		// Update with incremented count
		mutation := spanner.Update("medical_record_templates",
			[]string{"template_id", "usage_count"},
			[]interface{}{templateID, usageCount + 1},
		)

		return txn.BufferWrite([]*spanner.Mutation{mutation})
	})

	if err != nil {
		return fmt.Errorf("failed to increment usage count: %w", err)
	}

	return nil
}

// GetSystemTemplates retrieves all system templates
func (r *MedicalRecordTemplateRepository) GetSystemTemplates(ctx context.Context) ([]*models.MedicalRecordTemplate, error) {
	stmt := NewStatement(`SELECT
			template_id, template_name, template_description, specialty,
			soap_template, is_system_template, usage_count,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM medical_record_templates
		WHERE is_system_template = true AND deleted = false
		ORDER BY specialty, template_name`,
		nil)

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var templates []*models.MedicalRecordTemplate
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate system templates: %w", err)
		}

		template, err := scanTemplate(row)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// GetBySpecialty retrieves templates by specialty
func (r *MedicalRecordTemplateRepository) GetBySpecialty(ctx context.Context, specialty string) ([]*models.MedicalRecordTemplate, error) {
	stmt := NewStatement(`SELECT
			template_id, template_name, template_description, specialty,
			soap_template, is_system_template, usage_count,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM medical_record_templates
		WHERE specialty = @specialty AND deleted = false
		ORDER BY usage_count DESC, template_name ASC`,
		map[string]interface{}{
			"specialty": specialty,
		})

	iter := r.spannerRepo.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var templates []*models.MedicalRecordTemplate
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate templates: %w", err)
		}

		template, err := scanTemplate(row)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// scanTemplate scans a Spanner row into a MedicalRecordTemplate model
func scanTemplate(row *spanner.Row) (*models.MedicalRecordTemplate, error) {
	var template models.MedicalRecordTemplate

	// Nullable fields
	var description, specialty spanner.NullString
	var soapTemplateStr spanner.NullString
	var updatedBy spanner.NullString
	var deletedAt spanner.NullTime

	err := row.Columns(
		&template.TemplateID,
		&template.TemplateName,
		&description,
		&specialty,
		&soapTemplateStr,
		&template.IsSystemTemplate,
		&template.UsageCount,
		&template.CreatedAt,
		&template.CreatedBy,
		&template.UpdatedAt,
		&updatedBy,
		&template.Deleted,
		&deletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan template: %w", err)
	}

	// Convert nullable fields
	if description.Valid {
		template.TemplateDescription = &description.StringVal
	}
	if specialty.Valid {
		template.Specialty = &specialty.StringVal
	}
	if soapTemplateStr.Valid {
		template.SOAPTemplate = json.RawMessage(soapTemplateStr.StringVal)
	}
	if updatedBy.Valid {
		template.UpdatedBy = &updatedBy.StringVal
	}
	if deletedAt.Valid {
		template.DeletedAt = &deletedAt.Time
	}

	return &template, nil
}
