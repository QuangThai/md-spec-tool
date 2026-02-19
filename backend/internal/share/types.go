package share

import "time"

type ShareSummary struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Template  string `json:"template"`
	CreatedAt string `json:"created_at"`
}

// Event represents a resolution event in the share's event log
type Event struct {
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	CommentID string    `json:"comment_id"`
	Data      string    `json:"data"` // JSON-encoded additional data (author, etc.)
}
