package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/md-spec-tool/internal/models"
)

type TemplateRepository struct {
	pool *pgxpool.Pool
}

func NewTemplateRepository(pool *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{pool: pool}
}

func (r *TemplateRepository) Create(ctx context.Context, tmpl *models.Template) error {
	query := `
		INSERT INTO templates (user_id, name, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, tmpl.UserID, tmpl.Name, tmpl.Content).
		Scan(&tmpl.ID, &tmpl.CreatedAt, &tmpl.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*models.Template, error) {
	tmpl := &models.Template{}
	query := `
		SELECT id, user_id, name, content, created_at, updated_at
		FROM templates
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).
		Scan(&tmpl.ID, &tmpl.UserID, &tmpl.Name, &tmpl.Content, &tmpl.CreatedAt, &tmpl.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return tmpl, nil
}

func (r *TemplateRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Template, error) {
	query := `
		SELECT id, user_id, name, content, created_at, updated_at
		FROM templates
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.Template
	for rows.Next() {
		tmpl := &models.Template{}
		if err := rows.Scan(&tmpl.ID, &tmpl.UserID, &tmpl.Name, &tmpl.Content, &tmpl.CreatedAt, &tmpl.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, tmpl)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *TemplateRepository) GetDefaultTemplate(ctx context.Context) (*models.Template, error) {
	query := `
		SELECT id, user_id, name, content, created_at, updated_at
		FROM templates
		WHERE name = 'Default' AND user_id IS NULL
		LIMIT 1
	`

	tmpl := &models.Template{}
	err := r.pool.QueryRow(ctx, query).
		Scan(&tmpl.ID, &tmpl.UserID, &tmpl.Name, &tmpl.Content, &tmpl.CreatedAt, &tmpl.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default template: %w", err)
	}

	return tmpl, nil
}
