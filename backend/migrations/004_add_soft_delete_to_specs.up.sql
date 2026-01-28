ALTER TABLE specs ADD COLUMN deleted_at TIMESTAMP NULL;
CREATE INDEX idx_specs_user_id_not_deleted ON specs(user_id) WHERE deleted_at IS NULL;
