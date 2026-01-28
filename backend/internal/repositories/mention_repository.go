package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MentionRepository struct {
	pool *pgxpool.Pool
}

func NewMentionRepository(pool *pgxpool.Pool) *MentionRepository {
	return &MentionRepository{pool: pool}
}

// Create adds a mention record
func (r *MentionRepository) Create(ctx context.Context, commentID, mentionedUserID string) error {
	query := `
		INSERT INTO mentions (comment_id, mentioned_user_id)
		VALUES ($1, $2)
	`

	_, err := r.pool.Exec(ctx, query, commentID, mentionedUserID)
	return err
}

// GetByCommentID retrieves all mentioned user IDs for a comment
func (r *MentionRepository) GetByCommentID(ctx context.Context, commentID string) ([]string, error) {
	query := `SELECT mentioned_user_id FROM mentions WHERE comment_id = $1`

	rows, err := r.pool.Query(ctx, query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mentionedIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		mentionedIDs = append(mentionedIDs, id)
	}

	return mentionedIDs, rows.Err()
}

// DeleteByCommentID removes all mentions for a comment
func (r *MentionRepository) DeleteByCommentID(ctx context.Context, commentID string) error {
	query := `DELETE FROM mentions WHERE comment_id = $1`

	result, err := r.pool.Exec(ctx, query, commentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("no mentions found")
	}

	return nil
}
