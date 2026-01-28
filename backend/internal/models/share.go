package models

import "time"

type Share struct {
	ID                 string    `db:"id" json:"id"`
	SpecID             string    `db:"spec_id" json:"spec_id"`
	SharedWithUserID   string    `db:"shared_with_user_id" json:"shared_with_user_id"`
	SharedWithUsername string    `json:"shared_with_username"`                   // Populated from join
	PermissionLevel    string    `db:"permission_level" json:"permission_level"` // 'view' or 'edit'
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

type ShareRequest struct {
	SharedWithUserID string `json:"shared_with_user_id" binding:"required"`
	PermissionLevel  string `json:"permission_level" binding:"required,oneof=view edit"`
}

type ShareResponse struct {
	Message string   `json:"message"`
	Share   *Share   `json:"share,omitempty"`
	Shares  []*Share `json:"shares,omitempty"`
}
