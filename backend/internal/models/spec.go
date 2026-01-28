package models

import "time"

type Spec struct {
	ID         string     `db:"id"`
	UserID     string     `db:"user_id"`
	Title      string     `db:"title"`
	Content    string     `db:"content"` // Markdown content
	TemplateID *string    `db:"template_id"`
	Version    int        `db:"version"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"` // NULL if not deleted
}
