package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/md-spec-tool/internal/models"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// Create adds a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (user_id, actor_id, type, resource_type, resource_id, spec_id, message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query, notification.UserID, notification.ActorID, notification.Type,
		notification.ResourceType, notification.ResourceID, notification.SpecID, notification.Message).
		Scan(&notification.ID, &notification.CreatedAt)

	return err
}

// GetByUserID retrieves notifications for a user
func (r *NotificationRepository) GetByUserID(ctx context.Context, userID string, limit int, offset int) ([]*models.Notification, int, error) {
	// Get count
	var count int
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	// Get notifications with actor name
	query := `
		SELECT n.id, n.user_id, n.actor_id, u.username, n.type, n.resource_type, n.resource_id, n.spec_id, n.message, n.read, n.created_at, n.read_at
		FROM notifications n
		JOIN users u ON n.actor_id = u.id
		WHERE n.user_id = $1
		ORDER BY n.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		notif := &models.Notification{}
		err := rows.Scan(&notif.ID, &notif.UserID, &notif.ActorID, &notif.ActorName,
			&notif.Type, &notif.ResourceType, &notif.ResourceID, &notif.SpecID,
			&notif.Message, &notif.Read, &notif.CreatedAt, &notif.ReadAt)
		if err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, notif)
	}

	return notifications, count, rows.Err()
}

// GetUnreadCount returns unread notification count
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// MarkAsRead marks notifications as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationIDs []string) error {
	if len(notificationIDs) == 0 {
		return errors.New("no notification IDs provided")
	}

	query := `
		UPDATE notifications
		SET read = true, read_at = CURRENT_TIMESTAMP
		WHERE id = ANY($1)
	`

	_, err := r.pool.Exec(ctx, query, notificationIDs)
	return err
}

// MarkAllAsRead marks all notifications for user as read
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `
		UPDATE notifications
		SET read = true, read_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND read = false
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// Delete removes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("notification not found")
	}

	return nil
}
