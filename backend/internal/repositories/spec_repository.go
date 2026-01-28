package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/md-spec-tool/internal/models"
)

type SpecRepository struct {
	pool *pgxpool.Pool
}

func NewSpecRepository(pool *pgxpool.Pool) *SpecRepository {
	return &SpecRepository{pool: pool}
}

func (r *SpecRepository) Create(ctx context.Context, spec *models.Spec) error {
	query := `
		INSERT INTO specs (user_id, template_id, title, content, version)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	var templateID any = spec.TemplateID
	if spec.TemplateID == nil {
		templateID = nil
	}

	err := r.pool.QueryRow(ctx, query, spec.UserID, templateID, spec.Title, spec.Content, spec.Version).
		Scan(&spec.ID, &spec.CreatedAt, &spec.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create spec: %w", err)
	}

	return nil
}

func (r *SpecRepository) GetByID(ctx context.Context, id string) (*models.Spec, error) {
	spec := &models.Spec{}
	var templateID sql.NullString
	query := `
		SELECT id, user_id, template_id, title, content, version, created_at, updated_at, deleted_at
		FROM specs
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.pool.QueryRow(ctx, query, id).
		Scan(&spec.ID, &spec.UserID, &templateID, &spec.Title, &spec.Content, &spec.Version, &spec.CreatedAt, &spec.UpdatedAt, &spec.DeletedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get spec: %w", err)
	}

	if templateID.Valid {
		spec.TemplateID = &templateID.String
	}

	return spec, nil
}

func (r *SpecRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Spec, error) {
	query := `
		SELECT id, user_id, template_id, title, content, version, created_at, updated_at, deleted_at
		FROM specs
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list specs: %w", err)
	}
	defer rows.Close()

	var specs []*models.Spec
	for rows.Next() {
		spec := &models.Spec{}
		var templateID sql.NullString
		if err := rows.Scan(&spec.ID, &spec.UserID, &templateID, &spec.Title, &spec.Content, &spec.Version, &spec.CreatedAt, &spec.UpdatedAt, &spec.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan spec: %w", err)
		}
		if templateID.Valid {
			spec.TemplateID = &templateID.String
		}
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return specs, nil
}

func (r *SpecRepository) GetVersions(ctx context.Context, specID string) ([]*models.Spec, error) {
	query := `
		SELECT id, user_id, template_id, title, content, version, created_at, updated_at, deleted_at
		FROM specs
		WHERE id = $1 AND deleted_at IS NULL
		ORDER BY version DESC
	`

	rows, err := r.pool.Query(ctx, query, specID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer rows.Close()

	var specs []*models.Spec
	for rows.Next() {
		spec := &models.Spec{}
		var templateID sql.NullString
		if err := rows.Scan(&spec.ID, &spec.UserID, &templateID, &spec.Title, &spec.Content, &spec.Version, &spec.CreatedAt, &spec.UpdatedAt, &spec.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan spec: %w", err)
		}
		if templateID.Valid {
			spec.TemplateID = &templateID.String
		}
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return specs, nil
}

// Update updates a spec and increments version
func (r *SpecRepository) Update(ctx context.Context, spec *models.Spec) error {
	query := `
		UPDATE specs
		SET title = $2, content = $3, version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING version, updated_at
	`

	err := r.pool.QueryRow(ctx, query, spec.ID, spec.Title, spec.Content).
		Scan(&spec.Version, &spec.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update spec: %w", err)
	}

	return nil
}

// SoftDelete marks a spec as deleted without removing it from database
func (r *SpecRepository) SoftDelete(ctx context.Context, id string) error {
	query := `
		UPDATE specs
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete spec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("spec not found or already deleted")
	}

	return nil
}

// SearchByTitle searches specs by title pattern
func (r *SpecRepository) SearchByTitle(ctx context.Context, userID, titlePattern string) ([]*models.Spec, error) {
	query := `
		SELECT id, user_id, template_id, title, content, version, created_at, updated_at, deleted_at
		FROM specs
		WHERE user_id = $1 AND title ILIKE $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID, "%"+titlePattern+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search specs: %w", err)
	}
	defer rows.Close()

	var specs []*models.Spec
	for rows.Next() {
		spec := &models.Spec{}
		var templateID sql.NullString
		if err := rows.Scan(&spec.ID, &spec.UserID, &templateID, &spec.Title, &spec.Content, &spec.Version, &spec.CreatedAt, &spec.UpdatedAt, &spec.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan spec: %w", err)
		}
		if templateID.Valid {
			spec.TemplateID = &templateID.String
		}
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return specs, nil
}
