-- Rollback Phase 7: Remove comment edit history tracking

ALTER TABLE activity_logs DROP COLUMN IF EXISTS filters_applied;

DROP INDEX IF EXISTS idx_comments_updated_at;
ALTER TABLE comments DROP COLUMN IF EXISTS updated_at;
ALTER TABLE comments DROP COLUMN IF EXISTS edit_count;

DROP INDEX IF EXISTS idx_comment_edits_created_at;
DROP INDEX IF EXISTS idx_comment_edits_comment_id;
DROP TABLE IF EXISTS comment_edits;
