package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/md-spec-tool/internal/models"
)

type CommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

// Create adds a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (spec_id, user_id, content, parent_comment_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	
	err := r.pool.QueryRow(ctx, query, comment.SpecID, comment.UserID, comment.Content, comment.ParentCommentID).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	
	return err
}

// GetByID retrieves a comment by ID
func (r *CommentRepository) GetByID(ctx context.Context, id string) (*models.Comment, error) {
	query := `
		SELECT c.id, c.spec_id, c.user_id, u.username, c.content, c.parent_comment_id, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`
	
	comment := &models.Comment{}
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&comment.ID, &comment.SpecID, &comment.UserID, &comment.Username,
			&comment.Content, &comment.ParentCommentID, &comment.CreatedAt, &comment.UpdatedAt)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("comment not found")
		}
		return nil, err
	}
	
	return comment, nil
}

// GetBySpecID retrieves all top-level comments for a spec with usernames (threads handled separately)
func (r *CommentRepository) GetBySpecID(ctx context.Context, specID string) ([]*models.Comment, error) {
	query := `
		SELECT c.id, c.spec_id, c.user_id, u.username, c.content, c.parent_comment_id, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.spec_id = $1 AND c.parent_comment_id IS NULL
		ORDER BY c.created_at DESC
	`
	
	rows, err := r.pool.Query(ctx, query, specID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		err := rows.Scan(&comment.ID, &comment.SpecID, &comment.UserID, &comment.Username,
			&comment.Content, &comment.ParentCommentID, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	
	return comments, rows.Err()
}

// GetReplies retrieves all replies to a comment
func (r *CommentRepository) GetReplies(ctx context.Context, parentCommentID string) ([]*models.Comment, error) {
	query := `
		SELECT c.id, c.spec_id, c.user_id, u.username, c.content, c.parent_comment_id, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.parent_comment_id = $1
		ORDER BY c.created_at ASC
	`
	
	rows, err := r.pool.Query(ctx, query, parentCommentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var replies []*models.Comment
	for rows.Next() {
		reply := &models.Comment{}
		err := rows.Scan(&reply.ID, &reply.SpecID, &reply.UserID, &reply.Username,
			&reply.Content, &reply.ParentCommentID, &reply.CreatedAt, &reply.UpdatedAt)
		if err != nil {
			return nil, err
		}
		replies = append(replies, reply)
	}
	
	return replies, rows.Err()
}

// Update updates a comment
func (r *CommentRepository) Update(ctx context.Context, id, content string) error {
	query := `
		UPDATE comments
		SET content = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id
	`

	var returnedID string
	err := r.pool.QueryRow(ctx, query, content, id).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("comment not found")
		}
		return err
	}

	return nil
}

// Delete removes a comment
func (r *CommentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM comments WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("comment not found")
	}

	return nil
}

// GetCommentCountBySpec returns the number of comments for a spec
func (r *CommentRepository) GetCommentCountBySpec(ctx context.Context, specID string) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE spec_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, specID).Scan(&count)

	return count, err
}

// PHASE 7: Comment edit history methods

// CreateCommentEdit creates a record of a comment edit
func (r *CommentRepository) CreateCommentEdit(ctx context.Context, commentID, userID, previousContent, newContent string) error {
	query := `
		INSERT INTO comment_edits (id, comment_id, edited_by_user_id, previous_content, new_content)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, commentID, userID, previousContent, newContent)
	return err
}

// GetCommentEdits retrieves all edits for a comment
func (r *CommentRepository) GetCommentEdits(ctx context.Context, commentID string) ([]*models.CommentEdit, error) {
	query := `
		SELECT ce.id, ce.comment_id, ce.edited_by_user_id, ce.previous_content, ce.new_content, ce.created_at, u.username
		FROM comment_edits ce
		JOIN users u ON ce.edited_by_user_id = u.id
		WHERE ce.comment_id = $1
		ORDER BY ce.created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edits []*models.CommentEdit
	for rows.Next() {
		edit := &models.CommentEdit{}
		err := rows.Scan(&edit.ID, &edit.CommentID, &edit.EditedByUserID, &edit.PreviousContent,
			&edit.NewContent, &edit.CreatedAt, &edit.EditedByName)
		if err != nil {
			return nil, err
		}
		edits = append(edits, edit)
	}

	return edits, rows.Err()
}

// UpdateCommentWithEdit updates comment content and increments edit count
func (r *CommentRepository) UpdateCommentWithEdit(ctx context.Context, commentID, newContent string) error {
	query := `
		UPDATE comments
		SET content = $1, updated_at = CURRENT_TIMESTAMP, edit_count = edit_count + 1
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, newContent, commentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("comment not found")
	}

	return nil
}
