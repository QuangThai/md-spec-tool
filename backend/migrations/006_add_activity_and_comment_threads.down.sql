-- Drop mention tracking
DROP TABLE IF EXISTS mentions;

-- Drop notifications
DROP TABLE IF EXISTS notifications;

-- Remove reply support from comments
ALTER TABLE comments DROP COLUMN IF EXISTS parent_comment_id;

-- Drop activity logs
DROP TABLE IF EXISTS activity_logs;
