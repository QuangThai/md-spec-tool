-- Phase 7: Add comment edit history tracking

-- Create comment_edits table for immutable edit history
CREATE TABLE comment_edits (
  id UUID PRIMARY KEY,
  comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
  edited_by_user_id UUID NOT NULL REFERENCES users(id),
  previous_content TEXT NOT NULL,
  new_content TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_comment_edits_comment_id ON comment_edits(comment_id);
CREATE INDEX idx_comment_edits_created_at ON comment_edits(created_at DESC);

-- Add columns to comments table for edit tracking
ALTER TABLE comments ADD COLUMN updated_at TIMESTAMP;
ALTER TABLE comments ADD COLUMN edit_count INT DEFAULT 0;

-- Create indexes on comments for edit queries
CREATE INDEX idx_comments_updated_at ON comments(updated_at DESC);

-- Add filters_applied column to activity_logs (for audit trail of filter queries)
ALTER TABLE activity_logs ADD COLUMN filters_applied JSONB;
