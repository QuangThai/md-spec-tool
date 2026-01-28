DROP INDEX idx_specs_user_id_not_deleted;
ALTER TABLE specs DROP COLUMN deleted_at;
