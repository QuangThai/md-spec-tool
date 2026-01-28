package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/md-spec-tool/internal/models"
)

type ShareRepository struct {
	pool *pgxpool.Pool
}

func NewShareRepository(pool *pgxpool.Pool) *ShareRepository {
	return &ShareRepository{pool: pool}
}

// Create adds a new share record
func (r *ShareRepository) Create(ctx context.Context, share *models.Share) error {
	query := `
		INSERT INTO shares (spec_id, shared_with_user_id, permission_level)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, share.SpecID, share.SharedWithUserID, share.PermissionLevel).
		Scan(&share.ID, &share.CreatedAt, &share.UpdatedAt)

	if err != nil {
		if err.Error() == "duplicate key value violates unique constraint \"shares_spec_id_shared_with_user_id_key\"" {
			return errors.New("spec already shared with this user")
		}
		return err
	}

	return nil
}

// Delete removes a share
func (r *ShareRepository) Delete(ctx context.Context, specID, sharedWithUserID string) error {
	query := `DELETE FROM shares WHERE spec_id = $1 AND shared_with_user_id = $2`

	result, err := r.pool.Exec(ctx, query, specID, sharedWithUserID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("share not found")
	}

	return nil
}

// GetBySpecID retrieves all shares for a spec with username details
func (r *ShareRepository) GetBySpecID(ctx context.Context, specID string) ([]*models.Share, error) {
	query := `
		SELECT s.id, s.spec_id, s.shared_with_user_id, u.username, s.permission_level, s.created_at, s.updated_at
		FROM shares s
		JOIN users u ON s.shared_with_user_id = u.id
		WHERE s.spec_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, specID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []*models.Share
	for rows.Next() {
		share := &models.Share{}
		err := rows.Scan(&share.ID, &share.SpecID, &share.SharedWithUserID, &share.SharedWithUsername,
			&share.PermissionLevel, &share.CreatedAt, &share.UpdatedAt)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	return shares, rows.Err()
}

// GetPermission returns the permission level for a shared spec
func (r *ShareRepository) GetPermission(ctx context.Context, specID, userID string) (string, error) {
	query := `
		SELECT permission_level FROM shares
		WHERE spec_id = $1 AND shared_with_user_id = $2
	`

	var permission string
	err := r.pool.QueryRow(ctx, query, specID, userID).Scan(&permission)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil // Not shared with this user
		}
		return "", err
	}

	return permission, nil
}

// UpdatePermission updates the permission level
func (r *ShareRepository) UpdatePermission(ctx context.Context, specID, userID, permissionLevel string) error {
	query := `
		UPDATE shares
		SET permission_level = $1, updated_at = CURRENT_TIMESTAMP
		WHERE spec_id = $2 AND shared_with_user_id = $3
	`

	result, err := r.pool.Exec(ctx, query, permissionLevel, specID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("share not found")
	}

	return nil
}

// GetSharedSpecsForUser retrieves specs shared with a user
func (r *ShareRepository) GetSharedSpecsForUser(ctx context.Context, userID string) ([]*models.Spec, error) {
	query := `
		SELECT s.id, s.user_id, s.title, s.content, s.template_id, s.version, s.created_at, s.updated_at, s.deleted_at
		FROM specs s
		JOIN shares sh ON s.id = sh.spec_id
		WHERE sh.shared_with_user_id = $1 AND s.deleted_at IS NULL
		ORDER BY s.updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var specs []*models.Spec
	for rows.Next() {
		spec := &models.Spec{}
		var templateID sql.NullString
		err := rows.Scan(&spec.ID, &spec.UserID, &spec.Title, &spec.Content, &templateID,
			&spec.Version, &spec.CreatedAt, &spec.UpdatedAt, &spec.DeletedAt)
		if err != nil {
			return nil, err
		}
		if templateID.Valid {
			spec.TemplateID = &templateID.String
		}
		specs = append(specs, spec)
	}

	return specs, rows.Err()
}
