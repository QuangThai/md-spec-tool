package models

import "time"

type Comment struct {
	ID               string       `db:"id" json:"id"`
	SpecID           string       `db:"spec_id" json:"spec_id"`
	UserID           string       `db:"user_id" json:"user_id"`
	Username         string       `json:"username"` // Populated from join
	Content          string       `db:"content" json:"content"`
	ParentCommentID  *string      `db:"parent_comment_id" json:"parent_comment_id"` // For threaded replies
	CreatedAt        time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt        *time.Time   `db:"updated_at" json:"updated_at"`                // NEW: Track last edit
	EditCount        int          `db:"edit_count" json:"edit_count"`                // NEW: Number of edits
	Replies          []*Comment    `json:"replies,omitempty"`                         // Threaded replies
	MentionedUserIDs []string      `json:"mentioned_user_ids,omitempty"`
	EditHistory      []*CommentEdit `json:"edit_history,omitempty"`                   // NEW: Edit audit trail
}

type CommentEdit struct {
	ID              string    `db:"id" json:"id"`
	CommentID       string    `db:"comment_id" json:"comment_id"`
	EditedByUserID  string    `db:"edited_by_user_id" json:"edited_by_user_id"`
	PreviousContent string    `db:"previous_content" json:"previous_content"`
	NewContent      string    `db:"new_content" json:"new_content"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	EditedByName    string    `json:"edited_by_name"` // Populated from join
}

type CommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type CommentEditRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type CommentResponse struct {
	Message  string     `json:"message"`
	Comment  *Comment   `json:"comment,omitempty"`
	Comments []*Comment `json:"comments,omitempty"`
}
