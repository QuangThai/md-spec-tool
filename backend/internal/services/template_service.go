package services

import (
	"context"

	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/repositories"
)

type TemplateService struct {
	repo *repositories.TemplateRepository
}

func NewTemplateService(repo *repositories.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

// GetTemplate fetches a template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, id string) (*models.Template, error) {
	return s.repo.GetByID(ctx, id)
}

// ListTemplates returns all templates for a user
func (s *TemplateService) ListTemplates(ctx context.Context, userID string) ([]*models.Template, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// CreateTemplate creates a new template
func (s *TemplateService) CreateTemplate(ctx context.Context, tmpl *models.Template) error {
	return s.repo.Create(ctx, tmpl)
}

// GetDefaultTemplate returns the system default template
func (s *TemplateService) GetDefaultTemplate(ctx context.Context) (*models.Template, error) {
	return s.repo.GetDefaultTemplate(ctx)
}
