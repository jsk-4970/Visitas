package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/visitas/backend/internal/models"
	"google.golang.org/api/iterator"
)

// SocialProfileRepository handles social profile data operations
type SocialProfileRepository struct {
	client *spanner.Client
}

// NewSocialProfileRepository creates a new social profile repository
func NewSocialProfileRepository(spannerRepo *SpannerRepository) *SocialProfileRepository {
	return &SocialProfileRepository{
		client: spannerRepo.client,
	}
}

// CreateSocialProfile creates a new social profile record
func (r *SocialProfileRepository) CreateSocialProfile(ctx context.Context, req *models.PatientSocialProfileCreateRequest, createdBy string) (*models.PatientSocialProfile, error) {
	profileID := uuid.New().String()
	now := time.Now()

	// Marshal content to JSONB
	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	mutation := spanner.InsertMap("patient_social_profiles", map[string]interface{}{
		"profile_id":       profileID,
		"patient_id":       req.PatientID,
		"profile_version":  1,
		"content":          string(contentJSON),
		"valid_from":       req.ValidFrom,
		"valid_to":         nil,
		"assessed_by":      req.AssessedBy,
		"assessed_at":      req.AssessedAt,
		"assessment_notes": req.AssessmentNotes,
		"created_at":       now,
		"created_by":       createdBy,
		"updated_at":       now,
		"updated_by":       createdBy,
		"deleted":          false,
		"deleted_at":       nil,
	})

	_, err = r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to insert social profile: %w", err)
	}

	profile, err := r.GetSocialProfileByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created social profile: %w", err)
	}

	return profile, nil
}

// GetSocialProfileByID retrieves a social profile by ID
func (r *SocialProfileRepository) GetSocialProfileByID(ctx context.Context, profileID string) (*models.PatientSocialProfile, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			profile_id, patient_id, profile_version, content,
			lives_alone, requires_caregiver_support,
			valid_from, valid_to,
			assessed_by, assessed_at, assessment_notes,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM patient_social_profiles
		WHERE profile_id = @profileID AND deleted = false`,
		Params: map[string]interface{}{
			"profileID": profileID,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("social profile not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query social profile: %w", err)
	}

	profile, err := scanSocialProfile(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan social profile: %w", err)
	}

	return profile, nil
}

// GetCurrentSocialProfile retrieves the current valid social profile for a patient
func (r *SocialProfileRepository) GetCurrentSocialProfile(ctx context.Context, patientID string) (*models.PatientSocialProfile, error) {
	now := time.Now()

	stmt := spanner.Statement{
		SQL: `SELECT
			profile_id, patient_id, profile_version, content,
			lives_alone, requires_caregiver_support,
			valid_from, valid_to,
			assessed_by, assessed_at, assessment_notes,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM patient_social_profiles
		WHERE patient_id = @patientID
			AND deleted = false
			AND valid_from <= @now
			AND (valid_to IS NULL OR valid_to > @now)
		ORDER BY profile_version DESC
		LIMIT 1`,
		Params: map[string]interface{}{
			"patientID": patientID,
			"now":       now,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no current social profile found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query social profile: %w", err)
	}

	profile, err := scanSocialProfile(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan social profile: %w", err)
	}

	return profile, nil
}

// GetSocialProfileHistory retrieves all social profiles for a patient
func (r *SocialProfileRepository) GetSocialProfileHistory(ctx context.Context, patientID string) ([]*models.PatientSocialProfile, error) {
	stmt := spanner.Statement{
		SQL: `SELECT
			profile_id, patient_id, profile_version, content,
			lives_alone, requires_caregiver_support,
			valid_from, valid_to,
			assessed_by, assessed_at, assessment_notes,
			created_at, created_by, updated_at, updated_by,
			deleted, deleted_at
		FROM patient_social_profiles
		WHERE patient_id = @patientID AND deleted = false
		ORDER BY profile_version DESC`,
		Params: map[string]interface{}{
			"patientID": patientID,
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var profiles []*models.PatientSocialProfile
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate social profiles: %w", err)
		}

		profile, err := scanSocialProfile(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan social profile: %w", err)
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// UpdateSocialProfile updates a social profile record
func (r *SocialProfileRepository) UpdateSocialProfile(ctx context.Context, profileID string, req *models.PatientSocialProfileUpdateRequest, updatedBy string) (*models.PatientSocialProfile, error) {
	now := time.Now()
	updates := make(map[string]interface{})
	updates["profile_id"] = profileID
	updates["updated_at"] = now
	updates["updated_by"] = updatedBy

	if req.Content != nil {
		contentJSON, err := json.Marshal(*req.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal content: %w", err)
		}
		updates["content"] = string(contentJSON)
	}

	if req.ValidTo != nil {
		updates["valid_to"] = *req.ValidTo
	}

	if req.AssessmentNotes != nil {
		updates["assessment_notes"] = *req.AssessmentNotes
	}

	mutation := spanner.UpdateMap("patient_social_profiles", updates)

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to update social profile: %w", err)
	}

	profile, err := r.GetSocialProfileByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated social profile: %w", err)
	}

	return profile, nil
}

// DeleteSocialProfile performs a soft delete on a social profile record
func (r *SocialProfileRepository) DeleteSocialProfile(ctx context.Context, profileID, deletedBy string) error {
	now := time.Now()

	mutation := spanner.UpdateMap("patient_social_profiles", map[string]interface{}{
		"profile_id": profileID,
		"deleted":    true,
		"deleted_at": now,
		"updated_at": now,
		"updated_by": deletedBy,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to delete social profile: %w", err)
	}

	return nil
}

// scanSocialProfile scans a Spanner row into a PatientSocialProfile model
func scanSocialProfile(row *spanner.Row) (*models.PatientSocialProfile, error) {
	var profile models.PatientSocialProfile
	var contentStr string

	err := row.Columns(
		&profile.ProfileID,
		&profile.PatientID,
		&profile.ProfileVersion,
		&contentStr,
		&profile.LivesAlone,
		&profile.RequiresCaregiverSupport,
		&profile.ValidFrom,
		&profile.ValidTo,
		&profile.AssessedBy,
		&profile.AssessedAt,
		&profile.AssessmentNotes,
		&profile.CreatedAt,
		&profile.CreatedBy,
		&profile.UpdatedAt,
		&profile.UpdatedBy,
		&profile.Deleted,
		&profile.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	profile.Content = json.RawMessage(contentStr)

	return &profile, nil
}
