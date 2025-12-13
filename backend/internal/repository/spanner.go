package repository

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"github.com/visitas/backend/internal/config"
)

type SpannerRepository struct {
	client *spanner.Client
}

func NewSpannerRepository(ctx context.Context, cfg *config.Config) (*SpannerRepository, error) {
	// Note: SPANNER_EMULATOR_HOST environment variable is read automatically by the Spanner client

	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		cfg.ProjectID, cfg.SpannerInstance, cfg.SpannerDatabase)

	client, err := spanner.NewClient(ctx, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create spanner client: %w", err)
	}

	return &SpannerRepository{
		client: client,
	}, nil
}

func (r *SpannerRepository) Close() {
	r.client.Close()
}

// Patient repository methods will be implemented in patient_repository.go
