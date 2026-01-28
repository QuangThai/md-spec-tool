package services

import (
	"context"

	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/repositories"
)

type SpecService struct {
	repo *repositories.SpecRepository
}

func NewSpecService(repo *repositories.SpecRepository) *SpecService {
	return &SpecService{repo: repo}
}

// SaveSpec saves a new spec document
func (s *SpecService) SaveSpec(ctx context.Context, spec *models.Spec) error {
	return s.repo.Create(ctx, spec)
}

// GetSpec retrieves a spec by ID
func (s *SpecService) GetSpec(ctx context.Context, id string) (*models.Spec, error) {
	return s.repo.GetByID(ctx, id)
}

// ListSpecs returns all specs for a user
func (s *SpecService) ListSpecs(ctx context.Context, userID string) ([]*models.Spec, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// GetVersions returns all versions of a spec
func (s *SpecService) GetVersions(ctx context.Context, specID string) ([]*models.Spec, error) {
	return s.repo.GetVersions(ctx, specID)
}

// UpdateSpec updates a spec and increments version
func (s *SpecService) UpdateSpec(ctx context.Context, spec *models.Spec) error {
	return s.repo.Update(ctx, spec)
}

// DeleteSpec marks a spec as deleted (soft delete)
func (s *SpecService) DeleteSpec(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}

// SearchSpecs searches for specs by title
func (s *SpecService) SearchSpecs(ctx context.Context, userID, titlePattern string) ([]*models.Spec, error) {
	return s.repo.SearchByTitle(ctx, userID, titlePattern)
}
