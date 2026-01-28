package services

import (
	"context"

	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/repositories"
)

type NotificationService struct {
	notificationRepo *repositories.NotificationRepository
	mentionRepo      *repositories.MentionRepository
	userRepo         *repositories.UserRepository
}

func NewNotificationService(notificationRepo *repositories.NotificationRepository, mentionRepo *repositories.MentionRepository, userRepo *repositories.UserRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		mentionRepo:      mentionRepo,
		userRepo:         userRepo,
	}
}

// NotifyMention creates notifications for mentioned users
func (s *NotificationService) NotifyMention(ctx context.Context, commentID, specID, actorID string, mentionedUserIDs []string) error {
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil {
		return err
	}

	for _, mentionedUserID := range mentionedUserIDs {
		// Don't notify the author
		if mentionedUserID == actorID {
			continue
		}

		notification := &models.Notification{
			UserID:       mentionedUserID,
			ActorID:      actorID,
			Type:         "mention",
			ResourceType: "comment",
			ResourceID:   commentID,
			SpecID:       &specID,
			Message:      stringPtr(actor.Email + " mentioned you in a comment"),
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			return err
		}
	}

	return nil
}

// NotifyReply creates notification for parent comment author
func (s *NotificationService) NotifyReply(ctx context.Context, parentCommentID, specID, actorID string) error {
	// Note: In real implementation, fetch parent comment from repo to get parent author
	// For now, this is a placeholder

	return nil
}

// NotifyShare creates notification for shared user
func (s *NotificationService) NotifyShare(ctx context.Context, specID, specTitle, sharedWithUserID, actorID string) error {
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		UserID:       sharedWithUserID,
		ActorID:      actorID,
		Type:         "share",
		ResourceType: "spec",
		ResourceID:   specID,
		SpecID:       &specID,
		Message:      stringPtr(actor.Email + " shared " + specTitle + " with you"),
	}

	return s.notificationRepo.Create(ctx, notification)
}

// GetNotifications retrieves user notifications
func (s *NotificationService) GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, int, int, error) {
	notifications, total, err := s.notificationRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}

	unreadCount, err := s.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}

	return notifications, total, unreadCount, nil
}

// MarkAsRead marks notifications as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationIDs []string) error {
	return s.notificationRepo.MarkAsRead(ctx, notificationIDs)
}

// MarkAllAsRead marks all user notifications as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, id string) error {
	return s.notificationRepo.Delete(ctx, id)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
