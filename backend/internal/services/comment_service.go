package services

import (
	"context"
	"errors"

	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/repositories"
)

type CommentService struct {
	commentRepo        *repositories.CommentRepository
	specRepo           *repositories.SpecRepository
	shareRepo          *repositories.ShareRepository
	mentionRepo        *repositories.MentionRepository
	notificationRepo   *repositories.NotificationRepository
	userRepo           *repositories.UserRepository
}

func NewCommentService(commentRepo *repositories.CommentRepository, specRepo *repositories.SpecRepository, shareRepo *repositories.ShareRepository, mentionRepo *repositories.MentionRepository, notificationRepo *repositories.NotificationRepository, userRepo *repositories.UserRepository) *CommentService {
	return &CommentService{
		commentRepo:      commentRepo,
		specRepo:         specRepo,
		shareRepo:        shareRepo,
		mentionRepo:      mentionRepo,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

// CanCommentOnSpec checks if user can comment on a spec
func (s *CommentService) CanCommentOnSpec(ctx context.Context, specID, userID string) (bool, error) {
	spec, err := s.specRepo.GetByID(ctx, specID)
	if err != nil {
		return false, err
	}

	// Owner can always comment
	if spec.UserID == userID {
		return true, nil
	}

	// Check if shared with "edit" permission
	permission, err := s.shareRepo.GetPermission(ctx, specID, userID)
	if err != nil {
		return false, err
	}

	// Can comment if shared with view or edit permission
	return permission == "view" || permission == "edit", nil
}

// AddComment adds a new comment to a spec
func (s *CommentService) AddComment(ctx context.Context, specID, userID, content string) (*models.Comment, error) {
	// Verify user can comment
	canComment, err := s.CanCommentOnSpec(ctx, specID, userID)
	if err != nil {
		return nil, err
	}

	if !canComment {
		return nil, errors.New("user not allowed to comment on this spec")
	}

	comment := &models.Comment{
		SpecID:  specID,
		UserID:  userID,
		Content: content,
	}

	err = s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Fetch with username
	return s.commentRepo.GetByID(ctx, comment.ID)
}

// GetComments returns all comments for a spec
func (s *CommentService) GetComments(ctx context.Context, specID string) ([]*models.Comment, error) {
	return s.commentRepo.GetBySpecID(ctx, specID)
}

// UpdateComment updates a comment (only by owner)
func (s *CommentService) UpdateComment(ctx context.Context, commentID, userID, content string) (*models.Comment, error) {
	// Get the comment to verify ownership
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if comment.UserID != userID {
		return nil, errors.New("only comment owner can update")
	}

	err = s.commentRepo.Update(ctx, commentID, content)
	if err != nil {
		return nil, err
	}

	// Fetch updated comment
	return s.commentRepo.GetByID(ctx, commentID)
}

// DeleteComment deletes a comment (only by owner or spec owner)
func (s *CommentService) DeleteComment(ctx context.Context, commentID, userID string) error {
	// Get the comment to verify ownership or spec ownership
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Check if user is comment owner
	if comment.UserID == userID {
		return s.commentRepo.Delete(ctx, commentID)
	}

	// Check if user is spec owner
	spec, err := s.specRepo.GetByID(ctx, comment.SpecID)
	if err != nil {
		return err
	}

	if spec.UserID != userID {
		return errors.New("only comment owner or spec owner can delete")
	}

	return s.commentRepo.Delete(ctx, commentID)
}

// GetCommentCount returns the number of comments on a spec
func (s *CommentService) GetCommentCount(ctx context.Context, specID string) (int, error) {
	return s.commentRepo.GetCommentCountBySpec(ctx, specID)
}

// GetCommentsWithReplies returns comments with threaded replies
func (s *CommentService) GetCommentsWithReplies(ctx context.Context, specID string) ([]*models.Comment, error) {
	comments, err := s.commentRepo.GetBySpecID(ctx, specID)
	if err != nil {
		return nil, err
	}

	// Load replies for each comment
	for _, comment := range comments {
		replies, err := s.commentRepo.GetReplies(ctx, comment.ID)
		if err == nil {
			comment.Replies = replies
		}

		// Load mentions
		mentions, err := s.mentionRepo.GetByCommentID(ctx, comment.ID)
		if err == nil {
			comment.MentionedUserIDs = mentions
		}
	}

	return comments, nil
}

// AddReply adds a reply to a comment
func (s *CommentService) AddReply(ctx context.Context, specID, parentCommentID, userID, content string) (*models.Comment, error) {
	// Verify parent comment exists and belongs to same spec
	parentComment, err := s.commentRepo.GetByID(ctx, parentCommentID)
	if err != nil {
		return nil, err
	}

	if parentComment.SpecID != specID {
		return nil, errors.New("parent comment does not belong to this spec")
	}

	// Verify user can comment
	canComment, err := s.CanCommentOnSpec(ctx, specID, userID)
	if err != nil {
		return nil, err
	}

	if !canComment {
		return nil, errors.New("user not allowed to comment on this spec")
	}

	comment := &models.Comment{
		SpecID:          specID,
		UserID:          userID,
		Content:         content,
		ParentCommentID: &parentCommentID,
	}

	err = s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	return s.commentRepo.GetByID(ctx, comment.ID)
}

// PHASE 7: Comment edit methods

// EditComment allows a comment author to edit their comment
func (s *CommentService) EditComment(ctx context.Context, commentID, userID, newContent string) (*models.Comment, error) {
	// Validate new content
	if newContent == "" {
		return nil, errors.New("comment content cannot be empty")
	}
	if len(newContent) > 2000 {
		return nil, errors.New("comment content cannot exceed 2000 characters")
	}

	// Get the comment to verify ownership
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if comment.UserID != userID {
		return nil, errors.New("only comment author can edit")
	}

	// Prevent duplicate edits (no change)
	if comment.Content == newContent {
		return comment, nil
	}

	// Create edit history record
	err = s.commentRepo.CreateCommentEdit(ctx, commentID, userID, comment.Content, newContent)
	if err != nil {
		return nil, err
	}

	// Update the comment with new content and increment edit count
	err = s.commentRepo.UpdateCommentWithEdit(ctx, commentID, newContent)
	if err != nil {
		return nil, err
	}

	// Return updated comment
	return s.commentRepo.GetByID(ctx, commentID)
}

// GetCommentWithEditHistory retrieves a comment with its edit history
func (s *CommentService) GetCommentWithEditHistory(ctx context.Context, commentID string, includeHistory bool) (*models.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if includeHistory {
		edits, err := s.commentRepo.GetCommentEdits(ctx, commentID)
		if err == nil {
			comment.EditHistory = edits
		}
	}

	return comment, nil
}

// GetCommentEdits retrieves the edit history for a comment
func (s *CommentService) GetCommentEdits(ctx context.Context, commentID string) ([]*models.CommentEdit, error) {
	return s.commentRepo.GetCommentEdits(ctx, commentID)
}
