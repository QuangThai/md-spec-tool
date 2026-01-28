package models

import "time"

type Notification struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	ActorID      string    `db:"actor_id" json:"actor_id"`
	ActorName    string    `json:"actor_name"` // From JOIN
	Type         string    `db:"type" json:"type"`             // 'mention', 'reply', 'share'
	ResourceType string    `db:"resource_type" json:"resource_type"`
	ResourceID   string    `db:"resource_id" json:"resource_id"`
	SpecID       *string   `db:"spec_id" json:"spec_id"`
	Message      *string   `db:"message" json:"message"`
	Read         bool      `db:"read" json:"read"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	ReadAt       *time.Time `db:"read_at" json:"read_at"`
}

type NotificationResponse struct {
	Notifications []*Notification `json:"notifications"`
	UnreadCount   int              `json:"unread_count"`
}

type MarkNotificationReadRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}
