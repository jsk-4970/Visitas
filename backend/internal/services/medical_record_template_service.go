package services

import (
	"context"
	"fmt"

	"github.com/visitas/backend/internal/models"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/pkg/logger"
)

// MedicalRecordTemplateService handles business logic for medical record templates
type MedicalRecordTemplateService struct {
	templateRepo *repository.MedicalRecordTemplateRepository
}

// NewMedicalRecordTemplateService creates a new template service
func NewMedicalRecordTemplateService(templateRepo *repository.MedicalRecordTemplateRepository) *MedicalRecordTemplateService {
	return &MedicalRecordTemplateService{
		templateRepo: templateRepo,
	}
}

// CreateTemplate creates a new medical record template
func (s *MedicalRecordTemplateService) CreateTemplate(ctx context.Context, req *models.MedicalRecordTemplateCreateRequest, createdBy string) (*models.MedicalRecordTemplate, error) {
	// Validate specialty if provided
	if req.Specialty != nil {
		validSpecialties := map[string]bool{
			"general":           true,
			"internal_medicine": true,
			"neurology":         true,
			"palliative_care":   true,
		}
		if !validSpecialties[*req.Specialty] {
			logger.WarnContext(ctx, "Invalid specialty", map[string]interface{}{
				"specialty": *req.Specialty,
			})
			return nil, fmt.Errorf("invalid specialty: %s", *req.Specialty)
		}
	}

	// Validate SOAP template structure
	if len(req.SOAPTemplate) == 0 {
		return nil, fmt.Errorf("soap_template is required")
	}

	// Create template
	template, err := s.templateRepo.Create(ctx, req, createdBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create template", err, map[string]interface{}{
			"template_name": req.TemplateName,
			"created_by":    createdBy,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Template created successfully", map[string]interface{}{
		"template_id":   template.TemplateID,
		"template_name": template.TemplateName,
		"created_by":    createdBy,
	})

	return template, nil
}

// GetTemplate retrieves a template by ID
func (s *MedicalRecordTemplateService) GetTemplate(ctx context.Context, templateID string) (*models.MedicalRecordTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get template", err, map[string]interface{}{
			"template_id": templateID,
		})
		return nil, err
	}

	return template, nil
}

// ListTemplates retrieves templates with filters
func (s *MedicalRecordTemplateService) ListTemplates(ctx context.Context, filter *models.MedicalRecordTemplateFilter) ([]*models.MedicalRecordTemplate, error) {
	templates, err := s.templateRepo.List(ctx, filter)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list templates", err, map[string]interface{}{})
		return nil, err
	}

	logger.InfoContext(ctx, "Templates listed successfully", map[string]interface{}{
		"count": len(templates),
	})

	return templates, nil
}

// UpdateTemplate updates a template
func (s *MedicalRecordTemplateService) UpdateTemplate(ctx context.Context, templateID string, req *models.MedicalRecordTemplateUpdateRequest, updatedBy string) (*models.MedicalRecordTemplate, error) {
	// Validate specialty if provided
	if req.Specialty != nil {
		validSpecialties := map[string]bool{
			"general":           true,
			"internal_medicine": true,
			"neurology":         true,
			"palliative_care":   true,
		}
		if !validSpecialties[*req.Specialty] {
			return nil, fmt.Errorf("invalid specialty: %s", *req.Specialty)
		}
	}

	// Update template
	template, err := s.templateRepo.Update(ctx, templateID, req, updatedBy)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update template", err, map[string]interface{}{
			"template_id": templateID,
			"updated_by":  updatedBy,
		})
		return nil, err
	}

	logger.InfoContext(ctx, "Template updated successfully", map[string]interface{}{
		"template_id": templateID,
		"updated_by":  updatedBy,
	})

	return template, nil
}

// DeleteTemplate soft-deletes a template
func (s *MedicalRecordTemplateService) DeleteTemplate(ctx context.Context, templateID, deletedBy string) error {
	// Check if template exists and is not a system template
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return err
	}

	if template.IsSystemTemplate {
		return fmt.Errorf("cannot delete system template")
	}

	err = s.templateRepo.Delete(ctx, templateID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete template", err, map[string]interface{}{
			"template_id": templateID,
			"deleted_by":  deletedBy,
		})
		return err
	}

	logger.InfoContext(ctx, "Template deleted successfully", map[string]interface{}{
		"template_id": templateID,
		"deleted_by":  deletedBy,
	})

	return nil
}

// GetSystemTemplates retrieves all system templates
func (s *MedicalRecordTemplateService) GetSystemTemplates(ctx context.Context) ([]*models.MedicalRecordTemplate, error) {
	templates, err := s.templateRepo.GetSystemTemplates(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get system templates", err, map[string]interface{}{})
		return nil, err
	}

	return templates, nil
}

// GetTemplatesBySpecialty retrieves templates by specialty
func (s *MedicalRecordTemplateService) GetTemplatesBySpecialty(ctx context.Context, specialty string) ([]*models.MedicalRecordTemplate, error) {
	// Validate specialty
	validSpecialties := map[string]bool{
		"general":           true,
		"internal_medicine": true,
		"neurology":         true,
		"palliative_care":   true,
	}
	if !validSpecialties[specialty] {
		return nil, fmt.Errorf("invalid specialty: %s", specialty)
	}

	templates, err := s.templateRepo.GetBySpecialty(ctx, specialty)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get templates by specialty", err, map[string]interface{}{
			"specialty": specialty,
		})
		return nil, err
	}

	return templates, nil
}

// IncrementUsageCount increments the usage count of a template
func (s *MedicalRecordTemplateService) IncrementUsageCount(ctx context.Context, templateID string) error {
	return s.templateRepo.IncrementUsageCount(ctx, templateID)
}
